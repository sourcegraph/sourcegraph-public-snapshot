package graphql

import (
	"context"

	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
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
