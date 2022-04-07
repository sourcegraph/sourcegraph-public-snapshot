package worker

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/worker/scheduler"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type WorkerJob struct{}

func NewWorkerJob() job.Job {
	return &WorkerJob{}
}

func (j *WorkerJob) Config() []env.Config {
	return []env.Config{
		scheduler.ConfigInst,
	}
}

func (j *WorkerJob) Routines(ctx context.Context) ([]goroutine.BackgroundRoutine, error) {
	return []goroutine.BackgroundRoutine{
		scheduler.NewScheduler(),
	}, nil
}
