package cleanup

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type DBStore interface {
	basestore.ShareableStore

	Transact(ctx context.Context) (DBStore, error)
	Done(err error) error

	DeleteUploadsWithoutRepository(ctx context.Context, now time.Time) (map[int]int, error)
	DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (map[int]int, error)
	DeleteUploadsStuckUploading(ctx context.Context, uploadedBefore time.Time) (int, error)
	SoftDeleteExpiredUploads(ctx context.Context) (int, error)
	GetUploads(ctx context.Context, opts dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error)
	DeleteOldAuditLogs(ctx context.Context, maxAge time.Duration, now time.Time) (int, error)
	HardDeleteUploadByID(ctx context.Context, ids ...int) error
}

type LSIFStore interface {
	Transact(ctx context.Context) (LSIFStore, error)
	Done(err error) error
	Clear(ctx context.Context, bundleIDs ...int) error
}

type UploadService interface {
	StaleSourcedCommits(ctx context.Context, threshold time.Duration, limit int, now time.Time) ([]shared.SourcedCommits, error)
	DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (int, int, int, error)
	UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (int, int, error)
}

type DBStoreShim struct{ *dbstore.Store }

func (s DBStoreShim) Transact(ctx context.Context) (DBStore, error) { return s, nil }

type LSIFStoreShim struct{ *lsifstore.Store }

func (s LSIFStoreShim) Transact(ctx context.Context) (LSIFStore, error) { return s, nil }
