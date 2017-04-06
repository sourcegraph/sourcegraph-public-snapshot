package git

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/sourcegraph/zap"
	"github.com/sourcegraph/zap/internal/pkg/pathutil"
	"github.com/sourcegraph/zap/ot"
	"github.com/sourcegraph/zap/pkg/gitutil"
)

func CreateTreeForOp(logger log.Logger, gitRepo interface {
	ReadBlob(snapshot, name string) ([]byte, string, string, error)
	ListTreeFull(string) (*gitutil.Tree, error)
	FileInfoForPath(rev, path string) (string, string, error)
	HashObject(typ, path string, data []byte) (string, error)
	CreateTree(basePath string, entries []*gitutil.TreeEntry) (string, error)
}, fbuf FileSystem, base string, op ot.WorkspaceOp) (string, error) {
	// As an optimization, if op does not depend on the Git tree, then
	// do not compute the Git tree.
	var tree *gitutil.Tree
	if OpDependsOnGitTree(op) {
		var err error
		tree, err = gitRepo.ListTreeFull(base)
		if err != nil {
			return "", err
		}
	}

	panicOnSomeErrors := os.Getenv("WORKSPACE_APPLY_ERRORS_FATAL") != ""

	fileOrigin := func(name string) string {
		if src, ok := op.Copy[name]; ok {
			return src
		}
		for src, dst := range op.Rename {
			if dst == name {
				return src
			}
		}
		return name
	}

	updateGitFile := func(filename string, newData []byte) error {
		if strings.HasPrefix(filename, "/") || strings.HasPrefix(filename, "#") {
			panic(fmt.Sprintf("expected stripped filename, got %q", filename))
		}
		e := tree.Get(filename)
		newOID, err := gitRepo.HashObject("blob", filename, newData)
		if err != nil {
			return err
		}
		return tree.ApplyChanges([]*gitutil.ChangedFile{{
			Status:  "M",
			SrcMode: e.Mode, DstMode: e.Mode,
			SrcSHA: e.OID, DstSHA: newOID,
			SrcPath: filename,
		}})
	}

	for _, f := range op.Save {
		if !isBufferPath(f) {
			panic(fmt.Sprintf("op.Save file %q must be a buffer path", f))
		}
		data, err := fbuf.ReadFile(stripFileOrBufferPath(f))
		if err != nil {
			return "", err
		}
		if err := updateGitFile(stripFileOrBufferPath(f), data); err != nil {
			return "", err
		}
		if err := fbuf.Remove(stripFileOrBufferPath(f)); err != nil {
			return "", err
		}
	}
	for dst, src := range op.Copy {
		if isBufferPath(src) {
			panic("not yet implemented")
		}
		if isBufferPath(dst) {
			data, _, _, err := gitRepo.ReadBlob(base, stripFileOrBufferPath(src))
			if err != nil {
				return "", err
			}
			if err := fbuf.Exists(stripFileOrBufferPath(dst)); !os.IsNotExist(err) {
				err = fmt.Errorf("copy %q to %q: destination file %q already exists", src, dst, dst)
				if panicOnSomeErrors {
					panic(err)
				}
				return "", err
			}
			if err := fbuf.WriteFile(stripFileOrBufferPath(dst), data, 0666); err != nil {
				return "", err
			}
		} else {
			// TODO(sqs): handle copy-then-modify
			mode, sha, err := gitRepo.FileInfoForPath(base, stripFileOrBufferPath(src))
			if err != nil {
				return "", err
			}
			if err := tree.ApplyChanges([]*gitutil.ChangedFile{{
				Status:  "C",
				SrcMode: mode, DstMode: mode,
				SrcSHA: sha, DstSHA: sha,
				SrcPath: stripFileOrBufferPath(src), DstPath: stripFileOrBufferPath(dst),
			}}); err != nil {
				level.Error(logger).Log("tree-apply-changes-failed", err, "src", src, "dst", dst, "base", base, "op", op)
				return "", err
			}
		}
	}
	for src, dst := range op.Rename {
		if isBufferPath(src) || isBufferPath(dst) {
			panic("not yet implemented")
		}

		// TODO(sqs): handle rename-then-modify
		mode, sha, err := gitRepo.FileInfoForPath(base, stripFileOrBufferPath(src))
		if err != nil {
			return "", err
		}
		if err := tree.ApplyChanges([]*gitutil.ChangedFile{{
			Status:  "R",
			SrcMode: mode, DstMode: mode,
			SrcSHA: sha, DstSHA: sha,
			SrcPath: stripFileOrBufferPath(src), DstPath: stripFileOrBufferPath(dst),
		}}); err != nil {
			return "", err
		}
	}
	created := make(map[string]struct{}, len(op.Create))
	for _, f := range op.Create {
		created[f] = struct{}{}

		if isBufferPath(f) {
			if err := fbuf.WriteFile(stripFileOrBufferPath(f), nil, 0666); err != nil {
				return "", err
			}
		} else {
			// Ensure we have the object for the empty blob.
			//
			// NOTE: This *almost* always produces the
			// gitutil.SHAEmptyBlob oid, but you can hack gitattributes to
			// make this return something else, and we want to avoid
			// making assumptions about your git repo that could ever be
			// violated.
			//
			// TODO(sqs): could optimize by leaving the dstSHA blank for
			// newly created files that we have nonzero edits for (we will
			// compute the dstSHA again for those files anyway).
			oid, err := gitRepo.HashObject("blob", stripFileOrBufferPath(f), nil)
			if err != nil {
				return "", err
			}

			// We will fill in the dstSHA below in Edit when we have the
			// file contents, if we have edits.
			if err := tree.ApplyChanges([]*gitutil.ChangedFile{{
				Status:  "A",
				SrcMode: gitutil.SHAAllZeros, DstMode: gitutil.RegularFileNonExecutableMode,
				SrcSHA: gitutil.SHAAllZeros, DstSHA: oid,
				SrcPath: stripFileOrBufferPath(f),
			}}); err != nil {
				return "", err
			}
		}
	}
	for _, f := range op.Delete {
		if isBufferPath(f) {
			if err := fbuf.Remove(stripFileOrBufferPath(f)); err != nil {
				return "", err
			}
		} else {
			mode, sha, err := gitRepo.FileInfoForPath(base, stripFileOrBufferPath(f))
			if err != nil {
				return "", err
			}
			if err := tree.ApplyChanges([]*gitutil.ChangedFile{{
				Status:  "D",
				SrcMode: mode, DstMode: gitutil.SHAAllZeros,
				SrcSHA: sha, DstSHA: gitutil.SHAAllZeros,
				SrcPath: stripFileOrBufferPath(f),
			}}); err != nil {
				return "", err
			}
		}
	}

	for _, f := range op.Truncate {
		if isBufferPath(f) {
			if err := fbuf.WriteFile(stripFileOrBufferPath(f), nil, 0666); err != nil {
				return "", err
			}
		} else {
			if err := updateGitFile(stripFileOrBufferPath(f), nil); err != nil {
				return "", err
			}
		}
	}
	for f, edits := range op.Edit {
		if len(edits) == 0 {
			continue
		}

		var data []byte
		var err error
		if _, created := created[f]; created {
			// no data yet
			data = []byte{}
		} else if isBufferPath(f) {
			data, err = fbuf.ReadFile(stripFileOrBufferPath(f))
		} else {
			f0 := fileOrigin(f)
			if !isFilePath(f0) {
				panic(fmt.Sprintf("not implemented: edit of a disk file %q derived from buffer file %q", f, f0))
			}
			data, _, _, err = gitRepo.ReadBlob(base, stripFileOrBufferPath(f0))
		}
		if err != nil {
			return "", err
		}

		doc := ot.Doc(string(data))
		if err := doc.Apply(edits); err != nil {
			err := zap.Errorf(zap.ErrorCodeInvalidOp, "apply OT edit to %s @ %s: %s (doc: %q, op: %v)", f, base, err, data, op)
			if panicOnSomeErrors {
				level.Error(logger).Log("PANIC-BELOW", "")
				panic(err)
			}
			return "", err
		}

		if isBufferPath(f) {
			if err := fbuf.WriteFile(stripFileOrBufferPath(f), []byte(string(doc)), 0666); err != nil {
				return "", err
			}
		} else {
			if err := updateGitFile(stripFileOrBufferPath(f), []byte(string(doc))); err != nil {
				return "", err
			}
		}
	}

	if tree == nil {
		return "", nil // indicates no new tree SHA was created
	}
	return gitRepo.CreateTree("", tree.Root)
}

