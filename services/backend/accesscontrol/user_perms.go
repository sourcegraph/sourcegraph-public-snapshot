package accesscontrol

import (
	"context"
	"fmt"
	"strings"

	gogithub "github.com/sourcegraph/go-github/github"
	"golang.org/x/oauth2"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/google"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/google.golang.org/api/source/v1"
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

	// UnsafeDangerousGetByURI is like GetByURI except it *does not* consult
	// GitHub in order to determine if the user has access to the specified
	// repository. It is the caller's responsibility to ensure the returned
	// repo can be displayed to the user.
	UnsafeDangerousGetByURI(ctx context.Context, repo string) (*sourcegraph.Repo, error)
}

// VerifyUserHasReadAccess checks if the user in the current context
// is authorized to make write requests to this server.
//
// This method always returns nil when the user has write access,
// and returns a non-nil error when access cannot be granted.
// If the cmdline flag auth.restrict-write-access is set, this method
// will check if the authenticated user has admin privileges.
func VerifyUserHasReadAccess(ctx context.Context, method string, repo interface{}) error {
	return VerifyActorHasReadAccess(ctx, auth.ActorFromContext(ctx), method, repo)
}

// VerifyUserHasWriteAccess checks if the user in the current context
// is authorized to make write requests to this server.
//
// This method always returns nil when the user has write access,
// and returns a non-nil error when access cannot be granted.
// If the cmdline flag auth.restrict-write-access is set, this method
// will check if the authenticated user has admin privileges.
func VerifyUserHasWriteAccess(ctx context.Context, method string, repo interface{}) error {
	return VerifyActorHasWriteAccess(ctx, auth.ActorFromContext(ctx), method, repo)
}

// VerifyUserHasWriteAccess checks if the user in the current context
// is authorized to make admin requests to this server.
func VerifyUserHasAdminAccess(ctx context.Context, method string) error {
	return VerifyActorHasAdminAccess(ctx, auth.ActorFromContext(ctx), method)
}

// VerifyUserSelfOrAdmin checks if the user in the current context has
// the given uid, or if the actor has admin access on the server.
// This check should be used in cases where a request should succeed only
// if the request is for the user's own information, or if the ctx actor is an admin.
func VerifyUserSelfOrAdmin(ctx context.Context, method string, uid string) error {
	if uid != "" && auth.ActorFromContext(ctx).UID == uid {
		return nil
	}

	return VerifyUserHasAdminAccess(ctx, method)
}

