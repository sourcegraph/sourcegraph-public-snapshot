package queryrunner

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// NewCleaner returns a background goroutine which will periodically find jobs left in the
// "completed" or "failed" state that are over 12 hours old and removes them.
//
// This is particularly important because the historical enqueuer can produce e.g.
// num_series*num_repos*num_timeframes jobs (example: 20*40,000*6 in an average case) which
// can quickly add up to be millions of jobs left in a "completed" state in the DB.
func NewCleaner(ctx context.Context, workerBaseStore *basestore.Store, observationContext *observation.Context) goroutine.BackgroundRoutine {
	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"insights_query_runner_cleaner",
		metrics.WithCountHelp("Total number of insights queryrunner cleaner executions"),
	)
	operation := observationContext.Operation(observation.Op{
		Name:    "QueryRunner.Cleaner.Run",
		Metrics: metrics,
	})

	// We look for jobs to cleanup every hour.
	return goroutine.NewPeriodicGoroutineWithMetrics(ctx, 1*time.Hour, goroutine.NewHandlerWithErrorMessage(
		"insights_query_runner_cleaner",
		func(ctx context.Context) error {
			// TODO(slimsag): future: recording the number of jobs cleaned up in a metric would be nice.
			_, err := cleanJobs(ctx, workerBaseStore)
			return err
		},
	), operation)
}

// cleanJobs
func cleanJobs(ctx context.Context, workerBaseStore *basestore.Store) (numCleaned int, err error) {
	numCleaned, _, err = basestore.ScanFirstInt(workerBaseStore.Query(
		ctx,
		sqlf.Sprintf(cleanJobsFmtStr, time.Now().Add(-12*time.Hour)),
	))
	return
}

const cleanJobsFmtStr = `
-- source: enterprise/internal/insights/background/queryrunner/cleaner.go:cleanJobs
WITH deleted AS (
	DELETE FROM insights_query_runner_jobs WHERE (state='completed' OR state='failed') AND started_at >= %s RETURNING *
) SELECT count(*) FROM deleted
`
