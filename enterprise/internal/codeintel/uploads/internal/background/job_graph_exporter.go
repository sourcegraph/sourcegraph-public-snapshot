package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewRankingGraphExporter(
	observationCtx *observation.Context,
	uploadsService UploadService,
	numRankingRoutines int,
	interval time.Duration,
	batchSize int,
	rankingJobEnabled bool,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"rank.graph-exporter", "exports SCIP data to ranking defintions and reference tables",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			if err := uploadsService.ExportRankingGraph(ctx, numRankingRoutines, batchSize, rankingJobEnabled); err != nil {
				return err
			}

			// Need to replace this pre-deployment
			// if err := uploadsService.VacuumRankingGraph(ctx); err != nil {
			// 	return err
			// }

			return nil
		}))
}

func NewRankingGraphMapper(
	observationCtx *observation.Context,
	uploadsService UploadService,
	numRankingRoutines int,
	interval time.Duration,
	rankingJobEnabled bool,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"rank.graph-mapper", "maps definitions and references data to path_counts_inputs table in store",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			if err := uploadsService.MapRankingGraph(ctx, numRankingRoutines, rankingJobEnabled); err != nil {
				return err
			}
			return nil
		}))
}
