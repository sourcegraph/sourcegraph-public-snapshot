package search

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type searchJob struct{}

func NewSearchJob() job.Job {
	return &searchJob{}
}

func (j *searchJob) Description() string {
	return ""
}

func (j *searchJob) Config() []env.Config {
	return nil
}

func (j *searchJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	observationCtx = observation.NewContext(observationCtx.Logger.Scoped("routines", "exhaustive search job routines"))
	workCtx := actor.WithInternalActor(context.Background())

	workerStore, err := InitExhaustiveSearchWorkerStore()
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
		NewExhaustiveSearchWorker(workCtx, observationCtx, workerStore),
	}, nil
}
