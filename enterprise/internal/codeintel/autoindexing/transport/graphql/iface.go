package graphql

import (
	"context"
	"time"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	uploadshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

type AutoIndexingService interface {
	GetIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int) (_ shared.IndexConfiguration, _ bool, err error)
	GetIndexes(ctx context.Context, opts shared.GetIndexesOptions) (_ []types.Index, _ int, err error)
	GetIndexByID(ctx context.Context, id int) (_ types.Index, _ bool, err error)
	GetIndexesByIDs(ctx context.Context, ids ...int) (_ []types.Index, err error)
	GetRecentIndexesSummary(ctx context.Context, repositoryID int) (summaries []shared.IndexesWithRepositoryNamespace, err error)
	GetLastIndexScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error)
	UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, data []byte) (err error)
	DeleteIndexByID(ctx context.Context, id int) (_ bool, err error)
	DeleteIndexes(ctx context.Context, opts shared.DeleteIndexesOptions) (err error)
	ReindexIndexByID(ctx context.Context, id int) (err error)
	ReindexIndexes(ctx context.Context, opts shared.ReindexIndexesOptions) (err error)
	GetInferenceScript(ctx context.Context) (script string, err error)
	SetInferenceScript(ctx context.Context, script string) (err error)

	InferIndexJobHintsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string) ([]config.IndexJobHint, error)
	InferIndexJobsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string, bypassLimit bool) ([]config.IndexJob, error)
	InferIndexConfiguration(ctx context.Context, repositoryID int, commit string, bypassLimit bool) (_ *config.IndexConfiguration, hints []config.IndexJobHint, err error)
	QueueIndexes(ctx context.Context, repositoryID int, rev, configuration string, force, bypassLimit bool) (_ []types.Index, err error)

	GetListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) (_ []*gitdomain.Tag, err error)
	ListFiles(ctx context.Context, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error)

	GetSupportedByCtags(ctx context.Context, filepath string, repoName api.RepoName) (bool, string, error)
	GetLanguagesRequestedBy(ctx context.Context, userID int) (_ []string, err error)
	SetRequestLanguageSupport(ctx context.Context, userID int, language string) (err error)

	GetUnsafeDB() database.DB
}

type UploadsService interface {
	GetLastUploadRetentionScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error)
	GetRecentUploadsSummary(ctx context.Context, repositoryID int) (upload []uploadshared.UploadsWithRepositoryNamespace, err error)
	GetUploads(ctx context.Context, opts uploadshared.GetUploadsOptions) (uploads []types.Upload, totalCount int, err error)
	GetAuditLogsForUpload(ctx context.Context, uploadID int) (_ []types.UploadLog, err error)
	GetListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) (_ []*gitdomain.Tag, err error)
	GetUploadDocumentsForPath(ctx context.Context, bundleID int, pathPattern string) ([]string, int, error)
	GetUploadsByIDs(ctx context.Context, ids ...int) (_ []types.Upload, err error)
	GetUploadByID(ctx context.Context, id int) (_ types.Upload, _ bool, err error)
	DeleteUploadByID(ctx context.Context, id int) (_ bool, err error)
	DeleteUploads(ctx context.Context, opts uploadshared.DeleteUploadsOptions) (err error)
	ReindexUploads(ctx context.Context, opts uploadshared.ReindexUploadsOptions) error
	ReindexUploadByID(ctx context.Context, id int) error
}

type PolicyService interface {
	GetRetentionPolicyOverview(ctx context.Context, upload types.Upload, matchesOnly bool, first int, after int64, query string, now time.Time) (matches []types.RetentionPolicyMatchCandidate, totalCount int, err error)
}
