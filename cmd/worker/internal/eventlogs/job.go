package eventlogs

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/metrics"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type eventLogsJob struct{}

func NewEventLogsJanitorJob() job.Job {
	return &eventLogsJob{}
}

func (e eventLogsJob) Description() string {
	return "deletes old event logs from postgres"
}

func (e eventLogsJob) Config() []env.Config {
	return nil
}

func (e eventLogsJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
			newEventLogsJob(observationCtx, db),
			newSecurityEventLogsJob(observationCtx, db),
		},
		nil
}

func newEventLogsJob(observationCtx *observation.Context, db database.DB) goroutine.BackgroundRoutine {
	handler := goroutine.HandlerFunc(func(ctx context.Context) error {
		return deleteOldEventLogsInPostgres(ctx, db)
	})

	operation := observationCtx.Operation(observation.Op{
		Name: "event_logs.janitor.run",
		Metrics: metrics.NewREDMetrics(
			observationCtx.Registerer,
			"event_logs_janitor",
			metrics.WithCountHelp("Total number of event_logs janitor executions"),
		),
	})

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		handler,
		goroutine.WithName("delete_old_event_logs"),
		goroutine.WithDescription("deleting expired rows from event_logs table"),
		goroutine.WithInterval(time.Hour),
		goroutine.WithOperation(operation),
	)
}

func newSecurityEventLogsJob(observationCtx *observation.Context, db database.DB) goroutine.BackgroundRoutine {
	handler := goroutine.HandlerFunc(func(ctx context.Context) error {
		return deleteOldSecurityEventLogsInPostgres(ctx, db)
	})

	operation := observationCtx.Operation(observation.Op{
		Name: "security_event_logs.janitor.run",
		Metrics: metrics.NewREDMetrics(
			observationCtx.Registerer,
			"security_event_logs_janitor",
			metrics.WithCountHelp("Total number of security_event_logs janitor executions"),
		),
	})

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		handler,
		goroutine.WithName("delete_old_security_event_logs"),
		goroutine.WithDescription("deleting expired rows from security_event_logs table"),
		goroutine.WithInterval(time.Hour),
		goroutine.WithInitialDelay(time.Hour),
		goroutine.WithOperation(operation),
	)
}
