package janitor

import "github.com/prometheus/client_golang/prometheus"

type JanitorMetrics struct {
	FailedUploads prometheus.Counter
	DeadDumps     prometheus.Counter
	OldDumps      prometheus.Counter
	Errors        prometheus.Counter
}

func (jm JanitorMetrics) MustRegister(r prometheus.Registerer) {
	r.MustRegister(jm.FailedUploads)
	r.MustRegister(jm.DeadDumps)
	r.MustRegister(jm.OldDumps)
	r.MustRegister(jm.Errors)
}

func NewJanitorMetrics() JanitorMetrics {
	return JanitorMetrics{
		FailedUploads: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "src",
			Subsystem: "precise-code-intel-bundle-manager",
			Name:      "janitor_failed_uploads",
			Help:      "Total number of failed upload removed",
		}),
		DeadDumps: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "src",
			Subsystem: "precise-code-intel-bundle-manager",
			Name:      "janitor_dead_dumps",
			Help:      "Total number of dead dumps removed",
		}),
		OldDumps: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "src",
			Subsystem: "precise-code-intel-bundle-manager",
			Name:      "janitor_old_dumps",
			Help:      "Total number of old dumps removed",
		}),
		Errors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "src",
			Subsystem: "precise-code-intel-bundle-manager",
			Name:      "janitor_errors",
			Help:      "Total number of errors when running the janitor",
		}),
	}
}
