package installplugin_test

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"roblox-sync/sync-agent/internal/installplugin"
)

func writeFixture(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func command(env installplugin.Env, args ...string) (stderr string, code int) {
	var errBuf bytes.Buffer
	code = installplugin.Command(env, args, &errBuf)
	return errBuf.String(), code
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#plugin-install-helper
// Spec-Anchor: proj-spec/chapters/10.1-install.md#plugin-destination
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#platform
func TestInstallpluginResolvesWindowsDest_WhenGOOSIsWindows(t *testing.T) {
	root := t.TempDir()
	local := filepath.Join(root, "AppData", "Local")
	src := writeFixture(t, root, "rbx-sync.rbxmx", "<roblox/>")
	env := installplugin.Env{
		GOOS:           "windows",
		LocalAppData:   local,
		ExecutableDir:  root,
		LookPath:       func(string) (string, error) { return "", os.ErrNotExist },
		CombinedOutput: func(string, ...string) ([]byte, error) { t.Fatal("cmd.exe must not run on Windows dest"); return nil, nil },
		Glob:           func(string) ([]string, error) { t.Fatal("WSL glob must not run on Windows dest"); return nil, nil },
	}
	stderr, code := command(env, "install", "--file", src)
	if code != 0 {
		t.Fatalf("exit %d, want 0; stderr=%q", code, stderr)
	}
	dest := filepath.Join(local, "Roblox", "Plugins", "rbx-sync.rbxmx")
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(got) != "<roblox/>" {
		t.Fatalf("dest body %q, want fixture", got)
	}
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#plugin-install-helper
// Spec-Anchor: proj-spec/chapters/10.1-install.md#plugin-destination
func TestInstallpluginResolvesWslDest_WhenCmdAndWslpathSucceed(t *testing.T) {
	root := t.TempDir()
	src := writeFixture(t, root, "rbx-sync.rbxmx", "<roblox/>")
	plugins := filepath.Join(root, "Plugins")
	if err := os.MkdirAll(plugins, 0o755); err != nil {
		t.Fatal(err)
	}
	winLocal := `C:\Users\InjectedUser\AppData\Local`
	unixLocal := filepath.ToSlash(root)
	env := installplugin.Env{
		GOOS:          "linux",
		Home:          root,
		ExecutableDir: root,
		LookPath: func(file string) (string, error) {
			if file == "cmd.exe" {
				return "/mnt/c/Windows/System32/cmd.exe", nil
			}
			return "", os.ErrNotExist
		},
		CombinedOutput: func(name string, args ...string) ([]byte, error) {
			joined := strings.Join(append([]string{name}, args...), " ")
			if strings.Contains(joined, "cmd.exe") && strings.Contains(joined, "%LOCALAPPDATA%") {
				return []byte(winLocal + "\r\n"), nil
			}
			if name == "wslpath" && len(args) >= 2 && args[0] == "-u" && strings.Contains(args[1], "InjectedUser") {
				return []byte(unixLocal + "\n"), nil
			}
			t.Fatalf("unexpected command %s %v", name, args)
			return nil, nil
		},
		Glob: func(string) ([]string, error) {
			t.Fatal("unique /mnt/c glob must not run when wslpath succeeds")
			return nil, nil
		},
	}
	stderr, code := command(env, "install", "--file", src)
	if code != 0 {
		t.Fatalf("exit %d, want 0; stderr=%q", code, stderr)
	}
	dest := filepath.Join(root, "Roblox", "Plugins", "rbx-sync.rbxmx")
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(got) != "<roblox/>" {
		t.Fatalf("dest body %q", got)
	}
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#plugin-install-helper
func TestInstallpluginResolvesWslDest_WhenUniqueMntPluginsDirExists(t *testing.T) {
	root := t.TempDir()
	src := writeFixture(t, root, "rbx-sync.rbxmx", "<roblox/>")
	only := filepath.Join(root, "mnt", "c", "Users", "OnlyOne", "AppData", "Local", "Roblox", "Plugins")
	if err := os.MkdirAll(only, 0o755); err != nil {
		t.Fatal(err)
	}
	env := installplugin.Env{
		GOOS:          "linux",
		Home:          root,
		ExecutableDir: root,
		LookPath:      func(string) (string, error) { return "", os.ErrNotExist },
		CombinedOutput: func(string, ...string) ([]byte, error) {
			t.Fatal("cmd.exe must not run when LookPath fails")
			return nil, nil
		},
		Glob: func(pattern string) ([]string, error) {
			if pattern != "/mnt/c/Users/*/AppData/Local/Roblox/Plugins" {
				t.Fatalf("glob pattern %q", pattern)
			}
			return []string{only}, nil
		},
	}
	stderr, code := command(env, "install", "--file", src)
	if code != 0 {
		t.Fatalf("exit %d, want 0; stderr=%q", code, stderr)
	}
	dest := filepath.Join(only, "rbx-sync.rbxmx")
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("dest missing: %v", err)
	}
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#plugin-install-helper
// Spec-Anchor: proj-spec/chapters/07.3-cli.md#error-handling
func TestInstallpluginExitsDestMissing_WhenMntPluginsDirIsNotUnique(t *testing.T) {
	root := t.TempDir()
	src := writeFixture(t, root, "rbx-sync.rbxmx", "<roblox/>")
	env := installplugin.Env{
		GOOS:          "linux",
		Home:          root,
		ExecutableDir: root,
		LookPath:      func(string) (string, error) { return "", os.ErrNotExist },
		Glob: func(string) ([]string, error) {
			return []string{
				"/mnt/c/Users/A/AppData/Local/Roblox/Plugins",
				"/mnt/c/Users/B/AppData/Local/Roblox/Plugins",
			}, nil
		},
	}
	stderr, code := command(env, "install", "--file", src)
	if code != 1 {
		t.Fatalf("exit %d, want 1; stderr=%q", code, stderr)
	}
	if !strings.Contains(stderr, `%LOCALAPPDATA%\Roblox\Plugins`) {
		t.Fatalf("stderr %q must print expected Windows path", stderr)
	}
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#plugin-install-helper
// Spec-Anchor: proj-spec/chapters/07.3-cli.md#error-handling
func TestInstallpluginExitsDestMissing_WhenLinuxHasNoWindowsPlugins(t *testing.T) {
	root := t.TempDir()
	src := writeFixture(t, root, "rbx-sync.rbxmx", "<roblox/>")
	env := installplugin.Env{
		GOOS:          "linux",
		Home:          root,
		ExecutableDir: root,
		LookPath:      func(string) (string, error) { return "", os.ErrNotExist },
		Glob:          func(string) ([]string, error) { return nil, nil },
	}
	stderr, code := command(env, "install", "--file", src)
	if code != 1 {
		t.Fatalf("exit %d, want 1; stderr=%q", code, stderr)
	}
	if !strings.Contains(stderr, `%LOCALAPPDATA%\Roblox\Plugins`) {
		t.Fatalf("stderr %q must print expected Windows path", stderr)
	}
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#plugin-install-helper
func TestInstallpluginSourceOmitsHardcodedUsername_WhenInspected(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	usersSep := regexp.MustCompile(`(?i)[/\\]Users[/\\]([A-Za-z0-9._-]+)`)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		body, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range usersSep.FindAllSubmatch(body, -1) {
			if string(m[1]) == "*" {
				continue
			}
			t.Fatalf("%s hard-codes username %q", e.Name(), m[1])
		}
	}
}

// Spec-Anchor: proj-spec/chapters/07.3-cli.md#plugin-install-flags
func TestInstallpluginResolvesFile_WhenBesideBinary(t *testing.T) {
	root := t.TempDir()
	src := writeFixture(t, root, "rbx-sync.rbxmx", "beside")
	local := filepath.Join(root, "Local")
	env := installplugin.Env{
		GOOS:          "windows",
		LocalAppData:  local,
		ExecutableDir: root,
	}
	stderr, code := command(env, "install")
	if code != 0 {
		t.Fatalf("exit %d, want 0; stderr=%q", code, stderr)
	}
	got, err := os.ReadFile(filepath.Join(local, "Roblox", "Plugins", "rbx-sync.rbxmx"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "beside" {
		t.Fatalf("copied %q, want beside-binary default", got)
	}
	_ = src
}

// Spec-Anchor: proj-spec/chapters/07.3-cli.md#plugin-install-flags
func TestInstallpluginResolvesFile_WhenLinuxShareExists(t *testing.T) {
	root := t.TempDir()
	src := writeFixture(t, filepath.Join(root, ".local", "share", "rbx-sync"), "rbx-sync.rbxmx", "share")
	only := filepath.Join(root, "Plugins")
	if err := os.MkdirAll(only, 0o755); err != nil {
		t.Fatal(err)
	}
	env := installplugin.Env{
		GOOS:          "linux",
		Home:          root,
		ExecutableDir: filepath.Join(root, "bin"),
		LookPath:      func(string) (string, error) { return "", os.ErrNotExist },
		Glob:          func(string) ([]string, error) { return []string{only}, nil },
	}
	if err := os.MkdirAll(env.ExecutableDir, 0o755); err != nil {
		t.Fatal(err)
	}
	stderr, code := command(env, "install")
	if code != 0 {
		t.Fatalf("exit %d, want 0; stderr=%q", code, stderr)
	}
	got, err := os.ReadFile(filepath.Join(only, "rbx-sync.rbxmx"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "share" {
		t.Fatalf("copied %q, want linux share default", got)
	}
	_ = src
}

// Spec-Anchor: proj-spec/chapters/07.3-cli.md#plugin-install-flags
func TestInstallpluginPrefersBinaryDir_WhenBothDefaultsExist(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "rbx-sync.rbxmx", "beside")
	writeFixture(t, filepath.Join(root, ".local", "share", "rbx-sync"), "rbx-sync.rbxmx", "share")
	only := filepath.Join(root, "Plugins")
	if err := os.MkdirAll(only, 0o755); err != nil {
		t.Fatal(err)
	}
	env := installplugin.Env{
		GOOS:          "linux",
		Home:          root,
		ExecutableDir: root,
		LookPath:      func(string) (string, error) { return "", os.ErrNotExist },
		Glob:          func(string) ([]string, error) { return []string{only}, nil },
	}
	stderr, code := command(env, "install")
	if code != 0 {
		t.Fatalf("exit %d, want 0; stderr=%q", code, stderr)
	}
	got, err := os.ReadFile(filepath.Join(only, "rbx-sync.rbxmx"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "beside" {
		t.Fatalf("copied %q, want binary-dir first", got)
	}
}

// Spec-Anchor: proj-spec/chapters/07.3-cli.md#plugin-install-flags
func TestInstallpluginUsesFileFlag_WhenProvided(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "rbx-sync.rbxmx", "beside")
	override := writeFixture(t, root, "custom.rbxmx", "custom")
	local := filepath.Join(root, "Local")
	env := installplugin.Env{
		GOOS:          "windows",
		LocalAppData:  local,
		ExecutableDir: root,
	}
	stderr, code := command(env, "install", "--file", override)
	if code != 0 {
		t.Fatalf("exit %d, want 0; stderr=%q", code, stderr)
	}
	got, err := os.ReadFile(filepath.Join(local, "Roblox", "Plugins", "rbx-sync.rbxmx"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "custom" {
		t.Fatalf("copied %q, want --file override", got)
	}
}

// Spec-Anchor: proj-spec/chapters/07.3-cli.md#plugin-install-flags
func TestInstallpluginSkipsLinuxShare_WhenWindowsProcess(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, filepath.Join(root, ".local", "share", "rbx-sync"), "rbx-sync.rbxmx", "share")
	local := filepath.Join(root, "Local")
	env := installplugin.Env{
		GOOS:          "windows",
		Home:          root,
		LocalAppData:  local,
		ExecutableDir: filepath.Join(root, "bin"),
	}
	if err := os.MkdirAll(env.ExecutableDir, 0o755); err != nil {
		t.Fatal(err)
	}
	stderr, code := command(env, "install")
	if code == 0 {
		t.Fatalf("windows default must not use linux share; stderr=%q", stderr)
	}
}

// Spec-Anchor: proj-spec/chapters/07.3-cli.md#commands
func TestInstallpluginUnknownSubcommandExitsTwo_WhenNotInstall(t *testing.T) {
	stderr, code := command(installplugin.Env{GOOS: "windows"}, "remove")
	if code != 2 {
		t.Fatalf("exit %d, want 2; stderr=%q", code, stderr)
	}
	if strings.TrimSpace(stderr) == "" {
		t.Fatal("unknown plugin subcommand must write stderr")
	}
}
