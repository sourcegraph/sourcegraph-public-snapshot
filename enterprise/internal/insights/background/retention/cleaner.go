package retention

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// NewCleaner returns a routine that deletes completed retention records older than a week.
// We enqueue a new retention row every time a series is handled in the queryrunner, so we do not want records to pile
// up too much.
func NewCleaner(ctx context.Context, observationCtx *observation.Context, workerBaseStore *basestore.Store) goroutine.BackgroundRoutine {
	operation := observationCtx.Operation(observation.Op{
		Name: "DataRetention.Cleaner.Run",
		Metrics: metrics.NewREDMetrics(
			observationCtx.Registerer,
			"insights_data_retention_job_cleaner",
			metrics.WithCountHelp("Total number of insights data retention cleaner executions"),
		),
	})

	// We look for jobs to clean up every hour.
	return goroutine.NewPeriodicGoroutine(
		ctx,
		goroutine.HandlerFunc(
			func(ctx context.Context) error {
				return cleanJobs(ctx, workerBaseStore)
			},
		),
		goroutine.WithName("insights.data_retention_job_cleaner"),
		goroutine.WithDescription("removes completed data retention jobs"),
		goroutine.WithInterval(1*time.Hour),
		goroutine.WithOperation(operation),
	)
}

func cleanJobs(ctx context.Context, workerBaseStore *basestore.Store) error {
	return workerBaseStore.Exec(
		ctx,
		sqlf.Sprintf(cleanJobsFmtStr, time.Now().Add(-168*time.Hour)),
	)
}

const cleanJobsFmtStr = `
DELETE FROM insights_data_retention_jobs WHERE state='completed' AND started_at <= %s
`
