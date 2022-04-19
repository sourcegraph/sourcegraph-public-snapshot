package dependencies

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/lockfiles"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Store interface {
	ListDependencyRepos(ctx context.Context, opts store.ListDependencyReposOpts) ([]store.DependencyRepo, error)
	UpsertDependencyRepos(ctx context.Context, deps []store.DependencyRepo) ([]store.DependencyRepo, error)
}

type LockfilesService interface {
	ListDependencies(ctx context.Context, repo api.RepoName, rev string) ([]reposource.PackageDependency, error)
}

type GitService = lockfiles.GitService

type Syncer interface {
	// Sync will lazily sync the repos that have been inserted into the database but have not yet been
	// cloned. See repos.Syncer.SyncRepo.
	Sync(ctx context.Context, repo api.RepoName) error
}

// ErrorSyncer should be used when constructing a dependencies service that only uses the system-level
// behaviors (e.g., listing repositories to sync since the last request). This syncer value will issue
// errors on invocation indicating that gitserver/repoupdater services are not expected to invoke such
// methods.
var ErrorSyncer Syncer = &errorSyncer{}

type errorSyncer struct{}

func (s *errorSyncer) Sync(ctx context.Context, repo api.RepoName) error {
	return errors.Newf("codeintel/dependencies/Syncer should not be required by service methods called from gitserver or repoupdater")
}
