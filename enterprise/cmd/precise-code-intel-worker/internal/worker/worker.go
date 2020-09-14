package worker

import (
	"context"
	"database/sql"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/metrics"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

func NewWorker(
	s store.Store,
	codeIntelDB *sql.DB,
	bundleManagerClient bundles.BundleManagerClient,
	gitserverClient gitserver.Client,
	pollInterval time.Duration,
	numProcessorRoutines int,
	budgetMax int64,
	metrics metrics.WorkerMetrics,
) *workerutil.Worker {
	rootContext := actor.WithActor(context.Background(), &actor.Actor{Internal: true})

	handler := &handler{
		store:               s,
		codeIntelDB:         codeIntelDB,
		bundleManagerClient: bundleManagerClient,
		gitserverClient:     gitserverClient,
		metrics:             metrics,
		enableBudget:        budgetMax > 0,
		budgetRemaining:     budgetMax,
	}

	return dbworker.NewWorker(rootContext, store.WorkerutilUploadStore(s), dbworker.WorkerOptions{
		Handler:     handler,
		NumHandlers: numProcessorRoutines,
		Interval:    pollInterval,
		Metrics: workerutil.WorkerMetrics{
			HandleOperation: metrics.ProcessOperation,
		},
	})
}
