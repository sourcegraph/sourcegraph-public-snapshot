package background

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	UploadRecordsRemoved prometheus.Counter
	IndexRecordsRemoved  prometheus.Counter
	UploadDataRemoved    prometheus.Counter
	UploadResets         prometheus.Counter
	UploadResetFailures  prometheus.Counter
	IndexResets          prometheus.Counter
	IndexResetFailures   prometheus.Counter
	Errors               prometheus.Counter
}

func NewMetrics(r prometheus.Registerer) Metrics {
	uploadRecordsRemoved := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_codeintel_background_upload_records_removed_total",
		Help: "The number of codeintel upload records removed.",
	})
	r.MustRegister(uploadRecordsRemoved)

	indexRecordsRemoved := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_codeintel_background_index_records_removed_total",
		Help: "The number of codeintel index records removed.",
	})
	r.MustRegister(indexRecordsRemoved)

	uploadDataRemoved := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_codeintel_background_upload_data_removed_total",
		Help: "The number of indexes for which records in the codeintel db were removed.",
	})
	r.MustRegister(uploadDataRemoved)

	uploadResets := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_codeintel_background_upload_resets_total",
		Help: "The number of upload record resets.",
	})
	r.MustRegister(uploadResets)

	uploadResetFailures := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_codeintel_background_upload_reset_failures_total",
		Help: "The number of upload reset failures.",
	})
	r.MustRegister(uploadResetFailures)

	indexResets := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_codeintel_background_index_resets_total",
		Help: "The number of index records reset.",
	})
	r.MustRegister(indexResets)

	indexResetFailures := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_codeintel_background_index_reset_failures_total",
		Help: "The number of index reset failures.",
	})
	r.MustRegister(indexResetFailures)

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_codeintel_background_errors_total",
		Help: "The number of errors that occur during a codeintel background job.",
	})
	r.MustRegister(errors)

	return Metrics{
		UploadRecordsRemoved: uploadRecordsRemoved,
		IndexRecordsRemoved:  indexRecordsRemoved,
		UploadDataRemoved:    uploadDataRemoved,
		UploadResets:         uploadResets,
		UploadResetFailures:  uploadResetFailures,
		IndexResets:          indexResets,
		IndexResetFailures:   indexResetFailures,
		Errors:               errors,
	}
}
