package background

import (
	"context"

	"github.com/sourcegraph/log"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewBackgroundJobs(logger log.Logger, db edb.EnterpriseDB) []goroutine.BackgroundRoutine {
	logger = logger.Scoped("BackgroundJobs", "code monitors background jobs")

	codeMonitorsStore := db.CodeMonitors()

	triggerMetrics := newMetricsForTriggerQueries(logger)
	actionMetrics := newActionMetrics(logger)

	// Create a new context. Each background routine will wrap this with
	// a cancellable context that is canceled when Stop() is called.
	ctx := context.Background()
	return []goroutine.BackgroundRoutine{
		newTriggerQueryEnqueuer(ctx, codeMonitorsStore),
		newTriggerJobsLogDeleter(ctx, codeMonitorsStore),
		newTriggerQueryRunner(ctx, logger.Scoped("TriggerQueryRunner", ""), db, triggerMetrics),
		newTriggerQueryResetter(ctx, logger.Scoped("TriggerQueryResetter", ""), codeMonitorsStore, triggerMetrics),
		newActionRunner(ctx, logger.Scoped("ActionRunner", ""), codeMonitorsStore, actionMetrics),
		newActionJobResetter(ctx, logger.Scoped("ActionJobResetter", ""), codeMonitorsStore, actionMetrics),
	}
}
