package mapfs

import (
	"os"
	"path/filepath"
	"strings"
)

func (fs *FS) Nodes() ([]Node, error) {
	var out []Node
	for _, r := range fs.roots {
		abs := fs.rootAbs(r)
		info, err := os.Stat(abs)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, ioError(err)
		}
		if !info.IsDir() {
			return nil, projectInvalid("disk root is not a directory")
		}
		err = filepath.WalkDir(abs, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			rel, err := filepath.Rel(abs, path)
			if err != nil {
				return err
			}
			if d.IsDir() {
				if rel == "." {
					return nil
				}
				hasInit, err := directoryHasInit(path)
				if err != nil {
					return err
				}
				if hasInit {
					return nil
				}
				out = append(out, Node{
					NodeID:    r.ID + "/" + posixJoinRel(rel),
					ClassName: "Folder",
					DiskPath:  posixRel(fs.projectDir, path),
				})
				return nil
			}
			parentName := filepath.Base(filepath.Dir(path))
			classified, err := ClassifySourceFile(d.Name(), parentName)
			if err == errNotSource {
				return nil
			}
			if err != nil {
				return err
			}
			var instanceRel string
			if classified.Init {
				if rel == d.Name() {
					instanceRel = classified.InstanceName
				} else {
					instanceRel = posixJoinRel(filepath.Dir(rel))
				}
			} else {
				dirRel := filepath.Dir(rel)
				if dirRel == "." {
					instanceRel = classified.InstanceName
				} else {
					instanceRel = posixJoinRel(dirRel) + "/" + classified.InstanceName
				}
			}
			out = append(out, Node{
				NodeID:    r.ID + "/" + instanceRel,
				ClassName: classified.ClassName,
				DiskPath:  posixRel(fs.projectDir, path),
			})
			return nil
		})
		if err != nil {
			return nil, mapWalkErr(err)
		}
		if err := validateTree(abs); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func posixJoinRel(rel string) string {
	return filepath.ToSlash(rel)
}

func mapWalkErr(err error) error {
	var e *Error
	if errorsAs(err, &e) {
		return err
	}
	return ioError(err)
}

func directoryHasInit(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, ioError(err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if _, ok := initFiles[e.Name()]; ok {
			return true, nil
		}
	}
	return false, nil
}

func validateTree(rootAbs string) error {
	return filepath.WalkDir(rootAbs, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			return nil
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return ioError(err)
		}
		initCount := 0
		seen := map[string]int{}
		parentName := filepath.Base(path)
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			classified, err := ClassifySourceFile(e.Name(), parentName)
			if err == errNotSource {
				continue
			}
			if err != nil {
				return err
			}
			if classified.Init {
				initCount++
			}
			key := classified.InstanceName + "\x00" + classified.ClassName
			seen[key]++
		}
		if initCount > 1 {
			return projectInvalid("directory has more than one init source file")
		}
		for _, n := range seen {
			if n > 1 {
				return projectInvalid("duplicate lua and luau sources")
			}
		}
		return nil
	})
}

func (fs *FS) ReadSource(nodeID string) (string, error) {
	n, err := fs.node(nodeID)
	if err != nil {
		return "", err
	}
	if n.ClassName == "Folder" {
		return "", projectInvalid("folder has no source")
	}
	raw, err := os.ReadFile(filepath.Join(fs.projectDir, filepath.FromSlash(n.DiskPath)))
	if err != nil {
		return "", ioError(err)
	}
	return stripBOM(string(raw)), nil
}

func (fs *FS) node(nodeID string) (Node, error) {
	nodes, err := fs.Nodes()
	if err != nil {
		return Node{}, err
	}
	for _, n := range nodes {
		if n.NodeID == nodeID {
			return n, nil
		}
	}
	return Node{}, notMapped(nodeID)
}

func findExistingLua(parentAbs, instanceName, className string) bool {
	rel := InverseDiskRelative(className, instanceName, false, true)
	_, err := os.Stat(filepath.Join(parentAbs, filepath.FromSlash(rel)))
	return err == nil
}

func mappedChildrenPresent(parentAbs, instanceName string) bool {
	dir := filepath.Join(parentAbs, instanceName)
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			return true
		}
		if _, err := ClassifySourceFile(e.Name(), instanceName); err == nil {
			return true
		}
	}
	return false
}

func isDiskRootAbs(fs *FS, abs string) bool {
	clean := filepath.Clean(abs)
	for _, r := range fs.roots {
		if filepath.Clean(fs.rootAbs(r)) == clean {
			return true
		}
	}
	return false
}

func (fs *FS) pruneEmpty(start string) error {
	abs := filepath.Clean(start)
	for {
		if isDiskRootAbs(fs, abs) {
			return nil
		}
		if abs == filepath.Clean(fs.projectDir) {
			return nil
		}
		entries, err := os.ReadDir(abs)
		if err != nil {
			if os.IsNotExist(err) {
				abs = filepath.Dir(abs)
				continue
			}
			return ioError(err)
		}
		if len(entries) > 0 {
			return nil
		}
		parent := filepath.Dir(abs)
		if err := os.Remove(abs); err != nil {
			return ioError(err)
		}
		abs = parent
	}
}

func splitRel(rest string) []string {
	return strings.Split(rest, "/")
}
