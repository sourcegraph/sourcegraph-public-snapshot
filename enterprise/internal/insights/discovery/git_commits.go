package discovery

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

func NewGitCommitClient(db database.DB) *GitCommitClient {
	return &GitCommitClient{
		cachedFirstCommit: NewCachedGitFirstEverCommit(),
		db:                db,
	}
}

type GitCommitClient struct {
	cachedFirstCommit *CachedGitFirstEverCommit
	db                database.DB
}

func (g *GitCommitClient) FirstCommit(ctx context.Context, repoName api.RepoName) (*gitdomain.Commit, error) {
	return g.cachedFirstCommit.GitFirstEverCommit(ctx, g.db, repoName)
}
func (g *GitCommitClient) RecentCommits(ctx context.Context, repoName api.RepoName, target time.Time) ([]*gitdomain.Commit, error) {
	return gitserver.NewClient(g.db).Commits(ctx, repoName, gitserver.CommitsOptions{N: 1, Before: target.Format(time.RFC3339), DateOrder: true}, authz.DefaultSubRepoPermsChecker)
}
