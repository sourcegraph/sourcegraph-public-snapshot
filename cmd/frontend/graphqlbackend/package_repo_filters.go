package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type inputPackageFilter struct {
	NameFilter *struct {
		PackageGlob string
	}
	VersionFilter *struct {
		PackageName string
		VersionGlob string
	}
}

type filterMatchingResolver struct {
	packageResolver *packageRepoReferenceConnectionResolver
	versionResolver *packageRepoReferenceVersionConnectionResolver
}

func (r *filterMatchingResolver) ToPackageRepoReferenceConnection() (*packageRepoReferenceConnectionResolver, bool) {
	return r.packageResolver, r.packageResolver != nil
}

func (r *filterMatchingResolver) ToPackageRepoReferenceVersionConnection() (*packageRepoReferenceVersionConnectionResolver, bool) {
	return r.versionResolver, r.versionResolver != nil
}

func (r *schemaResolver) PackageRepoReferencesMatchingFilter(ctx context.Context, args struct {
	Kind   string
	Filter inputPackageFilter
	gqlutil.ConnectionArgs
	After *string
},
) (_ *filterMatchingResolver, err error) {
	if args.Filter.NameFilter == nil && args.Filter.VersionFilter == nil {
		return nil, errors.New("must provide either nameFilter or versionFilter")
	}

	if args.Filter.NameFilter != nil && args.Filter.VersionFilter != nil {
		return nil, errors.New("cannot provide both a name filter and version filter")
	}

	limit := int(args.GetFirst())

	var after int
	if args.After != nil {
		if err = relay.UnmarshalSpec(graphql.ID(*args.After), &after); err != nil {
			return nil, err
		}
	}

	depsService := dependencies.NewService(observation.NewContext(r.logger), r.db)

	matchingPkgs, totalCount, hasMore, err := depsService.PackagesOrVersionsMatchingFilter(ctx, shared.MinimalPackageFilter{
		PackageScheme: externalServiceToPackageSchemeMap[args.Kind],
		NameFilter:    args.Filter.NameFilter,
		VersionFilter: args.Filter.VersionFilter,
	}, limit, after)

	if args.Filter.NameFilter != nil {
		return &filterMatchingResolver{
			packageResolver: &packageRepoReferenceConnectionResolver{
				db:      r.db,
				deps:    matchingPkgs,
				hasMore: hasMore,
				total:   totalCount,
			},
		}, err
	}

	var versions []shared.PackageRepoRefVersion
	if len(matchingPkgs) == 1 {
		versions = matchingPkgs[0].Versions
	}
	return &filterMatchingResolver{
		versionResolver: &packageRepoReferenceVersionConnectionResolver{
			versions: versions,
			hasMore:  hasMore,
			total:    totalCount,
		},
	}, err
}

func (r *schemaResolver) PackageRepoFilters(ctx context.Context, args struct {
	Behaviour *string
	Kind      *string
},
) (resolvers *[]*packageRepoFilterResolver, err error) {
	var opts dependencies.ListPackageRepoRefFiltersOpts

	if args.Behaviour != nil {
		opts.Behaviour = *args.Behaviour
	}

	if args.Kind != nil {
		opts.PackageScheme = externalServiceToPackageSchemeMap[*args.Kind]
	}

	depsService := dependencies.NewService(observation.NewContext(r.logger), r.db)
	filters, _, err := depsService.ListPackageRepoFilters(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, "error listing package repo filters")
	}

	resolvers = new([]*packageRepoFilterResolver)
	*resolvers = make([]*packageRepoFilterResolver, 0, len(filters))

	for _, filter := range filters {
		*resolvers = append(*resolvers, &packageRepoFilterResolver{
			filter: filter,
		})
	}

	return resolvers, nil
}

type packageRepoFilterResolver struct {
	filter dependencies.PackageRepoFilter
}

