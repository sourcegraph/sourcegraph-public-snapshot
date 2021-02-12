package background

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func StartBackgroundJobs(ctx context.Context, db *sql.DB) {
	codeMonitorsStore := codemonitors.NewStore(db)

	triggerMetrics := newMetricsForTriggerQueries()
	actionMetrics := newActionMetrics()

	routines := []goroutine.BackgroundRoutine{
		newTriggerQueryEnqueuer(ctx, codeMonitorsStore),
		newTriggerJobsLogDeleter(ctx, codeMonitorsStore),
		newTriggerQueryRunner(ctx, codeMonitorsStore, triggerMetrics),
		newTriggerQueryResetter(ctx, codeMonitorsStore, triggerMetrics),
		newActionRunner(ctx, codeMonitorsStore, actionMetrics),
		newActionJobResetter(ctx, codeMonitorsStore, actionMetrics),
	}
	go goroutine.MonitorBackgroundRoutines(ctx, routines...)
}
