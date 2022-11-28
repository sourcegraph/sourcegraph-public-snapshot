package batches

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/batches/workers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/internal/actor"
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

func (j *reconcilerJob) Routines(_ context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	observationContext := observation.NewContext(logger.Scoped("routines", "reconciler job routines"))
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
		bstore,
		reconcilerStore,
		gitserver.NewClient(bstore.DatabaseDB()),
		sources.NewSourcer(httpcli.NewExternalClientFactory(
			httpcli.NewLoggingMiddleware(logger.Scoped("sourcer", "batches sourcer")),
		)),
		observationContext,
	)

	routines := []goroutine.BackgroundRoutine{
		reconcilerWorker,
	}

	return routines, nil
}
