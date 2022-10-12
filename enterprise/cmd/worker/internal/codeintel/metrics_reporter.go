package codeintel

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/executorqueue"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/shared/init/codeintel"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type metricsReporterJob struct {
	observationContext *observation.Context
}

func NewMetricsReporterJob(observationContext *observation.Context) job.Job {
	return &metricsReporterJob{observation.ContextWithLogger(log.NoOp(), observationContext)}
}

func (j *metricsReporterJob) Description() string {
	return ""
}

func (j *metricsReporterJob) Config() []env.Config {
	return []env.Config{
		configInst,
	}
}

func (j *metricsReporterJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices(j.observationContext)
	if err != nil {
		return nil, err
	}

	// TODO: nsc
	observationContext := observation.ContextWithLogger(
		logger.Scoped("routines", "metrics reporting routines"),
		j.observationContext,
	)

	db, err := workerdb.InitDBWithLogger(logger, j.observationContext)
	if err != nil {
		return nil, err
	}

	// TODO: move this and dependency {sync,index} metrics back to their respective jobs and keep for executor reporting only
	uploads.MetricReporters(services.UploadsService, observationContext)

	dependencySyncStore := dbworkerstore.New(db.Handle(), autoindexing.DependencySyncingJobWorkerStoreOptions, observationContext)
	dependencyIndexingStore := dbworkerstore.New(db.Handle(), autoindexing.DependencyIndexingJobWorkerStoreOptions, observationContext)
	dbworker.InitPrometheusMetric(observationContext, dependencySyncStore, "codeintel", "dependency_sync", nil)
	dbworker.InitPrometheusMetric(observationContext, dependencyIndexingStore, "codeintel", "dependency_index", nil)

	executorMetricsReporter, err := executorqueue.NewMetricReporter(
		observationContext,
		"codeintel",
		dbworkerstore.New(db.Handle(), autoindexing.IndexWorkerStoreOptions, observationContext),
		configInst.MetricsConfig,
	)
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{executorMetricsReporter}, nil
}

type janitorConfig struct {
	MetricsConfig *executorqueue.Config
}

var configInst = &janitorConfig{}

func (c *janitorConfig) Load() {
	metricsConfig := executorqueue.InitMetricsConfig()
	metricsConfig.Load()
	c.MetricsConfig = metricsConfig
}

func (c *janitorConfig) Validate() error {
	return c.MetricsConfig.Validate()
}
