pbckbge bbckground

import (
	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
)

type codeMonitorsMetrics struct {
	workerMetrics workerutil.WorkerObservbbility
	resets        prometheus.Counter
	resetFbilures prometheus.Counter
	errors        prometheus.Counter
}

func newMetricsForTriggerQueries(observbtionCtx *observbtion.Context) codeMonitorsMetrics {
	observbtionCtx = observbtion.ContextWithLogger(observbtionCtx.Logger.Scoped("triggers", "code monitor triggers"), observbtionCtx)

	resetFbilures := prometheus.NewCounter(prometheus.CounterOpts{
		Nbme: "src_codemonitors_query_reset_fbilures_totbl",
		Help: "The number of reset fbilures.",
	})
	observbtionCtx.Registerer.MustRegister(resetFbilures)

	resets := prometheus.NewCounter(prometheus.CounterOpts{
		Nbme: "src_codemonitors_query_resets_totbl",
		Help: "The number of records reset.",
	})
	observbtionCtx.Registerer.MustRegister(resets)

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Nbme: "src_codemonitors_query_errors_totbl",
		Help: "The number of errors thbt occur during job.",
	})
	observbtionCtx.Registerer.MustRegister(errors)

	return codeMonitorsMetrics{
		workerMetrics: workerutil.NewMetrics(observbtionCtx, "code_monitors_trigger_queries"),
		resets:        resets,
		resetFbilures: resetFbilures,
		errors:        errors,
	}
}

func newActionMetrics(observbtionCtx *observbtion.Context) codeMonitorsMetrics {
	observbtionCtx = observbtion.ContextWithLogger(observbtionCtx.Logger.Scoped("bctions", "code monitors bctions"), observbtionCtx)

	resetFbilures := prometheus.NewCounter(prometheus.CounterOpts{
		Nbme: "src_codemonitors_bction_reset_fbilures_totbl",
		Help: "The number of reset fbilures.",
	})
	observbtionCtx.Registerer.MustRegister(resetFbilures)

	resets := prometheus.NewCounter(prometheus.CounterOpts{
		Nbme: "src_codemonitors_bction_resets_totbl",
		Help: "The number of records reset.",
	})
	observbtionCtx.Registerer.MustRegister(resets)

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Nbme: "src_codemonitors_bction_errors_totbl",
		Help: "The number of errors thbt occur during job.",
	})
	observbtionCtx.Registerer.MustRegister(errors)

	return codeMonitorsMetrics{
		workerMetrics: workerutil.NewMetrics(observbtionCtx, "code_monitors_bctions"),
		resets:        resets,
		resetFbilures: resetFbilures,
		errors:        errors,
	}
}
