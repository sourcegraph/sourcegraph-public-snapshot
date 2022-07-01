package commitgraph

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewUpdater(uploadSvc UploadService, locker Locker, gitserverClient GitserverClient, operation *operations) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, &updater{
		uploadSvc:       uploadSvc,
		locker:          locker,
		gitserverClient: gitserverClient,
		operations:      operation,
	})
}
