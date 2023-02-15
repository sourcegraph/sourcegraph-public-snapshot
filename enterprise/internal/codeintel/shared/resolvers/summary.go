package sharedresolvers

import (
	"context"
	"sort"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type InferredAvailableIndexers struct {
	Indexer types.CodeIntelIndexer
	Roots   []string
}

type summaryResolver struct {
	summary          shared.Summary
	locationResolver *CachedLocationResolver
}

func NewSummaryResolver(summary shared.Summary, locationResolver *CachedLocationResolver) resolverstubs.CodeIntelSummaryResolver {
	return &summaryResolver{
		summary:          summary,
		locationResolver: locationResolver,
	}
}

func (r *summaryResolver) NumRepositoriesWithCodeIntelligence(ctx context.Context) (int32, error) {
	return int32(r.summary.NumRepositoriesWithCodeIntelligence), nil
}

func (r *summaryResolver) RepositoriesWithErrors(ctx context.Context) ([]resolverstubs.CodeIntelRepositoryWithErrorResolver, error) {
	var resolvers []resolverstubs.CodeIntelRepositoryWithErrorResolver
	for _, repositoryWithCount := range r.summary.RepositoryIDsWithErrors {
		resolver, err := r.locationResolver.Repository(ctx, api.RepoID(repositoryWithCount.RepositoryID))
		if err != nil {
			return nil, err
		}

		resolvers = append(resolvers, &codeIntelRepositoryWithErrorResolver{
			repositoryResolver: resolver,
			count:              repositoryWithCount.Count,
		})
	}

	return resolvers, nil
}

func (r *summaryResolver) RepositoriesWithConfiguration(ctx context.Context) ([]resolverstubs.CodeIntelRepositoryWithConfigurationResolver, error) {
	var resolvers []resolverstubs.CodeIntelRepositoryWithConfigurationResolver
	for _, repositoryWithAvailableIndexers := range r.summary.RepositoryIDsWithConfiguration {
		resolver, err := r.locationResolver.Repository(ctx, api.RepoID(repositoryWithAvailableIndexers.RepositoryID))
		if err != nil {
			return nil, err
		}

		resolvers = append(resolvers, &codeIntelRepositoryWithConfigurationResolver{
			repositoryResolver: resolver,
			availableIndexers:  repositoryWithAvailableIndexers.AvailableIndexers,
		})
	}

	return resolvers, nil
}

type codeIntelRepositoryWithErrorResolver struct {
	repositoryResolver resolverstubs.RepositoryResolver
	count              int
}

func (r *codeIntelRepositoryWithErrorResolver) Repository() resolverstubs.RepositoryResolver {
	return r.repositoryResolver
}
func (r *codeIntelRepositoryWithErrorResolver) Count() int32 {
	return int32(r.count)
}

type codeIntelRepositoryWithConfigurationResolver struct {
	repositoryResolver resolverstubs.RepositoryResolver
	availableIndexers  map[string]shared.AvailableIndexer
}

func (r *codeIntelRepositoryWithConfigurationResolver) Repository() resolverstubs.RepositoryResolver {
	return r.repositoryResolver
}

func (r *codeIntelRepositoryWithConfigurationResolver) Indexers() []resolverstubs.IndexerWithCountResolver {
	var resolvers []resolverstubs.IndexerWithCountResolver
	for indexer, meta := range r.availableIndexers {
		resolvers = append(resolvers, &indexerWithCountResolver{
			indexer: types.NewCodeIntelIndexerResolver(indexer),
			count:   int32(len(meta.Roots)),
		})
	}

	return resolvers
}

type indexerWithCountResolver struct {
	indexer resolverstubs.CodeIntelIndexerResolver
	count   int32
}

func (r *indexerWithCountResolver) Indexer() resolverstubs.CodeIntelIndexerResolver { return r.indexer }
func (r *indexerWithCountResolver) Count() int32                                    { return r.count }

type repositorySummaryResolver struct {
	autoindexingSvc   AutoIndexingService
	uploadsSvc        UploadsService
	policySvc         PolicyService
	summary           RepositorySummary
	availableIndexers []InferredAvailableIndexers
	prefetcher        *Prefetcher
	locationResolver  *CachedLocationResolver
	errTracer         *observation.ErrCollector
}

func NewRepositorySummaryResolver(
	autoindexingSvc AutoIndexingService,
	uploadsSvc UploadsService,
	policySvc PolicyService,
	summary RepositorySummary,
	availableIndexers []InferredAvailableIndexers,
	prefetcher *Prefetcher,
	errTracer *observation.ErrCollector,
) resolverstubs.CodeIntelRepositorySummaryResolver {
	db := autoindexingSvc.GetUnsafeDB()
	return &repositorySummaryResolver{
		autoindexingSvc:   autoindexingSvc,
		uploadsSvc:        uploadsSvc,
		policySvc:         policySvc,
		summary:           summary,
		availableIndexers: availableIndexers,
		prefetcher:        prefetcher,
		locationResolver:  NewCachedLocationResolver(db, gitserver.NewClient()),
		errTracer:         errTracer,
	}
}

func (r *repositorySummaryResolver) RecentUploads() []resolverstubs.LSIFUploadsWithRepositoryNamespaceResolver {
	resolvers := make([]resolverstubs.LSIFUploadsWithRepositoryNamespaceResolver, 0, len(r.summary.RecentUploads))
	for _, upload := range r.summary.RecentUploads {
		uploadResolvers := make([]resolverstubs.LSIFUploadResolver, 0, len(upload.Uploads))
		for _, u := range upload.Uploads {
			uploadResolvers = append(uploadResolvers, NewUploadResolver(r.uploadsSvc, r.autoindexingSvc, r.policySvc, u, r.prefetcher, r.locationResolver, r.errTracer))
		}

		resolvers = append(resolvers, NewLSIFUploadsWithRepositoryNamespaceResolver(upload, uploadResolvers))
	}

	return resolvers
}

func (r *repositorySummaryResolver) AvailableIndexers() []resolverstubs.InferredAvailableIndexersResolver {
	resolvers := make([]resolverstubs.InferredAvailableIndexersResolver, 0, len(r.availableIndexers))
	for _, indexer := range r.availableIndexers {
		resolvers = append(resolvers, resolverstubs.NewInferredAvailableIndexersResolver(types.NewCodeIntelIndexerResolverFrom(indexer.Indexer), indexer.Roots))
	}
	return resolvers
}

func (r *repositorySummaryResolver) RecentIndexes() []resolverstubs.LSIFIndexesWithRepositoryNamespaceResolver {
	resolvers := make([]resolverstubs.LSIFIndexesWithRepositoryNamespaceResolver, 0, len(r.summary.RecentIndexes))
	for _, index := range r.summary.RecentIndexes {
		indexResolvers := make([]resolverstubs.LSIFIndexResolver, 0, len(index.Indexes))
		for _, idx := range index.Indexes {
			indexResolvers = append(indexResolvers, NewIndexResolver(r.autoindexingSvc, r.uploadsSvc, r.policySvc, idx, r.prefetcher, r.locationResolver, r.errTracer))
		}
		resolvers = append(resolvers, NewLSIFIndexesWithRepositoryNamespaceResolver(index, indexResolvers))
	}

	return resolvers
}

func (r *repositorySummaryResolver) RecentActivity(ctx context.Context) ([]resolverstubs.PreciseIndexResolver, error) {
	uploadIDs := map[int]struct{}{}
	var resolvers []resolverstubs.PreciseIndexResolver
	for _, recentUploads := range r.summary.RecentUploads {
		for _, upload := range recentUploads.Uploads {
			upload := upload

			resolver, err := NewPreciseIndexResolver(ctx, r.autoindexingSvc, r.uploadsSvc, r.policySvc, r.prefetcher, r.locationResolver, r.errTracer, &upload, nil)
			if err != nil {
				return nil, err
			}

			uploadIDs[upload.ID] = struct{}{}
			resolvers = append(resolvers, resolver)
		}
	}
	for _, recentIndexes := range r.summary.RecentIndexes {
		for _, index := range recentIndexes.Indexes {
			index := index

			if index.AssociatedUploadID != nil {
				if _, ok := uploadIDs[*index.AssociatedUploadID]; ok {
					continue
				}
			}

			resolver, err := NewPreciseIndexResolver(ctx, r.autoindexingSvc, r.uploadsSvc, r.policySvc, r.prefetcher, r.locationResolver, r.errTracer, nil, &index)
			if err != nil {
				return nil, err
			}

			resolvers = append(resolvers, resolver)
		}
	}

	sort.Slice(resolvers, func(i, j int) bool { return resolvers[i].ID() < resolvers[j].ID() })
	return resolvers, nil
}

func (r *repositorySummaryResolver) LastUploadRetentionScan() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.summary.LastUploadRetentionScan)
}

func (r *repositorySummaryResolver) LastIndexScan() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.summary.LastIndexScan)
}
