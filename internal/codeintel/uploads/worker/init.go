package worker

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/worker/cleanup"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/worker/commitgraph"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/worker/expiration"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type WorkerJob struct{}

func NewWorkerJob() job.Job {
	return &WorkerJob{}
}

func (j *WorkerJob) Config() []env.Config {
	return []env.Config{
		cleanup.ConfigInst,
		commitgraph.ConfigInst,
		expiration.ConfigInst,
	}
}

func (j *WorkerJob) Routines(ctx context.Context) ([]goroutine.BackgroundRoutine, error) {
	return []goroutine.BackgroundRoutine{
		cleanup.NewJanitor(),
		commitgraph.NewUpdater(),
		expiration.NewExpirer(),
	}, nil
}
