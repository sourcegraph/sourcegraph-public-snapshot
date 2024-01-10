package janitor

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

type metrics struct {
	reconcilerWorkerResetterMetrics                  dbworker.ResetterMetrics
	bulkProcessorWorkerResetterMetrics               dbworker.ResetterMetrics
	batchSpecResolutionWorkerResetterMetrics         dbworker.ResetterMetrics
	batchSpecWorkspaceExecutionWorkerResetterMetrics dbworker.ResetterMetrics
}

func NewMetrics(observationCtx *observation.Context) *metrics {
	return &metrics{
		reconcilerWorkerResetterMetrics:                  makeResetterMetrics(observationCtx, "batch_changes_reconciler"),
		bulkProcessorWorkerResetterMetrics:               makeResetterMetrics(observationCtx, "batch_changes_bulk_processor"),
		batchSpecResolutionWorkerResetterMetrics:         makeResetterMetrics(observationCtx, "batch_changes_batch_spec_resolution_worker_resetter"),
		batchSpecWorkspaceExecutionWorkerResetterMetrics: makeResetterMetrics(observationCtx, "batch_spec_workspace_execution_worker_resetter"),
	}
}

func makeResetterMetrics(observationCtx *observation.Context, workerName string) dbworker.ResetterMetrics {
	resetFailures := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("src_%s_reset_failures_total", workerName),
		Help: "The number of reset failures.",
	})
	observationCtx.Registerer.MustRegister(resetFailures)

	resets := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("src_%s_resets_total", workerName),
		Help: "The number of records reset.",
	})
	observationCtx.Registerer.MustRegister(resets)

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("src_%s_reset_errors_total", workerName),
		Help: "The number of errors that occur when resetting records.",
	})
	observationCtx.Registerer.MustRegister(errors)
	return dbworker.ResetterMetrics{
		RecordResets:        resets,
		RecordResetFailures: resetFailures,
		Errors:              errors,
	}
}
