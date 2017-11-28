package handlerutil

import (
	"net/http"

	"context"

	"github.com/gorilla/mux"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
)

// GetRepo gets the repo (from the reposSvc) specified in the URL's
// RepoSpec route param. Callers should ideally check for a return error of type
// URLMovedError and handle this scenario by warning or redirecting the user.
func GetRepo(ctx context.Context, vars map[string]string) (*sourcegraph.Repo, error) {
	origRepo := routevar.ToRepo(vars)

	repo, err := backend.Repos.GetByURI(ctx, origRepo)
	if err != nil {
		return nil, err
	}

	if origRepo != repo.URI {
		return nil, &URLMovedError{repo.URI}
	}

	return repo, nil
}

// getRepoRev resolves the RepoRevSpec and commit specified in the
// route vars.
func getRepoRev(ctx context.Context, vars map[string]string, repoID int32) (sourcegraph.RepoRevSpec, error) {
	repoRev := routevar.ToRepoRev(vars)
	res, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
		Repo: repoID,
		Rev:  repoRev.Rev,
	})
	if err != nil {
		return sourcegraph.RepoRevSpec{}, err
	}

	return sourcegraph.RepoRevSpec{Repo: repoID, CommitID: res.CommitID}, nil
}

// GetRepoAndRev returns the Repo and the RepoRevSpec for a repository. It may
// also return custom error URLMovedError to allow special handling of this case,
// such as for example redirecting the user.
func GetRepoAndRev(ctx context.Context, vars map[string]string) (repo *sourcegraph.Repo, repoRevSpec sourcegraph.RepoRevSpec, err error) {
	repo, err = GetRepo(ctx, vars)
	if err != nil {
		return repo, repoRevSpec, err
	}
	repoRevSpec.Repo = repo.ID

	repoRevSpec, err = getRepoRev(ctx, vars, repo.ID)
	return repo, repoRevSpec, err
}

// RedirectToNewRepoURI writes an HTTP redirect response with a
// Location that matches the request's location except with the
// RepoSpec route var updated to refer to newRepoURI (instead of the
// originally requested repo URI).
func RedirectToNewRepoURI(w http.ResponseWriter, r *http.Request, newRepoURI string) error {
	origVars := mux.Vars(r)
	origVars["Repo"] = newRepoURI

	var pairs []string
	for k, v := range origVars {
		pairs = append(pairs, k, v)
	}
	destURL, err := mux.CurrentRoute(r).URLPath(pairs...)
	if err != nil {
		return err
	}

	http.Redirect(w, r, destURL.String(), http.StatusMovedPermanently)
	return nil
}
