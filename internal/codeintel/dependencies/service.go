package dependencies

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/packagefilters"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	MinimalPackageRepoRefVersion = shared.MinimalPackageRepoRefVersion
	PackageRepoFilter            = shared.PackageRepoFilter
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

func (s *Service) ListPackageRepoRefs(ctx context.Context, opts ListDependencyReposOpts) (_ []PackageRepoReference, total int, hasMore bool, err error) {
	ctx, _, endObservation := s.operations.listPackageRepos.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("scheme", opts.Scheme),
		attribute.String("name", string(opts.Name)),
		attribute.Bool("exactOnly", opts.ExactNameOnly),
		attribute.Int("after", opts.After),
		attribute.Int("limit", opts.Limit),
	}})
	defer endObservation(1, observation.Args{})

	storeopts := store.ListDependencyReposOpts{
		Scheme:         opts.Scheme,
		Name:           opts.Name,
		After:          opts.After,
		Limit:          opts.Limit,
		IncludeBlocked: opts.IncludeBlocked,
	}

	if opts.ExactNameOnly {
		storeopts.Fuzziness = store.FuzzinessExactMatch
	} else {
		storeopts.Fuzziness = store.FuzzinessWildcard
	}

	return s.store.ListPackageRepoRefs(ctx, storeopts)
}

func (s *Service) InsertPackageRepoRefs(ctx context.Context, deps []MinimalPackageRepoRef) (_ []shared.PackageRepoReference, _ []shared.PackageRepoRefVersion, err error) {
	ctx, _, endObservation := s.operations.insertPackageRepoRefs.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("packageRepoRefs", len(deps)),
	}})
	defer endObservation(1, observation.Args{})

	return s.store.InsertPackageRepoRefs(ctx, deps)
}

func (s *Service) DeletePackageRepoRefsByID(ctx context.Context, ids ...int) (err error) {
	ctx, _, endObservation := s.operations.deletePackageRepoRefsByID.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("packageRepoRefs", len(ids)),
	}})
	defer endObservation(1, observation.Args{})

	return s.store.DeletePackageRepoRefsByID(ctx, ids...)
}

func (s *Service) DeletePackageRepoRefVersionsByID(ctx context.Context, ids ...int) (err error) {
	ctx, _, endObservation := s.operations.deletePackageRepoRefVersionsByID.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("packageRepoRefVersions", len(ids)),
	}})
	defer endObservation(1, observation.Args{})

	return s.store.DeletePackageRepoRefVersionsByID(ctx, ids...)
}

type ListPackageRepoRefFiltersOpts struct {
	IDs            []int
	PackageScheme  string
	Behaviour      string
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

func (s *Service) ListPackageRepoFilters(ctx context.Context, opts ListPackageRepoRefFiltersOpts) (_ []shared.PackageRepoFilter, hasMore bool, err error) {
	ctx, _, endObservation := s.operations.listPackageRepoFilters.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("numPackageRepoFilterIDs", len(opts.IDs)),
		attribute.String("packageScheme", opts.PackageScheme),
		attribute.Int("after", opts.After),
		attribute.Int("limit", opts.Limit),
		attribute.String("behaviour", opts.Behaviour),
	}})
	defer endObservation(1, observation.Args{})
	return s.store.ListPackageRepoRefFilters(ctx, store.ListPackageRepoRefFiltersOpts(opts))
}

func (s *Service) CreatePackageRepoFilter(ctx context.Context, input shared.MinimalPackageFilter) (filter *shared.PackageRepoFilter, err error) {
	ctx, _, endObservation := s.operations.createPackageRepoFilter.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("packageScheme", input.PackageScheme),
		attribute.String("behaviour", deref(input.Behaviour)),
		attribute.String("versionFilter", fmt.Sprintf("%+v", input.VersionFilter)),
		attribute.String("nameFilter", fmt.Sprintf("%+v", input.NameFilter)),
	}})
	defer func() {
		endObservation(1, observation.Args{Attrs: []attribute.KeyValue{
			attribute.Int("filterID", filter.ID),
		}})
	}()
	return s.store.CreatePackageRepoFilter(ctx, input)
}

func (s *Service) UpdatePackageRepoFilter(ctx context.Context, filter shared.PackageRepoFilter) (err error) {
	ctx, _, endObservation := s.operations.updatePackageRepoFilter.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("id", filter.ID),
		attribute.String("packageScheme", filter.PackageScheme),
		attribute.String("behaviour", filter.Behaviour),
		attribute.String("versionFilter", fmt.Sprintf("%+v", filter.VersionFilter)),
		attribute.String("nameFilter", fmt.Sprintf("%+v", filter.NameFilter)),
	}})
	defer endObservation(1, observation.Args{})
	return s.store.UpdatePackageRepoFilter(ctx, filter)
}

func (s *Service) DeletePackageRepoFilter(ctx context.Context, id int) (err error) {
	ctx, _, endObservation := s.operations.deletePackageRepoFilter.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})
	return s.store.DeletePacakgeRepoFilter(ctx, id)
}

