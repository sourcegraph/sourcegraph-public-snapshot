package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type packageFilter struct {
	NameFilter *struct {
		PackageGlob string
	}
	VersionFilter *struct {
		PackageName string
		VersionGlob string
	}
}

func (r *schemaResolver) PackageRepoReferencesMatchingFilter(ctx context.Context, args struct {
	Kind   string
	Filter packageFilter
	graphqlutil.ConnectionArgs
	After *string
},
) (_ *packageRepoReferenceConnectionResolver, err error) {
	if args.Filter.NameFilter == nil && args.Filter.VersionFilter == nil {
		return nil, errors.New("must provide either nameFilter or versionFilter")
	}

	if args.Filter.NameFilter != nil && args.Filter.VersionFilter != nil {
		return nil, errors.New("cannot provide both a name filter and version filter")
	}

	limit := int(args.GetFirst())

	var after int
	if args.After != nil {
		if after, err = graphqlutil.DecodeIntCursor(args.After); err != nil {
			return nil, err
		}
	}

	depsService := dependencies.NewService(observation.NewContext(r.logger), r.db)

	matchingPkgs, totalCount, err := depsService.PackagesOrVersionsMatchingFilter(ctx, shared.MinimalPackageFilter{
		PackageScheme: externalServiceToPackageSchemeMap[args.Kind],
		NameFilter:    args.Filter.NameFilter,
		VersionFilter: args.Filter.VersionFilter,
	}, limit, after)

	return &packageRepoReferenceConnectionResolver{
		db:    r.db,
		deps:  matchingPkgs,
		total: totalCount,
	}, err
}

func (r *schemaResolver) AddPackageRepoFilter(ctx context.Context, args struct {
	Behaviour string
	Kind      string
	Filter    packageFilter
},
) (*EmptyResponse, error) {
	if args.Filter.NameFilter == nil && args.Filter.VersionFilter == nil {
		return nil, errors.New("must provide either nameFilter or versionFilter")
	}

	if args.Filter.NameFilter != nil && args.Filter.VersionFilter != nil {
		return nil, errors.New("cannot provide both a name filter and version filter")
	}

	depsService := dependencies.NewService(observation.NewContext(r.logger), r.db)

	filter := shared.PackageFilter{
		Behaviour:     args.Behaviour,
		PackageScheme: externalServiceToPackageSchemeMap[args.Kind],
		NameFilter:    args.Filter.NameFilter,
		VersionFilter: args.Filter.VersionFilter,
	}

	return &EmptyResponse{}, depsService.CreatePackageRepoFilter(ctx, filter)
}

func (r *schemaResolver) UpdatePackageRepoFilter(ctx context.Context, args *struct {
	ID        graphql.ID
	Behaviour string
	Kind      string
	Filter    packageFilter
},
) (*EmptyResponse, error) {
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

	return &EmptyResponse{}, depsService.UpdatePackageRepoFilter(ctx, shared.PackageFilter{
		ID:            filterID,
		Behaviour:     args.Behaviour,
		PackageScheme: externalServiceToPackageSchemeMap[args.Kind],
		NameFilter:    args.Filter.NameFilter,
		VersionFilter: args.Filter.VersionFilter,
	})
}

func (r *schemaResolver) DeletePackageRepoFilter(ctx context.Context, args struct{ ID graphql.ID }) (*EmptyResponse, error) {
	depsService := dependencies.NewService(observation.NewContext(r.logger), r.db)

	var filterID int
	if err := relay.UnmarshalSpec(args.ID, &filterID); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, depsService.DeletePacakgeRepoFilter(ctx, filterID)
}
