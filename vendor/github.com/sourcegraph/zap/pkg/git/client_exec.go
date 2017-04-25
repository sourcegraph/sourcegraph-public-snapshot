package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/sourcegraph/zap/internal/pkg/backoff"
	"github.com/sourcegraph/zap/ot"
	"github.com/sourcegraph/zap/pkg/diff"
	"github.com/sourcegraph/zap/pkg/gitutil"
)

type gitRepo interface {
	ReadBlob(snapshot, name string) ([]byte, string, string, error)
	UpdateSymbolicRef(name, ref string) error
	ObjectNameSHA(arg string) (string, error)
	WorktreeDir() string
	GitDir() string
	IsValidRev(string) (bool, error)
	RemoteForBranchOrZapDefaultRemote(string) (string, error)
	RemoteURL(remote string) (string, error)
	MakeCommit(ctx context.Context, parent string, onlyIfChangedFiles bool) (string, []*gitutil.ChangedFile, error)
	Fetch(remote, refspec string) (bool, error)
	Push(remote, refspec string, force bool) error
	ConfigGetOne(name string) (string, error)
	ConfigSet(name, value string) error
	CreateCommitFromTree(ctx context.Context, tree, snapshot string, isRootCommit bool) (string, error)
	Reset(typ, rev string) error
	ListTreeFull(string) (*gitutil.Tree, error)
	FileInfoForPath(rev, path string) (string, string, error)
	HashObject(typ, path string, data []byte) (string, error)
	CreateTree(basePath string, entries []*gitutil.TreeEntry) (string, error)
}

type FileSystem interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, mode os.FileMode) error
	Rename(oldpath, newpath string) error
	Remove(name string) error
	Exists(name string) error
}

var TestApplyToWorktree func(ops ot.Ops) (unapplied ot.Ops, err error)

// ApplyToWorktree applies an op to the workspace. It may modify files
// on the file system. If an error occurs, it returns the unapplied
// remainder of the op. For example, if the op creates 3 files (f1,
// f2, and f3, in that order), and creating f2 fails, then it returns
// an op that creates f2 and f3 (because that part of the op was not
// successfully applied). Callers can use this to iteratively apply
// until there are no more changes in the op.
func ApplyToWorktree(ctx context.Context, logger log.Logger, gitRepo interface {
	ReadBlob(snapshot, name string) ([]byte, string, string, error)
	IsValidRev(string) (bool, error)
	RemoteForBranchOrZapDefaultRemote(string) (string, error)
	Fetch(string, string) (bool, error)
	Reset(string, string) error
}, fdisk, fbuf FileSystem, snapshot, gitBranch string, ops ot.Ops) (unapplied ot.Ops, err error) {
	if TestApplyToWorktree != nil {
		return TestApplyToWorktree(ops)
	}

	// OT_TODO: Do we need this check?
	// unapplied = ops.DeepCopy()

	fsForPath := func(path string) FileSystem {
		if isBufferPath(path) {
			return fbuf
		}
		return fdisk
	}

	// TODO(sqs): everything here is racy wrt local fs changes

	// TODO(sqs): check that paths are inside the dir/repo

	// TODO(sqs): make this read from prev snapshot and write to new
	// snapshot to avoid race conditions
	readFile := func(name string) ([]byte, error) {
		if isBufferPath(name) {
			return fbuf.ReadFile(stripFileOrBufferPath(name))
		}
		data, _, _, err := gitRepo.ReadBlob(snapshot, stripFilePath(name))
		return data, err
	}
	writeFile := func(name string, data []byte) error {
		return fsForPath(name).WriteFile(stripFileOrBufferPath(name), data, 0666)
	}

	created := make(map[string]struct{})
	for i, iop := range ops {
		unapplied = ops[i:]
		switch op := iop.(type) {
		case ot.FileCopy:
			src, dst := op.Src, op.Dst
			data, err := readFile(src)
			if err != nil {
				return unapplied, err
			}
			if err := writeFile(dst, data); err != nil {
				return unapplied, err
			}
		case ot.FileRename:
			src, dst := op.Src, op.Dst
			if err := checkRemotePath(src); err != nil {
				return unapplied, err
			}
			if err := checkRemotePath(dst); err != nil {
				return unapplied, err
			}
			if isBufferPath(src) || isBufferPath(dst) {
				return unapplied, fmt.Errorf("rename of buffer files not supported: %q -> %q", src, dst)
			}
			if err := fdisk.Rename(stripFilePath(src), stripFilePath(dst)); err != nil {
				return unapplied, err
			}
		case ot.FileCreate:
			f := op.File
			created[f] = struct{}{}
			if err := fsForPath(f).Exists(stripFileOrBufferPath(f)); !os.IsNotExist(err) {
				return unapplied, &os.PathError{Op: "Create", Path: f, Err: os.ErrExist}
			}
			if err := writeFile(f, nil); err != nil {
				return unapplied, err
			}
		case ot.FileDelete:
			f := op.File
			if err := fsForPath(f).Remove(stripFileOrBufferPath(f)); err != nil {
				return unapplied, err
			}
		case ot.FileTruncate:
			f := op.File
			if err := writeFile(f, nil); err != nil {
				return unapplied, err
			}
		case ot.FileEdit:
			f, edits := op.File, op.Edits
			if len(edits) == 0 {
				continue
			}
			var prevData []byte
			if _, justCreated := created[f]; !justCreated {
				var err error
				prevData, err = readFile(f)
				if err != nil {
					return unapplied, err
				}
			}
			doc := ot.Doc(string(prevData))

			if err := doc.Apply(edits); err != nil {
				return unapplied, &os.PathError{Op: "Edit", Path: f, Err: err}
			}
			if err := writeFile(f, []byte(string(doc))); err != nil {
				return unapplied, err
			}
		case ot.GitHead:
			if err := FetchAndCheck(ctx, gitRepo, gitBranch, "refs/zap/"+op.Commit, op.Commit); err != nil {
				return unapplied, err
			}

			// This acquires a lock and might conflict with other things
			// (error message of the form "... /index.lock"), like the
			// user running `git status`. Retry if it fails at first.
			if err := backoff.RetryNotifyWithContext(ctx, func(ctx context.Context) error {
				return gitRepo.Reset("mixed", op.Commit)
			}, GitBackOff(), nil); err != nil {
				return unapplied, err
			}
		}
	}
	return ot.Ops{}, nil
}

