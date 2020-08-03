package janitor

import "github.com/prometheus/client_golang/prometheus"

type JanitorMetrics struct {
	IndexRecordsRemoved prometheus.Counter
	Errors              prometheus.Counter
}

func NewJanitorMetrics(r prometheus.Registerer) JanitorMetrics {
	indexRecordsRemoved := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_indexer_janitor_index_records_removed_total",
		Help: "Total number of index records removed",
	})
	r.MustRegister(indexRecordsRemoved)

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_indexer_janitor_errors_total",
		Help: "Total number of errors when running the janitor",
	})
	r.MustRegister(errors)

	return JanitorMetrics{
		IndexRecordsRemoved: indexRecordsRemoved,
		Errors:              errors,
	}
}
