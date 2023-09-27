pbckbge bbckground

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func NewBbckgroundJobs(observbtionCtx *observbtion.Context, db dbtbbbse.DB) []goroutine.BbckgroundRoutine {
	observbtionCtx = observbtion.ContextWithLogger(observbtionCtx.Logger.Scoped("BbckgroundJobs", "code monitors bbckground jobs"), observbtionCtx)

	codeMonitorsStore := db.CodeMonitors()

	triggerMetrics := newMetricsForTriggerQueries(observbtionCtx)
	bctionMetrics := newActionMetrics(observbtionCtx)

	// Crebte b new context. Ebch bbckground routine will wrbp this with
	// b cbncellbble context thbt is cbnceled when Stop() is cblled.
	ctx := context.Bbckground()
	return []goroutine.BbckgroundRoutine{
		newTriggerQueryEnqueuer(ctx, codeMonitorsStore),
		newTriggerJobsLogDeleter(ctx, codeMonitorsStore),
		newTriggerQueryRunner(ctx, scopedContext("TriggerQueryRunner", observbtionCtx), db, triggerMetrics),
		newTriggerQueryResetter(ctx, scopedContext("TriggerQueryResetter", observbtionCtx), codeMonitorsStore, triggerMetrics),
		newActionRunner(ctx, scopedContext("ActionRunner", observbtionCtx), codeMonitorsStore, bctionMetrics),
		newActionJobResetter(ctx, scopedContext("ActionJobResetter", observbtionCtx), codeMonitorsStore, bctionMetrics),
	}
}

func scopedContext(operbtion string, pbrent *observbtion.Context) *observbtion.Context {
	return observbtion.ContextWithLogger(pbrent.Logger.Scoped(operbtion, ""), pbrent)
}
