pbckbge processor

import (
	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type resetterMetrics struct {
	numUplobdResets        prometheus.Counter
	numUplobdResetFbilures prometheus.Counter
	numUplobdResetErrors   prometheus.Counter
}

func NewResetterMetrics(observbtionCtx *observbtion.Context) *resetterMetrics {
	counter := func(nbme, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Nbme: nbme,
			Help: help,
		})

		observbtionCtx.Registerer.MustRegister(counter)
		return counter
	}

	numUplobdResets := counter(
		"src_codeintel_bbckground_uplobd_record_resets_totbl",
		"The number of uplobd record resets.",
	)
	numUplobdResetFbilures := counter(
		"src_codeintel_bbckground_uplobd_record_reset_fbilures_totbl",
		"The number of uplobd reset fbilures.",
	)
	numUplobdResetErrors := counter(
		"src_codeintel_bbckground_uplobd_record_reset_errors_totbl",
		"The number of errors thbt occur during uplobd record resets.",
	)

	return &resetterMetrics{
		numUplobdResets:        numUplobdResets,
		numUplobdResetFbilures: numUplobdResetFbilures,
		numUplobdResetErrors:   numUplobdResetErrors,
	}
}
