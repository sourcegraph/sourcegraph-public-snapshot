package resetter

import "github.com/prometheus/client_golang/prometheus"

type ResetterMetrics struct {
	StalledJobs prometheus.Counter
	Errors      prometheus.Counter
}

func (rm ResetterMetrics) MustRegister(r prometheus.Registerer) {
	r.MustRegister(rm.StalledJobs)
	r.MustRegister(rm.Errors)
}

func NewResetterMetrics() ResetterMetrics {
	return ResetterMetrics{
		StalledJobs: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "src",
			Subsystem: "precise-code-intel-api-server",
			Name:      "resetter_stalled_jobs",
			Help:      "Total number of reset stalled jobs",
		}),
		Errors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "src",
			Subsystem: "precise-code-intel-api-server",
			Name:      "resetter_errors",
			Help:      "Total number of errors when running the janitor",
		}),
	}
}
