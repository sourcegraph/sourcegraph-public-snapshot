package background

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// newBatchSpecWorkspaceExecutionWorkerResetter creates a dbworker.Resetter that re-enqueues
// lost batch_spec_workspace_execution_jobs for processing.
func newBatchSpecWorkspaceExecutionWorkerResetter(workerStore dbworkerstore.Store, metrics batchChangesMetrics) *dbworker.Resetter {
	options := dbworker.ResetterOptions{
		Name:     "batch_spec_workspace_execution_worker_resetter",
		Interval: 1 * time.Minute,
		Metrics:  metrics.batchSpecWorkspaceExecutionWorkerResetterMetrics,
	}

	resetter := dbworker.NewResetter(workerStore, options)
	return resetter
}
