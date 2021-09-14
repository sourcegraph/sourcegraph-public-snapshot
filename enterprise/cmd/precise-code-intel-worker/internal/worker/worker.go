package worker

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// UploadHeartbeatInterval is the duration between heartbeat updates to the upload job records.
const UploadHeartbeatInterval = time.Second

func NewWorker(
	dbStore DBStore,
	workerStore dbworkerstore.Store,
	lsifStore LSIFStore,
	repoStore *database.RepoStore,
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
		workerStore:     workerStore,
		lsifStore:       lsifStore,
		repoStore:       repoStore,
		uploadStore:     uploadStore,
		gitserverClient: gitserverClient,
		enableBudget:    budgetMax > 0,
		budgetRemaining: budgetMax,
	}

	return dbworker.NewWorker(rootContext, workerStore, handler, workerutil.WorkerOptions{
		Name:              "precise_code_intel_upload_worker",
		NumHandlers:       numProcessorRoutines,
		Interval:          pollInterval,
		HeartbeatInterval: UploadHeartbeatInterval,
		Metrics:           workerMetrics,
	})
}
