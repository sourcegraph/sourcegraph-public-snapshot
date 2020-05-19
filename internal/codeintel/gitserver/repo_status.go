package gitserver

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

func RepoStatus(repositoryID int, commit string) (cloneInProgress, notFound bool, err error) {
	repo, err := backend.Repos.Get(context.Background(), api.RepoID(repositoryID))
	if err != nil {
		return false, false, errors.Wrap(err, "backend.Repos.Get")
	}

	if _, err := backend.Repos.ResolveRev(context.Background(), repo, commit); err != nil{
		if vcs.IsCloneInProgress(err) {
			return true, false, nil
		}

		return false, false, errors.Wrap(err, "backend.Repos.ResolveRev")
	}

	return false, false, nil
}
