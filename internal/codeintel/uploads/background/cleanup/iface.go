package cleanup

import (
	"context"
	"time"

	sharedIndexes "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	sharedUploads "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type UploadService interface {
	GetStaleSourcedCommits(ctx context.Context, threshold time.Duration, limit int, now time.Time) ([]sharedUploads.SourcedCommits, error)
	DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (int, int, error)
	UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (int, error)

	DeleteUploadsWithoutRepository(ctx context.Context, now time.Time) (map[int]int, error)
	DeleteUploadsStuckUploading(ctx context.Context, uploadedBefore time.Time) (int, error)
	SoftDeleteExpiredUploads(ctx context.Context) (int, error)
	HardDeleteExpiredUploads(ctx context.Context) (int, error)

	DeleteOldAuditLogs(ctx context.Context, maxAge time.Duration, now time.Time) (int, error)

	// Workerutil
	WorkerutilStore() dbworkerstore.Store
}

type AutoIndexingService interface {
	GetStaleSourcedCommits(ctx context.Context, threshold time.Duration, limit int, now time.Time) ([]sharedIndexes.SourcedCommits, error)
	DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (int, error)
	UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (int, error)

	DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (map[int]int, error)
}
