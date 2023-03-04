package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewSymbolExporter(
	observationCtx *observation.Context,
	rankingService RankingService,
	numRoutines int,
	interval time.Duration,
	batchSize int,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"ranking.symbol-exporter", "exports SCIP data to ranking definitions and reference tables",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			if err := rankingService.ExportRankingGraph(ctx, numRoutines, batchSize); err != nil {
				return err
			}

			if err := rankingService.VacuumRankingGraph(ctx); err != nil {
				return err
			}

			return nil
		}),
	)
}

func NewMapper(
	observationCtx *observation.Context,
	rankingService RankingService,
	interval time.Duration,
) goroutine.BackgroundRoutine {
	operations := newMapperOperations(observationCtx)

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"ranking.file-reference-count-mapper", "maps definitions and references data to path_counts_inputs table in store",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			numReferenceRecordsProcessed, numInputsInserted, err := rankingService.MapRankingGraph(ctx)
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
	rankingService RankingService,
	interval time.Duration,
) goroutine.BackgroundRoutine {
	operations := newReducerOperations(observationCtx)

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"ranking.file-reference-count-reducer", "reduces path_counts_inputs into a count of paths per repository and stores it in path_ranks table in store.",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			numPathRanksInserted, numPathCountsInputsProcessed, err := rankingService.ReduceRankingGraph(ctx)
			if err != nil {
				return err
			}

			operations.numPathCountsInputsRowsProcessed.Add(numPathCountsInputsProcessed)
			operations.numPathRanksInserted.Add(numPathRanksInserted)
			return nil
		}),
	)
}
