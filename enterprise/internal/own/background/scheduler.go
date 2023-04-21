package background

import (
	"context"
	"fmt"

	"github.com/keegancsmith/sqlf"

	logger "github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func GetOwnIndexSchedulerRoutines(db database.DB, observationCtx *observation.Context) (routines []goroutine.BackgroundRoutine) {
	redMetrics := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"own_background_index_scheduler",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("own.background.index.scheduler.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           redMetrics,
		})
	}

	for _, jobType := range IndexJobTypes {
		operation := op(jobType.Name)
		routines = append(routines, goroutine.NewPeriodicGoroutineWithMetrics(context.Background(), jobType.Name, "", jobType.RefreshInterval, &ownRepoIndexSchedulerJob{store: basestore.NewWithHandle(db.Handle()), jobType: jobType, observationCtx: observationCtx}, operation))
	}
	return routines
}

type ownRepoIndexSchedulerJob struct {
	store          *basestore.Store
	jobType        IndexJobType
	observationCtx *observation.Context
}

func (o *ownRepoIndexSchedulerJob) Handle(ctx context.Context) error {
	lgr := o.observationCtx.Logger
	lgr.Info("Scheduling repo indexes for own job", logger.String("job-name", o.jobType.Name))

	// convert duration to hours to match the query
	hours := int(o.jobType.IndexInterval.Hours())

	query := sqlf.Sprintf(fmt.Sprintf(ownIndexRepoQuery, o.jobType.Id, hours, o.jobType.Id))
	val, err := o.store.ExecResult(ctx, query)
	if err != nil {
		return errors.Wrapf(err, "ownRepoIndexSchedulerJob.Handle %s", o.jobType.Name)
	}

	rows, _ := val.RowsAffected()
	lgr.Info("Own index job scheduled", logger.String("job-name", o.jobType.Name), logger.Int64("row-count", rows))
	return nil
}

// Every X duration the scheduler will run and try to index repos for each job type. It will obey the following rules:
// 1. ignore jobs in progress or still in retry-backoff
// 2. ignore repos that have indexed more recently than the configured index interval for the job, ex. 24 hours
// 3. add all remaining cloned repos to the queue
// This means each (job, repo) tuple will only be index maximum once in a single interval duration
var ownIndexRepoQuery = `
WITH ineligible_repos AS (SELECT repo_id
                          FROM own_background_jobs
                          WHERE job_type = %d
                              AND (state IN ('failed', 'completed') AND finished_at > NOW() - INTERVAL '%d hour')
                             OR (state IN ('processing', 'errored')))
insert into own_background_jobs (repo_id, job_type) (SELECT gr.repo_id, %d
FROM gitserver_repos gr
WHERE gr.repo_id NOT IN (SELECT * FROM ineligible_repos) and gr.clone_status = 'cloned');
`
