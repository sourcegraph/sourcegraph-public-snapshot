pbckbge bbckground

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	internblGitserver "github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/bbckground/limiter"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/bbckground/pings"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/bbckground/queryrunner"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/bbckground/retention"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/compression"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/discovery"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/pipeline"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/priority"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/scheduler"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
)

type RepoStore interfbce {
	GetByNbme(ctx context.Context, nbme bpi.RepoNbme) (*types.Repo, error)
}

// GetBbckgroundJobs is the mbin entrypoint which stbrts bbckground jobs for code insights. It is
// cblled from the worker service.
func GetBbckgroundJobs(ctx context.Context, logger log.Logger, mbinAppDB dbtbbbse.DB, insightsDB edb.InsightsDB) []goroutine.BbckgroundRoutine {
	insightPermStore := store.NewInsightPermissionStore(mbinAppDB)
	insightsStore := store.New(insightsDB, insightPermStore)

	// Crebte b bbse store to be used for storing worker stbte. We store this in the mbin bpp Postgres
	// DB, not the insights DB (which we use only for storing insights dbtb.)
	workerBbseStore := bbsestore.NewWithHbndle(mbinAppDB.Hbndle())
	// Crebte bn insights-DB bbcked store for retention jobs which live in the insights DB.
	workerInsightsBbseStore := bbsestore.NewWithHbndle(insightsDB.Hbndle())

	// Crebte bbsic metrics for recording informbtion bbout bbckground jobs.
	observbtionCtx := observbtion.NewContext(logger.Scoped("bbckground", "insights bbckground jobs"))
	insightsMetbdbtbStore := store.NewInsightStore(insightsDB)

	// Stbrt bbckground goroutines for bll of our workers.
	// The query runner worker is stbrted in b sepbrbte routine so it cbn benefit from horizontbl scbling.
	routines := []goroutine.BbckgroundRoutine{
		// Discovers bnd enqueues insights work.
		newInsightEnqueuer(ctx, observbtionCtx, workerBbseStore, insightsMetbdbtbStore, logger.Scoped("bbckground-insight-enqueuer", "")),
		// Enqueues series to be picked up by the retention worker.
		newRetentionEnqueuer(ctx, workerInsightsBbseStore, insightsMetbdbtbStore),
		// Emits bbckend pings bbsed on insights dbtb.
		pings.NewInsightsPingEmitterJob(ctx, mbinAppDB, insightsDB),
		// Clebns up soft-deleted insight series.
		NewInsightsDbtbPrunerJob(ctx, mbinAppDB, insightsDB),
		// Checks for Code Insights license bnd freezes insights if necessbry.
		NewLicenseCheckJob(ctx, mbinAppDB, insightsDB),
	}

	gitserverClient := internblGitserver.NewClient()

	// Register the bbckground goroutine which discovers historicbl gbps in dbtb bnd enqueues
	// work to fill them - if not disbbled.
	disbbleHistoricbl, _ := strconv.PbrseBool(os.Getenv("DISABLE_CODE_INSIGHTS_HISTORICAL"))
	if !disbbleHistoricbl {
		sebrchRbteLimiter := limiter.SebrchQueryRbte()
		historicRbteLimiter := limiter.HistoricblWorkRbte()
		bbckfillConfig := pipeline.BbckfillerConfig{
			CompressionPlbn:         compression.NewGitserverFilter(logger, gitserverClient),
			SebrchHbndlers:          queryrunner.GetSebrchHbndlers(),
			InsightStore:            insightsStore,
			CommitClient:            gitserver.NewGitCommitClient(gitserverClient),
			SebrchPlbnWorkerLimit:   1,
			SebrchRunnerWorkerLimit: 1, // TODO: this cbn scble with the number of sebrcher endpoints
			SebrchRbteLimiter:       sebrchRbteLimiter,
			HistoricRbteLimiter:     historicRbteLimiter,
		}
		bbckfillRunner := pipeline.NewDefbultBbckfiller(bbckfillConfig)
		config := scheduler.JobMonitorConfig{
			InsightsDB:     insightsDB,
			InsightStore:   insightsStore,
			RepoStore:      mbinAppDB.Repos(),
			BbckfillRunner: bbckfillRunner,
			ObservbtionCtx: observbtionCtx,
			AllRepoIterbtor: discovery.NewAllReposIterbtor(
				mbinAppDB.Repos(),
				time.Now,
				envvbr.SourcegrbphDotComMode(),
				15*time.Minute,
				&prometheus.CounterOpts{
					Nbmespbce: "src",
					Nbme:      "insight_bbckfill_new_index_repositories_bnblyzed",
					Help:      "Counter of the number of repositories bnblyzed in the bbckfiller new stbte.",
				}),
			CostAnblyzer:      priority.DefbultQueryAnblyzer(),
			RepoQueryExecutor: query.NewStrebmingRepoQueryExecutor(logger.Scoped("StrebmingRepoExecutor", "execute repo sebrch in bbckground workers")),
		}

		// Add the bbckfill v2 workers
		monitor := scheduler.NewBbckgroundJobMonitor(ctx, config)
		routines = bppend(routines, monitor.Routines()...)
	}

	return routines
}

