package graphqlbackend

import (
	"context"
	"fmt"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/packagerepos"
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
	PackageReferenceKind string
	Filter               packageFilter
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

	var (
		filter      packagerepos.PackageMatcher
		nameToMatch string
	)
	if args.Filter.NameFilter != nil {
		filter, err = packagerepos.NewPackageNameGlob(args.Filter.NameFilter.PackageGlob)
	} else {
		filter, err = packagerepos.NewVersionGlob(args.Filter.VersionFilter.PackageName, args.Filter.VersionFilter.VersionGlob)
		nameToMatch = args.Filter.VersionFilter.PackageName
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to compile glob")
	}

	limit := int(args.GetFirst())

	var after int
	if args.After != nil {
		if after, err = graphqlutil.DecodeIntCursor(args.After); err != nil {
			return nil, err
		}
	}

	packageRepoScheme := externalServiceToPackageSchemeMap[args.PackageReferenceKind]

	depsService := dependencies.NewService(observation.NewContext(r.logger), r.db)

	matchingPkgs := make([]shared.PackageRepoReference, 0, limit)
	if args.Filter.NameFilter != nil {
		lastID := after

	gather:
		for limit == 0 || len(matchingPkgs) < limit {
			fmt.Println(limit == 0, len(matchingPkgs) < limit)
			pkgs, _, err := depsService.ListPackageRepoRefs(ctx, dependencies.ListDependencyReposOpts{
				Scheme: packageRepoScheme,
				After:  lastID,
				Limit:  limit,
			})
			if err != nil {
				return nil, errors.Wrap(err, "failed to list package repo references")
			}

			if len(pkgs) == 0 {
				break
			}

			lastID = pkgs[len(pkgs)-1].ID

			for _, pkg := range pkgs {
				if filter.Matches(string(pkg.Name), "") {
					pkg.Versions = nil
					matchingPkgs = append(matchingPkgs, pkg)
				}
				if limit != 0 && len(matchingPkgs) == limit {
					break gather
				}
			}
		}
	} else {
		pkgs, _, err := depsService.ListPackageRepoRefs(ctx, dependencies.ListDependencyReposOpts{
			Scheme:        packageRepoScheme,
			Name:          reposource.PackageName(nameToMatch),
			ExactNameOnly: true,
			Limit:         1,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to list package repo references")
		}

		if len(pkgs) == 0 {
			return nil, errors.Newf("package repo reference not found for name %q", nameToMatch)
		}

		pkg := pkgs[0]
		versions := pkg.Versions[:0]
		for _, version := range pkg.Versions {
			if filter.Matches(string(pkg.Name), version.Version) {
				versions = append(versions, version)
			}
		}
		pkg.Versions = versions
		matchingPkgs = append(matchingPkgs, pkg)
	}

	return &packageRepoReferenceConnectionResolver{
		db:   r.db,
		deps: matchingPkgs,
		// bit of a lie lol
		total: len(matchingPkgs),
	}, nil
}

func (r *schemaResolver) AddPackageRepoFilter(ctx context.Context, args struct {
	Behaviour            string
	PackageReferenceKind string
	Filter               packageFilter
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
		Behaviour:       args.Behaviour,
		ExternalService: args.PackageReferenceKind,
	}

	if args.Filter.NameFilter != nil {
		filter.NameFilter = &struct{ PackageGlob string }{
			PackageGlob: args.Filter.NameFilter.PackageGlob,
		}
	} else {
		filter.VersionFilter = &struct {
			PackageName string
			VersionGlob string
		}{
			PackageName: args.Filter.VersionFilter.PackageName,
			VersionGlob: args.Filter.VersionFilter.VersionGlob,
		}
	}

	return &EmptyResponse{}, depsService.CreatePackageRepoFilter(ctx, filter)
}

func (r *schemaResolver) UpdatePackageRepoFilter(ctx context.Context, args *struct {
	ID     graphql.ID
	Filter packageFilter
},
) (*EmptyResponse, error) {
	return nil, nil
}

func (r *schemaResolver) DeletePackageRepoFilter(ctx context.Context, id graphql.ID) (*EmptyResponse, error) {
	return nil, nil
}
