package resolvers

import (
	"context"
	"time"

	autoindexingShared "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	autoindexinggraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/transport/graphql"
	codenavgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/transport/graphql"
	policiesgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/transport/graphql"
	uploadsShared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	uploadsgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

type DBStore interface {
	RequestLanguageSupport(ctx context.Context, userID int, language string) error
	LanguagesRequestedBy(ctx context.Context, userID int) ([]string, error)
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

type UploadsResolver interface {
	GetUploadsByIDs(ctx context.Context, ids ...int) (_ []uploadsShared.Upload, err error)
	GetUploadDocumentsForPath(ctx context.Context, bundleID int, pathPattern string) ([]string, int, error)
	GetRecentUploadsSummary(ctx context.Context, repositoryID int) (upload []uploadsShared.UploadsWithRepositoryNamespace, err error)
	GetLastUploadRetentionScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error)
	DeleteUploadByID(ctx context.Context, id int) (_ bool, err error)
	GetAuditLogsForUpload(ctx context.Context, uploadID int) (_ []uploadsShared.UploadLog, err error)
	UploadsConnectionResolverFromFactory(opts uploadsShared.GetUploadsOptions) *uploadsgraphql.UploadsResolver
	CommitGraphResolverFromFactory(ctx context.Context, repositoryID int) *uploadsgraphql.CommitGraphResolver
}
