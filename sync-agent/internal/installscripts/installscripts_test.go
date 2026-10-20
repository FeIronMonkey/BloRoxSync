package installscripts_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(file)
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "proj-spec", "chapters")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "scripts")); err == nil {
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

func scriptPath(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "scripts", name)
	st, err := os.Stat(path)
	if err != nil {
		t.Fatalf("installer script missing: %s (%v)", path, err)
	}
	if st.IsDir() {
		t.Fatalf("installer script is a directory: %s", path)
	}
	return path
}

func lookInterp(names ...string) (string, bool) {
	for _, n := range names {
		if p, err := exec.LookPath(n); err == nil {
			return p, true
		}
	}
	return "", false
}

func requirePwsh(t *testing.T) string {
	t.Helper()
	p, ok := lookInterp("pwsh", "powershell")
	if !ok {
		t.Fatal("PowerShell (pwsh or powershell) is required to run install.ps1 tests")
	}
	return p
}

func requireSh(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		gitBash := filepath.Join("C:\\", "Program Files", "Git", "bin", "bash.exe")
		if _, err := os.Stat(gitBash); err == nil {
			return gitBash
		}
	}
	p, ok := lookInterp("bash", "sh")
	if !ok {
		t.Fatal("POSIX sh/bash is required to run install.sh tests")
	}
	lower := strings.ToLower(p)
	if strings.Contains(lower, `windowsapps\bash.exe`) || strings.Contains(lower, `system32\bash.exe`) {
		t.Fatal("WSL bash.exe cannot run the repo install.sh with Windows paths; install Git bash or run tests on Linux")
	}
	return p
}

type isoEnv struct {
	home  string
	local string
	path  string
	env   []string
}

func isolate(t *testing.T) isoEnv {
	t.Helper()
	home := t.TempDir()
	local := filepath.Join(home, "AppData", "Local")
	if err := os.MkdirAll(local, 0o755); err != nil {
		t.Fatal(err)
	}
	path := os.Getenv("PATH")
	// Drop user-local bin so Linux dry-run/install must print the export hint.
	filtered := make([]string, 0)
	for _, p := range filepath.SplitList(path) {
		if strings.Contains(filepath.ToSlash(p), "/.local/bin") {
			continue
		}
		filtered = append(filtered, p)
	}
	path = strings.Join(filtered, string(os.PathListSeparator))
	env := []string{
		"HOME=" + home,
		"USERPROFILE=" + home,
		"LOCALAPPDATA=" + local,
		"PATH=" + path,
		"SystemRoot=" + os.Getenv("SystemRoot"),
		"SYSTEMROOT=" + os.Getenv("SYSTEMROOT"),
		"WINDIR=" + os.Getenv("WINDIR"),
		"COMSPEC=" + os.Getenv("COMSPEC"),
		"PATHEXT=" + os.Getenv("PATHEXT"),
		"TEMP=" + t.TempDir(),
		"TMP=" + t.TempDir(),
		"ProgramFiles=" + os.Getenv("ProgramFiles"),
		"ProgramW6432=" + os.Getenv("ProgramW6432"),
		"PSModulePath=" + os.Getenv("PSModulePath"),
	}
	if goRoot := os.Getenv("GOROOT"); goRoot != "" {
		env = append(env, "GOROOT="+goRoot)
	}
	return isoEnv{home: home, local: local, path: path, env: env}
}

func (e isoEnv) with(kvs ...string) []string {
	return append(append([]string{}, e.env...), kvs...)
}

func runPs1(t *testing.T, interp, script string, env []string, args ...string) (out string, code int) {
	t.Helper()
	cmdArgs := []string{"-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", script}
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.Command(interp, cmdArgs...)
	cmd.Env = env
	cmd.Dir = filepath.Dir(script)
	b, err := cmd.CombinedOutput()
	code = 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			t.Fatalf("run install.ps1: %v\n%s", err, b)
		}
	}
	return string(b), code
}

