package git

import (
	"context"
	"io"

	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

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
