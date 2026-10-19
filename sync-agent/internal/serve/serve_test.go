package serve_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"roblox-sync/sync-agent/internal/cli"
	"roblox-sync/sync-agent/internal/mapfs"
	"roblox-sync/sync-agent/internal/serve"
)

const testToken = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(file)
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "proj-spec", "chapters")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "sync-agent")); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("could not locate repository root")
	return ""
}

func writeExampleProject(t *testing.T, dir string) {
	t.Helper()
	path := filepath.Join(repoRoot(t), "sync-protocol", "fixtures", "v1", "mapping", "valid-project-example.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	project, err := json.MarshalIndent(doc["project"], "", "  ")
	if err != nil {
		t.Fatalf("marshal project: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rbx-sync.json"), project, 0o644); err != nil {
		t.Fatalf("write rbx-sync.json: %v", err)
	}
}

func newTestServer(t *testing.T, dir string) *httptest.Server {
	t.Helper()
	fsys, err := mapfs.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	h := serve.New(fsys, testToken, cli.AgentVersion, cli.ProtocolVersion).Handler()
	return httptest.NewServer(h)
}

func doJSON(t *testing.T, method, url, body string, headers map[string]string) (status int, raw []byte, header http.Header) {
	t.Helper()
	var rdr io.Reader
	if body != "" || method == http.MethodPost {
		rdr = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, rdr)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, raw, resp.Header
}

func errorCode(t *testing.T, body []byte) string {
	t.Helper()
	var env struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode envelope: %v body=%s", err, body)
	}
	return env.Error.Code
}

