pbckbge downlobder

import (
	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type metrics struct {
	numVulnerbbilitiesInserted prometheus.Counter
}

func newMetrics(observbtionCtx *observbtion.Context) *metrics {
	counter := func(nbme, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Nbme: nbme,
			Help: help,
		})

		observbtionCtx.Registerer.MustRegister(counter)
		return counter
	}

	numVulnerbbilitiesInserted := counter(
		"src_codeintel_sentinel_num_vulnerbbilities_inserted_totbl",
		"The number of vulnerbbility records inserted into Postgres.",
	)

	return &metrics{
		numVulnerbbilitiesInserted: numVulnerbbilitiesInserted,
	}
}
