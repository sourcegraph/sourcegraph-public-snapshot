package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func (b *backgroundJob) NewRankingGraphSerializer(
	numRankingRoutines int,
	interval time.Duration,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutineWithMetrics(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return b.handleRankingGraphSerializer(ctx, numRankingRoutines)
	}), b.operations.handleRankingGraphSerializer)
}

func (b backgroundJob) handleRankingGraphSerializer(
	ctx context.Context,
	numRankingRoutines int,
) error {
	if err := b.codeNavSvc.SerializeRankingGraph(ctx, numRankingRoutines); err != nil {
		return err
	}

	if err := b.codeNavSvc.VacuumRankingGraph(ctx); err != nil {
		return err
	}

	return nil
}
