package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type ExecutorStore interface {
	basestore.ShareableStore
	WithTransact(context.Context, func(ExecutorStore) error) error
	Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error)
	With(basestore.ShareableStore) ExecutorStore

	// List returns a set of executor activity records matching the given options.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is permitted to view executor details
	// (e.g., a site-admin).
	List(ctx context.Context, args ExecutorStoreListOptions) ([]types.Executor, error)

	// Count returns the number of executor activity records matching the given options.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is permitted to view executor details
	// (e.g., a site-admin).
	Count(ctx context.Context, args ExecutorStoreListOptions) (int, error)

	// GetByID returns an executor activity record by identifier. If no such record exists, a
	// false-valued flag is returned.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is permitted to view executor details
	// (e.g., a site-admin).
	GetByID(ctx context.Context, id int) (types.Executor, bool, error)

	// UpsertHeartbeat updates or creates an executor activity record for a particular executor instance.
	UpsertHeartbeat(ctx context.Context, executor types.Executor) error

	// DeleteInactiveHeartbeats deletes heartbeat records belonging to executor instances that have not pinged
	// the Sourcegraph instance in at least the given duration.
	DeleteInactiveHeartbeats(ctx context.Context, minAge time.Duration) error

	// GetByHostname returns an executor resolver for the given hostname, or
	// nil when there is no executor record matching the given hostname.
	//
	// 🚨 SECURITY: This always returns nil for non-site admins.
	GetByHostname(ctx context.Context, hostname string) (types.Executor, bool, error)
}

type ExecutorStoreListOptions struct {
	Query  string
	Active bool
	Offset int
	Limit  int
}

type executorStore struct {
	*basestore.Store
}

var _ ExecutorStore = (*executorStore)(nil)

// ExecutorsWith instantiates and returns a new ExecutorStore using the other store handle.
func ExecutorsWith(other basestore.ShareableStore) ExecutorStore {
	return &executorStore{
		Store: basestore.NewWithHandle(other.Handle()),
	}
}

func (s *executorStore) WithTransact(ctx context.Context, f func(ExecutorStore) error) error {
	return s.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		return f(&executorStore{Store: tx})
	})
}

func (s *executorStore) With(other basestore.ShareableStore) ExecutorStore {
	return &executorStore{Store: s.Store.With(other)}
}

func (s *executorStore) List(ctx context.Context, opts ExecutorStoreListOptions) (_ []types.Executor, err error) {
	return s.list(ctx, opts, timeutil.Now())
}

func (s *executorStore) list(ctx context.Context, opts ExecutorStoreListOptions, now time.Time) (_ []types.Executor, err error) {
	executors, err := scanExecutors(s.Query(ctx, sqlf.Sprintf(executorStoreListQuery, executorStoreListOptionsConditions(opts, now), opts.Limit, opts.Offset)))
	if err != nil {
		return nil, err
	}

	return executors, nil
}

func (s *executorStore) Count(ctx context.Context, opts ExecutorStoreListOptions) (int, error) {
	return s.count(ctx, opts, timeutil.Now())
}

func (s *executorStore) count(ctx context.Context, opts ExecutorStoreListOptions, now time.Time) (_ int, err error) {
	totalCount, _, err := basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf(executorStoreListCountQuery, executorStoreListOptionsConditions(opts, now))))
	if err != nil {
		return 0, err
	}

	return totalCount, nil
}

func executorStoreListOptionsConditions(opts ExecutorStoreListOptions, now time.Time) *sqlf.Query {
	conds := make([]*sqlf.Query, 0, 2)
	if opts.Query != "" {
		conds = append(conds, makeExecutorSearchCondition(opts.Query))
	}
	if opts.Active {
		conds = append(conds, sqlf.Sprintf("%s - h.last_seen_at <= '15 minutes'::interval", now))
	}

	whereConditions := sqlf.Sprintf("TRUE")
	if len(conds) > 0 {
		whereConditions = sqlf.Join(conds, " AND ")
	}
	return whereConditions
}

const executorStoreListCountQuery = `
SELECT COUNT(*)
FROM executor_heartbeats h
WHERE %s
`

const executorStoreListQuery = `
SELECT
	h.id,
	h.hostname,
	h.queue_name,
	h.queue_names,
	h.os,
	h.architecture,
	h.docker_version,
	h.executor_version,
	h.git_version,
	h.ignite_version,
	h.src_cli_version,
	h.first_seen_at,
	h.last_seen_at
FROM executor_heartbeats h
WHERE %s
ORDER BY h.first_seen_at DESC, h.id
LIMIT %s OFFSET %s
`

// makeExecutorSearchCondition returns a disjunction of LIKE clauses against all searchable columns of an executor.
func makeExecutorSearchCondition(term string) *sqlf.Query {
	searchableColumns := []string{
		"h.hostname",
		"h.queue_name",
		"h.os",
		"h.architecture",
		"h.docker_version",
		"h.executor_version",
		"h.git_version",
		"h.ignite_version",
		"h.src_cli_version",
	}

	var termConds []*sqlf.Query
	for _, column := range searchableColumns {
		termConds = append(termConds, sqlf.Sprintf(column+" ILIKE %s", "%"+term+"%"))
	}

	return sqlf.Sprintf("(%s)", sqlf.Join(termConds, " OR "))
}

