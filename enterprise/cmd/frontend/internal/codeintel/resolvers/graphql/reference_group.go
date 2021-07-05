package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
)

func filterLocations(locations []resolvers.AdjustedLocation, keep func(resolvers.AdjustedLocation) bool) (filtered []resolvers.AdjustedLocation) {
	filtered = locations[:0]
	for _, loc := range locations {
		if keep(loc) {
			filtered = append(filtered, loc)
		}
	}
	return filtered
}

type referenceGroupResolver struct {
	name      string
	locations []resolvers.AdjustedLocation

	locationResolver *CachedLocationResolver
}

func (r *referenceGroupResolver) Name() string {
	return r.name
}

func (r *referenceGroupResolver) Locations(ctx context.Context) (gql.LocationConnectionResolver, error) {
	return NewLocationConnectionResolver(r.locations, nil, r.locationResolver), nil
}
