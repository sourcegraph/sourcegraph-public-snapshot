package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/resolvers"
)

type DependencyConnectionResolver struct {
	dependencies []resolvers.AdjustedDependency
	totalCount   int
}

func NewDependencyConnectionResolver(dependencies []resolvers.AdjustedDependency, totalCount int) gql.DependencyConnectionResolver {
	return &DependencyConnectionResolver{
		dependencies: dependencies,
		totalCount:   totalCount,
	}
}

func (r *DependencyConnectionResolver) Nodes(ctx context.Context) ([]gql.DependencyResolver, error) {
	resolvers := make([]gql.DependencyResolver, 0, len(r.dependencies))
	for i := range r.dependencies {
		resolvers = append(resolvers, NewDependencyResolver(r.dependencies[i]))
	}
	return resolvers, nil
}

func (r *DependencyConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return int32(r.totalCount), nil
}

func (r *DependencyConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(len(r.dependencies) < r.totalCount), nil
}
