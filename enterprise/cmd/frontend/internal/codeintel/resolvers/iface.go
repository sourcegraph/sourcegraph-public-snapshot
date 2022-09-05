package resolvers

import (
	"context"
	"time"

	autoindexingShared "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	autoindexinggraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/transport/graphql"
	codenavgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/transport/graphql"
	policiesgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

type DBStore interface {
	gitserver.DBStore

	GetUploadByID(ctx context.Context, id int) (dbstore.Upload, bool, error)
	GetUploadsByIDs(ctx context.Context, ids ...int) ([]dbstore.Upload, error)
	GetUploads(ctx context.Context, opts dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error)
	DeleteUploadByID(ctx context.Context, id int) (bool, error)
	MarkRepositoryAsDirty(ctx context.Context, repositoryID int) error
	CommitGraphMetadata(ctx context.Context, repositoryID int) (stale bool, updatedAt *time.Time, _ error)
	RecentUploadsSummary(ctx context.Context, repositoryID int) ([]dbstore.UploadsWithRepositoryNamespace, error)
	LastUploadRetentionScanForRepository(ctx context.Context, repositoryID int) (*time.Time, error)
	RequestLanguageSupport(ctx context.Context, userID int, language string) error
	LanguagesRequestedBy(ctx context.Context, userID int) ([]string, error)
	GetAuditLogsForUpload(ctx context.Context, uploadID int) ([]dbstore.UploadLog, error)
}

type LSIFStore interface {
	DocumentPaths(ctx context.Context, bundleID int, path string) ([]string, int, error)
}

type CodeNavResolver interface {
	GitBlobLSIFDataResolverFactory(ctx context.Context, repo *types.Repo, commit, path, toolName string, exactPath bool) (_ codenavgraphql.GitBlobLSIFDataResolver, err error)
}

type PoliciesResolver interface {
	PolicyResolverFactory(ctx context.Context) (_ policiesgraphql.PolicyResolver, err error)
}

type AutoIndexingResolver interface {
	GetIndexByID(ctx context.Context, id int) (_ autoindexingShared.Index, _ bool, err error)
	GetIndexesByIDs(ctx context.Context, ids ...int) (_ []autoindexingShared.Index, err error)
	GetRecentIndexesSummary(ctx context.Context, repositoryID int) (summaries []autoindexingShared.IndexesWithRepositoryNamespace, err error)
	GetLastIndexScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error)
	DeleteIndexByID(ctx context.Context, id int) (err error)
	QueueAutoIndexJobsForRepo(ctx context.Context, repositoryID int, rev, configuration string) ([]autoindexingShared.Index, error)

	GetIndexConfiguration(ctx context.Context, repositoryID int) ([]byte, bool, error)                                        // GetIndexConfigurationByRepositoryID
	InferedIndexConfiguration(ctx context.Context, repositoryID int, commit string) (*config.IndexConfiguration, bool, error) // in the service InferIndexConfiguration first return
	InferedIndexConfigurationHints(ctx context.Context, repositoryID int, commit string) ([]config.IndexJobHint, error)       // in the service InferIndexConfiguration second return
	UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, configuration string) error                 // simple dbstore

	IndexConnectionResolverFromFactory(opts autoindexingShared.GetIndexesOptions) *autoindexinggraphql.IndexesResolver
}
