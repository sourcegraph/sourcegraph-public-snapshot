package background

import (
	"context"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func StartBackgroundJobs(ctx context.Context, db edb.EnterpriseDB) {
	routines := NewBackgroundJobs(ctx, db)
	go goroutine.MonitorBackgroundRoutines(ctx, routines...)
}

func NewBackgroundJobs(ctx context.Context, db edb.EnterpriseDB) []goroutine.BackgroundRoutine {
	codeMonitorsStore := db.CodeMonitors()

	triggerMetrics := newMetricsForTriggerQueries()
	actionMetrics := newActionMetrics()

	return []goroutine.BackgroundRoutine{
		newTriggerQueryEnqueuer(ctx, codeMonitorsStore),
		newTriggerJobsLogDeleter(ctx, codeMonitorsStore),
		newTriggerQueryRunner(ctx, db, triggerMetrics),
		newTriggerQueryResetter(ctx, codeMonitorsStore, triggerMetrics),
		newActionRunner(ctx, codeMonitorsStore, actionMetrics),
		newActionJobResetter(ctx, codeMonitorsStore, actionMetrics),
	}
}
