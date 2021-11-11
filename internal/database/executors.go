package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

// Executor describes an executor instance that has recently connected to Sourcegraph.
type Executor struct {
	ID         int
	Hostname   string
	LastSeenAt time.Time
}

type ExecutorStore interface {
	List(ctx context.Context, args ExecutorStoreListOptions) ([]Executor, int, error)
	GetByID(ctx context.Context, id int) (Executor, bool, error)
	Heartbeat(ctx context.Context, hostname string) error
	Transact(ctx context.Context) (ExecutorStore, error)
	Done(err error) error
	With(store basestore.ShareableStore) ExecutorStore
	basestore.ShareableStore
}

type executorStore struct {
	*basestore.Store
}

var _ ExecutorStore = (*executorStore)(nil)

// Executors instantiates and returns a new ExecutorStore with prepared statements.
func Executors(db dbutil.DB) ExecutorStore {
	return &executorStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

// ExecutorsWith instantiates and returns a new ExecutorStore using the other store handle.
func ExecutorsWith(other basestore.ShareableStore) ExecutorStore {
	return &executorStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *executorStore) With(other basestore.ShareableStore) ExecutorStore {
	return &executorStore{Store: s.Store.With(other)}
}

func (s *executorStore) Done(err error) error {
	return s.Store.Done(err)
}

func (s *executorStore) Transact(ctx context.Context) (ExecutorStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &executorStore{Store: txBase}, err
}

// scanExecutors reads executor objects from the given row object.
func scanExecutors(rows *sql.Rows, queryErr error) (_ []Executor, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var executors []Executor
	for rows.Next() {
		var id int
		var hostname string
		var lastSeenAt time.Time
		if err := rows.Scan(&id, &hostname, &lastSeenAt); err != nil {
			return nil, err
		}

		executors = append(executors, Executor{
			ID:         id,
			Hostname:   hostname,
			LastSeenAt: lastSeenAt,
		})
	}

	return executors, nil
}

// scanFirstExecutor scans a slice of executors from the return value of `*Store.query` and returns the first.
func scanFirstExecutor(rows *sql.Rows, err error) (Executor, bool, error) {
	executors, err := scanExecutors(rows, err)
	if err != nil || len(executors) == 0 {
		return Executor{}, false, err
	}
	return executors[0], true, nil
}

type ExecutorStoreListOptions struct {
	Offset int
	Limit  int
}

// TODO - test
// List returns a set of executor activity records matching the given options.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view executor details
// (e.g., a site-admin).
func (s *executorStore) List(ctx context.Context, opts ExecutorStoreListOptions) (_ []Executor, _ int, err error) {
	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer func() { err = tx.Done(err) }()

	totalCount, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(executorStoreListCountQuery)))
	if err != nil {
		return nil, 0, err
	}

	executors, err := scanExecutors(tx.Query(ctx, sqlf.Sprintf(executorStoreListQuery, opts.Limit, opts.Offset)))
	if err != nil {
		return nil, 0, err
	}

	return executors, totalCount, nil
}

const executorStoreListCountQuery = `
-- source: internal/database/executors.go:List
SELECT COUNT(*)
FROM executor_heartbeats h
`

const executorStoreListQuery = `
-- source: internal/database/executors.go:List
SELECT
	h.id,
	h.hostname,
	h.last_seen_at
FROM executor_heartbeats h
ORDER BY h.last_seen_at DESC
LIMIT %s OFFSET %s
`

// TODO - test
// GetByID returns an executor activity record by identifier. If no such record exists, a
// false-valued flag is returned.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view executor details
// (e.g., a site-admin).
func (s *executorStore) GetByID(ctx context.Context, id int) (Executor, bool, error) {
	return scanFirstExecutor(s.Query(ctx, sqlf.Sprintf(executorStoreGetByIDQuery, id)))
}

const executorStoreGetByIDQuery = `
-- source: internal/database/executors.go:GetByID
SELECT
	h.id,
	h.hostname,
	h.last_seen_at
FROM executor_heartbeats h
WHERE h.id = %s
`

// Heartbeat updates or creates an executor activity record for a particular executor instance.
func (s *executorStore) Heartbeat(ctx context.Context, hostname string) error {
	return s.heartbeat(ctx, hostname, timeutil.Now())
}

// TODO - test
func (s *executorStore) heartbeat(ctx context.Context, hostname string, now time.Time) error {
	return s.Exec(ctx, sqlf.Sprintf(executorStoryHeartbeatQuery, hostname, now, now))
}

const executorStoryHeartbeatQuery = `
-- source: internal/database/executors.go:HeartbeatHeartbeat
INSERT INTO executor_heartbeats (hostname, last_seen_at)
VALUES (%s, %s)
ON CONFLICT (hostname) DO UPDATE SET last_seen_at = %s
`
