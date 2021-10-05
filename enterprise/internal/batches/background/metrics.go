package background

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

type batchChangesMetrics struct {
	reconcilerWorkerMetrics            workerutil.WorkerMetrics
	bulkProcessorWorkerMetrics         workerutil.WorkerMetrics
	reconcilerWorkerResetterMetrics    dbworker.ResetterMetrics
	bulkProcessorWorkerResetterMetrics dbworker.ResetterMetrics

	batchSpecResolutionWorkerMetrics         workerutil.WorkerMetrics
	batchSpecResolutionWorkerResetterMetrics dbworker.ResetterMetrics

	batchSpecWorkspaceExecutionWorkerResetterMetrics dbworker.ResetterMetrics
}

func newMetrics(observationContext *observation.Context) batchChangesMetrics {
	return batchChangesMetrics{
		reconcilerWorkerMetrics:            workerutil.NewMetrics(observationContext, "batch_changes_reconciler", nil),
		bulkProcessorWorkerMetrics:         workerutil.NewMetrics(observationContext, "batch_changes_bulk_processor", nil),
		reconcilerWorkerResetterMetrics:    makeResetterMetrics(observationContext, "batch_changes_reconciler"),
		bulkProcessorWorkerResetterMetrics: makeResetterMetrics(observationContext, "batch_changes_bulk_processor"),

		batchSpecResolutionWorkerMetrics:         workerutil.NewMetrics(observationContext, "batch_changes_batch_spec_resolution_worker", nil),
		batchSpecResolutionWorkerResetterMetrics: makeResetterMetrics(observationContext, "batch_changes_batch_spec_resolution_worker_resetter"),

		batchSpecWorkspaceExecutionWorkerResetterMetrics: makeResetterMetrics(observationContext, "batch_spec_workspace_execution_worker_resetter"),
	}
}

func makeResetterMetrics(observationContext *observation.Context, workerName string) dbworker.ResetterMetrics {
	resetFailures := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("src_%s_reset_failures_total", workerName),
		Help: "The number of reset failures.",
	})
	observationContext.Registerer.MustRegister(resetFailures)

	resets := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("src_%s_resets_total", workerName),
		Help: "The number of records reset.",
	})
	observationContext.Registerer.MustRegister(resets)

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("src_%s_reset_errors_total", workerName),
		Help: "The number of errors that occur when resetting records.",
	})
	observationContext.Registerer.MustRegister(errors)
	return dbworker.ResetterMetrics{
		RecordResets:        resets,
		RecordResetFailures: resetFailures,
		Errors:              errors,
	}
}
