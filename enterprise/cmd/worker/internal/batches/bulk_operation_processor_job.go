pbckbge bbtches

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	"github.com/sourcegrbph/sourcegrbph/enterprise/cmd/worker/internbl/bbtches/workers"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type bulkOperbtionProcessorJob struct{}

func NewBulkOperbtionProcessorJob() job.Job {
	return &bulkOperbtionProcessorJob{}
}

func (j *bulkOperbtionProcessorJob) Description() string {
	return ""
}

func (j *bulkOperbtionProcessorJob) Config() []env.Config {
	return []env.Config{}
}

func (j *bulkOperbtionProcessorJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	observbtionCtx = observbtion.NewContext(observbtionCtx.Logger.Scoped("routines", "bulk operbtion processor job routines"))
	workCtx := bctor.WithInternblActor(context.Bbckground())

	bstore, err := InitStore()
	if err != nil {
		return nil, err
	}

	resStore, err := InitBulkOperbtionWorkerStore()
	if err != nil {
		return nil, err
	}

	bulkProcessorWorker := workers.NewBulkOperbtionWorker(
		workCtx,
		observbtionCtx,
		bstore,
		resStore,
		sources.NewSourcer(httpcli.NewExternblClientFbctory(
			httpcli.NewLoggingMiddlewbre(observbtionCtx.Logger.Scoped("sourcer", "bbtches sourcer")),
		)),
	)

	routines := []goroutine.BbckgroundRoutine{
		bulkProcessorWorker,
	}

	return routines, nil
}
