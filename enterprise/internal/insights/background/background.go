package background

import (
	"context"
	"database/sql"
	"os"
	"strconv"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/compression"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

// GetBackgroundJobs is the main entrypoint which starts background jobs for code insights. It is
// called from the worker service.
func GetBackgroundJobs(ctx context.Context, mainAppDB *sql.DB, insightsDB *sql.DB) []goroutine.BackgroundRoutine {
	insightPermStore := store.NewInsightPermissionStore(mainAppDB)
	insightsStore := store.New(insightsDB, insightPermStore)

	// Create a base store to be used for storing worker state. We store this in the main app Postgres
	// DB, not the TimescaleDB (which we use only for storing insights data.)
	workerBaseStore := basestore.NewWithDB(mainAppDB, sql.TxOptions{})

	// Create basic metrics for recording information about background jobs.
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}
	queryRunnerWorkerMetrics, queryRunnerResetterMetrics := newWorkerMetrics(observationContext, "insights_search_queue")

	insightsMetadataStore := store.NewInsightStore(insightsDB)

	workerStore := queryrunner.CreateDBWorkerStore(workerBaseStore, observationContext)

	// Start background goroutines for all of our workers.
	routines := []goroutine.BackgroundRoutine{
		// Register the background goroutine which discovers and enqueues insights work.
		newInsightEnqueuer(ctx, workerBaseStore, insightsMetadataStore, observationContext),

		// Register the query-runner worker and resetter, which executes search queries and records
		// results to TimescaleDB.
		queryrunner.NewWorker(ctx, workerStore, insightsStore, queryRunnerWorkerMetrics),
		queryrunner.NewResetter(ctx, workerStore, queryRunnerResetterMetrics),
		// disabling the cleaner job while we debug mismatched results from historical insights
		queryrunner.NewCleaner(ctx, workerBaseStore, observationContext),

		// TODO(slimsag): future: register another worker here for webhook querying.
	}

	// todo(insights) add setting to disable this indexer
	routines = append(routines, compression.NewCommitIndexerWorker(ctx, mainAppDB, insightsDB))

	// Register the background goroutine which discovers historical gaps in data and enqueues
	// work to fill them - if not disabled.
	disableHistorical, _ := strconv.ParseBool(os.Getenv("DISABLE_CODE_INSIGHTS_HISTORICAL"))
	if !disableHistorical {
		routines = append(routines, newInsightHistoricalEnqueuer(ctx, workerBaseStore, insightsMetadataStore, insightsStore, observationContext))
	}

	routines = append(routines, discovery.NewMigrateSettingInsightsJob(ctx, mainAppDB, insightsDB))

	return routines
}

// newWorkerMetrics returns a basic set of metrics to be used for a worker and its resetter:
//
// * WorkerMetrics records worker operations & number of jobs.
// * ResetterMetrics records the number of jobs that got reset because workers timed out / took too
//   long.
//
// Individual insights workers may then _also_ want to register their own metrics, if desired, in
// their NewWorker functions.
func newWorkerMetrics(observationContext *observation.Context, workerName string) (workerutil.WorkerMetrics, dbworker.ResetterMetrics) {
	workerMetrics := workerutil.NewMetrics(observationContext, workerName+"_processor", nil)
	resetterMetrics := dbworker.NewMetrics(observationContext, workerName)
	return workerMetrics, *resetterMetrics
}
