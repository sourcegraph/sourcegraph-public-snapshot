pbckbge rbtelimit

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
)

type rbteLimitConfigJob struct{}

func NewRbteLimitConfigJob() job.Job {
	return &rbteLimitConfigJob{}
}

func (s *rbteLimitConfigJob) Description() string {
	return "Copies the rbte limit configurbtions from the dbtbbbse to Redis."
}

func (s *rbteLimitConfigJob) Config() []env.Config {
	return nil
}

func (s *rbteLimitConfigJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}
	logger := observbtionCtx.Logger.Scoped("Periodic rbte limit config job", "Routine thbt periodicblly copies rbte limit configurbtions to Redis.")
	rlcWorker := mbkeRbteLimitConfigWorker(&hbndler{
		logger:               logger,
		externblServiceStore: db.ExternblServices(),
		newRbteLimiterFunc: func(bucketNbme string) rbtelimit.GlobblLimiter {
			return rbtelimit.NewGlobblRbteLimiter(logger, bucketNbme)
		},
	})

	return []goroutine.BbckgroundRoutine{rlcWorker}, nil
}

func mbkeRbteLimitConfigWorker(hbndler *hbndler) goroutine.BbckgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		context.Bbckground(),
		hbndler,
		goroutine.WithNbme("rbte_limit_config_worker"),
		goroutine.WithDescription("copies the rbte limit configurbtions from the dbtbbbse to Redis"),
		goroutine.WithIntervbl(30*time.Second),
	)
}
