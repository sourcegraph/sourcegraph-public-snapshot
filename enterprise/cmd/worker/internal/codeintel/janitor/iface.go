package janitor

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type DBStore interface {
	basestore.ShareableStore

	Transact(ctx context.Context) (DBStore, error)
	Done(err error) error

	GetUploads(ctx context.Context, opts dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error)
	HardDeleteUploadByID(ctx context.Context, ids ...int) error
	GetConfigurationPolicies(ctx context.Context, opts dbstore.GetConfigurationPoliciesOptions) ([]dbstore.ConfigurationPolicy, int, error)
	SelectRepositoriesForRetentionScan(ctx context.Context, processDelay time.Duration, limit int) ([]int, error)
	CommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) ([]string, *string, error)
	UpdateUploadRetention(ctx context.Context, protectedIDs, expiredIDs []int) error
	SoftDeleteExpiredUploads(ctx context.Context) (int, error)
	DirtyRepositories(ctx context.Context) (map[int]int, error)
	DeleteUploadsStuckUploading(ctx context.Context, uploadedBefore time.Time) (int, error)
	DeleteOldAuditLogs(ctx context.Context, maxAge time.Duration, now time.Time) (int, error)
}

type DBStoreShim struct {
	*dbstore.Store
}

func (s *DBStoreShim) Transact(ctx context.Context) (DBStore, error) {
	store, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &DBStoreShim{store}, nil
}

type LSIFStore interface {
	Transact(ctx context.Context) (LSIFStore, error)
	Done(err error) error

	Clear(ctx context.Context, bundleIDs ...int) error
}

type LSIFStoreShim struct {
	*lsifstore.Store
}

func (s *LSIFStoreShim) Transact(ctx context.Context) (LSIFStore, error) {
	store, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &LSIFStoreShim{store}, nil
}
