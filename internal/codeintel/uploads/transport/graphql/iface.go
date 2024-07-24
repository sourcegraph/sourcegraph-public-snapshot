package graphql

import (
	"context"
	"time"

	autoindexingshared "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	policiesshared "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	uploadshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
)

type UploadsService interface {
	GetAutoIndexJobsByIDs(ctx context.Context, ids ...int) (_ []shared.AutoIndexJob, err error)
	GetUploadsByIDs(ctx context.Context, ids ...int) (_ []shared.Upload, err error)
	GetAutoIndexJobs(ctx context.Context, opts uploadshared.GetAutoIndexJobsOptions) (_ []uploadsshared.AutoIndexJob, _ int, err error)
	GetUploads(ctx context.Context, opts uploadshared.GetUploadsOptions) (uploads []shared.Upload, totalCount int, err error)
	GetAuditLogsForUpload(ctx context.Context, uploadID int) (_ []shared.UploadLog, err error)
	GetAutoIndexJobByID(ctx context.Context, id int) (_ uploadsshared.AutoIndexJob, _ bool, err error)
	DeleteAutoIndexJobByID(ctx context.Context, id int) (_ bool, err error)
	DeleteAutoIndexJobs(ctx context.Context, opts uploadshared.DeleteAutoIndexJobsOptions) (err error)
	SetRerunAutoIndexJobByID(ctx context.Context, id int) (err error)
	SetRerunAutoIndexJobs(ctx context.Context, opts uploadshared.SetRerunAutoIndexJobsOptions) (err error)
	GetIndexers(ctx context.Context, opts uploadshared.GetIndexersOptions) ([]string, error)
	GetUploadByID(ctx context.Context, id int) (_ shared.Upload, _ bool, err error)
	DeleteUploadByID(ctx context.Context, id int) (_ bool, err error)
	DeleteUploads(ctx context.Context, opts uploadshared.DeleteUploadsOptions) (err error)
	ReindexUploads(ctx context.Context, opts uploadshared.ReindexUploadsOptions) error
	ReindexUploadByID(ctx context.Context, id int) error
	GetCommitGraphMetadata(ctx context.Context, repositoryID int) (stale bool, updatedAt *time.Time, err error)
	GetRecentUploadsSummary(ctx context.Context, repositoryID int) ([]uploadshared.UploadsWithRepositoryNamespace, error)
	GetLastUploadRetentionScanForRepository(ctx context.Context, repositoryID int) (*time.Time, error)
	GetRecentAutoIndexJobsSummary(ctx context.Context, repositoryID int) ([]uploadshared.GroupedAutoIndexJobs, error)
	NumRepositoriesWithCodeIntelligence(ctx context.Context) (int, error)
	RepositoryIDsWithErrors(ctx context.Context, offset, limit int) (_ []uploadshared.RepositoryWithCount, totalCount int, err error)
}

type AutoIndexingService interface {
	RepositoryIDsWithConfiguration(ctx context.Context, offset, limit int) (_ []uploadshared.RepositoryWithAvailableIndexers, totalCount int, err error)
	InferIndexJobsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string, localOverrideScript string, bypassLimit bool) (*autoindexingshared.InferenceResult, error)
	GetLastIndexScanForRepository(ctx context.Context, repositoryID int) (*time.Time, error)
}

type PolicyService interface {
	GetRetentionPolicyOverview(ctx context.Context, upload shared.Upload, matchesOnly bool, first int, after int64, query string, now time.Time) (matches []policiesshared.RetentionPolicyMatchCandidate, totalCount int, err error)
}
