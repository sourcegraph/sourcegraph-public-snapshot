package resetter

import "github.com/prometheus/client_golang/prometheus"

type ResetterMetrics struct {
	StalledJobs prometheus.Counter
	Errors      prometheus.Counter
}

func NewResetterMetrics(r prometheus.Registerer) ResetterMetrics {
	stalledJobs := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "precise_code_intel_api_server",
		Name:      "resetter_stalled_jobs",
		Help:      "Total number of reset stalled jobs",
	})

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "precise_code_intel_api_server",
		Name:      "resetter_errors",
		Help:      "Total number of errors when running the janitor",
	})

	if r != nil {
		r.MustRegister(stalledJobs)
		r.MustRegister(errors)
	}

	return ResetterMetrics{
		StalledJobs: stalledJobs,
		Errors:      errors,
	}
}
