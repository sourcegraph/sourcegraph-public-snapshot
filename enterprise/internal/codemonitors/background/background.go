package background

import (
	"context"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewBackgroundJobs(db edb.EnterpriseDB, observationContext *observation.Context) []goroutine.BackgroundRoutine {
	observationContext = observation.ContextWithLogger(observationContext.Logger.Scoped("BackgroundJobs", "code monitors background jobs"), observationContext)

	codeMonitorsStore := db.CodeMonitors()

	triggerMetrics := newMetricsForTriggerQueries(observationContext)
	actionMetrics := newActionMetrics(observationContext)

	// Create a new context. Each background routine will wrap this with
	// a cancellable context that is canceled when Stop() is called.
	ctx := context.Background()
	return []goroutine.BackgroundRoutine{
		newTriggerQueryEnqueuer(ctx, codeMonitorsStore),
		newTriggerJobsLogDeleter(ctx, codeMonitorsStore),
		newTriggerQueryRunner(ctx, db, triggerMetrics, scopedContext("TriggerQueryRunner", observationContext)),
		newTriggerQueryResetter(ctx, codeMonitorsStore, triggerMetrics, scopedContext("TriggerQueryResetter", observationContext)),
		newActionRunner(ctx, codeMonitorsStore, actionMetrics, scopedContext("ActionRunner", observationContext)),
		newActionJobResetter(ctx, codeMonitorsStore, actionMetrics, scopedContext("ActionJobResetter", observationContext)),
	}
}

func scopedContext(operation string, parent *observation.Context) *observation.Context {
	return observation.ContextWithLogger(parent.Logger.Scoped(operation, ""), parent)
}
