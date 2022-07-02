package commitgraph

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewUpdater(uploadSvc UploadService, operation *operations) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, &updater{
		uploadSvc:  uploadSvc,
		operations: operation,
	})
}
