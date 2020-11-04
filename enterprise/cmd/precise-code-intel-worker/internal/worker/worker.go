package worker

import (
	"context"
	"database/sql"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/metrics"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/postgres"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	uploadstore "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/upload_store"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

func NewWorker(
	s store.Store,
	codeIntelDB *sql.DB,
	uploadStore uploadstore.Store,
	gitserverClient gitserverClient,
	pollInterval time.Duration,
	numProcessorRoutines int,
	budgetMax int64,
	metrics metrics.WorkerMetrics,
	observationContext *observation.Context,
) *workerutil.Worker {
	rootContext := actor.WithActor(context.Background(), &actor.Actor{Internal: true})

	handler := &handler{
		store:           s,
		uploadStore:     uploadStore,
		gitserverClient: gitserverClient,
		metrics:         metrics,
		enableBudget:    budgetMax > 0,
		budgetRemaining: budgetMax,
		createStore: func(id int) persistence.Store {
			return persistence.NewObserved(postgres.NewStore(codeIntelDB, id), observationContext)
		},
	}

	return dbworker.NewWorker(rootContext, store.WorkerutilUploadStore(s), handler, workerutil.WorkerOptions{
		NumHandlers: numProcessorRoutines,
		Interval:    pollInterval,
		Metrics: workerutil.WorkerMetrics{
			HandleOperation: metrics.ProcessOperation,
		},
	})
}
