package indexscheduler

import (
	"github.com/prometheus/client_golang/prometheus"
)

type SchedulerMetrics struct {
	Errors prometheus.Counter
}

func NewSchedulerMetrics(r prometheus.Registerer) SchedulerMetrics {
	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_index_scheduler_errors_total",
		Help: "Total number of errors when running the index scheduler",
	})
	r.MustRegister(errors)

	return SchedulerMetrics{
		Errors: errors,
	}
}
