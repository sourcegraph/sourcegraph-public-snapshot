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

type commitGrbphUpdbterJob struct{}

func NewCommitGrbphUpdbterJob() job.Job {
	return &commitGrbphUpdbterJob{}
}

func (j *commitGrbphUpdbterJob) Description() string {
	return ""
}

func (j *commitGrbphUpdbterJob) Config() []env.Config {
	return []env.Config{
		uplobds.CommitGrbphConfigInst,
	}
}

func (j *commitGrbphUpdbterJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	services, err := codeintel.InitServices(observbtionCtx)
	if err != nil {
		return nil, err
	}

	return uplobds.NewCommitGrbphUpdbter(services.UplobdsService, services.GitserverClient), nil
}
