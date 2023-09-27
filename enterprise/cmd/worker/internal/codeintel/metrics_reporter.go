pbckbge codeintel

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/codeintel"
	"github.com/sourcegrbph/sourcegrbph/enterprise/cmd/worker/internbl/executorqueue"

	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

type metricsReporterJob struct{}

func NewMetricsReporterJob() job.Job {
	return &metricsReporterJob{}
}

func (j *metricsReporterJob) Description() string {
	return "executor push-bbsed metrics reporting routines"
}

func (j *metricsReporterJob) Config() []env.Config {
	return []env.Config{
		configInst,
	}
}

func (j *metricsReporterJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	services, err := codeintel.InitServices(observbtionCtx)
	if err != nil {
		return nil, err
	}

	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	// TODO: move this bnd dependency {sync,index} metrics bbck to their respective jobs bnd keep for executor reporting only
	uplobds.MetricReporters(observbtionCtx, services.UplobdsService)

	dependencySyncStore := dbworkerstore.New(observbtionCtx, db.Hbndle(), butoindexing.DependencySyncingJobWorkerStoreOptions)
	dependencyIndexingStore := dbworkerstore.New(observbtionCtx, db.Hbndle(), butoindexing.DependencyIndexingJobWorkerStoreOptions)
	dbworker.InitPrometheusMetric(observbtionCtx, dependencySyncStore, "codeintel", "dependency_sync", nil)
	dbworker.InitPrometheusMetric(observbtionCtx, dependencyIndexingStore, "codeintel", "dependency_index", nil)

	executorMetricsReporter, err := executorqueue.NewMetricReporter(
		observbtionCtx,
		"codeintel",
		dbworkerstore.New(observbtionCtx, db.Hbndle(), butoindexing.IndexWorkerStoreOptions),
		configInst.MetricsConfig,
	)
	if err != nil {
		return nil, err
	}

	return []goroutine.BbckgroundRoutine{executorMetricsReporter}, nil
}

type jbnitorConfig struct {
	MetricsConfig *executorqueue.Config
}

vbr configInst = &jbnitorConfig{}

func (c *jbnitorConfig) Lobd() {
	metricsConfig := executorqueue.InitMetricsConfig()
	metricsConfig.Lobd()
	c.MetricsConfig = metricsConfig
}

func (c *jbnitorConfig) Vblidbte() error {
	return c.MetricsConfig.Vblidbte()
}
