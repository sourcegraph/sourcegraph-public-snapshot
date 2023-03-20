package sharedresolvers

import (
	"context"
	"sort"
	"strconv"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type InferredAvailableIndexers struct {
	Indexer types.CodeIntelIndexer
	Roots   []string
}

type summaryResolver struct {
	autoindexSvc     AutoIndexingService
	locationResolver *CachedLocationResolver
}

func NewSummaryResolver(autoindexSvc AutoIndexingService, locationResolver *CachedLocationResolver) resolverstubs.CodeIntelSummaryResolver {
	return &summaryResolver{
		autoindexSvc:     autoindexSvc,
		locationResolver: locationResolver,
	}
}

func (r *summaryResolver) NumRepositoriesWithCodeIntelligence(ctx context.Context) (int32, error) {
	numRepositoriesWithCodeIntelligence, err := r.autoindexSvc.NumRepositoriesWithCodeIntelligence(ctx)
	if err != nil {
		return 0, err
	}

	return int32(numRepositoriesWithCodeIntelligence), nil
}

func (r *summaryResolver) RepositoriesWithErrors(ctx context.Context, args *resolverstubs.RepositoriesWithErrorsArgs) (resolverstubs.CodeIntelRepositoryWithErrorConnectionResolver, error) {
	pageSize := 25
	if args.First != nil {
		pageSize = int(*args.First)
	}

	offset := 0
	if args.After != nil {
		after, _ := strconv.Atoi(*args.After)
		offset = after
	}

	repositoryIDsWithErrors, totalCount, err := r.autoindexSvc.RepositoryIDsWithErrors(ctx, offset, pageSize)
	if err != nil {
		return nil, err
	}

	var resolvers []resolverstubs.CodeIntelRepositoryWithErrorResolver
	for _, repositoryWithCount := range repositoryIDsWithErrors {
		resolver, err := r.locationResolver.Repository(ctx, api.RepoID(repositoryWithCount.RepositoryID))
		if err != nil {
			return nil, err
		}

		resolvers = append(resolvers, &codeIntelRepositoryWithErrorResolver{
			repositoryResolver: resolver,
			count:              repositoryWithCount.Count,
		})
	}

	endCursor := ""
	if newOffset := offset + pageSize; newOffset < totalCount {
		endCursor = strconv.Itoa(newOffset)
	}

	return &codeIntelRepositoryWithErrorConnectionResolver{
		nodes:      resolvers,
		totalCount: totalCount,
		endCursor:  endCursor,
	}, nil
}

type codeIntelRepositoryWithErrorConnectionResolver struct {
	nodes      []resolverstubs.CodeIntelRepositoryWithErrorResolver
	totalCount int
	endCursor  string
}

func (r *codeIntelRepositoryWithErrorConnectionResolver) Nodes() []resolverstubs.CodeIntelRepositoryWithErrorResolver {
	return r.nodes
}

func (r *codeIntelRepositoryWithErrorConnectionResolver) TotalCount() *int32 {
	v := int32(r.totalCount)
	return &v
}

func (r *codeIntelRepositoryWithErrorConnectionResolver) PageInfo() resolverstubs.PageInfo {
	if r.endCursor != "" {
		return &pageInfo{hasNextPage: true, endCursor: &r.endCursor}
	}

	return &pageInfo{hasNextPage: false}
}

func (r *summaryResolver) RepositoriesWithConfiguration(ctx context.Context, args *resolverstubs.RepositoriesWithConfigurationArgs) (resolverstubs.CodeIntelRepositoryWithConfigurationConnectionResolver, error) {
	pageSize := 25
	if args.First != nil {
		pageSize = int(*args.First)
	}

	offset := 0
	if args.After != nil {
		after, _ := strconv.Atoi(*args.After)
		offset = after
	}

	repositoryIDsWithConfiguration, totalCount, err := r.autoindexSvc.RepositoryIDsWithConfiguration(ctx, offset, pageSize)
	if err != nil {
		return nil, err
	}

	var resolvers []resolverstubs.CodeIntelRepositoryWithConfigurationResolver
	for _, repositoryWithAvailableIndexers := range repositoryIDsWithConfiguration {
		resolver, err := r.locationResolver.Repository(ctx, api.RepoID(repositoryWithAvailableIndexers.RepositoryID))
		if err != nil {
			return nil, err
		}

		resolvers = append(resolvers, &codeIntelRepositoryWithConfigurationResolver{
			repositoryResolver: resolver,
			availableIndexers:  repositoryWithAvailableIndexers.AvailableIndexers,
		})
	}

	endCursor := ""
	if newOffset := offset + pageSize; newOffset < totalCount {
		endCursor = strconv.Itoa(newOffset)
	}

	return &codeIntelRepositoryWithConfigurationConnectionResolver{
		nodes:      resolvers,
		totalCount: totalCount,
		endCursor:  endCursor,
	}, nil
}

type codeIntelRepositoryWithConfigurationConnectionResolver struct {
	nodes      []resolverstubs.CodeIntelRepositoryWithConfigurationResolver
	totalCount int
	endCursor  string
}

func (r *codeIntelRepositoryWithConfigurationConnectionResolver) Nodes() []resolverstubs.CodeIntelRepositoryWithConfigurationResolver {
	return r.nodes
}

func (r *codeIntelRepositoryWithConfigurationConnectionResolver) TotalCount() *int32 {
	v := int32(r.totalCount)
	return &v
}

func (r *codeIntelRepositoryWithConfigurationConnectionResolver) PageInfo() resolverstubs.PageInfo {
	if r.endCursor != "" {
		return &pageInfo{hasNextPage: true, endCursor: &r.endCursor}
	}

	return &pageInfo{hasNextPage: false}
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
			indexer: types.NewCodeIntelIndexerResolver(indexer, ""),
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
	uploadsSvc        UploadsService
	policySvc         PolicyService
	gitserverClient   gitserver.Client
	siteAdminChecker  SiteAdminChecker
	repoStore         database.RepoStore
	summary           RepositorySummary
	availableIndexers []InferredAvailableIndexers
	limitErr          error
	prefetcher        *Prefetcher
	locationResolver  *CachedLocationResolver
	errTracer         *observation.ErrCollector
}

func NewRepositorySummaryResolver(
	uploadsSvc UploadsService,
	policySvc PolicyService,
	gitserverClient gitserver.Client,
	siteAdminChecker SiteAdminChecker,
	repoStore database.RepoStore,
	locationResolver *CachedLocationResolver,
	summary RepositorySummary,
	availableIndexers []InferredAvailableIndexers,
	limitErr error,
	prefetcher *Prefetcher,
	errTracer *observation.ErrCollector,
) resolverstubs.CodeIntelRepositorySummaryResolver {
	return &repositorySummaryResolver{
		uploadsSvc:        uploadsSvc,
		policySvc:         policySvc,
		gitserverClient:   gitserverClient,
		siteAdminChecker:  siteAdminChecker,
		repoStore:         repoStore,
		summary:           summary,
		availableIndexers: availableIndexers,
		limitErr:          limitErr,
		prefetcher:        prefetcher,
		locationResolver:  locationResolver,
		errTracer:         errTracer,
	}
}

func (r *repositorySummaryResolver) AvailableIndexers() []resolverstubs.InferredAvailableIndexersResolver {
	resolvers := make([]resolverstubs.InferredAvailableIndexersResolver, 0, len(r.availableIndexers))
	for _, indexer := range r.availableIndexers {
		resolvers = append(resolvers, resolverstubs.NewInferredAvailableIndexersResolver(types.NewCodeIntelIndexerResolverFrom(indexer.Indexer, ""), indexer.Roots))
	}
	return resolvers
}

func (r *repositorySummaryResolver) RecentActivity(ctx context.Context) ([]resolverstubs.PreciseIndexResolver, error) {
	uploadIDs := map[int]struct{}{}
	var resolvers []resolverstubs.PreciseIndexResolver
	for _, recentUploads := range r.summary.RecentUploads {
		for _, upload := range recentUploads.Uploads {
			upload := upload

			resolver, err := NewPreciseIndexResolver(ctx, r.uploadsSvc, r.policySvc, r.gitserverClient, r.prefetcher, r.siteAdminChecker, r.repoStore, r.locationResolver, r.errTracer, &upload, nil)
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

			resolver, err := NewPreciseIndexResolver(ctx, r.uploadsSvc, r.policySvc, r.gitserverClient, r.prefetcher, r.siteAdminChecker, r.repoStore, r.locationResolver, r.errTracer, nil, &index)
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

func (r *repositorySummaryResolver) LimitError() *string {
	if r.limitErr != nil {
		m := r.limitErr.Error()
		return &m
	}

	return nil
}
