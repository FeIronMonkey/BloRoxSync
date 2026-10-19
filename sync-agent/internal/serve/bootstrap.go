package serve

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
)

type wireNode struct {
	NodeID    string  `json:"nodeId"`
	ClassName string  `json:"className"`
	Source    *string `json:"source"`
}

func (s *Server) handleCompare(w http.ResponseWriter, r *http.Request) {
	if !requireJSONContentType(w, r) {
		return
	}
	payload, err := decodeObject(r)
	if err != nil {
		writeEnvelope(w, http.StatusBadRequest, "BAD_REQUEST", "body fails schema or is invalid JSON")
		return
	}
	rawNodes, ok := payload["nodes"]
	if !ok {
		writeEnvelope(w, http.StatusBadRequest, "BAD_REQUEST", "body fails schema or is invalid JSON")
		return
	}
	var studio []wireNode
	if err := json.Unmarshal(rawNodes, &studio); err != nil {
		writeEnvelope(w, http.StatusBadRequest, "BAD_REQUEST", "body fails schema or is invalid JSON")
		return
	}
	disk, err := s.snapshotNodes()
	if err != nil {
		writeMapfsError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"equal": snapshotsEqual(disk, studio)})
}

func (s *Server) handleBootstrapApply(w http.ResponseWriter, r *http.Request) {
	if !requireJSONContentType(w, r) {
		return
	}
	payload, err := decodeObject(r)
	if err != nil {
		writeEnvelope(w, http.StatusBadRequest, "BAD_REQUEST", "body fails schema or is invalid JSON")
		return
	}
	var direction string
	if raw, ok := payload["direction"]; ok {
		if err := json.Unmarshal(raw, &direction); err != nil {
			writeEnvelope(w, http.StatusBadRequest, "BAD_REQUEST", "body fails schema or is invalid JSON")
			return
		}
	}
	_, hasNodes := payload["nodes"]
	switch direction {
	case "disk-to-studio":
		if hasNodes {
			writeEnvelope(w, http.StatusBadRequest, "BAD_REQUEST", "disk-to-studio must not include nodes")
			return
		}
		nodes, err := s.snapshotNodes()
		if err != nil {
			writeMapfsError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"nodes": nodes})
	case "studio-to-disk":
		if !hasNodes {
			writeEnvelope(w, http.StatusBadRequest, "BAD_REQUEST", "studio-to-disk requires nodes")
			return
		}
		var studio []wireNode
		if err := json.Unmarshal(payload["nodes"], &studio); err != nil {
			writeEnvelope(w, http.StatusBadRequest, "BAD_REQUEST", "body fails schema or is invalid JSON")
			return
		}
		if err := s.applyStudioSnapshot(studio); err != nil {
			writeMapfsError(w, err)
			return
		}
		nodes, err := s.snapshotNodes()
		if err != nil {
			writeMapfsError(w, err)
			return
		}
		applied := make([]map[string]any, 0, len(nodes))
		for _, n := range nodes {
			applied = append(applied, map[string]any{
				"nodeId":   n["nodeId"],
				"revision": 1,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"applied": applied})
	default:
		writeEnvelope(w, http.StatusBadRequest, "BAD_REQUEST", "body fails schema or is invalid JSON")
	}
}

func (s *Server) applyStudioSnapshot(studio []wireNode) error {
	disk, err := s.Tree.Nodes()
	if err != nil {
		return err
	}
	keep := make(map[string]struct{}, len(studio))
	for _, n := range studio {
		keep[n.NodeID] = struct{}{}
		source := ""
		if n.Source != nil {
			source = *n.Source
		}
		if err := s.Tree.Upsert(n.NodeID, n.ClassName, source); err != nil {
			return err
		}
	}
	var omitted []string
	for _, n := range disk {
		if _, ok := keep[n.NodeID]; ok {
			continue
		}
		omitted = append(omitted, n.NodeID)
	}
	sort.Slice(omitted, func(i, j int) bool {
		return len(omitted[i]) > len(omitted[j])
	})
	for _, id := range omitted {
		if err := s.Tree.Delete(id); err != nil {
			return err
		}
	}
	return nil
}

func snapshotsEqual(disk []map[string]any, studio []wireNode) bool {
	if len(disk) != len(studio) {
		return false
	}
	byID := make(map[string]map[string]any, len(disk))
	for _, n := range disk {
		id, _ := n["nodeId"].(string)
		byID[id] = n
	}
	for _, s := range studio {
		d, ok := byID[s.NodeID]
		if !ok {
			return false
		}
		class, _ := d["className"].(string)
		if class != s.ClassName {
			return false
		}
		if s.ClassName == "Folder" {
			continue
		}
		diskSrc, _ := d["source"].(string)
		studioSrc := ""
		if s.Source != nil {
			studioSrc = *s.Source
		}
		if normalizeLF(diskSrc) != normalizeLF(studioSrc) {
			return false
		}
	}
	return true
}

func normalizeLF(source string) string {
	source = strings.TrimPrefix(source, "\uFEFF")
	source = strings.ReplaceAll(source, "\r\n", "\n")
	source = strings.ReplaceAll(source, "\r", "\n")
	return source
}

func decodeObject(r *http.Request) (map[string]json.RawMessage, error) {
	var payload map[string]json.RawMessage
	if err := decodeJSON(r, &payload); err != nil {
		return nil, err
	}
	if payload == nil {
		payload = map[string]json.RawMessage{}
	}
	return payload, nil
}