func authed(path string) string {
	return path + "?accessToken=" + testToken
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#routes
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
func TestServeHealthOmitsTokenAndProject_WhenUnauthenticated(t *testing.T) {
	dir := t.TempDir()
	writeExampleProject(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body, header := doJSON(t, http.MethodGet, srv.URL+"/v1/health", "", nil)
	if status != http.StatusOK {
		t.Fatalf("status %d, want 200 body=%s", status, body)
	}
	if !strings.Contains(header.Get("Content-Type"), "application/json") {
		t.Fatalf("Content-Type %q, want application/json", header.Get("Content-Type"))
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got["ok"] != true {
		t.Fatalf("ok=%v, want true", got["ok"])
	}
	if got["protocolVersion"] != float64(1) {
		t.Fatalf("protocolVersion=%v, want 1", got["protocolVersion"])
	}
	text := string(body)
	if strings.Contains(text, testToken) || strings.Contains(text, "accessToken") {
		t.Fatalf("health leaked token: %s", text)
	}
	if strings.Contains(text, dir) || strings.Contains(strings.ToLower(text), "rbx-sync.json") {
		t.Fatalf("health leaked project path: %s", text)
	}
}

// Spec-Anchor: proj-spec/chapters/06.1-sync-protocol.md#auth
// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#get-v1snapshot
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
func TestServeSnapshotUnauthorized_WhenTokenMissing(t *testing.T) {
	dir := t.TempDir()
	writeExampleProject(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body, _ := doJSON(t, http.MethodGet, srv.URL+"/v1/snapshot", "", nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("status %d, want 401 body=%s", status, body)
	}
	if errorCode(t, body) != "UNAUTHORIZED" {
		t.Fatalf("code %s, want UNAUTHORIZED body=%s", errorCode(t, body), body)
	}
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1session
func TestServeSessionMismatch_WhenProtocolDiffers(t *testing.T) {
	dir := t.TempDir()
	writeExampleProject(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body, _ := doJSON(t, http.MethodPost, srv.URL+authed("/v1/session"),
		`{"protocolVersion":2,"pluginVersion":"0.0.1"}`,
		map[string]string{"Content-Type": "application/json"})
	if status != http.StatusConflict {
		t.Fatalf("status %d, want 409 body=%s", status, body)
	}
	if errorCode(t, body) != "PROTOCOL_MISMATCH" {
		t.Fatalf("code %s, want PROTOCOL_MISMATCH body=%s", errorCode(t, body), body)
	}
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1session
// Spec-Anchor: proj-spec/chapters/08-domain-model.md#root-handshake
func TestServeSessionOmitsDiskPaths_WhenProtocolMatches(t *testing.T) {
	dir := t.TempDir()
	writeExampleProject(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body, header := doJSON(t, http.MethodPost, srv.URL+authed("/v1/session"),
		`{"protocolVersion":1,"pluginVersion":"0.0.1","extra":"ignored"}`,
		map[string]string{"Content-Type": "application/json"})
	if status != http.StatusOK {
		t.Fatalf("status %d, want 200 body=%s", status, body)
	}
	if !strings.Contains(header.Get("Content-Type"), "application/json") {
		t.Fatalf("Content-Type %q", header.Get("Content-Type"))
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	roots, _ := got["roots"].([]any)
	if len(roots) == 0 {
		t.Fatalf("roots missing: %s", body)
	}
	for _, rawRoot := range roots {
		root, _ := rawRoot.(map[string]any)
		if _, ok := root["disk"]; ok {
			t.Fatalf("roots must omit disk paths: %s", body)
		}
		if root["id"] == nil || root["studio"] == nil {
			t.Fatalf("roots need id and studio: %s", body)
		}
	}
	if strings.Contains(string(body), `"disk"`) {
		t.Fatalf("response leaked disk: %s", body)
	}
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#get-v1snapshot
// Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#revision-initialization
func TestServeSnapshotReturnsNodes_WhenAuthorized(t *testing.T) {
	dir := t.TempDir()
	writeExampleProject(t, dir)
	src := filepath.Join(dir, "src", "server")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "Foo.luau"), []byte("print('hi')\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body, _ := doJSON(t, http.MethodGet, srv.URL+authed("/v1/snapshot"), "", nil)
	if status != http.StatusOK {
		t.Fatalf("status %d, want 200 body=%s", status, body)
	}
	var got struct {
		Nodes []map[string]any `json:"nodes"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, n := range got.Nodes {
		if n["nodeId"] == "server/Foo" {
			found = true
			if n["className"] != "ModuleScript" {
				t.Fatalf("className %v", n["className"])
			}
			if n["source"] != "print('hi')\n" {
				t.Fatalf("source %v", n["source"])
			}
			if n["revision"] != float64(1) {
				t.Fatalf("revision %v, want 1 on fresh serve", n["revision"])
			}
		}
	}
	if !found {
		t.Fatalf("missing server/Foo in %s", body)
	}
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#routes
// Spec-Anchor: proj-spec/chapters/06.1-sync-protocol.md#error-codes
func TestServeUnlistedPathNotFound_WhenRouteMissing(t *testing.T) {
	dir := t.TempDir()
	writeExampleProject(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body, _ := doJSON(t, http.MethodGet, srv.URL+authed("/v1/nope"), "", nil)
	if status != http.StatusNotFound {
		t.Fatalf("status %d, want 404 body=%s", status, body)
	}
	if errorCode(t, body) != "NOT_FOUND" {
		t.Fatalf("code %s, want NOT_FOUND body=%s", errorCode(t, body), body)
	}
	if bytes.Contains(body, []byte("BAD_REQUEST")) {
		t.Fatalf("unlisted path must not use BAD_REQUEST: %s", body)
	}
}

// Spec-Anchor: proj-spec/chapters/03-architecture.md#trust-boundary
// Spec-Anchor: proj-spec/chapters/06.1-sync-protocol.md#auth
func TestServeUnlistedPathUnauthorized_WhenTokenMissing(t *testing.T) {
	dir := t.TempDir()
	writeExampleProject(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body, _ := doJSON(t, http.MethodGet, srv.URL+"/v1/nope", "", nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("status %d, want 401 body=%s", status, body)
	}
	if errorCode(t, body) != "UNAUTHORIZED" {
		t.Fatalf("code %s, want UNAUTHORIZED body=%s", errorCode(t, body), body)
	}
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#error-handling
func TestServeSessionBadRequest_WhenJSONInvalid(t *testing.T) {
	dir := t.TempDir()
	writeExampleProject(t, dir)
	srv := newTestServer(t, dir)
	defer srv.Close()
	status, body, _ := doJSON(t, http.MethodPost, srv.URL+authed("/v1/session"),
		`{`, map[string]string{"Content-Type": "application/json"})
	if status != http.StatusBadRequest {
		t.Fatalf("status %d, want 400 body=%s", status, body)
	}
	if errorCode(t, body) != "BAD_REQUEST" {
		t.Fatalf("code %s, want BAD_REQUEST body=%s", errorCode(t, body), body)
	}
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#error-handling
type unreadableTree struct{}

func (unreadableTree) Nodes() ([]mapfs.Node, error) {
	return nil, &mapfs.Error{Code: mapfs.CodeIOError, Message: "permission denied"}
}
func (unreadableTree) ReadSource(string) (string, error) {
	return "", &mapfs.Error{Code: mapfs.CodeIOError, Message: "permission denied"}
}
func (unreadableTree) ProjectName() string { return "example" }
func (unreadableTree) StudioRoots() []mapfs.StudioRoot {
	return []mapfs.StudioRoot{{ID: "server", Studio: "ServerScriptService"}}
}
func (unreadableTree) Upsert(string, string, string) error {
	return &mapfs.Error{Code: mapfs.CodeIOError, Message: "permission denied"}
}
func (unreadableTree) Delete(string) error {
	return &mapfs.Error{Code: mapfs.CodeIOError, Message: "permission denied"}
}
func (unreadableTree) Rename(string, string, string) error {
	return &mapfs.Error{Code: mapfs.CodeIOError, Message: "permission denied"}
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#error-handling
func TestServeSnapshotIOError_WhenRootUnreadable(t *testing.T) {
	h := serve.New(unreadableTree{}, testToken, cli.AgentVersion, cli.ProtocolVersion).Handler()
	srv := httptest.NewServer(h)
	defer srv.Close()
	status, body, _ := doJSON(t, http.MethodGet, srv.URL+authed("/v1/snapshot"), "", nil)
	if status != http.StatusInternalServerError {
		t.Fatalf("snapshot status %d, want 500 body=%s", status, body)
	}
	if errorCode(t, body) != "IO_ERROR" {
		t.Fatalf("snapshot code %s, want IO_ERROR body=%s", errorCode(t, body), body)
	}
	healthStatus, healthBody, _ := doJSON(t, http.MethodGet, srv.URL+"/v1/health", "", nil)
	if healthStatus != http.StatusOK {
		t.Fatalf("health should stay up, status %d body=%s", healthStatus, healthBody)
	}
	applyStatus, applyBody, _ := doJSON(t, http.MethodPost, srv.URL+authed("/v1/apply"),
		`{"changes":[]}`, map[string]string{"Content-Type": "application/json"})
	if applyStatus != http.StatusInternalServerError {
		t.Fatalf("apply status %d, want 500 body=%s", applyStatus, applyBody)
	}
	if errorCode(t, applyBody) != "IO_ERROR" {
		t.Fatalf("apply code %s, want IO_ERROR body=%s", errorCode(t, applyBody), applyBody)
	}
}
