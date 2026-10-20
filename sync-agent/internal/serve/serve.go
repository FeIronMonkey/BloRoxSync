package serve

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"roblox-sync/sync-agent/internal/mapfs"
	"roblox-sync/sync-agent/internal/watch"
)

const DefaultHeartbeat = 15 * time.Second

const jsonContentType = "application/json"

type Tree interface {
	Nodes() ([]mapfs.Node, error)
	ReadSource(nodeID string) (string, error)
	ProjectName() string
	StudioRoots() []mapfs.StudioRoot
	Upsert(nodeID, className, source string) error
	Delete(nodeID string) error
	Rename(oldNodeID, newNodeID, className string) error
	DiskRoots() []string
	AbsFromNode(n mapfs.Node) string
	DiskAbs(nodeID string) (string, error)
	PredictWriteAbs(nodeID, className string) (string, error)
}

type Server struct {
	Tree            Tree
	Token           string
	AgentVersion    string
	ProtocolVersion int

	HeartbeatEvery time.Duration

	mu          sync.Mutex
	revisions   map[string]int
	conflicted  map[string]bool
	log         []changeEvent
	cursor      int
	clients     []*sseClient
	diskIndex   map[string]mapfs.Node
	watcher     *watch.Watcher
	watchCancel context.CancelFunc
}

func New(tree Tree, token, agentVersion string, protocolVersion int) *Server {
	return &Server{
		Tree:            tree,
		Token:           token,
		AgentVersion:    agentVersion,
		ProtocolVersion: protocolVersion,
		revisions:       map[string]int{},
		conflicted:      map[string]bool{},
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/health", s.handleHealth)
	mux.HandleFunc("POST /v1/session", s.authed(s.handleSession))
	mux.HandleFunc("POST /v1/bootstrap/compare", s.authed(s.handleCompare))
	mux.HandleFunc("POST /v1/bootstrap/apply", s.authed(s.handleBootstrapApply))
	mux.HandleFunc("GET /v1/snapshot", s.authed(s.handleSnapshot))
	mux.HandleFunc("POST /v1/apply", s.authed(s.handleApply))
	mux.HandleFunc("GET /v1/events", s.authed(s.handleEvents))
	mux.HandleFunc("GET /v1/changes", s.authed(s.handleChanges))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/health" && r.URL.Query().Get("accessToken") != s.Token {
			writeEnvelope(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing or wrong accessToken")
			return
		}
		h, pattern := mux.Handler(r)
		if pattern == "" || pattern == "/" {
			writeEnvelope(w, http.StatusNotFound, "NOT_FOUND", "not found")
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (s *Server) authed(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("accessToken") != s.Token {
			writeEnvelope(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing or wrong accessToken")
			return
		}
		next(w, r)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":              true,
		"protocolVersion": s.ProtocolVersion,
	})
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	if !requireJSONContentType(w, r) {
		return
	}
	var req struct {
		ProtocolVersion int    `json:"protocolVersion"`
		PluginVersion   string `json:"pluginVersion"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeEnvelope(w, http.StatusBadRequest, "BAD_REQUEST", "body fails schema or is invalid JSON")
		return
	}
	if strings.TrimSpace(req.PluginVersion) == "" {
		writeEnvelope(w, http.StatusBadRequest, "BAD_REQUEST", "body fails schema or is invalid JSON")
		return
	}
	if req.ProtocolVersion != s.ProtocolVersion {
		writeEnvelope(w, http.StatusConflict, "PROTOCOL_MISMATCH", "plugin protocolVersion does not match rbx-sync")
		return
	}
	s.endStreams("replaced")
	roots := s.Tree.StudioRoots()
	outRoots := make([]map[string]string, 0, len(roots))
	for _, root := range roots {
		outRoots = append(outRoots, map[string]string{
			"id":     root.ID,
			"studio": root.Studio,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"protocolVersion": s.ProtocolVersion,
		"agentVersion":    s.AgentVersion,
		"projectName":     s.Tree.ProjectName(),
		"roots":           outRoots,
	})
}

func (s *Server) handleSnapshot(w http.ResponseWriter, _ *http.Request) {
	nodes, err := s.snapshotNodes()
	if err != nil {
		writeMapfsError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"nodes": nodes})
}

func (s *Server) snapshotNodes() ([]map[string]any, error) {
	nodes, err := s.Tree.Nodes()
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureRevisionsLocked(nodes)
	out := make([]map[string]any, 0, len(nodes))
	for _, n := range nodes {
		item := map[string]any{
			"nodeId":    n.NodeID,
			"className": n.ClassName,
			"revision":  s.revisions[n.NodeID],
		}
		if n.ClassName != "Folder" {
			src, err := s.Tree.ReadSource(n.NodeID)
			if err != nil {
				return nil, err
			}
			item["source"] = src
		}
		out = append(out, item)
	}
	return out, nil
}

func (s *Server) ensureRevisionsLocked(nodes []mapfs.Node) {
	if s.revisions == nil {
		s.revisions = map[string]int{}
	}
	if s.conflicted == nil {
		s.conflicted = map[string]bool{}
	}
	for _, n := range nodes {
		if _, ok := s.revisions[n.NodeID]; !ok {
			s.revisions[n.NodeID] = 1
		}
	}
}

func (s *Server) resetRevisionsLocked(nodes []mapfs.Node) {
	s.revisions = map[string]int{}
	s.conflicted = map[string]bool{}
	for _, n := range nodes {
		s.revisions[n.NodeID] = 1
	}
}

func writeMapfsError(w http.ResponseWriter, err error) {
	var e *mapfs.Error
	if errors.As(err, &e) {
		switch e.Code {
		case mapfs.CodeIOError:
			writeEnvelope(w, http.StatusInternalServerError, "IO_ERROR", "filesystem read or write failed")
			return
		case mapfs.CodePayloadTooLarge:
			writeEnvelope(w, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "source exceeds 1 MiB")
			return
		case mapfs.CodeNotMapped:
			writeEnvelope(w, http.StatusBadRequest, "NOT_MAPPED", "nodeId outside configured roots")
			return
		}
	}
	writeEnvelope(w, http.StatusInternalServerError, "INTERNAL", "internal error")
}

func requireJSONContentType(w http.ResponseWriter, r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	if ct != jsonContentType && !strings.HasPrefix(ct, jsonContentType+";") {
		writeEnvelope(w, http.StatusBadRequest, "BAD_REQUEST", "body fails schema or is invalid JSON")
		return false
	}
	return true
}

func decodeJSON(r *http.Request, dest any) error {
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dest); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("extra JSON")
	}
	return nil
}

func writeEnvelope(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
