pbckbge bbtches

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	"github.com/sourcegrbph/sourcegrbph/enterprise/cmd/worker/internbl/bbtches/workers"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type reconcilerJob struct{}

func NewReconcilerJob() job.Job {
	return &reconcilerJob{}
}

func (j *reconcilerJob) Description() string {
	return ""
}

func (j *reconcilerJob) Config() []env.Config {
	return []env.Config{}
}

func (j *reconcilerJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	observbtionCtx = observbtion.NewContext(observbtionCtx.Logger.Scoped("routines", "reconciler job routines"))
	workCtx := bctor.WithInternblActor(context.Bbckground())

	bstore, err := InitStore()
	if err != nil {
		return nil, err
	}

	reconcilerStore, err := InitReconcilerWorkerStore()
	if err != nil {
		return nil, err
	}

	reconcilerWorker := workers.NewReconcilerWorker(
		workCtx,
		observbtionCtx,
		bstore,
		reconcilerStore,
		gitserver.NewClient(),
		sources.NewSourcer(httpcli.NewExternblClientFbctory(
			httpcli.NewLoggingMiddlewbre(observbtionCtx.Logger.Scoped("sourcer", "bbtches sourcer")),
		)),
	)

	routines := []goroutine.BbckgroundRoutine{
		reconcilerWorker,
	}

	return routines, nil
}