func posixPath(p string) string {
	p = filepath.ToSlash(p)
	if runtime.GOOS == "windows" && len(p) >= 2 && p[1] == ':' {
		return "/" + strings.ToLower(string(p[0])) + p[2:]
	}
	return p
}

func containsPath(out, p string) bool {
	if strings.Contains(out, p) {
		return true
	}
	slash := filepath.ToSlash(p)
	if strings.Contains(filepath.ToSlash(out), slash) {
		return true
	}
	return strings.Contains(out, posixPath(p))
}

func runSh(t *testing.T, interp, script string, env []string, args ...string) (out string, code int) {
	t.Helper()
	cmdArgs := append([]string{posixPath(script)}, args...)
	cmd := exec.Command(interp, cmdArgs...)
	cmd.Env = env
	cmd.Dir = filepath.Dir(script)
	b, err := cmd.CombinedOutput()
	code = 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			t.Fatalf("run install.sh: %v\n%s", err, b)
		}
	}
	return string(b), code
}

func assertNoInstallWrites(t *testing.T, e isoEnv) {
	t.Helper()
	for _, p := range []string{
		filepath.Join(e.local, "rbx-sync", "rbx-sync.exe"),
		filepath.Join(e.local, "rbx-sync", "rbx-sync.rbxmx"),
		filepath.Join(e.home, ".local", "bin", "rbx-sync"),
		filepath.Join(e.home, ".local", "share", "rbx-sync", "rbx-sync.rbxmx"),
	} {
		if _, err := os.Stat(p); err == nil {
			t.Errorf("dry-run must not write %s", p)
		}
	}
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func writeFile(t *testing.T, path string, body []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, body, 0o755); err != nil {
		t.Fatal(err)
	}
}

func serveRelease(t *testing.T, files map[string][]byte, sums string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	for name, body := range files {
		n, b := name, body
		mux.HandleFunc("/"+n, func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(b)
		})
	}
	mux.HandleFunc("/SHA256SUMS", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, sums)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func sumsFile(entries map[string][]byte) string {
	var b strings.Builder
	for name, body := range entries {
		fmt.Fprintf(&b, "%s  %s\n", sha256Hex(body), name)
	}
	return b.String()
}

func stubBytes(t *testing.T, pluginExit int) []byte {
	t.Helper()
	dir := t.TempDir()
	src := filepath.Join(dir, "main.go")
	writeFile(t, src, []byte(fmt.Sprintf(`package main
import (
	"fmt"
	"os"
)
func main() {
	if len(os.Args) >= 3 && os.Args[1] == "plugin" && os.Args[2] == "install" {
		os.Exit(%d)
	}
	if len(os.Args) >= 2 && os.Args[1] == "version" {
		fmt.Print("agent=0.0.0 protocol=1\n")
	}
}
`, pluginExit)))
	out := filepath.Join(dir, "stub")
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", out, src)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build stub: %v\n%s", err, b)
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func writeBinOnPath(t *testing.T, dir, name, script string) {
	t.Helper()
	path := filepath.Join(dir, name)
	writeFile(t, path, []byte(script))
}

// Spec-Anchor: proj-spec/chapters/10.1-install.md#one-command-user-install
// Spec-Anchor: proj-spec/chapters/05-file-tree.md#installed-files-user-machine-after-101
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#platform
// Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d10--one-command-install-from-release-assets
// Spec-Anchor: proj-spec/chapters/02-goals-and-non-goals.md#goals
func TestInstallPs1DryRunPrintsWindowsDest_WhenDryRun(t *testing.T) {
	script := scriptPath(t, "install.ps1")
	interp := requirePwsh(t)
	e := isolate(t)
	out, code := runPs1(t, interp, script, e.with(), "-DryRun")
	if code != 0 {
		t.Fatalf("exit %d, want 0; output:\n%s", code, out)
	}
	for _, want := range []string{
		"rbx-sync-windows-amd64.exe",
		"rbx-sync.rbxmx",
		"new",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("dry-run output missing %q\n%s", want, out)
		}
	}
	slashOut := filepath.ToSlash(out)
	if !strings.Contains(slashOut, "rbx-sync/rbx-sync.exe") && !strings.Contains(out, `rbx-sync\rbx-sync.exe`) {
		t.Errorf("dry-run must print windows dest binary path\n%s", out)
	}
	if !strings.Contains(slashOut, "rbx-sync/rbx-sync.rbxmx") && !strings.Contains(out, `rbx-sync\rbx-sync.rbxmx`) {
		t.Errorf("dry-run must print windows dest plugin path\n%s", out)
	}
	if !strings.Contains(strings.ToLower(out), "path") {
		t.Errorf("dry-run must mention PATH\n%s", out)
	}
	if strings.Contains(out, "rbx-sync-linux-amd64") {
		t.Errorf("Windows installer must not select the linux asset\n%s", out)
	}
	assertNoInstallWrites(t, e)
}

