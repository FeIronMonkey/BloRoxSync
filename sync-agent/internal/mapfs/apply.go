package mapfs

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (fs *FS) Upsert(nodeID, className, source string) error {
	r, segments, err := fs.parseMapped(nodeID)
	if err != nil {
		return err
	}
	if className == "Folder" {
		abs := filepath.Join(fs.rootAbs(r), filepath.Join(segments...))
		if err := os.MkdirAll(abs, 0o755); err != nil {
			return ioError(err)
		}
		return nil
	}
	source = normalizeSource(source)
	if len(source) > maxSourceBytes {
		return &Error{Code: CodePayloadTooLarge, Message: "source exceeds 1 MiB"}
	}
	if err := fs.ensureParents(r, segments[:len(segments)-1]); err != nil {
		return err
	}
	dest, err := fs.destFile(r, segments, className)
	if err != nil {
		return err
	}
	return atomicWrite(dest, source)
}

func (fs *FS) Delete(nodeID string) error {
	n, err := fs.node(nodeID)
	if err != nil {
		return err
	}
	abs := filepath.Join(fs.projectDir, filepath.FromSlash(n.DiskPath))
	info, err := os.Stat(abs)
	if err != nil {
		return ioError(err)
	}
	if info.IsDir() {
		if err := os.RemoveAll(abs); err != nil {
			return ioError(err)
		}
	} else if err := os.Remove(abs); err != nil {
		return ioError(err)
	}
	return fs.pruneEmpty(filepath.Dir(abs))
}

func (fs *FS) Rename(oldNodeID, newNodeID, className string) error {
	oldNode, err := fs.node(oldNodeID)
	if err != nil {
		return err
	}
	oldAbs := filepath.Join(fs.projectDir, filepath.FromSlash(oldNode.DiskPath))
	oldRootID, _, err := splitNodeID(oldNodeID)
	if err != nil {
		return err
	}
	newRoot, newSegs, err := fs.parseMapped(newNodeID)
	if err != nil {
		return err
	}
	if err := fs.ensureParents(newRoot, newSegs[:len(newSegs)-1]); err != nil {
		return err
	}
	dest, err := fs.destFile(newRoot, newSegs, className)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return ioError(err)
	}
	if oldRootID == newRoot.ID {
		if err := replaceRename(oldAbs, dest); err != nil {
			return ioError(err)
		}
		return fs.pruneEmpty(filepath.Dir(oldAbs))
	}
	if err := copyFile(oldAbs, dest); err != nil {
		return err
	}
	if err := os.Remove(oldAbs); err != nil {
		return ioError(err)
	}
	return fs.pruneEmpty(filepath.Dir(oldAbs))
}

func (fs *FS) parseMapped(nodeID string) (root, []string, error) {
	rootID, rest, err := splitNodeID(nodeID)
	if err != nil {
		return root{}, nil, err
	}
	r, ok := fs.rootByID(rootID)
	if !ok {
		return root{}, nil, notMapped(nodeID)
	}
	return r, splitRel(rest), nil
}

func (fs *FS) destFile(r root, segments []string, className string) (string, error) {
	parentAbs := fs.rootAbs(r)
	if len(segments) > 1 {
		parentAbs = filepath.Join(parentAbs, filepath.Join(segments[:len(segments)-1]...))
	}
	name := segments[len(segments)-1]
	children := mappedChildrenPresent(parentAbs, name)
	existingLua := findExistingLua(parentAbs, name, className)
	rel := InverseDiskRelative(className, name, children, existingLua)
	return filepath.Join(parentAbs, filepath.FromSlash(strings.TrimSuffix(rel, "/"))), nil
}

func (fs *FS) ensureParents(r root, prefixes []string) error {
	parentAbs := fs.rootAbs(r)
	for _, seg := range prefixes {
		if err := convertLeafToInit(parentAbs, seg); err != nil {
			return err
		}
		parentAbs = filepath.Join(parentAbs, seg)
		if err := os.MkdirAll(parentAbs, 0o755); err != nil {
			return ioError(err)
		}
	}
	return nil
}

func convertLeafToInit(parentAbs, name string) error {
	type cand struct {
		file  string
		class string
		lua   bool
	}
	candidates := []cand{
		{name + ".luau", "ModuleScript", false},
		{name + ".lua", "ModuleScript", true},
		{name + ".server.luau", "Script", false},
		{name + ".server.lua", "Script", true},
		{name + ".client.luau", "LocalScript", false},
		{name + ".client.lua", "LocalScript", true},
	}
	for _, c := range candidates {
		p := filepath.Join(parentAbs, c.file)
		st, err := os.Stat(p)
		if err != nil || st.IsDir() {
			continue
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return ioError(err)
		}
		rel := InverseDiskRelative(c.class, name, true, c.lua)
		dest := filepath.Join(parentAbs, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return ioError(err)
		}
		if err := atomicWrite(dest, string(data)); err != nil {
			return err
		}
		if err := os.Remove(p); err != nil {
			return ioError(err)
		}
		return nil
	}
	return nil
}

func atomicWrite(path string, body string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return ioError(err)
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".rbx-sync-*")
	if err != nil {
		return ioError(err)
	}
	tmp := f.Name()
	if _, err := io.WriteString(f, body); err != nil {
		f.Close()
		_ = os.Remove(tmp)
		return ioError(err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return ioError(err)
	}
	if err := commitAtomic(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return ioError(err)
	}
	return nil
}

// commitAtomic is replaceRename in production. Tests swap it to observe temp+rename.
var commitAtomic = replaceRename

func replaceRename(src, dest string) error {
	if err := os.Rename(src, dest); err == nil {
		return nil
	}
	_ = os.Remove(dest)
	return os.Rename(src, dest)
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return ioError(err)
	}
	defer in.Close()
	data, err := io.ReadAll(in)
	if err != nil {
		return ioError(err)
	}
	return atomicWrite(dest, string(data))
}
