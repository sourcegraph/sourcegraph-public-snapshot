package cleanup

import (
	"context"
	"time"

	sharedIndexes "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	sharedUploads "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type DBStore interface {
	basestore.ShareableStore

	Transact(ctx context.Context) (DBStore, error)
	Done(err error) error
}

type UploadService interface {
	StaleSourcedCommits(ctx context.Context, threshold time.Duration, limit int, now time.Time) ([]sharedUploads.SourcedCommits, error)
	DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (int, int, error)
	UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (int, error)

	DeleteUploadsWithoutRepository(ctx context.Context, now time.Time) (map[int]int, error)
	DeleteUploadsStuckUploading(ctx context.Context, uploadedBefore time.Time) (int, error)
	SoftDeleteExpiredUploads(ctx context.Context) (int, error)
	HardDeleteExpiredUploads(ctx context.Context) (int, error)

	DeleteOldAuditLogs(ctx context.Context, maxAge time.Duration, now time.Time) (int, error)
}

type AutoIndexingService interface {
	StaleSourcedCommits(ctx context.Context, threshold time.Duration, limit int, now time.Time) ([]sharedIndexes.SourcedCommits, error)
	DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (int, error)
	UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (int, error)

	DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (map[int]int, error)
}

type DBStoreShim struct{ *dbstore.Store }

func (s DBStoreShim) Transact(ctx context.Context) (DBStore, error) { return s, nil }
