package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mappingFixtureDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(protocolRoot(t), "fixtures", "v1", "mapping")
}

func readMappingCase(t *testing.T, name string) map[string]any {
	t.Helper()
	path := filepath.Join(mappingFixtureDir(t), name)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read mapping fixture %s: %v", name, err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode mapping fixture %s: %v", name, err)
	}
	return doc
}

func mappingProject(t *testing.T, doc map[string]any, name string) map[string]any {
	t.Helper()
	proj, ok := doc["project"].(map[string]any)
	if !ok {
		t.Fatalf("%s missing project object", name)
	}
	return proj
}

func mappingRows(t *testing.T, doc map[string]any, name string) []map[string]any {
	t.Helper()
	raw, ok := doc["rows"].([]any)
	if !ok {
		t.Fatalf("%s missing rows array", name)
	}
	out := make([]map[string]any, 0, len(raw))
	for i, item := range raw {
		row, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("%s rows[%d] is not an object", name, i)
		}
		out = append(out, row)
	}
	return out
}

func stringField(t *testing.T, obj map[string]any, key, name string) string {
	t.Helper()
	v, ok := obj[key].(string)
	if !ok {
		t.Fatalf("%s missing string %s", name, key)
	}
	return v
}

func projectRoots(t *testing.T, proj map[string]any, name string) []map[string]any {
	t.Helper()
	raw, ok := proj["roots"].([]any)
	if !ok {
		t.Fatalf("%s project.roots missing", name)
	}
	out := make([]map[string]any, 0, len(raw))
	for i, item := range raw {
		root, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("%s roots[%d] is not an object", name, i)
		}
		out = append(out, root)
	}
	return out
}

func findRow(rows []map[string]any, key, want string) (map[string]any, bool) {
	for _, row := range rows {
		if v, ok := row[key].(string); ok && v == want {
			return row, true
		}
	}
	return nil, false
}

