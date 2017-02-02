package gitutil

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitExecutor executes a git subcommand (with input and args) in a
// git repository.
type GitExecutor interface {
	Exec(input []byte, args ...string) ([]byte, error)
}

// BareRepo refers to a bare git repository directory (which can be
// the .git directory of a non-bare git repository).
type BareRepo struct {
	GitExecutor
}

// BareFSRepo is a bare repo that exists on the local disk at a known
// directory.
type BareFSRepo struct {
	Dir string
	BareRepo
}

func (r BareFSRepo) GitDir() string {
	if r.Dir == "" {
		panic("git repo dir is empty")
	}
	return r.Dir
}

// BareRepoDir creates a new git BareRepo/BareFSRepo wrapper for the
// bare git repo at dir.
func BareRepoDir(dir string) BareFSRepo {
	return BareFSRepo{
		Dir: dir,
		BareRepo: BareRepo{
			GitExecutor: gitExecutorFunc(func(input []byte, args ...string) ([]byte, error) {
				var f func(*exec.Cmd)
				if input != nil {
					f = func(c *exec.Cmd) { c.Stdin = bytes.NewReader(input) }
				}
				out, err := execCustomGitCommand(dir, f, args...)
				return []byte(out), err
			}),
		},
	}
}

type gitExecutorFunc func(input []byte, args ...string) ([]byte, error)

func (f gitExecutorFunc) Exec(input []byte, args ...string) ([]byte, error) {
	return f(input, args...)
}

func (r BareRepo) ObjectNameSHA(arg string) (string, error) {
	if !strings.Contains(arg, "^") {
		panic("arg should have a type-peeling operator like ^{commit}, or else all 40-hex-char args will be treated as valid even if they don't refer to anything")
	}
	if err := checkArgSafety(arg); err != nil {
		return "", err
	}
	sha, err := r.Exec(nil, "rev-parse", "--verify", arg)
	if err != nil && strings.Contains(arg, "4b825") {
		panic("X")
	}
	return string(bytes.TrimSpace(sha)), err
}

func (r BareRepo) Fetch(remote, refspec string) (remoteRefNotFound bool, err error) {
	if err := checkArgSafety(remote); err != nil {
		return false, err
	}
	if err := checkArgSafety(refspec); err != nil {
		return false, err
	}

	_, err = r.Exec(nil, "fetch", "--quiet", remote, refspec)
	if err != nil && strings.Contains(err.Error(), "fatal: Couldn't find remote ref refs/") {
		remoteRefNotFound = true
	}
	return
}

func (r BareRepo) Push(remote, refspec string, force bool) error {
	if err := checkArgSafety(remote); err != nil {
		return err
	}
	if err := checkArgSafety(refspec); err != nil {
		return err
	}

	args := []string{"push", "--quiet"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, remote, refspec)
	_, err := r.Exec(nil, args...)
	return err
}

func (r BareRepo) IsValidRev(rev string) (bool, error) {
	if strings.Contains(rev, "^") {
		panic("likely incorrect usage of gitIsValidRev; ^{commit} is automatically added to rev, so rev should not contain ^: " + rev)
	}
	if err := checkArgSafety(rev); err != nil {
		return false, err
	}
	_, err := r.Exec(nil, "rev-parse", "--quiet", "--verify", rev+"^{commit}")
	if err == nil {
		return true, nil
	}
	if strings.Contains(err.Error(), "exit status 1") {
		return false, nil
	}
	return false, err
}

func (r BareRepo) RemoteURL(remote string) (string, error) {
	url, err := r.Exec(nil, "config", "remote."+remote+".url")
	return string(bytes.TrimSpace(url)), err
}

type NoConfigKeyError struct {
	Key string
}

func (e *NoConfigKeyError) Error() string {
	return fmt.Sprintf("git config has no config key %q", string(e.Key))
}

func (r BareRepo) ConfigGetOne(name string) (string, error) {
	values, err := r.ConfigGetAll(name)
	if err != nil {
		return "", err
	}
	if len(values) == 0 {
		return "", &NoConfigKeyError{name}
	}
	if len(values) > 1 {
		return "", fmt.Errorf("git config %q has %d values (expected 1)", name, len(values))
	}
	return values[0], nil
}

