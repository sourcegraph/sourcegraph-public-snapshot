package background

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/sourcegraph/log"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	edb "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	internalGitserver "github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/insights/background/limiter"
	"github.com/sourcegraph/sourcegraph/internal/insights/background/pings"
	"github.com/sourcegraph/sourcegraph/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/internal/insights/background/retention"
	"github.com/sourcegraph/sourcegraph/internal/insights/compression"
	"github.com/sourcegraph/sourcegraph/internal/insights/discovery"
	"github.com/sourcegraph/sourcegraph/internal/insights/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/insights/pipeline"
	"github.com/sourcegraph/sourcegraph/internal/insights/priority"
	"github.com/sourcegraph/sourcegraph/internal/insights/query"
	"github.com/sourcegraph/sourcegraph/internal/insights/scheduler"
	"github.com/sourcegraph/sourcegraph/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

type RepoStore interface {
	GetByName(ctx context.Context, name api.RepoName) (*types.Repo, error)
}

// GetBackgroundJobs is the main entrypoint which starts background jobs for code insights. It is
// called from the worker service.
func GetBackgroundJobs(ctx context.Context, logger log.Logger, mainAppDB database.DB, insightsDB edb.InsightsDB) []goroutine.BackgroundRoutine {
	insightPermStore := store.NewInsightPermissionStore(mainAppDB)
	insightsStore := store.New(insightsDB, insightPermStore)

	// Create a base store to be used for storing worker state. We store this in the main app Postgres
	// DB, not the insights DB (which we use only for storing insights data.)
	workerBaseStore := basestore.NewWithHandle(mainAppDB.Handle())
	// Create an insights-DB backed store for retention jobs which live in the insights DB.
	workerInsightsBaseStore := basestore.NewWithHandle(insightsDB.Handle())

	// Create basic metrics for recording information about background jobs.
	observationCtx := observation.NewContext(logger.Scoped("background"))
	insightsMetadataStore := store.NewInsightStore(insightsDB)

	// Start background goroutines for all of our workers.
	// The query runner worker is started in a separate routine so it can benefit from horizontal scaling.
	routines := []goroutine.BackgroundRoutine{
		// Discovers and enqueues insights work.
		newInsightEnqueuer(ctx, observationCtx, workerBaseStore, insightsMetadataStore, logger.Scoped("background-insight-enqueuer")),
		// Enqueues series to be picked up by the retention worker.
		newRetentionEnqueuer(ctx, workerInsightsBaseStore, insightsMetadataStore),
		// Emits backend pings based on insights data.
		pings.NewInsightsPingEmitterJob(ctx, mainAppDB, insightsDB),
		// Cleans up soft-deleted insight series.
		NewInsightsDataPrunerJob(ctx, mainAppDB, insightsDB),
		// Checks for Code Insights license and freezes insights if necessary.
		NewLicenseCheckJob(ctx, mainAppDB, insightsDB),
	}

	gitserverClient := internalGitserver.NewClient("insights")

	// Register the background goroutine which discovers historical gaps in data and enqueues
	// work to fill them - if not disabled.
	disableHistorical, _ := strconv.ParseBool(os.Getenv("DISABLE_CODE_INSIGHTS_HISTORICAL"))
	if !disableHistorical {
		searchRateLimiter := limiter.SearchQueryRate()
		historicRateLimiter := limiter.HistoricalWorkRate()
		backfillConfig := pipeline.BackfillerConfig{
			CompressionPlan:         compression.NewGitserverFilter(logger, gitserverClient.Scoped("compressionfilter")),
			SearchHandlers:          queryrunner.GetSearchHandlers(),
			InsightStore:            insightsStore,
			CommitClient:            gitserver.NewGitCommitClient(gitserverClient.Scoped("commitclient")),
			SearchPlanWorkerLimit:   1,
			SearchRunnerWorkerLimit: 1, // TODO: this can scale with the number of searcher endpoints
			SearchRateLimiter:       searchRateLimiter,
			HistoricRateLimiter:     historicRateLimiter,
		}
		backfillRunner := pipeline.NewDefaultBackfiller(backfillConfig)
		config := scheduler.JobMonitorConfig{
			InsightsDB:     insightsDB,
			InsightStore:   insightsStore,
			RepoStore:      mainAppDB.Repos(),
			BackfillRunner: backfillRunner,
			ObservationCtx: observationCtx,
			AllRepoIterator: discovery.NewAllReposIterator(
				mainAppDB.Repos(),
				time.Now,
				15*time.Minute,
				&prometheus.CounterOpts{
					Namespace: "src",
					Name:      "insight_backfill_new_index_repositories_analyzed",
					Help:      "Counter of the number of repositories analyzed in the backfiller new state.",
				}),
			CostAnalyzer:      priority.DefaultQueryAnalyzer(),
			RepoQueryExecutor: query.NewStreamingRepoQueryExecutor(logger.Scoped("StreamingRepoExecutor")),
		}

		// Add the backfill v2 workers
		monitor := scheduler.NewBackgroundJobMonitor(ctx, config)
		routines = append(routines, monitor.Routines()...)
	}

	return routines
}

