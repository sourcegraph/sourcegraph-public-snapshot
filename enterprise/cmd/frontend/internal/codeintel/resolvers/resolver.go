package resolvers

import (
	"context"
	"time"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	executor "github.com/sourcegraph/sourcegraph/internal/services/executors/transport/graphql"
	symbolsClient "github.com/sourcegraph/sourcegraph/internal/symbols"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

// Resolver is the main interface to code intel-related operations exposed to the GraphQL API.
// This resolver consolidates the logic for code intel operations and is not itself concerned
// with GraphQL/API specifics (auth, validation, marshaling, etc.). This resolver is wrapped
// by a symmetrics resolver in this package's graphql subpackage, which is exposed directly
// by the API.
type Resolver interface {
	// TODO: Move to uploads resolver.
	GetUploadByID(ctx context.Context, id int) (dbstore.Upload, bool, error)
	GetUploadsByIDs(ctx context.Context, ids ...int) ([]dbstore.Upload, error)
	DeleteUploadByID(ctx context.Context, uploadID int) error
	GetUploadDocumentsForPath(ctx context.Context, uploadID int, pathPrefix string) ([]string, int, error)
	CommitGraph(ctx context.Context, repositoryID int) (gql.CodeIntelligenceCommitGraphResolver, error)
	UploadConnectionResolver(opts dbstore.GetUploadsOptions) *UploadsResolver
	AuditLogsForUpload(ctx context.Context, id int) ([]dbstore.UploadLog, error)
	RepositorySummary(ctx context.Context, repositoryID int) (RepositorySummary, error)

	// TODO: Move to autoindex service.
	GetIndexByID(ctx context.Context, id int) (dbstore.Index, bool, error)
	GetIndexesByIDs(ctx context.Context, ids ...int) ([]dbstore.Index, error)
	DeleteIndexByID(ctx context.Context, id int) error
	IndexConfiguration(ctx context.Context, repositoryID int) ([]byte, bool, error)
	InferedIndexConfiguration(ctx context.Context, repositoryID int, commit string) (*config.IndexConfiguration, bool, error)
	InferedIndexConfigurationHints(ctx context.Context, repositoryID int, commit string) ([]config.IndexJobHint, error)
	UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, configuration string) error
	QueueAutoIndexJobsForRepo(ctx context.Context, repositoryID int, rev, configuration string) ([]dbstore.Index, error)
	IndexConnectionResolver(opts dbstore.GetIndexesOptions) *IndexesResolver

	// TODO: Move to codenav service.
	SupportedByCtags(ctx context.Context, filepath string, repo api.RepoName) (bool, string, error)
	RequestLanguageSupport(ctx context.Context, userID int, language string) error
	RequestedLanguageSupport(ctx context.Context, userID int) ([]string, error)

	ExecutorResolver() executor.Resolver
	CodeNavResolver() CodeNavResolver
	PoliciesResolver() PoliciesResolver
}

type RepositorySummary struct {
	RecentUploads           []dbstore.UploadsWithRepositoryNamespace
	RecentIndexes           []dbstore.IndexesWithRepositoryNamespace
	LastUploadRetentionScan *time.Time
	LastIndexScan           *time.Time
}

type resolver struct {
	dbStore       DBStore
	lsifStore     LSIFStore
	indexEnqueuer IndexEnqueuer
	symbolsClient *symbolsClient.Client

	executorResolver executor.Resolver
	codenavResolver  CodeNavResolver
	policiesResolver PoliciesResolver
}

// NewResolver creates a new resolver with the given services.
func NewResolver(
	dbStore DBStore,
	lsifStore LSIFStore,
	indexEnqueuer IndexEnqueuer,
	symbolsClient *symbolsClient.Client,
	codenavResolver CodeNavResolver,
	executorResolver executor.Resolver,
	policiesResolver PoliciesResolver,
) Resolver {
	return &resolver{
		dbStore:       dbStore,
		lsifStore:     lsifStore,
		indexEnqueuer: indexEnqueuer,
		symbolsClient: symbolsClient,

		executorResolver: executorResolver,
		codenavResolver:  codenavResolver,
		policiesResolver: policiesResolver,
	}
}

func (r *resolver) CodeNavResolver() CodeNavResolver {
	return r.codenavResolver
}

func (r *resolver) PoliciesResolver() PoliciesResolver {
	return r.policiesResolver
}

func (r *resolver) ExecutorResolver() executor.Resolver {
	return r.executorResolver
}

func (r *resolver) GetUploadByID(ctx context.Context, id int) (dbstore.Upload, bool, error) {
	return r.dbStore.GetUploadByID(ctx, id)
}

func (r *resolver) GetIndexByID(ctx context.Context, id int) (dbstore.Index, bool, error) {
	return r.dbStore.GetIndexByID(ctx, id)
}

