package janitor

import "github.com/prometheus/client_golang/prometheus"

type JanitorMetrics struct {
	OldUploads    prometheus.Counter
	OrphanedDumps prometheus.Counter
	EvictedDumps  prometheus.Counter
	Errors        prometheus.Counter
}

func NewJanitorMetrics(r prometheus.Registerer) JanitorMetrics {
	oldUploads := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "precise_code_intel_bundle_manager",
		Name:      "janitor_old_uploads",
		Help:      "Total number of old upload removed",
	})

	orphanedDumps := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "precise_code_intel_bundle_manager",
		Name:      "janitor_orphaned_dumps",
		Help:      "Total number of orphaned dumps removed",
	})

	evictedDumps := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "precise_code_intel_bundle_manager",
		Name:      "janitor_old_dumps",
		Help:      "Total number of dumps evicted from disk",
	})

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "precise_code_intel_bundle_manager",
		Name:      "janitor_errors",
		Help:      "Total number of errors when running the janitor",
	})

	if r != nil {
		r.MustRegister(oldUploads)
		r.MustRegister(orphanedDumps)
		r.MustRegister(evictedDumps)
		r.MustRegister(errors)
	}

	return JanitorMetrics{
		OldUploads:    oldUploads,
		OrphanedDumps: orphanedDumps,
		EvictedDumps:  evictedDumps,
		Errors:        errors,
	}
}
