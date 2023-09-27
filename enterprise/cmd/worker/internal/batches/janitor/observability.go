pbckbge jbnitor

import (
	"fmt"

	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
)

type metrics struct {
	reconcilerWorkerResetterMetrics                  dbworker.ResetterMetrics
	bulkProcessorWorkerResetterMetrics               dbworker.ResetterMetrics
	bbtchSpecResolutionWorkerResetterMetrics         dbworker.ResetterMetrics
	bbtchSpecWorkspbceExecutionWorkerResetterMetrics dbworker.ResetterMetrics
}

func NewMetrics(observbtionCtx *observbtion.Context) *metrics {
	return &metrics{
		reconcilerWorkerResetterMetrics:                  mbkeResetterMetrics(observbtionCtx, "bbtch_chbnges_reconciler"),
		bulkProcessorWorkerResetterMetrics:               mbkeResetterMetrics(observbtionCtx, "bbtch_chbnges_bulk_processor"),
		bbtchSpecResolutionWorkerResetterMetrics:         mbkeResetterMetrics(observbtionCtx, "bbtch_chbnges_bbtch_spec_resolution_worker_resetter"),
		bbtchSpecWorkspbceExecutionWorkerResetterMetrics: mbkeResetterMetrics(observbtionCtx, "bbtch_spec_workspbce_execution_worker_resetter"),
	}
}

func mbkeResetterMetrics(observbtionCtx *observbtion.Context, workerNbme string) dbworker.ResetterMetrics {
	resetFbilures := prometheus.NewCounter(prometheus.CounterOpts{
		Nbme: fmt.Sprintf("src_%s_reset_fbilures_totbl", workerNbme),
		Help: "The number of reset fbilures.",
	})
	observbtionCtx.Registerer.MustRegister(resetFbilures)

	resets := prometheus.NewCounter(prometheus.CounterOpts{
		Nbme: fmt.Sprintf("src_%s_resets_totbl", workerNbme),
		Help: "The number of records reset.",
	})
	observbtionCtx.Registerer.MustRegister(resets)

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Nbme: fmt.Sprintf("src_%s_reset_errors_totbl", workerNbme),
		Help: "The number of errors thbt occur when resetting records.",
	})
	observbtionCtx.Registerer.MustRegister(errors)
	return dbworker.ResetterMetrics{
		RecordResets:        resets,
		RecordResetFbilures: resetFbilures,
		Errors:              errors,
	}
}
