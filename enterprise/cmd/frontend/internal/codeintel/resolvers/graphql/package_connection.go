package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
)

type PackageConnectionResolver struct {
	packages   []resolvers.AdjustedPackage
	totalCount int
}

func NewPackageConnectionResolver(packages []resolvers.AdjustedPackage, totalCount int) gql.PackageConnectionResolver {
	return &PackageConnectionResolver{
		packages:   packages,
		totalCount: totalCount,
	}
}

func (r *PackageConnectionResolver) Nodes(ctx context.Context) ([]gql.PackageResolver, error) {
	resolvers := make([]gql.PackageResolver, 0, len(r.packages))
	for i := range r.packages {
		resolvers = append(resolvers, NewPackageResolver(r.packages[i]))
	}
	return resolvers, nil
}

func (r *PackageConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return int32(r.totalCount), nil
}

func (r *PackageConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(len(r.packages) < r.totalCount), nil
}
