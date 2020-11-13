package background

import (
	"context"
	"regexp"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
)

type DBStore interface {
	basestore.ShareableStore
	gitserver.DBStore

	Handle() *basestore.TransactableHandle
	Transact(ctx context.Context) (DBStore, error)
	Done(err error) error

	Lock(ctx context.Context, key int, blocking bool) (bool, dbstore.UnlockFunc, error)
	GetUploads(ctx context.Context, opts dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error)
	DeleteUploadsWithoutRepository(ctx context.Context, now time.Time) (map[int]int, error)
	HardDeleteUploadByID(ctx context.Context, ids ...int) error
	SoftDeleteOldDumps(ctx context.Context, maxAge time.Duration, now time.Time) (int, error)
	DirtyRepositories(ctx context.Context) (map[int]int, error)
	CalculateVisibleUploads(ctx context.Context, repositoryID int, graph map[string][]string, tipCommit string, dirtyToken int) error
	IndexableRepositories(ctx context.Context, opts dbstore.IndexableRepositoryQueryOptions) ([]dbstore.IndexableRepository, error)
	UpdateIndexableRepository(ctx context.Context, indexableRepository dbstore.UpdateableIndexableRepository, now time.Time) error
	ResetIndexableRepositories(ctx context.Context, lastUpdatedBefore time.Time) error
	IsQueued(ctx context.Context, repositoryID int, commit string) (bool, error)
	InsertIndex(ctx context.Context, index dbstore.Index) (int, error)
	DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (map[int]int, error)
	RepoUsageStatistics(ctx context.Context) ([]dbstore.RepoUsageStatistics, error)
	GetRepositoriesWithIndexConfiguration(ctx context.Context) ([]int, error)
	GetIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int) (dbstore.IndexConfiguration, bool, error)
	DeleteUploadsStuckUploading(ctx context.Context, uploadedBefore time.Time) (_ int, err error)
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
	Head(ctx context.Context, repositoryID int) (string, error)
	ListFiles(ctx context.Context, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error)
	FileExists(ctx context.Context, repositoryID int, commit, file string) (bool, error)
	RawContents(ctx context.Context, repositoryID int, commit, file string) ([]byte, error)
	CommitGraph(ctx context.Context, repositoryID int, options gitserver.CommitGraphOptions) (map[string][]string, error)
}
