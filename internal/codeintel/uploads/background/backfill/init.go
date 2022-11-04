package backfill

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewCommittedAtBackfiller(uploadSvc UploadService) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, &committedAtBackfiller{
		uploadSvc: uploadSvc,
		batchSize: ConfigInst.BatchSize,
	})
}
