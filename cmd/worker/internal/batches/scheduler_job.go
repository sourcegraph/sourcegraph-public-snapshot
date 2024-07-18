package batches

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/batches"
	"github.com/sourcegraph/sourcegraph/internal/batches/scheduler"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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

func (j *schedulerJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	if !batches.IsEnabled() {
		return nil, nil
	}
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
