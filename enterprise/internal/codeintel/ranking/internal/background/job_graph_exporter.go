package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewSymbolExporter(
	observationCtx *observation.Context,
	store store.Store,
	lsifstore lsifstore.LsifStore,
	numRoutines int,
	interval time.Duration,
	readBatchSize int,
	writeBatchSize int,
) goroutine.BackgroundRoutine {
	metrics := NewMetrics(observationCtx)

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"ranking.symbol-exporter", "exports SCIP data to ranking definitions and reference tables",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			if err := exportRankingGraph(ctx, store, lsifstore, metrics, observationCtx.Logger, numRoutines, readBatchSize, writeBatchSize); err != nil {
				return err
			}

			if err := vacuumRankingGraph(ctx, store, metrics); err != nil {
				return err
			}

			return nil
		}),
	)
}

func NewMapper(
	observationCtx *observation.Context,
	store store.Store,
	interval time.Duration,
	batchSize int,
) goroutine.BackgroundRoutine {
	operations := newMapperOperations(observationCtx)

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"ranking.file-reference-count-mapper", "maps definitions and references data to path_counts_inputs table in store",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			numReferenceRecordsProcessed, numInputsInserted, err := mapRankingGraph(ctx, store, batchSize)
			if err != nil {
				return err
			}

			operations.numReferenceRecordsProcessed.Add(float64(numReferenceRecordsProcessed))
			operations.numInputsInserted.Add(float64(numInputsInserted))
			return nil
		}),
	)
}

func NewReducer(
	observationCtx *observation.Context,
	store store.Store,
	interval time.Duration,
	batchSize int,
) goroutine.BackgroundRoutine {
	operations := newReducerOperations(observationCtx)

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"ranking.file-reference-count-reducer", "reduces path_counts_inputs into a count of paths per repository and stores it in path_ranks table in store.",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			numPathRanksInserted, numPathCountsInputsProcessed, err := reduceRankingGraph(ctx, store, batchSize)
			if err != nil {
				return err
			}

			operations.numPathCountsInputsRowsProcessed.Add(numPathCountsInputsProcessed)
			operations.numPathRanksInserted.Add(numPathRanksInserted)
			return nil
		}),
	)
}
