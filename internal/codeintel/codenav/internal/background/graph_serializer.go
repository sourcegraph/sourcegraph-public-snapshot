package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func (b *backgroundJob) NewRankingGraphSerializer(
	interval time.Duration,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutineWithMetrics(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return b.handleRankingGraphSerializer(ctx)
	}), b.operations.handleRankingGraphSerializer)
}

func (b backgroundJob) handleRankingGraphSerializer(
	ctx context.Context,
) error {
	return b.codeNavSvc.SerializeRankingGraph(ctx)
}
