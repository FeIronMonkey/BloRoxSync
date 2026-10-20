package release_test

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"roblox-sync/sync-agent/internal/release"
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
			return dir
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

func sha256File(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// Spec-Anchor: proj-spec/chapters/04-stack.md#cross-compilation-rbx-sync
// Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d8--rbx-sync-language-go
func TestReleaseCrossCompilesWindowsAndLinux_WhenCgoDisabled(t *testing.T) {
	root := repoRoot(t)
	out := t.TempDir()
	win := filepath.Join(out, "rbx-sync.exe")
	linux := filepath.Join(out, "rbx-sync")
	if err := release.CrossCompile(filepath.Join(root, "sync-agent"), win, linux); err != nil {
		t.Fatalf("cross-compile: %v", err)
	}
	winBody, err := os.ReadFile(win)
	if err != nil {
		t.Fatal(err)
	}
	linuxBody, err := os.ReadFile(linux)
	if err != nil {
		t.Fatal(err)
	}
	if len(winBody) < 2 || winBody[0] != 'M' || winBody[1] != 'Z' {
		t.Fatalf("windows artifact is not a PE executable")
	}
	if len(linuxBody) < 4 || linuxBody[0] != 0x7f || linuxBody[1] != 'E' || linuxBody[2] != 'L' || linuxBody[3] != 'F' {
		t.Fatalf("linux artifact is not an ELF executable")
	}
}

// Spec-Anchor: proj-spec/chapters/10.1-install.md#release-assets
func TestReleaseRecipeWritesAssetsAndSums_WhenOutDirEmpty(t *testing.T) {
	root := repoRoot(t)
	out := t.TempDir()
	if err := release.Package(filepath.Join(root, "sync-agent"), filepath.Join(root, "studio-plugin", "src"), out); err != nil {
		t.Fatalf("package release: %v", err)
	}
	names := []string{
		"rbx-sync-windows-amd64.exe",
		"rbx-sync-linux-amd64",
		"rbx-sync.rbxmx",
		"SHA256SUMS",
	}
	for _, n := range names {
		if _, err := os.Stat(filepath.Join(out, n)); err != nil {
			t.Errorf("missing %s: %v", n, err)
		}
	}
	sumsPath := filepath.Join(out, "SHA256SUMS")
	f, err := os.Open(sumsPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	found := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			t.Fatalf("SHA256SUMS line must be hash and path, got %q", line)
		}
		found[fields[1]] = strings.ToLower(fields[0])
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"rbx-sync-windows-amd64.exe", "rbx-sync-linux-amd64", "rbx-sync.rbxmx"} {
		want := sha256File(t, filepath.Join(out, n))
		got, ok := found[n]
		if !ok {
			t.Errorf("SHA256SUMS missing %s", n)
			continue
		}
		if got != want {
			t.Errorf("SHA256SUMS hash for %s does not match file", n)
		}
	}
}

// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#out-of-scope-checks-must-remain-true
// Spec-Anchor: proj-spec/chapters/02-goals-and-non-goals.md#non-goals
func TestRepoShipsNoEditorExtension_WhenTreeWalked(t *testing.T) {
	root := repoRoot(t)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		slash := filepath.ToSlash(rel)
		if d.IsDir() {
			base := d.Name()
			if base == ".git" || base == "node_modules" || base == "tmp" {
				return filepath.SkipDir
			}
			return nil
		}
		name := strings.ToLower(d.Name())
		if strings.HasSuffix(name, ".vsix") {
			t.Errorf("editor extension artifact shipped: %s", slash)
		}
		if name == "package.json" {
			body, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			lower := strings.ToLower(string(body))
			if strings.Contains(lower, `"vscode"`) || strings.Contains(lower, "engines") && strings.Contains(lower, "vscode") {
				t.Errorf("VS Code/Cursor extension manifest shipped: %s", slash)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#out-of-scope-checks-must-remain-true
// Spec-Anchor: proj-spec/chapters/04-stack.md#rbx-sync-dependencies
// Spec-Anchor: proj-spec/chapters/12-testing.md#source-ban-in-tests
func TestRepoDoesNotDependOnOtherSyncTools_WhenModulesAndTreeWalked(t *testing.T) {
	root := repoRoot(t)
	banned := []string{"github.com/rojo-rbx", "github.com/argon-rbx", "rojo-rbx/", "argon-rbx/"}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "tmp" {
				return filepath.SkipDir
			}
			return nil
		}
		name := d.Name()
		if name == "go.mod" || name == "go.sum" || strings.HasSuffix(name, "_test.go") {
			if strings.HasSuffix(filepath.ToSlash(path), "internal/release/release_test.go") {
				return nil
			}
			body, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			text := string(body)
			for _, b := range banned {
				if strings.Contains(text, b) {
					t.Errorf("%s imports or snapshots another sync tool (%s)", path, b)
				}
			}
		}
		if name == "default.project.json" {
			t.Errorf("other-tool project format shipped: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// Spec-Anchor: proj-spec/chapters/10.1-install.md#contributor-build-optional
// Spec-Anchor: proj-spec/chapters/05-file-tree.md#sync-agent-target-after-implementation
func TestReleaseCommandExists_WhenReleaseRecipeDocumented(t *testing.T) {
	root := repoRoot(t)
	mainPath := filepath.Join(root, "sync-agent", "cmd", "package-release", "main.go")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("release recipe command missing: %v", err)
	}
}
