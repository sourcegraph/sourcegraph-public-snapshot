package store

import (
	"context"
	"time"

	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Store provides the interface for autoindexing storage.
type Store interface {
	Transact(ctx context.Context) (Store, error)
	Done(err error) error

	InsertIndexes(ctx context.Context, indexes []types.Index) ([]types.Index, error)
	GetLastIndexScanForRepository(ctx context.Context, repositoryID int) (*time.Time, error)
	IsQueued(ctx context.Context, repositoryID int, commit string) (bool, error)
	IsQueuedRootIndexer(ctx context.Context, repositoryID int, commit string, root string, indexer string) (bool, error)
	QueueRepoRev(ctx context.Context, repositoryID int, commit string) error
	GetQueuedRepoRev(ctx context.Context, batchSize int) ([]RepoRev, error)
	MarkRepoRevsAsProcessed(ctx context.Context, ids []int) error
	GetIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int) (shared.IndexConfiguration, bool, error)
	UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, data []byte) error
	GetInferenceScript(ctx context.Context) (string, error)
	SetInferenceScript(ctx context.Context, script string) error
	RepositoryIDsWithConfiguration(ctx context.Context, offset, limit int) ([]shared.RepositoryWithAvailableIndexers, int, error)
	TopRepositoriesToConfigure(ctx context.Context, limit int) ([]shared.RepositoryWithCount, error)
	SetConfigurationSummary(ctx context.Context, repositoryID int, numEvents int, availableIndexers map[string]shared.AvailableIndexer) error
	TruncateConfigurationSummary(ctx context.Context, numRecordsToRetain int) error
	GetRepositoriesForIndexScan(ctx context.Context, table, column string, processDelay time.Duration, allowGlobalPolicies bool, repositoryMatchLimit *int, limit int, now time.Time) ([]int, error)
	InsertDependencyIndexingJob(ctx context.Context, uploadID int, externalServiceKind string, syncTime time.Time) (int, error)
	GetRepoName(ctx context.Context, repositoryID int) (string, error)
}

// store manages the autoindexing store.
type store struct {
	db         *basestore.Store
	logger     logger.Logger
	operations *operations
}

// New returns a new autoindexing store.
func New(observationCtx *observation.Context, db database.DB) Store {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		logger:     logger.Scoped("autoindexing.store", ""),
		operations: newOperations(observationCtx),
	}
}

func (s *store) Transact(ctx context.Context) (Store, error) {
	return s.transact(ctx)
}

func (s *store) transact(ctx context.Context) (*store, error) {
	tx, err := s.db.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &store{
		logger:     s.logger,
		db:         tx,
		operations: s.operations,
	}, nil
}

func (s *store) Done(err error) error {
	return s.db.Done(err)
}