// Spec-Anchor: proj-spec/chapters/10.1-install.md#one-command-user-install
// Spec-Anchor: proj-spec/chapters/05-file-tree.md#installed-files-user-machine-after-101
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#platform
// Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d10--one-command-install-from-release-assets
// Spec-Anchor: proj-spec/chapters/02-goals-and-non-goals.md#goals
func TestInstallShDryRunPrintsLinuxDest_WhenDryRun(t *testing.T) {
	script := scriptPath(t, "install.sh")
	interp := requireSh(t)
	e := isolate(t)
	out, code := runSh(t, interp, script, e.with("WSL_DISTRO_NAME=Ubuntu"), "--dry-run")
	if code != 0 {
		t.Fatalf("exit %d, want 0; output:\n%s", code, out)
	}
	for _, want := range []string{
		"rbx-sync-linux-amd64",
		"rbx-sync.rbxmx",
		filepath.ToSlash(filepath.Join(e.home, ".local", "bin", "rbx-sync")),
		filepath.ToSlash(filepath.Join(e.home, ".local", "share", "rbx-sync", "rbx-sync.rbxmx")),
		`export PATH="$HOME/.local/bin:$PATH"`,
	} {
		if !containsPath(out, want) && !strings.Contains(out, want) {
			t.Errorf("dry-run output missing %q\n%s", want, out)
		}
	}
	if strings.Contains(out, "rbx-sync-windows-amd64.exe") {
		t.Errorf("WSL/Linux installer must use the linux asset, not windows\n%s", out)
	}
	assertNoInstallWrites(t, e)
}

// Spec-Anchor: proj-spec/chapters/10.1-install.md#one-command-user-install
// Spec-Anchor: proj-spec/chapters/10.1-install.md#release-assets
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#platform
func TestInstallPs1RefusesReplace_WhenChecksumMismatches(t *testing.T) {
	script := scriptPath(t, "install.ps1")
	interp := requirePwsh(t)
	e := isolate(t)
	destDir := filepath.Join(e.local, "rbx-sync")
	oldBin := []byte("OLD-EXE")
	oldPlug := []byte("OLD-RBXMX")
	writeFile(t, filepath.Join(destDir, "rbx-sync.exe"), oldBin)
	writeFile(t, filepath.Join(destDir, "rbx-sync.rbxmx"), oldPlug)

	files := map[string][]byte{
		"rbx-sync-windows-amd64.exe": []byte("NEW-EXE"),
		"rbx-sync.rbxmx":             []byte("NEW-PLUGIN"),
	}
	bad := strings.Repeat("0", 64)
	srv := serveRelease(t, files, bad+"  rbx-sync-windows-amd64.exe\n"+bad+"  rbx-sync.rbxmx\n")
	out, code := runPs1(t, interp, script, e.with("RBX_SYNC_RELEASE_BASE="+srv.URL))
	if code == 0 {
		t.Fatalf("checksum mismatch must exit non-zero; output:\n%s", out)
	}
	gotBin, _ := os.ReadFile(filepath.Join(destDir, "rbx-sync.exe"))
	gotPlug, _ := os.ReadFile(filepath.Join(destDir, "rbx-sync.rbxmx"))
	if string(gotBin) != string(oldBin) {
		t.Errorf("binary replaced on checksum mismatch")
	}
	if string(gotPlug) != string(oldPlug) {
		t.Errorf("plugin replaced on checksum mismatch")
	}
}

