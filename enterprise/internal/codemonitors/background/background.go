package background

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func StartBackgroundJobs(ctx context.Context, db *sql.DB) {
	codeMonitorsStore := codemonitors.NewStore(db)

	metrics := newMetricsForTriggerQueries()

	routines := []goroutine.BackgroundRoutine{
		newTriggerQueryEnqueuer(ctx, codeMonitorsStore),
		newTriggerJobsLogDeleter(ctx, codeMonitorsStore),
		newTriggerQueryRunner(ctx, codeMonitorsStore, metrics),
		newTriggerQueryResetter(ctx, codeMonitorsStore, metrics),
	}
	go goroutine.MonitorBackgroundRoutines(ctx, routines...)
}
