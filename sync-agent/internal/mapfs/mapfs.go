package mapfs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	CodeProjectInvalid  = "PROJECT_INVALID"
	CodePayloadTooLarge = "PAYLOAD_TOO_LARGE"
	CodeNotMapped       = "NOT_MAPPED"
	CodeIOError         = "IO_ERROR"
	maxSourceBytes      = 1 << 20
)

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message == "" {
		return e.Code
	}
	return e.Code + ": " + e.Message
}

func projectInvalid(msg string) error {
	return &Error{Code: CodeProjectInvalid, Message: msg}
}

func notMapped(msg string) error {
	return &Error{Code: CodeNotMapped, Message: msg}
}

func ioError(err error) error {
	if err == nil {
		return nil
	}
	var e *Error
	if errorsAs(err, &e) {
		return err
	}
	return &Error{Code: CodeIOError, Message: err.Error()}
}

func errorsAs(err error, target **Error) bool {
	if err == nil {
		return false
	}
	e, ok := err.(*Error)
	if !ok {
		return false
	}
	*target = e
	return true
}

type Node struct {
	NodeID    string
	ClassName string
	DiskPath  string
}

type Classified struct {
	InstanceName string
	ClassName    string
	Init         bool
}

type root struct {
	ID     string
	Disk   string
	Studio string
}

type StudioRoot struct {
	ID     string
	Studio string
}

type FS struct {
	projectDir string
	name       string
	roots      []root
}

type projectJSON struct {
	ProtocolVersion int    `json:"protocolVersion"`
	Name            string `json:"name"`
	Roots           []struct {
		ID     string `json:"id"`
		Disk   string `json:"disk"`
		Studio string `json:"studio"`
	} `json:"roots"`
}

var rootIDPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)

func Open(projectDir string) (*FS, error) {
	raw, err := os.ReadFile(filepath.Join(projectDir, "rbx-sync.json"))
	if err != nil {
		return nil, projectInvalid("rbx-sync.json missing or unreadable")
	}
	var doc projectJSON
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, projectInvalid("rbx-sync.json is not valid JSON")
	}
	if err := validateProject(doc); err != nil {
		return nil, err
	}
	fs := &FS{projectDir: projectDir, name: doc.Name}
	for _, r := range doc.Roots {
		fs.roots = append(fs.roots, root{ID: r.ID, Disk: r.Disk, Studio: r.Studio})
	}
	return fs, nil
}

func (fs *FS) ProjectName() string {
	return fs.name
}

func (fs *FS) StudioRoots() []StudioRoot {
	out := make([]StudioRoot, 0, len(fs.roots))
	for _, r := range fs.roots {
		out = append(out, StudioRoot{ID: r.ID, Studio: r.Studio})
	}
	return out
}

func (fs *FS) DiskRoots() []string {
	out := make([]string, 0, len(fs.roots))
	for _, r := range fs.roots {
		out = append(out, fs.rootAbs(r))
	}
	return out
}

func (fs *FS) AbsFromNode(n Node) string {
	return filepath.Join(fs.projectDir, filepath.FromSlash(n.DiskPath))
}

func (fs *FS) DiskAbs(nodeID string) (string, error) {
	n, err := fs.node(nodeID)
	if err != nil {
		return "", err
	}
	return fs.AbsFromNode(n), nil
}

func (fs *FS) PredictWriteAbs(nodeID, className string) (string, error) {
	r, segs, err := fs.parseMapped(nodeID)
	if err != nil {
		return "", err
	}
	if className == "Folder" {
		return filepath.Join(fs.rootAbs(r), filepath.Join(segs...)), nil
	}
	return fs.destFile(r, segs, className)
}

func validateProject(doc projectJSON) error {
	if doc.ProtocolVersion != 1 {
		return projectInvalid("protocolVersion must be 1")
	}
	if strings.TrimSpace(doc.Name) == "" {
		return projectInvalid("name must be non-empty")
	}
	if len(doc.Roots) == 0 {
		return projectInvalid("roots must be non-empty")
	}
	ids := map[string]struct{}{}
	studios := map[string]struct{}{}
	disks := make([]string, 0, len(doc.Roots))
	for _, r := range doc.Roots {
		if !rootIDPattern.MatchString(r.ID) {
			return projectInvalid("root id is invalid")
		}
		if _, ok := ids[r.ID]; ok {
			return projectInvalid("duplicate root id")
		}
		ids[r.ID] = struct{}{}
		if r.Studio == "" || r.Studio == "game" {
			return projectInvalid("studio path must be a service path")
		}
		if _, ok := studios[r.Studio]; ok {
			return projectInvalid("duplicate studio path")
		}
		studios[r.Studio] = struct{}{}
		if err := validateDisk(r.Disk); err != nil {
			return err
		}
		disks = append(disks, strings.Trim(r.Disk, "/"))
	}
	for i := 0; i < len(disks); i++ {
		for j := i + 1; j < len(disks); j++ {
			if diskOverlap(disks[i], disks[j]) {
				return projectInvalid("overlapping disk roots")
			}
		}
	}
	return nil
}

func validateDisk(disk string) error {
	if disk == "" || strings.Contains(disk, "\\") {
		return projectInvalid("disk paths must use POSIX separators")
	}
	if filepath.IsAbs(disk) {
		return projectInvalid("disk path must be relative")
	}
	for _, seg := range strings.Split(disk, "/") {
		if seg == ".." {
			return projectInvalid("disk path must not contain ..")
		}
	}
	return nil
}

func diskOverlap(a, b string) bool {
	if a == b {
		return true
	}
	return strings.HasPrefix(a+"/", b+"/") || strings.HasPrefix(b+"/", a+"/")
}

func (fs *FS) rootByID(id string) (root, bool) {
	for _, r := range fs.roots {
		if r.ID == id {
			return r, true
		}
	}
	return root{}, false
}

func (fs *FS) rootAbs(r root) string {
	return filepath.Join(fs.projectDir, filepath.FromSlash(r.Disk))
}

func posixRel(projectDir, abs string) string {
	rel, err := filepath.Rel(projectDir, abs)
	if err != nil {
		return filepath.ToSlash(abs)
	}
	return filepath.ToSlash(rel)
}
