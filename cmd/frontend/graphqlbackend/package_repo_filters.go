package graphqlbackend

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/packagerepos"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type packageMatcher struct {
	NameMatcher *struct {
		PackageGlob string
	}
	VersionMatcher *struct {
		PackageName string
		VersionGlob string
	}
}

func (r *schemaResolver) PackageReposMatches(ctx context.Context, args struct {
	PackageReferenceKind string
	Matcher              packageMatcher
	graphqlutil.ConnectionArgs
	After *string
},
) (*packageRepoReferenceConnectionResolver, error) {
	if args.Matcher.NameMatcher == nil && args.Matcher.VersionMatcher == nil {
		return nil, errors.New("must provide either nameMatcher or versionMatcher")
	}

	if args.Matcher.NameMatcher != nil && args.Matcher.VersionMatcher != nil {
		return nil, errors.New("cannot provide both a name matcher and version matcher")
	}

	kinds := []string{args.PackageReferenceKind}

	extsvcs, err := r.db.ExternalServices().List(ctx, database.ExternalServicesListOptions{Kinds: kinds})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list external services")
	}

	if len(extsvcs) == 0 {
		return nil, errors.Newf("no external service configured of kind %q", args.PackageReferenceKind)
	}

	var (
		matcher     packagerepos.PackageMatcher
		nameToMatch string
	)
	if args.Matcher.NameMatcher != nil {
		matcher, err = packagerepos.NewPackageNameGlob(args.Matcher.NameMatcher.PackageGlob)
	} else {
		matcher, err = packagerepos.NewVersionGlob(args.Matcher.VersionMatcher.PackageName, args.Matcher.VersionMatcher.VersionGlob)
		nameToMatch = args.Matcher.VersionMatcher.PackageName
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
	if args.Matcher.NameMatcher != nil {
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
				if matcher.Matches(string(pkg.Name), "") {
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
			if matcher.Matches(string(pkg.Name), version.Version) {
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

func (r *schemaResolver) AddPackageRepoMatcher(ctx context.Context, args struct {
	Behaviour            string
	PackageReferenceKind string
	Matcher              packageMatcher
},
) (*EmptyResponse, error) {
	if args.Matcher.NameMatcher == nil && args.Matcher.VersionMatcher == nil {
		return nil, errors.New("must provide either nameMatcher or versionMatcher")
	}

	if args.Matcher.NameMatcher != nil && args.Matcher.VersionMatcher != nil {
		return nil, errors.New("cannot provide both a name matcher and version matcher")
	}

	depsService := dependencies.NewService(observation.NewContext(r.logger), r.db)

	filter := shared.PackageFilter{
		Behaviour:       args.Behaviour,
		ExternalService: args.PackageReferenceKind,
	}

	if args.Matcher.NameMatcher != nil {
		filter.NameMatcher = &struct{ PackageGlob string }{
			PackageGlob: args.Matcher.NameMatcher.PackageGlob,
		}
	} else {
		filter.VersionMatcher = &struct {
			PackageName string
			VersionGlob string
		}{
			PackageName: args.Matcher.VersionMatcher.PackageName,
			VersionGlob: args.Matcher.VersionMatcher.VersionGlob,
		}
	}

	err := depsService.CreatePackageRepoFilter(ctx, filter)

	return &EmptyResponse{}, err
}
