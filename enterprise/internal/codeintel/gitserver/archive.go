package gitserver

import (
	"context"
	"io"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

// Archive retrieves a tar-formatted archive of the given commit.
func Archive(ctx context.Context, store store.Store, repositoryID int, commit string) (io.Reader, error) {
	repo, err := repositoryIDToRepo(ctx, store, repositoryID)
	if err != nil {
		return nil, err
	}

	if _, err := git.ResolveRevision(ctx, repo, nil, commit, git.ResolveRevisionOptions{}); err != nil {
		return nil, errors.Wrap(err, "git.ResolveRevision")
	}

	return gitserver.DefaultClient.Archive(ctx, repo, gitserver.ArchiveOptions{
		Treeish: commit,
		Format:  "tar",
	})
}
