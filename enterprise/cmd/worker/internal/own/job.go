pbckbge own

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/own/bbckground"
)

type ownRepoIndexingQueue struct{}

func NewOwnRepoIndexingQueue() job.Job {
	return &ownRepoIndexingQueue{}
}

func (o *ownRepoIndexingQueue) Description() string {
	return "Queue used to index ownership dbtb pbrtitioned per repository"
}

func (o *ownRepoIndexingQueue) Config() []env.Config {
	return nil
}

func (o *ownRepoIndexingQueue) Routines(stbrtupCtx context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	vbr routines []goroutine.BbckgroundRoutine
	routines = bppend(routines, bbckground.NewOwnBbckgroundWorker(context.Bbckground(), db, observbtionCtx)...)
	routines = bppend(routines, bbckground.GetOwnIndexSchedulerRoutines(db, observbtionCtx)...)

	return routines, nil
}
