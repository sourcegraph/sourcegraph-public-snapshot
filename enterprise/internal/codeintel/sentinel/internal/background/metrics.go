package background

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Metrics struct {
	numVulnerabilitiesInserted prometheus.Counter
}

func NewMetrics(observationCtx *observation.Context) *Metrics {
	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})

		observationCtx.Registerer.MustRegister(counter)
		return counter
	}

	numVulnerabilitiesInserted := counter(
		"src_codeintel_sentinel_num_vulnerabilities_inserted_total",
		"The number of vulnerability records inserted into Postgres.",
	)

	return &Metrics{
		numVulnerabilitiesInserted: numVulnerabilitiesInserted,
	}
}
