package resolvers

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
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
	GetIndexByID(ctx context.Context, id int) (dbstore.Index, bool, error)
	GetIndexesByIDs(ctx context.Context, ids ...int) ([]dbstore.Index, error)
	GetIndexes(ctx context.Context, opts dbstore.GetIndexesOptions) ([]dbstore.Index, int, error)
	DeleteIndexByID(ctx context.Context, id int) (bool, error)

	GetIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int) (dbstore.IndexConfiguration, bool, error)
	UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, data []byte) error
	RecentUploadsSummary(ctx context.Context, repositoryID int) ([]dbstore.UploadsWithRepositoryNamespace, error)
	RecentIndexesSummary(ctx context.Context, repositoryID int) ([]dbstore.IndexesWithRepositoryNamespace, error)
	LastUploadRetentionScanForRepository(ctx context.Context, repositoryID int) (*time.Time, error)
	LastIndexScanForRepository(ctx context.Context, repositoryID int) (*time.Time, error)
	RequestLanguageSupport(ctx context.Context, userID int, language string) error
	LanguagesRequestedBy(ctx context.Context, userID int) ([]string, error)
	GetAuditLogsForUpload(ctx context.Context, uploadID int) ([]dbstore.UploadLog, error)
}

type LSIFStore interface {
	DocumentPaths(ctx context.Context, bundleID int, path string) ([]string, int, error)
}

type IndexEnqueuer interface {
	QueueIndexes(ctx context.Context, repositoryID int, rev, configuration string, force, bypassLimit bool) ([]dbstore.Index, error)
	InferIndexConfiguration(ctx context.Context, repositoryID int, commit string, bypassLimit bool) (*config.IndexConfiguration, []config.IndexJobHint, error)
}

type CodeNavResolver interface {
	GitBlobLSIFDataResolverFactory(ctx context.Context, repo *types.Repo, commit, path, toolName string, exactPath bool) (_ codenavgraphql.GitBlobLSIFDataResolver, err error)
}

type PoliciesResolver interface {
	PolicyResolverFactory(ctx context.Context) (_ policiesgraphql.PolicyResolver, err error)
}

type (
	RepoUpdaterClient       = autoindexing.RepoUpdaterClient
	EnqueuerDBStore         = autoindexing.DBStore
	EnqueuerGitserverClient = autoindexing.GitserverClient
)
