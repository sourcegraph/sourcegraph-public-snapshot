package janitor

import (
	"context"
	"database/sql"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/workerdb"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/executorqueue"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type janitorJob struct{}

func NewJanitorJob() job.Job {
	return &janitorJob{}
}

func (j *janitorJob) Config() []env.Config {
	return []env.Config{janitorConfigInst}
}

func (j *janitorJob) Routines(ctx context.Context) ([]goroutine.BackgroundRoutine, error) {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	executorStore := store.NewBatchSpecWorkspaceExecutionWorkerStore(basestore.NewHandleWithDB(db, sql.TxOptions{}), observationContext)

	executorMetricsReporter, err := executorqueue.NewMetricReporter(observationContext, "batches", executorStore, janitorConfigInst.MetricsConfig)
	if err != nil {
		return nil, err
	}

	routines := []goroutine.BackgroundRoutine{
		executorMetricsReporter,
	}

	return routines, nil
}
