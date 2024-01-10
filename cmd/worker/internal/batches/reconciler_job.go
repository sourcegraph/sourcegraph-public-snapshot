package batches

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/internal/batches/workers"
	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type reconcilerJob struct{}

func NewReconcilerJob() job.Job {
	return &reconcilerJob{}
}

func (j *reconcilerJob) Description() string {
	return ""
}

func (j *reconcilerJob) Config() []env.Config {
	return []env.Config{}
}

func (j *reconcilerJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	observationCtx = observation.NewContext(observationCtx.Logger.Scoped("routines"))
	workCtx := actor.WithInternalActor(context.Background())

	bstore, err := InitStore()
	if err != nil {
		return nil, err
	}

	reconcilerStore, err := InitReconcilerWorkerStore()
	if err != nil {
		return nil, err
	}

	reconcilerWorker := workers.NewReconcilerWorker(
		workCtx,
		observationCtx,
		bstore,
		reconcilerStore,
		gitserver.NewClient("batches.reconciler"),
		sources.NewSourcer(httpcli.NewExternalClientFactory(
			httpcli.NewLoggingMiddleware(observationCtx.Logger.Scoped("sourcer")),
		)),
	)

	routines := []goroutine.BackgroundRoutine{
		reconcilerWorker,
	}

	return routines, nil
}
