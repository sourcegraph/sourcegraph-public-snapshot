package localstore

import (
	"context"
	"fmt"
	"strings"

	gogithub "github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
)

// ErrRepoNotFound indicates that the repo does not exist or that the user has no access to that
// repo. Those two cases are not differentiated to avoid leaking repo existence information.
var ErrRepoNotFound = legacyerr.Errorf(legacyerr.NotFound, "repo not found")

// verifyActorHasRepoURIAccess checks if the given actor is authorized to access
// the given repository with repoURI. The access check is performed by delegating
// the access check to external providers as necessary, based on the host of repoURI.
func verifyActorHasRepoURIAccess(ctx context.Context, actor *auth.Actor, method string, repoURI string) bool {
	if accesscontrol.Skip(ctx) {
		return true
	}

	switch {
	case strings.HasPrefix(strings.ToLower(repoURI), "github.com/"):
		// Perform GitHub repository authorization check by delegating to GitHub API.
		if _, err := github.ReposFromContext(ctx).Get(ctx, repoURI); err == nil {
			return true
		}
		return false

	default:
		// Unless something above explicitly grants access, by default, access is denied.
		// This is a safer default.
		return false
	}
}

// verifyUserHasReadAccessAll verifies checks if the current actor
// can access these repositories. This method implements a more
// efficient way of verifying permissions on a set of repositories.
// (Calling VerifyHasRepoAccess on each individual repository in a
// long list of repositories incurs too many GitHub API requests.)
// Unlike other authentication checking functions in this package,
// this function assumes that the list of repositories passed in has a
// correct `Private` field. This method does not incur a GitHub API
// call for public repositories.
//
// Unlike other access functions, this function does not return an
// error when there is a permission-denied error for one of the
// repositories. Instead, the first return value is the list of
// repositories to which access is allowed (input order is preserved).
// If permission was denied for any repository, this list will be
// shorter than the repos argument. If there is any error in
// determining the list of allowed repositories, the second return
// value will be non-nil error.
func verifyUserHasReadAccessAll(ctx context.Context, method string, repos []*sourcegraph.Repo) (allowed []*sourcegraph.Repo, err error) {
	if accesscontrol.Skip(ctx) {
		return repos, nil
	}

	hasPrivate := false
	for _, repo := range repos {
		if repo.Private {
			hasPrivate = true
			break
		}
	}

	privateGHRepoURIs := map[string]struct{}{}
	if hasPrivate {
		ghrepos, err := github.ListAllGitHubRepos(ctx, &gogithub.RepositoryListOptions{Type: "private"})
		if err != nil {
			return nil, fmt.Errorf("could not list all accessible GitHub repositories: %s", err)
		}
		for _, ghrepo := range ghrepos {
			privateGHRepoURIs[ghrepo.URI] = struct{}{}
		}
	}

	for _, repo := range repos {
		if repo.Private {
			if _, authorized := privateGHRepoURIs[repo.URI]; authorized {
				allowed = append(allowed, repo)
			}
		} else {
			allowed = append(allowed, repo)
		}
	}

	return allowed, nil
}
