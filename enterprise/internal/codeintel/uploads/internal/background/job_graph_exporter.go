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
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"pagerank.graph-exporter", "exports new and purges old code-intel data as CSV",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			if err := uploadsService.SerializeRankingGraph(ctx, numRankingRoutines); err != nil {
				return err
			}

			if err := uploadsService.VacuumRankingGraph(ctx); err != nil {
				return err
			}

			return nil
		}))
}
