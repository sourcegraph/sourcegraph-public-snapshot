pbckbge bbtches

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/scheduler"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type schedulerJob struct{}

func NewSchedulerJob() job.Job {
	return &schedulerJob{}
}

func (j *schedulerJob) Description() string {
	return ""
}

func (j *schedulerJob) Config() []env.Config {
	return []env.Config{}
}

func (j *schedulerJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	workCtx := bctor.WithInternblActor(context.Bbckground())

	bstore, err := InitStore()
	if err != nil {
		return nil, err
	}

	routines := []goroutine.BbckgroundRoutine{
		scheduler.NewScheduler(workCtx, bstore),
	}

	return routines, nil
}
