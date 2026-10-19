package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const draft202012 = "https://json-schema.org/draft/2020-12/schema"

var schemaFiles = []string{
	"project.schema.json",
	"http.schema.json",
	"events.schema.json",
}

var errorCodes = []struct {
	code   string
	status int
}{
	{"UNAUTHORIZED", 401},
	{"PROTOCOL_MISMATCH", 409},
	{"PROJECT_INVALID", 400},
	{"BAD_REQUEST", 400},
	{"NOT_FOUND", 404},
	{"NOT_MAPPED", 400},
	{"CONFLICT", 409},
	{"SESSION_REPLACED", 409},
	{"IO_ERROR", 500},
	{"PAYLOAD_TOO_LARGE", 413},
	{"INTERNAL", 500},
}

var camelKey = regexp.MustCompile(`^[a-z][A-Za-z0-9]*$`)

func protocolRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(file)
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "protocolVersion.txt")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("could not locate sync-protocol root")
	return ""
}

func schemaDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(protocolRoot(t), "schema", "v1")
}

func httpFixtureDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(protocolRoot(t), "fixtures", "v1", "http")
}

func readJSONFile(t *testing.T, path string) any {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	doc, err := jsonschema.UnmarshalJSON(f)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return doc
}

func compileSchema(t *testing.T, path, fragment string) *jsonschema.Schema {
	t.Helper()
	doc := readJSONFile(t, path)
	url := filepath.ToSlash(path)
	c := jsonschema.NewCompiler()
	if err := c.AddResource(url, doc); err != nil {
		t.Fatalf("add schema %s: %v", url, err)
	}
	loc := url
	if fragment != "" {
		loc = url + fragment
	}
	sch, err := c.Compile(loc)
	if err != nil {
		t.Fatalf("compile %s: %v", loc, err)
	}
	return sch
}

func validateFile(t *testing.T, sch *jsonschema.Schema, path string) {
	t.Helper()
	inst := readJSONFile(t, path)
	if err := sch.Validate(inst); err != nil {
		t.Fatalf("validate %s: %v", path, err)
	}
}

func validateInstance(t *testing.T, sch *jsonschema.Schema, inst any, name string) {
	t.Helper()
	if err := sch.Validate(inst); err != nil {
		t.Fatalf("validate %s: %v", name, err)
	}
}

func mustReject(t *testing.T, sch *jsonschema.Schema, inst any, name string) {
	t.Helper()
	if err := sch.Validate(inst); err == nil {
		t.Fatalf("expected %s to be rejected", name)
	}
}

func walkObjectKeys(t *testing.T, v any, onKey func(key string)) {
	t.Helper()
	switch n := v.(type) {
	case map[string]any:
		for k, child := range n {
			onKey(k)
			walkObjectKeys(t, child, onKey)
		}
	case []any:
		for _, child := range n {
			walkObjectKeys(t, child, onKey)
		}
	}
}

func propertyKeysFromSchema(t *testing.T, v any, onKey func(key string)) {
	t.Helper()
	m, ok := v.(map[string]any)
	if !ok {
		if arr, ok := v.([]any); ok {
			for _, child := range arr {
				propertyKeysFromSchema(t, child, onKey)
			}
		}
		return
	}
	if props, ok := m["properties"].(map[string]any); ok {
		for k, child := range props {
			onKey(k)
			propertyKeysFromSchema(t, child, onKey)
		}
	}
	for k, child := range m {
		if k == "properties" {
			continue
		}
		propertyKeysFromSchema(t, child, onKey)
	}
}

// Spec-Anchor: proj-spec/chapters/06.1-sync-protocol.md#protocolversion
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#protocol-and-mapping
func TestProtocolRecordsVersion1_WhenProtocolVersionFileExists(t *testing.T) {
	path := filepath.Join(protocolRoot(t), "protocolVersion.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read protocolVersion.txt: %v", err)
	}
	got := strings.TrimSpace(string(data))
	if got != "1" {
		t.Fatalf("protocolVersion.txt = %q, want 1", got)
	}
}

