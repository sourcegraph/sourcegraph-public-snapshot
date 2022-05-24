package executors

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/services/executors"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

type janitorJob struct{}

func NewJanitorJob() job.Job {
	return &janitorJob{}
}

func (j *janitorJob) Description() string {
	return ""
}

func (j *janitorJob) Config() []env.Config {
	return []env.Config{janitorConfigInst}
}

func (j *janitorJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	routines := []goroutine.BackgroundRoutine{
		goroutine.NewPeriodicGoroutine(context.Background(), janitorConfigInst.CleanupTaskInterval, goroutine.HandlerFunc(func(ctx context.Context) error {
			return executors.New(db).DeleteInactiveHeartbeats(ctx, janitorConfigInst.HeartbeatRecordsMaxAge)
		})),
	}

	return routines, nil
}
