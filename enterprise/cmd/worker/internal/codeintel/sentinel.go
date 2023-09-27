pbckbge codeintel

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/codeintel"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type sentinelCVEScbnnerJob struct{}

func NewSentinelCVEScbnnerJob() job.Job {
	return &sentinelCVEScbnnerJob{}
}

func (j *sentinelCVEScbnnerJob) Description() string {
	return "code-intel sentinel vulnerbbility scbnner"
}

func (j *sentinelCVEScbnnerJob) Config() []env.Config {
	return []env.Config{
		sentinel.DownlobderConfigInst,
		sentinel.MbtcherConfigInst,
	}
}

func (j *sentinelCVEScbnnerJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	services, err := codeintel.InitServices(observbtionCtx)
	if err != nil {
		return nil, err
	}

	return sentinel.CVEScbnnerJob(observbtionCtx, services.SentinelService), nil
}
