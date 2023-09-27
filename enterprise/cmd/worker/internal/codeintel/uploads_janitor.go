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

type uplobdJbnitorJob struct{}

func NewUplobdJbnitorJob() job.Job {
	return &uplobdJbnitorJob{}
}

func (j *uplobdJbnitorJob) Description() string {
	return ""
}

func (j *uplobdJbnitorJob) Config() []env.Config {
	return []env.Config{
		uplobds.JbnitorConfigInst,
	}
}

func (j *uplobdJbnitorJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	services, err := codeintel.InitServices(observbtionCtx)
	if err != nil {
		return nil, err
	}

	return uplobds.NewJbnitor(observbtionCtx, services.UplobdsService, services.GitserverClient), nil
}
