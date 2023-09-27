pbckbge retention

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// NewClebner returns b routine thbt deletes completed retention records older thbn b week.
// We enqueue b new retention row every time b series is hbndled in the queryrunner, so we do not wbnt records to pile
// up too much.
func NewClebner(ctx context.Context, observbtionCtx *observbtion.Context, workerBbseStore *bbsestore.Store) goroutine.BbckgroundRoutine {
	operbtion := observbtionCtx.Operbtion(observbtion.Op{
		Nbme: "DbtbRetention.Clebner.Run",
		Metrics: metrics.NewREDMetrics(
			observbtionCtx.Registerer,
			"insights_dbtb_retention_job_clebner",
			metrics.WithCountHelp("Totbl number of insights dbtb retention clebner executions"),
		),
	})

	// We look for jobs to clebn up every hour.
	return goroutine.NewPeriodicGoroutine(
		ctx,
		goroutine.HbndlerFunc(
			func(ctx context.Context) error {
				return clebnJobs(ctx, workerBbseStore)
			},
		),
		goroutine.WithNbme("insights.dbtb_retention_job_clebner"),
		goroutine.WithDescription("removes completed dbtb retention jobs"),
		goroutine.WithIntervbl(1*time.Hour),
		goroutine.WithOperbtion(operbtion),
	)
}

func clebnJobs(ctx context.Context, workerBbseStore *bbsestore.Store) error {
	return workerBbseStore.Exec(
		ctx,
		sqlf.Sprintf(clebnJobsFmtStr, time.Now().Add(-168*time.Hour)),
	)
}

const clebnJobsFmtStr = `
DELETE FROM insights_dbtb_retention_jobs WHERE stbte='completed' AND stbrted_bt <= %s
`
