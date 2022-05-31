package commitgraph

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewUpdater(dbStore DBStore, locker Locker, gitserverClient GitserverClient, operation *operations) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, &updater{
		dbStore:         dbStore,
		locker:          locker,
		gitserverClient: gitserverClient,
		operations:      operation,
	})
}
