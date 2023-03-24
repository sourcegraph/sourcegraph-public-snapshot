package graphql

import (
	"context"

	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

func NewLocationConnectionResolver(locations []types.UploadLocation, cursor *string, locationResolver *sharedresolvers.CachedLocationResolver) resolverstubs.LocationConnectionResolver {
	return resolverstubs.NewLazyConnectionResolver(func(ctx context.Context) ([]resolverstubs.LocationResolver, error) {
		return resolveLocations(ctx, locationResolver, locations)
	}, encodeCursor(cursor))
}
