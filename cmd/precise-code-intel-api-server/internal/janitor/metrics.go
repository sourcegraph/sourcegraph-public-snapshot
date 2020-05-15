package janitor

import "github.com/prometheus/client_golang/prometheus"

type JanitorMetrics struct {
	UploadRecordsRemoved prometheus.Counter
	Errors               prometheus.Counter
}

func NewJanitorMetrics(r prometheus.Registerer) JanitorMetrics {
	uploadRecordsRemoved := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_api_server_janitor_upload_records_removed_total",
		Help: "Total number of processed upload records removed (with no corresponding bundle file)",
	})
	r.MustRegister(uploadRecordsRemoved)

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_api_server_janitor_errors_total",
		Help: "Total number of errors when running the janitor",
	})
	r.MustRegister(errors)

	return JanitorMetrics{
		UploadRecordsRemoved: uploadRecordsRemoved,
		Errors:               errors,
	}
}
