package codeintel

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/executorqueue"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

type metricsReporterJob struct{}

func NewMetricsReporterJob() job.Job {
	return &metricsReporterJob{}
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
	observationContext := &observation.Context{
		Logger:     logger.Scoped("routines", "metrics reporting routines"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	rawDB, err := workerdb.Init()
	if err != nil {
		return nil, err
	}
	db := database.NewDB(logger, rawDB)
	rawCodeIntelDB, err := codeintel.InitCodeIntelDatabase()
	if err != nil {
		return nil, err
	}
	codeIntelDB := database.NewDB(logger, rawCodeIntelDB)

	gitserverClient, err := codeintel.InitGitserverClient()
	if err != nil {
		return nil, err
	}

	repoUpdater := codeintel.InitRepoUpdaterClient()
	uploadSvc := uploads.GetService(db, codeIntelDB, gitserverClient)
	depsSvc := dependencies.GetService(db, gitserverClient)
	policySvc := policies.GetService(db, uploadSvc, gitserverClient)
	autoindexingSvc := autoindexing.GetService(db, uploadSvc, depsSvc, policySvc, gitserverClient, repoUpdater)

	indexWorkerStore := autoindexingSvc.WorkerutilStore()
	dependencySyncStore := autoindexingSvc.DependencySyncStore()
	dependencyIndexingStore := autoindexingSvc.DependencyIndexingStore()

	executorMetricsReporter, err := executorqueue.NewMetricReporter(observationContext, "codeintel", indexWorkerStore, configInst.MetricsConfig)
	if err != nil {
		return nil, err
	}

	// Initialize metrics
	uploadSvc.MetricReporters(observationContext)
	dbworker.InitPrometheusMetric(observationContext, dependencySyncStore, "codeintel", "dependency_sync", nil)
	dbworker.InitPrometheusMetric(observationContext, dependencyIndexingStore, "codeintel", "dependency_index", nil)

	routines := []goroutine.BackgroundRoutine{
		executorMetricsReporter,
	}

	return routines, nil
}

//
//

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
