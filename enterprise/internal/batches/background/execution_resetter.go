package background

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// newBatchSpecExecutionResetter creates a dbworker.Resetter that re-enqueues
// lost batch_spec_execution jobs for processing.
func newBatchSpecExecutionResetter(workerStore dbworkerstore.Store, metrics batchChangesMetrics) *dbworker.Resetter {
	options := dbworker.ResetterOptions{
		Name:     "batch_spec_executor_resetter",
		Interval: 1 * time.Minute,
		Metrics:  metrics.executionResetterMetrics,
	}

	resetter := dbworker.NewResetter(workerStore, options)
	return resetter
}