func CreateWorktreeSnapshotCommit(ctx context.Context, gitRepo interface {
	MakeCommit(ctx context.Context, parent string, onlyIfChangedFiles bool) (string, []*gitutil.ChangedFile, error)
}, parent string) (string, []*gitutil.ChangedFile, error) {
	commitID, changes, err := gitRepo.MakeCommit(ctx, parent, true)
	if err != nil {
		return "", nil, err
	}
	if len(changes) == 0 {
		return parent, nil, nil
	}
	return commitID, changes, nil
}

// OpDependsOnGitTree reports whether applying op requires the Git
// tree to be known. Mere edits to buffered files do not require the
// Git tree. This function allows us to optimize elsewhere by not
// always reading in the full Git tree.
func OpDependsOnGitTree(op ot.WorkspaceOp) bool {
	if op.Noop() {
		return false
	}
	if len(op.Save) > 0 {
		return true
	}
	for src, dst := range op.Copy {
		if pathutil.IsFilePath(src) {
			return true
		}
		if pathutil.IsFilePath(dst) {
			return true
		}
	}
	for src, dst := range op.Rename {
		if pathutil.IsFilePath(src) {
			return true
		}
		if pathutil.IsFilePath(dst) {
			return true
		}
	}
	for _, f := range op.Create {
		if pathutil.IsFilePath(f) {
			return true
		}
	}
	for _, f := range op.Delete {
		if pathutil.IsFilePath(f) {
			return true
		}
	}
	for _, f := range op.Truncate {
		if pathutil.IsFilePath(f) {
			return true
		}
	}
	for f := range op.Edit {
		if pathutil.IsFilePath(f) {
			return true
		}
	}
	if op.GitHead != "" {
		return true
	}
	return false
}

// OpAffectsGitTree reports whether the op modifies the Git tree.
func OpAffectsGitTree(op ot.WorkspaceOp) bool {
	if op.Noop() {
		return false
	}
	if len(op.Save) > 0 {
		return true
	}
	for dst := range op.Copy {
		if pathutil.IsFilePath(dst) {
			return true
		}
	}
	for src, dst := range op.Rename {
		if pathutil.IsFilePath(src) {
			return true
		}
		if pathutil.IsFilePath(dst) {
			return true
		}
	}
	for _, f := range op.Create {
		if pathutil.IsFilePath(f) {
			return true
		}
	}
	for _, f := range op.Delete {
		if pathutil.IsFilePath(f) {
			return true
		}
	}
	for _, f := range op.Truncate {
		if pathutil.IsFilePath(f) {
			return true
		}
	}
	for f := range op.Edit {
		if pathutil.IsFilePath(f) {
			return true
		}
	}
	return false
}
