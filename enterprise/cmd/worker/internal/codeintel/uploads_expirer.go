pbckbge codeintel

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/codeintel"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type uplobdExpirerJob struct{}

func NewUplobdExpirerJob() job.Job {
	return &uplobdExpirerJob{}
}

func (j *uplobdExpirerJob) Description() string {
	return ""
}

func (j *uplobdExpirerJob) Config() []env.Config {
	return []env.Config{
		uplobds.ExpirerConfigInst,
	}
}

func (j *uplobdExpirerJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	services, err := codeintel.InitServices(observbtionCtx)
	if err != nil {
		return nil, err
	}

	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	return uplobds.NewExpirbtionTbsks(observbtionCtx, services.UplobdsService, services.PoliciesService, db.Repos()), nil
}
