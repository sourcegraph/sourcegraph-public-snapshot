package janitor

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func NewReconcilerWorkerResetter(workerStore dbworkerstore.Store, metrics *metrics) *dbworker.Resetter {
	options := dbworker.ResetterOptions{
		Name:     "batches_reconciler_worker_resetter",
		Interval: 1 * time.Minute,
		Metrics:  metrics.reconcilerWorkerResetterMetrics,
	}

	resetter := dbworker.NewResetter(workerStore, options)
	return resetter
}

// NewBulkOperationWorkerResetter creates a dbworker.Resetter that reenqueues lost jobs
// for processing.
func NewBulkOperationWorkerResetter(workerStore dbworkerstore.Store, metrics *metrics) *dbworker.Resetter {
	options := dbworker.ResetterOptions{
		Name:     "batches_bulk_worker_resetter",
		Interval: 1 * time.Minute,
		Metrics:  metrics.bulkProcessorWorkerResetterMetrics,
	}

	resetter := dbworker.NewResetter(workerStore, options)
	return resetter
}

// NewBatchSpecWorkspaceExecutionWorkerResetter creates a dbworker.Resetter that re-enqueues
// lost batch_spec_workspace_execution_jobs for processing.
func NewBatchSpecWorkspaceExecutionWorkerResetter(workerStore dbworkerstore.Store, metrics *metrics) *dbworker.Resetter {
	options := dbworker.ResetterOptions{
		Name:     "batch_spec_workspace_execution_worker_resetter",
		Interval: 1 * time.Minute,
		Metrics:  metrics.batchSpecWorkspaceExecutionWorkerResetterMetrics,
	}

	resetter := dbworker.NewResetter(workerStore, options)
	return resetter
}

func NewBatchSpecWorkspaceResolutionWorkerResetter(workerStore dbworkerstore.Store, metrics *metrics) *dbworker.Resetter {
	options := dbworker.ResetterOptions{
		Name:     "batch_changes_batch_spec_resolution_worker_resetter",
		Interval: 1 * time.Minute,
		Metrics:  metrics.batchSpecResolutionWorkerResetterMetrics,
	}

	resetter := dbworker.NewResetter(workerStore, options)
	return resetter
}