func (r *resolver) GetUploadsByIDs(ctx context.Context, ids ...int) ([]dbstore.Upload, error) {
	return r.dbStore.GetUploadsByIDs(ctx, ids...)
}

func (r *resolver) GetIndexesByIDs(ctx context.Context, ids ...int) ([]dbstore.Index, error) {
	return r.dbStore.GetIndexesByIDs(ctx, ids...)
}

func (r *resolver) UploadConnectionResolver(opts dbstore.GetUploadsOptions) *UploadsResolver {
	return NewUploadsResolver(r.dbStore, opts)
}

func (r *resolver) IndexConnectionResolver(opts dbstore.GetIndexesOptions) *IndexesResolver {
	return NewIndexesResolver(r.dbStore, opts)
}

func (r *resolver) DeleteUploadByID(ctx context.Context, uploadID int) error {
	_, err := r.dbStore.DeleteUploadByID(ctx, uploadID)
	return err
}

func (r *resolver) DeleteIndexByID(ctx context.Context, id int) error {
	_, err := r.dbStore.DeleteIndexByID(ctx, id)
	return err
}

func (r *resolver) CommitGraph(ctx context.Context, repositoryID int) (gql.CodeIntelligenceCommitGraphResolver, error) {
	stale, updatedAt, err := r.dbStore.CommitGraphMetadata(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	return NewCommitGraphResolver(stale, updatedAt), nil
}

func (r *resolver) GetUploadDocumentsForPath(ctx context.Context, uploadID int, pathPattern string) ([]string, int, error) {
	return r.lsifStore.DocumentPaths(ctx, uploadID, pathPattern)
}

func (r *resolver) QueueAutoIndexJobsForRepo(ctx context.Context, repositoryID int, rev, configuration string) ([]dbstore.Index, error) {
	return r.indexEnqueuer.QueueIndexes(ctx, repositoryID, rev, configuration, true, true)
}

func (r *resolver) IndexConfiguration(ctx context.Context, repositoryID int) ([]byte, bool, error) {
	configuration, exists, err := r.dbStore.GetIndexConfigurationByRepositoryID(ctx, repositoryID)
	if err != nil {
		return nil, false, err
	}
	if !exists {
		return nil, false, nil
	}

	return configuration.Data, true, nil
}

func (r *resolver) InferedIndexConfiguration(ctx context.Context, repositoryID int, commit string) (*config.IndexConfiguration, bool, error) {
	maybeConfig, _, err := r.indexEnqueuer.InferIndexConfiguration(ctx, repositoryID, commit, true)
	if err != nil || maybeConfig == nil {
		return nil, false, err
	}

	return maybeConfig, true, nil
}

func (r *resolver) InferedIndexConfigurationHints(ctx context.Context, repositoryID int, commit string) ([]config.IndexJobHint, error) {
	_, hints, err := r.indexEnqueuer.InferIndexConfiguration(ctx, repositoryID, commit, true)
	if err != nil {
		return nil, err
	}

	return hints, nil
}

func (r *resolver) UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, configuration string) error {
	if _, err := config.UnmarshalJSON([]byte(configuration)); err != nil {
		return err
	}

	return r.dbStore.UpdateIndexConfigurationByRepositoryID(ctx, repositoryID, []byte(configuration))
}

func (r *resolver) SupportedByCtags(ctx context.Context, filepath string, repoName api.RepoName) (bool, string, error) {
	mappings, err := r.symbolsClient.ListLanguageMappings(ctx, repoName)
	if err != nil {
		return false, "", err
	}

	for language, globs := range mappings {
		for _, glob := range globs {
			if glob.Match(filepath) {
				return true, language, nil
			}
		}
	}

	return false, "", nil
}

func (r *resolver) RepositorySummary(ctx context.Context, repositoryID int) (RepositorySummary, error) {
	recentUploads, err := r.dbStore.RecentUploadsSummary(ctx, repositoryID)
	if err != nil {
		return RepositorySummary{}, err
	}

	recentIndexes, err := r.dbStore.RecentIndexesSummary(ctx, repositoryID)
	if err != nil {
		return RepositorySummary{}, err
	}

	lastUploadRetentionScan, err := r.dbStore.LastUploadRetentionScanForRepository(ctx, repositoryID)
	if err != nil {
		return RepositorySummary{}, err
	}

	lastIndexScan, err := r.dbStore.LastIndexScanForRepository(ctx, repositoryID)
	if err != nil {
		return RepositorySummary{}, err
	}

	return RepositorySummary{
		RecentUploads:           recentUploads,
		RecentIndexes:           recentIndexes,
		LastUploadRetentionScan: lastUploadRetentionScan,
		LastIndexScan:           lastIndexScan,
	}, nil
}