func (s *Service) IsPackageRepoVersionAllowed(ctx context.Context, scheme string, pkg reposource.PackageName, version string) (allowed bool, err error) {
	ctx, _, endObservation := s.operations.isPackageRepoVersionAllowed.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("packageScheme", scheme),
		attribute.String("name", string(pkg)),
		attribute.String("version", version),
	}})
	defer endObservation(1, observation.Args{})

	filters, _, err := s.store.ListPackageRepoRefFilters(ctx, store.ListPackageRepoRefFiltersOpts{
		PackageScheme:  scheme,
		IncludeDeleted: false,
	})
	if err != nil {
		return false, err
	}

	packageFilters, err := packagefilters.NewFilterLists(filters)
	if err != nil {
		return false, err
	}

	return packagefilters.IsVersionedPackageAllowed(scheme, pkg, version, packageFilters), nil
}

func (s *Service) IsPackageRepoAllowed(ctx context.Context, scheme string, pkg reposource.PackageName) (allowed bool, err error) {
	ctx, _, endObservation := s.operations.isPackageRepoAllowed.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("packageScheme", scheme),
		attribute.String("name", string(pkg)),
	}})
	defer endObservation(1, observation.Args{})

	filters, _, err := s.store.ListPackageRepoRefFilters(ctx, store.ListPackageRepoRefFiltersOpts{
		PackageScheme:  scheme,
		IncludeDeleted: false,
	})
	if err != nil {
		return false, err
	}

	packageFilters, err := packagefilters.NewFilterLists(filters)
	if err != nil {
		return false, err
	}

	return packagefilters.IsPackageAllowed(scheme, pkg, packageFilters), nil
}

func (s *Service) PackagesOrVersionsMatchingFilter(ctx context.Context, filter shared.MinimalPackageFilter, limit, after int) (_ []shared.PackageRepoReference, _ int, hasMore bool, err error) {
	ctx, _, endObservation := s.operations.pkgsOrVersionsMatchingFilter.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("packageScheme", filter.PackageScheme),
		attribute.String("versionFilter", fmt.Sprintf("%+v", filter.VersionFilter)),
		attribute.String("nameFilter", fmt.Sprintf("%+v", filter.NameFilter)),
	}})
	defer endObservation(1, observation.Args{})

	var (
		totalCount   int
		matchingPkgs = make([]shared.PackageRepoReference, 0, limit)
	)

	if filter.NameFilter != nil {
		// we dont use a compiled glob when checking name filters as we can do a hugely more efficient regex search
		// in postgres instead of paging through every single package to do a glob check here
		nameRegex, err := packagefilters.GlobToRegex(filter.NameFilter.PackageGlob)
		if err != nil {
			return nil, 0, false, errors.Wrap(err, "failed to compile glob")
		}

		var lastID int
		for {
			pkgs, _, _, err := s.store.ListPackageRepoRefs(ctx, store.ListDependencyReposOpts{
				Scheme: filter.PackageScheme,
				// we filter down here else we have to page through a potentially huge number of non-matching packages
				Name:      reposource.PackageName(nameRegex),
				Fuzziness: store.FuzzinessRegex,
				// doing this so we don't have to load everything in at once
				Limit:          500,
				After:          lastID,
				IncludeBlocked: true,
			})
			if err != nil {
				return nil, 0, false, errors.Wrap(err, "failed to list package repo references")
			}

			if len(pkgs) == 0 {
				break
			}

			lastID = pkgs[len(pkgs)-1].ID

			totalCount += len(pkgs)

			for _, pkg := range pkgs {
				if pkg.ID <= after {
					continue
				}
				if len(matchingPkgs) == limit {
					// once we've reached the limit but are hitting more, we know theres more
					hasMore = true
					continue
				}
				pkg.Versions = nil
				matchingPkgs = append(matchingPkgs, pkg)
			}
		}
	} else {
		matcher, err := packagefilters.NewVersionGlob(filter.VersionFilter.PackageName, filter.VersionFilter.VersionGlob)
		if err != nil {
			return nil, 0, false, errors.Wrap(err, "failed to compile glob")
		}
		nameToMatch := filter.VersionFilter.PackageName

		pkgs, _, _, err := s.store.ListPackageRepoRefs(ctx, store.ListDependencyReposOpts{
			Scheme:    filter.PackageScheme,
			Name:      reposource.PackageName(nameToMatch),
			Fuzziness: store.FuzzinessExactMatch,
			// should only have 1 matching package ref
			Limit:          1,
			IncludeBlocked: true,
		})
		if err != nil {
			return nil, 0, false, errors.Wrap(err, "failed to list package repo references")
		}

		if len(pkgs) == 0 {
			return nil, 0, false, errors.Newf("package repo reference not found for name %q", nameToMatch)
		}

		pkg := pkgs[0]
		versions := pkg.Versions[:0]
		for _, version := range pkg.Versions {
			if matcher.Matches(pkg.Name, version.Version) {
				totalCount++
				if version.ID <= after {
					continue
				}
				if len(versions) == limit {
					// once we've reached the limit but are hitting more, we know theres more
					hasMore = true
					continue
				}
				versions = append(versions, version)
			}
		}
		pkg.Versions = versions
		matchingPkgs = append(matchingPkgs, pkg)
	}

	return matchingPkgs, totalCount, hasMore, nil
}
