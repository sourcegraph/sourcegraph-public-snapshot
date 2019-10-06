package server

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
)

const shaAllZeros = "0000000000000000000000000000000000000000"

// Tree represents a git tree (i.e., "git ls-tree -r -d").
type Tree struct {
	byPath map[string]*TreeEntry // key is path without leading or trailing slashes (e.g., "d", "d1/d2", "d/f")
	Root   []*TreeEntry
}

func (t *Tree) Add(path string, e *TreeEntry) error {
	if path == "" || path == "." {
		panic(fmt.Sprintf("bad path: %q", path))
	}

	if t.byPath == nil {
		t.byPath = map[string]*TreeEntry{}
	}
	if _, exists := t.byPath[path]; exists {
		return fmt.Errorf("Tree.Add: path already exists: %q", path)
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
	return nil
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
func (t *Tree) createOrDirtyAncestors(path string) error {
	dir := filepath.Dir(path)
	if dir == "." {
		return nil
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
			if err := t.Add(p, ancestor); err != nil {
				return err
			}
		}
		ancestor.OID = "" // mark dirty
	}
	return nil
}

func (t *Tree) ApplyChanges(changes []*ChangedFile) error {
	for _, f := range changes {
		status := f.Status[0] // see "git diff-index --help" RAW OUTPUT FORMAT section for values

		if status == 'M' { // M=in-place edit
			src := t.Get(f.SrcPath)
			src.Mode = f.DstMode
			src.OID = f.DstSHA
			if err := t.createOrDirtyAncestors(f.SrcPath); err != nil {
				return err
			}
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
			if err := t.createOrDirtyAncestors(path); err != nil {
				return err
			}
			if err := t.Add(path, e); err != nil {
				return err
			}
		}

		if status == 'D' || status == 'R' { // D=delete, R=rename-edit
			if err := t.createOrDirtyAncestors(f.SrcPath); err != nil {
				return err
			}
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

const RegularFileNonExecutableMode = "100644"

func objectTypeForMode(modeStr string) (string, error) {
	switch modeStr {
	case "040000": // directory
		return "tree", nil
	case RegularFileNonExecutableMode, "100755": // regular file
		return "blob", nil
	case "120000": // symlink
		return "blob", nil
	case "160000": // submodule
		return "commit", nil
	default:
		return "", fmt.Errorf("unrecognized git mode %q", modeStr)
	}
}

func createCommitFromTree(ctx context.Context, repoDir string, tree string, isRootCommit bool, req protocol.CreateCommitFromPatchRequest) (oid string, err error) {
	args := []string{"commit-tree", "-m", req.CommitInfo.Message}
	if !isRootCommit {
		parent, err := objectNameSHA(ctx, repoDir, string(req.BaseCommit)+"^{commit}")
		if err != nil {
			return "", err
		}
		args = append(args, "-p", parent)
	}
	args = append(args, tree)
	cmd := exec.CommandContext(ctx, "git", args...)
	setGitEnvForCommit(cmd, req)
	cmd.Dir = repoDir
	oidBytes, err := cmd.Output()
	return string(bytes.TrimSpace(oidBytes)), err
}

func objectNameSHA(ctx context.Context, repoDir string, arg string) (string, error) {
	if !strings.Contains(arg, "^") {
		panic("arg should have a type-peeling operator like ^{commit}, or else all 40-hex-char args will be treated as valid even if they don't refer to anything")
	}
	if err := checkSpecArgSafety(arg); err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--verify", arg)
	cmd.Dir = repoDir
	sha, err := cmd.Output()
	return string(bytes.TrimSpace(sha)), err
}

func createCommitFromPatch(ctx context.Context, repoDir string, req protocol.CreateCommitFromPatchRequest) (string, error) {
	tree, err := createTreeFromPatch(ctx, repoDir, string(req.BaseCommit), []byte(req.Patch))
	if err != nil {
		return "", err
	}
	// TODO!(sqs): set committer, message, etc
	return createCommitFromTree(ctx, repoDir, tree, false, req)
}

func createTreeFromPatch(ctx context.Context, repoDir string, base string, rawDiff []byte) (string, error) {
	tree, err := listTreeFull(ctx, repoDir, base)
	if err != nil {
		return "", err
	}

	updateGitFile := func(filename string, newData []byte) error {
		if strings.HasPrefix(filename, "/") || strings.HasPrefix(filename, "#") {
			panic(fmt.Sprintf("expected stripped filename, got %q", filename))
		}
		e := tree.Get(filename)
		newOID, err := hashObject(ctx, repoDir, "blob", filename, newData)
		if err != nil {
			return err
		}
		return tree.ApplyChanges([]*ChangedFile{{
			Status:  "M",
			SrcMode: e.Mode, DstMode: e.Mode,
			SrcSHA: e.OID, DstSHA: newOID,
			SrcPath: filename,
		}})
	}

	fileDiffs, err := diff.ParseMultiFileDiff(rawDiff)
	if err != nil {
		return "", err
	}

	for _, fileDiff := range fileDiffs {
		fileDiff.OrigName = strings.TrimPrefix(fileDiff.OrigName, "a/")
		fileDiff.NewName = strings.TrimPrefix(fileDiff.NewName, "b/")
		switch {
		case fileDiff.OrigName != "/dev/null" && fileDiff.NewName != "/dev/null": // modify file
			if fileDiff.OrigName != fileDiff.NewName {
				panic("TODO!(sqs): renames not handled")
			}
			origData, _, _, err := readBlob(ctx, repoDir, base, fileDiff.OrigName)
			if err != nil {
				return "", err
			}
			newData := applyPatch(origData, fileDiff)
			if err := updateGitFile(fileDiff.OrigName, newData); err != nil {
				return "", err
			}

		case fileDiff.OrigName == "/dev/null": // create file
			newData := applyPatch(nil, fileDiff)

			// Ensure we have the object for the empty blob.
			//
			// NOTE: This *almost* always produces the
			// SHAEmptyBlob oid, but you can hack gitattributes to
			// make this return something else, and we want to avoid
			// making assumptions about your git repo that could ever be
			// violated.
			//
			// TODO(sqs): could optimize by leaving the dstSHA blank for
			// newly created files that we have nonzero edits for (we will
			// compute the dstSHA again for those files anyway).
			oid, err := hashObject(ctx, repoDir, "blob", fileDiff.NewName, newData)
			if err != nil {
				return "", err
			}

			// We will fill in the dstSHA below in Edit when we have the
			// file contents, if we have edits.
			if err := tree.ApplyChanges([]*ChangedFile{{
				Status:  "A",
				SrcMode: shaAllZeros, DstMode: RegularFileNonExecutableMode,
				SrcSHA: shaAllZeros, DstSHA: oid,
				SrcPath: fileDiff.NewName,
			}}); err != nil {
				return "", err
			}

		case fileDiff.NewName == "/dev/null": // delete file
			mode, sha, err := fileInfoForPath(ctx, repoDir, base, fileDiff.OrigName)
			if err != nil {
				return "", err
			}
			if err := tree.ApplyChanges([]*ChangedFile{{
				Status:  "D",
				SrcMode: mode, DstMode: shaAllZeros,
				SrcSHA: sha, DstSHA: shaAllZeros,
				SrcPath: fileDiff.OrigName,
			}}); err != nil {
				return "", err
			}

		default:
			panic("unhandled")
		}
	}

	if tree == nil {
		return "", nil // indicates no new tree SHA was created
	}
	return createTree(ctx, repoDir, "", tree.Root)
}

func listTreeFull(ctx context.Context, repoDir string, head string) (*Tree, error) {
	if err := checkSpecArgSafety(head); err != nil {
		return nil, err
	}

	// This is pretty fast, even on large repositories (45ms on the
	// Sourcegraph repository).
	cmd := exec.CommandContext(ctx, "git", "ls-tree", "-r", "-t", "-z", "--full-tree", "--full-name", head)
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var t Tree
	for _, line := range splitNullsBytes(out) {
		mode, typ, oid, path, err := parseLsTreeLine(line)
		if err != nil {
			return nil, err
		}
		t.Add(path, &TreeEntry{
			Mode: mode,
			Type: typ,
			OID:  oid,
			Name: filepath.Base(path),
		})
	}
	return &t, nil
}

func readBlob(ctx context.Context, repoDir string, treeish, path string) (data []byte, mode, oid string, err error) {
	mode, oid, err = fileInfoForPath(ctx, repoDir, treeish, path)
	if err != nil {
		return nil, "", "", err
	}
	typ, err := objectTypeForMode(mode)
	if err != nil || typ != "blob" {
		return nil, "", "", &os.PathError{Op: "gitReadBlob (tree: " + treeish + ")", Path: path, Err: os.ErrInvalid}
	}

	if err := checkSpecArgSafety(treeish); err != nil {
		return nil, "", "", err
	}
	if strings.Contains(treeish, ":") {
		return nil, "", "", fmt.Errorf("bad treeish arg (contains ':'): %q", treeish)
	}
	if err := checkSpecArgSafety(path); err != nil {
		return nil, "", "", err
	}
	cmd := exec.CommandContext(ctx, "git", "cat-file", typ, oid)
	cmd.Dir = repoDir
	contents, err := cmd.Output()
	if err != nil {
		return nil, "", "", err
	}
	return contents, mode, oid, nil
}

func fileInfoForPath(ctx context.Context, repoDir string, treeish, path string) (mode, oid string, err error) {
	cmd := exec.CommandContext(ctx, "git", "ls-tree", "-z", "-t", "--full-name", "--full-tree", "--", treeish, path)
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return "", "", err
	}
	for _, line := range splitNullsBytes(out) {
		mode, _, oid, path2, err := parseLsTreeLine(line)
		if err != nil {
			return "", "", err
		}
		if path2 == path {
			return mode, oid, nil
		}
	}
	return "", "", &os.PathError{Op: "gitFileInfoForPath (tree: " + treeish + ")", Path: path, Err: os.ErrNotExist}
}

func createTree(ctx context.Context, repoDir string, basePath string, entries []*TreeEntry) (string, error) {
	var buf bytes.Buffer // output in the "git ls-tree" format
	for _, e := range entries {
		// Entries that were added, and the ancestor trees thereof,
		// have empty or all-zero SHAs.
		if e.OID == "" || e.OID == shaAllZeros {
			path := filepath.Join(basePath, e.Name)

			switch e.Type {
			case "blob":
				return "", fmt.Errorf("tree entry blob at %q must have OID set when creating tree in bare repo (OID is %q)", path, e.OID)

			case "tree":
				var err error
				e.OID, err = createTree(ctx, repoDir, path, e.Entries)
				if err != nil {
					return "", err
				}

			default:
				// There are no known cases that this should happen
				// for, but this case is handled with an error message
				// just in case.
				//
				// This is only triggered when e.oid is zeroed out,
				// which should never happen for submodules (e.typ ==
				// "commit").
				return "", fmt.Errorf("repository contains unsupported tree entry type %q at %q", e.Type, path)
			}
		}

		fmt.Fprintf(&buf, "%s %s %s\t%s\x00", e.Mode, e.Type, e.OID, e.Name)
	}

	cmd := exec.CommandContext(ctx, "git", "mktree", "-z")
	cmd.Dir = repoDir
	cmd.Stdin = &buf
	oidBytes, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(oidBytes)), nil
}

func hashObject(ctx context.Context, repoDir string, typ, path string, data []byte) (oid string, err error) {
	if err := checkSpecArgSafety(typ); err != nil {
		return "", err
	}
	if err := checkSpecArgSafety(path); err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, "git", "hash-object", "-t", typ, "-w", "--stdin", "--path", path)
	cmd.Dir = repoDir
	cmd.Stdin = bytes.NewReader(data)
	oidBytes, err := cmd.Output()
	return string(bytes.TrimSpace(oidBytes)), err
}

func parseLsTreeLine(line []byte) (mode, typ, oid, path string, err error) {
	partsTab := bytes.SplitN(line, []byte("\t"), 2)
	if len(partsTab) != 2 {
		err = fmt.Errorf("bad ls-tree line: %q", line)
		return
	}

	path = string(partsTab[1])
	partsFirst := bytes.Split(partsTab[0], []byte(" "))
	if len(partsFirst) != 3 {
		err = fmt.Errorf("bad ls-tree line section (before first TAB): %q", partsTab[0])
		return
	}
	mode = string(partsFirst[0])
	typ = string(partsFirst[1])
	oid = string(partsFirst[2])
	return
}

func splitNullsBytes(s []byte) [][]byte {
	if len(s) == 0 {
		return nil
	}
	if s[len(s)-1] == '\x00' {
		s = s[:len(s)-1]
	}
	if len(s) == 0 {
		return nil
	}
	return bytes.Split(s, []byte("\x00"))
}

// TODO!(sqs): this is pretty hacky and not comprehensive, also TODO!(sqs): check bounds and return error not panic
func applyPatch(data []byte, fileDiff *diff.FileDiff) []byte {
	lines := bytes.SplitAfter(data, []byte("\n"))
	for _, hunk := range fileDiff.Hunks {
		hunkLines := bytes.SplitAfter(hunk.Body, []byte("\n"))
		origAdjust := 0
		for _, hunkLine := range hunkLines {
			if len(hunkLine) == 0 {
				continue
			}
			switch hunkLine[0] {
			case '-': // remove line
				l := int(hunk.NewStartLine-1) + origAdjust
				lines = append(lines[:l], lines[l+1:]...)
			case '+': // add line
				l := int(hunk.NewStartLine-1) + origAdjust
				lines = append(lines, nil)
				copy(lines[l+1:], lines[l:])
				lines[l] = hunkLine[1:]
				origAdjust++
			case ' ': // unchanged
				origAdjust++
			default:
				panic("unhandled")
			}
		}
	}
	return bytes.Join(lines, nil)
}
