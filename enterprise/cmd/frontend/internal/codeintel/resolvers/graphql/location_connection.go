package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type LocationConnectionResolver struct {
	locations        []AdjustedLocation
	cursor           *string
	locationResolver *CachedLocationResolver
}

func NewLocationConnectionResolver(locations []AdjustedLocation, cursor *string, locationResolver *CachedLocationResolver) gql.LocationConnectionResolver {
	return &LocationConnectionResolver{
		locations:        locations,
		cursor:           cursor,
		locationResolver: locationResolver,
	}
}

func (r *LocationConnectionResolver) Nodes(ctx context.Context) ([]gql.LocationResolver, error) {
	return resolveLocations(ctx, r.locationResolver, r.locations)
}

func (r *LocationConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.EncodeCursor(r.cursor), nil
}