func findRoot(roots []map[string]any, id string) (map[string]any, bool) {
	return findRow(roots, "id", id)
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#project-file
// Spec-Anchor: proj-spec/chapters/11-examples.md#example-project-file
func TestMappingValidProjectsValidate_WhenFixturePackExists(t *testing.T) {
	sch := compileSchema(t, filepath.Join(schemaDir(t), "project.schema.json"), "")
	for _, name := range []string{"valid-project-example.json", "valid-project-demo.json"} {
		doc := readMappingCase(t, name)
		if stringField(t, doc, "kind", name) != "validProject" {
			t.Errorf("%s kind = %q, want validProject", name, doc["kind"])
		}
		proj := mappingProject(t, doc, name)
		validateInstance(t, sch, proj, name+".project")
		if name == "valid-project-example.json" {
			if stringField(t, proj, "name", name) != "example" {
				t.Errorf("%s project.name = %q, want example", name, proj["name"])
			}
			if _, ok := findRoot(projectRoots(t, proj, name), "client"); !ok {
				t.Errorf("%s missing client root from §6.4 example", name)
			}
		}
		if name == "valid-project-demo.json" {
			if stringField(t, proj, "name", name) != "demo" {
				t.Errorf("%s project.name = %q, want demo", name, proj["name"])
			}
			if _, ok := findRoot(projectRoots(t, proj, name), "client"); ok {
				t.Errorf("%s must match §11 demo (server + shared only)", name)
			}
		}
		ids := map[string]struct{}{}
		studios := map[string]struct{}{}
		disks := []string{}
		for _, root := range projectRoots(t, proj, name) {
			id := stringField(t, root, "id", name)
			if _, dup := ids[id]; dup {
				t.Errorf("%s duplicate root id %q", name, id)
			}
			ids[id] = struct{}{}
			studio := stringField(t, root, "studio", name)
			if _, dup := studios[studio]; dup {
				t.Errorf("%s duplicate studio %q", name, studio)
			}
			studios[studio] = struct{}{}
			disk := stringField(t, root, "disk", name)
			if strings.Contains(disk, "..") {
				t.Errorf("%s disk %q contains ..", name, disk)
			}
			if strings.Contains(disk, `\`) {
				t.Errorf("%s disk %q uses backslash; rbx-sync.json must use /", name, disk)
			}
			disks = append(disks, disk)
		}
		for i, a := range disks {
			for j, b := range disks {
				if i >= j {
					continue
				}
				if a == b || strings.HasPrefix(a+"/", b+"/") || strings.HasPrefix(b+"/", a+"/") {
					t.Errorf("%s overlapping disk %q and %q", name, a, b)
				}
			}
		}
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#project-file
func TestMappingProjectRejectsDotDot_WhenDiskHasParentSegment(t *testing.T) {
	name := "invalid-dotdot-disk.json"
	doc := readMappingCase(t, name)
	if stringField(t, doc, "expectedError", name) != "PROJECT_INVALID" {
		t.Errorf("%s expectedError = %q, want PROJECT_INVALID", name, doc["expectedError"])
	}
	sch := compileSchema(t, filepath.Join(schemaDir(t), "project.schema.json"), "")
	mustReject(t, sch, mappingProject(t, doc, name), name+".project")
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#project-file
func TestMappingProjectRejectsBackslashDisk_WhenSeparatorsAreNotPosix(t *testing.T) {
	name := "invalid-backslash-disk.json"
	doc := readMappingCase(t, name)
	if stringField(t, doc, "expectedError", name) != "PROJECT_INVALID" {
		t.Errorf("%s expectedError = %q, want PROJECT_INVALID", name, doc["expectedError"])
	}
	sch := compileSchema(t, filepath.Join(schemaDir(t), "project.schema.json"), "")
	mustReject(t, sch, mappingProject(t, doc, name), name+".project")
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#project-file
func TestMappingProjectRequiresUniqueStudio_WhenInvalidFixtureExists(t *testing.T) {
	name := "invalid-duplicate-studio.json"
	doc := readMappingCase(t, name)
	if stringField(t, doc, "expectedError", name) != "PROJECT_INVALID" {
		t.Errorf("%s expectedError = %q, want PROJECT_INVALID", name, doc["expectedError"])
	}
	if stringField(t, doc, "reason", name) != "duplicateStudio" {
		t.Errorf("%s reason = %q, want duplicateStudio", name, doc["reason"])
	}
	studios := map[string]int{}
	for _, root := range projectRoots(t, mappingProject(t, doc, name), name) {
		studios[stringField(t, root, "studio", name)]++
	}
	dup := false
	for _, n := range studios {
		if n > 1 {
			dup = true
		}
	}
	if !dup {
		t.Fatalf("%s does not encode duplicate studio targets", name)
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#project-file
func TestMappingProjectRequiresUniqueRootIds_WhenInvalidFixtureExists(t *testing.T) {
	name := "invalid-duplicate-root-id.json"
	doc := readMappingCase(t, name)
	if stringField(t, doc, "expectedError", name) != "PROJECT_INVALID" {
		t.Errorf("%s expectedError = %q, want PROJECT_INVALID", name, doc["expectedError"])
	}
	ids := map[string]int{}
	for _, root := range projectRoots(t, mappingProject(t, doc, name), name) {
		ids[stringField(t, root, "id", name)]++
	}
	dup := false
	for _, n := range ids {
		if n > 1 {
			dup = true
		}
	}
	if !dup {
		t.Fatalf("%s does not encode duplicate root ids", name)
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#project-file
func TestMappingProjectRequiresNonOverlappingDisk_WhenInvalidFixtureExists(t *testing.T) {
	name := "invalid-overlapping-disk.json"
	doc := readMappingCase(t, name)
	if stringField(t, doc, "expectedError", name) != "PROJECT_INVALID" {
		t.Errorf("%s expectedError = %q, want PROJECT_INVALID", name, doc["expectedError"])
	}
	if stringField(t, doc, "reason", name) != "overlappingDisk" {
		t.Errorf("%s reason = %q, want overlappingDisk", name, doc["reason"])
	}
	disks := []string{}
	for _, root := range projectRoots(t, mappingProject(t, doc, name), name) {
		disks = append(disks, stringField(t, root, "disk", name))
	}
	overlap := false
	for i, a := range disks {
		for j, b := range disks {
			if i >= j {
				continue
			}
			if a == b || strings.HasPrefix(a+"/", b+"/") || strings.HasPrefix(b+"/", a+"/") {
				overlap = true
			}
		}
	}
	if !overlap {
		t.Fatalf("%s does not encode overlapping disk directories", name)
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#source-file-names
// Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d9--independent-mapping-names-server--client--init
func TestMappingSourceNamesEncodeClassTable_WhenFixtureExists(t *testing.T) {
	name := "source-names.json"
	doc := readMappingCase(t, name)
	rows := mappingRows(t, doc, name)
	want := []struct {
		fileName       string
		instanceName   string
		className      string
		nameIsParent   bool
	}{
		{"Foo.luau", "Foo", "ModuleScript", false},
		{"Foo.lua", "Foo", "ModuleScript", false},
		{"Foo.server.luau", "Foo", "Script", false},
		{"Foo.server.lua", "Foo", "Script", false},
		{"Foo.client.luau", "Foo", "LocalScript", false},
		{"Foo.client.lua", "Foo", "LocalScript", false},
		{"init.luau", "", "ModuleScript", true},
		{"init.lua", "", "ModuleScript", true},
		{"init.server.luau", "", "Script", true},
		{"init.server.lua", "", "Script", true},
		{"init.client.luau", "", "LocalScript", true},
		{"init.client.lua", "", "LocalScript", true},
	}
	for _, row := range want {
		got, ok := findRow(rows, "fileName", row.fileName)
		if !ok {
			t.Errorf("missing source-name row for %s", row.fileName)
			continue
		}
		if stringField(t, got, "className", row.fileName) != row.className {
			t.Errorf("%s className = %q, want %q", row.fileName, got["className"], row.className)
		}
		parentName, _ := got["nameIsParentDirectory"].(bool)
		if parentName != row.nameIsParent {
			t.Errorf("%s nameIsParentDirectory = %v, want %v", row.fileName, parentName, row.nameIsParent)
		}
		if !row.nameIsParent {
			if stringField(t, got, "instanceName", row.fileName) != row.instanceName {
				t.Errorf("%s instanceName = %q, want %q", row.fileName, got["instanceName"], row.instanceName)
			}
		}
	}
	initServer, ok := findRow(rows, "fileName", "init.server.luau")
	if ok {
		if stringField(t, initServer, "className", "init.server.luau") != "Script" {
			t.Errorf("init.server.luau must be Script, not ModuleScript named init.server")
		}
		if v, _ := initServer["instanceName"].(string); v == "init.server" {
			t.Errorf("init.server.luau instanceName must be parent directory, not init.server")
		}
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#nodeid
// Spec-Anchor: proj-spec/chapters/11-examples.md#example-mapping
func TestMappingNodeIdsMatchSpecExamples_WhenFixtureExists(t *testing.T) {
	name := "nodeid-examples.json"
	doc := readMappingCase(t, name)
	rows := mappingRows(t, doc, name)
	want := []struct {
		diskPath  string
		nodeId    string
		className string
	}{
		{"src/server/Foo.server.luau", "server/Foo", "Script"},
		{"src/server/Util.luau", "server/Util", "ModuleScript"},
		{"src/server/Bar/init.luau", "server/Bar", "ModuleScript"},
		{"src/server/Bar/Child.luau", "server/Bar/Child", "ModuleScript"},
		{"src/server/Qux/", "server/Qux", "Folder"},
		{"src/server/Main.server.luau", "server/Main", "Script"},
		{"src/server/Lib/init.luau", "server/Lib", "ModuleScript"},
		{"src/server/Lib/Util.luau", "server/Lib/Util", "ModuleScript"},
		{"src/shared/Config.luau", "shared/Config", "ModuleScript"},
	}
	for _, row := range want {
		got, ok := findRow(rows, "diskPath", row.diskPath)
		if !ok {
			t.Errorf("missing nodeId row for %s", row.diskPath)
			continue
		}
		if stringField(t, got, "nodeId", row.diskPath) != row.nodeId {
			t.Errorf("%s nodeId = %q, want %q", row.diskPath, got["nodeId"], row.nodeId)
		}
		if stringField(t, got, "className", row.diskPath) != row.className {
			t.Errorf("%s className = %q, want %q", row.diskPath, got["className"], row.className)
		}
		if strings.HasPrefix(stringField(t, got, "nodeId", row.diskPath), "/") {
			t.Errorf("%s nodeId has a leading slash", row.diskPath)
		}
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#nodeid
func TestMappingStudioPathWalksFromService_WhenNodeIdExamplesExist(t *testing.T) {
	name := "nodeid-examples.json"
	doc := readMappingCase(t, name)
	rows := mappingRows(t, doc, name)
	child, ok := findRow(rows, "nodeId", "server/Bar/Child")
	if !ok {
		t.Fatal("missing server/Bar/Child")
	}
	got := stringField(t, child, "studioPath", "server/Bar/Child")
	if got != "ServerScriptService.Bar.Child" {
		t.Errorf("studioPath = %q, want ServerScriptService.Bar.Child", got)
	}
	shared, ok := findRow(rows, "nodeId", "shared/Config")
	if !ok {
		t.Fatal("missing shared/Config")
	}
	got = stringField(t, shared, "studioPath", "shared/Config")
	if got != "ReplicatedStorage.Shared.Config" {
		t.Errorf("shared/Config studioPath = %q, want ReplicatedStorage.Shared.Config", got)
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#inverse-mapping-roblox-studio--disk
func TestMappingInversePrefersLuau_WhenCreatingFiles(t *testing.T) {
	name := "inverse-examples.json"
	doc := readMappingCase(t, name)
	rows := mappingRows(t, doc, name)
	want := []struct {
		className       string
		mappedChildren  bool
		existingLua     bool
		diskRelative    string
	}{
		{"ModuleScript", false, false, "Name.luau"},
		{"Script", false, false, "Name.server.luau"},
		{"LocalScript", false, false, "Name.client.luau"},
		{"ModuleScript", true, false, "Name/init.luau"},
		{"Script", true, false, "Name/init.server.luau"},
		{"LocalScript", true, false, "Name/init.client.luau"},
		{"Folder", false, false, "Name/"},
		{"ModuleScript", false, true, "Name.lua"},
	}
	for _, row := range want {
		found := false
		for _, got := range rows {
			className := stringField(t, got, "className", name)
			children, _ := got["mappedChildren"].(bool)
			existing, _ := got["existingLua"].(bool)
			if className == row.className && children == row.mappedChildren && existing == row.existingLua {
				found = true
				if stringField(t, got, "diskRelative", className) != row.diskRelative {
					t.Errorf("%s children=%v existingLua=%v diskRelative = %q, want %q",
						row.className, row.mappedChildren, row.existingLua, got["diskRelative"], row.diskRelative)
				}
			}
		}
		if !found {
			t.Errorf("missing inverse row class=%s mappedChildren=%v existingLua=%v",
				row.className, row.mappedChildren, row.existingLua)
		}
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#source-file-names
func TestMappingDuplicateLuaLuauIsProjectInvalid_WhenSameDirectory(t *testing.T) {
	name := "invalid-duplicate-lua-luau.json"
	doc := readMappingCase(t, name)
	if stringField(t, doc, "expectedError", name) != "PROJECT_INVALID" {
		t.Errorf("%s expectedError = %q, want PROJECT_INVALID", name, doc["expectedError"])
	}
	if stringField(t, doc, "reason", name) != "duplicateLuaLuau" {
		t.Errorf("%s reason = %q, want duplicateLuaLuau", name, doc["reason"])
	}
	files, ok := doc["files"].([]any)
	if !ok || len(files) != 2 {
		t.Fatalf("%s files must list Foo.luau and Foo.lua", name)
	}
	joined := ""
	for _, f := range files {
		s, _ := f.(string)
		joined += s + " "
	}
	if !strings.Contains(joined, "Foo.luau") || !strings.Contains(joined, "Foo.lua") {
		t.Errorf("%s files = %v, want Foo.luau and Foo.lua", name, files)
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#source-file-names
func TestMappingAtMostOneInit_WhenDirectoryHasTwoInits(t *testing.T) {
	name := "invalid-multiple-init.json"
	doc := readMappingCase(t, name)
	if stringField(t, doc, "expectedError", name) != "PROJECT_INVALID" {
		t.Errorf("%s expectedError = %q, want PROJECT_INVALID", name, doc["expectedError"])
	}
	if stringField(t, doc, "reason", name) != "multipleInit" {
		t.Errorf("%s reason = %q, want multipleInit", name, doc["reason"])
	}
	files, ok := doc["files"].([]any)
	if !ok || len(files) < 2 {
		t.Fatalf("%s must list two init.* files", name)
	}
	initCount := 0
	for _, f := range files {
		s, _ := f.(string)
		base := filepath.Base(s)
		if strings.HasPrefix(base, "init.") {
			initCount++
		}
	}
	if initCount < 2 {
		t.Errorf("%s files do not encode two init.* sources", name)
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#source-file-names
func TestMappingForbiddenStemsAreProjectInvalid_WhenStemIllegal(t *testing.T) {
	name := "invalid-forbidden-stems.json"
	doc := readMappingCase(t, name)
	if stringField(t, doc, "expectedError", name) != "PROJECT_INVALID" {
		t.Errorf("%s expectedError = %q, want PROJECT_INVALID", name, doc["expectedError"])
	}
	if stringField(t, doc, "reason", name) != "forbiddenStem" {
		t.Errorf("%s reason = %q, want forbiddenStem", name, doc["reason"])
	}
	stems, ok := doc["stems"].([]any)
	if !ok {
		t.Fatalf("%s missing stems array", name)
	}
	want := []string{"", ".", "..", "a/b", `a\b`, "a:b", "a*b", "a?b", `a"b`, "a<b", "a>b", "a|b"}
	got := map[string]struct{}{}
	for _, s := range stems {
		str, _ := s.(string)
		got[str] = struct{}{}
	}
	for _, stem := range want {
		if _, ok := got[stem]; !ok {
			t.Errorf("missing forbidden stem %q", stem)
		}
	}
}

// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#source-file-names
// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#nodeid
func TestMappingFolderVsInitParent_WhenDirectoryHasOrLacksInit(t *testing.T) {
	name := "folder-vs-init-parent.json"
	doc := readMappingCase(t, name)
	rows := mappingRows(t, doc, name)
	bar, ok := findRow(rows, "nodeId", "server/Bar")
	if !ok {
		t.Fatal("missing server/Bar")
	}
	if stringField(t, bar, "className", "server/Bar") != "ModuleScript" {
		t.Errorf("Bar with init.luau className = %q, want ModuleScript", bar["className"])
	}
	child, ok := findRow(rows, "nodeId", "server/Bar/Child")
	if !ok {
		t.Fatal("missing server/Bar/Child")
	}
	if stringField(t, child, "parentNodeId", "server/Bar/Child") != "server/Bar" {
		t.Errorf("Child parentNodeId = %q, want server/Bar (init instance, not Folder)", child["parentNodeId"])
	}
	qux, ok := findRow(rows, "nodeId", "server/Qux")
	if !ok {
		t.Fatal("missing server/Qux")
	}
	if stringField(t, qux, "className", "server/Qux") != "Folder" {
		t.Errorf("Qux without init className = %q, want Folder", qux["className"])
	}
	qchild, ok := findRow(rows, "nodeId", "server/Qux/Child")
	if !ok {
		t.Fatal("missing server/Qux/Child")
	}
	if stringField(t, qchild, "parentNodeId", "server/Qux/Child") != "server/Qux" {
		t.Errorf("Qux/Child parentNodeId = %q, want server/Qux Folder", qchild["parentNodeId"])
	}
}

// Impl-Only: structural — mapping pack lives under sync-protocol fixtures (G7 location)
func TestMappingFixturesLiveInProtocolRepo_WhenPackRequired(t *testing.T) {
	dir := mappingFixtureDir(t)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read mapping dir: %v", err)
	}
	n := 0
	for _, ent := range entries {
		if !ent.IsDir() && strings.HasSuffix(ent.Name(), ".json") {
			n++
		}
	}
	if n < 12 {
		t.Fatalf("mapping fixture pack has %d json files, want at least 12", n)
	}
}