// Spec-Anchor: proj-spec/chapters/04-stack.md#chosen-stack
// Spec-Anchor: proj-spec/chapters/05-file-tree.md#sync-protocol-target-after-implementation
func TestSchemaExistsAsDraft202012_WhenV1SchemaFilesPresent(t *testing.T) {
	dir := schemaDir(t)
	for _, name := range schemaFiles {
		path := filepath.Join(dir, name)
		doc := readJSONFile(t, path)
		obj, ok := doc.(map[string]any)
		if !ok {
			t.Fatalf("%s is not a JSON object", name)
		}
		got, _ := obj["$schema"].(string)
		if got != draft202012 {
			t.Errorf("%s $schema = %q, want %s", name, got, draft202012)
		}
		compileSchema(t, path, "")
	}
}

// Spec-Anchor: proj-spec/chapters/06.1-sync-protocol.md#error-codes
// Spec-Anchor: proj-spec/chapters/06.1-sync-protocol.md#error-envelope
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#protocol-and-mapping
func TestErrorGoldensCoverEveryCode_WhenHttpFixturesExist(t *testing.T) {
	dir := httpFixtureDir(t)
	httpSchema := filepath.Join(schemaDir(t), "http.schema.json")
	sch := compileSchema(t, httpSchema, "#/$defs/httpErrorGolden")
	seen := map[string]int{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read http fixtures: %v", err)
	}
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasPrefix(ent.Name(), "error-") || !strings.HasSuffix(ent.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, ent.Name())
		validateFile(t, sch, path)
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		var golden struct {
			Status int `json:"status"`
			Body   struct {
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			} `json:"body"`
		}
		if err := json.Unmarshal(raw, &golden); err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
		if golden.Body.Error.Code == "" {
			t.Errorf("%s missing error.code", ent.Name())
			continue
		}
		seen[golden.Body.Error.Code] = golden.Status
	}
	for _, row := range errorCodes {
		status, ok := seen[row.code]
		if !ok {
			t.Errorf("no http golden for error code %s", row.code)
			continue
		}
		if status != row.status {
			t.Errorf("code %s status = %d, want %d", row.code, status, row.status)
		}
	}
}

// Spec-Anchor: proj-spec/chapters/06.1-sync-protocol.md#error-codes
// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#routes
func TestErrorGoldenUsesNotFound404_WhenUnlistedPath(t *testing.T) {
	path := filepath.Join(httpFixtureDir(t), "error-NOT_FOUND.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read unlisted-path golden: %v", err)
	}
	var golden struct {
		Status int `json:"status"`
		Body   struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		} `json:"body"`
	}
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	if golden.Status != 404 {
		t.Errorf("unlisted-path status = %d, want 404", golden.Status)
	}
	if golden.Body.Error.Code != "NOT_FOUND" {
		t.Errorf("unlisted-path code = %q, want NOT_FOUND", golden.Body.Error.Code)
	}
	if golden.Body.Error.Message != "not found" {
		t.Errorf("unlisted-path message = %q, want %q", golden.Body.Error.Message, "not found")
	}
}

// Spec-Anchor: proj-spec/chapters/08-domain-model.md#node
// Spec-Anchor: proj-spec/chapters/04-stack.md#json-conventions
func TestNodeShapeValidates_WhenSnapshotFieldsPresent(t *testing.T) {
	httpSchema := filepath.Join(schemaDir(t), "http.schema.json")
	sch := compileSchema(t, httpSchema, "#/$defs/node")
	dir := httpFixtureDir(t)
	validateFile(t, sch, filepath.Join(dir, "node-script.json"))
	validateFile(t, sch, filepath.Join(dir, "node-folder.json"))
	mustReject(t, sch, map[string]any{
		"nodeId":    "server/Foo",
		"className": "Script",
	}, "script without source")
	mustReject(t, sch, map[string]any{
		"nodeId":    "server/Qux",
		"className": "Folder",
		"source":    "nope",
	}, "folder with source")
}

// Spec-Anchor: proj-spec/chapters/08-domain-model.md#change
// Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1apply
func TestChangeShapeValidates_WhenApplyOpPresent(t *testing.T) {
	httpSchema := filepath.Join(schemaDir(t), "http.schema.json")
	sch := compileSchema(t, httpSchema, "#/$defs/change")
	validateFile(t, sch, filepath.Join(httpFixtureDir(t), "change-upsert.json"))
	mustReject(t, sch, map[string]any{
		"op":           "upsert",
		"nodeId":       "server/Dir",
		"className":    "Folder",
		"source":       "",
		"baseRevision": 0,
	}, "folder upsert with source")
}

