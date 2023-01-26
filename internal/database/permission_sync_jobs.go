package database

import (
	"context"
	"strconv"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

const CancellationReasonHigherPriority = "A job with higher priority was added."

type PermissionSyncJobPriority int

const (
	LowPriorityPermissionSync    PermissionSyncJobPriority = 0
	MediumPriorityPermissionSync PermissionSyncJobPriority = 5
	HighPriorityPermissionSync   PermissionSyncJobPriority = 10
)

type PermissionSyncJobOpts struct {
	Priority          PermissionSyncJobPriority
	InvalidateCaches  bool
	ProcessAfter      time.Time
	Reason            string
	TriggeredByUserID int32
}

type PermissionSyncJobStore interface {
	basestore.ShareableStore
	With(other basestore.ShareableStore) PermissionSyncJobStore
	// Transact begins a new transaction and make a new PermissionSyncJobStore over it.
	Transact(ctx context.Context) (PermissionSyncJobStore, error)
	Done(err error) error

	CreateUserSyncJob(ctx context.Context, user int32, opts PermissionSyncJobOpts) error
	CreateRepoSyncJob(ctx context.Context, repo api.RepoID, opts PermissionSyncJobOpts) error

	List(ctx context.Context, opts ListPermissionSyncJobOpts) ([]*PermissionSyncJob, error)
	CancelQueuedJob(ctx context.Context, reason string, id int) error
}

type permissionSyncJobStore struct {
	logger log.Logger
	*basestore.Store
}

var _ PermissionSyncJobStore = (*permissionSyncJobStore)(nil)

func PermissionSyncJobsWith(logger log.Logger, other basestore.ShareableStore) PermissionSyncJobStore {
	return &permissionSyncJobStore{logger: logger, Store: basestore.NewWithHandle(other.Handle())}
}

func (s *permissionSyncJobStore) With(other basestore.ShareableStore) PermissionSyncJobStore {
	return &permissionSyncJobStore{logger: s.logger, Store: s.Store.With(other)}
}

func (s *permissionSyncJobStore) Transact(ctx context.Context) (PermissionSyncJobStore, error) {
	return s.transact(ctx)
}

func (s *permissionSyncJobStore) transact(ctx context.Context) (*permissionSyncJobStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &permissionSyncJobStore{Store: txBase}, err
}

func (s *permissionSyncJobStore) Done(err error) error {
	return s.Store.Done(err)
}

func (s *permissionSyncJobStore) CreateUserSyncJob(ctx context.Context, user int32, opts PermissionSyncJobOpts) error {
	job := &PermissionSyncJob{
		UserID:            int(user),
		Priority:          opts.Priority,
		InvalidateCaches:  opts.InvalidateCaches,
		Reason:            opts.Reason,
		TriggeredByUserID: opts.TriggeredByUserID,
	}
	if !opts.ProcessAfter.IsZero() {
		job.ProcessAfter = opts.ProcessAfter
	}
	return s.createSyncJob(ctx, job)
}

func (s *permissionSyncJobStore) CreateRepoSyncJob(ctx context.Context, repo api.RepoID, opts PermissionSyncJobOpts) error {
	job := &PermissionSyncJob{
		RepositoryID:      int(repo),
		Priority:          opts.Priority,
		InvalidateCaches:  opts.InvalidateCaches,
		Reason:            opts.Reason,
		TriggeredByUserID: opts.TriggeredByUserID,
	}
	if !opts.ProcessAfter.IsZero() {
		job.ProcessAfter = opts.ProcessAfter
	}
	return s.createSyncJob(ctx, job)
}

const permissionSyncJobCreateQueryFmtstr = `
INSERT INTO permission_sync_jobs (
	reason,
	triggered_by_user_id,
	process_after,
	repository_id,
	user_id,
	priority,
	invalidate_caches
)
VALUES (
	%s,
	%s,
	%s,
	%s,
	%s,
	%s,
	%s
)
ON CONFLICT DO NOTHING
RETURNING %s
`

// createSyncJob inserts a postponed (`process_after IS NOT NULL`) sync job right
// away and checks new sync jobs without provided delay for duplicates.
func (s *permissionSyncJobStore) createSyncJob(ctx context.Context, job *PermissionSyncJob) error {
	if job.ProcessAfter.IsZero() {
		// sync jobs without delay are checked for duplicates
		return s.checkDuplicateAndCreateSyncJob(ctx, job)
	}
	return s.create(ctx, job)
}

func (s *permissionSyncJobStore) create(ctx context.Context, job *PermissionSyncJob) error {
	q := sqlf.Sprintf(
		permissionSyncJobCreateQueryFmtstr,
		job.Reason,
		dbutil.NewNullInt32(job.TriggeredByUserID),
		dbutil.NullTimeColumn(job.ProcessAfter),
		dbutil.NewNullInt(job.RepositoryID),
		dbutil.NewNullInt(job.UserID),
		job.Priority,
		job.InvalidateCaches,
		sqlf.Join(PermissionSyncJobColumns, ", "),
	)

	return scanPermissionSyncJob(job, s.QueryRow(ctx, q))
}

// checkDuplicateAndCreateSyncJob adds a new perms sync job with `process_after
// IS NULL` if there is no present duplicate of it.
//
// Duplicates are handled in this way:
//
// 1) If there is no existing job for given user/repo ID in a queued state, we
// insert right away.
//
// 2) If there is an existing job with lower priority, we cancel it and insert a
// new one with higher priority.
//
// 3) If there is an existing job with higher priority, we don't insert new job.
func (s *permissionSyncJobStore) checkDuplicateAndCreateSyncJob(ctx context.Context, job *PermissionSyncJob) (err error) {
	tx, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Done(err)
	}()
	opts := ListPermissionSyncJobOpts{UserID: job.UserID, RepoID: job.RepositoryID, State: "queued", NotCanceled: true, NullProcessAfter: true}
	syncJobs, err := tx.List(ctx, opts)
	if err != nil {
		return err
	}
	// Job doesn't exist -- create it
	if len(syncJobs) == 0 {
		return tx.create(ctx, job)
	}
	// Database constraint guarantees that we have at most 1 job with NULL
	// `process_after` value.
	existingJob := syncJobs[0]

	// Existing job with higher priority should not be overridden. Existing
	// priority job shouldn't be overridden by another same priority job.
	if existingJob.Priority >= job.Priority {
		logField := "repositoryID"
		id := strconv.Itoa(job.RepositoryID)
		if job.RepositoryID == 0 {
			logField = "userID"
			id = strconv.Itoa(job.UserID)
		}
		s.logger.Debug(
			"Permissions sync job is not added because a job with similar or higher priority already exists",
			log.String(logField, id),
		)
		return nil
	}

	err = tx.CancelQueuedJob(ctx, CancellationReasonHigherPriority, existingJob.ID)
	if err != nil && !errcode.IsNotFound(err) {
		return err
	}
	return tx.create(ctx, job)
}

