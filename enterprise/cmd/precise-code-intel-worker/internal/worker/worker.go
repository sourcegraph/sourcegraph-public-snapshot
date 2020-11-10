package worker

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/metrics"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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
	metrics metrics.WorkerMetrics,
	observationContext *observation.Context,
) *workerutil.Worker {
	rootContext := actor.WithActor(context.Background(), &actor.Actor{Internal: true})

	handler := &handler{
		dbStore:         dbStore,
		lsifStore:       lsifStore,
		uploadStore:     uploadStore,
		gitserverClient: gitserverClient,
		metrics:         metrics,
		enableBudget:    budgetMax > 0,
		budgetRemaining: budgetMax,
	}

	return dbworker.NewWorker(rootContext, store.WorkerutilUploadStore(dbStore), handler, workerutil.WorkerOptions{
		NumHandlers: numProcessorRoutines,
		Interval:    pollInterval,
		Metrics: workerutil.WorkerMetrics{
			HandleOperation: metrics.ProcessOperation,
		},
	})
}
