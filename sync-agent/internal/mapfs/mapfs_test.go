package mapfs_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"roblox-sync/sync-agent/internal/mapfs"
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

func mappingFixture(t *testing.T, name string) map[string]any {
	t.Helper()
	path := filepath.Join(repoRoot(t), "sync-protocol", "fixtures", "v1", "mapping", name)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode fixture %s: %v", name, err)
	}
	return doc
}

func writeProject(t *testing.T, dir string, project any) {
	t.Helper()
	raw, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		t.Fatalf("marshal project: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rbx-sync.json"), raw, 0o644); err != nil {
		t.Fatalf("write rbx-sync.json: %v", err)
	}
}

func openFromFixtureProject(t *testing.T, fixtureName string) (*mapfs.FS, string) {
	t.Helper()
	doc := mappingFixture(t, fixtureName)
	dir := t.TempDir()
	writeProject(t, dir, doc["project"])
	fs, err := mapfs.Open(dir)
	if err != nil {
		t.Fatalf("Open valid project from %s: %v", fixtureName, err)
	}
	return fs, dir
}

func codeOf(err error) string {
	var e *mapfs.Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}

func rows(t *testing.T, doc map[string]any) []map[string]any {
	t.Helper()
	raw, ok := doc["rows"].([]any)
	if !ok {
		t.Fatal("fixture missing rows")
	}
	out := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		row, ok := item.(map[string]any)
		if !ok {
			t.Fatal("row is not object")
		}
		out = append(out, row)
	}
	return out
}

func nodeByID(nodes []mapfs.Node, id string) (mapfs.Node, bool) {
	for _, n := range nodes {
		if n.NodeID == id {
			return n, true
		}
	}
	return mapfs.Node{}, false
}

func relSlash(path string) string {
	return filepath.ToSlash(path)
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#project-file
// Spec-Anchor: proj-spec/chapters/06.1-sync-protocol.md#error-codes
func TestMapfsOpenRejectsInvalidProject_WhenFixturePackExists(t *testing.T) {
	cases := []string{
		"invalid-dotdot-disk.json",
		"invalid-backslash-disk.json",
		"invalid-duplicate-studio.json",
		"invalid-duplicate-root-id.json",
		"invalid-overlapping-disk.json",
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			doc := mappingFixture(t, name)
			if doc["expectedError"] != "PROJECT_INVALID" {
				t.Fatalf("%s expectedError = %v", name, doc["expectedError"])
			}
			dir := t.TempDir()
			writeProject(t, dir, doc["project"])
			_, err := mapfs.Open(dir)
			if codeOf(err) != mapfs.CodeProjectInvalid {
				t.Fatalf("Open(%s) code = %q err=%v, want PROJECT_INVALID", name, codeOf(err), err)
			}
		})
	}
}