func (r BareRepo) ConfigGetAll(name string) ([]string, error) {
	if err := checkArgSafety(name); err != nil {
		return nil, err
	}
	value, err := r.Exec(nil, "config", "--get-all", "-z", name)
	value = bytes.TrimSpace(value)
	if err != nil && len(value) == 0 && strings.Contains(err.Error(), "exit status 1") {
		// When getting all values, 0 values is not an error.
		return nil, nil
	}
	return splitNullsBytesToStrings(value), err
}

func (r BareRepo) ConfigSet(name, value string) error {
	if err := checkArgSafety(name); err != nil {
		return err
	}
	if err := checkArgSafety(value); err != nil {
		return err
	}
	_, err := r.Exec(nil, "config", name, value)
	return err
}

func (r BareRepo) ConfigAdd(name, value string) error {
	if err := checkArgSafety(name); err != nil {
		return err
	}
	if err := checkArgSafety(value); err != nil {
		return err
	}
	_, err := r.Exec(nil, "config", "--add", name, value)
	return err
}

func (r BareRepo) RemoteForBranchOrZapDefaultRemote(branch string) (string, error) {
	remote, err := r.remoteForBranch(branch)
	if _, ok := err.(*NoUpstreamError); ok {
		return r.ConfigGetOne("zap.defaultRemote")
	}
	return remote, err
}

// remoteForBranch gets the remote that is configured as the named
// branch's upstream. NOTE: You should always use
// RemoteForBranchOrZapDefaultRemote to use the zap.defaultRemote
// setting if no remote exists (which happens frequently, such as on
// newly created and not-yet-pushed branches).
func (r BareRepo) remoteForBranch(branch string) (string, error) {
	remoteBytes, err := r.Exec(nil, "config", "branch."+branch+".remote")
	remote := string(bytes.TrimSpace(remoteBytes))
	if err != nil {
		if remote == "" && strings.Contains(err.Error(), "exit status 1") {
			return "", &NoUpstreamError{Branch: branch}
		}
		return "", err
	}
	exists, err := r.RemoteExists(remote)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("remote does not exist: %q", remote)
	}
	return remote, nil
}

type NoUpstreamError struct{ Branch string }

func (e *NoUpstreamError) Error() string {
	return fmt.Sprintf("branch %q has no upstream configured", e.Branch)
}

func (r BareRepo) RemoteExists(remote string) (bool, error) {
	_, err := r.Exec(nil, "config", "remote."+remote+".url")
	if err == nil {
		return true, nil
	}
	if strings.Contains(err.Error(), "exit status 1") {
		return false, nil
	}
	return false, err
}

