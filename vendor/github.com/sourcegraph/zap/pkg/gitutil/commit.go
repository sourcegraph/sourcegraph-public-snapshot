package gitutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func makeCommitFromDir(gitRepo Worktree, onlyIfChangedFiles bool) (commitID string, err error) {
	head, err := gitRepo.HEADOrDevNullTree()
	if err != nil {
		return
	}
	commitID, _, err = gitRepo.MakeCommit(head, onlyIfChangedFiles)
	return
}

// Tree represents a git tree (i.e., "git ls-tree -r -d").
type Tree struct {
	byPath map[string]*TreeEntry // key is path without leading or trailing slashes (e.g., "d", "d1/d2", "d/f")
	Root   []*TreeEntry
}

func (t *Tree) Add(path string, e *TreeEntry) {
	if path == "" || path == "." {
		panic(fmt.Sprintf("bad path: %q", path))
	}

	if t.byPath == nil {
		t.byPath = map[string]*TreeEntry{}
	}
	if _, exists := t.byPath[path]; exists {
		panic(fmt.Sprintf("already exists: %q", path))
	}
	t.byPath[path] = e

	dir := filepath.Dir(path)
	if dir == "." {
		t.Root = append(t.Root, e)
	} else {
		parent := t.byPath[dir]
		if parent == nil {
			panic(fmt.Sprintf("no parent tree for %q", path))
		}
		parent.Entries = append(parent.Entries, e)
	}
}

// remove removes path from the tree. It does not prune trees that are
// empty after the removal.
func (t *Tree) remove(path string) {
	var parentEntries *[]*TreeEntry
	if dir := filepath.Dir(path); dir == "." {
		parentEntries = &t.Root
	} else {
		parentEntries = &t.byPath[dir].Entries
	}

	name := filepath.Base(path)
	for i, e := range *parentEntries {
		if e.Name == name {
			a := *parentEntries

			// Delete item without causing memory leak (see
			// https://github.com/golang/go/wiki/SliceTricks).
			copy(a[i:], a[i+1:])
			a[len(a)-1] = nil
			a = a[:len(a)-1]

			*parentEntries = a
			break
		}
	}
}

func (t *Tree) Get(path string) *TreeEntry {
	return t.byPath[path]
}

// createOrDirtyAncestors creates all ancestor trees of path (if they
// don't already exist). If any ancestor trees already exist, it marks
// them as having changed. Their oids are zeroed out and need to be
// recomputed later.
func (t *Tree) createOrDirtyAncestors(path string) {
	dir := filepath.Dir(path)
	if dir == "." {
		return
	}

	comps := strings.Split(dir, string(os.PathSeparator))
	var p string // ancestor path
	for i, c := range comps {
		if i == 0 {
			p = c
		} else {
			p += string(os.PathSeparator) + c
		}

		ancestor, present := t.byPath[p]
		if !present {
			// Create ancestor's entry in its parent if it doesn't yet
			// exist.
			ancestor = &TreeEntry{
				Mode: "040000",
				Type: "tree",
				Name: filepath.Base(p),
			}
			t.Add(p, ancestor)
		}
		ancestor.OID = "" // mark dirty
	}
}

func (t *Tree) ApplyChanges(changes []*ChangedFile) error {
	for _, f := range changes {
		status := f.Status[0] // see "git diff-index --help" RAW OUTPUT FORMAT section for values

		if status == 'M' { // M=in-place edit
			src := t.Get(f.SrcPath)
			src.Mode = f.DstMode
			src.OID = f.DstSHA
			t.createOrDirtyAncestors(f.SrcPath)
		}

		if status == 'A' || status == 'C' || status == 'R' { // A=create, C=copy-edit, R=rename-edit
			var path string
			if status == 'A' {
				path = f.SrcPath // "git diff-index" calls created files' paths their src path not dst path
			} else {
				path = f.DstPath
			}

			typ, err := objectTypeForMode(f.DstMode)
			if err != nil {
				return err
			}

			e := &TreeEntry{
				Mode: f.DstMode,
				Type: typ,
				OID:  f.DstSHA,
				Name: filepath.Base(path),
			}
			t.createOrDirtyAncestors(path)
			t.Add(path, e)
		}

		if status == 'D' || status == 'R' { // D=delete, R=rename-edit
			t.createOrDirtyAncestors(f.SrcPath)
			t.remove(f.SrcPath)
		}
	}
	return nil
}

type TreeEntry struct {
	Mode    string
	Type    string // object type (blob, tree, etc.)
	OID     string
	Name    string
	Entries []*TreeEntry
}

func (e TreeEntry) String() string {
	s := fmt.Sprintf("%s %s %s %s", e.Mode, e.Type, e.OID, e.Name)
	if len(e.Entries) > 0 {
		entryNames := make([]string, len(e.Entries))
		for i, c := range e.Entries {
			entryNames[i] = c.Name
		}
		s += fmt.Sprintf(" [children: %s]", strings.Join(entryNames, " "))
	}
	return s
}

// ChangedFile represents a line in "git diff-index" output.
type ChangedFile struct {
	Status           string
	SrcMode, DstMode string
	SrcSHA, DstSHA   string
	SrcPath, DstPath string
}

func (c *ChangedFile) String() string {
	switch c.Status {
	case "A":
		return fmt.Sprintf("add %s", c.SrcPath)
	case "C":
		return fmt.Sprintf("copy %s -> %s", c.SrcPath, c.DstPath)
	case "D":
		return fmt.Sprintf("delete %s", c.SrcPath)
	case "M":
		return fmt.Sprintf("mod %s", c.SrcPath)
	case "R":
		return fmt.Sprintf("rename %s -> %s", c.SrcPath, c.DstPath)
	case "T":
		return fmt.Sprintf("change type %s %s -> %s", c.SrcPath, c.SrcMode, c.DstMode)
	case "U":
		return fmt.Sprintf("unmerged %s", c.SrcPath)
	default:
		return fmt.Sprintf("%s %s", c.Status, c.SrcPath)
	}
}
