package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
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
