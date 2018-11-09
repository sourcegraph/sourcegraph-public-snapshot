package handlerutil

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
	"github.com/sourcegraph/sourcegraph/pkg/routevar"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
)

// GetRepo gets the repo (from the reposSvc) specified in the URL's
// Repo route param. Callers should ideally check for a return error of type
// URLMovedError and handle this scenario by warning or redirecting the user.
func GetRepo(ctx context.Context, vars map[string]string) (*types.Repo, error) {
	origRepo := routevar.ToRepo(vars)
	repo, err := backend.Repos.GetByName(ctx, origRepo)
	if e, ok := err.(backend.ErrRepoSeeOther); ok {
		return nil, &URLMovedError{NewURL: e.RedirectURL}
	}
	if errcode.IsNotFound(err) || errors.Cause(err) == repoupdater.ErrNotFound {
		return nil, &vcs.RepoNotExistError{Repo: origRepo, CloneInProgress: false}
	}
	if err != nil {
		return nil, err
	}
	if origRepo != repo.Name {
		return nil, &URLMovedError{NewURL: "/" + string(repo.Name)}
	}
	if repo.Name == "github.com/sourcegraphtest/Always500Test" {
		return nil, errors.New("error caused by Always500Test repo name")
	}
	return repo, nil
}

// getRepoRev resolves the repository and commit specified in the route vars.
func getRepoRev(ctx context.Context, vars map[string]string, repoID api.RepoID) (api.RepoID, api.CommitID, error) {
	repoRev := routevar.ToRepoRev(vars)
	repo, err := backend.Repos.Get(ctx, repoID)
	if err != nil {
		return repoID, "", err
	}
	commitID, err := backend.Repos.ResolveRev(ctx, repo, repoRev.Rev)
	if err != nil {
		return repoID, "", err
	}

	return repoID, commitID, nil
}

// GetRepoAndRev returns the repo object and the commit ID for a repository. It may
// also return custom error URLMovedError to allow special handling of this case,
// such as for example redirecting the user.
func GetRepoAndRev(ctx context.Context, vars map[string]string) (*types.Repo, api.CommitID, error) {
	repo, err := GetRepo(ctx, vars)
	if err != nil {
		return repo, "", err
	}

	_, commitID, err := getRepoRev(ctx, vars, repo.ID)
	return repo, commitID, err
}
