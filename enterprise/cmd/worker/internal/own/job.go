package own

import (
	"context"
	"fmt"

	"github.com/keegancsmith/sqlf"

	logger "github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own/background"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ownRepoIndexingQueue struct{}

func NewOwnRepoIndexingQueue() job.Job {
	return &ownRepoIndexingQueue{}
}

func (o *ownRepoIndexingQueue) Description() string {
	return "Queue used by Sourcegraph Own to index ownership data partitioned per repository"
}

func (o *ownRepoIndexingQueue) Config() []env.Config {
	return nil
}

func (o *ownRepoIndexingQueue) Routines(startupCtx context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	var routines []goroutine.BackgroundRoutine
	routines = append(routines, background.NewOwnBackgroundWorker(context.Background(), db, observationCtx)...)
	for _, jobType := range background.IndexJobTypes {
		routines = append(routines, goroutine.NewPeriodicGoroutine(context.Background(), jobType.Name, "", jobType.RefreshInterval, &ownRepoIndexSchedulerJob{store: basestore.NewWithHandle(db.Handle()), jobType: jobType, observationCtx: observationCtx}))
	}

	return routines, nil
}

type ownRepoIndexSchedulerJob struct {
	store          *basestore.Store
	jobType        background.IndexJobType
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