// Spec-Anchor: proj-spec/chapters/10.1-install.md#one-command-user-install
// Spec-Anchor: proj-spec/chapters/10.1-install.md#release-assets
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#platform
func TestInstallShRefusesReplace_WhenChecksumMismatches(t *testing.T) {
	script := scriptPath(t, "install.sh")
	interp := requireSh(t)
	e := isolate(t)
	binPath := filepath.Join(e.home, ".local", "bin", "rbx-sync")
	plugPath := filepath.Join(e.home, ".local", "share", "rbx-sync", "rbx-sync.rbxmx")
	oldBin := []byte("OLD-BIN")
	oldPlug := []byte("OLD-RBXMX")
	writeFile(t, binPath, oldBin)
	writeFile(t, plugPath, oldPlug)

	files := map[string][]byte{
		"rbx-sync-linux-amd64": []byte("NEW-BIN"),
		"rbx-sync.rbxmx":       []byte("NEW-PLUGIN"),
	}
	bad := strings.Repeat("0", 64)
	srv := serveRelease(t, files, bad+"  rbx-sync-linux-amd64\n"+bad+"  rbx-sync.rbxmx\n")
	out, code := runSh(t, interp, script, e.with("RBX_SYNC_RELEASE_BASE="+srv.URL))
	if code == 0 {
		t.Fatalf("checksum mismatch must exit non-zero; output:\n%s", out)
	}
	gotBin, _ := os.ReadFile(binPath)
	gotPlug, _ := os.ReadFile(plugPath)
	if string(gotBin) != string(oldBin) {
		t.Errorf("binary replaced on checksum mismatch")
	}
	if string(gotPlug) != string(oldPlug) {
		t.Errorf("plugin replaced on checksum mismatch")
	}
}

// Spec-Anchor: proj-spec/chapters/10.1-install.md#one-command-user-install
func TestInstallPs1HonorsReleaseBase_WhenEnvSet(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("install.ps1 writes rbx-sync.exe; execute plugin/install paths on Windows")
	}
	script := scriptPath(t, "install.ps1")
	interp := requirePwsh(t)
	e := isolate(t)
	stub := stubBytes(t, 0)
	files := map[string][]byte{
		"rbx-sync-windows-amd64.exe": stub,
		"rbx-sync.rbxmx":             []byte("<roblox version=\"4\"/>"),
	}
	srv := serveRelease(t, files, sumsFile(files))
	out, code := runPs1(t, interp, script, e.with("RBX_SYNC_RELEASE_BASE="+srv.URL))
	if code != 0 {
		t.Fatalf("exit %d, want 0; output:\n%s", code, out)
	}
	bin := filepath.Join(e.local, "rbx-sync", "rbx-sync.exe")
	plug := filepath.Join(e.local, "rbx-sync", "rbx-sync.rbxmx")
	if _, err := os.Stat(bin); err != nil {
		t.Fatalf("windows dest binary missing: %v\n%s", err, out)
	}
	if _, err := os.Stat(plug); err != nil {
		t.Fatalf("windows dest plugin missing: %v\n%s", err, out)
	}
}

