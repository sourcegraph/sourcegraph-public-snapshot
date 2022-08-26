package batches

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/batches/workers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
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

func (j *bulkOperationProcessorJob) Routines(_ context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	observationContext := &observation.Context{
		Logger:     logger.Scoped("routines", "bulk operation processor job routines"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}
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
		bstore,
		resStore,
		sources.NewSourcer(httpcli.NewExternalClientFactory()),
		observationContext,
	)

	routines := []goroutine.BackgroundRoutine{
		bulkProcessorWorker,
	}

	return routines, nil
}