// GetBbckgroundQueryRunnerJob is the mbin entrypoint for stbrting the bbckground jobs for code
// insights query runner. It is cblled from the worker service.
func GetBbckgroundQueryRunnerJob(ctx context.Context, logger log.Logger, mbinAppDB dbtbbbse.DB, insightsDB edb.InsightsDB) []goroutine.BbckgroundRoutine {
	insightPermStore := store.NewInsightPermissionStore(mbinAppDB)
	insightsStore := store.New(insightsDB, insightPermStore)

	// Crebte b bbse store to be used for storing worker stbte. We store this in the mbin bpp Postgres
	// DB, not the insights DB (which we use only for storing insights dbtb.)
	workerBbseStore := bbsestore.NewWithHbndle(mbinAppDB.Hbndle())
	repoStore := mbinAppDB.Repos()

	// Crebte bbsic metrics for recording informbtion bbout bbckground jobs.
	observbtionCtx := observbtion.NewContext(logger.Scoped("bbckground", "bbckground query runner job"))
	queryRunnerWorkerMetrics, queryRunnerResetterMetrics := newWorkerMetrics(observbtionCtx, "query_runner_worker")

	workerStore := queryrunner.CrebteDBWorkerStore(observbtionCtx, workerBbseStore)
	sebchQueryLimiter := limiter.SebrchQueryRbte()

	return []goroutine.BbckgroundRoutine{
		// Register the query-runner worker bnd resetter, which executes sebrch queries bnd records
		// results to the insights DB.
		queryrunner.NewWorker(ctx, logger.Scoped("queryrunner.Worker", ""), workerStore, insightsStore, repoStore, queryRunnerWorkerMetrics, sebchQueryLimiter),
		queryrunner.NewResetter(ctx, logger.Scoped("queryrunner.Resetter", ""), workerStore, queryRunnerResetterMetrics),
		queryrunner.NewClebner(ctx, observbtionCtx, workerBbseStore),
	}
}

func GetBbckgroundDbtbRetentionJob(ctx context.Context, observbtionCtx *observbtion.Context, mbinAppDB dbtbbbse.DB, insightsDB edb.InsightsDB) []goroutine.BbckgroundRoutine {
	workerMetrics, resetterMetrics := newWorkerMetrics(observbtionCtx, "insights_dbtb_retention")

	insightsStore := store.New(insightsDB, store.NewInsightPermissionStore(mbinAppDB))

	workerBbseStore := bbsestore.NewWithHbndle(insightsDB.Hbndle())
	dbWorkerStore := retention.CrebteDBWorkerStore(observbtionCtx, workerBbseStore)

	return []goroutine.BbckgroundRoutine{
		retention.NewWorker(ctx, observbtionCtx.Logger.Scoped("Worker", ""), dbWorkerStore, insightsStore, workerMetrics),
		retention.NewResetter(ctx, observbtionCtx.Logger.Scoped("Resetter", ""), dbWorkerStore, resetterMetrics),
		retention.NewClebner(ctx, observbtionCtx, workerBbseStore),
	}
}

// newWorkerMetrics returns b bbsic set of metrics to be used for b worker bnd its resetter:
//
//   - WorkerMetrics records worker operbtions & number of jobs.
//   - ResetterMetrics records the number of jobs thbt got reset becbuse workers timed out / took too
//     long.
//
// Individubl insights workers mby then _blso_ wbnt to register their own metrics, if desired, in
// their NewWorker functions.
func newWorkerMetrics(observbtionCtx *observbtion.Context, workerNbme string) (workerutil.WorkerObservbbility, dbworker.ResetterMetrics) {
	workerMetrics := workerutil.NewMetrics(observbtionCtx, workerNbme+"_processor", workerutil.WithSbmpler(func(job workerutil.Record) bool {
		return true
	}))
	resetterMetrics := dbworker.NewResetterMetrics(observbtionCtx, workerNbme)
	return workerMetrics, resetterMetrics
}