// VerifyClientSelfOrAdmin checks if the client in the current context has
// the given id, or if the actor has admin access on the server.
// This check should be used in cases where a request should succeed only
// if the request is for the client's own information, or if the ctx actor is an admin.
func VerifyClientSelfOrAdmin(ctx context.Context, method string, clientID string) error {
	return VerifyUserHasAdminAccess(ctx, method)
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
func VerifyActorHasRepoURIAccess(ctx context.Context, actor *auth.Actor, method string, repoID int32, repoURI string) bool {
	if skip(ctx) {
		return true
	}

	switch {
	case strings.HasPrefix(strings.ToLower(repoURI), "github.com/"):
		// Perform GitHub repository authorization check by delegating to GitHub API.
		return verifyActorHasGitHubRepoAccess(ctx, actor, repoURI)

	case strings.HasPrefix(strings.ToLower(repoURI), "source.developers.google.com/p/"):
		// Perform GCP repository authorization check by delegating to GCP API.
		return VerifyActorHasGCPRepoAccess(ctx, actor, repoURI)

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

// VerifyActorHasGCPRepoAccess checks if the given actor is authorized to access
// the given GCP mirrored repository.
func VerifyActorHasGCPRepoAccess(ctx context.Context, actor *auth.Actor, repoURI string) bool {
	if !actor.GoogleConnected {
		return false
	}

	googleRefreshToken, err := auth.FetchGoogleRefreshToken(ctx, actor.UID)
	if err != nil {
		return false
	}
	client := google.Default.Client(ctx, &oauth2.Token{
		RefreshToken: googleRefreshToken.Token,
	})
	s, err := source.New(client)
	if err != nil {
		return false
	}

	// Parse "source.developers.google.com/p/projectID/r/repoName" repoURI format.
	els := strings.SplitN(repoURI, "/", 6)
	if len(els) != 5 { // It's expected to have exactly 5 elements.
		return false
	}
	projectID := els[2] // projectID is at index 2.
	repoName := els[4]  // repoName is at index 4.

	if _, err := s.Projects.Repos.Get(projectID, repoName).Do(); err == nil {
		return true
	}

	return false
}

func getRepo(ctx context.Context, repoIDOrURI interface{}) (repoID int32, repoURI string, err error) {
	repoURI, _ = repoIDOrURI.(string)
	repoID, _ = repoIDOrURI.(int32)

	var repoObj *sourcegraph.Repo
	if repoID != 0 && repoURI == "" {
		repoObj, err = Repos.Get(ctx, repoID)
	} else if repoURI != "" && repoID == 0 {
		repoObj, err = Repos.GetByURI(ctx, repoURI)
	}
	if err != nil {
		return
	}
	if repoObj != nil {
		repoID = repoObj.ID
		repoURI = repoObj.URI
	}
	return
}

// VerifyActorHasReadAccess checks if the given actor is authorized to make
// read requests to this server.
//
// Note that this function allows the caller to retrieve any user's
// access levels.  This is meant for trusted server code living
// outside the scope of a user request to verify user permissions. For
// all other cases, VerifyUserHasWriteAccess or
// VerifyUserHasAdminAccess should be used to authorize a user.
func VerifyActorHasReadAccess(ctx context.Context, actor *auth.Actor, method string, repo interface{}) error {
	if skip(ctx) {
		return nil
	}

	if repo != nil {
		repoID, repoURI, err := getRepo(ctx, repo)
		if err != nil {
			return err
		}
		// TODO: getRepo above already indirectly performs this access check, but outside of
		//       accesscontrol package, so it can't be relied on. Still, this is an opportunity
		//       to optimize, just need to refactor this in a better way.
		if !VerifyActorHasRepoURIAccess(ctx, actor, method, repoID, repoURI) {
			return ErrRepoNotFound
		}
	}
	return nil
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
	if skip(ctx) {
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

// VerifyUserHasReadAccessToDefRepoRefs filters out any DefRepoRefs which the
// user does not have access to (e.g. ones originating from private repos). It
// consults with the Postgres DB (Repos.Get) and GitHub (via VerifyUserHasReadAccessAll)
// in order to quickly filter the refs. The returned list is the ones the user
// has access to.
func VerifyUserHasReadAccessToDefRepoRefs(ctx context.Context, method string, repoRefs []*sourcegraph.DeprecatedDefRepoRef) ([]*sourcegraph.DeprecatedDefRepoRef, error) {
	// Build a list of repos that we must check for access.
	repos := make([]*sourcegraph.Repo, 0, len(repoRefs))
	for _, r := range repoRefs {
		// SECURITY: We must get the entire repo object by URI, or else we
		// would be missing the Private field thus making all private repos
		// public in the eyes of VerifyUserHasReadAccessAll. We do not use
		// standard Repos.GetByURI as that would perform a GitHub API request
		// for each repo.
		repo, err := Repos.UnsafeDangerousGetByURI(ctx, r.Repo)
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}

	// Perform the access check.
	allowed, err := VerifyUserHasReadAccessAll(ctx, method, repos)
	if err != nil {
		return nil, err
	}

	// Create a map of repo URIs we can access.
	allowedURIs := make(map[string]struct{}, len(allowed))
	for _, allowed := range allowed {
		allowedURIs[allowed.URI] = struct{}{}
	}

	// Formulate the final list of repoRefs the user can access.
	final := make([]*sourcegraph.DeprecatedDefRepoRef, 0, len(repoRefs))
	for _, repoRef := range repoRefs {
		if _, allowed := allowedURIs[repoRef.Repo]; !allowed {
			continue
		}
		final = append(final, repoRef)
	}
	return final, nil
}

// VerifyActorHasWriteAccess checks if the given actor is authorized to make
// write requests to this server.
//
// Note that this function allows the caller to retrieve any user's
// access levels.  This is meant for trusted server code living
// outside the scope of user requests to verify user permissions. For
// all other cases, VerifyUserHasWriteAccess should be used to
// authorize a user.
func VerifyActorHasWriteAccess(ctx context.Context, actor *auth.Actor, method string, repo interface{}) error {
	if skip(ctx) {
		return nil
	}

	repoID, repoURI, err := getRepo(ctx, repo)
	if err != nil {
		return err
	}

	// TODO: redesign the permissions model to avoid short-circuited "return nil"s.
	// (because it makes modifying authorization logic more error-prone.)

	if !actor.IsAuthenticated() {
		return legacyerr.Errorf(legacyerr.Unauthenticated, "write operation (%s) denied: not authenticated", method)
	}

	if !inAuthenticatedWriteWhitelist(method) {
		return legacyerr.Errorf(legacyerr.PermissionDenied, "write operation (%s) denied: user does not have write access", method)
	}

	if repoID != 0 && repoURI != "" {
		// TODO: getRepo above already indirectly performs this access check, but outside of
		//       accesscontrol package, so it can't be relied on. Still, this is an opportunity
		//       to optimize, just need to refactor this in a better way.
		if !VerifyActorHasRepoURIAccess(ctx, actor, method, repoID, repoURI) {
			return ErrRepoNotFound
		}
	}

	return nil
}

// VerifyActorHasAdminAccess checks if the given actor is authorized to make
// admin requests to this server.
//
// Note that this function allows the caller to retrieve any user's
// access levels.  This is meant for trusted server code living
// outside the scope of user requests to verify user permissions. For
// all other cases, VerifyUserHasAdminAccess should be used to
// authorize a user.
func VerifyActorHasAdminAccess(ctx context.Context, actor *auth.Actor, method string) error {
	if skip(ctx) {
		return nil
	}

	if !actor.IsAuthenticated() {
		return legacyerr.Errorf(legacyerr.Unauthenticated, "admin operation (%s) denied: not authenticated", method)
	}

	return legacyerr.Errorf(legacyerr.PermissionDenied, "admin operation (%s) denied: not authorized", method)
}

// inAuthenticatedWriteWhitelist reports if we always allow write access
// for method to any authenticated user.
func inAuthenticatedWriteWhitelist(method string) bool {
	switch method {
	case "MirrorRepos.cloneRepo":
		// This is used for read-only users to be able to trigger mirror clones
		// of public repositories, effectively "enabling" that repository.
		return true
	default:
		return false
	}
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

func skip(ctx context.Context) bool {
	v, _ := ctx.Value(insecureSkip).(bool)
	return v
}
