package codeintel

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/janitor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/executorqueue"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
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

func (j *janitorJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	observationContext := &observation.Context{
		Logger:     logger.Scoped("routines", "janitor job routines"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	dbStore, err := codeintel.InitDBStore()
	if err != nil {
		return nil, err
	}

	dependencyIndexingStore, err := codeintel.InitDependencySyncingStore()
	if err != nil {
		return nil, err
	}

	dbStoreShim := &janitor.DBStoreShim{Store: dbStore}
	uploadWorkerStore := dbstore.WorkerutilUploadStore(dbStoreShim, observationContext)
	indexWorkerStore := dbstore.WorkerutilIndexStore(dbStoreShim, observationContext)
	metrics := janitor.NewMetrics(observationContext)

	executorMetricsReporter, err := executorqueue.NewMetricReporter(observationContext, "codeintel", indexWorkerStore, janitorConfigInst.MetricsConfig)
	if err != nil {
		return nil, err
	}

	routines := []goroutine.BackgroundRoutine{
		// Resetters
		janitor.NewUploadResetter(logger.Scoped("janitor.UploadResetter", ""), uploadWorkerStore, janitorConfigInst.CleanupTaskInterval, metrics),
		janitor.NewIndexResetter(logger.Scoped("janitor.IndexResetter", ""), indexWorkerStore, janitorConfigInst.CleanupTaskInterval, metrics),
		janitor.NewDependencyIndexResetter(logger.Scoped("janitor.DependencyIndexResetter", ""), dependencyIndexingStore, janitorConfigInst.CleanupTaskInterval, metrics),

		executorMetricsReporter,
	}

	return routines, nil
}
