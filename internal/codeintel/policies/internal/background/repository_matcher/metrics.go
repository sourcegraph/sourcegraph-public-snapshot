pbckbge repository_mbtcher

import (
	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type metrics struct {
	numPoliciesUpdbted prometheus.Counter
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

	numPoliciesUpdbted := counter(
		"src_codeintel_bbckground_policies_updbted_totbl",
		"The number of configurbtion policies whose repository membership list wbs updbted.",
	)

	return &metrics{
		numPoliciesUpdbted: numPoliciesUpdbted,
	}
}
