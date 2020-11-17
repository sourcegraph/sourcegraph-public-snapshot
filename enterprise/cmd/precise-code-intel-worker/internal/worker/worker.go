package worker

import (
	"context"
	"time"

	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

func NewWorker(
	dbStore DBStore,
	lsifStore LSIFStore,
	uploadStore uploadstore.Store,
	gitserverClient GitserverClient,
	pollInterval time.Duration,
	numProcessorRoutines int,
	budgetMax int64,
	workerMetrics workerutil.WorkerMetrics,
) *workerutil.Worker {
	rootContext := actor.WithActor(context.Background(), &actor.Actor{Internal: true})

	handler := &handler{
		dbStore:         dbStore,
		lsifStore:       lsifStore,
		uploadStore:     uploadStore,
		gitserverClient: gitserverClient,
		enableBudget:    budgetMax > 0,
		budgetRemaining: budgetMax,
	}

	return dbworker.NewWorker(rootContext, store.WorkerutilUploadStore(dbStore), handler, workerutil.WorkerOptions{
		NumHandlers: numProcessorRoutines,
		Interval:    pollInterval,
		Metrics:     workerMetrics,
	})
}
