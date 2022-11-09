package sharedresolvers

import (
	"context"

	"github.com/opentracing/opentracing-go/log"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type IndexConnectionResolver struct {
	uploadsSvc       UploadsService
	autoindexingSvc  AutoIndexingService
	policySvc        PolicyService
	indexesResolver  *IndexesResolver
	prefetcher       *Prefetcher
	locationResolver *CachedLocationResolver
	errTracer        *observation.ErrCollector
}

func NewIndexConnectionResolver(autoindexingSvc AutoIndexingService, uploadsSvc UploadsService, policySvc PolicyService, indexesResolver *IndexesResolver, prefetcher *Prefetcher, errTracer *observation.ErrCollector) resolverstubs.LSIFIndexConnectionResolver {
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

func (r *IndexConnectionResolver) Nodes(ctx context.Context) ([]resolverstubs.LSIFIndexResolver, error) {
	if err := r.indexesResolver.Resolve(ctx); err != nil {
		return nil, err
	}

	resolvers := make([]resolverstubs.LSIFIndexResolver, 0, len(r.indexesResolver.Indexes))
	for i := range r.indexesResolver.Indexes {
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

func (r *IndexConnectionResolver) PageInfo(ctx context.Context) (_ resolverstubs.PageInfo, err error) {
	defer r.errTracer.Collect(&err, log.String("indexConnectionResolver.field", "pageInfo"))

	if err := r.indexesResolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return EncodeIntCursor(r.indexesResolver.NextOffset), nil
}