func (r *packageRepoFilterResolver) ID() graphql.ID {
	return relay.MarshalID("PackageRepoFilter", r.filter.ID)
}

func (r *packageRepoFilterResolver) Behaviour() string {
	return r.filter.Behaviour
}

func (r *packageRepoFilterResolver) Kind() string {
	return packageSchemeToExternalServiceMap[r.filter.PackageScheme]
}

func (r *packageRepoFilterResolver) NameFilter() *packageRepoNameFilterResolver {
	if r.filter.NameFilter != nil {
		return &packageRepoNameFilterResolver{*r.filter.NameFilter}
	}
	return nil
}

func (r *packageRepoFilterResolver) VersionFilter() *packageRepoVersionFilterResolver {
	if r.filter.VersionFilter != nil {
		return &packageRepoVersionFilterResolver{*r.filter.VersionFilter}
	}
	return nil
}

type packageRepoVersionFilterResolver struct {
	filter struct {
		PackageName string
		VersionGlob string
	}
}

func (r *packageRepoVersionFilterResolver) PackageName() string {
	return r.filter.PackageName
}

func (r *packageRepoVersionFilterResolver) VersionGlob() string {
	return r.filter.VersionGlob
}

type packageRepoNameFilterResolver struct {
	filter struct {
		PackageGlob string
	}
}

func (r *packageRepoNameFilterResolver) PackageGlob() string {
	return r.filter.PackageGlob
}

func (r *schemaResolver) AddPackageRepoFilter(ctx context.Context, args struct {
	Behaviour string
	Kind      string
	Filter    inputPackageFilter
},
) (*packageRepoFilterResolver, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	if args.Filter.NameFilter == nil && args.Filter.VersionFilter == nil {
		return nil, errors.New("must provide either nameFilter or versionFilter")
	}

	if args.Filter.NameFilter != nil && args.Filter.VersionFilter != nil {
		return nil, errors.New("cannot provide both a name filter and version filter")
	}

	depsService := dependencies.NewService(observation.NewContext(r.logger), r.db)

	filter := shared.MinimalPackageFilter{
		Behaviour:     &args.Behaviour,
		PackageScheme: externalServiceToPackageSchemeMap[args.Kind],
		NameFilter:    args.Filter.NameFilter,
		VersionFilter: args.Filter.VersionFilter,
	}

	newFilter, err := depsService.CreatePackageRepoFilter(ctx, filter)
	return &packageRepoFilterResolver{*newFilter}, err
}

func (r *schemaResolver) UpdatePackageRepoFilter(ctx context.Context, args *struct {
	ID        graphql.ID
	Behaviour string
	Kind      string
	Filter    inputPackageFilter
},
) (*EmptyResponse, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	if args.Filter.NameFilter == nil && args.Filter.VersionFilter == nil {
		return nil, errors.New("must provide either nameFilter or versionFilter")
	}

	if args.Filter.NameFilter != nil && args.Filter.VersionFilter != nil {
		return nil, errors.New("cannot provide both a name filter and version filter")
	}

	depsService := dependencies.NewService(observation.NewContext(r.logger), r.db)

	var filterID int
	if err := relay.UnmarshalSpec(args.ID, &filterID); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, depsService.UpdatePackageRepoFilter(ctx, shared.PackageRepoFilter{
		ID:            filterID,
		Behaviour:     args.Behaviour,
		PackageScheme: externalServiceToPackageSchemeMap[args.Kind],
		NameFilter:    args.Filter.NameFilter,
		VersionFilter: args.Filter.VersionFilter,
	})
}

func (r *schemaResolver) DeletePackageRepoFilter(ctx context.Context, args struct{ ID graphql.ID }) (*EmptyResponse, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	depsService := dependencies.NewService(observation.NewContext(r.logger), r.db)

	var filterID int
	if err := relay.UnmarshalSpec(args.ID, &filterID); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, depsService.DeletePackageRepoFilter(ctx, filterID)
}
