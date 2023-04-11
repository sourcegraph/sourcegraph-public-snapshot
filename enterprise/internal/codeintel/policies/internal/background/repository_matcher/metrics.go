package repository_matcher

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type metrics struct {
	numPoliciesUpdated prometheus.Counter
}

func newMetrics(observationCtx *observation.Context) *metrics {
	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})

		observationCtx.Registerer.MustRegister(counter)
		return counter
	}

	numPoliciesUpdated := counter(
		"src_codeintel_background_policies_updated_total",
		"The number of configuration policies whose repository membership list was updated.",
	)

	return &metrics{
		numPoliciesUpdated: numPoliciesUpdated,
	}
}
