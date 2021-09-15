package janitor

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type DBStore interface {
	basestore.ShareableStore

	Handle() *basestore.TransactableHandle
	Transact(ctx context.Context) (DBStore, error)
	Done(err error) error

	GetUploads(ctx context.Context, opts dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error)
	DeleteUploadsWithoutRepository(ctx context.Context, now time.Time) (map[int]int, error)
	HardDeleteUploadByID(ctx context.Context, ids ...int) error
	GetConfigurationPolicies(ctx context.Context, opts dbstore.GetConfigurationPoliciesOptions) ([]dbstore.ConfigurationPolicy, error)
	SelectRepositoriesForRetentionScan(ctx context.Context, processDelay time.Duration, limit int) (map[int]*time.Time, error)
	CommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) ([]string, *string, error)
	UpdateUploadRetention(ctx context.Context, protectedIDs, expiredIDs []int) error
	SoftDeleteExpiredUploads(ctx context.Context) (int, error)
	DirtyRepositories(ctx context.Context) (map[int]int, error)
	DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (map[int]int, error)
	DeleteUploadsStuckUploading(ctx context.Context, uploadedBefore time.Time) (int, error)
	StaleSourcedCommits(ctx context.Context, threshold time.Duration, limit int, now time.Time) ([]dbstore.SourcedCommits, error)
	RefreshCommitResolvability(ctx context.Context, repositoryID int, commit string, delete bool, now time.Time) (int, int, error)
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
	Clear(ctx context.Context, bundleIDs ...int) error
}

type GitserverClient interface {
	RefDescriptions(ctx context.Context, repositoryID int) (map[string][]gitserver.RefDescription, error)
	BranchesContaining(ctx context.Context, repositoryID int, commit string) ([]string, error)
}
