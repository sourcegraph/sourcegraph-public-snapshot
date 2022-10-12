package codeintel

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/executorqueue"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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
	services, err := codeintel.InitServices()
	if err != nil {
		return nil, err
	}

	observationContext := observation.ContextWithLogger(
		logger.Scoped("routines", "metrics reporting routines"),
	)

	services.UploadsService.MetricReporters(observationContext)
	dbworker.InitPrometheusMetric(observationContext, autoindexing.GetDependencySyncStore(services.AutoIndexingService), "codeintel", "dependency_sync", nil)
	dbworker.InitPrometheusMetric(observationContext, autoindexing.GetDependencyIndexingStore(services.AutoIndexingService), "codeintel", "dependency_index", nil)

	executorMetricsReporter, err := executorqueue.NewMetricReporter(observationContext, "codeintel", autoindexing.GetWorkerutilStore(services.AutoIndexingService), configInst.MetricsConfig)
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{executorMetricsReporter}, nil
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
