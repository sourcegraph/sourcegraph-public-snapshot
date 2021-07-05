package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type symbolUsageResolver struct {
	symbol *SymbolResolver

	locationResolver *CachedLocationResolver
}

func (r *symbolUsageResolver) References(ctx context.Context) (gql.LocationConnectionResolver, error) {
	return r.symbol.References(ctx)
}

func (r *symbolUsageResolver) ReferenceGroups(ctx context.Context) ([]gql.ReferenceGroupResolver, error) {
	var groups []gql.ReferenceGroupResolver

	locations, _, err := r.symbol.references(ctx)
	if err != nil {
		return nil, err
	}

	groups = append(groups, &referenceGroupResolver{
		name:             "My group",
		locations:        locations,
		locationResolver: r.locationResolver,
	})

	return groups, nil
}
