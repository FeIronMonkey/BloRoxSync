package serve_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func seedMappedTree(t *testing.T, dir string) {
	t.Helper()
	writeExampleProject(t, dir)
	src := filepath.Join(dir, "src", "server")
	if err := os.MkdirAll(filepath.Join(src, "Lib"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "Foo.luau"), []byte("print('hi')\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "Bar.luau"), []byte("print('bar')\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# unmapped\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "notes.txt"), []byte("keep me\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func postJSON(t *testing.T, srvURL, path, body string) (status int, raw []byte) {
	t.Helper()
	status, raw, _ = doJSON(t, http.MethodPost, srvURL+authed(path), body,
		map[string]string{"Content-Type": "application/json"})
	return status, raw
}

func matchingStudioSnapshot() string {
	return `{"nodes":[` +
		`{"nodeId":"server/Foo","className":"ModuleScript","source":"print('hi')\n"},` +
		`{"nodeId":"server/Bar","className":"ModuleScript","source":"print('bar')\n"},` +
		`{"nodeId":"server/Lib","className":"Folder"}` +
		`]}`
}

func decodeEqual(t *testing.T, body []byte) bool {
	t.Helper()
	var got struct {
		Equal *bool `json:"equal"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode: %v body=%s", err, body)
	}
	if got.Equal == nil {
		t.Fatalf("missing equal: %s", body)
	}
	return *got.Equal
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1bootstrapcompare
// Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#session-lifecycle
// Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#bootstrap
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
func TestServeCompareEqualTrue_WhenNodeIdClassAndSourceMatch(t *testing.T) {
	dir := t.TempDir()
	seedMappedTree(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body := postJSON(t, srv.URL, "/v1/bootstrap/compare", matchingStudioSnapshot())
	if status != http.StatusOK {
		t.Fatalf("status %d, want 200 body=%s", status, body)
	}
	if !decodeEqual(t, body) {
		t.Fatalf("equal=false, want true body=%s", body)
	}
	if _, err := os.Stat(filepath.Join(dir, "src", "server", "Bar.luau")); err != nil {
		t.Fatalf("compare must not write disk: %v", err)
	}
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1bootstrapcompare
func TestServeCompareEqualTrue_WhenStudioSourceHasCRLF(t *testing.T) {
	dir := t.TempDir()
	seedMappedTree(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	bodyJSON := `{"nodes":[` +
		`{"nodeId":"server/Foo","className":"ModuleScript","source":"print('hi')\r\n"},` +
		`{"nodeId":"server/Bar","className":"ModuleScript","source":"print('bar')\r\n"},` +
		`{"nodeId":"server/Lib","className":"Folder"}` +
		`]}`
	status, body := postJSON(t, srv.URL, "/v1/bootstrap/compare", bodyJSON)
	if status != http.StatusOK {
		t.Fatalf("status %d, want 200 body=%s", status, body)
	}
	if !decodeEqual(t, body) {
		t.Fatalf("LF-normalized sources must compare equal: %s", body)
	}
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1bootstrapcompare
func TestServeCompareEqualFalse_WhenSourceDiffers(t *testing.T) {
	dir := t.TempDir()
	seedMappedTree(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	bodyJSON := `{"nodes":[` +
		`{"nodeId":"server/Foo","className":"ModuleScript","source":"print('other')\n"},` +
		`{"nodeId":"server/Bar","className":"ModuleScript","source":"print('bar')\n"},` +
		`{"nodeId":"server/Lib","className":"Folder"}` +
		`]}`
	status, body := postJSON(t, srv.URL, "/v1/bootstrap/compare", bodyJSON)
	if status != http.StatusOK {
		t.Fatalf("status %d, want 200 body=%s", status, body)
	}
	if decodeEqual(t, body) {
		t.Fatalf("equal=true, want false body=%s", body)
	}
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1bootstrapcompare
func TestServeCompareEqualFalse_WhenClassNameDiffers(t *testing.T) {
	dir := t.TempDir()
	seedMappedTree(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	bodyJSON := `{"nodes":[` +
		`{"nodeId":"server/Foo","className":"Script","source":"print('hi')\n"},` +
		`{"nodeId":"server/Bar","className":"ModuleScript","source":"print('bar')\n"},` +
		`{"nodeId":"server/Lib","className":"Folder"}` +
		`]}`
	status, body := postJSON(t, srv.URL, "/v1/bootstrap/compare", bodyJSON)
	if status != http.StatusOK {
		t.Fatalf("status %d, want 200 body=%s", status, body)
	}
	if decodeEqual(t, body) {
		t.Fatalf("className mismatch must be unequal: %s", body)
	}
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1bootstrapcompare
func TestServeCompareEqualFalse_WhenMappedNodeOmitted(t *testing.T) {
	dir := t.TempDir()
	seedMappedTree(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	bodyJSON := `{"nodes":[` +
		`{"nodeId":"server/Foo","className":"ModuleScript","source":"print('hi')\n"},` +
		`{"nodeId":"server/Lib","className":"Folder"}` +
		`]}`
	status, body := postJSON(t, srv.URL, "/v1/bootstrap/compare", bodyJSON)
	if status != http.StatusOK {
		t.Fatalf("status %d, want 200 body=%s", status, body)
	}
	if decodeEqual(t, body) {
		t.Fatalf("omitted mapped node must be unequal: %s", body)
	}
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1bootstrapcompare
func TestServeCompareBadRequest_WhenNodesOmitted(t *testing.T) {
	dir := t.TempDir()
	seedMappedTree(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body := postJSON(t, srv.URL, "/v1/bootstrap/compare", `{}`)
	if status != http.StatusBadRequest {
		t.Fatalf("status %d, want 400 body=%s", status, body)
	}
	if errorCode(t, body) != "BAD_REQUEST" {
		t.Fatalf("code %s, want BAD_REQUEST body=%s", errorCode(t, body), body)
	}
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1bootstrapapply
// Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#revision-initialization
// Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#bootstrap
func TestServeBootstrapDiskToStudioReturnsRevision1_WhenFreshServe(t *testing.T) {
	dir := t.TempDir()
	seedMappedTree(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	before, err := os.ReadFile(filepath.Join(dir, "src", "server", "Foo.luau"))
	if err != nil {
		t.Fatal(err)
	}
	status, body := postJSON(t, srv.URL, "/v1/bootstrap/apply", `{"direction":"disk-to-studio"}`)
	if status != http.StatusOK {
		t.Fatalf("status %d, want 200 body=%s", status, body)
	}
	var got struct {
		Nodes []map[string]any `json:"nodes"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode: %v body=%s", err, body)
	}
	if _, ok := rawKey(body, "applied"); ok {
		t.Fatalf("disk-to-studio must return nodes, not applied: %s", body)
	}
	found := map[string]map[string]any{}
	for _, n := range got.Nodes {
		id, _ := n["nodeId"].(string)
		found[id] = n
		if n["revision"] != float64(1) {
			t.Fatalf("revision %v for %s, want 1", n["revision"], id)
		}
	}
	foo := found["server/Foo"]
	if foo == nil || foo["className"] != "ModuleScript" || foo["source"] != "print('hi')\n" {
		t.Fatalf("missing disk snapshot node server/Foo: %s", body)
	}
	if found["server/Lib"] == nil || found["server/Lib"]["className"] != "Folder" {
		t.Fatalf("missing Folder server/Lib: %s", body)
	}
	if _, hasSource := found["server/Lib"]["source"]; hasSource {
		t.Fatalf("Folder must omit source: %s", body)
	}
	after, err := os.ReadFile(filepath.Join(dir, "src", "server", "Foo.luau"))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("disk-to-studio must not write disk")
	}
}

func rawKey(body []byte, key string) (json.RawMessage, bool) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(body, &obj); err != nil {
		return nil, false
	}
	v, ok := obj[key]
	return v, ok
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1bootstrapapply
func TestServeBootstrapDiskToStudioBadRequest_WhenNodesPresent(t *testing.T) {
	dir := t.TempDir()
	seedMappedTree(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body := postJSON(t, srv.URL, "/v1/bootstrap/apply",
		`{"direction":"disk-to-studio","nodes":[]}`)
	if status != http.StatusBadRequest {
		t.Fatalf("status %d, want 400 body=%s", status, body)
	}
	if errorCode(t, body) != "BAD_REQUEST" {
		t.Fatalf("code %s, want BAD_REQUEST body=%s", errorCode(t, body), body)
	}
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1bootstrapapply
func TestServeBootstrapStudioToDiskBadRequest_WhenNodesOmitted(t *testing.T) {
	dir := t.TempDir()
	seedMappedTree(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body := postJSON(t, srv.URL, "/v1/bootstrap/apply",
		`{"direction":"studio-to-disk"}`)
	if status != http.StatusBadRequest {
		t.Fatalf("status %d, want 400 body=%s", status, body)
	}
	if errorCode(t, body) != "BAD_REQUEST" {
		t.Fatalf("code %s, want BAD_REQUEST body=%s", errorCode(t, body), body)
	}
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1bootstrapapply
// Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#revision-initialization
// Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#bootstrap
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#out-of-scope-checks-must-remain-true
// Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d3--scripts-and-folders-only-in-v1
func TestServeBootstrapStudioToDiskWritesMappedOnly_WhenNodesSubmitted(t *testing.T) {
	dir := t.TempDir()
	seedMappedTree(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	req := `{"direction":"studio-to-disk","nodes":[` +
		`{"nodeId":"server/Foo","className":"ModuleScript","source":"print('studio')\n"},` +
		`{"nodeId":"server/Lib","className":"Folder"}` +
		`]}`
	status, body := postJSON(t, srv.URL, "/v1/bootstrap/apply", req)
	if status != http.StatusOK {
		t.Fatalf("status %d, want 200 body=%s", status, body)
	}
	var got struct {
		Applied []struct {
			NodeID   string `json:"nodeId"`
			Revision int    `json:"revision"`
		} `json:"applied"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode: %v body=%s", err, body)
	}
	if _, ok := rawKey(body, "nodes"); ok {
		t.Fatalf("studio-to-disk must return applied, not nodes: %s", body)
	}
	applied := map[string]int{}
	for _, e := range got.Applied {
		applied[e.NodeID] = e.Revision
		if e.Revision != 1 {
			t.Fatalf("applied revision %d for %s, want 1", e.Revision, e.NodeID)
		}
	}
	if applied["server/Foo"] != 1 {
		t.Fatalf("missing applied server/Foo: %s", body)
	}
	if applied["server/Lib"] != 1 {
		t.Fatalf("missing applied server/Lib: %s", body)
	}
	if _, ok := applied["server/Bar"]; ok {
		t.Fatalf("omitted mapped node must not remain in applied: %s", body)
	}
	foo, err := os.ReadFile(filepath.Join(dir, "src", "server", "Foo.luau"))
	if err != nil {
		t.Fatal(err)
	}
	if string(foo) != "print('studio')\n" {
		t.Fatalf("Foo source %q", foo)
	}
	if _, err := os.Stat(filepath.Join(dir, "src", "server", "Bar.luau")); !os.IsNotExist(err) {
		t.Fatalf("omitted mapped Bar.luau must be deleted, err=%v", err)
	}
	readme, err := os.ReadFile(filepath.Join(dir, "README.md"))
	if err != nil {
		t.Fatalf("unmapped README.md must stay: %v", err)
	}
	if string(readme) != "# unmapped\n" {
		t.Fatalf("README.md mutated: %q", readme)
	}
	notes, err := os.ReadFile(filepath.Join(dir, "src", "server", "notes.txt"))
	if err != nil {
		t.Fatalf("unmapped notes.txt under mapped root must stay (A20): %v", err)
	}
	if string(notes) != "keep me\n" {
		t.Fatalf("notes.txt mutated: %q", notes)
	}
}
