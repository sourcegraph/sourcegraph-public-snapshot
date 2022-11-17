package batches

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/batches/janitor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/executorqueue"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type janitorJob struct{}

func NewJanitorJob() job.Job {
	return &janitorJob{}
}

func (j *janitorJob) Description() string {
	return ""
}

func (j *janitorJob) Config() []env.Config {
	return []env.Config{janitorConfigInst}
}

func (j *janitorJob) Routines(_ context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	observationContext := &observation.Context{
		Logger:     logger.Scoped("routines", "janitor job routines"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}
	workCtx := actor.WithInternalActor(context.Background())

	bstore, err := InitStore()
	if err != nil {
		return nil, err
	}

	janitorMetrics := janitor.NewMetrics(observationContext)

	reconcilerStore, err := InitReconcilerWorkerStore()
	if err != nil {
		return nil, err
	}
	bulkOperationStore, err := InitBulkOperationWorkerStore()
	if err != nil {
		return nil, err
	}
	workspaceExecutionStore, err := InitBatchSpecWorkspaceExecutionWorkerStore()
	if err != nil {
		return nil, err
	}
	workspaceResolutionStore, err := InitBatchSpecResolutionWorkerStore()
	if err != nil {
		return nil, err
	}

	executorMetricsReporter, err := executorqueue.NewMetricReporter(observationContext, "batches", workspaceExecutionStore, janitorConfigInst.MetricsConfig)
	if err != nil {
		return nil, err
	}

	routines := []goroutine.BackgroundRoutine{
		executorMetricsReporter,

		janitor.NewReconcilerWorkerResetter(
			logger.Scoped("ReconcilerWorkerResetter", ""),
			reconcilerStore,
			janitorMetrics,
		),
		janitor.NewBulkOperationWorkerResetter(
			logger.Scoped("BulkOperationWorkerResetter", ""),
			bulkOperationStore,
			janitorMetrics,
		),
		janitor.NewBatchSpecWorkspaceExecutionWorkerResetter(
			logger.Scoped("BatchSpecWorkspaceExecutionWorkerResetter", ""),
			workspaceExecutionStore,
			janitorMetrics,
		),
		janitor.NewBatchSpecWorkspaceResolutionWorkerResetter(
			logger.Scoped("BatchSpecWorkspaceResolutionWorkerResetter", ""),
			workspaceResolutionStore,
			janitorMetrics,
		),

		janitor.NewSpecExpirer(workCtx, bstore),
		janitor.NewCacheEntryCleaner(workCtx, bstore),
		janitor.NewChangesetDetachedCleaner(workCtx, bstore),
	}

	return routines, nil
}
