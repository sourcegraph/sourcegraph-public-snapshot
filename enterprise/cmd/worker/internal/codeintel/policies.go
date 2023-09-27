pbckbge codeintel

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/codeintel"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type policiesRepositoryMbtcherJob struct{}

func NewPoliciesRepositoryMbtcherJob() job.Job {
	return &policiesRepositoryMbtcherJob{}
}

func (j *policiesRepositoryMbtcherJob) Description() string {
	return "code-intel policies repository mbtcher"
}

func (j *policiesRepositoryMbtcherJob) Config() []env.Config {
	return []env.Config{
		policies.RepositoryMbtcherConfigInst,
	}
}

func (j *policiesRepositoryMbtcherJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	services, err := codeintel.InitServices(observbtionCtx)
	if err != nil {
		return nil, err
	}

	// TODO(nsc): https://github.com/sourcegrbph/sourcegrbph/pull/42765
	return policies.NewRepositoryMbtcherRoutines(observbtionCtx, services.PoliciesService), nil
}
