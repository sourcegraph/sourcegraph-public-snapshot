package graphql

import (
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/resolvers"
)

type DependencyResolver struct {
	dependency resolvers.AdjustedDependency
}

func NewDependencyResolver(dependency resolvers.AdjustedDependency) gql.DependencyResolver {
	return &DependencyResolver{
		dependency: dependency,
	}
}

func (r *DependencyResolver) LSIFName() string {
	return r.dependency.Dependency.Name
}

func (r *DependencyResolver) LSIFVersion() string {
	return r.dependency.Dependency.Version
}

func (r *DependencyResolver) LSIFManager() string {
	return r.dependency.Dependency.Manager
}