func (r BareRepo) HEADHasNoCommitsAndNextCommitWillBeRootCommit() (bool, error) {
	if _, err := r.ObjectNameSHA("HEAD^{commit}"); err != nil {
		if strings.Contains(err.Error(), "Needed a single revision") {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func (r BareRepo) ReadBlob(treeish, path string) (data []byte, mode, oid string, err error) {
	mode, oid, err = r.FileInfoForPath(treeish, path)
	if err != nil {
		return nil, "", "", err
	}
	typ, err := objectTypeForMode(mode)
	if err != nil || typ != "blob" {
		return nil, "", "", &os.PathError{Op: "gitReadBlob (tree: " + treeish + ")", Path: path, Err: os.ErrInvalid}
	}

	if err := checkArgSafety(treeish); err != nil {
		return nil, "", "", err
	}
	if strings.Contains(treeish, ":") {
		return nil, "", "", fmt.Errorf("bad treeish arg (contains ':'): %q", treeish)
	}
	if err := checkArgSafety(path); err != nil {
		return nil, "", "", err
	}
	contents, err := r.CatFile(typ, oid)
	if err != nil {
		return nil, "", "", err
	}
	return contents, mode, oid, nil
}

func (r BareRepo) CatFile(typ, oid string) ([]byte, error) {
	if err := checkArgSafety(typ); err != nil {
		return nil, err
	}
	if err := checkArgSafety(oid); err != nil {
		return nil, err
	}
	return r.Exec(nil, "cat-file", typ, oid)
}

func (r BareRepo) FileInfoForPath(treeish, path string) (mode, oid string, err error) {
	out, err := r.Exec(nil, "ls-tree", "-z", "-t", "--full-name", "--full-tree", "--", treeish, path)
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
	panic(fmt.Sprintf("ERROR: %q", path))
	return "", "", &os.PathError{Op: "gitFileInfoForPath (tree: " + treeish + ")", Path: path, Err: os.ErrNotExist}
}

func (r BareRepo) HEADOrDevNullTree() (string, error) {
	onRootCommit, err := r.HEADHasNoCommitsAndNextCommitWillBeRootCommit()
	if err != nil {
		return "", err
	}
	if onRootCommit {
		return r.DevNullTree()
	}
	return r.ObjectNameSHA("HEAD^{commit}")
}

func (r BareRepo) DevNullTree() (oid string, err error) {
	// This should be the same oid everywhere, but let's compute it
	// just in case.
	oidBytes, err := r.Exec(nil, "hash-object", "-t", "tree", "/dev/null")
	return string(bytes.TrimSpace(oidBytes)), err
}

func (r BareRepo) HashObject(typ, path string, data []byte) (oid string, err error) {
	if err := checkArgSafety(typ); err != nil {
		return "", err
	}
	if err := checkArgSafety(path); err != nil {
		return "", err
	}
	oidBytes, err := r.Exec(data, "hash-object", "-t", typ, "-w", "--stdin", "--path", path)
	return string(bytes.TrimSpace(oidBytes)), err
}

func (r BareRepo) UpdateRef(ref, value string) error {
	if err := checkArgSafety(ref); err != nil {
		return err
	}
	if err := checkArgSafety(value); err != nil {
		return err
	}
	_, err := r.Exec(nil, "update-ref", ref, value)
	return err
}

func (r BareRepo) ReadSymbolicRef(name string) (string, error) {
	if err := checkArgSafety(name); err != nil {
		return "", err
	}
	value, err := r.Exec(nil, "symbolic-ref", name)
	return string(bytes.TrimSpace(value)), err
}

func (r BareRepo) UpdateSymbolicRef(name, ref string) error {
	if err := checkArgSafety(name); err != nil {
		return err
	}
	if err := checkArgSafety(ref); err != nil {
		return err
	}
	_, err := r.Exec(nil, "symbolic-ref", name, ref)
	return err
}

func (r BareRepo) ListTreeFull(head string) (*Tree, error) {
	if err := checkArgSafety(head); err != nil {
		return nil, err
	}

	// This is pretty fast, even on large repositories (45ms on the
	// Sourcegraph repository).
	out, err := r.Exec(nil, "ls-tree", "-r", "-t", "-z", "--full-tree", "--full-name", head)
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

func (r BareRepo) CreateTree(basePath string, entries []*TreeEntry) (string, error) {
	var buf bytes.Buffer // output in the "git ls-tree" format
	for _, e := range entries {
		// Entries that were added, and the ancestor trees thereof,
		// have empty or all-zero SHAs.
		if e.OID == "" || e.OID == SHAAllZeros {
			path := filepath.Join(basePath, e.Name)

			switch e.Type {
			case "blob":
				return "", fmt.Errorf("tree entry blob at %q must have OID set when creating tree in bare repo (OID is %q)", path, e.OID)

			case "tree":
				var err error
				e.OID, err = r.CreateTree(path, e.Entries)
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

	const commitVerbose = false // DEV

	stdinBytes := buf.Bytes()
	oidBytes, err := r.Exec(stdinBytes, "mktree", "-z")
	if err != nil {
		if commitVerbose {
			return "", fmt.Errorf("%s\n\nstdin input follows:\n%s", err, stdinBytes)
		}
		return "", err
	}
	return string(bytes.TrimSpace(oidBytes)), nil
}

func (r BareRepo) CreateCommitFromTree(tree, parent string, isRootCommit bool) (oid string, err error) {
	args := []string{"commit-tree", "-m", "wip"}
	if !isRootCommit {
		parent, err := r.ObjectNameSHA(parent + "^{commit}")
		if err != nil {
			return "", err
		}
		args = append(args, "-p", parent)
	}
	args = append(args, tree)
	oidBytes, err := r.Exec(nil, args...)
	return string(bytes.TrimSpace(oidBytes)), err
}
