package dependencies

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Service encapsulates the resolution and persistence of dependencies at the repository and package levels.
type Service struct {
	store      store.Store
	operations *operations
}

func newService(observationCtx *observation.Context, store store.Store) *Service {
	return &Service{
		store:      store,
		operations: newOperations(observationCtx),
	}
}

type Repo = shared.Repo

type ListDependencyReposOpts struct {
	// Scheme is the moniker scheme to filter for e.g. 'gomod', 'npm' etc.
	Scheme string
	// Name is the package name to filter for e.g. '@types/node' etc.
	Name reposource.PackageName
	// After is the value predominantly used for pagination. When sorting by
	// newest first, this should be the ID of the last element in the previous
	// page, when excluding versions it should be the last package name in the
	// previous page.
	After any
	// Limit limits the size of the results set to be returned.
	Limit int
	// NewestFirst sorts by when a (package, version) was added to the list.
	// Incompatible with ExcludeVersions below.
	NewestFirst bool
	// ExcludeVersions returns one row for every package, instead of one for
	// every (package, version) tuple. Results will be sorted by name to make
	// pagination possible. Takes precedence over NewestFirst.
	ExcludeVersions bool
}

func (s *Service) ListDependencyRepos(ctx context.Context, opts ListDependencyReposOpts) (_ []Repo, err error) {
	ctx, _, endObservation := s.operations.listDependencyRepos.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.ListDependencyRepos(ctx, store.ListDependencyReposOpts(opts))
}

func (s *Service) UpsertDependencyRepos(ctx context.Context, deps []Repo) (_ []Repo, err error) {
	ctx, _, endObservation := s.operations.upsertDependencyRepos.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.UpsertDependencyRepos(ctx, deps)
}

func (s *Service) DeleteDependencyReposByID(ctx context.Context, ids ...int) (err error) {
	ctx, _, endObservation := s.operations.deleteDependencyReposByID.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteDependencyReposByID(ctx, ids...)
}
