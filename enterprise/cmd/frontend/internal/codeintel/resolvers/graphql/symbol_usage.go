package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type symbolUsageResolver struct {
	symbol *SymbolResolver
}

func (r *symbolUsageResolver) References(ctx context.Context) (gql.LocationConnectionResolver, error) {
	return r.symbol.References(ctx)
}
