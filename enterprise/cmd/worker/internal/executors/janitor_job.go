package executors

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type janitorJob struct {
	observationContext *observation.Context
}

func NewJanitorJob(observationContext *observation.Context) job.Job {
	return &janitorJob{observation.ContextWithLogger(log.NoOp(), observationContext)}
}

func (j *janitorJob) Description() string {
	return ""
}

func (j *janitorJob) Config() []env.Config {
	return []env.Config{janitorConfigInst}
}

func (j *janitorJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDBWithLogger(observation.ContextWithLogger(logger, j.observationContext))
	if err != nil {
		return nil, err
	}

	routines := []goroutine.BackgroundRoutine{
		goroutine.NewPeriodicGoroutine(context.Background(), janitorConfigInst.CleanupTaskInterval, goroutine.HandlerFunc(func(ctx context.Context) error {
			return db.Executors().DeleteInactiveHeartbeats(ctx, janitorConfigInst.HeartbeatRecordsMaxAge)
		})),
	}

	return routines, nil
}
