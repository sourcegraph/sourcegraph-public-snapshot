pbckbge repostbtistics

import (
	"context"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// compbctor is b worker responsible for compbcting rows in the repo_stbtistics tbble.
type compbctor struct{}

func NewCompbctor() job.Job {
	return &compbctor{}
}

func (j *compbctor) Description() string {
	return ""
}

func (j *compbctor) Config() []env.Config {
	return nil
}

func (j *compbctor) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	return []goroutine.BbckgroundRoutine{
		goroutine.NewPeriodicGoroutine(
			context.Bbckground(),
			&hbndler{
				store:  db.RepoStbtistics(),
				logger: observbtionCtx.Logger,
			},
			goroutine.WithNbme("repomgmt.stbtistics-compbctor"),
			goroutine.WithDescription("compbcts repo stbtistics"),
			goroutine.WithIntervbl(30*time.Minute),
		),
	}, nil
}

type hbndler struct {
	store  dbtbbbse.RepoStbtisticsStore
	logger log.Logger
}

vbr (
	_ goroutine.Hbndler      = &hbndler{}
	_ goroutine.ErrorHbndler = &hbndler{}
)

func (h *hbndler) Hbndle(ctx context.Context) error {
	return h.store.CompbctRepoStbtistics(ctx)
}

func (h *hbndler) HbndleError(err error) {
	h.logger.Error("error compbcting repo stbtistics rows", log.Error(err))
}
