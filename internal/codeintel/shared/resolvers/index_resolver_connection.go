package sharedresolvers

import (
	"context"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type LSIFIndexConnectionResolver interface {
	Nodes(ctx context.Context) ([]LSIFIndexResolver, error)
	TotalCount(ctx context.Context) (*int32, error)
	PageInfo(ctx context.Context) (*PageInfo, error)
}

type IndexConnectionResolver struct {
	uploadsSvc       UploadsService
	autoindexingSvc  AutoIndexingService
	policySvc        PolicyService
	indexesResolver  *IndexesResolver
	prefetcher       *Prefetcher
	locationResolver *CachedLocationResolver
	errTracer        *observation.ErrCollector
}

func NewIndexConnectionResolver(autoindexingSvc AutoIndexingService, uploadsSvc UploadsService, policySvc PolicyService, indexesResolver *IndexesResolver, prefetcher *Prefetcher, errTracer *observation.ErrCollector) LSIFIndexConnectionResolver {
	db := autoindexingSvc.GetUnsafeDB()
	return &IndexConnectionResolver{
		uploadsSvc:       uploadsSvc,
		autoindexingSvc:  autoindexingSvc,
		policySvc:        policySvc,
		indexesResolver:  indexesResolver,
		prefetcher:       prefetcher,
		locationResolver: NewCachedLocationResolver(db, gitserver.NewClient(db)),
		errTracer:        errTracer,
	}
}

func (r *IndexConnectionResolver) Nodes(ctx context.Context) ([]LSIFIndexResolver, error) {
	if err := r.indexesResolver.Resolve(ctx); err != nil {
		return nil, err
	}

	resolvers := make([]LSIFIndexResolver, 0, len(r.indexesResolver.Indexes))
	for i := range r.indexesResolver.Indexes {
		// index := convertSharedIndexToDBStoreIndex(r.indexesResolver.Indexes[i])
		resolvers = append(resolvers, NewIndexResolver(r.autoindexingSvc, r.uploadsSvc, r.policySvc, r.indexesResolver.Indexes[i], r.prefetcher, r.errTracer))
	}
	return resolvers, nil
}

func (r *IndexConnectionResolver) TotalCount(ctx context.Context) (_ *int32, err error) {
	defer r.errTracer.Collect(&err, log.String("indexConnectionResolver.field", "totalCount"))

	if err := r.indexesResolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return toInt32(&r.indexesResolver.TotalCount), nil
}

func (r *IndexConnectionResolver) PageInfo(ctx context.Context) (_ *PageInfo, err error) {
	defer r.errTracer.Collect(&err, log.String("indexConnectionResolver.field", "pageInfo"))

	if err := r.indexesResolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return EncodeIntCursor(r.indexesResolver.NextOffset), nil
}
