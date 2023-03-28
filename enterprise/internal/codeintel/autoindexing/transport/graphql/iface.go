package graphql

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	uploadshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	uploadsgraphql "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

type AutoIndexingService interface {
	sharedresolvers.AutoIndexingService

	GetIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int) (_ shared.IndexConfiguration, _ bool, err error)
	GetRecentIndexesSummary(ctx context.Context, repositoryID int) (summaries []shared.IndexesWithRepositoryNamespace, err error)
	GetLastIndexScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error)
	UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, data []byte) (err error)
	GetInferenceScript(ctx context.Context) (script string, err error)
	SetInferenceScript(ctx context.Context, script string) (err error)

	InferIndexJobsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string, localOverrideScript string, bypassLimit bool) ([]config.IndexJob, error)
	InferIndexConfiguration(ctx context.Context, repositoryID int, commit string, localOverrideScript string, bypassLimit bool) (_ *config.IndexConfiguration, hints []config.IndexJobHint, err error)
	QueueIndexes(ctx context.Context, repositoryID int, rev, configuration string, force bool, bypassLimit bool) (_ []types.Index, err error)

	GetSupportedByCtags(ctx context.Context, filepath string, repoName api.RepoName) (bool, string, error)
	GetLanguagesRequestedBy(ctx context.Context, userID int) (_ []string, err error)
	SetRequestLanguageSupport(ctx context.Context, userID int, language string) (err error)
}

type UploadsService interface {
	sharedresolvers.UploadsService
	uploadsgraphql.UploadsService

	GetIndexByID(ctx context.Context, id int) (_ types.Index, _ bool, err error)
	DeleteIndexByID(ctx context.Context, id int) (_ bool, err error)
	DeleteIndexes(ctx context.Context, opts uploadshared.DeleteIndexesOptions) (err error)
	ReindexIndexByID(ctx context.Context, id int) (err error)
	ReindexIndexes(ctx context.Context, opts uploadshared.ReindexIndexesOptions) (err error)

	GetLastUploadRetentionScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error)
	GetRecentUploadsSummary(ctx context.Context, repositoryID int) (upload []uploadshared.UploadsWithRepositoryNamespace, err error)
	GetIndexers(ctx context.Context, opts uploadshared.GetIndexersOptions) ([]string, error)
	GetUploadByID(ctx context.Context, id int) (_ types.Upload, _ bool, err error)
	DeleteUploadByID(ctx context.Context, id int) (_ bool, err error)
	DeleteUploads(ctx context.Context, opts uploadshared.DeleteUploadsOptions) (err error)
	ReindexUploads(ctx context.Context, opts uploadshared.ReindexUploadsOptions) error
	ReindexUploadByID(ctx context.Context, id int) error
}

type PolicyService = sharedresolvers.PolicyService
