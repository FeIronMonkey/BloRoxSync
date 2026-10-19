package mapfs

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func atomicRepoRoot(t *testing.T) string {
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

func atomicOpenExample(t *testing.T) (*FS, string) {
	t.Helper()
	path := filepath.Join(atomicRepoRoot(t), "sync-protocol", "fixtures", "v1", "mapping", "valid-project-example.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	proj, err := json.Marshal(doc["project"])
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rbx-sync.json"), proj, 0o644); err != nil {
		t.Fatal(err)
	}
	fs, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	return fs, dir
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#apply-from-roblox-studio
func TestMapfsUpsertIsAtomic_WhenWriteSucceeds(t *testing.T) {
	orig := commitAtomic
	t.Cleanup(func() { commitAtomic = orig })

	var sawTemp bool
	commitAtomic = func(src, dest string) error {
		if _, err := os.Stat(src); err != nil {
			t.Fatalf("temp missing before rename: %v", err)
		}
		if filepath.Dir(src) != filepath.Dir(dest) {
			t.Fatalf("temp %s not in dest dir of %s", src, dest)
		}
		if filepath.Base(dest) != "Foo.luau" {
			t.Fatalf("dest = %s, want Foo.luau", dest)
		}
		if _, err := os.Stat(dest); err == nil {
			t.Fatal("dest already existed before rename; expected create via temp")
		}
		sawTemp = true
		return orig(src, dest)
	}

	fs, dir := atomicOpenExample(t)
	payload := strings.Repeat("x", 4096)
	if err := fs.Upsert("server/Foo", "ModuleScript", payload); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if !sawTemp {
		t.Fatal("upsert did not commit via temp file + rename")
	}
	destDir := filepath.Join(dir, "src", "server")
	entries, err := os.ReadDir(destDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "Foo.luau" {
		t.Fatalf("dest entries = %v, want only Foo.luau (no leftover temp)", entries)
	}
	got, err := os.ReadFile(filepath.Join(destDir, "Foo.luau"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != payload {
		t.Fatalf("incomplete write, len=%d", len(got))
	}
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#apply-from-roblox-studio
func TestMapfsUpsertPreservesDest_WhenAtomicCommitFails(t *testing.T) {
	orig := commitAtomic
	t.Cleanup(func() { commitAtomic = orig })

	fs, dir := atomicOpenExample(t)
	if err := fs.Upsert("server/Foo", "ModuleScript", "old\n"); err != nil {
		t.Fatal(err)
	}

	commitAtomic = func(src, dest string) error {
		if _, err := os.Stat(src); err != nil {
			t.Fatalf("temp missing: %v", err)
		}
		return errors.New("forced commit failure")
	}
	err := fs.Upsert("server/Foo", "ModuleScript", "new-should-not-land\n")
	if err == nil {
		t.Fatal("expected commit failure")
	}

	dest := filepath.Join(dir, "src", "server", "Foo.luau")
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "old\n" {
		t.Fatalf("dest mutated despite failed atomic commit: %q", got)
	}
	entries, err := os.ReadDir(filepath.Join(dir, "src", "server"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != "Foo.luau" {
			t.Fatalf("leftover after failed commit: %s", e.Name())
		}
	}
}