// Spec-Anchor: proj-spec/chapters/10.1-install.md#one-command-user-install
func TestInstallShHonorsReleaseBase_WhenEnvSet(t *testing.T) {
	script := scriptPath(t, "install.sh")
	interp := requireSh(t)
	e := isolate(t)
	stub := stubBytes(t, 0)
	files := map[string][]byte{
		"rbx-sync-linux-amd64": stub,
		"rbx-sync.rbxmx":       []byte("<roblox version=\"4\"/>"),
	}
	srv := serveRelease(t, files, sumsFile(files))
	out, code := runSh(t, interp, script, e.with("RBX_SYNC_RELEASE_BASE="+srv.URL))
	if code != 0 {
		t.Fatalf("exit %d, want 0; output:\n%s", code, out)
	}
	bin := filepath.Join(e.home, ".local", "bin", "rbx-sync")
	plug := filepath.Join(e.home, ".local", "share", "rbx-sync", "rbx-sync.rbxmx")
	if _, err := os.Stat(bin); err != nil {
		t.Fatalf("linux dest binary missing: %v\n%s", err, out)
	}
	if _, err := os.Stat(plug); err != nil {
		t.Fatalf("linux dest plugin missing: %v\n%s", err, out)
	}
	if !containsPath(out, plug) {
		t.Errorf("unreachable plugins folder must print staged rbxmx path\n%s", out)
	}
}

// Spec-Anchor: proj-spec/chapters/10.1-install.md#one-command-user-install
func TestInstallPs1ExitsPluginStatus_WhenPluginsReachableAndInstallFails(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("install.ps1 plugin-install status is exercised on Windows")
	}
	script := scriptPath(t, "install.ps1")
	interp := requirePwsh(t)
	e := isolate(t)
	if err := os.MkdirAll(filepath.Join(e.local, "Roblox", "Plugins"), 0o755); err != nil {
		t.Fatal(err)
	}
	stub := stubBytes(t, 7)
	files := map[string][]byte{
		"rbx-sync-windows-amd64.exe": stub,
		"rbx-sync.rbxmx":             []byte("<roblox version=\"4\"/>"),
	}
	srv := serveRelease(t, files, sumsFile(files))
	out, code := runPs1(t, interp, script, e.with("RBX_SYNC_RELEASE_BASE="+srv.URL))
	if code != 7 {
		t.Fatalf("exit %d, want plugin install status 7; output:\n%s", code, out)
	}
}

// Spec-Anchor: proj-spec/chapters/10.1-install.md#one-command-user-install
func TestInstallShExitsPluginStatus_WhenPluginsReachableAndInstallFails(t *testing.T) {
	script := scriptPath(t, "install.sh")
	interp := requireSh(t)
	e := isolate(t)
	plugins := filepath.Join(e.home, "winlocal", "Roblox", "Plugins")
	if err := os.MkdirAll(plugins, 0o755); err != nil {
		t.Fatal(err)
	}
	binDir := filepath.Join(e.home, "fakes")
	unixLocal := filepath.ToSlash(filepath.Join(e.home, "winlocal"))
	writeBinOnPath(t, binDir, "cmd.exe", "#!/bin/sh\nprintf '%s\\n' 'C:\\Users\\FixtureUser\\AppData\\Local'\n")
	writeBinOnPath(t, binDir, "wslpath", "#!/bin/sh\nprintf '%s\\n' '"+unixLocal+"'\n")
	stub := stubBytes(t, 7)
	files := map[string][]byte{
		"rbx-sync-linux-amd64": stub,
		"rbx-sync.rbxmx":       []byte("<roblox version=\"4\"/>"),
	}
	srv := serveRelease(t, files, sumsFile(files))
	path := binDir + string(os.PathListSeparator) + e.path
	out, code := runSh(t, interp, script, e.with(
		"PATH="+path,
		"RBX_SYNC_RELEASE_BASE="+srv.URL,
	))
	if code != 7 {
		t.Fatalf("exit %d, want plugin install status 7; output:\n%s", code, out)
	}
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#plugin-install-helper
func TestInstallScriptsOmitHardCodedUsername_WhenRead(t *testing.T) {
	root := repoRoot(t)
	for _, name := range []string{"install.ps1", "install.sh"} {
		body, err := os.ReadFile(filepath.Join(root, "scripts", name))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		lower := strings.ToLower(string(body))
		if strings.Contains(lower, "/users/sende") || strings.Contains(lower, `\users\sende`) {
			t.Errorf("%s hard-codes a Windows username", name)
		}
		if strings.Contains(lower, "/mnt/c/users/sende") {
			t.Errorf("%s hard-codes a Windows username", name)
		}
	}
}
