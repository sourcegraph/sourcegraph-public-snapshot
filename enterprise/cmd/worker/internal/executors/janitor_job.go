package executors

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/workerdb"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type janitorJob struct{}

func NewJanitorJob() job.Job {
	return &janitorJob{}
}

func (j *janitorJob) Config() []env.Config {
	return []env.Config{janitorConfigInst}
}

func (j *janitorJob) Routines(ctx context.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	routines := []goroutine.BackgroundRoutine{
		goroutine.NewPeriodicGoroutine(context.Background(), janitorConfigInst.CleanupTaskInterval, goroutine.HandlerFunc(func(ctx context.Context) error {
			return database.NewDB(db).Executors().DeleteInactiveHeartbeats(ctx, janitorConfigInst.HeartbeatRecordsMaxAge)
		})),
	}

	return routines, nil
}
