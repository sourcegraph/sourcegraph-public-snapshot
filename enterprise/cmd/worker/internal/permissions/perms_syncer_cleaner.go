package permissions

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ job.Job = (*permissionSyncJobCleaner)(nil)

// permissionSyncJobCleaner is a worker responsible for cleaning up processed
// permission sync jobs.
type permissionSyncJobCleaner struct{}

func (p *permissionSyncJobCleaner) Description() string {
	return "Cleans up completed or failed permissions sync jobs"
}

func (p *permissionSyncJobCleaner) Config() []env.Config {
	return nil
}

const defaultCleanupInterval = time.Minute

var cleanupInterval = defaultCleanupInterval

var watchConfOnce = sync.Once{}

func loadCleanupIntervalFromConf() {
	seconds := conf.Get().PermissionsSyncJobCleanupInterval
	if seconds <= 0 {
		cleanupInterval = defaultCleanupInterval
	} else {
		cleanupInterval = time.Duration(seconds) * time.Second
	}
}

func (p *permissionSyncJobCleaner) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, errors.Wrap(err, "init DB")
	}

	m := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"permission_sync_job_worker_cleaner",
		metrics.WithCountHelp("Total number of permissions syncer cleaner executions."),
	)
	operation := observationCtx.Operation(observation.Op{
		Name:    "PermissionsSyncer.Cleaner.Run",
		Metrics: m,
	})

	watchConfOnce.Do(func() {
		conf.Watch(loadCleanupIntervalFromConf)
	})

	return []goroutine.BackgroundRoutine{
		goroutine.NewPeriodicGoroutine(
			context.Background(),
			goroutine.HandlerFunc(
				func(ctx context.Context) error {
					start := time.Now()
					cleanedJobs, err := cleanJobs(ctx, db)
					m.Observe(time.Since(start).Seconds(), float64(cleanedJobs), &err)
					return err
				},
			),
			goroutine.WithName("auth.permission_sync_job_cleaner"),
			goroutine.WithDescription(p.Description()),
			goroutine.WithIntervalFunc(func() time.Duration { return cleanupInterval }),
			goroutine.WithOperation(operation),
		),
	}, nil
}

func NewPermissionSyncJobCleaner() job.Job {
	return &permissionSyncJobCleaner{}
}

// cleanJobs runs an SQL query which finds and deletes all non-queued/processing
// permission sync jobs of users/repos which number exceeds `jobsToKeep`.
func cleanJobs(ctx context.Context, store database.DB) (int64, error) {
	jobsToKeep := 5
	if conf.Get().PermissionsSyncJobsHistorySize != nil {
		jobsToKeep = *conf.Get().PermissionsSyncJobsHistorySize
	}

	result, err := store.ExecContext(ctx, fmt.Sprintf(cleanJobsFmtStr, jobsToKeep))
	if err != nil {
		return 0, err
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return deleted, err
}

const cleanJobsFmtStr = `
-- CTE for fetching queued/processing jobs per repository_id/user_id and their row numbers

WITH job_history AS (
	SELECT id, repository_id, user_id, ROW_NUMBER() OVER (
		PARTITION BY repository_id, user_id
		ORDER BY finished_at DESC NULLS LAST
	) FROM permission_sync_jobs
	WHERE state NOT IN ('queued', 'processing')
)

-- Removing those jobs which count per repo/user exceeds a certain number

DELETE FROM permission_sync_jobs
WHERE id IN (
	SELECT id
	FROM job_history
	WHERE row_number > %d
)
`
