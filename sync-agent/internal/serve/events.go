package serve

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"roblox-sync/sync-agent/internal/mapfs"
	"roblox-sync/sync-agent/internal/watch"
)

type changeEvent struct {
	Cursor    int     `json:"cursor"`
	Origin    string  `json:"origin"`
	Op        string  `json:"op"`
	NodeID    string  `json:"nodeId"`
	ClassName string  `json:"className,omitempty"`
	Source    *string `json:"source,omitempty"`
	NewNodeID string  `json:"newNodeId,omitempty"`
	Revision  int     `json:"revision"`
}

type sseClient struct {
	ch   chan string
	done chan struct{}
	once sync.Once
}

func (c *sseClient) stop() {
	c.once.Do(func() { close(c.done) })
}

func (s *Server) heartbeatEvery() time.Duration {
	if s.HeartbeatEvery > 0 {
		return s.HeartbeatEvery
	}
	return DefaultHeartbeat
}

func (s *Server) StartWatch() error {
	ctx, cancel := context.WithCancel(context.Background())
	s.watchCancel = cancel
	roots := s.Tree.DiskRoots()
	for _, root := range roots {
		if err := os.MkdirAll(root, 0o755); err != nil {
			cancel()
			return err
		}
	}
	w, err := watch.Start(ctx, roots)
	if err != nil {
		cancel()
		return err
	}
	s.watcher = w
	if nodes, err := s.Tree.Nodes(); err == nil {
		s.mu.Lock()
		s.diskIndex = s.indexAbsLocked(nodes)
		s.mu.Unlock()
	}
	go s.consumeWatch()
	return nil
}

func (s *Server) consumeWatch() {
	if s.watcher == nil {
		return
	}
	for ev := range s.watcher.Events() {
		s.emitDiskPath(ev.Path)
	}
}

func (s *Server) Close(reason string) {
	s.endStreams(reason)
	if s.watchCancel != nil {
		s.watchCancel()
	}
	if s.watcher != nil {
		_ = s.watcher.Close()
	}
}

func (s *Server) noteDiskWrite(path string) {
	if s.watcher != nil {
		s.watcher.NoteWrite(path)
	}
}

func (s *Server) noteMappedWrite(nodeID, className string) {
	if path, err := s.Tree.PredictWriteAbs(nodeID, className); err == nil {
		s.noteDiskWrite(path)
	}
}

func (s *Server) noteMappedDelete(nodeID string) {
	if path, err := s.Tree.DiskAbs(nodeID); err == nil {
		s.noteDiskWrite(path)
	}
}

func (s *Server) refreshDiskIndex() {
	nodes, err := s.Tree.Nodes()
	if err != nil {
		return
	}
	s.mu.Lock()
	s.diskIndex = s.indexAbsLocked(nodes)
	s.mu.Unlock()
}

func (s *Server) indexAbsLocked(nodes []mapfs.Node) map[string]mapfs.Node {
	out := make(map[string]mapfs.Node, len(nodes))
	for _, n := range nodes {
		out[filepath.Clean(s.Tree.AbsFromNode(n))] = n
	}
	return out
}