func (s *executorStore) GetByID(ctx context.Context, id int) (types.Executor, bool, error) {
	preds := []*sqlf.Query{
		sqlf.Sprintf("h.id = %s", id),
	}
	return scanFirstExecutor(s.Query(ctx, sqlf.Sprintf(executorStoreGetQuery, sqlf.Join(preds, "AND"))))
}

func (s *executorStore) GetByHostname(ctx context.Context, hostname string) (types.Executor, bool, error) {
	preds := []*sqlf.Query{
		sqlf.Sprintf("h.hostname = %s", hostname),
	}
	return scanFirstExecutor(s.Query(ctx, sqlf.Sprintf(executorStoreGetQuery, sqlf.Join(preds, "AND"))))
}

const executorStoreGetQuery = `
SELECT
	h.id,
	h.hostname,
	h.queue_name,
	h.queue_names,
	h.os,
	h.architecture,
	h.docker_version,
	h.executor_version,
	h.git_version,
	h.ignite_version,
	h.src_cli_version,
	h.first_seen_at,
	h.last_seen_at
FROM
	executor_heartbeats h
WHERE
	%s
`

func (s *executorStore) UpsertHeartbeat(ctx context.Context, executor types.Executor) error {
	return s.upsertHeartbeat(ctx, executor, timeutil.Now())
}

func (s *executorStore) upsertHeartbeat(ctx context.Context, executor types.Executor, now time.Time) error {
	return s.Exec(ctx, sqlf.Sprintf(
		executorStoreUpsertHeartbeatQuery,

		executor.Hostname,
		dbutil.NullStringColumn(executor.QueueName),
		pq.Array(executor.QueueNames),
		executor.OS,
		executor.Architecture,
		executor.DockerVersion,
		executor.ExecutorVersion,
		executor.GitVersion,
		executor.IgniteVersion,
		executor.SrcCliVersion,
		now,
		now,
	))
}

const executorStoreUpsertHeartbeatQuery = `
INSERT INTO executor_heartbeats (
	hostname,
	queue_name,
	queue_names,
	os,
	architecture,
	docker_version,
	executor_version,
	git_version,
	ignite_version,
	src_cli_version,
	first_seen_at,
	last_seen_at
)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
ON CONFLICT (hostname) DO UPDATE
SET
	queue_name = EXCLUDED.queue_name,
	queue_names = EXCLUDED.queue_names,
	os = EXCLUDED.os,
	architecture = EXCLUDED.architecture,
	docker_version = EXCLUDED.docker_version,
	executor_version = EXCLUDED.executor_version,
	git_version = EXCLUDED.git_version,
	ignite_version = EXCLUDED.ignite_version,
	src_cli_version = EXCLUDED.src_cli_version,
	last_seen_at =EXCLUDED.last_seen_at
`

func (s *executorStore) DeleteInactiveHeartbeats(ctx context.Context, minAge time.Duration) error {
	return s.deleteInactiveHeartbeats(ctx, minAge, timeutil.Now())
}

func (s *executorStore) deleteInactiveHeartbeats(ctx context.Context, minAge time.Duration, now time.Time) error {
	return s.Exec(ctx, sqlf.Sprintf(executorStoreDeleteInactiveHeartbeatsQuery, now, minAge/time.Second))
}

const executorStoreDeleteInactiveHeartbeatsQuery = `
DELETE FROM executor_heartbeats
WHERE %s - last_seen_at >= %s * interval '1 second'
`

// scanExecutors reads executor objects from the given row object.
func scanExecutors(rows *sql.Rows, queryErr error) (_ []types.Executor, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var executors []types.Executor
	for rows.Next() {
		var executor types.Executor
		var sqlQueueName *string
		var sqlQueueNames pq.StringArray
		if err := rows.Scan(
			&executor.ID,
			&executor.Hostname,
			&sqlQueueName,
			&sqlQueueNames,
			&executor.OS,
			&executor.Architecture,
			&executor.DockerVersion,
			&executor.ExecutorVersion,
			&executor.GitVersion,
			&executor.IgniteVersion,
			&executor.SrcCliVersion,
			&executor.FirstSeenAt,
			&executor.LastSeenAt,
		); err != nil {
			return nil, err
		}

		if sqlQueueName != nil {
			executor.QueueName = *sqlQueueName
		}

		var queueNames []string
		for _, name := range sqlQueueNames {
			queueNames = append(queueNames, name)
		}
		executor.QueueNames = queueNames

		executors = append(executors, executor)
	}

	return executors, nil
}

// scanFirstExecutor scans a slice of executors from the return value of `*Store.query` and returns the first.
func scanFirstExecutor(rows *sql.Rows, err error) (types.Executor, bool, error) {
	executors, err := scanExecutors(rows, err)
	if err != nil || len(executors) == 0 {
		return types.Executor{}, false, err
	}
	return executors[0], true, nil
}
