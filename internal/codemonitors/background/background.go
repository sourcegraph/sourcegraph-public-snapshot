package background

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewBackgroundJobs(observationCtx *observation.Context, db database.DB) []goroutine.BackgroundRoutine {
	observationCtx = observation.ContextWithLogger(observationCtx.Logger.Scoped("BackgroundJobs"), observationCtx)

	codeMonitorsStore := db.CodeMonitors()

	triggerMetrics := newMetricsForTriggerQueries(observationCtx)
	actionMetrics := newActionMetrics(observationCtx)

	// Create a new context. Each background routine will wrap this with
	// a cancellable context that is canceled when Stop() is called.
	ctx := context.Background()
	return []goroutine.BackgroundRoutine{
		newTriggerQueryEnqueuer(ctx, codeMonitorsStore),
		newTriggerJobsLogDeleter(ctx, codeMonitorsStore),
		newTriggerQueryRunner(ctx, scopedContext("TriggerQueryRunner", observationCtx), db, triggerMetrics),
		newTriggerQueryResetter(ctx, scopedContext("TriggerQueryResetter", observationCtx), codeMonitorsStore, triggerMetrics),
		newActionRunner(ctx, scopedContext("ActionRunner", observationCtx), codeMonitorsStore, actionMetrics),
		newActionJobResetter(ctx, scopedContext("ActionJobResetter", observationCtx), codeMonitorsStore, actionMetrics),
	}
}

func scopedContext(operation string, parent *observation.Context) *observation.Context {
	return observation.ContextWithLogger(parent.Logger.Scoped(operation), parent)
}
