package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/services/executors/store"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// scanExecutors reads executor objects from the given row object.
func scanExecutors(rows *sql.Rows, queryErr error) (_ []types.Executor, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var executors []types.Executor
	for rows.Next() {
		var executor types.Executor
		if err := rows.Scan(
			&executor.ID,
			&executor.Hostname,
			&executor.QueueName,
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

func (s *ExecutorStore) List(ctx context.Context, opts store.ExecutorStoreListOptions) (_ []types.Executor, _ int, err error) {
	return s.list(ctx, opts, timeutil.Now())
}

func (s *ExecutorStore) list(ctx context.Context, opts store.ExecutorStoreListOptions, now time.Time) (_ []types.Executor, _ int, err error) {
	tx, err := s.db.Transact(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer func() { err = tx.Done(err) }()

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

	totalCount, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(executorStoreListCountQuery, whereConditions)))
	if err != nil {
		return nil, 0, err
	}

	executors, err := scanExecutors(tx.Query(ctx, sqlf.Sprintf(executorStoreListQuery, whereConditions, opts.Limit, opts.Offset)))
	if err != nil {
		return nil, 0, err
	}

	return executors, totalCount, nil
}

const executorStoreListCountQuery = `
-- source: internal/database/executors.go:List
SELECT COUNT(*)
FROM executor_heartbeats h
WHERE %s
`

const executorStoreListQuery = `
-- source: internal/database/executors.go:List
SELECT
	h.id,
	h.hostname,
	h.queue_name,
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

func (s *ExecutorStore) GetByID(ctx context.Context, id int) (types.Executor, bool, error) {
	return scanFirstExecutor(s.db.Query(ctx, sqlf.Sprintf(executorStoreGetByIDQuery, id)))
}

const executorStoreGetByIDQuery = `
-- source: internal/database/executors.go:GetByID
SELECT
	h.id,
	h.hostname,
	h.queue_name,
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
WHERE h.id = %s
`

func (s *ExecutorStore) GetByHostname(ctx context.Context, hostname string) (types.Executor, bool, error) {
	return scanFirstExecutor(s.db.Query(ctx, sqlf.Sprintf(executorStoreGetByHostnameQuery, hostname)))
}

const executorStoreGetByHostnameQuery = `
-- source: internal/database/executors.go:GetByHostname
SELECT
	h.id,
	h.hostname,
	h.queue_name,
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
WHERE h.hostname = %s
`

func (s *ExecutorStore) UpsertHeartbeat(ctx context.Context, executor types.Executor) error {
	return s.upsertHeartbeat(ctx, executor, timeutil.Now())
}

func (s *ExecutorStore) upsertHeartbeat(ctx context.Context, executor types.Executor, now time.Time) error {
	return s.db.Exec(ctx, sqlf.Sprintf(
		executorStoreUpsertHeartbeatQuery,

		// insert
		executor.Hostname,
		executor.QueueName,
		executor.OS,
		executor.Architecture,
		executor.DockerVersion,
		executor.ExecutorVersion,
		executor.GitVersion,
		executor.IgniteVersion,
		executor.SrcCliVersion,
		now,
		now,

		// update
		executor.QueueName,
		executor.OS,
		executor.Architecture,
		executor.DockerVersion,
		executor.ExecutorVersion,
		executor.GitVersion,
		executor.IgniteVersion,
		executor.SrcCliVersion,
		now,
	))
}

const executorStoreUpsertHeartbeatQuery = `
-- source: internal/database/executors.go:HeartbeatHeartbeat
INSERT INTO executor_heartbeats (
	hostname,
	queue_name,
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
VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
ON CONFLICT (hostname) DO UPDATE
SET
	queue_name = %s,
	os = %s,
	architecture = %s,
	docker_version = %s,
	executor_version = %s,
	git_version = %s,
	ignite_version = %s,
	src_cli_version = %s,
	last_seen_at = %s
`

func (s *ExecutorStore) DeleteInactiveHeartbeats(ctx context.Context, minAge time.Duration) error {
	return s.deleteInactiveHeartbeats(ctx, minAge, timeutil.Now())
}

func (s *ExecutorStore) deleteInactiveHeartbeats(ctx context.Context, minAge time.Duration, now time.Time) error {
	return s.db.Exec(ctx, sqlf.Sprintf(executorStoreDeleteInactiveHeartbeatsQuery, now, minAge/time.Second))
}

const executorStoreDeleteInactiveHeartbeatsQuery = `
-- source: internal/database/executors.go:DeleteInactiveHeartbeats
DELETE FROM executor_heartbeats
WHERE %s - last_seen_at >= %s * interval '1 second'
`
