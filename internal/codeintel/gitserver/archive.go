package gitserver

import (
	"context"
	"io"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

// Archive retrieves a tar-formatted archive of the given commit.
func Archive(ctx context.Context, db db.DB, repositoryID int, commit string) (io.Reader, error) {
	repo, err := repositoryIDToRepo(ctx, db, repositoryID)
	if err != nil {
		return nil, err
	}

	if _, err := git.ResolveRevision(ctx, repo, nil, commit, nil); err != nil {
		return nil, errors.Wrap(err, "git.ResolveRevision")
	}

	return gitserver.DefaultClient.Archive(ctx, repo, gitserver.ArchiveOptions{
		Treeish: commit,
		Format:  "tar",
	})
}
