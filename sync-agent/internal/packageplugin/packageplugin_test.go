package packageplugin_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"roblox-sync/sync-agent/internal/installplugin"
	"roblox-sync/sync-agent/internal/packageplugin"
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
			if _, err := os.Stat(filepath.Join(dir, "studio-plugin", "src")); err == nil {
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

// Spec-Anchor: proj-spec/chapters/05-file-tree.md#studio-plugin-target-after-implementation
// Spec-Anchor: proj-spec/chapters/10.1-install.md#contributor-build-optional
func TestPackagepluginWritesRbxmx_WhenSrcHasLuau(t *testing.T) {
	root := repoRoot(t)
	src := filepath.Join(root, "studio-plugin", "src")
	out := filepath.Join(t.TempDir(), "rbx-sync.rbxmx")
	if err := packageplugin.Package(src, out); err != nil {
		t.Fatalf("package: %v", err)
	}
	st, err := os.Stat(out)
	if err != nil {
		t.Fatalf("artifact missing: %v", err)
	}
	if st.IsDir() {
		t.Fatal("artifact must be a file")
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, `<roblox`) || !strings.Contains(text, `version="4"`) {
		t.Fatalf("artifact is not a version-4 rbxmx: %q", text[:min(120, len(text))])
	}
	if !strings.Contains(text, `class="Script"`) {
		t.Fatal("plugin rbxmx must contain a Script item")
	}
	mapping, err := os.ReadFile(filepath.Join(src, "mapping.luau"))
	if err != nil {
		t.Fatal(err)
	}
	session, err := os.ReadFile(filepath.Join(src, "session.luau"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, string(mapping)) {
		t.Fatal("rbxmx must embed mapping.luau source")
	}
	if !strings.Contains(text, string(session)) {
		t.Fatal("rbxmx must embed session.luau source")
	}
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#plugin-install-helper
// Spec-Anchor: proj-spec/chapters/10.1-install.md#plugin-destination
func TestPackagepluginArtifactCopies_WhenInstallpluginRuns(t *testing.T) {
	root := repoRoot(t)
	src := filepath.Join(root, "studio-plugin", "src")
	artifact := filepath.Join(t.TempDir(), "rbx-sync.rbxmx")
	if err := packageplugin.Package(src, artifact); err != nil {
		t.Fatalf("package: %v", err)
	}
	local := filepath.Join(t.TempDir(), "Local")
	env := installplugin.Env{
		GOOS:          "windows",
		LocalAppData:  local,
		ExecutableDir: t.TempDir(),
	}
	var stderr strings.Builder
	code := installplugin.Command(env, []string{"install", "--file", artifact}, &stderr)
	if code != 0 {
		t.Fatalf("plugin install exit %d; stderr=%q", code, stderr.String())
	}
	dest := filepath.Join(local, "Roblox", "Plugins", "rbx-sync.rbxmx")
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("copied artifact missing: %v", err)
	}
	want, err := os.ReadFile(artifact)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatal("installed file must match packaged rbxmx")
	}
}

// Spec-Anchor: proj-spec/chapters/10.1-install.md#contributor-build-optional
func TestPackagepluginCommandExists_WhenContributorBuildDocumented(t *testing.T) {
	root := repoRoot(t)
	mainPath := filepath.Join(root, "sync-agent", "cmd", "package-plugin", "main.go")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("contributor packager command missing: %v", err)
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go not on PATH; packager main.go exists")
	}
	cmd := exec.Command(goBin, "run", "./cmd/package-plugin", "-src", filepath.Join(root, "studio-plugin", "src"), "-out", filepath.Join(t.TempDir(), "rbx-sync.rbxmx"))
	cmd.Dir = filepath.Join(root, "sync-agent")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run ./cmd/package-plugin: %v\n%s", err, out)
	}
}