type notFoundError struct{ error }

func (e notFoundError) NotFound() bool { return true }

func (s *permissionSyncJobStore) CancelQueuedJob(ctx context.Context, reason string, id int) error {
	now := timeutil.Now()
	q := sqlf.Sprintf(`
UPDATE permission_sync_jobs
SET cancel = TRUE, state = 'canceled', finished_at = %s, cancellation_reason = %s
WHERE id = %s AND state = 'queued' AND cancel IS FALSE
`, now, reason, id)

	res, err := s.ExecResult(ctx, q)
	if err != nil {
		return err
	}
	af, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if af != 1 {
		return notFoundError{errors.Newf("sync job with id %d not found", id)}
	}
	return nil
}

type ListPermissionSyncJobOpts struct {
	ID                  int
	UserID              int
	RepoID              int
	Reason              string
	State               string
	NullProcessAfter    bool
	NotNullProcessAfter bool
	NotCanceled         bool
}

func (opts ListPermissionSyncJobOpts) sqlConds() []*sqlf.Query {
	conds := []*sqlf.Query{}

	if opts.ID != 0 {
		conds = append(conds, sqlf.Sprintf("id = %s", opts.ID))
	}
	if opts.UserID != 0 {
		conds = append(conds, sqlf.Sprintf("user_id = %s", opts.UserID))
	}
	if opts.RepoID != 0 {
		conds = append(conds, sqlf.Sprintf("repository_id = %s", opts.RepoID))
	}
	if opts.Reason != "" {
		conds = append(conds, sqlf.Sprintf("reason = %s", opts.Reason))
	}
	if opts.State != "" {
		conds = append(conds, sqlf.Sprintf("state = %s", opts.State))
	}
	if opts.NullProcessAfter {
		conds = append(conds, sqlf.Sprintf("process_after IS NULL"))
	}
	if opts.NotNullProcessAfter {
		conds = append(conds, sqlf.Sprintf("process_after IS NOT NULL"))
	}
	if opts.NotCanceled {
		conds = append(conds, sqlf.Sprintf("cancel = false"))
	}
	return conds
}

