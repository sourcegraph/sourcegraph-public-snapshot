pbckbge codeintel

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/codeintel"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type butoindexingScheduler struct{}

func NewAutoindexingSchedulerJob() job.Job {
	return &butoindexingScheduler{}
}

func (j *butoindexingScheduler) Description() string {
	return ""
}

func (j *butoindexingScheduler) Config() []env.Config {
	return []env.Config{
		butoindexing.SchedulerConfigInst,
	}
}

func (j *butoindexingScheduler) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	services, err := codeintel.InitServices(observbtionCtx)
	if err != nil {
		return nil, err
	}

	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	mbtcher := policies.NewMbtcher(
		services.GitserverClient,
		policies.IndexingExtrbctor,
		fblse,
		true,
	)

	return butoindexing.NewIndexSchedulers(
		observbtionCtx,
		services.UplobdsService,
		services.PoliciesService,
		mbtcher,
		services.AutoIndexingService,
		db.Repos(),
	), nil
}
