package webhookworker

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewCleaner(ctx context.Context, workerBaseStore *basestore.Store, observationContext *observation.Context) goroutine.BackgroundRoutine {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"webhook_build_worker_cleaner",
		metrics.WithCountHelp("Total number of webhookbuilder cleaner executions"),
	)
	operation := observationContext.Operation(observation.Op{
		Name:    "WebhookBuilder.Cleaner.Run",
		Metrics: metrics,
	})

	return goroutine.NewPeriodicGoroutineWithMetrics(ctx, 1*time.Hour, goroutine.NewHandlerWithErrorMessage(
		"webhook_build_worker_cleaner",
		func(ctx context.Context) error {
			_, err := cleanJobs(ctx, workerBaseStore)
			return err
		},
	), operation)
}

func cleanJobs(ctx context.Context, workerBaseStore *basestore.Store) (numCleaned int, err error) {
	numCleaned, _, err = basestore.ScanFirstInt(workerBaseStore.Query(
		ctx,
		sqlf.Sprintf(cleanJobsFmtStr, time.Now().Add(-168*time.Hour)), // 7 days
	))
	return
}

const cleanJobsFmtStr = `
WITH deleted AS (
	DELETE FROM webhook_build_jobs WHERE state='completed' OR state='failed' AND started_at <= %s RETURNING *
) SELECT count(*) from deleted
`
