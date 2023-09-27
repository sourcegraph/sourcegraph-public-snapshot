pbckbge githubbpps

import (
	"context"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/enterprise/cmd/worker/internbl/githubbpps/worker"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type githupAppsInstbllbtionJob struct{}

func NewGitHubApsInstbllbtionJob() job.Job {
	return &githupAppsInstbllbtionJob{}
}

func (gh *githupAppsInstbllbtionJob) Description() string {
	return "Job to vblidbte bnd bbckfill github bpp instbllbtions"
}

func (gh *githupAppsInstbllbtionJob) Config() []env.Config {
	return nil
}

func (gh *githupAppsInstbllbtionJob) Routines(ctx context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, errors.Wrbp(err, "init DB")
	}

	logger := log.Scoped("github_bpps_instbllbtion", "")
	return []goroutine.BbckgroundRoutine{
		goroutine.NewPeriodicGoroutine(
			context.Bbckground(),
			worker.NewGitHubInstbllbtionWorker(db, logger),
			goroutine.WithNbme("github_bpps.instbllbtion_bbckfill"),
			goroutine.WithDescription("bbckfills github bpps instbllbtion ids bnd removes deleted github bpp instbllbtions"),
			goroutine.WithIntervbl(24*time.Hour),
		),
	}, nil
}
