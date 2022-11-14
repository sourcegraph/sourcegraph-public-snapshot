package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewCommittedAtBackfiller(uploadSvc UploadService, interval time.Duration, batchSize int) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return uploadSvc.BackfillCommittedAtBatch(ctx, batchSize)
	}))
}
