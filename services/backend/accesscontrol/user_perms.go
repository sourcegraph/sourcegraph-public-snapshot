package accesscontrol

import (
	"context"
	"fmt"
	"strings"

	gogithub "github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
)

// ErrRepoNotFound indicates that the repo does not exist or that the user has no access to that
// repo. Those two cases are not differentiated to avoid leaking repo existence information.
var ErrRepoNotFound = legacyerr.Errorf(legacyerr.NotFound, "repo not found")

var Repos interface {
	// Get a repository.
	Get(ctx context.Context, repo int32) (*sourcegraph.Repo, error)

	// GetByURI a repository by its URI.
	GetByURI(ctx context.Context, repo string) (*sourcegraph.Repo, error)
}

// VerifyUserHasReadAccess checks if the user in the current context
// is authorized to make read requests to this server.
//
// This method always returns nil when the user has read access,
// and returns a non-nil error when access cannot be granted.
func VerifyUserHasReadAccess(ctx context.Context, method string, repoID int32) error {
	if mock(ctx) {
		return Mocks.VerifyUserHasReadAccess(ctx, method, repoID)
	} else if Skip(ctx) {
		return nil
	}

	if repoID != 0 {
		repo, err := Repos.Get(ctx, repoID)
		if err != nil {
			return err
		}
		// TODO: Repos.Get above already indirectly performs this access check, but outside of
		//       accesscontrol package, so it can't be relied on. Still, this is an opportunity
		//       to optimize, just need to refactor this in a better way.
		if repo.Private && !VerifyActorHasRepoURIAccess(ctx, auth.ActorFromContext(ctx), method, repo.URI) {
			return ErrRepoNotFound
		}
	}
	return nil
}

// VerifyActorHasRepoURIAccess checks if the given actor is authorized to access
// the given repository with repoURI. The access check is performed by delegating
// the access check to external providers as necessary, based on the host of repoURI.
// repoURI MUST begin with the hostname and not include schema. E.g., its value is
// like "github.com/user/repo" or "bitbucket.com/user/repo".
//
// NOTE: Only (*localstore.repos).Get/GetByURI method should call this
// func. All other callers should use
// Verify{User,Actor}Has{Read,Write}Access funcs. This func is
// specially designed to avoid infinite loops with
// (*localstore.repos).Get/GetByURI.
func VerifyActorHasRepoURIAccess(ctx context.Context, actor *auth.Actor, method string, repoURI string) bool {
	if mock(ctx) {
		return Mocks.VerifyActorHasRepoURIAccess(ctx, actor, method, repoURI)
	} else if Skip(ctx) {
		return true
	}

	switch {
	case strings.HasPrefix(strings.ToLower(repoURI), "github.com/"):
		// Perform GitHub repository authorization check by delegating to GitHub API.
		return verifyActorHasGitHubRepoAccess(ctx, actor, repoURI)

	default:
		// Unless something above explicitly grants access, by default, access is denied.
		// This is a safer default.
		return false
	}
}

// verifyActorHasGitHubRepoAccess checks if the given actor is authorized to access
// the given GitHub mirrored repository. repoURI MUST be of the form "github.com/user/repo",
// it MUST begin with "github.com/" (case insensitive). The access check is performed
// by delegating the access check to GitHub.
//
// NOTE: Only (*localstore.repos).Get/GetByURI method should call this
// func (indirectly, via VerifyActorHasRepoURIAccess). All other callers should use
// Verify{User,Actor}Has{Read,Write}Access funcs. This func is
// specially designed to avoid infinite loops with
// (*localstore.repos).Get/GetByURI.
//
// TODO: move to a security model that is more robust, readable, has
// better separation when dealing with multiple configurations, actor
// types, resource types and actions.
func verifyActorHasGitHubRepoAccess(ctx context.Context, actor *auth.Actor, repoURI string) bool {
	if repoURI == "" {
		panic("repoURI must be set")
	}
	if !strings.HasPrefix(strings.ToLower(repoURI), "github.com/") {
		panic(fmt.Errorf(`verifyActorHasGitHubRepoAccess: precondition not satisfied, repoURI %q does not begin with "github.com/" (case insensitive)`, repoURI))
	}

	if _, err := github.ReposFromContext(ctx).Get(ctx, repoURI); err == nil {
		return true
	}

	return false
}

// VerifyUserHasReadAccessAll verifies checks if the current actor
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
func VerifyUserHasReadAccessAll(ctx context.Context, method string, repos []*sourcegraph.Repo) (allowed []*sourcegraph.Repo, err error) {
	if mock(ctx) {
		return Mocks.VerifyUserHasReadAccessAll(ctx, method, repos)
	} else if Skip(ctx) {
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

// Allow skipping access checks when testing other packages.
type contextKey int

const insecureSkip contextKey = 0

// WithInsecureSkip skips all access checks performed using ctx or one
// of its descendants. It is INSECURE and should only be used during
// testing.
func WithInsecureSkip(ctx context.Context, skip bool) context.Context {
	return context.WithValue(ctx, insecureSkip, skip)
}

func Skip(ctx context.Context) bool {
	v, _ := ctx.Value(insecureSkip).(bool)
	return v
}

// Allow mocking of access control checks
const insecureMock contextKey = 1

// WithInsecureMock replaces all access checks with mocks. It
// supersedes WithInsecureSkip. It is INSECURE and should only be used
// during testing.
func WithInsecureMock(ctx context.Context, mock bool) context.Context {
	return context.WithValue(ctx, insecureMock, mock)
}

func mock(ctx context.Context) bool {
	v, _ := ctx.Value(insecureMock).(bool)
	return v
}
