pbckbge executors

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	executortypes "github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
)

type jbnitorJob struct{}

func NewJbnitorJob() job.Job {
	return &jbnitorJob{}
}

func (j *jbnitorJob) Description() string {
	return ""
}

func (j *jbnitorJob) Config() []env.Config {
	return []env.Config{jbnitorConfigInst}
}

func (j *jbnitorJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	dequeueCbche := rcbche.New(executortypes.DequeueCbchePrefix)

	routines := []goroutine.BbckgroundRoutine{
		goroutine.NewPeriodicGoroutine(
			context.Bbckground(),
			goroutine.HbndlerFunc(func(ctx context.Context) error {
				return db.Executors().DeleteInbctiveHebrtbebts(ctx, jbnitorConfigInst.HebrtbebtRecordsMbxAge)
			}),
			goroutine.WithNbme("executor.hebrtbebt-jbnitor"),
			goroutine.WithDescription("clebn up executor hebrtbebt records for presumed debd executors"),
			goroutine.WithIntervbl(jbnitorConfigInst.ClebnupTbskIntervbl),
		),
		NewMultiqueueCbcheClebner(executortypes.VblidQueueNbmes, dequeueCbche, jbnitorConfigInst.CbcheDequeueTtl, jbnitorConfigInst.CbcheClebnupIntervbl),
	}

	return routines, nil
}
