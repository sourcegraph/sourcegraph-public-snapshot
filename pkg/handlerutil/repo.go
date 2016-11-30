package handlerutil

import (
	"net/http"
	"strconv"
	"strings"

	"context"

	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
)

// RepoCommon holds all of the information necessary to render a
// repository page template. It is returned by GetRepoFromRequest. See also
// RepoRevCommon.
type RepoCommon struct {
	Repo       *sourcegraph.Repo
	RepoConfig *sourcegraph.RepoConfig
}

// RepoRevCommon holds all of the commit-specific information
// necessary to render a repository page template for a certain
// commit. It is returned by GetRepoAndRevFromRequest. It is assumed that pages
// rendered are also provided with repoCommon template data.
type RepoRevCommon struct {
	RepoRevSpec sourcegraph.RepoRevSpec
}

// GetRepoAndRevCommon returns the repository and RepoRevSpec based on
// the route vars. It may also return custom error types
// URLMovedError, NoVCSDataError, which callers should ideally check
// for.
func GetRepoAndRevCommon(ctx context.Context, vars map[string]string) (rc *RepoCommon, vc *RepoRevCommon, err error) {
	rc, err = GetRepoCommon(ctx, vars)
	if err != nil {
		return
	}

	vc = &RepoRevCommon{}
	vc.RepoRevSpec.Repo = rc.Repo.ID

	vc.RepoRevSpec, err = getRepoRev(ctx, vars, rc.Repo.ID, rc.Repo.DefaultBranch)
	if err != nil {
		cloneInProgress := legacyerr.ErrCode(err) == legacyerr.Unavailable && legacyerr.ErrorDesc(err) == vcs.RepoNotExistError{CloneInProgress: true}.Error()
		if noVCSData := legacyerr.ErrCode(err) == legacyerr.NotFound ||
			cloneInProgress ||
			strings.Contains(err.Error(), "has no default branch"); noVCSData {

			if cloneInProgress {
				return rc, vc, err
			} else if rev := vars["Rev"]; rev != "" && rev != "@" {
				err = vcs.ErrRevisionNotFound
			} else {
				err = &NoVCSDataError{RepoCommon: rc}
			}
		}
		return
	}

	return
}

// GetRepoCommon returns the repository and RepoSpec based on the
// route vars. Callers should ideally handle the custom error type
// URLMovedError.
func GetRepoCommon(ctx context.Context, vars map[string]string) (rc *RepoCommon, err error) {
	rc = &RepoCommon{}
	rc.Repo, err = GetRepo(ctx, vars)
	if err != nil {
		return
	}

	rc.RepoConfig, err = backend.Repos.GetConfig(ctx, &sourcegraph.RepoSpec{ID: rc.Repo.ID})
	return
}

// GetRepoID returns the Sourcegraph repository ID based on the route vars.
// If the repo path string is a numeric ID, it is returned immediately.
// Otherwise the repository ID is resolved via backend.Repos.Resolve.
// If the canonical path differs, a URLMovedError error is returned.
func GetRepoID(ctx context.Context, vars map[string]string) (int32, error) {
	origRepo := routevar.ToRepo(vars)

	// If the URL contains just a numeric ID, then just return that
	// without incurring a lookup. This does not check for the
	// existence of the repo, but the backend API will effectively
	// perform that check.
	id, err := strconv.Atoi(origRepo)
	if err == nil {
		return int32(id), nil
	}

	res, err := backend.Repos.Resolve(ctx, &sourcegraph.RepoResolveOp{Path: origRepo})
	if err != nil {
		return 0, err
	}

	// Check for redirect.
	if origRepo != "" && res.CanonicalPath != "" && origRepo != res.CanonicalPath {
		return 0, &URLMovedError{res.CanonicalPath}
	}

	return res.Repo, nil
}

// GetRepo gets the repo (from the reposSvc) specified in the URL's
// RepoSpec route param. Callers should ideally check for a return error of type
// URLMovedError and handle this scenario by warning or redirecting the user.
func GetRepo(ctx context.Context, vars map[string]string) (repo *sourcegraph.Repo, err error) {
	repoID, err := GetRepoID(ctx, vars)
	if err != nil {
		return nil, err
	}

	return backend.Repos.Get(ctx, &sourcegraph.RepoSpec{ID: repoID})
}

// getRepoRev resolves the RepoRevSpec and commit specified in the
// route vars. The provided defaultBranch is used if no rev is
// specified in the URL.
func getRepoRev(ctx context.Context, vars map[string]string, repoID int32, defaultRev string) (sourcegraph.RepoRevSpec, error) {
	repoRev := routevar.ToRepoRev(vars)
	if repoRev.Rev == "" {
		repoRev.Rev = defaultRev

		if repoRev.Rev == "" {
			log15.Warn("getRepoRev: no rev specified and repo has no default rev", "repo", repoRev.Repo)
		}
	}

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

	repoRevSpec, err = getRepoRev(ctx, vars, repo.ID, repo.DefaultBranch)
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