func (s *permissionSyncJobStore) List(ctx context.Context, opts ListPermissionSyncJobOpts) ([]*PermissionSyncJob, error) {
	conds := opts.sqlConds()

	var whereClause *sqlf.Query
	if len(conds) != 0 {
		whereClause = sqlf.Sprintf("WHERE %s", sqlf.Join(conds, "\n AND "))
	} else {
		whereClause = sqlf.Sprintf("")
	}

	q := sqlf.Sprintf(
		listPermissionSyncJobQueryFmtstr,
		sqlf.Join(PermissionSyncJobColumns, ", "),
		whereClause,
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var syncJobs []*PermissionSyncJob
	for rows.Next() {
		job, err := ScanPermissionSyncJob(rows)
		if err != nil {
			return nil, err
		}
		syncJobs = append(syncJobs, job)
	}

	return syncJobs, nil
}

const listPermissionSyncJobQueryFmtstr = `
SELECT %s
FROM permission_sync_jobs
%s -- whereClause
ORDER BY id ASC
`

type PermissionSyncJob struct {
	ID                 int
	State              string
	FailureMessage     *string
	Reason             string
	CancellationReason string
	TriggeredByUserID  int32
	QueuedAt           time.Time
	StartedAt          time.Time
	FinishedAt         time.Time
	ProcessAfter       time.Time
	NumResets          int
	NumFailures        int
	LastHeartbeatAt    time.Time
	ExecutionLogs      []workerutil.ExecutionLogEntry
	WorkerHostname     string
	Cancel             bool

	RepositoryID int
	UserID       int

	Priority         PermissionSyncJobPriority
	InvalidateCaches bool
}

func (j *PermissionSyncJob) RecordID() int { return j.ID }

var PermissionSyncJobColumns = []*sqlf.Query{
	sqlf.Sprintf("permission_sync_jobs.id"),
	sqlf.Sprintf("permission_sync_jobs.state"),
	sqlf.Sprintf("permission_sync_jobs.reason"),
	sqlf.Sprintf("permission_sync_jobs.cancellation_reason"),
	sqlf.Sprintf("permission_sync_jobs.triggered_by_user_id"),
	sqlf.Sprintf("permission_sync_jobs.failure_message"),
	sqlf.Sprintf("permission_sync_jobs.queued_at"),
	sqlf.Sprintf("permission_sync_jobs.started_at"),
	sqlf.Sprintf("permission_sync_jobs.finished_at"),
	sqlf.Sprintf("permission_sync_jobs.process_after"),
	sqlf.Sprintf("permission_sync_jobs.num_resets"),
	sqlf.Sprintf("permission_sync_jobs.num_failures"),
	sqlf.Sprintf("permission_sync_jobs.last_heartbeat_at"),
	sqlf.Sprintf("permission_sync_jobs.execution_logs"),
	sqlf.Sprintf("permission_sync_jobs.worker_hostname"),
	sqlf.Sprintf("permission_sync_jobs.cancel"),

	sqlf.Sprintf("permission_sync_jobs.repository_id"),
	sqlf.Sprintf("permission_sync_jobs.user_id"),

	sqlf.Sprintf("permission_sync_jobs.priority"),
	sqlf.Sprintf("permission_sync_jobs.invalidate_caches"),
}

func ScanPermissionSyncJob(s dbutil.Scanner) (*PermissionSyncJob, error) {
	var job PermissionSyncJob
	if err := scanPermissionSyncJob(&job, s); err != nil {
		return nil, err
	}
	return &job, nil
}

func scanPermissionSyncJob(job *PermissionSyncJob, s dbutil.Scanner) error {
	var executionLogs []dbworkerstore.ExecutionLogEntry

	if err := s.Scan(
		&job.ID,
		&job.State,
		&job.Reason,
		&dbutil.NullString{S: &job.CancellationReason},
		&dbutil.NullInt32{N: &job.TriggeredByUserID},
		&job.FailureMessage,
		&job.QueuedAt,
		&dbutil.NullTime{Time: &job.StartedAt},
		&dbutil.NullTime{Time: &job.FinishedAt},
		&dbutil.NullTime{Time: &job.ProcessAfter},
		&job.NumResets,
		&job.NumFailures,
		&dbutil.NullTime{Time: &job.LastHeartbeatAt},
		pq.Array(&executionLogs),
		&job.WorkerHostname,
		&job.Cancel,

		&dbutil.NullInt{N: &job.RepositoryID},
		&dbutil.NullInt{N: &job.UserID},

		&job.Priority,
		&job.InvalidateCaches,
	); err != nil {
		return err
	}

	for _, entry := range executionLogs {
		job.ExecutionLogs = append(job.ExecutionLogs, workerutil.ExecutionLogEntry(entry))
	}
	return nil
}
