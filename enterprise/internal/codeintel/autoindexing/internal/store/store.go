package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/keegancsmith/sqlf"
	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Store provides the interface for autoindexing storage.
type Store interface {
	// Transactions
	Transact(ctx context.Context) (Store, error)
	Done(err error) error

	Summary(ctx context.Context) (shared.Summary, error)

	// Commits
	ProcessStaleSourcedCommits(
		ctx context.Context,
		minimumTimeSinceLastCheck time.Duration,
		commitResolverBatchSize int,
		commitResolverMaximumCommitLag time.Duration,
		shouldDelete func(ctx context.Context, repositoryID int, commit string) (bool, error),
	) (indexesDeleted int, _ error)

	// Indexes
	InsertIndexes(ctx context.Context, indexes []types.Index) (_ []types.Index, err error)
	GetIndexes(ctx context.Context, opts shared.GetIndexesOptions) (_ []types.Index, _ int, err error)
	GetIndexByID(ctx context.Context, id int) (_ types.Index, _ bool, err error)
	GetIndexesByIDs(ctx context.Context, ids ...int) (_ []types.Index, err error)
	GetRecentIndexesSummary(ctx context.Context, repositoryID int) (summaries []shared.IndexesWithRepositoryNamespace, err error)
	GetLastIndexScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error)
	DeleteIndexByID(ctx context.Context, id int) (_ bool, err error)
	DeleteIndexes(ctx context.Context, opts shared.DeleteIndexesOptions) (err error)
	ReindexIndexByID(ctx context.Context, id int) (err error)
	ReindexIndexes(ctx context.Context, opts shared.ReindexIndexesOptions) (err error)
	DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (_ map[int]int, err error)
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

	GetUnsafeDB() database.DB

	GetRepoName(ctx context.Context, repositoryID int) (_ string, err error)
	TopRepositoriesToConfigure(ctx context.Context, limit int) (_ []int, err error)
	SetConfigurationSummary(ctx context.Context, repositoryID int, availableIndexers map[string]shared.AvailableIndexer) (err error)

	InsertDependencyIndexingJob(ctx context.Context, uploadID int, externalServiceKind string, syncTime time.Time) (id int, err error)
	ExpireFailedRecords(ctx context.Context, batchSize int, failedIndexMaxAge time.Duration, now time.Time) error
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

// GetUnsafeDB returns the underlying database handle. This is used by the
// resolvers that have the old convention of using the database handle directly.
func (s *store) GetUnsafeDB() database.DB {
	return database.NewDBWith(s.logger, s.db)
}

// TODO - test
func (s *store) Summary(ctx context.Context) (shared.Summary, error) {
	numRepositoriesWithCodeIntelligence, _, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(`
		SELECT COUNT(DISTINCT r.id)
		FROM lsif_uploads u
		JOIN repo r ON r.id = u.repository_id
		WHERE
			u.state = 'completed' AND
			r.deleted_at IS NULL AND
			r.blocked IS NULL
	`)))
	if err != nil {
		return shared.Summary{}, err
	}

	// TODO - combine to get count?
	repositoriesWithErrors, err := basestore.NewMapScanner(func(s dbutil.Scanner) (repositoryID int, count int, _ error) {
		if err := s.Scan(&repositoryID, &count); err != nil {
			return 0, 0, err
		}

		return repositoryID, count, nil
	})(s.db.Query(ctx, sqlf.Sprintf(`
		WITH
		ranked_completed_indexes AS (
			SELECT
				u.repository_id,
				u.state,
				RANK() OVER (PARTITION BY repository_id, root, indexer ORDER BY finished_at DESC) AS rank
			FROM lsif_indexes u
			WHERE u.state NOT IN ('queued', 'processing', 'deleted')
		),
		ranked_completed_uploads AS (
			SELECT
				u.repository_id,
				u.state,
				RANK() OVER (PARTITION BY repository_id, root, indexer ORDER BY finished_at DESC) AS rank
			FROM lsif_uploads u
			WHERE u.state NOT IN ('uploading', 'queued', 'processing', 'deleted')
		),
		ranked_uploads_and_indexes AS (
			SELECT repository_id FROM ranked_completed_indexes WHERE rank = 1 AND state = 'failed'
			UNION ALL
			SELECT repository_id FROM ranked_completed_uploads WHERE rank = 1 AND state = 'failed'
		)
		SELECT
			r.id,
			COUNT(*) AS count
		FROM repo r
		JOIN ranked_uploads_and_indexes rui ON rui.repository_id = r.id
		WHERE
			r.deleted_at IS NULL AND
			r.blocked IS NULL
		GROUP BY r.id
	`)))
	if err != nil {
		return shared.Summary{}, err
	}

	repositoryIDsWithConfiguration, err := basestore.NewMapScanner(func(s dbutil.Scanner) (
		repositoryID int,
		availableIndexers map[string]shared.AvailableIndexer,
		_ error,
	) {
		var payload string
		if err := s.Scan(&repositoryID, &payload); err != nil {
			return 0, nil, err
		}

		if err := json.Unmarshal([]byte(payload), &availableIndexers); err != nil {
			return 0, nil, err
		}

		return repositoryID, availableIndexers, nil
	})(s.db.Query(ctx, sqlf.Sprintf(`
		SELECT repository_id, available_indexers FROM cached_available_indexers
	`)))
	if err != nil {
		return shared.Summary{}, err
	}

	return shared.Summary{
		NumRepositoriesWithCodeIntelligence: numRepositoriesWithCodeIntelligence,
		RepositoryIDsWithErrors:             repositoriesWithErrors,
		RepositoryIDsWithConfiguration:      repositoryIDsWithConfiguration,
	}, nil
}
