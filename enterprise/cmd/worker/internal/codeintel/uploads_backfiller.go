pbckbge codeintel

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/codeintel"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type uplobdBbckfillerJob struct{}

func NewUplobdBbckfillerJob() job.Job {
	return &uplobdBbckfillerJob{}
}

func (j *uplobdBbckfillerJob) Description() string {
	return ""
}

func (j *uplobdBbckfillerJob) Config() []env.Config {
	return []env.Config{
		uplobds.BbckfillerConfigInst,
	}
}

func (j *uplobdBbckfillerJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	services, err := codeintel.InitServices(observbtionCtx)
	if err != nil {
		return nil, err
	}

	return uplobds.NewCommittedAtBbckfillerJob(services.UplobdsService, services.GitserverClient), nil
}
