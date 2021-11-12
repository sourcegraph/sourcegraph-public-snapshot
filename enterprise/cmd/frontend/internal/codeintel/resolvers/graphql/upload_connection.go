package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
)

type UploadConnectionResolver struct {
	resolver         resolvers.Resolver
	uploadsResolver  *resolvers.UploadsResolver
	prefetcher       *Prefetcher
	locationResolver *CachedLocationResolver
}

func NewUploadConnectionResolver(resolver resolvers.Resolver, uploadsResolver *resolvers.UploadsResolver, prefetcher *Prefetcher, locationResolver *CachedLocationResolver) gql.LSIFUploadConnectionResolver {
	return &UploadConnectionResolver{
		resolver:         resolver,
		uploadsResolver:  uploadsResolver,
		prefetcher:       prefetcher,
		locationResolver: locationResolver,
	}
}

func (r *UploadConnectionResolver) Nodes(ctx context.Context) ([]gql.LSIFUploadResolver, error) {
	if err := r.uploadsResolver.Resolve(ctx); err != nil {
		return nil, err
	}

	resolvers := make([]gql.LSIFUploadResolver, 0, len(r.uploadsResolver.Uploads))
	for i := range r.uploadsResolver.Uploads {
		resolvers = append(resolvers, NewUploadResolver(r.resolver, r.uploadsResolver.Uploads[i], r.prefetcher, r.locationResolver))
	}
	return resolvers, nil
}

func (r *UploadConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	if err := r.uploadsResolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return toInt32(&r.uploadsResolver.TotalCount), nil
}

func (r *UploadConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if err := r.uploadsResolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return graphqlutil.EncodeIntCursor(toInt32(r.uploadsResolver.NextOffset)), nil
}
