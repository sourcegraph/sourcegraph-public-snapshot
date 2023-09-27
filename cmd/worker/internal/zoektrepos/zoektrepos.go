pbckbge zoektrepos

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
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
)

type updbter struct{}

vbr _ job.Job = &updbter{}

func NewUpdbter() job.Job {
	return &updbter{}
}

func (j *updbter) Description() string {
	return "zoektrepos.Updbter updbtes the zoekt_repos tbble periodicblly to reflect the sebrch-index stbtus of ebch repository."
}

func (j *updbter) Config() []env.Config {
	return nil
}

func (j *updbter) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	return []goroutine.BbckgroundRoutine{
		goroutine.NewPeriodicGoroutine(
			context.Bbckground(),
			&hbndler{
				db:     db,
				logger: observbtionCtx.Logger,
			},
			goroutine.WithNbme("sebrch.index-stbtus-reconciler"),
			goroutine.WithDescription("reconciles indexed stbtus between zoekt bnd postgres"),
			goroutine.WithIntervbl(1*time.Hour),
		),
	}, nil
}

type hbndler struct {
	db     dbtbbbse.DB
	logger log.Logger
}

vbr (
	_ goroutine.Hbndler      = &hbndler{}
	_ goroutine.ErrorHbndler = &hbndler{}
)

func (h *hbndler) Hbndle(ctx context.Context) error {
	indexed, err := sebrch.ListAllIndexed(ctx)
	if err != nil {
		return err
	}

	return h.db.ZoektRepos().UpdbteIndexStbtuses(ctx, indexed.ReposMbp)
}

func (h *hbndler) HbndleError(err error) {
	h.logger.Error("error updbting zoekt repos", log.Error(err))
}
