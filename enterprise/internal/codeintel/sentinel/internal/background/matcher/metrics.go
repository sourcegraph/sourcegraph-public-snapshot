package matcher

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type metrics struct {
	numReferencesScanned    prometheus.Counter
	numVulnerabilityMatches prometheus.Counter
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

	numReferencesScanned := counter(
		"src_codeintel_sentinel_num_references_scanned_total",
		"The total number of references scanned for vulnerabilities.",
	)
	numVulnerabilityMatches := counter(
		"src_codeintel_sentinel_num_vulnerability_matches_total",
		"The total number of vulnerability matches found.",
	)

	return &metrics{
		numReferencesScanned:    numReferencesScanned,
		numVulnerabilityMatches: numVulnerabilityMatches,
	}
}
