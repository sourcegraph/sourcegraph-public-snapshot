package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewRankingGraphExporter(
	uploadsService UploadService,
	numRankingRoutines int,
	interval time.Duration,
	observationContext *observation.Context,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		if err := uploadsService.SerializeRankingGraph(ctx, numRankingRoutines); err != nil {
			return err
		}

		if err := uploadsService.VacuumRankingGraph(ctx); err != nil {
			return err
		}

		return nil
	}))
}
