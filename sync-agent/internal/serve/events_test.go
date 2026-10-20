package serve_test

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"roblox-sync/sync-agent/internal/cli"
	"roblox-sync/sync-agent/internal/mapfs"
	"roblox-sync/sync-agent/internal/serve"
)

func mkdirServerRoot(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "src", "server"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func newWatchedServer(t *testing.T, dir string) (*httptest.Server, *serve.Server) {
	t.Helper()
	writeExampleProject(t, dir)
	mkdirServerRoot(t, dir)
	fsys, err := mapfs.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	s := serve.New(fsys, testToken, cli.AgentVersion, cli.ProtocolVersion)
	s.HeartbeatEvery = 50 * time.Millisecond
	if err := s.StartWatch(); err != nil {
		t.Fatalf("StartWatch: %v", err)
	}
	t.Cleanup(func() { s.Close("shutdown") })
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	return ts, s
}

func pollChanges(t *testing.T, base, since string) (status int, cursor int, changes []map[string]any, header http.Header, body []byte) {
	t.Helper()
	path := "/v1/changes"
	if since != "omit" {
		path += "?since=" + since + "&accessToken=" + testToken
	} else {
		path += "?accessToken=" + testToken
	}
	status, body, header = doJSON(t, http.MethodGet, base+path, "", nil)
	if status != http.StatusOK {
		return status, 0, nil, header, body
	}
	var got struct {
		Cursor  int              `json:"cursor"`
		Changes []map[string]any `json:"changes"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode changes: %v body=%s", err, body)
	}
	return status, got.Cursor, got.Changes, header, body
}

func waitChanges(t *testing.T, base string, min int) []map[string]any {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	var last []map[string]any
	for time.Now().Before(deadline) {
		status, _, changes, _, body := pollChanges(t, base, "omit")
		if status != http.StatusOK {
			t.Fatalf("poll status %d body=%s", status, body)
		}
		last = changes
		if len(changes) >= min {
			return changes
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %d changes, last=%v", min, last)
	return nil
}

func writeServerScript(t *testing.T, dir, name, source string) string {
	t.Helper()
	path := filepath.Join(dir, "src", "server", name)
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func openSSE(t *testing.T, url string) (*http.Response, *bufio.Reader) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp, bufio.NewReader(resp.Body)
}

func readSSEEvent(t *testing.T, r *bufio.Reader, d time.Duration) (name, id, data string, ok bool) {
	t.Helper()
	type result struct {
		name, id, data string
		err            error
	}
	ch := make(chan result, 1)
	go func() {
		var evName, evID, evData string
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				ch <- result{err: err}
				return
			}
			line = strings.TrimRight(line, "\r\n")
			if line == "" {
				if evName != "" || evData != "" {
					ch <- result{name: evName, id: evID, data: evData}
					return
				}
				continue
			}
			switch {
			case strings.HasPrefix(line, "event:"):
				evName = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			case strings.HasPrefix(line, "id:"):
				evID = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
			case strings.HasPrefix(line, "data:"):
				evData = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			}
		}
	}()
	select {
	case res := <-ch:
		if res.err != nil {
			if res.err != io.EOF && res.err != io.ErrUnexpectedEOF {
				t.Logf("sse read: %v", res.err)
			}
			return "", "", "", false
		}
		return res.name, res.id, res.data, true
	case <-time.After(d):
		return "", "", "", false
	}
}

// Spec-Anchor: proj-spec/chapters/07.2-event-stream.md#get-v1events-preferred
func TestServeEventsContentType_WhenAuthorized(t *testing.T) {
	dir := t.TempDir()
	ts, _ := newWatchedServer(t, dir)
	resp, _ := openSSE(t, ts.URL+authed("/v1/events"))
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status %d body=%s", resp.StatusCode, body)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("Content-Type %q, want text/event-stream", ct)
	}
}

// Spec-Anchor: proj-spec/chapters/07.2-event-stream.md#error-handling
// Spec-Anchor: proj-spec/chapters/06.1-sync-protocol.md#auth
func TestServeEventsUnauthorized_WhenTokenMissing(t *testing.T) {
	dir := t.TempDir()
	ts, _ := newWatchedServer(t, dir)
	status, body, header := doJSON(t, http.MethodGet, ts.URL+"/v1/events", "", nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("status %d, want 401 body=%s", status, body)
	}
	if errorCode(t, body) != "UNAUTHORIZED" {
		t.Fatalf("code %s, want UNAUTHORIZED body=%s", errorCode(t, body), body)
	}
	if strings.Contains(header.Get("Content-Type"), "text/event-stream") {
		t.Fatalf("unauthorized stream must not send SSE body, Content-Type=%q body=%s", header.Get("Content-Type"), body)
	}
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#get-v1changes
// Spec-Anchor: proj-spec/chapters/07.2-event-stream.md#get-v1changes-fallback
// Spec-Anchor: proj-spec/chapters/13-performance.md#limits
func TestServeChangesReturnsImmediately_WhenNoEvents(t *testing.T) {
	dir := t.TempDir()
	ts, _ := newWatchedServer(t, dir)
	start := time.Now()
	status, cursor, changes, header, body := pollChanges(t, ts.URL, "omit")
	if status != http.StatusOK {
		t.Fatalf("status %d body=%s", status, body)
	}
	if time.Since(start) > 200*time.Millisecond {
		t.Fatalf("poll blocked for %s", time.Since(start))
	}
	if cursor != 0 {
		t.Fatalf("cursor %d, want 0", cursor)
	}
	if len(changes) != 0 {
		t.Fatalf("changes %v, want empty", changes)
	}
	if !strings.Contains(header.Get("Content-Type"), "application/json") {
		t.Fatalf("Content-Type %q", header.Get("Content-Type"))
	}
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#watch-to-roblox-studio
// Spec-Anchor: proj-spec/chapters/07.2-event-stream.md#event-object
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#disk--roblox-studio-g2
func TestServeChangesOriginDisk_WhenFileWritten(t *testing.T) {
	dir := t.TempDir()
	ts, _ := newWatchedServer(t, dir)
	writeServerScript(t, dir, "Foo.server.luau", "print(1)\n")
	changes := waitChanges(t, ts.URL, 1)
	got := changes[0]
	if got["origin"] != "disk" {
		t.Fatalf("origin %v, want disk", got["origin"])
	}
	if got["op"] != "upsert" {
		t.Fatalf("op %v, want upsert", got["op"])
	}
	if got["nodeId"] != "server/Foo" {
		t.Fatalf("nodeId %v, want server/Foo", got["nodeId"])
	}
	if got["className"] != "Script" {
		t.Fatalf("className %v", got["className"])
	}
	if got["source"] != "print(1)\n" {
		t.Fatalf("source %v", got["source"])
	}
	cur, _ := got["cursor"].(float64)
	if cur < 1 {
		t.Fatalf("cursor %v, want >= 1", got["cursor"])
	}
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#watch-to-roblox-studio
func TestServeChangesCursorIncreases_WhenTwoFilesWritten(t *testing.T) {
	dir := t.TempDir()
	ts, _ := newWatchedServer(t, dir)
	writeServerScript(t, dir, "Foo.server.luau", "print(1)\n")
	writeServerScript(t, dir, "Bar.server.luau", "print(2)\n")
	changes := waitChanges(t, ts.URL, 2)
	c1, _ := changes[0]["cursor"].(float64)
	c2, _ := changes[1]["cursor"].(float64)
	if c1 != 1 || c2 != 2 {
		t.Fatalf("cursors %v %v, want 1 then 2", changes[0]["cursor"], changes[1]["cursor"])
	}
}

// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#get-v1changes
// Spec-Anchor: proj-spec/chapters/07.2-event-stream.md#get-v1changes-fallback
func TestServeChangesReturnsBacklog_WhenSinceDefaultZero(t *testing.T) {
	dir := t.TempDir()
	ts, _ := newWatchedServer(t, dir)
	writeServerScript(t, dir, "Foo.server.luau", "print(1)\n")
	waitChanges(t, ts.URL, 1)
	status, cursor, changes, _, body := pollChanges(t, ts.URL, "0")
	if status != http.StatusOK {
		t.Fatalf("status %d body=%s", status, body)
	}
	if cursor != 1 || len(changes) != 1 {
		t.Fatalf("cursor=%d changes=%d body=%s", cursor, len(changes), body)
	}
}

// Spec-Anchor: proj-spec/chapters/07.2-event-stream.md#get-v1events-preferred
func TestServeEventsOmitsBacklog_WhenSinceOmitted(t *testing.T) {
	dir := t.TempDir()
	ts, _ := newWatchedServer(t, dir)
	writeServerScript(t, dir, "Foo.server.luau", "print(1)\n")
	waitChanges(t, ts.URL, 1)
	resp, r := openSSE(t, ts.URL+authed("/v1/events"))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		name, _, data, ok := readSSEEvent(t, r, time.Until(deadline)+10*time.Millisecond)
		if !ok {
			break
		}
		if name == "change" {
			t.Fatalf("SSE without since must not replay backlog, data=%s", data)
		}
	}
	writeServerScript(t, dir, "Bar.server.luau", "print(2)\n")
	found := false
	end := time.Now().Add(2 * time.Second)
	for time.Now().Before(end) {
		name, id, data, ok := readSSEEvent(t, r, time.Until(end))
		if !ok {
			break
		}
		if name != "change" {
			continue
		}
		var ev map[string]any
		if err := json.Unmarshal([]byte(data), &ev); err != nil {
			t.Fatalf("data %s: %v", data, err)
		}
		if ev["nodeId"] != "server/Bar" {
			t.Fatalf("live event nodeId %v, want server/Bar", ev["nodeId"])
		}
		if id != "2" {
			t.Fatalf("id %q, want 2", id)
		}
		found = true
		break
	}
	if !found {
		t.Fatal("expected live change after SSE subscribe")
	}
}

// Spec-Anchor: proj-spec/chapters/07.2-event-stream.md#get-v1events-preferred
func TestServeEventsReplaysBacklog_WhenSinceProvided(t *testing.T) {
	dir := t.TempDir()
	ts, _ := newWatchedServer(t, dir)
	writeServerScript(t, dir, "Foo.server.luau", "print(1)\n")
	waitChanges(t, ts.URL, 1)
	resp, r := openSSE(t, ts.URL+authed("/v1/events")+"&since=0")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	end := time.Now().Add(2 * time.Second)
	for time.Now().Before(end) {
		name, id, data, ok := readSSEEvent(t, r, time.Until(end))
		if !ok {
			break
		}
		if name != "change" {
			continue
		}
		if id != "1" {
			t.Fatalf("id %q, want 1", id)
		}
		var ev map[string]any
		if err := json.Unmarshal([]byte(data), &ev); err != nil {
			t.Fatal(err)
		}
		if ev["nodeId"] != "server/Foo" {
			t.Fatalf("nodeId %v", ev["nodeId"])
		}
		return
	}
	t.Fatal("expected backlog change for since=0")
}

// Spec-Anchor: proj-spec/chapters/07.2-event-stream.md#get-v1events-preferred
// Spec-Anchor: proj-spec/chapters/13-performance.md#limits
func TestServeEventsSendsHeartbeat_WhenStreamIdle(t *testing.T) {
	dir := t.TempDir()
	ts, _ := newWatchedServer(t, dir)
	resp, r := openSSE(t, ts.URL+authed("/v1/events"))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	name, _, data, ok := readSSEEvent(t, r, time.Second)
	if !ok {
		t.Fatal("expected heartbeat")
	}
	if name != "heartbeat" {
		t.Fatalf("event %q, want heartbeat", name)
	}
	if data != "{}" {
		t.Fatalf("data %q, want {}", data)
	}
}

// Spec-Anchor: proj-spec/chapters/13-performance.md#limits
func TestServeEventsHeartbeatIntervalIsFifteenSeconds_WhenDefault(t *testing.T) {
	if serve.DefaultHeartbeat != 15*time.Second {
		t.Fatalf("DefaultHeartbeat %s, want 15s", serve.DefaultHeartbeat)
	}
}

// Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#session-lifecycle
// Spec-Anchor: proj-spec/chapters/07.2-event-stream.md#get-v1events-preferred
func TestServeEventsSendsEndReplaced_WhenNewSession(t *testing.T) {
	dir := t.TempDir()
	ts, _ := newWatchedServer(t, dir)
	resp, r := openSSE(t, ts.URL+authed("/v1/events"))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	status, body, _ := doJSON(t, http.MethodPost, ts.URL+authed("/v1/session"),
		`{"protocolVersion":1,"pluginVersion":"0.0.1"}`,
		map[string]string{"Content-Type": "application/json"})
	if status != http.StatusOK {
		t.Fatalf("session status %d body=%s", status, body)
	}
	end := time.Now().Add(2 * time.Second)
	for time.Now().Before(end) {
		name, _, data, ok := readSSEEvent(t, r, time.Until(end))
		if !ok {
			break
		}
		if name != "end" {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			t.Fatal(err)
		}
		if payload["reason"] != "replaced" {
			t.Fatalf("reason %v, want replaced", payload["reason"])
		}
		return
	}
	t.Fatal("expected event:end replaced")
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#commands
// Spec-Anchor: proj-spec/chapters/07.2-event-stream.md#get-v1events-preferred
func TestServeEventsSendsEndShutdown_WhenServerCloses(t *testing.T) {
	dir := t.TempDir()
	writeExampleProject(t, dir)
	mkdirServerRoot(t, dir)
	fsys, err := mapfs.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	s := serve.New(fsys, testToken, cli.AgentVersion, cli.ProtocolVersion)
	s.HeartbeatEvery = time.Hour
	if err := s.StartWatch(); err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	resp, r := openSSE(t, ts.URL+authed("/v1/events"))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	time.Sleep(20 * time.Millisecond)
	s.Close("shutdown")
	end := time.Now().Add(2 * time.Second)
	for time.Now().Before(end) {
		name, _, data, ok := readSSEEvent(t, r, time.Until(end))
		if !ok {
			break
		}
		if name != "end" {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			t.Fatal(err)
		}
		if payload["reason"] != "shutdown" {
			t.Fatalf("reason %v, want shutdown", payload["reason"])
		}
		return
	}
	t.Fatal("expected event:end shutdown")
}

// Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#echo-suppression
// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#watch-to-roblox-studio
func TestServeBootstrapDoesNotEmitDiskEvent_WhenStudioToDisk(t *testing.T) {
	dir := t.TempDir()
	ts, _ := newWatchedServer(t, dir)
	status, body, _ := doJSON(t, http.MethodPost, ts.URL+authed("/v1/bootstrap/apply"),
		`{"direction":"studio-to-disk","nodes":[{"nodeId":"server/Foo","className":"Script","source":"print(1)\n"}]}`,
		map[string]string{"Content-Type": "application/json"})
	if status != http.StatusOK {
		t.Fatalf("bootstrap status %d body=%s", status, body)
	}
	time.Sleep(150 * time.Millisecond)
	st, _, changes, _, raw := pollChanges(t, ts.URL, "omit")
	if st != http.StatusOK {
		t.Fatalf("poll status %d body=%s", st, raw)
	}
	if len(changes) != 0 {
		t.Fatalf("bootstrap self-write must not emit disk events: %v", changes)
	}
}

// Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#echo-suppression
// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#watch-to-roblox-studio
func TestServeApplyDoesNotEmitDiskEvent_WhenSelfWrite(t *testing.T) {
	dir := t.TempDir()
	ts, _ := newWatchedServer(t, dir)
	status, body, _ := doJSON(t, http.MethodPost, ts.URL+authed("/v1/apply"),
		applyBody(`{"op":"upsert","nodeId":"server/Foo","className":"Script","source":"print(1)\n","baseRevision":0}`),
		map[string]string{"Content-Type": "application/json"})
	if status != http.StatusOK {
		t.Fatalf("apply status %d body=%s", status, body)
	}
	time.Sleep(150 * time.Millisecond)
	st, _, changes, _, raw := pollChanges(t, ts.URL, "omit")
	if st != http.StatusOK {
		t.Fatalf("poll status %d body=%s", st, raw)
	}
	if len(changes) != 0 {
		t.Fatalf("self-write must not emit disk events: %v", changes)
	}
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#construction--configuration
// Spec-Anchor: proj-spec/chapters/13-performance.md#limits
func TestServeChangesCoalescesWrites_WhenTwoWritesWithinDebounce(t *testing.T) {
	dir := t.TempDir()
	ts, _ := newWatchedServer(t, dir)
	path := filepath.Join(dir, "src", "server", "Foo.server.luau")
	if err := os.WriteFile(path, []byte("print(1)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("print(2)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	changes := waitChanges(t, ts.URL, 1)
	if len(changes) != 1 {
		t.Fatalf("want 1 coalesced change, got %d: %v", len(changes), changes)
	}
	if changes[0]["source"] != "print(2)\n" {
		t.Fatalf("source %v, want last write", changes[0]["source"])
	}
	time.Sleep(100 * time.Millisecond)
	_, _, again, _, _ := pollChanges(t, ts.URL, "omit")
	if len(again) != 1 {
		t.Fatalf("extra events after coalesce: %v", again)
	}
}
