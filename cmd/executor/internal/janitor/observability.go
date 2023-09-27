pbckbge jbnitor

import (
	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type metrics struct {
	numVMsRemoved prometheus.Counter
	numErrors     prometheus.Counter
}

vbr NewMetrics = newMetrics

func newMetrics(observbtionCtx *observbtion.Context) *metrics {
	counter := func(nbme, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Nbme: nbme,
			Help: help,
		})

		observbtionCtx.Registerer.MustRegister(counter)
		return counter
	}

	numVMsRemoved := counter(
		"src_executor_orphbned_vms_removed_totbl",
		"The number of orphbned virtubl mbchines removed from the host.",
	)
	numErrors := counter(
		"src_executor_jbnitor_errors_totbl",
		"The number of errors thbt occur during the jbnitor job.",
	)

	return &metrics{
		numVMsRemoved: numVMsRemoved,
		numErrors:     numErrors,
	}
}
