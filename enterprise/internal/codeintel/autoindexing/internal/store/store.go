package store

import (
	"context"
	"time"

	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Store interface {
	Transact(ctx context.Context) (Store, error)
	Done(err error) error

	// Index scanning/queueing
	GetRepositoriesForIndexScan(ctx context.Context, table, column string, processDelay time.Duration, allowGlobalPolicies bool, repositoryMatchLimit *int, limit int, now time.Time) ([]int, error)
	GetLastIndexScanForRepository(ctx context.Context, repositoryID int) (*time.Time, error)
	IsQueued(ctx context.Context, repositoryID int, commit string) (bool, error)
	IsQueuedRootIndexer(ctx context.Context, repositoryID int, commit string, root string, indexer string) (bool, error)
	QueueRepoRev(ctx context.Context, repositoryID int, commit string) error
	GetQueuedRepoRev(ctx context.Context, batchSize int) ([]RepoRev, error)
	MarkRepoRevsAsProcessed(ctx context.Context, ids []int) error

	// Manual index insertion
	InsertIndexes(ctx context.Context, indexes []uploadsshared.Index) ([]uploadsshared.Index, error)
	InsertDependencyIndexingJob(ctx context.Context, uploadID int, externalServiceKind string, syncTime time.Time) (int, error)

	// Inference configuration
	GetInferenceScript(ctx context.Context) (string, error)
	SetInferenceScript(ctx context.Context, script string) error

	// Repository configuration
	GetIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int) (shared.IndexConfiguration, bool, error)
	UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, data []byte) error

	// Summary building
	TopRepositoriesToConfigure(ctx context.Context, limit int) ([]uploadsshared.RepositoryWithCount, error)
	RepositoryIDsWithConfiguration(ctx context.Context, offset, limit int) ([]uploadsshared.RepositoryWithAvailableIndexers, int, error)
	SetConfigurationSummary(ctx context.Context, repositoryID int, numEvents int, availableIndexers map[string]uploadsshared.AvailableIndexer) error
	TruncateConfigurationSummary(ctx context.Context, numRecordsToRetain int) error
}

type store struct {
	db         *basestore.Store
	logger     logger.Logger
	operations *operations
}

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