var TestFetchAndCheck func(remoteOfBranch, refspec, desiredRev string) error

func FetchAndCheck(ctx context.Context, gitRepo interface {
	IsValidRev(string) (bool, error)
	RemoteForBranchOrZapDefaultRemote(string) (string, error)
	Fetch(string, string) (bool, error)
	Reset(typ, rev string) error
}, remoteOfBranch, refspec, desiredRev string) error {
	if TestFetchAndCheck != nil {
		return TestFetchAndCheck(remoteOfBranch, refspec, desiredRev)
	}

	return backoff.RetryNotifyWithContext(ctx, func(ctx context.Context) error {
		return gitutil.FetchAndCheck(ctx, gitRepo, remoteOfBranch, refspec, desiredRev)
	}, GitBackOff(), func(err error, d time.Duration) {
		if true {
			fmt.Fprintf(os.Stderr, "# retrying git fetch (waiting for other client's commit %s) after error: %s\n", desiredRev, err)
		}
	})
}

// checkRemotePath returns a non-nil error if path (which is assumed
// to have come from an external source, such as the Zap sync
// server) refers to a location outside of the workspace.
func checkRemotePath(path string) error {
	check := func(path string) error {
		path = filepath.Clean(path)
		if filepath.IsAbs(path) {
			return fmt.Errorf("bad absolute path %q", path)
		}
		if path == ".." || strings.HasPrefix(path, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("bad path %q (outside of root)", path)
		}
		if path == ".git" || strings.HasPrefix(path, ".git"+string(os.PathSeparator)) {
			return fmt.Errorf("bad path %q (in .git directory)", path)
		}
		return nil
	}

	if err := check(path); err != nil {
		return err
	}
	path, err := filepath.EvalSymlinks(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return check(path)
}

type ReadFileFunc func(name string) ([]byte, error)

var TestWorkspaceOpForChanges func(changes []*gitutil.ChangedFile, readFileA, readFileB ReadFileFunc) (ot.Ops, error)

func WorkspaceOpForChanges(changes []*gitutil.ChangedFile, readFileA, readFileB ReadFileFunc) (ot.Ops, error) {
	if TestWorkspaceOpForChanges != nil {
		return TestWorkspaceOpForChanges(changes, readFileA, readFileB)
	}

	var ops ot.Ops
	for _, c := range changes {
		// TODO(sqs): sanitize/clean these paths
		srcPath := c.SrcPath
		dstPath := c.DstPath

		if strings.HasPrefix(srcPath, "/") {
			panic(fmt.Sprintf("unexpected '/' prefix in srcPath %q", srcPath))
		}
		if strings.HasPrefix(dstPath, "/") {
			panic(fmt.Sprintf("unexpected '/' prefix in dstPath %q", dstPath))
		}

		switch c.Status[0] {
		case 'R':
			ops = append(ops, ot.FileRename{Src: srcPath, Dst: dstPath})

			// Apply the "edit" part of a rename-edit, if any.
			prevData, err := readFileA(srcPath)
			if err != nil {
				return ot.Ops{}, err
			}
			data, err := readFileB(dstPath)
			if err != nil {
				return ot.Ops{}, err
			}
			if edits := DiffOps([]rune(string(prevData)), []rune(string(data))); len(edits) > 0 {
				ops = append(ops, ot.FileEdit{File: dstPath, Edits: edits})
			}

		case 'D':
			ops = append(ops, ot.FileDelete{File: c.SrcPath})

		case 'A', 'C', 'M':
			if dstPath == "" {
				dstPath = srcPath
			}

			switch c.Status[0] {
			case 'A':
				ops = append(ops, ot.FileCreate{File: srcPath})
			case 'C':
				ops = append(ops, ot.FileCopy{Src: srcPath, Dst: dstPath})
			}

			var prevData []byte
			if c.Status[0] == 'C' || c.Status[0] == 'M' {
				var err error
				prevData, err = readFileA(srcPath)
				if err != nil {
					return ot.Ops{}, err
				}
			}
			data, err := readFileB(dstPath)
			if err != nil {
				return ot.Ops{}, err
			}
			if edits := DiffOps([]rune(string(prevData)), []rune(string(data))); len(edits) > 0 {
				ops = append(ops, ot.FileEdit{File: dstPath, Edits: edits})
			}
		}
	}
	return ops, nil
}

// DiffOps returns the diff between old and new as OT edit ops.
//
// DEV NOTE: Keep this in sync with other language implementations of
// diffOps.
func DiffOps(old, new []rune) ot.EditOps {
	change := diff.Runes(old, new)
	ops := make(ot.EditOps, 0, len(change)*2)
	var ret, del, ins int
	for _, c := range change {
		if r := c.A - ret - del; r > 0 {
			ops = append(ops, ot.EditOp{N: r})
			ret = c.A - del
		}
		if c.Del > 0 {
			ops = append(ops, ot.EditOp{N: -c.Del})
			del += c.Del
		}
		if c.Ins > 0 {
			ops = append(ops, ot.EditOp{S: string(new[c.B : c.B+c.Ins])})
			ins += c.Ins
		}
	}
	if r := len(new) - ret - ins; r > 0 {
		ops = append(ops, ot.EditOp{N: r})
	}
	if del > 0 || ins > 0 {
		return ot.MergeEditOps(ops)
	}
	return nil
}

var testPushGitRefToGitUpstream func(headOID, gitBranch string) error

func PushGitRefToGitUpstream(ctx context.Context, gitRepo interface {
	RemoteForBranchOrZapDefaultRemote(string) (string, error)
	Push(string, string, bool) error
}, headOID, gitBranch string) error {
	if testPushGitRefToGitUpstream != nil {
		return testPushGitRefToGitUpstream(headOID, gitBranch)
	}

	gitRemote, err := gitRepo.RemoteForBranchOrZapDefaultRemote(gitBranch)
	if err != nil {
		return err
	}
	return backoff.RetryNotifyWithContext(ctx, func(ctx context.Context) error {
		return gitRepo.Push(gitRemote, headOID+":refs/zap/"+headOID, true)
	}, GitBackOff(), nil)
}
