package janitor

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type metrics struct {
	numUploadRecordsRemoved prometheus.Counter
	numIndexRecordsRemoved  prometheus.Counter
	numUploadsPurged        prometheus.Counter
	numUploadResets         prometheus.Counter
	numUploadResetFailures  prometheus.Counter
	numIndexResets          prometheus.Counter
	numIndexResetFailures   prometheus.Counter
	numErrors               prometheus.Counter
}

var NewMetrics = newMetrics

func newMetrics(observationContext *observation.Context) *metrics {
	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})

		observationContext.Registerer.MustRegister(counter)
		return counter
	}

	numUploadRecordsRemoved := counter(
		"src_codeintel_background_upload_records_removed_total",
		"The number of codeintel upload records removed.",
	)
	numIndexRecordsRemoved := counter(
		"src_codeintel_background_index_records_removed_total",
		"The number of codeintel index records removed.",
	)
	numUploadsPurged := counter(
		"src_codeintel_background_uploads_purged_total",
		"The number of uploads for which records in the codeintel db were removed.",
	)
	numUploadResets := counter(
		"src_codeintel_background_upload_resets_total",
		"The number of upload record resets.",
	)
	numUploadResetFailures := counter(
		"src_codeintel_background_upload_reset_failures_total",
		"The number of upload reset failures.",
	)
	numIndexResets := counter(
		"src_codeintel_background_index_resets_total",
		"The number of index records reset.",
	)
	numIndexResetFailures := counter(
		"src_codeintel_background_index_reset_failures_total",
		"The number of index reset failures.",
	)
	numErrors := counter(
		"src_codeintel_background_errors_total",
		"The number of errors that occur during a codeintel background job.",
	)

	return &metrics{
		numUploadRecordsRemoved: numUploadRecordsRemoved,
		numIndexRecordsRemoved:  numIndexRecordsRemoved,
		numUploadsPurged:        numUploadsPurged,
		numUploadResets:         numUploadResets,
		numUploadResetFailures:  numUploadResetFailures,
		numIndexResets:          numIndexResets,
		numIndexResetFailures:   numIndexResetFailures,
		numErrors:               numErrors,
	}
}
