package background

import (
	"context"
	"fmt"
	"time"

	"github.com/derision-test/glock"
	"github.com/keegancsmith/sqlf"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/own/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type IndexJobType struct {
	Name            string
	IndexInterval   time.Duration
	RefreshInterval time.Duration
}

// QueuePerRepoIndexJobs is a slice of jobs that will automatically initialize and will queue up one index job per repo every IndexInterval.
var QueuePerRepoIndexJobs = []IndexJobType{
	{
		Name:            types.SignalRecentContributors,
		IndexInterval:   time.Hour * 24,
		RefreshInterval: time.Minute * 5,
	}, {
		Name:            types.Analytics,
		IndexInterval:   time.Hour * 24,
		RefreshInterval: time.Hour * 24,
	},
}

var repoCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "src",
	Name:      "own_background_index_scheduler_repos_queued_total",
	Help:      "Number of repositories queued for indexing in Sourcegraph Own",
}, []string{"op"})

func GetOwnIndexSchedulerRoutines(db database.DB, observationCtx *observation.Context) (routines []goroutine.BackgroundRoutine) {
	redMetrics := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"own_background_index_scheduler",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(jobType IndexJobType) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("own.background.index.scheduler.%s", jobType.Name),
			MetricLabelValues: []string{jobType.Name},
			Metrics:           redMetrics,
		})
	}

	makeRoutine := func(jobType IndexJobType, op *observation.Operation, handler goroutine.Handler) goroutine.BackgroundRoutine {
		return goroutine.NewPeriodicGoroutine(
			context.Background(),
			newFeatureFlagWrapper(db, jobType, op, handler),
			goroutine.WithName(jobType.Name),
			goroutine.WithDescription(""),
			goroutine.WithInterval(jobType.RefreshInterval),
			goroutine.WithOperation(op),
		)
	}

	for _, jobType := range QueuePerRepoIndexJobs {
		operation := op(jobType)
		routines = append(routines, makeRoutine(jobType, operation, newOwnRepoIndexSchedulerJob(db, jobType, operation.Logger)))
	}

	recent := IndexJobType{
		Name:            types.SignalRecentViews,
		RefreshInterval: time.Minute * 5,
	}
	routines = append(routines, makeRoutine(recent, op(recent), newRecentViewsIndexer(db, observationCtx.Logger)))

	return routines
}

type featureFlagWrapper struct {
	jobType IndexJobType
	logger  logger.Logger
	db      database.DB
	handler goroutine.Handler
}

func newFeatureFlagWrapper(db database.DB, jobType IndexJobType, op *observation.Operation, handler goroutine.Handler) *featureFlagWrapper {
	return &featureFlagWrapper{
		jobType: jobType,
		logger:  op.Logger,
		db:      db,
		handler: handler,
	}
}

func (f *featureFlagWrapper) Handle(ctx context.Context) error {
	logJobDisabled := func() {
		f.logger.Info("skipping own indexing job, job disabled", logger.String("job-name", f.jobType.Name))
	}

	config, err := loadConfig(ctx, f.jobType, f.db.OwnSignalConfigurations())
	if err != nil {
		return errors.Wrap(err, "loadConfig")
	}

	if !config.Enabled {
		logJobDisabled()
		return nil
	}
	// okay, so the job is enabled - proceed!
	f.logger.Info("Scheduling repo indexes for own job", logger.String("job-name", f.jobType.Name))
	return f.handler.Handle(ctx)
}

type ownRepoIndexSchedulerJob struct {
	store       *basestore.Store
	jobType     IndexJobType
	logger      logger.Logger
	clock       glock.Clock
	configStore database.SignalConfigurationStore
}

func newOwnRepoIndexSchedulerJob(db database.DB, jobType IndexJobType, logger logger.Logger) *ownRepoIndexSchedulerJob {
	store := basestore.NewWithHandle(db.Handle())
	return &ownRepoIndexSchedulerJob{jobType: jobType, store: store, logger: logger, clock: glock.NewRealClock(), configStore: db.OwnSignalConfigurations()}
}

func (o *ownRepoIndexSchedulerJob) Handle(ctx context.Context) error {
	// convert duration to hours to match the query
	after := o.clock.Now().Add(-1 * o.jobType.IndexInterval)

	query := sqlf.Sprintf(ownIndexRepoQuery, o.jobType.Name, after)
	val, err := o.store.ExecResult(ctx, query)
	if err != nil {
		return errors.Wrapf(err, "ownRepoIndexSchedulerJob.Handle %s", o.jobType.Name)
	}

	rows, _ := val.RowsAffected()
	o.logger.Info("Own index job scheduled", logger.String("job-name", o.jobType.Name), logger.Int64("row-count", rows))
	repoCounter.WithLabelValues(o.jobType.Name).Add(float64(rows))
	return nil
}

// Every X duration the scheduler will run and try to index repos for each job type. It will obey the following rules:
//  1. ignore jobs in progress, queued, or still in retry-backoff
//  2. ignore repos that have indexed more recently than the configured index interval for the job, ex. 24 hours
//     OR repos that are excluded from the signal configuration. All exclusions are pulled into the ineligible_repos CTE.
//  3. add all remaining cloned repos to the queue
//
// This means each (job, repo) tuple will only be index maximum once in a single interval duration
var ownIndexRepoQuery = `
WITH signal_config AS (SELECT * FROM own_signal_configurations WHERE name = %s LIMIT 1),
     ineligible_repos AS (SELECT repo_id
                          FROM own_background_jobs,
                               signal_config
                          WHERE job_type = signal_config.id
                              AND (state IN ('failed', 'completed') AND finished_at > %s) OR (state IN ('processing', 'errored', 'queued'))
                          UNION
                            SELECT repo.id FROM repo, signal_config WHERE repo.name ~~ ANY(signal_config.excluded_repo_patterns))
INSERT
INTO own_background_jobs (repo_id, job_type) (SELECT gr.repo_id, signal_config.id
                                              FROM gitserver_repos gr,
                                                   signal_config
                                              WHERE gr.repo_id NOT IN (SELECT * FROM ineligible_repos)
                                                AND gr.clone_status = 'cloned');`

func loadConfig(ctx context.Context, jobType IndexJobType, store database.SignalConfigurationStore) (database.SignalConfiguration, error) {
	configurations, err := store.LoadConfigurations(ctx, database.LoadSignalConfigurationArgs{Name: jobType.Name})
	if err != nil {
		return database.SignalConfiguration{}, errors.Wrap(err, "LoadConfigurations")
	} else if len(configurations) == 0 {
		return database.SignalConfiguration{}, errors.Newf("ownership signal configuration not found for name: %s\n", jobType.Name)
	}
	return configurations[0], nil
}
