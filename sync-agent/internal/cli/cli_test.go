package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"roblox-sync/sync-agent/internal/cli"
)

type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (w *lockedBuffer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.b.Write(p)
}

func (w *lockedBuffer) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.b.String()
}

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

func runCLI(t *testing.T, ctx context.Context, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	if ctx == nil {
		ctx = context.Background()
	}
	var outBuf, errBuf lockedBuffer
	code = cli.Run(ctx, args, &outBuf, &errBuf)
	return outBuf.String(), errBuf.String(), code
}

func freePort(t *testing.T, host string) int {
	t.Helper()
	ln, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
	if err != nil {
		t.Fatalf("listen %s: %v", host, err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

func waitServeStarted(t *testing.T, args []string) (stderr string, codeCh <-chan int, cancel context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	var outBuf, errBuf lockedBuffer
	ch := make(chan int, 1)
	go func() {
		ch <- cli.Run(ctx, args, &outBuf, &errBuf)
	}()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case code := <-ch:
			t.Fatalf("serve exited before start: code=%d stdout=%q stderr=%q", code, outBuf.String(), errBuf.String())
		default:
		}
		if strings.Contains(errBuf.String(), "accessToken=") {
			return errBuf.String(), ch, cancel
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	t.Fatalf("serve did not print accessToken=; stdout=%q stderr=%q", outBuf.String(), errBuf.String())
	return "", nil, cancel
}

// Spec-Anchor: proj-spec/chapters/07.3-cli.md#commands
// Spec-Anchor: proj-spec/chapters/03.1-repository-layout.md#versioning
func TestCliVersionPrintsAgentAndProtocol_WhenCommandIsVersion(t *testing.T) {
	stdout, stderr, code := runCLI(t, context.Background(), "version")
	if code != 0 {
		t.Fatalf("exit %d, want 0; stderr=%q", code, stderr)
	}
	line := strings.TrimSpace(stdout)
	re := regexp.MustCompile(`^agent=([0-9]+\.[0-9]+\.[0-9]+) protocol=([0-9]+)$`)
	m := re.FindStringSubmatch(line)
	if m == nil {
		t.Fatalf("stdout %q, want agent=<semver> protocol=<n>", stdout)
	}
	if m[2] != "1" {
		t.Fatalf("protocol=%s, want 1", m[2])
	}
}

// Spec-Anchor: proj-spec/chapters/07.3-cli.md#commands
func TestCliHelpExitsZero_WhenCommandIsHelp(t *testing.T) {
	stdout, stderr, code := runCLI(t, context.Background(), "help")
	if code != 0 {
		t.Fatalf("exit %d, want 0; stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "serve") || !strings.Contains(stdout, "version") || !strings.Contains(stdout, "plugin") {
		t.Fatalf("help stdout %q must describe usage", stdout)
	}
}

// Spec-Anchor: proj-spec/chapters/07.3-cli.md#commands
func TestCliPluginInstallIsKnownCommand_WhenFileMissing(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.rbxmx")
	_, stderr, code := runCLI(t, context.Background(), "plugin", "install", "--file", missing)
	if code == 2 {
		t.Fatalf("plugin install must not be unknown; stderr=%q", stderr)
	}
	if code != 1 {
		t.Fatalf("exit %d, want 1 for missing --file; stderr=%q", code, stderr)
	}
}

// Spec-Anchor: proj-spec/chapters/07.3-cli.md#commands
// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#error-handling
func TestCliUnknownCommandExitsTwo_WhenCommandIsUnknown(t *testing.T) {
	stdout, stderr, code := runCLI(t, context.Background(), "not-a-command")
	if code != 2 {
		t.Fatalf("exit %d, want 2; stdout=%q stderr=%q", code, stdout, stderr)
	}
	if strings.TrimSpace(stderr) == "" {
		t.Fatal("unknown command must write stderr")
	}
}

// Spec-Anchor: proj-spec/chapters/07.3-cli.md#error-handling
// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#construction--configuration
func TestCliServeExitsProjectInvalid_WhenJsonMissing(t *testing.T) {
	dir := t.TempDir()
	_, stderr, code := runCLI(t, context.Background(), "serve", "--project", dir, "--port", "29851")
	if code != 1 {
		t.Fatalf("exit %d, want 1; stderr=%q", code, stderr)
	}
	if !strings.Contains(stderr, "PROJECT_INVALID") {
		t.Fatalf("stderr %q must contain PROJECT_INVALID", stderr)
	}
}

// Spec-Anchor: proj-spec/chapters/07.3-cli.md#serve-flags
// Spec-Anchor: proj-spec/chapters/07.3-cli.md#error-handling
// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#construction--configuration
// Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d5--loopback-only-unprivileged-port-29850
// Spec-Anchor: proj-spec/chapters/10.2-windows-and-wsl.md#binding
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
func TestCliServeRejectsBind_WhenHostIsNotLoopback(t *testing.T) {
	dir := t.TempDir()
	writeExampleProject(t, dir)
	for _, host := range []string{"0.0.0.0", "192.0.2.1"} {
		_, stderr, code := runCLI(t, context.Background(), "serve", "--project", dir, "--bind", host, "--port", "29851")
		if code != 1 {
			t.Errorf("bind %s: exit %d, want 1; stderr=%q", host, code, stderr)
		}
		if !strings.Contains(stderr, "must be loopback") {
			t.Errorf("bind %s: stderr %q must say must be loopback", host, stderr)
		}
	}
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#construction--configuration
// Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d5--loopback-only-unprivileged-port-29850
func TestCliServeRejectsPort_WhenPortIs1194(t *testing.T) {
	dir := t.TempDir()
	writeExampleProject(t, dir)
	_, stderr, code := runCLI(t, context.Background(), "serve", "--project", dir, "--port", "1194")
	if code != 1 {
		t.Fatalf("exit %d, want 1; stderr=%q", code, stderr)
	}
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#construction--configuration
func TestCliServeRejectsPort_WhenPortIsPrivileged(t *testing.T) {
	dir := t.TempDir()
	writeExampleProject(t, dir)
	_, stderr, code := runCLI(t, context.Background(), "serve", "--project", dir, "--port", "1023")
	if code != 1 {
		t.Fatalf("exit %d, want 1; stderr=%q", code, stderr)
	}
}

// Spec-Anchor: proj-spec/chapters/07.3-cli.md#serve-flags
// Spec-Anchor: proj-spec/chapters/07.3-cli.md#error-handling
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
func TestCliServeExitsAddressInUse_WhenPortBusy(t *testing.T) {
	dir := t.TempDir()
	writeExampleProject(t, dir)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port
	_, stderr, code := runCLI(t, context.Background(), "serve", "--project", dir, "--bind", "127.0.0.1", "--port", strconv.Itoa(port))
	if code != 1 {
		t.Fatalf("exit %d, want 1; stderr=%q", code, stderr)
	}
	if !strings.Contains(stderr, "address in use") {
		t.Fatalf("stderr %q must mention address in use", stderr)
	}
}

// Spec-Anchor: proj-spec/chapters/07.3-cli.md#serve-flags
// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#construction--configuration
// Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d5--loopback-only-unprivileged-port-29850
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
func TestCliServeAcceptsBind_WhenHostIsIPv4Loopback(t *testing.T) {
	dir := t.TempDir()
	writeExampleProject(t, dir)
	port := freePort(t, "127.0.0.1")
	stderr, codeCh, cancel := waitServeStarted(t, []string{
		"serve", "--project", dir, "--bind", "127.0.0.1", "--port", strconv.Itoa(port),
	})
	defer cancel()
	wantListen := "listen=" + net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	if !strings.Contains(stderr, wantListen) {
		t.Fatalf("stderr %q must contain %s", stderr, wantListen)
	}
	cancel()
	select {
	case code := <-codeCh:
		if code != 0 {
			t.Fatalf("shutdown exit %d, want 0", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("serve did not exit after cancel")
	}
}

// Spec-Anchor: proj-spec/chapters/07.3-cli.md#serve-flags
// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#construction--configuration
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
func TestCliServeAcceptsBind_WhenHostIsIPv6Loopback(t *testing.T) {
	dir := t.TempDir()
	writeExampleProject(t, dir)
	port := freePort(t, "::1")
	stderr, codeCh, cancel := waitServeStarted(t, []string{
		"serve", "--project", dir, "--bind", "::1", "--port", strconv.Itoa(port),
	})
	defer cancel()
	wantListen := "listen=" + net.JoinHostPort("::1", strconv.Itoa(port))
	if !strings.Contains(stderr, wantListen) {
		t.Fatalf("stderr %q must contain %s", stderr, wantListen)
	}
	cancel()
	select {
	case code := <-codeCh:
		if code != 0 {
			t.Fatalf("shutdown exit %d, want 0", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("serve did not exit after cancel")
	}
}

// Spec-Anchor: proj-spec/chapters/07.3-cli.md#serve-flags
// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#construction--configuration
func TestCliServePrintsTokenAndListen_WhenServeStarts(t *testing.T) {
	dir := t.TempDir()
	writeExampleProject(t, dir)
	port := freePort(t, "127.0.0.1")
	stderr, codeCh, cancel := waitServeStarted(t, []string{
		"serve", "--project", dir, "--port", strconv.Itoa(port),
	})
	defer cancel()
	if !strings.Contains(stderr, "protocolVersion=1") {
		t.Fatalf("stderr %q must contain protocolVersion=1", stderr)
	}
	token := regexp.MustCompile(`accessToken=([0-9a-f]{64})(?:\s|$)`)
	m := token.FindStringSubmatch(stderr)
	if m == nil {
		t.Fatalf("stderr %q must contain accessToken= plus 64 lowercase hex chars", stderr)
	}
	if regexp.MustCompile(`accessToken=[0-9A-F]{64}`).MatchString(stderr) && !regexp.MustCompile(`accessToken=[0-9a-f]{64}`).MatchString(stderr) {
		t.Fatal("accessToken hex must be lowercase")
	}
	cancel()
	select {
	case <-codeCh:
	case <-time.After(2 * time.Second):
		t.Fatal("serve did not exit after cancel")
	}
}

// Spec-Anchor: proj-spec/chapters/10.2-windows-and-wsl.md#paths
func TestCliServeLoadsProject_WhenProjectPathUsesOSSeparators(t *testing.T) {
	dir := t.TempDir()
	writeExampleProject(t, dir)
	mixed := dir
	if runtime.GOOS == "windows" {
		mixed = strings.Replace(dir, `\`, `/`, 1)
	} else {
		mixed = filepath.ToSlash(dir)
	}
	port := freePort(t, "127.0.0.1")
	_, codeCh, cancel := waitServeStarted(t, []string{
		"serve", "--project", mixed, "--port", strconv.Itoa(port),
	})
	defer cancel()
	cancel()
	select {
	case code := <-codeCh:
		if code != 0 {
			t.Fatalf("shutdown exit %d, want 0", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("serve did not exit after cancel")
	}
}
