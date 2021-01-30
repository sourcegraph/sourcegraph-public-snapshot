package background

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func StartBackgroundJobs(ctx context.Context, db *sql.DB) {
	resolver := insights.InitResolver(ctx, db)

	triggerMetrics := newMetricsForTriggerQueries()
	actionMetrics := newActionMetrics()

	routines := []goroutine.BackgroundRoutine{
		/*
			newTriggerQueryEnqueuer(ctx, resolver),
			newTriggerJobsLogDeleter(ctx, resolver),
			newTriggerQueryRunner(ctx, resolver, triggerMetrics),
			newTriggerQueryResetter(ctx, resolver, triggerMetrics),
			newActionRunner(ctx, resolver, actionMetrics),
			newActionJobResetter(ctx, resolver, actionMetrics),
		*/
	}
	go goroutine.MonitorBackgroundRoutines(ctx, routines...)
}
