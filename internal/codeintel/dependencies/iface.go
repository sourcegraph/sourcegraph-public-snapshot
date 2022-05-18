package dependencies

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/lockfiles"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

type localGitService interface {
	GetCommits(ctx context.Context, repoCommits []api.RepoCommit, ignoreErrors bool) ([]*gitdomain.Commit, error)
}

type GitService interface {
	localGitService
	lockfiles.GitService
}

type LockfilesService interface {
	ListDependencies(ctx context.Context, repo api.RepoName, rev string) ([]reposource.PackageDependency, error)
}

type Syncer interface {
	// Sync will lazily sync the repos that have been inserted into the database but have not yet been
	// cloned. See repos.Syncer.SyncRepo.
	Sync(ctx context.Context, repo api.RepoName) error
}
