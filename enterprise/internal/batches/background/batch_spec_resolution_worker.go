package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// newBatchSpecResolutionWorker creates a dbworker.newWorker that fetches BatchSpecResolutionJobs
// specs and passes them to the batchSpecWorkspaceCreator.
func newBatchSpecResolutionWorker(
	ctx context.Context,
	s *store.Store,
	workerStore dbworkerstore.Store,
	metrics batchChangesMetrics,
) *workerutil.Worker {
	e := &batchSpecWorkspaceCreator{store: s}

	options := workerutil.WorkerOptions{
		Name:              "batch_changes_batch_spec_resolution_worker",
		NumHandlers:       5,
		Interval:          1 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           metrics.batchSpecResolutionWorkerMetrics,
	}

	worker := dbworker.NewWorker(ctx, workerStore, e.HandlerFunc(), options)
	return worker
}

func newBatchSpecResolutionWorkerResetter(workerStore dbworkerstore.Store, metrics batchChangesMetrics) *dbworker.Resetter {
	options := dbworker.ResetterOptions{
		Name:     "batch_changes_batch_spec_resolution_worker_resetter",
		Interval: 1 * time.Minute,
		Metrics:  metrics.batchSpecResolutionWorkerResetterMetrics,
	}

	resetter := dbworker.NewResetter(workerStore, options)
	return resetter
}
