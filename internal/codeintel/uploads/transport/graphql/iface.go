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
	GetIndexesByIDs(ctx context.Context, ids ...int) (_ []shared.Index, err error)
	GetUploadsByIDs(ctx context.Context, ids ...int) (_ []shared.Upload, err error)
	GetIndexes(ctx context.Context, opts uploadshared.GetIndexesOptions) (_ []uploadsshared.Index, _ int, err error)
	GetUploads(ctx context.Context, opts uploadshared.GetUploadsOptions) (uploads []shared.Upload, totalCount int, err error)
	GetAuditLogsForUpload(ctx context.Context, uploadID int) (_ []shared.UploadLog, err error)
	GetIndexByID(ctx context.Context, id int) (_ uploadsshared.Index, _ bool, err error)
	DeleteIndexByID(ctx context.Context, id int) (_ bool, err error)
	DeleteIndexes(ctx context.Context, opts uploadshared.DeleteIndexesOptions) (err error)
	ReindexIndexByID(ctx context.Context, id int) (err error)
	ReindexIndexes(ctx context.Context, opts uploadshared.ReindexIndexesOptions) (err error)
	GetIndexers(ctx context.Context, opts uploadshared.GetIndexersOptions) ([]string, error)
	GetUploadByID(ctx context.Context, id int) (_ shared.Upload, _ bool, err error)
	DeleteUploadByID(ctx context.Context, id int) (_ bool, err error)
	DeleteUploads(ctx context.Context, opts uploadshared.DeleteUploadsOptions) (err error)
	ReindexUploads(ctx context.Context, opts uploadshared.ReindexUploadsOptions) error
	ReindexUploadByID(ctx context.Context, id int) error
	GetCommitGraphMetadata(ctx context.Context, repositoryID int) (stale bool, updatedAt *time.Time, err error)
	GetRecentUploadsSummary(ctx context.Context, repositoryID int) ([]uploadshared.UploadsWithRepositoryNamespace, error)
	GetLastUploadRetentionScanForRepository(ctx context.Context, repositoryID int) (*time.Time, error)
	GetRecentIndexesSummary(ctx context.Context, repositoryID int) ([]uploadshared.IndexesWithRepositoryNamespace, error)
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
