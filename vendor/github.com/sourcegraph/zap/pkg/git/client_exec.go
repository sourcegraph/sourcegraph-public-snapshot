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
	MakeCommit(parent string, onlyIfChangedFiles bool) (string, []*gitutil.ChangedFile, error)
	Fetch(remote, refspec string) (bool, error)
	Push(remote, refspec string, force bool) error
	ConfigGetOne(name string) (string, error)
	ConfigSet(name, value string) error
	CreateCommitFromTree(tree, snapshot string, isRootCommit bool) (string, error)
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

var TestApplyToWorktree func(op ot.WorkspaceOp) error

func ApplyToWorktree(ctx context.Context, log *log.Context, gitRepo interface {
	WorktreeDir() string
	ReadBlob(snapshot, name string) ([]byte, string, string, error)
	IsValidRev(string) (bool, error)
	RemoteForBranchOrZapDefaultRemote(string) (string, error)
	Fetch(string, string) (bool, error)
	Reset(string, string) error
}, fdisk, fbuf FileSystem, snapshot, ref, gitBranch string, op ot.WorkspaceOp) error {
	if TestApplyToWorktree != nil {
		return TestApplyToWorktree(op)
	}

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

	for _, f := range op.Save {
		data, err := fbuf.ReadFile(stripBufferPath(f))
		if err != nil {
			return err
		}
		if err := fbuf.Remove(stripBufferPath(f)); err != nil {
			return err
		}
		if err := fdisk.WriteFile(stripBufferPath(f), data, 0666); err != nil {
			return err
		}
	}
	for dst, src := range op.Copy {
		data, err := readFile(src)
		if err != nil {
			return err
		}
		if err := writeFile(dst, data); err != nil {
			return err
		}
	}
	for src, dst := range op.Rename {
		if err := checkRemotePath(src); err != nil {
			return err
		}
		if err := checkRemotePath(dst); err != nil {
			return err
		}
		if isBufferPath(src) || isBufferPath(dst) {
			panic(fmt.Sprintf("rename of buffer files not supported: %q -> %q", src, dst))
		}
		if err := fdisk.Rename(stripFilePath(src), stripFilePath(dst)); err != nil {
			return err
		}
	}
	created := make(map[string]struct{}, len(op.Create))
	for _, f := range op.Create {
		created[f] = struct{}{}
		if err := fsForPath(f).Exists(stripFileOrBufferPath(f)); !os.IsNotExist(err) {
			return &os.PathError{Op: "Create", Path: f, Err: os.ErrExist}
		}
		if err := writeFile(f, nil); err != nil {
			return err
		}
	}
	for _, f := range op.Delete {
		if err := fsForPath(f).Remove(stripFileOrBufferPath(f)); err != nil {
			return err
		}
	}
	for _, f := range op.Truncate {
		if err := writeFile(f, nil); err != nil {
			return err
		}
	}
	for f, edits := range op.Edit {
		if len(edits) == 0 {
			continue
		}
		var prevData []byte
		if _, justCreated := created[f]; !justCreated {
			var err error
			prevData, err = readFile(f)
			if err != nil {
				return err
			}
		}
		doc := ot.Doc(prevData)

		/// DEBUG LOG
		if isBufferPath(f) {
			log.Log("apply-to-worktree--edit", f, "pre-contents", string(prevData), "edits", fmt.Sprint(edits))
		}
		// END DEBUG LOG

		if err := doc.Apply(edits); err != nil {
			return &os.PathError{Op: "Edit", Path: f, Err: err}
		}
		if err := writeFile(f, doc); err != nil {
			return err
		}
	}
	if op.GitHead != "" {
		if err := FetchAndCheck(ctx, gitRepo, gitBranch, "refs/zap/"+ref, op.GitHead); err != nil {
			return err
		}

		// This acquires a lock and might conflict with other things
		// (error message of the form "... /index.lock"), like the
		// user running `git status`. Retry if it fails at first.
		if err := backoff.RetryNotifyWithContext(ctx, func(ctx context.Context) error {
			fmt.Fprintf(os.Stderr, "# git reset %s\n", op.GitHead)
			return gitRepo.Reset("mixed", op.GitHead)
		}, GitBackOff(), nil); err != nil {
			return err
		}
	}
	return nil
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

func WorkspaceOpForChanges(changes []*gitutil.ChangedFile, readFileA, readFileB ReadFileFunc) (ot.WorkspaceOp, error) {
	var op ot.WorkspaceOp
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
			if op.Rename == nil {
				op.Rename = map[string]string{}
			}
			op.Rename[srcPath] = dstPath

			// Apply the "edit" part of a rename-edit, if any.
			prevData, err := readFileA(srcPath)
			if err != nil {
				return ot.WorkspaceOp{}, err
			}
			data, err := readFileB(dstPath)
			if err != nil {
				return ot.WorkspaceOp{}, err
			}
			if edits := diffOps(prevData, data); len(edits) > 0 {
				op.Edit = map[string]ot.EditOps{dstPath: edits}
			}

		case 'D':
			op.Delete = append(op.Delete, c.SrcPath)

		case 'A', 'C', 'M':
			if dstPath == "" {
				dstPath = srcPath
			}

			switch c.Status[0] {
			case 'A':
				op.Create = append(op.Create, srcPath)
			case 'C':
				if op.Copy == nil {
					op.Copy = map[string]string{}
				}
				op.Copy[dstPath] = srcPath
			}

			var prevData []byte
			if c.Status[0] == 'C' || c.Status[0] == 'M' {
				var err error
				prevData, err = readFileA(srcPath)
				if err != nil {
					return ot.WorkspaceOp{}, err
				}
			}
			data, err := readFileB(dstPath)
			if err != nil {
				return ot.WorkspaceOp{}, err
			}
			if edits := diffOps(prevData, data); len(edits) > 0 {
				if op.Edit == nil {
					op.Edit = map[string]ot.EditOps{}
				}
				op.Edit[dstPath] = edits
			}
		}
	}
	return op, nil
}

func diffOps(old, new []byte) ot.EditOps {
	change := diff.Bytes(old, new)
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

var testPushGitRefToGitUpstream func(headOID, ref, gitBranch string) error

func PushGitRefToGitUpstream(ctx context.Context, gitRepo gitRepo, headOID, ref, gitBranch string) error {
	if testPushGitRefToGitUpstream != nil {
		return testPushGitRefToGitUpstream(headOID, ref, gitBranch)
	}

	gitRemote, err := gitRepo.RemoteForBranchOrZapDefaultRemote(gitBranch)
	if err != nil {
		return err
	}

	return backoff.RetryNotifyWithContext(ctx, func(ctx context.Context) error {
		return gitRepo.Push(gitRemote, headOID+":refs/zap/"+ref, true)
	}, GitBackOff(), nil)
}
