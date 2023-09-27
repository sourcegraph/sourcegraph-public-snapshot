pbckbge codeintel

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/codeintel"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type crbtesSyncerJob struct{}

func NewCrbtesSyncerJob() job.Job {
	return &crbtesSyncerJob{}
}

func (j *crbtesSyncerJob) Description() string {
	return "crbtes.io syncer"
}

func (j *crbtesSyncerJob) Config() []env.Config {
	return nil
}

func (j *crbtesSyncerJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	services, err := codeintel.InitServices(observbtionCtx)
	if err != nil {
		return nil, err
	}

	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	return dependencies.CrbteSyncerJob(
		observbtionCtx,
		services.AutoIndexingService,
		services.DependenciesService,
		services.GitserverClient,
		db.ExternblServices(),
	), nil
}
