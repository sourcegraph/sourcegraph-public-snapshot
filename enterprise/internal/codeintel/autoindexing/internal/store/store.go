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
	// Transactions
	Transact(ctx context.Context) (Store, error)
	Done(err error) error

	// Indexes
	InsertIndexes(ctx context.Context, indexes []types.Index) (_ []types.Index, err error)
	GetRecentIndexesSummary(ctx context.Context, repositoryID int) (summaries []shared.IndexesWithRepositoryNamespace, err error)
	GetLastIndexScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error)
	IsQueued(ctx context.Context, repositoryID int, commit string) (_ bool, err error)
	IsQueuedRootIndexer(ctx context.Context, repositoryID int, commit string, root string, indexer string) (_ bool, err error)
	QueueRepoRev(ctx context.Context, repositoryID int, commit string) error
	GetQueuedRepoRev(ctx context.Context, batchSize int) ([]RepoRev, error)
	MarkRepoRevsAsProcessed(ctx context.Context, ids []int) error

	// Index configurations
	GetIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int) (_ shared.IndexConfiguration, _ bool, err error)
	UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, data []byte) (err error)
	GetInferenceScript(ctx context.Context) (script string, err error)
	SetInferenceScript(ctx context.Context, script string) (err error)

	// Language support
	GetLanguagesRequestedBy(ctx context.Context, userID int) (_ []string, err error)
	SetRequestLanguageSupport(ctx context.Context, userID int, language string) (err error)

	GetRepoName(ctx context.Context, repositoryID int) (_ string, err error)
	NumRepositoriesWithCodeIntelligence(ctx context.Context) (int, error)
	RepositoryIDsWithErrors(ctx context.Context, offset, limit int) (_ []shared.RepositoryWithCount, totalCount int, err error)
	RepositoryIDsWithConfiguration(ctx context.Context, offset, limit int) (_ []shared.RepositoryWithAvailableIndexers, totalCount int, err error)
	TopRepositoriesToConfigure(ctx context.Context, limit int) ([]shared.RepositoryWithCount, error)
	SetConfigurationSummary(ctx context.Context, repositoryID int, numEvents int, availableIndexers map[string]shared.AvailableIndexer) (err error)
	TruncateConfigurationSummary(ctx context.Context, numRecordsToRetain int) error

	InsertDependencyIndexingJob(ctx context.Context, uploadID int, externalServiceKind string, syncTime time.Time) (id int, err error)
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
