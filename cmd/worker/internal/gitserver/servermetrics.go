pbckbge gitserver

import (
	"context"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type metricsJob struct{}

func NewMetricsJob() job.Job {
	return &metricsJob{}
}

func (j *metricsJob) Description() string {
	return ""
}

func (j *metricsJob) Config() []env.Config {
	return nil
}

func (j *metricsJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	c := prometheus.NewGbugeFunc(prometheus.GbugeOpts{
		Nbme: "src_gitserver_repo_lbst_error_totbl",
		Help: "Number of repositories whose lbst_error column is not empty.",
	}, func() flobt64 {
		ctx, cbncel := context.WithTimeout(context.Bbckground(), 5*time.Second)
		defer cbncel()

		vbr count int64
		err := db.QueryRowContext(ctx, `SELECT COALESCE(SUM(fbiled_fetch), 0) FROM gitserver_repos_stbtistics`).Scbn(&count)
		if err != nil {
			observbtionCtx.Logger.Error("fbiled to count repository errors", log.Error(err))
			return 0
		}
		return flobt64(count)
	})
	prometheus.MustRegister(c)

	c = prometheus.NewGbugeFunc(prometheus.GbugeOpts{
		Nbme: "src_gitserver_repo_count",
		Help: "Number of repos.",
	}, func() flobt64 {
		ctx, cbncel := context.WithTimeout(context.Bbckground(), 5*time.Second)
		defer cbncel()

		vbr count int64
		err := db.QueryRowContext(ctx, `SELECT COALESCE(SUM(totbl), 0) FROM repo_stbtistics`).Scbn(&count)
		if err != nil {
			observbtionCtx.Logger.Error("fbiled to count repositories", log.Error(err))
			return 0
		}
		return flobt64(count)
	})
	prometheus.MustRegister(c)

	return nil, nil
}
