package backend

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

// CachedGitRepo returns a handle to the Git repository that does not know the remote URL. If
// knowing the remote URL is necessary to perform any operations (from method calls on the return
// value), those operations will fail. This occurs when the repository isn't cloned on gitserver or
// when an update is needed (eg in ResolveRevision).
func CachedGitRepo(repo *types.Repo) gitserver.Repo {
	if r := quickGitserverRepo(repo.URI); r != nil {
		return *r
	}
	return gitserver.Repo{Name: repo.URI}
}

// GitRepo returns a handle to the Git repository with the up-to-date (as of the time of this call)
// remote URL. See CachedGitRepo for when this is necessary vs. unnecessary.
func GitRepo(ctx context.Context, repo *types.Repo) (gitserver.Repo, error) {
	if gitserverRepo := quickGitserverRepo(repo.URI); gitserverRepo != nil {
		return *gitserverRepo, nil
	}

	result, err := repoupdater.DefaultClient.RepoLookup(ctx, protocol.RepoLookupArgs{
		Repo:         repo.URI,
		ExternalRepo: repo.ExternalRepo,
	})
	if err != nil {
		return gitserver.Repo{Name: repo.URI}, err
	}
	if result.Repo == nil {
		return gitserver.Repo{Name: repo.URI}, repoupdater.ErrNotFound
	}
	return gitserver.Repo{Name: result.Repo.URI, URL: result.Repo.VCS.URL}, nil
}

func quickGitserverRepo(repo api.RepoURI) *gitserver.Repo {
	// If it is possible to 100% correctly determine it statically, use a fast path. This is
	// used to avoid a RepoLookup call for public GitHub.com and GitLab.com repositories
	// (especially on Sourcegraph.com), which reduces rate limit pressure significantly.
	//
	// This fails for private repositories, which require authentication in the URL userinfo.

	switch {
	case strings.HasPrefix(strings.ToLower(string(repo)), "github.com/"):
		if envvar.SourcegraphDotComMode() || !conf.HasGitHubDotComToken() {
			return &gitserver.Repo{Name: repo, URL: "https://" + string(repo) + ".git"}
		}

	case strings.HasPrefix(strings.ToLower(string(repo)), "gitlab.com/"):
		if envvar.SourcegraphDotComMode() || !conf.HasGitLabDotComToken() {
			return &gitserver.Repo{Name: repo, URL: "https://" + string(repo) + ".git"}
		}
	}

	// Fall back to performing full RepoLookup, which will hit the code host.
	return nil
}

// ResolveRev will return the absolute commit for a commit-ish spec in a repo.
// If no rev is specified, HEAD is used.
// Error cases:
// * Repo does not exist: vcs.RepoNotExistError
// * Commit does not exist: git.RevisionNotFoundError
// * Empty repository: git.RevisionNotFoundError
// * The user does not have permission: errcode.IsNotFound
// * Other unexpected errors.
func (s *repos) ResolveRev(ctx context.Context, repo *types.Repo, rev string) (commitID api.CommitID, err error) {
	if Mocks.Repos.ResolveRev != nil {
		return Mocks.Repos.ResolveRev(ctx, repo, rev)
	}

	ctx, done := trace(ctx, "Repos", "ResolveRev", map[string]interface{}{"repo": repo.URI, "rev": rev}, &err)
	defer done()

	// Try to get latest remote URL, but continue even if that fails.
	grepo, err := GitRepo(ctx, repo)
	maybeLogRepoUpdaterError(repo, err)
	if err != nil && !isIgnorableRepoUpdaterError(err) {
		return "", err
	}
	return git.ResolveRevision(ctx, grepo, nil, rev, nil)
}

func (s *repos) GetCommit(ctx context.Context, repo *types.Repo, commitID api.CommitID) (res *git.Commit, err error) {
	if Mocks.Repos.GetCommit != nil {
		return Mocks.Repos.GetCommit(ctx, repo, commitID)
	}

	ctx, done := trace(ctx, "Repos", "GetCommit", map[string]interface{}{"repo": repo.URI, "commitID": commitID}, &err)
	defer done()

	log15.Debug("svc.local.repos.GetCommit", "repo", repo.URI, "commitID", commitID)

	if !git.IsAbsoluteRevision(string(commitID)) {
		return nil, errors.Errorf("non-absolute CommitID for Repos.GetCommit: %v", commitID)
	}

	// Try to get latest remote URL, but continue even if that fails.
	gitserverRepo, err := GitRepo(ctx, repo)
	maybeLogRepoUpdaterError(repo, err)
	if err != nil && !isIgnorableRepoUpdaterError(err) {
		return nil, err
	}
	return git.GetCommit(ctx, gitserverRepo, commitID)
}

func isIgnorableRepoUpdaterError(err error) bool {
	err = errors.Cause(err)
	return err == repoupdater.ErrNotFound || err == repoupdater.ErrUnauthorized || err == repoupdater.ErrTemporarilyUnavailable
}

func maybeLogRepoUpdaterError(repo *types.Repo, err error) {
	var msg string
	switch c := errors.Cause(err); c {
	case repoupdater.ErrNotFound:
		msg = "Repository host reported a repository as not found. If this repository was deleted on its origin, the site admin must explicitly delete it on Sourcegraph."
	case repoupdater.ErrUnauthorized:
		msg = "Repository host rejected as unauthorized an attempt to retrieve a repository's metadata. Check the repository host credentials in site configuration."
	case repoupdater.ErrTemporarilyUnavailable:
		msg = "Repository host was temporarily unavailable while retrieving repository information."
	}
	if msg != "" {
		log15.Warn(msg+" Consult repo-updater logs for more information.", "repo", repo.URI)
	}
}
