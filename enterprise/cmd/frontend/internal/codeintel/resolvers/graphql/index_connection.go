package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
)

type IndexConnectionResolver struct {
	resolver         resolvers.Resolver
	indexesResolver  *resolvers.IndexesResolver
	prefetcher       *Prefetcher
	locationResolver *CachedLocationResolver
}

func NewIndexConnectionResolver(resolver resolvers.Resolver, indexesResolver *resolvers.IndexesResolver, prefetcher *Prefetcher, locationResolver *CachedLocationResolver) gql.LSIFIndexConnectionResolver {
	return &IndexConnectionResolver{
		resolver:         resolver,
		indexesResolver:  indexesResolver,
		prefetcher:       prefetcher,
		locationResolver: locationResolver,
	}
}

func (r *IndexConnectionResolver) Nodes(ctx context.Context) ([]gql.LSIFIndexResolver, error) {
	if err := r.indexesResolver.Resolve(ctx); err != nil {
		return nil, err
	}

	resolvers := make([]gql.LSIFIndexResolver, 0, len(r.indexesResolver.Indexes))
	for i := range r.indexesResolver.Indexes {
		resolvers = append(resolvers, NewIndexResolver(r.resolver, r.indexesResolver.Indexes[i], r.prefetcher, r.locationResolver))
	}
	return resolvers, nil
}

func (r *IndexConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	if err := r.indexesResolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return toInt32(&r.indexesResolver.TotalCount), nil
}

func (r *IndexConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if err := r.indexesResolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return graphqlutil.EncodeIntCursor(toInt32(r.indexesResolver.NextOffset)), nil
}
