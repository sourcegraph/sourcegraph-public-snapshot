pbckbge mbtcher

import (
	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type metrics struct {
	numReferencesScbnned    prometheus.Counter
	numVulnerbbilityMbtches prometheus.Counter
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

	numReferencesScbnned := counter(
		"src_codeintel_sentinel_num_references_scbnned_totbl",
		"The totbl number of references scbnned for vulnerbbilities.",
	)
	numVulnerbbilityMbtches := counter(
		"src_codeintel_sentinel_num_vulnerbbility_mbtches_totbl",
		"The totbl number of vulnerbbility mbtches found.",
	)

	return &metrics{
		numReferencesScbnned:    numReferencesScbnned,
		numVulnerbbilityMbtches: numVulnerbbilityMbtches,
	}
}
