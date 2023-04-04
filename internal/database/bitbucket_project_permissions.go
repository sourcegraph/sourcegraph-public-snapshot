package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// BitbucketProjectPermissionsStore is used by the BitbucketProjectPermissions worker
// to apply permissions asynchronously.
type BitbucketProjectPermissionsStore interface {
	basestore.ShareableStore
	With(other basestore.ShareableStore) BitbucketProjectPermissionsStore
	Enqueue(ctx context.Context, projectKey string, externalServiceID int64, permissions []types.UserPermission, unrestricted bool) (int, error)
	WithTransact(context.Context, func(BitbucketProjectPermissionsStore) error) error
	ListJobs(ctx context.Context, opt ListJobsOptions) ([]*types.BitbucketProjectPermissionJob, error)
}

type bitbucketProjectPermissionsStore struct {
	*basestore.Store
}

// BitbucketProjectPermissionsStoreWith instantiates and returns a new BitbucketProjectPermissionsStore using
// the other store handle.
func BitbucketProjectPermissionsStoreWith(other basestore.ShareableStore) BitbucketProjectPermissionsStore {
	return &bitbucketProjectPermissionsStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *bitbucketProjectPermissionsStore) With(other basestore.ShareableStore) BitbucketProjectPermissionsStore {
	return &bitbucketProjectPermissionsStore{Store: s.Store.With(other)}
}

func (s *bitbucketProjectPermissionsStore) copy() *bitbucketProjectPermissionsStore {
	return &bitbucketProjectPermissionsStore{
		Store: s.Store,
	}
}

func (s *bitbucketProjectPermissionsStore) WithTransact(ctx context.Context, f func(BitbucketProjectPermissionsStore) error) error {
	return s.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		c := s.copy()
		c.Store = tx
		return f(c)
	})
}

// Enqueue a job to apply permissions to a Bitbucket project, returning its jobID.
// The job will be enqueued to the BitbucketProjectPermissions worker.
// If a non-empty permissions slice is passed, unrestricted has to be false, and vice versa.
func (s *bitbucketProjectPermissionsStore) Enqueue(ctx context.Context, projectKey string, externalServiceID int64, permissions []types.UserPermission, unrestricted bool) (int, error) {
	if len(permissions) > 0 && unrestricted {
		return 0, errors.New("cannot specify permissions when unrestricted is true")
	}
	if len(permissions) == 0 && !unrestricted {
		return 0, errors.New("must specify permissions when unrestricted is false")
	}

	var perms []userPermission
	for _, perm := range permissions {
		perms = append(perms, userPermission(perm))
	}

	var jobID int
	err := s.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		// ensure we don't enqueue a job for the same project twice.
		// if so, cancel the existing jobs and enqueue a new one.
		// this doesn't apply to running jobs.
		err := tx.Exec(ctx, sqlf.Sprintf(`--sql
UPDATE explicit_permissions_bitbucket_projects_jobs SET state = 'canceled' WHERE project_key = %s AND external_service_id = %s AND state = 'queued'
`, projectKey, externalServiceID))
		if err != nil && err != sql.ErrNoRows {
			return err
		}

		err = tx.QueryRow(ctx, sqlf.Sprintf(`--sql
INSERT INTO
	explicit_permissions_bitbucket_projects_jobs (project_key, external_service_id, permissions, unrestricted)
VALUES (%s, %s, %s, %s) RETURNING id
	`, projectKey, externalServiceID, pq.Array(perms), unrestricted)).Scan(&jobID)
		if err != nil {
			return err
		}

		return nil
	})
	return jobID, err
}

var BitbucketProjectPermissionsColumnExpressions = []*sqlf.Query{
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.id"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.state"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.failure_message"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.queued_at"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.started_at"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.finished_at"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.process_after"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.num_resets"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.num_failures"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.last_heartbeat_at"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.execution_logs"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.worker_hostname"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.project_key"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.external_service_id"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.permissions"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.unrestricted"),
}

type ListJobsOptions struct {
	ProjectKeys []string
	State       string
	Count       int32
}

// ListJobs returns a list of types.BitbucketProjectPermissionJob for a given set
// of query options: ListJobsOptions
func (s *bitbucketProjectPermissionsStore) ListJobs(
	ctx context.Context,
	opt ListJobsOptions,
) (jobs []*types.BitbucketProjectPermissionJob, err error) {
	query := listWorkerJobsQuery(opt)

	rows, err := s.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var job *types.BitbucketProjectPermissionJob
		job, err = ScanBitbucketProjectPermissionJob(rows)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}

	return
}

func ScanBitbucketProjectPermissionJob(rows dbutil.Scanner) (*types.BitbucketProjectPermissionJob, error) {
	var job types.BitbucketProjectPermissionJob
	var executionLogs []executor.ExecutionLogEntry
	var permissions []userPermission

	if err := rows.Scan(
		&job.ID,
		&job.State,
		&job.FailureMessage,
		&job.QueuedAt,
		&job.StartedAt,
		&job.FinishedAt,
		&job.ProcessAfter,
		&job.NumResets,
		&job.NumFailures,
		&dbutil.NullTime{Time: &job.LastHeartbeatAt},
		pq.Array(&executionLogs),
		&job.WorkerHostname,
		&job.ProjectKey,
		&job.ExternalServiceID,
		pq.Array(&permissions),
		&job.Unrestricted,
	); err != nil {
		return nil, err
	}

	for _, entry := range executionLogs {
		logEntry := entry
		job.ExecutionLogs = append(job.ExecutionLogs, &logEntry)
	}

	for _, perm := range permissions {
		job.Permissions = append(job.Permissions, types.UserPermission(perm))
	}
	return &job, nil
}

const maxJobsCount = 500

func listWorkerJobsQuery(opt ListJobsOptions) *sqlf.Query {
	var where []*sqlf.Query

	q := `
SELECT id, state, failure_message, queued_at, started_at, finished_at, process_after, num_resets, num_failures, last_heartbeat_at, execution_logs, worker_hostname, project_key, external_service_id, permissions, unrestricted
FROM explicit_permissions_bitbucket_projects_jobs
%%s
ORDER BY queued_at DESC
LIMIT %d
`

	// we don't want to accept too many projects, that's why the input slice is trimmed
	if len(opt.ProjectKeys) != 0 {
		keys := opt.ProjectKeys
		if len(opt.ProjectKeys) > maxJobsCount {
			keys = keys[:maxJobsCount]
		}
		keyQueries := make([]*sqlf.Query, 0, len(keys))
		for _, key := range keys {
			keyQueries = append(keyQueries, sqlf.Sprintf("%s", key))
		}

		where = append(where, sqlf.Sprintf("project_key IN (%s)", sqlf.Join(keyQueries, ",")))
	}

	if opt.State != "" {
		where = append(where, sqlf.Sprintf("state = %s", opt.State))
	}

	whereClause := sqlf.Sprintf("")
	if len(where) != 0 {
		whereClause = sqlf.Sprintf("WHERE %s", sqlf.Join(where, "AND"))
	}

	var limitNum int32 = 100

	if opt.Count > 0 && opt.Count < maxJobsCount {
		limitNum = opt.Count
	} else if opt.Count >= maxJobsCount {
		limitNum = maxJobsCount
	}

	return sqlf.Sprintf(fmt.Sprintf(q, limitNum), whereClause)
}

type userPermission types.UserPermission

func (p *userPermission) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.Errorf("value is not []byte: %T", value)
	}

	return json.Unmarshal(b, &p)
}

func (p userPermission) Value() (driver.Value, error) {
	return json.Marshal(p)
}
