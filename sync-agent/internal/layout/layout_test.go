package layout_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

var packages = []string{"sync-protocol", "sync-agent", "studio-plugin"}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(file)
	for i := 0; i < 8; i++ {
		if isRepoRoot(dir) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("could not locate repository root from layout tests")
	return ""
}

func isRepoRoot(dir string) bool {
	if !dirExists(filepath.Join(dir, "proj-spec", "chapters")) {
		return false
	}
	if !dirExists(filepath.Join(dir, "sync-agent")) {
		return false
	}
	return true
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func stripMarkdownEmphasis(s string) string {
	return strings.ReplaceAll(s, "*", "")
}

// Spec-Anchor: proj-spec/chapters/03.1-repository-layout.md#packages
// Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d2--one-git-repository-three-packages
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#packages-and-process
func TestLayoutHasPackages_WhenRepoRootResolved(t *testing.T) {
	root := repoRoot(t)
	for _, pkg := range packages {
		path := filepath.Join(root, pkg)
		if !dirExists(path) {
			t.Errorf("package directory missing: %s", pkg)
		}
	}
}

// Spec-Anchor: proj-spec/chapters/03.1-repository-layout.md#packages
// Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d2--one-git-repository-three-packages
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#packages-and-process
func TestLayoutRejectsNestedGit_WhenPackageDirsExist(t *testing.T) {
	root := repoRoot(t)
	for _, pkg := range packages {
		nested := filepath.Join(root, pkg, ".git")
		if _, err := os.Stat(nested); err == nil {
			t.Errorf("package %s must not contain a nested .git", pkg)
		} else if !os.IsNotExist(err) {
			t.Errorf("stat %s: %v", nested, err)
		}
	}
}

// Spec-Anchor: proj-spec/chapters/03.1-repository-layout.md#packages
// Spec-Anchor: proj-spec/chapters/02-goals-and-non-goals.md#goals
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#packages-and-process
func TestLayoutRepeatsSourceBan_WhenRootAndPackageDocsExist(t *testing.T) {
	root := repoRoot(t)
	docs := []string{
		filepath.Join(root, "AGENTS.md"),
		filepath.Join(root, "README.md"),
	}
	for _, pkg := range packages {
		docs = append(docs,
			filepath.Join(root, pkg, "AGENTS.md"),
			filepath.Join(root, pkg, "README.md"),
		)
	}
	ban := regexp.MustCompile(`(?i)must\s+not\s+read`)
	for _, path := range docs {
		plain := stripMarkdownEmphasis(readFile(t, path))
		if !ban.MatchString(plain) {
			t.Errorf("%s must repeat the source ban (must not read … code)", path)
		}
		if !strings.Contains(strings.ToLower(plain), "source") {
			t.Errorf("%s must mention source in the source-ban text", path)
		}
	}
}

// Spec-Anchor: proj-spec/chapters/05-file-tree.md#this-repository-now
func TestLayoutHasMitLicense_WhenRepoRootResolved(t *testing.T) {
	root := repoRoot(t)
	body := readFile(t, filepath.Join(root, "LICENSE"))
	if !strings.Contains(body, "MIT License") {
		t.Errorf("LICENSE must be the MIT License")
	}
	if !strings.Contains(body, "Permission is hereby granted, free of charge") {
		t.Errorf("LICENSE must include the MIT grant")
	}
}

// Spec-Anchor: proj-spec/chapters/02-goals-and-non-goals.md#goals
// Spec-Anchor: proj-spec/chapters/05-file-tree.md#this-repository-now
// Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#packages-and-process
func TestLayoutListsAllowedSources_WhenReferencesExists(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "references.md")
	body := readFile(t, path)
	if !strings.Contains(body, "https://create.roblox.com") {
		t.Errorf("references.md must list official Roblox Creator Hub sources")
	}
	if !strings.Contains(strings.ToLower(body), "httpservice") {
		t.Errorf("references.md must list HttpService as an allowed research source")
	}
	plain := strings.ToLower(stripMarkdownEmphasis(body))
	if !strings.Contains(plain, "must not") && !strings.Contains(plain, "do not") {
		t.Errorf("references.md must document the source ban")
	}
}

// Spec-Anchor: proj-spec/chapters/04-stack.md#chosen-stack
// Spec-Anchor: proj-spec/chapters/05-file-tree.md#sync-agent-target-after-implementation
// Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d8--rbx-sync-language-go
func TestLayoutHasGoModule_WhenSyncAgentExists(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "sync-agent", "go.mod")
	body := readFile(t, path)
	if !regexp.MustCompile(`(?m)^module\s+\S+`).MatchString(body) {
		t.Errorf("sync-agent/go.mod must declare a module path")
	}
	if !regexp.MustCompile(`(?m)^go 1\.(2[3-9]|[3-9][0-9])`).MatchString(body) {
		t.Errorf("sync-agent/go.mod must pin Go 1.23 or newer")
	}
}

