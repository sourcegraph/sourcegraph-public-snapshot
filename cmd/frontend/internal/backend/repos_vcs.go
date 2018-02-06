package backend

import (
	"context"
	"fmt"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"

	"github.com/pkg/errors"

	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repoupdater"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/gitcmd"
)

// RemoteVCS returns a handle to the underlying Git repository (on gitserver) that is cloned or updated on demand
// (if needed), after checking with the repository's original code host to obtain the latest Git remote URL for the
// repository.
//
// When to use RemoteVCS vs. CachedVCS?
//
// A caller should use RemoteVCS if it:
//  - prefers to wait (on a Git remote update) instead of fail if the Git repository is out of date
//  - needs to ensure that the repository still exists on the original code host
//
// A caller should use CachedVCS if it:
//  - prefers to fail instead of block (on a Git remote update) if the Git repository is out of date
//  - doesn't care if the repository still exists on the original code host
func (repos) RemoteVCS(ctx context.Context, repo *types.Repo) (vcs.Repository, error) {
	if Mocks.Repos.VCS != nil {
		return Mocks.Repos.VCS(repo.URI)
	}
	gitserverRepo, err := (repos{}).GitserverRepoInfo(ctx, repo)
	return (repos{}).VCS(gitserverRepo), err
}

// CachedVCS returns a handle to the underlying Git repository on gitserver, without attempting to check that the
// repository exists on its original code host or that gitserver's mirror is up to date.
//
// See (repos).RemoteVCS for guidance on when to use CachedVCS vs. RemoteVCS.
func (repos) CachedVCS(repo *types.Repo) vcs.Repository {
	if Mocks.Repos.VCS != nil {
		vcsrepo, err := Mocks.Repos.VCS(repo.URI)
		if err != nil {
			panic(fmt.Sprintf("CachedVCS: Mock.Repos.VCS(%q) returned error: %s", repo.URI, err))
		}
		return vcsrepo
	}
	gitserverRepo := quickGitserverRepoInfo(repo.URI)
	if gitserverRepo == nil {
		gitserverRepo = &gitserver.Repo{Name: repo.URI}
	}
	return (repos{}).VCS(*gitserverRepo)
}

// VCS returns a handle to the Git repository specified by repo. Callers, unless they already have a gitserver.Repo
// value, should use either RemoteVCS or CachedVCS instead of this method.
func (repos) VCS(repo gitserver.Repo) vcs.Repository {
	return gitcmd.Open(repo.Name, repo.URL)
}

func (repos) GitserverRepoInfo(ctx context.Context, repo *types.Repo) (gitserver.Repo, error) {
	if gitserverRepo := quickGitserverRepoInfo(repo.URI); gitserverRepo != nil {
		return *gitserverRepo, nil
	}

	result, err := repoupdater.DefaultClient.RepoLookup(ctx, protocol.RepoLookupArgs{
		Repo:         repo.URI,
		ExternalRepo: repo.ExternalRepo,
	})
	if err != nil {
		return gitserver.Repo{Name: repo.URI}, err
	}
	return gitserver.Repo{Name: result.Repo.URI, URL: result.Repo.VCS.URL}, nil
}

func quickGitserverRepoInfo(repo api.RepoURI) *gitserver.Repo {
	// If it is possible to 100% correctly determine it statically, use a fast path. This is
	// used to avoid a RepoLookup call for public GitHub.com and GitLab.com repositories
	// (especially on Sourcegraph.com), which reduces rate limit pressure significantly.
	//
	// This fails for private repositories, which require authentication in the URL userinfo.

	switch {
	case strings.HasPrefix(strings.ToLower(string(repo)), "github.com/"):
		if envvar.SourcegraphDotComMode() || !conf.HasGitHubDotComToken() {
			return &gitserver.Repo{Name: repo, URL: "https://" + string(repo)}
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
// * Commit does not exist: vcs.ErrRevisionNotFound
// * Empty repository: vcs.ErrRevisionNotFound
// * The user does not have permission: errcode.IsNotFound
// * Other unexpected errors.
func (s *repos) ResolveRev(ctx context.Context, repo *types.Repo, rev string) (commitID api.CommitID, err error) {
	if Mocks.Repos.ResolveRev != nil {
		return Mocks.Repos.ResolveRev(ctx, repo, rev)
	}

	ctx, done := trace(ctx, "Repos", "ResolveRev", map[string]interface{}{"repo": repo.URI, "rev": rev}, &err)
	defer done()

	// Try to get latest remote URL, but continue even if that fails.
	vcsrepo, err := Repos.RemoteVCS(ctx, repo)
	if err != nil && !isIgnorableRepoUpdaterError(err) {
		return "", err
	}
	return vcsrepo.ResolveRevision(ctx, rev, nil)
}

func (s *repos) GetCommit(ctx context.Context, repo *types.Repo, commitID api.CommitID) (res *vcs.Commit, err error) {
	if Mocks.Repos.GetCommit != nil {
		return Mocks.Repos.GetCommit(ctx, repo, commitID)
	}

	ctx, done := trace(ctx, "Repos", "GetCommit", map[string]interface{}{"repo": repo.URI, "commitID": commitID}, &err)
	defer done()

	log15.Debug("svc.local.repos.GetCommit", "repo", repo.URI, "commitID", commitID)

	if !isAbsCommitID(commitID) {
		return nil, errors.Errorf("non-absolute CommitID for Repos.GetCommit: %v", commitID)
	}

	// Try to get latest remote URL, but continue even if that fails.
	vcsrepo, err := Repos.RemoteVCS(ctx, repo)
	if err != nil && !isIgnorableRepoUpdaterError(err) {
		return nil, err
	}
	return vcsrepo.GetCommit(ctx, commitID)
}

func isIgnorableRepoUpdaterError(err error) bool {
	return errors.Cause(err) == repoupdater.ErrNotFound || errors.Cause(err) == repoupdater.ErrUnauthorized
}

func isAbsCommitID(commitID api.CommitID) bool { return len(commitID) == 40 }
