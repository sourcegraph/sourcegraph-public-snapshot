package batches

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/internal/batches/workers"
	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/batches"
	"github.com/sourcegraph/sourcegraph/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type bulkOperationProcessorJob struct{}

func NewBulkOperationProcessorJob() job.Job {
	return &bulkOperationProcessorJob{}
}

func (j *bulkOperationProcessorJob) Description() string {
	return ""
}

func (j *bulkOperationProcessorJob) Config() []env.Config {
	return []env.Config{}
}

func (j *bulkOperationProcessorJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	if !batches.IsEnabled() {
		return nil, nil
	}
	observationCtx = observation.NewContext(observationCtx.Logger.Scoped("routines"))
	workCtx := actor.WithInternalActor(context.Background())

	bstore, err := InitStore()
	if err != nil {
		return nil, err
	}

	resStore, err := InitBulkOperationWorkerStore()
	if err != nil {
		return nil, err
	}

	bulkProcessorWorker := workers.NewBulkOperationWorker(
		workCtx,
		observationCtx,
		bstore,
		resStore,
		sources.NewSourcer(httpcli.NewExternalClientFactory(
			httpcli.NewLoggingMiddleware(observationCtx.Logger.Scoped("sourcer")),
		)),
	)

	routines := []goroutine.BackgroundRoutine{
		bulkProcessorWorker,
	}

	return routines, nil
}
