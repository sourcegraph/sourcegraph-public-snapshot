package janitor

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type metrics struct {
	// Data retention metrics
	numRepositoriesScanned          prometheus.Counter
	numUploadsScanned               prometheus.Counter
	numCommitsScanned               prometheus.Counter
	numUploadsExpired               prometheus.Counter
	numUploadRecordsRemoved         prometheus.Counter
	numIndexRecordsRemoved          prometheus.Counter
	numUploadsPurged                prometheus.Counter
	numDocumentSearchRecordsRemoved prometheus.Counter
	numErrors                       prometheus.Counter

	// Resetter metrics
	numUploadResets                 prometheus.Counter
	numUploadResetFailures          prometheus.Counter
	numUploadResetErrors            prometheus.Counter
	numIndexResets                  prometheus.Counter
	numIndexResetFailures           prometheus.Counter
	numIndexResetErrors             prometheus.Counter
	numDependencyIndexResets        prometheus.Counter
	numDependencyIndexResetFailures prometheus.Counter
	numDependencyIndexResetErrors   prometheus.Counter
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

	numRepositoriesScanned := counter(
		"src_codeintel_background_repositories_scanned_total",
		"The number of repositories scanned for data retention.",
	)
	numUploadsScanned := counter(
		"src_codeintel_background_upload_records_scanned_total",
		"The number of codeintel upload records scanned for data retention.",
	)
	numCommitsScanned := counter(
		"src_codeintel_background_commits_scanned_total",
		"The number of commits reachable from a codeintel upload record scanned for data retention.",
	)
	numUploadsExpired := counter(
		"src_codeintel_background_upload_records_expired_total",
		"The number of codeintel upload records marked as expired.",
	)
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
		"The number of uploads for which records in the codeintel database were removed.",
	)
	numDocumentSearchRecordsRemoved := counter(
		"src_codeintel_background_documentation_search_records_removed_total",
		"The number of documentation search records removed.",
	)
	numErrors := counter(
		"src_codeintel_background_errors_total",
		"The number of errors that occur during a codeintel expiration job.",
	)

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

	numIndexResets := counter(
		"src_codeintel_background_index_record_resets_total",
		"The number of index records reset.",
	)
	numIndexResetFailures := counter(
		"src_codeintel_background_index_record_reset_failures_total",
		"The number of dependency index reset failures.",
	)
	numIndexResetErrors := counter(
		"src_codeintel_background_index_record_reset_errors_total",
		"The number of errors that occur during index records reset.",
	)

	numDependencyIndexResets := counter(
		"src_codeintel_background_dependency_index_record_resets_total",
		"The number of dependency index records reset.",
	)
	numDependencyIndexResetFailures := counter(
		"src_codeintel_background_dependency_index_record_reset_failures_total",
		"The number of index reset failures.",
	)
	numDependencyIndexResetErrors := counter(
		"src_codeintel_background_dependency_index_record_reset_errors_total",
		"The number of errors that occur during dependency index records reset.",
	)

	return &metrics{
		numRepositoriesScanned:          numRepositoriesScanned,
		numUploadsScanned:               numUploadsScanned,
		numCommitsScanned:               numCommitsScanned,
		numUploadsExpired:               numUploadsExpired,
		numUploadRecordsRemoved:         numUploadRecordsRemoved,
		numIndexRecordsRemoved:          numIndexRecordsRemoved,
		numUploadsPurged:                numUploadsPurged,
		numDocumentSearchRecordsRemoved: numDocumentSearchRecordsRemoved,
		numErrors:                       numErrors,
		numUploadResets:                 numUploadResets,
		numUploadResetFailures:          numUploadResetFailures,
		numUploadResetErrors:            numUploadResetErrors,
		numIndexResets:                  numIndexResets,
		numIndexResetFailures:           numIndexResetFailures,
		numIndexResetErrors:             numIndexResetErrors,
		numDependencyIndexResets:        numDependencyIndexResets,
		numDependencyIndexResetFailures: numDependencyIndexResetFailures,
		numDependencyIndexResetErrors:   numDependencyIndexResetErrors,
	}
}
