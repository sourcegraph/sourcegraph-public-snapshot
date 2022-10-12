package store

import (
	"context"
	"time"

	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// Store provides the interface for autoindexing storage.
type Store interface {
	// Transactions
	Transact(ctx context.Context) (Store, error)
	Done(err error) error

	// Commits
	GetStaleSourcedCommits(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int, now time.Time) (_ []shared.SourcedCommits, err error)
	UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (indexesUpdated int, err error)
	DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration) (indexesDeleted int, err error)

	// Indexes
	InsertIndexes(ctx context.Context, indexes []types.Index) (_ []types.Index, err error)
	GetIndexes(ctx context.Context, opts types.GetIndexesOptions) (_ []types.Index, _ int, err error)
	GetIndexByID(ctx context.Context, id int) (_ types.Index, _ bool, err error)
	GetIndexesByIDs(ctx context.Context, ids ...int) (_ []types.Index, err error)
	GetRecentIndexesSummary(ctx context.Context, repositoryID int) (summaries []shared.IndexesWithRepositoryNamespace, err error)
	GetLastIndexScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error)
	DeleteIndexByID(ctx context.Context, id int) (_ bool, err error)
	DeleteIndexes(ctx context.Context, opts types.DeleteIndexesOptions) (err error)
	DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (_ map[int]int, err error)
	IsQueued(ctx context.Context, repositoryID int, commit string) (_ bool, err error)
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

	GetUnsafeDB() database.DB

	WorkerutilStore(observationContext *observation.Context) dbworkerstore.Store
	WorkerutilDependencySyncStore(observationContext *observation.Context) dbworkerstore.Store
	WorkerutilDependencyIndexStore(observationContext *observation.Context) dbworkerstore.Store
	InsertDependencyIndexingJob(ctx context.Context, uploadID int, externalServiceKind string, syncTime time.Time) (id int, err error)
}

// store manages the autoindexing store.
type store struct {
	db         *basestore.Store
	logger     logger.Logger
	operations *operations
}

// New returns a new autoindexing store.
func New(db database.DB, observationContext *observation.Context) Store {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		logger:     logger.Scoped("autoindexing.store", ""),
		operations: newOperations(observationContext),
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

// GetUnsafeDB returns the underlying database handle. This is used by the
// resolvers that have the old convention of using the database handle directly.
func (s *store) GetUnsafeDB() database.DB {
	return database.NewDBWith(s.logger, s.db)
}
