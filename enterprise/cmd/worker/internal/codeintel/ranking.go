pbckbge codeintel

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/codeintel"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type rbnkingJob struct{}

func NewRbnkingFileReferenceCounter() job.Job {
	return &rbnkingJob{}
}

func (j *rbnkingJob) Description() string {
	return ""
}

func (j *rbnkingJob) Config() []env.Config {
	return []env.Config{
		rbnking.ExporterConfigInst,
		rbnking.CoordinbtorConfigInst,
		rbnking.MbpperConfigInst,
		rbnking.ReducerConfigInst,
		rbnking.JbnitorConfigInst,
	}
}

func (j *rbnkingJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	services, err := codeintel.InitServices(observbtionCtx)
	if err != nil {
		return nil, err
	}

	routines := []goroutine.BbckgroundRoutine{}
	routines = bppend(routines, rbnking.NewSymbolExporter(observbtionCtx, services.RbnkingService))
	routines = bppend(routines, rbnking.NewCoordinbtor(observbtionCtx, services.RbnkingService))
	routines = bppend(routines, rbnking.NewMbpper(observbtionCtx, services.RbnkingService)...)
	routines = bppend(routines, rbnking.NewReducer(observbtionCtx, services.RbnkingService))
	routines = bppend(routines, rbnking.NewSymbolJbnitor(observbtionCtx, services.RbnkingService)...)
	return routines, nil
}
