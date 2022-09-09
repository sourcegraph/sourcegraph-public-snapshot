package graphql

import (
	"context"

	"github.com/opentracing/opentracing-go/log"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	uploadsgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type UploadConnectionResolver struct {
	db               database.DB
	gitserver        GitserverClient
	resolver         resolvers.Resolver
	uploadsResolver  *uploadsgraphql.UploadsResolver
	prefetcher       *Prefetcher
	locationResolver *CachedLocationResolver
	errTracer        *observation.ErrCollector
}

func NewUploadConnectionResolver(db database.DB, gitserver GitserverClient, resolver resolvers.Resolver, uploadsResolver *uploadsgraphql.UploadsResolver, prefetcher *Prefetcher, locationResolver *CachedLocationResolver, errTracer *observation.ErrCollector) gql.LSIFUploadConnectionResolver {
	return &UploadConnectionResolver{
		resolver:         resolver,
		db:               db,
		gitserver:        gitserver,
		uploadsResolver:  uploadsResolver,
		prefetcher:       prefetcher,
		locationResolver: locationResolver,
		errTracer:        errTracer,
	}
}

func (r *UploadConnectionResolver) Nodes(ctx context.Context) (_ []gql.LSIFUploadResolver, err error) {
	defer r.errTracer.Collect(&err, log.String("uploadConnectionResolver.field", "nodes"))

	if err := r.uploadsResolver.Resolve(ctx); err != nil {
		return nil, err
	}

	resolvers := make([]gql.LSIFUploadResolver, 0, len(r.uploadsResolver.Uploads))
	for i := range r.uploadsResolver.Uploads {
		upload := convertSharedUploadsToDBStoreUploads(r.uploadsResolver.Uploads[i])
		resolvers = append(resolvers, NewUploadResolver(r.db, r.gitserver, r.resolver, upload, r.prefetcher, r.locationResolver, r.errTracer))
	}
	return resolvers, nil
}

func (r *UploadConnectionResolver) TotalCount(ctx context.Context) (_ *int32, err error) {
	defer r.errTracer.Collect(&err, log.String("uploadConnectionResolver.field", "totalCount"))

	if err := r.uploadsResolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return toInt32(&r.uploadsResolver.TotalCount), nil
}

func (r *UploadConnectionResolver) PageInfo(ctx context.Context) (_ *graphqlutil.PageInfo, err error) {
	defer r.errTracer.Collect(&err, log.String("uploadConnectionResolver.field", "pageInfo"))

	if err := r.uploadsResolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return graphqlutil.EncodeIntCursor(toInt32(r.uploadsResolver.NextOffset)), nil
}