// GetBackgroundQueryRunnerJob is the main entrypoint for starting the background jobs for code
// insights query runner. It is called from the worker service.
func GetBackgroundQueryRunnerJob(ctx context.Context, logger log.Logger, mainAppDB database.DB, insightsDB edb.InsightsDB) []goroutine.BackgroundRoutine {
	insightPermStore := store.NewInsightPermissionStore(mainAppDB)
	insightsStore := store.New(insightsDB, insightPermStore)

	// Create a base store to be used for storing worker state. We store this in the main app Postgres
	// DB, not the insights DB (which we use only for storing insights data.)
	workerBaseStore := basestore.NewWithHandle(mainAppDB.Handle())
	repoStore := mainAppDB.Repos()

	// Create basic metrics for recording information about background jobs.
	observationCtx := observation.NewContext(logger.Scoped("background"))
	queryRunnerWorkerMetrics, queryRunnerResetterMetrics := newWorkerMetrics(observationCtx, "query_runner_worker")

	workerStore := queryrunner.CreateDBWorkerStore(observationCtx, workerBaseStore)
	searchQueryLimiter := limiter.SearchQueryRate()

	return []goroutine.BackgroundRoutine{
		// Register the query-runner worker and resetter, which executes search queries and records
		// results to the insights DB.
		queryrunner.NewWorker(ctx, logger.Scoped("queryrunner.Worker"), workerStore, insightsStore, repoStore, queryRunnerWorkerMetrics, searchQueryLimiter),
		queryrunner.NewResetter(ctx, logger.Scoped("queryrunner.Resetter"), workerStore, queryRunnerResetterMetrics),
		queryrunner.NewCleaner(ctx, observationCtx, workerBaseStore),
	}
}

func GetBackgroundDataRetentionJob(ctx context.Context, observationCtx *observation.Context, mainAppDB database.DB, insightsDB edb.InsightsDB) []goroutine.BackgroundRoutine {
	workerMetrics, resetterMetrics := newWorkerMetrics(observationCtx, "insights_data_retention")

	insightsStore := store.New(insightsDB, store.NewInsightPermissionStore(mainAppDB))

	workerBaseStore := basestore.NewWithHandle(insightsDB.Handle())
	dbWorkerStore := retention.CreateDBWorkerStore(observationCtx, workerBaseStore)

	return []goroutine.BackgroundRoutine{
		retention.NewWorker(ctx, observationCtx.Logger.Scoped("Worker"), dbWorkerStore, insightsStore, workerMetrics),
		retention.NewResetter(ctx, observationCtx.Logger.Scoped("Resetter"), dbWorkerStore, resetterMetrics),
		retention.NewCleaner(ctx, observationCtx, workerBaseStore),
	}
}

// newWorkerMetrics returns a basic set of metrics to be used for a worker and its resetter:
//
//   - WorkerMetrics records worker operations & number of jobs.
//   - ResetterMetrics records the number of jobs that got reset because workers timed out / took too
//     long.
//
// Individual insights workers may then _also_ want to register their own metrics, if desired, in
// their NewWorker functions.
func newWorkerMetrics(observationCtx *observation.Context, workerName string) (workerutil.WorkerObservability, dbworker.ResetterMetrics) {
	workerMetrics := workerutil.NewMetrics(observationCtx, workerName+"_processor", workerutil.WithSampler(func(job workerutil.Record) bool {
		return true
	}))
	resetterMetrics := dbworker.NewResetterMetrics(observationCtx, workerName)
	return workerMetrics, resetterMetrics
}
