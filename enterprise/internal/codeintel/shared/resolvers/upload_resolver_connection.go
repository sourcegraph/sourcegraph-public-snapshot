package sharedresolvers

import (
	"context"

	"github.com/opentracing/opentracing-go/log"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type UploadConnectionResolver struct {
	uploadsSvc       UploadsService
	autoindexingSvc  AutoIndexingService
	policySvc        PolicyService
	uploadsResolver  *UploadsResolver
	prefetcher       *Prefetcher
	locationResolver *CachedLocationResolver
	traceErrs        *observation.ErrCollector
}

func NewUploadConnectionResolver(uploadsSvc UploadsService, autoindexingSvc AutoIndexingService, policySvc PolicyService, uploadsResolver *UploadsResolver, prefetcher *Prefetcher, traceErrs *observation.ErrCollector) resolverstubs.LSIFUploadConnectionResolver {
	db := autoindexingSvc.GetUnsafeDB()
	return &UploadConnectionResolver{
		uploadsSvc:       uploadsSvc,
		autoindexingSvc:  autoindexingSvc,
		policySvc:        policySvc,
		uploadsResolver:  uploadsResolver,
		prefetcher:       prefetcher,
		locationResolver: NewCachedLocationResolver(db, gitserver.NewClient(db)),
		traceErrs:        traceErrs,
	}
}

func (r *UploadConnectionResolver) Nodes(ctx context.Context) (_ []resolverstubs.LSIFUploadResolver, err error) {
	defer r.traceErrs.Collect(&err, log.String("uploadConnectionResolver.field", "nodes"))

	if err := r.uploadsResolver.Resolve(ctx); err != nil {
		return nil, err
	}

	resolvers := make([]resolverstubs.LSIFUploadResolver, 0, len(r.uploadsResolver.Uploads))
	for i := range r.uploadsResolver.Uploads {
		resolvers = append(resolvers, NewUploadResolver(r.uploadsSvc, r.autoindexingSvc, r.policySvc, r.uploadsResolver.Uploads[i], r.prefetcher, r.traceErrs))
	}
	return resolvers, nil
}

func (r *UploadConnectionResolver) TotalCount(ctx context.Context) (_ *int32, err error) {
	defer r.traceErrs.Collect(&err, log.String("uploadConnectionResolver.field", "totalCount"))

	if err := r.uploadsResolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return toInt32(&r.uploadsResolver.TotalCount), nil
}

func (r *UploadConnectionResolver) PageInfo(ctx context.Context) (_ resolverstubs.PageInfo, err error) {
	defer r.traceErrs.Collect(&err, log.String("uploadConnectionResolver.field", "pageInfo"))

	if err := r.uploadsResolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return EncodeIntCursor(toInt32(r.uploadsResolver.NextOffset)), nil
}
