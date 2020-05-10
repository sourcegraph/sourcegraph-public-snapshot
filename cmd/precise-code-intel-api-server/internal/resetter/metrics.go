package resetter

import "github.com/prometheus/client_golang/prometheus"

type ResetterMetrics struct {
	Count  prometheus.Counter
	Errors prometheus.Counter
}

func NewResetterMetrics(r prometheus.Registerer) ResetterMetrics {
	count := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "upload_queue",
		Name:      "resets_total",
		Help:      "Total number of uploads put back into queued state",
	})
	r.MustRegister(count)

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "upload_queue",
		Name:      "reset_errors_total",
		Help:      "Total number of errors when running the upload resetter",
	})
	r.MustRegister(errors)

	return ResetterMetrics{
		Count:  count,
		Errors: errors,
	}
}
