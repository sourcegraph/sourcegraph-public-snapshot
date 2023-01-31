package dependencies

import (
	"context"

	"github.com/opentracing/opentracing-go/log"

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

type (
	PackageRepoReference  = shared.PackageRepoReference
	PackageRepoRefVersion = shared.PackageRepoRefVersion
	MinimalPackageRepoRef = shared.MinimalPackageRepoRef
)

type ListDependencyReposOpts struct {
	// Scheme is the moniker scheme to filter for e.g. 'gomod', 'npm' etc.
	Scheme string
	// Name is the package name to filter for e.g. '@types/node' etc.
	Name reposource.PackageName

	// ExactNameOnly enables exact name matching instead of substring.
	ExactNameOnly bool
	// After is the value predominantly used for pagination. When sorting by
	// newest first, this should be the ID of the last element in the previous
	// page, when excluding versions it should be the last package name in the
	// previous page.
	After int
	// Limit limits the size of the results set to be returned.
	Limit int
	// MostRecentlyUpdated sorts by when a package was updated (either created or
	// a new version added).
	MostRecentlyUpdated bool
}

func (s *Service) ListPackageRepoRefs(ctx context.Context, opts ListDependencyReposOpts) (_ []PackageRepoReference, total int, err error) {
	ctx, _, endObservation := s.operations.listPackageRepos.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("scheme", opts.Scheme),
		log.String("name", string(opts.Name)),
		log.Bool("exactOnly", opts.ExactNameOnly),
		log.Int("after", opts.After),
		log.Int("limit", opts.Limit),
		log.Bool("mostRecentlyUpdated", opts.MostRecentlyUpdated),
	}})
	defer endObservation(1, observation.Args{})

	return s.store.ListPackageRepoRefs(ctx, store.ListDependencyReposOpts(opts))
}

func (s *Service) InsertPackageRepoRefs(ctx context.Context, deps []MinimalPackageRepoRef) (_ []shared.PackageRepoReference, _ []shared.PackageRepoRefVersion, err error) {
	ctx, _, endObservation := s.operations.upsertPackageRepoRefs.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("packageRepoRefs", len(deps)),
	}})
	defer endObservation(1, observation.Args{})

	return s.store.InsertPackageRepoRefs(ctx, deps)
}

func (s *Service) DeletePackageRepoRefsByID(ctx context.Context, ids ...int) (err error) {
	ctx, _, endObservation := s.operations.deletePackageRepoRefsByID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("packageRepoRefs", len(ids)),
	}})
	defer endObservation(1, observation.Args{})

	return s.store.DeletePackageRepoRefsByID(ctx, ids...)
}

func (s *Service) DeletePackageRepoRefVersionsByID(ctx context.Context, ids ...int) (err error) {
	ctx, _, endObservation := s.operations.deletePackageRepoRefVersionsByID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("packageRepoRefVersions", len(ids)),
	}})
	defer endObservation(1, observation.Args{})

	return s.store.DeletePackageRepoRefVersionsByID(ctx, ids...)
}
