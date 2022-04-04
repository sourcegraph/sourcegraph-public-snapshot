package git

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

// DiffPath returns a position-ordered slice of changes (additions or deletions)
// of the given path between the given source and target commits.
func DiffPath(ctx context.Context, db database.DB, repo api.RepoName, sourceCommit, targetCommit, path string, checker authz.SubRepoPermissionChecker) ([]*diff.Hunk, error) {
	a := actor.FromContext(ctx)
	if hasAccess, err := authz.FilterActorPath(ctx, checker, a, repo, path); err != nil {
		return nil, err
	} else if !hasAccess {
		return nil, os.ErrNotExist
	}
	reader, err := execReader(ctx, db, repo, []string{"diff", sourceCommit, targetCommit, "--", path})
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	output, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if len(output) == 0 {
		return nil, nil
	}

	d, err := diff.NewFileDiffReader(bytes.NewReader(output)).Read()
	if err != nil {
		return nil, err
	}
	return d.Hunks, nil
}

// DiffSymbols performs a diff command which is expected to be parsed by our symbols package
func DiffSymbols(ctx context.Context, db database.DB, repo api.RepoName, commitA, commitB api.CommitID) ([]byte, error) {
	command := gitserver.NewClient(db).Command("git", "diff", "-z", "--name-status", "--no-renames", string(commitA), string(commitB))
	command.Repo = repo
	return command.Output(ctx)
}

type DiffFileIterator struct {
	rdr  io.ReadCloser
	mfdr *diff.MultiFileDiffReader
}

func (i *DiffFileIterator) Close() error {
	return i.rdr.Close()
}

// Next returns the next file diff. If no more diffs are available, the diff
// will be nil and the error will be io.EOF.
func (i *DiffFileIterator) Next() (*diff.FileDiff, error) {
	return i.mfdr.ReadFile()
}
