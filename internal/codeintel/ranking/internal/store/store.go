package store

import (
	"context"
	"time"

	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Store interface {
	WithTransaction(ctx context.Context, f func(tx Store) error) error

	// Metadata
	Summaries(ctx context.Context) ([]shared.Summary, error)

	// Retrieval
	GetStarRank(ctx context.Context, repoName api.RepoName) (float64, error)
	GetDocumentRanks(ctx context.Context, repoName api.RepoName) (map[string]float64, bool, error)
	GetReferenceCountStatistics(ctx context.Context) (logmean float64, _ error)
	CoverageCounts(ctx context.Context, graphKey string) (_ shared.CoverageCounts, err error)
	LastUpdatedAt(ctx context.Context, repoIDs []api.RepoID) (map[api.RepoID]time.Time, error)

	// Export uploads (metadata tracking) + cleanup
	GetUploadsForRanking(ctx context.Context, graphKey, objectPrefix string, batchSize int) ([]uploadsshared.ExportedUpload, error)
	VacuumAbandonedExportedUploads(ctx context.Context, graphKey string, batchSize int) (int, error)
	SoftDeleteStaleExportedUploads(ctx context.Context, graphKey string) (numExportedUploadRecordsScanned int, numStaleExportedUploadRecordsDeleted int, _ error)
	VacuumDeletedExportedUploads(ctx context.Context, derivativeGraphKey string) (int, error)

	// Exported data (raw)
	InsertDefinitionsForRanking(ctx context.Context, graphKey string, definitions chan shared.RankingDefinitions) error
	InsertReferencesForRanking(ctx context.Context, graphKey string, batchSize int, exportedUploadID int, references chan [16]byte) error
	InsertInitialPathRanks(ctx context.Context, exportedUploadID int, documentPaths []string, batchSize int, graphKey string) error

	// Graph keys
	DerivativeGraphKey(ctx context.Context) (string, time.Time, bool, error)
	BumpDerivativeGraphKey(ctx context.Context) error
	DeleteRankingProgress(ctx context.Context, graphKey string) error

	// Coordinates mapper+reducer phases
	Coordinate(ctx context.Context, derivativeGraphKey string) error

	// Mapper behavior + cleanup
	InsertPathCountInputs(ctx context.Context, derivativeGraphKey string, batchSize int) (numReferenceRecordsProcessed int, numInputsInserted int, err error)
	InsertInitialPathCounts(ctx context.Context, derivativeGraphKey string, batchSize int) (numInitialPathsProcessed int, numInitialPathRanksInserted int, err error)
	VacuumStaleProcessedReferences(ctx context.Context, derivativeGraphKey string, batchSize int) (processedReferencesDeleted int, _ error)
	VacuumStaleProcessedPaths(ctx context.Context, derivativeGraphKey string, batchSize int) (processedPathsDeleted int, _ error)
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
		logger:     logger.Scoped("ranking.store"),
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
