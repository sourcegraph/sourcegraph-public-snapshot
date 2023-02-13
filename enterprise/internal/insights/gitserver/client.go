package gitserver

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

func NewGitCommitClient() *GitCommitClient {
	return &GitCommitClient{
		cachedFirstCommit: NewCachedGitFirstEverCommit(),
		Gitclient:         gitserver.NewClient(),
	}
}

type GitCommitClient struct {
	cachedFirstCommit *CachedGitFirstEverCommit
	Gitclient         gitserver.Client
}

func (g *GitCommitClient) FirstCommit(ctx context.Context, repoName api.RepoName) (*gitdomain.Commit, error) {
	return g.cachedFirstCommit.GitFirstEverCommit(ctx, repoName)
}
func (g *GitCommitClient) RecentCommits(ctx context.Context, repoName api.RepoName, target time.Time, revision string) ([]*gitdomain.Commit, error) {
	options := gitserver.CommitsOptions{N: 1, Before: target.Format(time.RFC3339), DateOrder: true}
	if len(revision) > 0 {
		options.Range = revision
	}
	return g.Gitclient.Commits(ctx, authz.DefaultSubRepoPermsChecker, repoName, options)
}
