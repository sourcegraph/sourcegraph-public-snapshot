package store

import (
	"context"
	"time"

	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Store interface {
	// Retrieval
	GetStarRank(ctx context.Context, repoName api.RepoName) (float64, error)
	GetDocumentRanks(ctx context.Context, repoName api.RepoName) (map[string]float64, bool, error)
	GetReferenceCountStatistics(ctx context.Context) (logmean float64, _ error)
	LastUpdatedAt(ctx context.Context, repoIDs []api.RepoID) (map[api.RepoID]time.Time, error)

	// Export uploads (metadata tracking) + cleanup
	GetUploadsForRanking(ctx context.Context, graphKey, objectPrefix string, batchSize int) ([]uploadsshared.ExportedUpload, error)
	ProcessStaleExportedUploads(ctx context.Context, graphKey string, batchSize int, deleter func(ctx context.Context, objectPrefix string) error) (totalDeleted int, _ error)

	// Export definitions + cleanup
	InsertDefinitionsForRanking(ctx context.Context, rankingGraphKey string, definitions chan shared.RankingDefinitions) error
	VacuumAbandonedDefinitions(ctx context.Context, graphKey string, batchSize int) (int, error)
	VacuumStaleDefinitions(ctx context.Context, graphKey string) (numDefinitionRecordsScanned int, numStaleDefinitionRecordsDeleted int, _ error)

	// Export references + cleanup
	InsertReferencesForRanking(ctx context.Context, rankingGraphKey string, batchSize int, uploadID int, references chan string) error
	VacuumAbandonedReferences(ctx context.Context, graphKey string, batchSize int) (int, error)
	VacuumStaleReferences(ctx context.Context, graphKey string) (numReferenceRecordsScanned int, numStaleReferenceRecordsDeleted int, _ error)

	// Export upload paths + cleanup
	InsertInitialPathRanks(ctx context.Context, uploadID int, documentPaths chan string, batchSize int, graphKey string) error
	VacuumAbandonedInitialPathCounts(ctx context.Context, graphKey string, batchSize int) (int, error)
	VacuumStaleInitialPaths(ctx context.Context, graphKey string) (numPathRecordsScanned int, numStalePathRecordsDeleted int, _ error)

	// Mapper behavior + cleanup
	InsertPathCountInputs(ctx context.Context, rankingGraphKey string, batchSize int) (numReferenceRecordsProcessed int, numInputsInserted int, err error)
	InsertInitialPathCounts(ctx context.Context, derivativeGraphKey string, batchSize int) (numInitialPathsProcessed int, numInitialPathRanksInserted int, err error)
	VacuumStaleGraphs(ctx context.Context, derivativeGraphKey string, batchSize int) (inputRecordsDeleted int, _ error)

	// Reducer behavior + cleanup
	InsertPathRanks(ctx context.Context, graphKey string, batchSize int) (numInputsProcessed int, numPathRanksInserted int, _ error)
	VacuumStaleRanks(ctx context.Context, derivativeGraphKey string) (rankRecordsScanned int, rankRecordsSDeleted int, _ error)
}

type store struct {
	db         *basestore.Store
	logger     logger.Logger
	operations *operations
}

// New returns a new ranking store.
func New(observationCtx *observation.Context, db database.DB) Store {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		logger:     logger.Scoped("ranking.store", ""),
		operations: newOperations(observationCtx),
	}
}

func (s *store) WithTransaction(ctx context.Context, f func(s Store) error) error {
	return s.withTransaction(ctx, func(s *store) error { return f(s) })
}

func (s *store) withTransaction(ctx context.Context, f func(s *store) error) error {
	return basestore.InTransaction[*store](ctx, s, f)
}

func (s *store) Transact(ctx context.Context) (*store, error) {
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
