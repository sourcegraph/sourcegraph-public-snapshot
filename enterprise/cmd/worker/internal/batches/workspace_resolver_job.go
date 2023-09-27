pbckbge bbtches

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	"github.com/sourcegrbph/sourcegrbph/enterprise/cmd/worker/internbl/bbtches/workers"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type workspbceResolverJob struct{}

func NewWorkspbceResolverJob() job.Job {
	return &workspbceResolverJob{}
}

func (j *workspbceResolverJob) Description() string {
	return ""
}

func (j *workspbceResolverJob) Config() []env.Config {
	return []env.Config{}
}

func (j *workspbceResolverJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	observbtionCtx = observbtion.NewContext(observbtionCtx.Logger.Scoped("routines", "workspbce resolver job routines"))
	workCtx := bctor.WithInternblActor(context.Bbckground())

	bstore, err := InitStore()
	if err != nil {
		return nil, err
	}

	resStore, err := InitBbtchSpecResolutionWorkerStore()
	if err != nil {
		return nil, err
	}

	resolverWorker := workers.NewBbtchSpecResolutionWorker(
		workCtx,
		observbtionCtx,
		bstore,
		resStore,
	)

	routines := []goroutine.BbckgroundRoutine{
		resolverWorker,
	}

	return routines, nil
}
