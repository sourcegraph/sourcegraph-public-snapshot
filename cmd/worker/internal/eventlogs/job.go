package eventlogs

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/bg"
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
			NewEventLogsJob(observationCtx.Logger, db),
			NewSecurityEventLogsJob(observationCtx.Logger, db),
		},
		nil
}

func NewEventLogsJob(logger log.Logger, db database.DB) goroutine.BackgroundRoutine {
	handler := goroutine.HandlerFunc(func(ctx context.Context) error {
		return bg.DeleteOldEventLogsInPostgres(ctx, logger, db)
	})

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		handler,
		goroutine.WithName("delete_old_event_logs"),
		goroutine.WithDescription("deleting expired rows from event_logs table"),
		goroutine.WithInterval(time.Hour),
	)
}

func NewSecurityEventLogsJob(logger log.Logger, db database.DB) goroutine.BackgroundRoutine {
	handler := goroutine.HandlerFunc(func(ctx context.Context) error {
		return bg.DeleteOldSecurityEventLogsInPostgres(ctx, logger, db)
	})

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		handler,
		goroutine.WithName("delete_old_security_event_logs"),
		goroutine.WithDescription("deleting expired rows from security_event_logs table"),
		goroutine.WithInterval(time.Hour),
		goroutine.WithInitialDelay(time.Hour),
	)
}