// Spec-Anchor: proj-spec/chapters/03.1-repository-layout.md#versioning
// Spec-Anchor: proj-spec/chapters/05-file-tree.md#sync-protocol-target-after-implementation
func TestLayoutRecordsProtocolVersion1_WhenProtocolPackageExists(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "sync-protocol", "protocolVersion.txt")
	body := readFile(t, path)
	trimmed := strings.TrimRight(body, "\r\n")
	if trimmed != "1" {
		t.Errorf("protocolVersion.txt must be a single line 1, got %q", body)
	}
	if strings.Count(trimmed, "\n") != 0 {
		t.Errorf("protocolVersion.txt must be a single line")
	}
}

// Spec-Anchor: proj-spec/chapters/05-file-tree.md#install-scripts-target-after-implementation
// Spec-Anchor: proj-spec/chapters/05-file-tree.md#sync-protocol-target-after-implementation
// Spec-Anchor: proj-spec/chapters/05-file-tree.md#sync-agent-target-after-implementation
// Spec-Anchor: proj-spec/chapters/05-file-tree.md#studio-plugin-target-after-implementation
func TestLayoutHasTargetDirectories_WhenScaffoldingRequired(t *testing.T) {
	root := repoRoot(t)
	dirs := []string{
		filepath.Join("sync-protocol", "schema", "v1"),
		filepath.Join("sync-protocol", "fixtures", "v1", "mapping"),
		filepath.Join("sync-protocol", "fixtures", "v1", "http"),
		filepath.Join("sync-agent", "cmd", "rbx-sync"),
		filepath.Join("sync-agent", "internal", "cli"),
		filepath.Join("sync-agent", "internal", "serve"),
		filepath.Join("sync-agent", "internal", "mapfs"),
		filepath.Join("sync-agent", "internal", "watch"),
		filepath.Join("sync-agent", "internal", "installplugin"),
		filepath.Join("studio-plugin", "src"),
		filepath.Join("studio-plugin", "tests"),
		"scripts",
	}
	for _, rel := range dirs {
		path := filepath.Join(root, rel)
		if !dirExists(path) {
			t.Errorf("target directory missing: %s", rel)
		}
	}
}

// Spec-Anchor: proj-spec/chapters/04-stack.md#cross-compilation-rbx-sync
// Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d8--rbx-sync-language-go
func TestLayoutDocumentsCgoDisabled_WhenSyncAgentReadmeExists(t *testing.T) {
	root := repoRoot(t)
	body := readFile(t, filepath.Join(root, "sync-agent", "README.md"))
	if !strings.Contains(body, "CGO_ENABLED=0") {
		t.Errorf("sync-agent/README.md must document CGO_ENABLED=0")
	}
}

// Spec-Anchor: proj-spec/chapters/12-testing.md#layers
func TestLayoutRunsLayerCommands_WhenCIWorkflowExists(t *testing.T) {
	root := repoRoot(t)
	workflowDir := filepath.Join(root, ".github", "workflows")
	entries, err := os.ReadDir(workflowDir)
	if err != nil {
		t.Fatalf("CI workflow directory: %v", err)
	}
	var combined strings.Builder
	foundYAML := false
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
			continue
		}
		foundYAML = true
		combined.WriteString(readFile(t, filepath.Join(workflowDir, name)))
		combined.WriteByte('\n')
	}
	if !foundYAML {
		t.Fatal("no CI workflow YAML under .github/workflows")
	}
	body := combined.String()
	stepName := regexp.MustCompile(`(?m)^\s*-?\s*name:\s*L([1-4])\b`)
	seen := map[string]bool{}
	for _, m := range stepName.FindAllStringSubmatch(body, -1) {
		seen["L"+m[1]] = true
	}
	for _, layer := range []string{"L1", "L2", "L3", "L4"} {
		if !seen[layer] {
			t.Errorf("CI workflows must invoke a step named %s", layer)
		}
	}
}

// Spec-Anchor: proj-spec/chapters/04-stack.md#cross-compilation-rbx-sync
// Spec-Anchor: proj-spec/chapters/05-file-tree.md#sync-agent-target-after-implementation
// Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d8--rbx-sync-language-go
func TestLayoutBuildsRbxSyncStub_WhenGoModuleExists(t *testing.T) {
	root := repoRoot(t)
	agent := filepath.Join(root, "sync-agent")
	out := filepath.Join(t.TempDir(), "rbx-sync-stub")
	cmd := exec.Command("go", "build", "-o", out, "./cmd/rbx-sync")
	cmd.Dir = agent
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build ./cmd/rbx-sync (CGO_ENABLED=0) failed: %v\n%s", err, output)
	}
}
