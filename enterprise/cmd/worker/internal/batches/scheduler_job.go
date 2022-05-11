package batches

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/scheduler"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

type schedulerJob struct{}

func NewSchedulerJob() job.Job {
	return &schedulerJob{}
}

func (j *schedulerJob) Description() string {
	return ""
}

func (j *schedulerJob) Config() []env.Config {
	return []env.Config{}
}

func (j *schedulerJob) Routines(_ context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	workCtx := actor.WithInternalActor(context.Background())

	bstore, err := InitStore()
	if err != nil {
		return nil, err
	}

	routines := []goroutine.BackgroundRoutine{
		scheduler.NewScheduler(workCtx, bstore),
	}

	return routines, nil
}
