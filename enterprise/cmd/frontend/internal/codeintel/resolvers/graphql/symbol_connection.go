package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
)

type SymbolConnectionResolver struct {
	symbols    []resolvers.AdjustedSymbol
	root       *resolvers.AdjustedSymbol
	totalCount int

	locationResolver *CachedLocationResolver
	newQueryResolver newQueryResolver
}

func NewSymbolConnectionResolver(symbols []resolvers.AdjustedSymbol, root *resolvers.AdjustedSymbol, totalCount int, locationResolver *CachedLocationResolver, newQueryResolver newQueryResolver) gql.SymbolConnectionResolver {
	return &SymbolConnectionResolver{
		symbols:          symbols,
		totalCount:       totalCount,
		locationResolver: locationResolver,
		newQueryResolver: newQueryResolver,
	}
}

func (r *SymbolConnectionResolver) Nodes(ctx context.Context) ([]gql.SymbolResolver, error) {
	resolvers := make([]gql.SymbolResolver, 0, len(r.symbols))
	for i := range r.symbols {
		resolvers = append(resolvers, NewSymbolResolver(r.symbols[i], r.root, r.locationResolver, r.newQueryResolver))
	}
	return resolvers, nil
}

func (r *SymbolConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return int32(r.totalCount), nil
}

func (r *SymbolConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(len(r.symbols) < r.totalCount), nil
}
