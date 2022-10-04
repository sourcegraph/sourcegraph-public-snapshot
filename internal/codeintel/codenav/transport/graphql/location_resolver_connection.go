package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/sharedresolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
)

type LocationConnectionResolver interface {
	Nodes(ctx context.Context) ([]LocationResolver, error)
	PageInfo(ctx context.Context) (*PageInfo, error)
}

type locationConnectionResolver struct {
	locations        []types.UploadLocation
	cursor           *string
	locationResolver *sharedresolvers.CachedLocationResolver
}

func NewLocationConnectionResolver(locations []types.UploadLocation, cursor *string, locationResolver *sharedresolvers.CachedLocationResolver) LocationConnectionResolver {
	return &locationConnectionResolver{
		locations:        locations,
		cursor:           cursor,
		locationResolver: locationResolver,
	}
}

func (r *locationConnectionResolver) Nodes(ctx context.Context) ([]LocationResolver, error) {
	return resolveLocations(ctx, r.locationResolver, r.locations)
}

func (r *locationConnectionResolver) PageInfo(ctx context.Context) (*PageInfo, error) {
	return EncodeCursor(r.cursor), nil
}
