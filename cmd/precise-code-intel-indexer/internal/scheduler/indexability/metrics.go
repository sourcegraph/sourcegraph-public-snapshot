package indexabilityscheduler

import (
	"github.com/prometheus/client_golang/prometheus"
)

type SchedulerMetrics struct {
	Errors prometheus.Counter
}

func NewSchedulerMetrics(r prometheus.Registerer) SchedulerMetrics {
	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_indexability_scheduler_errors_total",
		Help: "Total number of errors when running the indexability scheduler",
	})
	r.MustRegister(errors)

	return SchedulerMetrics{
		Errors: errors,
	}
}
