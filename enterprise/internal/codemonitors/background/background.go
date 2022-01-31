package background

import (
	"context"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func StartBackgroundJobs(ctx context.Context, db edb.EnterpriseDB) {
	codeMonitorsStore := db.CodeMonitors()

	triggerMetrics := newMetricsForTriggerQueries()
	actionMetrics := newActionMetrics()

	routines := []goroutine.BackgroundRoutine{
		newTriggerQueryEnqueuer(ctx, codeMonitorsStore),
		newTriggerJobsLogDeleter(ctx, codeMonitorsStore),
		newTriggerQueryRunner(ctx, db, triggerMetrics),
		newTriggerQueryResetter(ctx, codeMonitorsStore, triggerMetrics),
		newActionRunner(ctx, codeMonitorsStore, actionMetrics),
		newActionJobResetter(ctx, codeMonitorsStore, actionMetrics),
	}
	go goroutine.MonitorBackgroundRoutines(ctx, routines...)
}