// Spec-Anchor: proj-spec/chapters/08-domain-model.md#event
// Spec-Anchor: proj-spec/chapters/07.2-event-stream.md#event-object
func TestEventShapeValidates_WhenDiskOriginChange(t *testing.T) {
	eventsSchema := filepath.Join(schemaDir(t), "events.schema.json")
	sch := compileSchema(t, eventsSchema, "")
	validateFile(t, sch, filepath.Join(httpFixtureDir(t), "event-upsert.json"))
}

// Spec-Anchor: proj-spec/chapters/08-domain-model.md#root-handshake
func TestRootShapeOmitsDisk_WhenHandshakeRootsPresent(t *testing.T) {
	httpSchema := filepath.Join(schemaDir(t), "http.schema.json")
	sch := compileSchema(t, httpSchema, "#/$defs/root")
	validateFile(t, sch, filepath.Join(httpFixtureDir(t), "root-handshake.json"))
	mustReject(t, sch, map[string]any{
		"id":     "server",
		"studio": "ServerScriptService",
		"disk":   "src/server",
	}, "handshake root with disk path")
}

// Spec-Anchor: proj-spec/chapters/08-domain-model.md#project-file
// Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#project-file
func TestProjectFileShapeValidates_WhenRequiredFieldsPresent(t *testing.T) {
	projectSchema := filepath.Join(schemaDir(t), "project.schema.json")
	sch := compileSchema(t, projectSchema, "")
	validateFile(t, sch, filepath.Join(httpFixtureDir(t), "project-file.json"))
}

// Spec-Anchor: proj-spec/chapters/04-stack.md#json-conventions
func TestWireKeysAreCamelCase_WhenSchemasAndGoldensExist(t *testing.T) {
	dir := schemaDir(t)
	for _, name := range schemaFiles {
		path := filepath.Join(dir, name)
		doc := readJSONFile(t, path)
		propertyKeysFromSchema(t, doc, func(key string) {
			if !camelKey.MatchString(key) {
				t.Errorf("%s property %q is not camelCase", name, key)
			}
		})
	}
	entries, err := os.ReadDir(httpFixtureDir(t))
	if err != nil {
		t.Fatalf("read http fixtures: %v", err)
	}
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".json") {
			continue
		}
		path := filepath.Join(httpFixtureDir(t), ent.Name())
		doc := readJSONFile(t, path)
		walkObjectKeys(t, doc, func(key string) {
			if strings.HasPrefix(key, "$") {
				return
			}
			if !camelKey.MatchString(key) {
				t.Errorf("%s key %q is not camelCase", ent.Name(), key)
			}
		})
	}
}

// Spec-Anchor: proj-spec/chapters/04-stack.md#json-conventions
// Spec-Anchor: proj-spec/chapters/08-domain-model.md#node
func TestRevisionIsJsonNumber_WhenNodeAndEventGoldensExist(t *testing.T) {
	httpSchema := filepath.Join(schemaDir(t), "http.schema.json")
	nodeSch := compileSchema(t, httpSchema, "#/$defs/node")
	mustReject(t, nodeSch, map[string]any{
		"nodeId":    "server/Foo",
		"className": "Script",
		"source":    "print(1)\n",
		"revision":  "1",
	}, "revision as string")
	nodePath := filepath.Join(httpFixtureDir(t), "node-script.json")
	raw, err := os.ReadFile(nodePath)
	if err != nil {
		t.Fatalf("read node-script.json: %v", err)
	}
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()
	var node map[string]any
	if err := dec.Decode(&node); err != nil {
		t.Fatalf("decode node-script.json: %v", err)
	}
	rev, ok := node["revision"].(json.Number)
	if !ok {
		t.Fatalf("node-script.json revision type %T, want json.Number", node["revision"])
	}
	n, err := rev.Int64()
	if err != nil {
		t.Fatalf("revision is not an integer: %v", err)
	}
	if n < 0 || n >= (1<<53) {
		t.Fatalf("revision %d outside [0, 2^53)", n)
	}
}
