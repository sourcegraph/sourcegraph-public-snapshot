package processor

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type resetterMetrics struct {
	numUploadResets        prometheus.Counter
	numUploadResetFailures prometheus.Counter
	numUploadResetErrors   prometheus.Counter
}

func NewResetterMetrics(observationCtx *observation.Context) *resetterMetrics {
	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})

		observationCtx.Registerer.MustRegister(counter)
		return counter
	}

	numUploadResets := counter(
		"src_codeintel_background_upload_record_resets_total",
		"The number of upload record resets.",
	)
	numUploadResetFailures := counter(
		"src_codeintel_background_upload_record_reset_failures_total",
		"The number of upload reset failures.",
	)
	numUploadResetErrors := counter(
		"src_codeintel_background_upload_record_reset_errors_total",
		"The number of errors that occur during upload record resets.",
	)

	return &resetterMetrics{
		numUploadResets:        numUploadResets,
		numUploadResetFailures: numUploadResetFailures,
		numUploadResetErrors:   numUploadResetErrors,
	}
}
