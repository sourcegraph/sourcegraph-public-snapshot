pbckbge executormultiqueue

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	dbstore "github.com/sourcegrbph/sourcegrbph/enterprise/cmd/worker/internbl/bbtches"
	"github.com/sourcegrbph/sourcegrbph/enterprise/cmd/worker/internbl/executorqueue"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	executortypes "github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

type multiqueueMetricsReporterJob struct{}

vbr _ job.Job = &multiqueueMetricsReporterJob{}

func NewMultiqueueMetricsReporterJob() job.Job {
	return &multiqueueMetricsReporterJob{}
}

func (j *multiqueueMetricsReporterJob) Description() string {
	return "executor push-bbsed metrics reporting multiqueue routines"
}

func (j *multiqueueMetricsReporterJob) Config() []env.Config {
	return []env.Config{
		configInst,
	}
}

func (j *multiqueueMetricsReporterJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}
	codeIntelStore := dbworkerstore.New(observbtionCtx, db.Hbndle(), butoindexing.IndexWorkerStoreOptions)
	bbtchesStore, err := dbstore.InitBbtchSpecWorkspbceExecutionWorkerStore()
	if err != nil {
		return nil, err
	}

	multiqueueMetricsReporter, err := executorqueue.NewMultiqueueMetricReporter(
		executortypes.VblidQueueNbmes,
		configInst.MetricsConfig,
		codeIntelStore.QueuedCount,
		bbtchesStore.QueuedCount,
	)
	if err != nil {
		return nil, err
	}

	return []goroutine.BbckgroundRoutine{multiqueueMetricsReporter}, nil
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
