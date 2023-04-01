package graphql

import (
	"context"
	"time"

	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	uploadshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

type UploadsService interface {
	sharedresolvers.UploadsService

	GetIndexes(ctx context.Context, opts uploadshared.GetIndexesOptions) (_ []types.Index, _ int, err error)
	GetUploads(ctx context.Context, opts uploadshared.GetUploadsOptions) (uploads []shared.Upload, totalCount int, err error)
	GetAuditLogsForUpload(ctx context.Context, uploadID int) (_ []shared.UploadLog, err error)
	GetIndexByID(ctx context.Context, id int) (_ types.Index, _ bool, err error)
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
	InferIndexJobsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string, localOverrideScript string, bypassLimit bool) ([]config.IndexJob, error)
	GetLastIndexScanForRepository(ctx context.Context, repositoryID int) (*time.Time, error)
}

type PolicyService interface {
	GetRetentionPolicyOverview(ctx context.Context, upload shared.Upload, matchesOnly bool, first int, after int64, query string, now time.Time) (matches []types.RetentionPolicyMatchCandidate, totalCount int, err error)
}
