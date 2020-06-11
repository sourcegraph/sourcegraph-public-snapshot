package indexabilityupdater

import (
	"github.com/prometheus/client_golang/prometheus"
)

type UpdaterMetrics struct {
	Errors prometheus.Counter
}

func NewUpdaterMetrics(r prometheus.Registerer) UpdaterMetrics {
	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_indexability_updater_errors_total",
		Help: "Total number of errors when running the indexability updater",
	})
	r.MustRegister(errors)

	return UpdaterMetrics{
		Errors: errors,
	}
}
