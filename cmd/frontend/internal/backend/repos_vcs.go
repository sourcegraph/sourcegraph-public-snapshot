package backend

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	"gopkg.in/inconshreveable/log15.v2"
)

// CachedGitRepoTmp is like CachedGitRepo, but instead of returning a handle to the gitserver repo
// (the new way), it returns the *git.Repository struct with a bunch of VCS methods (the old
// way). It will be removed once all *git.Repository methods are unpeeled to funcs in package vcs.
func CachedGitRepoTmp(repo *types.Repo) *git.Repository {
	if r := quickGitserverRepo(repo.URI); r != nil {
		return git.Open(r.Name, r.URL)
	}
	return git.Open(repo.URI, "")
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
	gitserverRepo, err := GitRepo(ctx, repo)
	if err != nil {
		return "", err
	}
	if err != nil && !isIgnorableRepoUpdaterError(err) {
		return "", err
	}
	return git.Open(gitserverRepo.Name, gitserverRepo.URL).ResolveRevision(ctx, nil, rev, nil)
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
	if err != nil {
		return nil, err
	}
	if err != nil && !isIgnorableRepoUpdaterError(err) {
		return nil, err
	}
	return git.Open(gitserverRepo.Name, gitserverRepo.URL).GetCommit(ctx, commitID)
}

func isIgnorableRepoUpdaterError(err error) bool {
	return errors.Cause(err) == repoupdater.ErrNotFound || errors.Cause(err) == repoupdater.ErrUnauthorized
}
