package dependencies

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/lockfiles"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

type Store interface {
	LockfileDependencies(ctx context.Context, repoName, commit string) ([]shared.PackageDependency, bool, error)
	UpsertLockfileDependencies(ctx context.Context, repoName, commit string, deps []shared.PackageDependency) error
	SelectRepoRevisionsToResolve(ctx context.Context, batchSize int, minimumCheckInterval time.Duration) (map[string][]string, error)
	UpdateResolvedRevisions(ctx context.Context, repoRevsToResolvedRevs map[string]map[string]string) error
	ListDependencyRepos(ctx context.Context, opts store.ListDependencyReposOpts) ([]Repo, error)
	UpsertDependencyRepos(ctx context.Context, deps []Repo) ([]Repo, error)
	DeleteDependencyReposByID(ctx context.Context, ids ...int) error
}

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
