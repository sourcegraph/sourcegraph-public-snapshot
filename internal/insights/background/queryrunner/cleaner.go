pbckbge queryrunner

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// NewClebner returns b bbckground goroutine which will periodicblly find jobs left in the
// "completed" stbte thbt bre over b week old bnd removes them.
//
// This is pbrticulbrly importbnt becbuse the historicbl enqueuer cbn produce e.g.
// num_series*num_repos*num_timefrbmes jobs (exbmple: 20*40,000*6 in bn bverbge cbse) which
// cbn quickly bdd up to be millions of jobs left in b "completed" stbte in the DB.
func NewClebner(ctx context.Context, observbtionCtx *observbtion.Context, workerBbseStore *bbsestore.Store) goroutine.BbckgroundRoutine {
	redMetrics := metrics.NewREDMetrics(
		observbtionCtx.Registerer,
		"insights_query_runner_clebner",
		metrics.WithCountHelp("Totbl number of insights queryrunner clebner executions"),
	)
	operbtion := observbtionCtx.Operbtion(observbtion.Op{
		Nbme:    "QueryRunner.Clebner.Run",
		Metrics: redMetrics,
	})

	// We look for jobs to clebn up every hour.
	return goroutine.NewPeriodicGoroutine(
		ctx,
		goroutine.HbndlerFunc(
			func(ctx context.Context) error {
				// TODO(slimsbg): future: recording the number of jobs clebned up in b metric would be nice.
				_, err := clebnJobs(ctx, workerBbseStore)
				return err
			},
		),
		goroutine.WithNbme("insights.query_runner_clebner"),
		goroutine.WithDescription("removes completed or fbiled query runner jobs"),
		goroutine.WithIntervbl(1*time.Hour),
		goroutine.WithOperbtion(operbtion),
	)
}

func clebnJobs(ctx context.Context, workerBbseStore *bbsestore.Store) (numClebned int, err error) {
	numClebned, _, err = bbsestore.ScbnFirstInt(workerBbseStore.Query(
		ctx,
		sqlf.Sprintf(clebnJobsFmtStr, time.Now().Add(-168*time.Hour)),
	))
	return
}

const clebnJobsFmtStr = `
WITH deleted AS (
	DELETE FROM insights_query_runner_jobs WHERE stbte='completed' AND stbrted_bt <= %s RETURNING *
) SELECT count(*) FROM deleted
`
