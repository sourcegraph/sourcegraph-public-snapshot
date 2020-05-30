package gitserver

import (
	"context"
	"os"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

// FileExists determines whether a file exists in a particular commit of a repository.
func FileExists(ctx context.Context, db db.DB, repositoryID int, commit, file string) (bool, error) {
	repo, err := repositoryIDToRepo(ctx, db, repositoryID)
	if err != nil {
		return false, err
	}

	if _, err := git.ResolveRevision(ctx, repo, nil, commit, nil); err != nil {
		return false, errors.Wrap(err, "git.ResolveRevision")
	}

	if _, err := git.Stat(ctx, repo, api.CommitID(commit), file); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
