package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
)

type DocSymbolConnectionResolver struct {
	symbols []gql.DocSymbolResolver
}

func NewDocSymbolConnectionResolver(symbols []*resolvers.AdjustedSymbol) gql.DocSymbolConnectionResolver {
	symbolResolvers := make([]gql.DocSymbolResolver, len(symbols))
	for i := range symbols {
		symbolResolvers[i] = &docSymbolResolver{adjustedSymbol: symbols[i]}
	}
	return &DocSymbolConnectionResolver{symbols: symbolResolvers}
}

func (r *DocSymbolConnectionResolver) Nodes(ctx context.Context) ([]gql.DocSymbolResolver, error) {
	return r.symbols, nil
}

func (r *DocSymbolConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}

type docSymbolResolver struct {
	adjustedSymbol *resolvers.AdjustedSymbol
}

func (r *docSymbolResolver) Name(ctx context.Context) (string, error) {
	return r.adjustedSymbol.Text, nil
}

func (r *docSymbolResolver) Children(ctx context.Context) ([]gql.DocSymbolResolver, error) {
	return []gql.DocSymbolResolver{ /* TODO */ }, nil
}
