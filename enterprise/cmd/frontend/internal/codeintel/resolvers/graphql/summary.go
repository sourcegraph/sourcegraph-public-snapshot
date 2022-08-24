package graphql

import (
	"strings"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type repositorySummaryResolver struct {
	db               database.DB
	resolver         resolvers.Resolver
	gitserver        GitserverClient
	summary          resolvers.RepositorySummary
	prefetcher       *Prefetcher
	locationResolver *CachedLocationResolver
	errTracer        *observation.ErrCollector
}

func NewRepositorySummaryResolver(
	db database.DB,
	resolver resolvers.Resolver,
	gitserver GitserverClient,
	summary resolvers.RepositorySummary,
	prefetcher *Prefetcher,
	locationResolver *CachedLocationResolver,
	errTracer *observation.ErrCollector,
) gql.CodeIntelRepositorySummaryResolver {
	return &repositorySummaryResolver{
		db:               db,
		resolver:         resolver,
		gitserver:        gitserver,
		summary:          summary,
		prefetcher:       prefetcher,
		locationResolver: locationResolver,
		errTracer:        errTracer,
	}
}

func (r *repositorySummaryResolver) RecentUploads() []gql.LSIFUploadsWithRepositoryNamespaceResolver {
	resolvers := make([]gql.LSIFUploadsWithRepositoryNamespaceResolver, 0, len(r.summary.RecentUploads))
	for _, upload := range r.summary.RecentUploads {
		uploadResolvers := make([]gql.LSIFUploadResolver, 0, len(upload.Uploads))
		for _, upload := range upload.Uploads {
			uploadResolvers = append(uploadResolvers, NewUploadResolver(r.db, r.gitserver, r.resolver, upload, r.prefetcher, r.locationResolver, r.errTracer))
		}

		resolvers = append(resolvers, NewLSIFUploadsWithRepositoryNamespaceResolver(upload, uploadResolvers))
	}

	return resolvers
}

func (r *repositorySummaryResolver) RecentIndexes() []gql.LSIFIndexesWithRepositoryNamespaceResolver {
	resolvers := make([]gql.LSIFIndexesWithRepositoryNamespaceResolver, 0, len(r.summary.RecentIndexes))
	for _, index := range r.summary.RecentIndexes {
		indexResolvers := make([]gql.LSIFIndexResolver, 0, len(index.Indexes))
		for _, u := range index.Indexes {
			upload := convertSharedIndexToDBStoreIndex(u)
			indexResolvers = append(indexResolvers, NewIndexResolver(r.db, r.gitserver, r.resolver, upload, r.prefetcher, r.locationResolver, r.errTracer))
		}
		dbstoreIndex := convertSharedIndexesWithRepositoryNamespaceToDBStoreIndexesWithRepositoryNamespace(index)
		resolvers = append(resolvers, NewLSIFIndexesWithRepositoryNamespaceResolver(dbstoreIndex, indexResolvers))
	}

	return resolvers
}

func (r *repositorySummaryResolver) LastUploadRetentionScan() *gql.DateTime {
	return gql.DateTimeOrNil(r.summary.LastUploadRetentionScan)
}

func (r *repositorySummaryResolver) LastIndexScan() *gql.DateTime {
	return gql.DateTimeOrNil(r.summary.LastIndexScan)
}

type LSIFUploadsWithRepositoryNamespaceResolver struct {
	uploadsSummary  dbstore.UploadsWithRepositoryNamespace
	uploadResolvers []gql.LSIFUploadResolver
}

func NewLSIFUploadsWithRepositoryNamespaceResolver(
	uploadsSummary dbstore.UploadsWithRepositoryNamespace,
	uploadResolvers []gql.LSIFUploadResolver,
) gql.LSIFUploadsWithRepositoryNamespaceResolver {
	return &LSIFUploadsWithRepositoryNamespaceResolver{
		uploadsSummary:  uploadsSummary,
		uploadResolvers: uploadResolvers,
	}
}

func (r *LSIFUploadsWithRepositoryNamespaceResolver) Root() string {
	return r.uploadsSummary.Root
}

func (r *LSIFUploadsWithRepositoryNamespaceResolver) Indexer() gql.CodeIntelIndexerResolver {
	for _, indexer := range allIndexers {
		if indexer.Name() == r.uploadsSummary.Indexer {
			return indexer
		}
	}

	return &codeIntelIndexerResolver{name: r.uploadsSummary.Indexer}
}

func (r *LSIFUploadsWithRepositoryNamespaceResolver) Uploads() []gql.LSIFUploadResolver {
	return r.uploadResolvers
}

type LSIFIndexesWithRepositoryNamespaceResolver struct {
	indexesSummary dbstore.IndexesWithRepositoryNamespace
	indexResolvers []gql.LSIFIndexResolver
}

func NewLSIFIndexesWithRepositoryNamespaceResolver(
	indexesSummary dbstore.IndexesWithRepositoryNamespace,
	indexResolvers []gql.LSIFIndexResolver,
) gql.LSIFIndexesWithRepositoryNamespaceResolver {
	return &LSIFIndexesWithRepositoryNamespaceResolver{
		indexesSummary: indexesSummary,
		indexResolvers: indexResolvers,
	}
}

func (r *LSIFIndexesWithRepositoryNamespaceResolver) Root() string {
	return r.indexesSummary.Root
}

func (r *LSIFIndexesWithRepositoryNamespaceResolver) Indexer() gql.CodeIntelIndexerResolver {
	// drop the tag if it exists
	if idx, ok := imageToIndexer[strings.Split(r.indexesSummary.Indexer, ":")[0]]; ok {
		return idx
	}

	return &codeIntelIndexerResolver{name: r.indexesSummary.Indexer}
}

func (r *LSIFIndexesWithRepositoryNamespaceResolver) Indexes() []gql.LSIFIndexResolver {
	return r.indexResolvers
}
