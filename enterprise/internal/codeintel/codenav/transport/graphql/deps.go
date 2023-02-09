package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

// TODO - the good things
func (r *rootResolver) Dependencies(ctx context.Context, args *resolverstubs.DependenciesArgs) (_ []resolverstubs.DependencyDescriptionResolver, err error) {
	repositoryID, err := graphqlbackend.UnmarshalRepositoryID(args.Repository)
	if err != nil {
		return nil, err
	}

	deps, err := r.svc.GetDependencies(ctx, int(repositoryID))
	if err != nil {
		return nil, err
	}

	var resolvers []resolverstubs.DependencyDescriptionResolver
	for _, dep := range deps {
		resolvers = append(resolvers, NewDependencyResolver(dep.Manager, dep.Name, dep.Version))
	}
	return resolvers, err
}

func (r *rootResolver) DependencyOccurrences(ctx context.Context, args *resolverstubs.DependencyOccurrencesArgs) (_ []resolverstubs.LocationResolver, err error) {
	repositoryID, err := graphqlbackend.UnmarshalRepositoryID(args.Repository)
	if err != nil {
		return nil, err
	}

	occurrences, err := r.svc.GetDependencyOccurrences(ctx, int(repositoryID), args.Manager, args.Name, args.Version)
	if err != nil {
		return nil, err
	}

	locationResolver := sharedresolvers.NewCachedLocationResolver(r.svc.GetUnsafeDB(), gitserver.NewClient())

	var locations []resolverstubs.LocationResolver
	for _, location := range occurrences {
		treeResolver, err := locationResolver.Path(ctx, api.RepoID(location.RepositoryID), location.Commit, location.Path)
		if err != nil || treeResolver == nil {
			return nil, err
		}

		lspRange := convertRange(location.Range)
		locations = append(locations, NewLocationResolver(treeResolver, &lspRange))
	}

	return locations, nil
}

type dependencyDescriptionResolver struct {
	manager string
	name    string
	version string
}

func NewDependencyResolver(manager, name, version string) resolverstubs.DependencyDescriptionResolver {
	return &dependencyDescriptionResolver{
		manager: manager,
		name:    name,
		version: version,
	}
}

func (r *dependencyDescriptionResolver) Manager() string { return r.manager }
func (r *dependencyDescriptionResolver) Name() string    { return r.name }
func (r *dependencyDescriptionResolver) Version() string { return r.version }
