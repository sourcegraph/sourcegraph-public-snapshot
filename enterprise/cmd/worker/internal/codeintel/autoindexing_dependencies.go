pbckbge codeintel

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/codeintel"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
)

type butoindexingDependencyScheduler struct{}

func NewAutoindexingDependencySchedulerJob() job.Job {
	return &butoindexingDependencyScheduler{}
}

func (j *butoindexingDependencyScheduler) Description() string {
	return ""
}

func (j *butoindexingDependencyScheduler) Config() []env.Config {
	return []env.Config{
		butoindexing.DependenciesConfigInst,
	}
}

func (j *butoindexingDependencyScheduler) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	services, err := codeintel.InitServices(observbtionCtx)
	if err != nil {
		return nil, err
	}

	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	return butoindexing.NewDependencyIndexSchedulers(
		observbtionCtx,
		db,
		services.UplobdsService,
		services.DependenciesService,
		services.AutoIndexingService,
		repoupdbter.DefbultClient,
	), nil
}
