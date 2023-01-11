package graphql

import (
	"context"

	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

type locationConnectionResolver struct {
	locations        []types.UploadLocation
	cursor           *string
	locationResolver *sharedresolvers.CachedLocationResolver
}

func NewLocationConnectionResolver(locations []types.UploadLocation, cursor *string, locationResolver *sharedresolvers.CachedLocationResolver) resolverstubs.LocationConnectionResolver {
	return &locationConnectionResolver{
		locations:        locations,
		cursor:           cursor,
		locationResolver: locationResolver,
	}
}

func (r *locationConnectionResolver) Nodes(ctx context.Context) ([]resolverstubs.LocationResolver, error) {
	return resolveLocations(ctx, r.locationResolver, r.locations)
}

func (r *locationConnectionResolver) PageInfo(ctx context.Context) (resolverstubs.PageInfo, error) {
	return EncodeCursor(r.cursor), nil
}