func (s *Server) emitDiskPath(abs string) {
	nodes, err := s.Tree.Nodes()
	if err != nil {
		return
	}
	absClean := filepath.Clean(abs)
	var match *mapfs.Node
	for i := range nodes {
		if filepath.Clean(s.Tree.AbsFromNode(nodes[i])) == absClean {
			n := nodes[i]
			match = &n
			break
		}
	}
	var source string
	if match != nil && match.ClassName != "Folder" {
		source, err = s.Tree.ReadSource(match.NodeID)
		if err != nil {
			return
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.diskIndex == nil {
		s.diskIndex = map[string]mapfs.Node{}
	}
	if match != nil {
		rev := s.revisions[match.NodeID]
		if rev == 0 {
			rev = 1
		} else {
			rev++
		}
		s.revisions[match.NodeID] = rev
		ev := changeEvent{
			Origin:    "disk",
			Op:        "upsert",
			NodeID:    match.NodeID,
			ClassName: match.ClassName,
			Revision:  rev,
		}
		if match.ClassName != "Folder" {
			src := source
			ev.Source = &src
		}
		s.appendEventLocked(ev)
		s.diskIndex[absClean] = *match
		return
	}
	if n, ok := s.diskIndex[absClean]; ok {
		rev := s.revisions[n.NodeID] + 1
		if s.revisions[n.NodeID] == 0 {
			rev = 1
		}
		delete(s.revisions, n.NodeID)
		s.appendEventLocked(changeEvent{
			Origin:   "disk",
			Op:       "delete",
			NodeID:   n.NodeID,
			Revision: rev,
		})
		delete(s.diskIndex, absClean)
	}
}

func (s *Server) appendEventLocked(ev changeEvent) {
	s.cursor++
	ev.Cursor = s.cursor
	s.log = append(s.log, ev)
	msg := formatSSE("change", strconv.Itoa(ev.Cursor), mustJSON(ev))
	for _, c := range s.clients {
		select {
		case c.ch <- msg:
		default:
		}
	}
}

func (s *Server) eventsAfterLocked(since int) []changeEvent {
	out := make([]changeEvent, 0)
	for _, ev := range s.log {
		if ev.Cursor > since {
			out = append(out, ev)
		}
	}
	return out
}

func (s *Server) endStreams(reason string) {
	s.mu.Lock()
	clients := s.clients
	s.clients = nil
	s.mu.Unlock()
	msg := formatSSE("end", "", mustJSON(map[string]string{"reason": reason}))
	for _, c := range clients {
		select {
		case c.ch <- msg:
		case <-time.After(2 * time.Second):
		}
	}
}

func (s *Server) removeClient(c *sseClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.clients[:0]
	for _, x := range s.clients {
		if x != c {
			out = append(out, x)
		}
	}
	s.clients = out
}

func (s *Server) handleChanges(w http.ResponseWriter, r *http.Request) {
	since := 0
	if raw := r.URL.Query().Get("since"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			writeEnvelope(w, http.StatusBadRequest, "BAD_REQUEST", "body fails schema or is invalid JSON")
			return
		}
		since = n
	}
	s.mu.Lock()
	changes := s.eventsAfterLocked(since)
	cursor := s.cursor
	s.mu.Unlock()
	if changes == nil {
		changes = []changeEvent{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"cursor":  cursor,
		"changes": changes,
	})
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeEnvelope(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		return
	}
	since, hasSince := 0, false
	if raw := r.URL.Query().Get("since"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			writeEnvelope(w, http.StatusBadRequest, "BAD_REQUEST", "body fails schema or is invalid JSON")
			return
		}
		since, hasSince = n, true
	}

	c := &sseClient{ch: make(chan string, 32), done: make(chan struct{})}
	s.mu.Lock()
	if !hasSince {
		since = s.cursor
	}
	backlog := s.eventsAfterLocked(since)
	s.clients = append(s.clients, c)
	s.mu.Unlock()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	for _, ev := range backlog {
		if _, err := fmt.Fprint(w, formatSSE("change", strconv.Itoa(ev.Cursor), mustJSON(ev))); err != nil {
			s.removeClient(c)
			return
		}
	}
	flusher.Flush()

	tick := time.NewTicker(s.heartbeatEvery())
	defer tick.Stop()
	defer s.removeClient(c)
	defer c.stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-c.done:
			return
		case msg := <-c.ch:
			if _, err := fmt.Fprint(w, msg); err != nil {
				return
			}
			flusher.Flush()
			if isEndEvent(msg) {
				return
			}
		case <-tick.C:
			if _, err := fmt.Fprint(w, formatSSE("heartbeat", "", "{}")); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func isEndEvent(msg string) bool {
	return len(msg) >= 10 && msg[:10] == "event: end"
}

func formatSSE(event, id, data string) string {
	out := ""
	if id != "" {
		out += "id: " + id + "\n"
	}
	out += "event: " + event + "\n"
	out += "data: " + data + "\n\n"
	return out
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}
