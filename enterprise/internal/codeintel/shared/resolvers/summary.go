package sharedresolvers

import (
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type repositorySummaryResolver struct {
	autoindexingSvc  AutoIndexingService
	uploadsSvc       UploadsService
	policySvc        PolicyService
	summary          RepositorySummary
	prefetcher       *Prefetcher
	locationResolver *CachedLocationResolver
	errTracer        *observation.ErrCollector
}

func NewRepositorySummaryResolver(autoindexingSvc AutoIndexingService, uploadsSvc UploadsService, policySvc PolicyService, summary RepositorySummary, prefetcher *Prefetcher, errTracer *observation.ErrCollector) resolverstubs.CodeIntelRepositorySummaryResolver {
	db := autoindexingSvc.GetUnsafeDB()
	return &repositorySummaryResolver{
		autoindexingSvc:  autoindexingSvc,
		uploadsSvc:       uploadsSvc,
		policySvc:        policySvc,
		summary:          summary,
		prefetcher:       prefetcher,
		locationResolver: NewCachedLocationResolver(db, gitserver.NewClient(db)),
		errTracer:        errTracer,
	}
}

func (r *repositorySummaryResolver) RecentUploads() []resolverstubs.LSIFUploadsWithRepositoryNamespaceResolver {
	resolvers := make([]resolverstubs.LSIFUploadsWithRepositoryNamespaceResolver, 0, len(r.summary.RecentUploads))
	for _, upload := range r.summary.RecentUploads {
		uploadResolvers := make([]resolverstubs.LSIFUploadResolver, 0, len(upload.Uploads))
		for _, u := range upload.Uploads {
			uploadResolvers = append(uploadResolvers, NewUploadResolver(r.uploadsSvc, r.autoindexingSvc, r.policySvc, u, r.prefetcher, r.errTracer))
		}

		resolvers = append(resolvers, NewLSIFUploadsWithRepositoryNamespaceResolver(upload, uploadResolvers))
	}

	return resolvers
}

func (r *repositorySummaryResolver) RecentIndexes() []resolverstubs.LSIFIndexesWithRepositoryNamespaceResolver {
	resolvers := make([]resolverstubs.LSIFIndexesWithRepositoryNamespaceResolver, 0, len(r.summary.RecentIndexes))
	for _, index := range r.summary.RecentIndexes {
		indexResolvers := make([]resolverstubs.LSIFIndexResolver, 0, len(index.Indexes))
		for _, idx := range index.Indexes {
			indexResolvers = append(indexResolvers, NewIndexResolver(r.autoindexingSvc, r.uploadsSvc, r.policySvc, idx, r.prefetcher, r.errTracer))
		}
		resolvers = append(resolvers, NewLSIFIndexesWithRepositoryNamespaceResolver(index, indexResolvers))
	}

	return resolvers
}

func (r *repositorySummaryResolver) LastUploadRetentionScan() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.summary.LastUploadRetentionScan)
}

func (r *repositorySummaryResolver) LastIndexScan() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.summary.LastIndexScan)
}