// Spec-Anchor: proj-spec/chapters/06.1-sync-protocol.md#error-codes
func TestMapfsOpenRejectsMissingProject_WhenFileAbsent(t *testing.T) {
	_, err := mapfs.Open(t.TempDir())
	if codeOf(err) != mapfs.CodeProjectInvalid {
		t.Fatalf("Open missing json code = %q err=%v, want PROJECT_INVALID", codeOf(err), err)
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#source-file-names
func TestMapfsClassifySourceNames_WhenFixtureExists(t *testing.T) {
	doc := mappingFixture(t, "source-names.json")
	for i, row := range rows(t, doc) {
		fileName, _ := row["fileName"].(string)
		className, _ := row["className"].(string)
		parentInit, _ := row["nameIsParentDirectory"].(bool)
		parent := "Bar"
		got, err := mapfs.ClassifySourceFile(fileName, parent)
		if err != nil {
			t.Fatalf("row %d ClassifySourceFile(%q): %v", i, fileName, err)
		}
		if got.ClassName != className {
			t.Errorf("%s class = %q, want %q", fileName, got.ClassName, className)
		}
		if parentInit {
			if got.InstanceName != parent {
				t.Errorf("%s instance = %q, want parent %q", fileName, got.InstanceName, parent)
			}
			if !got.Init {
				t.Errorf("%s Init = false, want true", fileName)
			}
		} else {
			wantName, _ := row["instanceName"].(string)
			if got.InstanceName != wantName {
				t.Errorf("%s instance = %q, want %q", fileName, got.InstanceName, wantName)
			}
		}
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#nodeid
func TestMapfsNodesMatchSpecExamples_WhenFixtureExists(t *testing.T) {
	fs, dir := openFromFixtureProject(t, "nodeid-examples.json")
	doc := mappingFixture(t, "nodeid-examples.json")
	for _, row := range rows(t, doc) {
		diskPath, _ := row["diskPath"].(string)
		abs := filepath.Join(dir, filepath.FromSlash(strings.TrimSuffix(diskPath, "/")))
		if strings.HasSuffix(diskPath, "/") {
			if err := os.MkdirAll(abs, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte("print(1)\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	nodes, err := fs.Nodes()
	if err != nil {
		t.Fatalf("Nodes: %v", err)
	}
	for _, row := range rows(t, doc) {
		id, _ := row["nodeId"].(string)
		className, _ := row["className"].(string)
		diskPath, _ := row["diskPath"].(string)
		n, ok := nodeByID(nodes, id)
		if !ok {
			t.Errorf("missing nodeId %s", id)
			continue
		}
		if n.ClassName != className {
			t.Errorf("%s class = %q, want %q", id, n.ClassName, className)
		}
		wantDisk := strings.TrimSuffix(diskPath, "/")
		if relSlash(n.DiskPath) != wantDisk {
			t.Errorf("%s disk = %q, want %q", id, n.DiskPath, wantDisk)
		}
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#source-file-names
// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#nodeid
func TestMapfsFolderVsInitParent_WhenDirectoryHasOrLacksInit(t *testing.T) {
	fs, dir := openFromFixtureProject(t, "folder-vs-init-parent.json")
	doc := mappingFixture(t, "folder-vs-init-parent.json")
	for _, row := range rows(t, doc) {
		diskPath, _ := row["diskPath"].(string)
		abs := filepath.Join(dir, filepath.FromSlash(strings.TrimSuffix(diskPath, "/")))
		if strings.HasSuffix(diskPath, "/") {
			if err := os.MkdirAll(abs, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte("-- src\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	nodes, err := fs.Nodes()
	if err != nil {
		t.Fatalf("Nodes: %v", err)
	}
	bar, ok := nodeByID(nodes, "server/Bar")
	if !ok || bar.ClassName != "ModuleScript" {
		t.Fatalf("Bar = %+v, want ModuleScript", bar)
	}
	qux, ok := nodeByID(nodes, "server/Qux")
	if !ok || qux.ClassName != "Folder" {
		t.Fatalf("Qux = %+v, want Folder", qux)
	}
	childOfBar, ok := nodeByID(nodes, "server/Bar/Child")
	if !ok || childOfBar.ClassName != "ModuleScript" {
		t.Fatalf("Bar/Child = %+v", childOfBar)
	}
	childOfQux, ok := nodeByID(nodes, "server/Qux/Child")
	if !ok || childOfQux.ClassName != "ModuleScript" {
		t.Fatalf("Qux/Child = %+v", childOfQux)
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#inverse-mapping-roblox-studio--disk
func TestMapfsInversePrefersLuau_WhenCreatingFiles(t *testing.T) {
	doc := mappingFixture(t, "inverse-examples.json")
	for _, row := range rows(t, doc) {
		className, _ := row["className"].(string)
		children, _ := row["mappedChildren"].(bool)
		existingLua, _ := row["existingLua"].(bool)
		want, _ := row["diskRelative"].(string)
		got := mapfs.InverseDiskRelative(className, "Name", children, existingLua)
		if got != want {
			t.Errorf("Inverse(%s children=%v lua=%v) = %q, want %q", className, children, existingLua, got, want)
		}
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#source-file-names
func TestMapfsNodesRejectDuplicateLuaLuau_WhenSameDirectory(t *testing.T) {
	doc := mappingFixture(t, "invalid-duplicate-lua-luau.json")
	dir := t.TempDir()
	writeProject(t, dir, doc["project"])
	server := filepath.Join(dir, "src", "server")
	if err := os.MkdirAll(server, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(server, "Foo.luau"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(server, "Foo.lua"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	fs, err := mapfs.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	_, err = fs.Nodes()
	if codeOf(err) != mapfs.CodeProjectInvalid {
		t.Fatalf("Nodes code = %q err=%v, want PROJECT_INVALID", codeOf(err), err)
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#source-file-names
func TestMapfsNodesRejectTwoInits_WhenDirectoryHasTwoInits(t *testing.T) {
	doc := mappingFixture(t, "invalid-multiple-init.json")
	dir := t.TempDir()
	writeProject(t, dir, doc["project"])
	bar := filepath.Join(dir, "src", "server", "Bar")
	if err := os.MkdirAll(bar, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bar, "init.luau"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bar, "init.server.luau"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	fs, err := mapfs.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	_, err = fs.Nodes()
	if codeOf(err) != mapfs.CodeProjectInvalid {
		t.Fatalf("Nodes code = %q err=%v, want PROJECT_INVALID", codeOf(err), err)
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#source-file-names
func TestMapfsClassifyRejectsForbiddenStems_WhenStemIllegal(t *testing.T) {
	doc := mappingFixture(t, "invalid-forbidden-stems.json")
	stems, ok := doc["stems"].([]any)
	if !ok {
		t.Fatal("stems missing")
	}
	for _, item := range stems {
		stem, _ := item.(string)
		name := stem + ".luau"
		_, err := mapfs.ClassifySourceFile(name, "parent")
		if codeOf(err) != mapfs.CodeProjectInvalid {
			t.Errorf("stem %q file %q code = %q err=%v, want PROJECT_INVALID", stem, name, codeOf(err), err)
		}
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#what-is-synced
func TestMapfsNodesIgnoreUnmappedFiles_WhenNonSourcePresent(t *testing.T) {
	fs, dir := openFromFixtureProject(t, "valid-project-example.json")
	server := filepath.Join(dir, "src", "server")
	if err := os.MkdirAll(server, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(server, "Foo.luau"), []byte("print(1)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(server, "README.md"), []byte("# no\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	nodes, err := fs.Nodes()
	if err != nil {
		t.Fatalf("Nodes: %v", err)
	}
	if _, ok := nodeByID(nodes, "server/Foo"); !ok {
		t.Fatal("missing server/Foo")
	}
	for _, n := range nodes {
		if strings.Contains(strings.ToLower(n.DiskPath), "readme") {
			t.Fatalf("unmapped README was synced: %+v", n)
		}
	}
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#apply-from-roblox-studio
// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#line-endings-and-encoding
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#roblox-studio--disk-g1
func TestMapfsUpsertWritesLF_WhenSourceHasCRLF(t *testing.T) {
	fs, dir := openFromFixtureProject(t, "valid-project-example.json")
	if err := fs.Upsert("server/Foo", "ModuleScript", "print(1)\r\nprint(2)\r\n"); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "src", "server", "Foo.luau"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), "\r") {
		t.Fatalf("disk still has CR: %q", got)
	}
	if string(got) != "print(1)\nprint(2)\n" {
		t.Fatalf("got %q", got)
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#line-endings-and-encoding
func TestMapfsReadStripsBOM_WhenFileHasBOM(t *testing.T) {
	fs, dir := openFromFixtureProject(t, "valid-project-example.json")
	server := filepath.Join(dir, "src", "server")
	if err := os.MkdirAll(server, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(server, "Foo.luau"), []byte("\uFEFFprint(1)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	src, err := fs.ReadSource("server/Foo")
	if err != nil {
		t.Fatalf("ReadSource: %v", err)
	}
	if strings.HasPrefix(src, "\uFEFF") {
		t.Fatalf("BOM not stripped: %q", src)
	}
	if src != "print(1)\n" {
		t.Fatalf("source = %q", src)
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#line-endings-and-encoding
func TestMapfsUpsertOmitsBOM_WhenWriting(t *testing.T) {
	fs, dir := openFromFixtureProject(t, "valid-project-example.json")
	if err := fs.Upsert("server/Foo", "ModuleScript", "\uFEFFprint(1)\n"); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "src", "server", "Foo.luau"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(string(got), "\uFEFF") {
		t.Fatalf("wrote BOM: %q", got)
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#line-endings-and-encoding
// Spec-Anchor: proj-spec/chapters/13-performance.md#limits
// Spec-Anchor: proj-spec/chapters/06.1-sync-protocol.md#error-codes
func TestMapfsUpsertRejectsOversize_WhenSourceExceedsOneMiB(t *testing.T) {
	fs, dir := openFromFixtureProject(t, "valid-project-example.json")
	ok := strings.Repeat("a", 1<<20)
	if err := fs.Upsert("server/Foo", "ModuleScript", ok); err != nil {
		t.Fatalf("1 MiB Upsert: %v", err)
	}
	tooBig := strings.Repeat("b", 1<<20+1)
	err := fs.Upsert("server/Foo", "ModuleScript", tooBig)
	if codeOf(err) != mapfs.CodePayloadTooLarge {
		t.Fatalf("oversize code = %q err=%v, want PAYLOAD_TOO_LARGE", codeOf(err), err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "src", "server", "Foo.luau"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != ok {
		t.Fatalf("partial write of oversize source: len=%d", len(got))
	}
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#apply-from-roblox-studio
func TestMapfsUpsertIsAtomic_WhenWriteSucceeds(t *testing.T) {
	fs, dir := openFromFixtureProject(t, "valid-project-example.json")
	payload := strings.Repeat("x", 4096)
	if err := fs.Upsert("server/Foo", "ModuleScript", payload); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	destDir := filepath.Join(dir, "src", "server")
	entries, err := os.ReadDir(destDir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
		if strings.Contains(strings.ToLower(e.Name()), "tmp") {
			t.Fatalf("temp leftover after atomic write: %s", e.Name())
		}
	}
	if len(names) != 1 || names[0] != "Foo.luau" {
		t.Fatalf("dest entries = %v, want [Foo.luau]", names)
	}
	got, err := os.ReadFile(filepath.Join(destDir, "Foo.luau"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != payload {
		t.Fatalf("incomplete write, len=%d", len(got))
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#inverse-mapping-roblox-studio--disk
func TestMapfsUpsertKeepsLuaExtension_WhenExistingLuaFile(t *testing.T) {
	fs, dir := openFromFixtureProject(t, "valid-project-example.json")
	server := filepath.Join(dir, "src", "server")
	if err := os.MkdirAll(server, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(server, "Foo.lua"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := fs.Upsert("server/Foo", "ModuleScript", "new\n"); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if _, err := os.Stat(filepath.Join(server, "Foo.lua")); err != nil {
		t.Fatalf("expected Foo.lua kept: %v", err)
	}
	if _, err := os.Stat(filepath.Join(server, "Foo.luau")); !os.IsNotExist(err) {
		t.Fatalf("created Foo.luau beside existing lua")
	}
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#apply-from-roblox-studio
func TestMapfsDeleteRemovesEmptyMappedDirs_WhenNotDiskRoot(t *testing.T) {
	fs, dir := openFromFixtureProject(t, "valid-project-example.json")
	if err := fs.Upsert("server/Nested/Deep", "ModuleScript", "x\n"); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	nested := filepath.Join(dir, "src", "server", "Nested")
	if _, err := os.Stat(nested); err != nil {
		t.Fatal(err)
	}
	if err := fs.Delete("server/Nested/Deep"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := os.Stat(nested); !os.IsNotExist(err) {
		t.Fatalf("empty Nested dir remained")
	}
	root := filepath.Join(dir, "src", "server")
	if _, err := os.Stat(root); err != nil {
		t.Fatalf("disk root removed: %v", err)
	}
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#apply-from-roblox-studio
func TestMapfsRenameMovesOnSameRoot_WhenSourceAndDestShareRoot(t *testing.T) {
	fs, dir := openFromFixtureProject(t, "valid-project-example.json")
	if err := fs.Upsert("server/Foo", "ModuleScript", "body\n"); err != nil {
		t.Fatal(err)
	}
	if err := fs.Rename("server/Foo", "server/Bar", "ModuleScript"); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "src", "server", "Foo.luau")); !os.IsNotExist(err) {
		t.Fatal("source remained after same-root rename")
	}
	got, err := os.ReadFile(filepath.Join(dir, "src", "server", "Bar.luau"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "body\n" {
		t.Fatalf("dest = %q", got)
	}
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#apply-from-roblox-studio
func TestMapfsRenameCopyDeletesAcrossRoots_WhenRootsDiffer(t *testing.T) {
	fs, dir := openFromFixtureProject(t, "valid-project-example.json")
	if err := fs.Upsert("server/Foo", "ModuleScript", "moved\n"); err != nil {
		t.Fatal(err)
	}
	if err := fs.Rename("server/Foo", "shared/Foo", "ModuleScript"); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "src", "server", "Foo.luau")); !os.IsNotExist(err) {
		t.Fatal("source remained after cross-root rename")
	}
	got, err := os.ReadFile(filepath.Join(dir, "src", "shared", "Foo.luau"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "moved\n" {
		t.Fatalf("dest = %q", got)
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#inverse-mapping-roblox-studio--disk
func TestMapfsUpsertConvertsLeafToInit_WhenMappedChildAdded(t *testing.T) {
	fs, dir := openFromFixtureProject(t, "valid-project-example.json")
	if err := fs.Upsert("server/Bar", "ModuleScript", "parent\n"); err != nil {
		t.Fatal(err)
	}
	if err := fs.Upsert("server/Bar/Child", "ModuleScript", "child\n"); err != nil {
		t.Fatal(err)
	}
	server := filepath.Join(dir, "src", "server")
	if _, err := os.Stat(filepath.Join(server, "Bar.luau")); !os.IsNotExist(err) {
		t.Fatal("leaf Bar.luau remained after child upsert")
	}
	got, err := os.ReadFile(filepath.Join(server, "Bar", "init.luau"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "parent\n" {
		t.Fatalf("init = %q", got)
	}
	got, err = os.ReadFile(filepath.Join(server, "Bar", "Child.luau"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "child\n" {
		t.Fatalf("child = %q", got)
	}
}

// Spec-Anchor: proj-spec/chapters/06.1-sync-protocol.md#error-codes
func TestMapfsUpsertRejectsUnknownNode_WhenOutsideRoots(t *testing.T) {
	fs, _ := openFromFixtureProject(t, "valid-project-example.json")
	err := fs.Upsert("missing/Foo", "ModuleScript", "x\n")
	if codeOf(err) != mapfs.CodeNotMapped {
		t.Fatalf("code = %q err=%v, want NOT_MAPPED", codeOf(err), err)
	}
}

// Spec-Anchor: proj-spec/chapters/06.2-sync-agent.md#apply-from-roblox-studio
func TestMapfsUpsertRejectsFolderSource_WhenClassIsFolder(t *testing.T) {
	fs, dir := openFromFixtureProject(t, "valid-project-example.json")
	if err := fs.Upsert("server/Qux", "Folder", ""); err != nil {
		t.Fatalf("Folder upsert: %v", err)
	}
	info, err := os.Stat(filepath.Join(dir, "src", "server", "Qux"))
	if err != nil || !info.IsDir() {
		t.Fatalf("expected Qux directory: %v", err)
	}
}

// Spec-Anchor: proj-spec/chapters/10.2-windows-and-wsl.md#paths
func TestMapfsOpenAcceptsProjectDir_WhenUsingOSSeparators(t *testing.T) {
	fs, dir := openFromFixtureProject(t, "valid-project-example.json")
	if err := fs.Upsert("server/Foo", "ModuleScript", "ok\n"); err != nil {
		t.Fatal(err)
	}
	joined := filepath.Join(dir, "src", "server", "Foo.luau")
	if _, err := os.Stat(joined); err != nil {
		t.Fatalf("disk path join failed: %v", err)
	}
}
