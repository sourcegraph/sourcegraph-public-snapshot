package graphql

import (
	"context"

	"github.com/opentracing/opentracing-go/log"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type UploadConnectionResolver struct {
	db               database.DB
	resolver         resolvers.Resolver
	uploadsResolver  *resolvers.UploadsResolver
	prefetcher       *Prefetcher
	locationResolver *CachedLocationResolver
	op               *observation.Operation
}

func NewUploadConnectionResolver(db database.DB, resolver resolvers.Resolver, uploadsResolver *resolvers.UploadsResolver, prefetcher *Prefetcher, locationResolver *CachedLocationResolver, op *observation.Operation) gql.LSIFUploadConnectionResolver {
	return &UploadConnectionResolver{
		resolver:         resolver,
		uploadsResolver:  uploadsResolver,
		prefetcher:       prefetcher,
		locationResolver: locationResolver,
		op:               op,
	}
}

func (r *UploadConnectionResolver) Nodes(ctx context.Context) ([]gql.LSIFUploadResolver, error) {
	if err := r.uploadsResolver.Resolve(ctx); err != nil {
		return nil, err
	}

	resolvers := make([]gql.LSIFUploadResolver, 0, len(r.uploadsResolver.Uploads))
	for i := range r.uploadsResolver.Uploads {
		resolvers = append(resolvers, NewUploadResolver(r.db, r.resolver, r.uploadsResolver.Uploads[i], r.prefetcher, r.locationResolver, r.op))
	}
	return resolvers, nil
}

func (r *UploadConnectionResolver) TotalCount(ctx context.Context) (_ *int32, err error) {
	ctx, endObservation := r.op.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("uploadConnectionResolver.field", "totalCount"),
	}})
	defer endObservation(1, observation.Args{})

	if err := r.uploadsResolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return toInt32(&r.uploadsResolver.TotalCount), nil
}

func (r *UploadConnectionResolver) PageInfo(ctx context.Context) (_ *graphqlutil.PageInfo, err error) {
	ctx, endObservation := r.op.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("uploadConnectionResolver.field", "pageInfo"),
	}})
	defer endObservation(1, observation.Args{})

	if err := r.uploadsResolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return graphqlutil.EncodeIntCursor(toInt32(r.uploadsResolver.NextOffset)), nil
}
