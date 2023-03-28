package graphql

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	autoindexingShared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	uploadsShared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *rootResolver) CodeIntelSummary(ctx context.Context) (_ resolverstubs.CodeIntelSummaryResolver, err error) {
	ctx, _, endObservation := r.operations.codeIntelSummary.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	return newSummaryResolver(r.autoindexSvc, r.locationResolverFactory.Create()), nil
}

func (r *rootResolver) RepositorySummary(ctx context.Context, repoID graphql.ID) (_ resolverstubs.CodeIntelRepositorySummaryResolver, err error) {
	ctx, errTracer, endObservation := r.operations.repositorySummary.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repoID", string(repoID)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	id, err := resolverstubs.UnmarshalID[int](repoID)
	if err != nil {
		return nil, err
	}

	lastUploadRetentionScan, err := r.uploadSvc.GetLastUploadRetentionScanForRepository(ctx, id)
	if err != nil {
		return nil, err
	}

	lastIndexScan, err := r.autoindexSvc.GetLastIndexScanForRepository(ctx, id)
	if err != nil {
		return nil, err
	}

	recentUploads, err := r.uploadSvc.GetRecentUploadsSummary(ctx, id)
	if err != nil {
		return nil, err
	}

	recentIndexes, err := r.autoindexSvc.GetRecentIndexesSummary(ctx, id)
	if err != nil {
		return nil, err
	}

	// Create blocklist for indexes that have already been uploaded.
	blocklist := map[string]struct{}{}
	for _, u := range recentUploads {
		key := shared.GetKeyForLookup(u.Indexer, u.Root)
		blocklist[key] = struct{}{}
	}
	for _, u := range recentIndexes {
		key := shared.GetKeyForLookup(u.Indexer, u.Root)
		blocklist[key] = struct{}{}
	}

	commit := "HEAD"
	var limitErr error

	indexJobs, err := r.autoindexSvc.InferIndexJobsFromRepositoryStructure(ctx, id, commit, "", false)
	if err != nil {
		if !autoindexing.IsLimitError(err) {
			return nil, err
		}

		limitErr = errors.Append(limitErr, err)
	}
	// indexJobHints, err := r.autoindexSvc.InferIndexJobHintsFromRepositoryStructure(ctx, repoID, commit)
	// if err != nil {
	// 	if !errors.As(err, &inference.LimitError{}) {
	// 		return nil, err
	// 	}

	// 	limitErr = errors.Append(limitErr, err)
	// }

	inferredAvailableIndexers := map[string]shared.AvailableIndexer{}
	inferredAvailableIndexers = shared.PopulateInferredAvailableIndexers(indexJobs, blocklist, inferredAvailableIndexers)
	// inferredAvailableIndexers = shared.PopulateInferredAvailableIndexers(indexJobHints, blocklist, inferredAvailableIndexers)

	inferredAvailableIndexersResolver := make([]inferredAvailableIndexers2, 0, len(inferredAvailableIndexers))
	for _, indexer := range inferredAvailableIndexers {
		inferredAvailableIndexersResolver = append(inferredAvailableIndexersResolver,
			inferredAvailableIndexers2{
				Indexer: indexer.Indexer,
				Roots:   indexer.Roots,
			},
		)
	}

	summary := RepositorySummary{
		RecentUploads:           recentUploads,
		RecentIndexes:           recentIndexes,
		LastUploadRetentionScan: lastUploadRetentionScan,
		LastIndexScan:           lastIndexScan,
	}

	return newRepositorySummaryResolver(
		r.uploadSvc,
		r.policySvc,
		r.gitserverClient,
		r.siteAdminChecker,
		r.repoStore,
		r.locationResolverFactory.Create(),
		summary,
		inferredAvailableIndexersResolver,
		limitErr,
		r.prefetcherFactory.Create(),
		errTracer,
	), nil
}

//
//

type summaryResolver struct {
	autoindexSvc     AutoIndexingService
	locationResolver *sharedresolvers.CachedLocationResolver
}

func newSummaryResolver(autoindexSvc AutoIndexingService, locationResolver *sharedresolvers.CachedLocationResolver) resolverstubs.CodeIntelSummaryResolver {
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

	return resolverstubs.NewCursorWithTotalCountConnectionResolver(resolvers, endCursor, int32(totalCount)), nil
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

	return resolverstubs.NewCursorWithTotalCountConnectionResolver(resolvers, endCursor, int32(totalCount)), nil
}

//
//

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

type RepositorySummary struct {
	RecentUploads           []uploadsShared.UploadsWithRepositoryNamespace
	RecentIndexes           []autoindexingShared.IndexesWithRepositoryNamespace
	LastUploadRetentionScan *time.Time
	LastIndexScan           *time.Time
}

//
//

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

//
//

type repositorySummaryResolver struct {
	uploadsSvc        UploadsService
	policySvc         PolicyService
	gitserverClient   gitserver.Client
	siteAdminChecker  sharedresolvers.SiteAdminChecker
	repoStore         database.RepoStore
	summary           RepositorySummary
	availableIndexers []inferredAvailableIndexers2
	limitErr          error
	prefetcher        *sharedresolvers.Prefetcher
	locationResolver  *sharedresolvers.CachedLocationResolver
	errTracer         *observation.ErrCollector
}

type inferredAvailableIndexers2 struct {
	Indexer types.CodeIntelIndexer
	Roots   []string
}

func newRepositorySummaryResolver(
	uploadsSvc UploadsService,
	policySvc PolicyService,
	gitserverClient gitserver.Client,
	siteAdminChecker sharedresolvers.SiteAdminChecker,
	repoStore database.RepoStore,
	locationResolver *sharedresolvers.CachedLocationResolver,
	summary RepositorySummary,
	availableIndexers []inferredAvailableIndexers2,
	limitErr error,
	prefetcher *sharedresolvers.Prefetcher,
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
		resolvers = append(resolvers, newInferredAvailableIndexersResolver(types.NewCodeIntelIndexerResolverFrom(indexer.Indexer, ""), indexer.Roots))
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

//
//

type inferredAvailableIndexersResolver struct {
	indexer resolverstubs.CodeIntelIndexerResolver
	roots   []string
}

func newInferredAvailableIndexersResolver(indexer resolverstubs.CodeIntelIndexerResolver, roots []string) resolverstubs.InferredAvailableIndexersResolver {
	return &inferredAvailableIndexersResolver{
		indexer: indexer,
		roots:   roots,
	}
}

func (r *inferredAvailableIndexersResolver) Indexer() resolverstubs.CodeIntelIndexerResolver {
	return r.indexer
}

func (r *inferredAvailableIndexersResolver) Roots() []string {
	return r.roots
}

func (r *inferredAvailableIndexersResolver) RootsWithKeys() []resolverstubs.RootsWithKeyResolver {
	var resolvers []resolverstubs.RootsWithKeyResolver
	for _, root := range r.roots {
		resolvers = append(resolvers, &rootWithKeyResolver{
			root: root,
			key:  comparisonKey(root, r.indexer.Name()),
		})
	}

	return resolvers
}

type rootWithKeyResolver struct {
	root string
	key  string
}

func (r *rootWithKeyResolver) Root() string {
	return r.root
}

func (r *rootWithKeyResolver) ComparisonKey() string {
	return r.key
}

//
//

func comparisonKey(root, indexer string) string {
	hash := sha256.New()
	_, _ = hash.Write([]byte(strings.Join([]string{root, indexer}, "\x00")))
	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}
