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
// "completed" state that are over a week old and removes them.
//
// This is particularly important because the historical enqueuer can produce e.g.
// num_series*num_repos*num_timeframes jobs (example: 20*40,000*6 in an average case) which
// can quickly add up to be millions of jobs left in a "completed" state in the DB.
func NewCleaner(ctx context.Context, observationCtx *observation.Context, workerBaseStore *basestore.Store) goroutine.BackgroundRoutine {
	redMetrics := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"insights_query_runner_cleaner",
		metrics.WithCountHelp("Total number of insights queryrunner cleaner executions"),
	)
	operation := observationCtx.Operation(observation.Op{
		Name:    "QueryRunner.Cleaner.Run",
		Metrics: redMetrics,
	})

	// We look for jobs to clean up every hour.
	return goroutine.NewPeriodicGoroutine(
		ctx,
		goroutine.HandlerFunc(
			func(ctx context.Context) error {
				// TODO(slimsag): future: recording the number of jobs cleaned up in a metric would be nice.
				_, err := cleanJobs(ctx, workerBaseStore)
				return err
			},
		),
		goroutine.WithName("insights.query_runner_cleaner"),
		goroutine.WithDescription("removes completed or failed query runner jobs"),
		goroutine.WithInterval(1*time.Hour),
		goroutine.WithOperation(operation),
	)
}

func cleanJobs(ctx context.Context, workerBaseStore *basestore.Store) (numCleaned int, err error) {
	numCleaned, _, err = basestore.ScanFirstInt(workerBaseStore.Query(
		ctx,
		sqlf.Sprintf(cleanJobsFmtStr, time.Now().Add(-168*time.Hour)),
	))
	return
}

const cleanJobsFmtStr = `
WITH deleted AS (
	DELETE FROM insights_query_runner_jobs WHERE state='completed' AND started_at <= %s RETURNING *
) SELECT count(*) FROM deleted
`
