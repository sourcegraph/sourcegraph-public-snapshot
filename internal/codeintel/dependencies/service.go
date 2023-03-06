package dependencies

import (
	"context"
	"fmt"

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
	PackageRepoReference         = shared.PackageRepoReference
	PackageRepoRefVersion        = shared.PackageRepoRefVersion
	MinimalPackageRepoRef        = shared.MinimalPackageRepoRef
	MinimialVersionedPackageRepo = shared.MinimialVersionedPackageRepo
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
	// IncludeBlocked also includes those that would not be synced due to filter rules
	IncludeBlocked bool
}

func (s *Service) ListPackageRepoRefs(ctx context.Context, opts ListDependencyReposOpts) (_ []PackageRepoReference, total int, err error) {
	ctx, _, endObservation := s.operations.listPackageRepos.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("scheme", opts.Scheme),
		log.String("name", string(opts.Name)),
		log.Bool("exactOnly", opts.ExactNameOnly),
		log.Int("after", opts.After),
		log.Int("limit", opts.Limit),
	}})
	defer endObservation(1, observation.Args{})

	return s.store.ListPackageRepoRefs(ctx, store.ListDependencyReposOpts(opts))
}

func (s *Service) InsertPackageRepoRefs(ctx context.Context, deps []MinimalPackageRepoRef) (_ []shared.PackageRepoReference, _ []shared.PackageRepoRefVersion, err error) {
	ctx, _, endObservation := s.operations.insertPackageRepoRefs.With(ctx, &err, observation.Args{LogFields: []log.Field{
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

type ListPackageRepoRefFiltersOpts struct {
	IDs            []int
	PackageScheme  string
	IncludeDeleted bool
	After          int
	Limit          int
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (s *Service) ListPackageRepoRefFilters(ctx context.Context, opts ListPackageRepoRefFiltersOpts) (_ []shared.PackageFilter, err error) {
	ctx, _, endObservation := s.operations.listPackageRepoFilters.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("numPackageRepoFilterIDs", len(opts.IDs)),
		log.String("packageScheme", opts.PackageScheme),
		log.Int("after", opts.After),
		log.Int("limit", opts.Limit),
	}})
	defer endObservation(1, observation.Args{})
	return s.store.ListPackageRepoRefFilters(ctx, store.ListPackageRepoRefFiltersOpts(opts))
}

func (s *Service) CreatePackageRepoFilter(ctx context.Context, filter shared.MinimalPackageFilter) (err error) {
	ctx, _, endObservation := s.operations.createPackageRepoFilter.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("packageScheme", filter.PackageScheme),
		log.String("behaviour", deref(filter.Behaviour)),
		log.String("versionFilter", fmt.Sprintf("%+v", filter.VersionFilter)),
		log.String("nameFilter", fmt.Sprintf("%+v", filter.NameFilter)),
	}})
	defer endObservation(1, observation.Args{})
	return s.store.CreatePackageRepoFilter(ctx, filter)
}

func (s *Service) UpdatePackageRepoFilter(ctx context.Context, filter shared.PackageFilter) (err error) {
	ctx, _, endObservation := s.operations.updatePackageRepoFilter.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", filter.ID),
		log.String("packageScheme", filter.PackageScheme),
		log.String("behaviour", filter.Behaviour),
		log.String("versionFilter", fmt.Sprintf("%+v", filter.VersionFilter)),
		log.String("nameFilter", fmt.Sprintf("%+v", filter.NameFilter)),
	}})
	defer endObservation(1, observation.Args{})
	return s.store.UpdatePackageRepoFilter(ctx, filter)
}

func (s *Service) DeletePackageRepoFilter(ctx context.Context, id int) (err error) {
	ctx, _, endObservation := s.operations.deletePackageRepoFilter.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})
	return s.store.DeletePacakgeRepoFilter(ctx, id)
}

func (s *Service) IsPackageRepoVersionAllowed(ctx context.Context, scheme string, pkg reposource.PackageName, version string) (allowed bool, err error) {
	ctx, _, endObservation := s.operations.isPackageRepoVersionAllowed.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("packageScheme", scheme),
		log.String("name", string(pkg)),
		log.String("version", version),
	}})
	defer endObservation(1, observation.Args{})
	return s.store.IsPackageRepoVersionAllowed(ctx, scheme, pkg, version)
}

func (s *Service) IsPackageRepoAllowed(ctx context.Context, scheme string, pkg reposource.PackageName) (allowed bool, err error) {
	ctx, _, endObservation := s.operations.isPackageRepoAllowed.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("packageScheme", scheme),
		log.String("name", string(pkg)),
	}})
	defer endObservation(1, observation.Args{})
	return s.store.IsPackageRepoAllowed(ctx, scheme, pkg)
}

func (s *Service) PackagesOrVersionsMatchingFilter(ctx context.Context, filter shared.MinimalPackageFilter, limit, after int) (_ []shared.PackageRepoReference, _ int, err error) {
	ctx, _, endObservation := s.operations.pkgsOrVersionsMatchingFilter.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("packageScheme", filter.PackageScheme),
		log.String("versionFilter", fmt.Sprintf("%+v", filter.VersionFilter)),
		log.String("nameFilter", fmt.Sprintf("%+v", filter.NameFilter)),
	}})
	defer endObservation(1, observation.Args{})
	return s.store.PackagesOrVersionsMatchingFilter(ctx, filter, limit, after)
}
