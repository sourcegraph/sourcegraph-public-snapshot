package commitgraph

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewUpdater(dbStore DBStore, uploadSvc UploadService, locker Locker, gitserverClient GitserverClient, operation *operations) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, &updater{
		dbStore:         dbStore,
		uploadSvc:       uploadSvc,
		locker:          locker,
		gitserverClient: gitserverClient,
		operations:      operation,
	})
}
