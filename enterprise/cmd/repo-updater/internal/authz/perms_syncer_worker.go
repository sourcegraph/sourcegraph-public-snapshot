package authz

import (
	"context"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/group"
)

func MakePermsSyncerWorker(ctx context.Context, observationCtx *observation.Context, syncer permsSyncer) *permsSyncerWorker {
	syncGroups := map[requestType]group.ContextGroup{
		requestTypeUser: group.New().WithContext(ctx).WithMaxConcurrency(syncUsersMaxConcurrency()),
		requestTypeRepo: group.New().WithContext(ctx).WithMaxConcurrency(1),
	}

	return &permsSyncerWorker{
		logger:     observationCtx.Logger.Scoped("PermsSyncerWorker", "Permission sync worker"),
		syncer:     syncer,
		syncGroups: syncGroups,
	}
}

type permsSyncer interface {
	syncPerms(ctx context.Context, syncGroups map[requestType]group.ContextGroup, request *syncRequest)
}

type permsSyncerWorker struct {
	logger     log.Logger
	syncer     permsSyncer
	syncGroups map[requestType]group.ContextGroup
}

func (h *permsSyncerWorker) Handle(_ context.Context, _ log.Logger, record *database.PermissionSyncJob) error {
	prio := priorityLow
	if record.HighPriority {
		prio = priorityHigh
	}

	reqType := requestTypeUser
	reqID := int32(record.UserID)
	if record.RepositoryID != 0 {
		reqType = requestTypeRepo
		reqID = int32(record.RepositoryID)
	}

	h.logger.Info(
		"Handling permission sync job",
		log.String("type", reqType.String()),
		log.Int32("id", reqID),
		log.String("priority", prio.String()),
	)

	// We use a background context here because right now syncPerms is an async operation.
	//
	// Later we can change the max concurrency on the worker though instead of using
	// the concurrency groups
	syncCtx := actor.WithInternalActor(context.Background())
	h.syncer.syncPerms(syncCtx, h.syncGroups, &syncRequest{requestMeta: &requestMeta{
		Priority: prio,
		Type:     reqType,
		ID:       reqID,
		Options: authz.FetchPermsOptions{
			InvalidateCaches: record.InvalidateCaches,
		},
		// TODO(sashaostrikov): Fill this out
		NoPerms: false,
	}})

	return nil
}

func MakeStore(observationCtx *observation.Context, dbHandle basestore.TransactableHandle) dbworkerstore.Store[*database.PermissionSyncJob] {
	return dbworkerstore.New(observationCtx, dbHandle, dbworkerstore.Options[*database.PermissionSyncJob]{
		Name:              "permission_sync_job_worker_store",
		TableName:         "permission_sync_jobs",
		ColumnExpressions: database.PermissionSyncJobColumns,
		Scan:              dbworkerstore.BuildWorkerScan(database.ScanPermissionSyncJob),
		// TODO(sashaostrikov): We need to take `NextSyncAt`/`process_after` into account
		OrderByExpression: sqlf.Sprintf("permission_sync_jobs.high_priority, permission_sync_jobs.repository_id, permission_sync_jobs.user_id"),
		MaxNumResets:      5,
		StalledMaxAge:     time.Second * 30,
	})
}

func MakeWorker(ctx context.Context, observationCtx *observation.Context, workerStore dbworkerstore.Store[*database.PermissionSyncJob], permsSyncer *PermsSyncer) *workerutil.Worker[*database.PermissionSyncJob] {
	handler := MakePermsSyncerWorker(ctx, observationCtx, permsSyncer)

	return dbworker.NewWorker[*database.PermissionSyncJob](ctx, workerStore, handler, workerutil.WorkerOptions{
		Name:              "permission_sync_job_worker",
		Interval:          time.Second, // Poll for a job once per second
		HeartbeatInterval: 10 * time.Second,
		Metrics:           workerutil.NewMetrics(observationCtx, "permission_sync_job_worker"),

		// Process only one job at a time (per instance).
		// TODO(sashaostrikov): This should be changed once the handler above is not async anymore.
		NumHandlers: 1,
	})
}

func MakeResetter(observationCtx *observation.Context, workerStore dbworkerstore.Store[*database.PermissionSyncJob]) *dbworker.Resetter[*database.PermissionSyncJob] {
	return dbworker.NewResetter(observationCtx.Logger, workerStore, dbworker.ResetterOptions{
		Name:     "permission_sync_job_worker_resetter",
		Interval: time.Second * 30, // Check for orphaned jobs every 30 seconds
		Metrics:  dbworker.NewResetterMetrics(observationCtx, "permission_sync_job_worker"),
	})
}

// MakeCleaner returns a background goroutine which will periodically find and
// remove permission sync jobs of repos/users which exceed the history preserving
// threshold in site config (PermissionsSyncJobsHistorySize).
func MakeCleaner(ctx context.Context, observationCtx *observation.Context, db database.DB) goroutine.BackgroundRoutine {
	m := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"permission_sync_job_worker_cleaner",
		metrics.WithCountHelp("Total number of permissions syncer cleaner executions."),
	)
	operation := observationCtx.Operation(observation.Op{
		Name:    "PermissionsSyncer.Cleaner.Run",
		Metrics: m,
	})
	jobsToKeep := 5
	if conf.Get().PermissionsSyncJobsHistorySize != nil {
		jobsToKeep = *conf.Get().PermissionsSyncJobsHistorySize
	}

	return goroutine.NewPeriodicGoroutineWithMetrics(
		ctx, "authz.permission_sync_job_worker_cleaner", "removes completed or failed permissions sync jobs",
		1*time.Hour, goroutine.HandlerFunc(
			func(ctx context.Context) error {
				start := time.Now()
				cleanedJobs, err := cleanJobs(ctx, db, jobsToKeep)
				m.Observe(time.Since(start).Seconds(), cleanedJobs, &err)
				return err
			},
		), operation,
	)
}

// cleanJobs runs an SQL query which finds and deletes all non-queued/processing
// permission sync jobs of users/repos which number exceeds `jobsToKeep`.
func cleanJobs(ctx context.Context, store database.DB, jobsToKeep int) (float64, error) {
	result, err := store.ExecContext(
		ctx,
		fmt.Sprintf(cleanJobsFmtStr, jobsToKeep),
	)
	if err != nil {
		return 0, err
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return float64(deleted), err
}

const cleanJobsFmtStr = `
-- CTE for fetching queued/processing jobs per repository_id/user_id and their row numbers

WITH job_history AS (
	SELECT id, repository_id, user_id, ROW_NUMBER() OVER (
		PARTITION BY repository_id, user_id
		ORDER BY id
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
