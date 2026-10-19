package serve_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func applyBody(changes string) string {
	return `{"changes":[` + changes + `]}`
}

func appliedRevision(t *testing.T, body []byte, nodeID string) int {
	t.Helper()
	var got struct {
		Applied []struct {
			NodeID   string `json:"nodeId"`
			Revision int    `json:"revision"`
		} `json:"applied"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode applied: %v body=%s", err, body)
	}
	for _, e := range got.Applied {
		if e.NodeID == nodeID {
			return e.Revision
		}
	}
	t.Fatalf("missing applied %s: %s", nodeID, body)
	return 0
}

func snapshotRevision(t *testing.T, srvURL, nodeID string) int {
	t.Helper()
	status, body, _ := doJSON(t, http.MethodGet, srvURL+authed("/v1/snapshot"), "", nil)
	if status != http.StatusOK {
		t.Fatalf("snapshot status %d body=%s", status, body)
	}
	var got struct {
		Nodes []struct {
			NodeID   string `json:"nodeId"`
			Revision int    `json:"revision"`
		} `json:"nodes"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	for _, n := range got.Nodes {
		if n.NodeID == nodeID {
			return n.Revision
		}
	}
	t.Fatalf("missing snapshot node %s: %s", nodeID, body)
	return 0
}

func conflictCurrent(t *testing.T, body []byte, nodeID string) int {
	t.Helper()
	var got struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
		Applied []struct {
			NodeID   string `json:"nodeId"`
			Revision int    `json:"revision"`
		} `json:"applied"`
		Conflicts []struct {
			NodeID          string `json:"nodeId"`
			CurrentRevision int    `json:"currentRevision"`
		} `json:"conflicts"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode conflict: %v body=%s", err, body)
	}
	if got.Error.Code != "CONFLICT" {
		t.Fatalf("code %s, want CONFLICT body=%s", got.Error.Code, body)
	}
	for _, c := range got.Conflicts {
		if c.NodeID == nodeID {
			return c.CurrentRevision
		}
	}
	t.Fatalf("missing conflict %s: %s", nodeID, body)
	return 0
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1apply
// Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#live-apply
// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#apply-from-roblox-studio
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#roblox-studio--disk-g1
func TestServeApplyUpsertsSource_WhenBaseRevisionMatches(t *testing.T) {
	dir := t.TempDir()
	seedMappedTree(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body := postJSON(t, srv.URL, "/v1/apply", applyBody(
		`{"op":"upsert","nodeId":"server/Foo","className":"ModuleScript","source":"print('live')\n","baseRevision":1}`,
	))
	if status != http.StatusOK {
		t.Fatalf("status %d, want 200 body=%s", status, body)
	}
	if appliedRevision(t, body, "server/Foo") != 2 {
		t.Fatalf("applied revision want 2 body=%s", body)
	}
	got, err := os.ReadFile(filepath.Join(dir, "src", "server", "Foo.luau"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "print('live')\n" {
		t.Fatalf("disk source %q", got)
	}
	if snapshotRevision(t, srv.URL, "server/Foo") != 2 {
		t.Fatalf("snapshot revision want 2")
	}
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1apply
// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#apply-from-roblox-studio
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#roblox-studio--disk-g1
func TestServeApplyDeletesFile_WhenBaseRevisionMatches(t *testing.T) {
	dir := t.TempDir()
	seedMappedTree(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body := postJSON(t, srv.URL, "/v1/apply", applyBody(
		`{"op":"delete","nodeId":"server/Bar","baseRevision":1}`,
	))
	if status != http.StatusOK {
		t.Fatalf("status %d, want 200 body=%s", status, body)
	}
	if appliedRevision(t, body, "server/Bar") != 2 {
		t.Fatalf("applied revision want 2 body=%s", body)
	}
	if _, err := os.Stat(filepath.Join(dir, "src", "server", "Bar.luau")); !os.IsNotExist(err) {
		t.Fatalf("Bar.luau still present: %v", err)
	}
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1apply
// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#apply-from-roblox-studio
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#roblox-studio--disk-g1
func TestServeApplyRenameAcksNewNodeId_WhenBaseRevisionMatches(t *testing.T) {
	dir := t.TempDir()
	seedMappedTree(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body := postJSON(t, srv.URL, "/v1/apply", applyBody(
		`{"op":"rename","nodeId":"server/Foo","newNodeId":"server/Renamed","baseRevision":1}`,
	))
	if status != http.StatusOK {
		t.Fatalf("status %d, want 200 body=%s", status, body)
	}
	if appliedRevision(t, body, "server/Renamed") != 2 {
		t.Fatalf("rename ack must use newNodeId revision 2 body=%s", body)
	}
	var got struct {
		Applied []struct {
			NodeID string `json:"nodeId"`
		} `json:"applied"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	for _, e := range got.Applied {
		if e.NodeID == "server/Foo" {
			t.Fatalf("rename ack must not use old nodeId: %s", body)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "src", "server", "Foo.luau")); !os.IsNotExist(err) {
		t.Fatalf("old file still present: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "src", "server", "Renamed.luau")); err != nil {
		t.Fatalf("new file missing: %v", err)
	}
}

// Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#live-apply
// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1apply
func TestServeApplyCreatesNode_WhenBaseRevisionIsZeroAndAbsent(t *testing.T) {
	dir := t.TempDir()
	seedMappedTree(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body := postJSON(t, srv.URL, "/v1/apply", applyBody(
		`{"op":"upsert","nodeId":"server/New","className":"ModuleScript","source":"print('new')\n","baseRevision":0}`,
	))
	if status != http.StatusOK {
		t.Fatalf("status %d, want 200 body=%s", status, body)
	}
	if appliedRevision(t, body, "server/New") != 1 {
		t.Fatalf("new node revision want 1 body=%s", body)
	}
	got, err := os.ReadFile(filepath.Join(dir, "src", "server", "New.luau"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "print('new')\n" {
		t.Fatalf("disk source %q", got)
	}
}

// Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#live-apply
// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1apply
// Spec-Anchor: proj-spec/chapters/06.1-sync-protocol.md#error-codes
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
func TestServeApplyReturnsConflict_WhenBaseRevisionStale(t *testing.T) {
	dir := t.TempDir()
	seedMappedTree(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body := postJSON(t, srv.URL, "/v1/apply", applyBody(
		`{"op":"upsert","nodeId":"server/Foo","className":"ModuleScript","source":"print('stale')\n","baseRevision":0}`,
	))
	if status != http.StatusConflict {
		t.Fatalf("status %d, want 409 body=%s", status, body)
	}
	if conflictCurrent(t, body, "server/Foo") != 1 {
		t.Fatalf("currentRevision want 1 body=%s", body)
	}
	got, err := os.ReadFile(filepath.Join(dir, "src", "server", "Foo.luau"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "print('hi')\n" {
		t.Fatalf("conflict must not write stale node: %q", got)
	}
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1apply
// Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#live-apply
func TestServeApplyAppliesOtherEntries_WhenOneNodeConflicts(t *testing.T) {
	dir := t.TempDir()
	seedMappedTree(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body := postJSON(t, srv.URL, "/v1/apply", applyBody(
		`{"op":"upsert","nodeId":"server/Foo","className":"ModuleScript","source":"print('stale')\n","baseRevision":0},`+
			`{"op":"upsert","nodeId":"server/Bar","className":"ModuleScript","source":"print('ok')\n","baseRevision":1}`,
	))
	if status != http.StatusConflict {
		t.Fatalf("status %d, want 409 body=%s", status, body)
	}
	if conflictCurrent(t, body, "server/Foo") != 1 {
		t.Fatalf("Foo conflict currentRevision want 1 body=%s", body)
	}
	if appliedRevision(t, body, "server/Bar") != 2 {
		t.Fatalf("Bar should apply body=%s", body)
	}
	foo, err := os.ReadFile(filepath.Join(dir, "src", "server", "Foo.luau"))
	if err != nil {
		t.Fatal(err)
	}
	if string(foo) != "print('hi')\n" {
		t.Fatalf("Foo must stay original: %q", foo)
	}
	bar, err := os.ReadFile(filepath.Join(dir, "src", "server", "Bar.luau"))
	if err != nil {
		t.Fatal(err)
	}
	if string(bar) != "print('ok')\n" {
		t.Fatalf("Bar disk %q", bar)
	}
}

// Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#conflict-ui
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
func TestServeApplyRejectsForce_WhenFirstTryLive(t *testing.T) {
	dir := t.TempDir()
	seedMappedTree(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body := postJSON(t, srv.URL, "/v1/apply", applyBody(
		`{"op":"upsert","nodeId":"server/Foo","className":"ModuleScript","source":"print('force')\n","baseRevision":1,"force":true}`,
	))
	if status != http.StatusBadRequest {
		t.Fatalf("status %d, want 400 body=%s", status, body)
	}
	if errorCode(t, body) != "BAD_REQUEST" {
		t.Fatalf("code %s, want BAD_REQUEST body=%s", errorCode(t, body), body)
	}
	got, err := os.ReadFile(filepath.Join(dir, "src", "server", "Foo.luau"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "print('hi')\n" {
		t.Fatalf("first-try force must not write: %q", got)
	}
	status, body = postJSON(t, srv.URL, "/v1/apply", applyBody(
		`{"op":"upsert","nodeId":"server/Foo","className":"ModuleScript","source":"print('live')\n","baseRevision":1}`,
	))
	if status != http.StatusOK {
		t.Fatalf("matching apply after rejected force status %d body=%s", status, body)
	}
}

// Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#conflict-ui
// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1apply
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
func TestServeApplyForceWrites_WhenAfterConflict(t *testing.T) {
	dir := t.TempDir()
	seedMappedTree(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body := postJSON(t, srv.URL, "/v1/apply", applyBody(
		`{"op":"upsert","nodeId":"server/Foo","className":"ModuleScript","source":"print('stale')\n","baseRevision":0}`,
	))
	if status != http.StatusConflict {
		t.Fatalf("setup conflict status %d body=%s", status, body)
	}
	cur := conflictCurrent(t, body, "server/Foo")
	force, err := json.Marshal(map[string]any{
		"changes": []map[string]any{{
			"op":           "upsert",
			"nodeId":       "server/Foo",
			"className":    "ModuleScript",
			"source":       "print('keep studio')\n",
			"baseRevision": cur,
			"force":        true,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	status, body = postJSON(t, srv.URL, "/v1/apply", string(force))
	if status != http.StatusOK {
		t.Fatalf("force status %d, want 200 body=%s", status, body)
	}
	if appliedRevision(t, body, "server/Foo") != cur+1 {
		t.Fatalf("force revision want %d body=%s", cur+1, body)
	}
	got, err := os.ReadFile(filepath.Join(dir, "src", "server", "Foo.luau"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "print('keep studio')\n" {
		t.Fatalf("force disk %q", got)
	}
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1apply
func TestServeApplyBadRequest_WhenFolderHasSource(t *testing.T) {
	dir := t.TempDir()
	seedMappedTree(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body := postJSON(t, srv.URL, "/v1/apply", applyBody(
		`{"op":"upsert","nodeId":"server/Lib","className":"Folder","source":"","baseRevision":1}`,
	))
	if status != http.StatusBadRequest {
		t.Fatalf("status %d, want 400 body=%s", status, body)
	}
	if errorCode(t, body) != "BAD_REQUEST" {
		t.Fatalf("code %s, want BAD_REQUEST body=%s", errorCode(t, body), body)
	}
	status, body = postJSON(t, srv.URL, "/v1/apply", applyBody(
		`{"op":"upsert","nodeId":"server/ExtraFolder","className":"Folder","baseRevision":0}`,
	))
	if status != http.StatusOK {
		t.Fatalf("Folder without source status %d, want 200 body=%s", status, body)
	}
	info, err := os.Stat(filepath.Join(dir, "src", "server", "ExtraFolder"))
	if err != nil || !info.IsDir() {
		t.Fatalf("Folder upsert should create directory: %v", err)
	}
}

// Spec-Anchor: proj-spec/chapters/13-performance.md#limits
// Spec-Anchor: proj-spec/chapters/06.1-sync-protocol.md#error-codes
// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1apply
func TestServeApplyPayloadTooLarge_WhenSourceExceedsOneMiB(t *testing.T) {
	dir := t.TempDir()
	seedMappedTree(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	huge := strings.Repeat("a", (1<<20)+1)
	payload, err := json.Marshal(map[string]any{
		"changes": []map[string]any{{
			"op":           "upsert",
			"nodeId":       "server/Foo",
			"className":    "ModuleScript",
			"source":       huge,
			"baseRevision": 1,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	status, body := postJSON(t, srv.URL, "/v1/apply", string(payload))
	if status != http.StatusRequestEntityTooLarge {
		t.Fatalf("status %d, want 413 body=%s", status, body)
	}
	if errorCode(t, body) != "PAYLOAD_TOO_LARGE" {
		t.Fatalf("code %s, want PAYLOAD_TOO_LARGE body=%s", errorCode(t, body), body)
	}
	got, err := os.ReadFile(filepath.Join(dir, "src", "server", "Foo.luau"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "print('hi')\n" {
		t.Fatalf("oversize must not write: %q", got)
	}
}

// Spec-Anchor: proj-spec/chapters/06.1-sync-protocol.md#error-codes
// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1apply
func TestServeApplyNotMapped_WhenNodeIdOutsideRoots(t *testing.T) {
	dir := t.TempDir()
	seedMappedTree(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body := postJSON(t, srv.URL, "/v1/apply", applyBody(
		`{"op":"upsert","nodeId":"other/Foo","className":"ModuleScript","source":"print('x')\n","baseRevision":0}`,
	))
	if status != http.StatusBadRequest {
		t.Fatalf("status %d, want 400 body=%s", status, body)
	}
	if errorCode(t, body) != "NOT_MAPPED" {
		t.Fatalf("code %s, want NOT_MAPPED body=%s", errorCode(t, body), body)
	}
}
