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
	List(ctx context.Context) ([]Executor, int, error)
	GetByID(ctx context.Context, id int) (Executor, bool, error)
	Heartbeat(ctx context.Context, hostname string) error
	Transact(ctx context.Context) (ExecutorStore, error)
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

// TODO - document
// TODO - test
//
// TODO - redocument security concern
// 🚨 SECURITY: The caller must ensure that the actor is permitted to create tokens for the
// specified user (i.e., that the actor is either the user or a site admin).
func (s *executorStore) List(ctx context.Context) ([]Executor, int, error) {
	executors, err := scanExecutors(s.Query(ctx, sqlf.Sprintf(executorStoreListQuery)))
	if err != nil {
		return nil, 0, err
	}

	// TODO - paginate
	// TODO - query count
	return executors, len(executors), nil
}

const executorStoreListQuery = `
-- source: internal/database/executors.go:List
SELECT
	h.id,
	h.hostname,
	h.last_seen_at
FROM executor_heartbeats h
`

// TODO - document
// TODO - test
//
// TODO - redocument security concern
// 🚨 SECURITY: The caller must ensure that the actor is permitted to create tokens for the
// specified user (i.e., that the actor is either the user or a site admin).
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

// TODO - document
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
