package graphql

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

type UploadService interface {
	GetCommitGraphMetadata(ctx context.Context, repositoryID int) (stale bool, updatedAt *time.Time, err error)
	GetAuditLogsForUpload(ctx context.Context, uploadID int) (_ []types.UploadLog, err error)
	GetListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) (_ []*gitdomain.Tag, err error)
	GetUploadDocumentsForPath(ctx context.Context, bundleID int, pathPattern string) (_ []string, _ int, err error)
	GetUploads(ctx context.Context, opts types.GetUploadsOptions) (uploads []types.Upload, totalCount int, err error)
	GetUploadsByIDs(ctx context.Context, ids ...int) (_ []types.Upload, err error)
	DeleteUploadByID(ctx context.Context, id int) (_ bool, err error)
	DeleteUploads(ctx context.Context, opts types.DeleteUploadsOptions) (err error)
}

type AutoIndexingService interface {
	GetIndexByID(ctx context.Context, id int) (_ types.Index, _ bool, err error)
	GetIndexes(ctx context.Context, opts types.GetIndexesOptions) (_ []types.Index, _ int, err error)
	GetIndexesByIDs(ctx context.Context, ids ...int) (_ []types.Index, err error)
	GetListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) (_ []*gitdomain.Tag, err error)
	GetUnsafeDB() database.DB
}

type PolicyService interface {
	GetRetentionPolicyOverview(ctx context.Context, upload types.Upload, matchesOnly bool, first int, after int64, query string, now time.Time) (matches []types.RetentionPolicyMatchCandidate, totalCount int, err error)
}
