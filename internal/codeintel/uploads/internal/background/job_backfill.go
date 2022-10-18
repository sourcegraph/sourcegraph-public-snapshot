package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func (b backgroundJob) NewCommittedAtBackfiller(interval time.Duration, batchSize int) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return b.uploadSvc.BackfillCommittedAtBatch(ctx, batchSize)
	}))
}
