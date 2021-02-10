package background

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

// metrics describes all Prometheus metrics to be recorded during the background execution of
// workers.
type metrics struct {
	// workerMetrics records worker operations & number of jobs.
	workerMetrics workerutil.WorkerMetrics

	// resetterMetrics records the number of jobs that got reset because workers timed out / took
	// too long.
	resetterMetrics dbworker.ResetterMetrics
}

func newMetrics(observationContext *observation.Context) *metrics {
	workerResets := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_insights_worker_resets_total",
		Help: "The number of times work took too long and was reset for retry later.",
	})
	observationContext.Registerer.MustRegister(workerResets)

	workerResetFailures := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_insights_worker_reset_failures_total",
		Help: "The number of times work took too long so many times that retries will no longer happen.",
	})
	observationContext.Registerer.MustRegister(workerResetFailures)

	workerErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_insights_worker_errors_total",
		Help: "The number of errors that occurred during a worker job.",
	})

	return &metrics{
		workerMetrics: workerutil.NewMetrics(observationContext, "insights", nil),
		resetterMetrics: dbworker.ResetterMetrics{
			RecordResets:        workerResets,
			RecordResetFailures: workerResetFailures,
			Errors:              workerErrors,
		},
	}
}
