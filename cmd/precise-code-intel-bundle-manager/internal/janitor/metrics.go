package janitor

import "github.com/prometheus/client_golang/prometheus"

type JanitorMetrics struct {
	OldUploads    prometheus.Counter
	OrphanedDumps prometheus.Counter
	EvictedDumps  prometheus.Counter
	Errors        prometheus.Counter
}

func (jm JanitorMetrics) MustRegister(r prometheus.Registerer) {
	r.MustRegister(jm.OldUploads)
	r.MustRegister(jm.OrphanedDumps)
	r.MustRegister(jm.EvictedDumps)
	r.MustRegister(jm.Errors)
}

func NewJanitorMetrics() JanitorMetrics {
	return JanitorMetrics{
		OldUploads: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "src",
			Subsystem: "precise-code-intel-bundle-manager",
			Name:      "janitor_old_uploads",
			Help:      "Total number of old upload removed",
		}),
		OrphanedDumps: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "src",
			Subsystem: "precise-code-intel-bundle-manager",
			Name:      "janitor_orphaned_dumps",
			Help:      "Total number of orphaned dumps removed",
		}),
		EvictedDumps: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "src",
			Subsystem: "precise-code-intel-bundle-manager",
			Name:      "janitor_old_dumps",
			Help:      "Total number of dumps evicted from disk",
		}),
		Errors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "src",
			Subsystem: "precise-code-intel-bundle-manager",
			Name:      "janitor_errors",
			Help:      "Total number of errors when running the janitor",
		}),
	}
}
