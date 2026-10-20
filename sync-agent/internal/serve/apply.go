package serve

import (
	"encoding/json"
	"net/http"
	"strings"

	"roblox-sync/sync-agent/internal/mapfs"
)

const maxSourceBytes = 1 << 20

type applyChange struct {
	Op           string  `json:"op"`
	NodeID       string  `json:"nodeId"`
	NewNodeID    string  `json:"newNodeId"`
	ClassName    string  `json:"className"`
	Source       *string `json:"source"`
	BaseRevision *uint64 `json:"baseRevision"`
	Force        bool    `json:"force"`
}

func (s *Server) handleApply(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.Tree.Nodes()
	if err != nil {
		writeMapfsError(w, err)
		return
	}
	if !requireJSONContentType(w, r) {
		return
	}
	payload, err := decodeObject(r)
	if err != nil {
		writeEnvelope(w, http.StatusBadRequest, "BAD_REQUEST", "body fails schema or is invalid JSON")
		return
	}
	rawChanges, ok := payload["changes"]
	if !ok {
		writeEnvelope(w, http.StatusBadRequest, "BAD_REQUEST", "body fails schema or is invalid JSON")
		return
	}
	var changes []applyChange
	if err := json.Unmarshal(rawChanges, &changes); err != nil {
		writeEnvelope(w, http.StatusBadRequest, "BAD_REQUEST", "body fails schema or is invalid JSON")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureRevisionsLocked(nodes)

	for _, c := range changes {
		if code, msg, status := s.validateChangeLocked(c); code != "" {
			writeEnvelope(w, status, code, msg)
			return
		}
	}

	applied := make([]map[string]any, 0)
	conflicts := make([]map[string]any, 0)
	byID := nodesByID(nodes)
	for _, c := range changes {
		entry, conflict, err := s.applyOneLocked(c, byID)
		if err != nil {
			writeMapfsError(w, err)
			return
		}
		if conflict != nil {
			conflicts = append(conflicts, conflict)
			continue
		}
		applied = append(applied, entry)
		if c.Op == "rename" {
			if n, ok := byID[c.NodeID]; ok {
				delete(byID, c.NodeID)
				n.NodeID = c.NewNodeID
				byID[c.NewNodeID] = n
			}
		} else if c.Op == "delete" {
			delete(byID, c.NodeID)
		} else {
			src := ""
			if c.Source != nil {
				src = *c.Source
			}
			byID[c.NodeID] = mapfs.Node{NodeID: c.NodeID, ClassName: c.ClassName, DiskPath: src}
		}
	}
	if len(conflicts) > 0 {
		writeJSON(w, http.StatusConflict, map[string]any{
			"error": map[string]any{
				"code":    "CONFLICT",
				"message": "stale baseRevision",
			},
			"applied":   applied,
			"conflicts": conflicts,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"applied": applied})
}

func (s *Server) validateChangeLocked(c applyChange) (code, msg string, status int) {
	if c.NodeID == "" || c.BaseRevision == nil {
		return "BAD_REQUEST", "body fails schema or is invalid JSON", http.StatusBadRequest
	}
	switch c.Op {
	case "upsert":
		if !validClass(c.ClassName) {
			return "BAD_REQUEST", "body fails schema or is invalid JSON", http.StatusBadRequest
		}
		if c.ClassName == "Folder" {
			if c.Source != nil {
				return "BAD_REQUEST", "body fails schema or is invalid JSON", http.StatusBadRequest
			}
		} else if c.Source == nil {
			return "BAD_REQUEST", "body fails schema or is invalid JSON", http.StatusBadRequest
		} else if len(normalizeLF(*c.Source)) > maxSourceBytes {
			return "PAYLOAD_TOO_LARGE", "source exceeds 1 MiB", http.StatusRequestEntityTooLarge
		}
	case "delete":
	case "rename":
		if c.NewNodeID == "" {
			return "BAD_REQUEST", "body fails schema or is invalid JSON", http.StatusBadRequest
		}
		if !s.nodeMappedLocked(c.NewNodeID) {
			return "NOT_MAPPED", "nodeId outside configured roots", http.StatusBadRequest
		}
	default:
		return "BAD_REQUEST", "body fails schema or is invalid JSON", http.StatusBadRequest
	}
	if !s.nodeMappedLocked(c.NodeID) {
		return "NOT_MAPPED", "nodeId outside configured roots", http.StatusBadRequest
	}
	if c.Force && !s.conflicted[c.NodeID] {
		return "BAD_REQUEST", "force is valid only after CONFLICT for that nodeId", http.StatusBadRequest
	}
	return "", "", 0
}

func (s *Server) applyOneLocked(c applyChange, byID map[string]mapfs.Node) (applied, conflict map[string]any, err error) {
	current, present := s.revisions[c.NodeID]
	base := int(*c.BaseRevision)
	match := present && base == current
	create := !present && base == 0 && c.Op == "upsert"
	if !match && !create {
		if !present {
			current = 0
		}
		s.conflicted[c.NodeID] = true
		return nil, map[string]any{
			"nodeId":          c.NodeID,
			"currentRevision": current,
		}, nil
	}
	newRev := base + 1
	ackID := c.NodeID
	switch c.Op {
	case "upsert":
		src := ""
		if c.Source != nil {
			src = *c.Source
		}
		s.noteMappedWrite(c.NodeID, c.ClassName)
		if err := s.Tree.Upsert(c.NodeID, c.ClassName, src); err != nil {
			return nil, nil, err
		}
		s.revisions[c.NodeID] = newRev
	case "delete":
		s.noteMappedDelete(c.NodeID)
		if err := s.Tree.Delete(c.NodeID); err != nil {
			return nil, nil, err
		}
		delete(s.revisions, c.NodeID)
	case "rename":
		n, ok := byID[c.NodeID]
		if !ok {
			s.conflicted[c.NodeID] = true
			return nil, map[string]any{
				"nodeId":          c.NodeID,
				"currentRevision": current,
			}, nil
		}
		s.noteMappedDelete(c.NodeID)
		s.noteMappedWrite(c.NewNodeID, n.ClassName)
		if err := s.Tree.Rename(c.NodeID, c.NewNodeID, n.ClassName); err != nil {
			return nil, nil, err
		}
		delete(s.revisions, c.NodeID)
		s.revisions[c.NewNodeID] = newRev
		ackID = c.NewNodeID
	default:
		return nil, nil, nil
	}
	delete(s.conflicted, c.NodeID)
	if c.Op == "rename" {
		delete(s.conflicted, c.NewNodeID)
	}
	return map[string]any{"nodeId": ackID, "revision": newRev}, nil, nil
}

func (s *Server) nodeMappedLocked(nodeID string) bool {
	i := strings.IndexByte(nodeID, '/')
	if i <= 0 || i == len(nodeID)-1 {
		return false
	}
	rootID := nodeID[:i]
	for _, r := range s.Tree.StudioRoots() {
		if r.ID == rootID {
			return true
		}
	}
	return false
}

func validClass(className string) bool {
	switch className {
	case "Folder", "Script", "LocalScript", "ModuleScript":
		return true
	default:
		return false
	}
}

func nodesByID(nodes []mapfs.Node) map[string]mapfs.Node {
	out := make(map[string]mapfs.Node, len(nodes))
	for _, n := range nodes {
		out[n.NodeID] = n
	}
	return out
}
