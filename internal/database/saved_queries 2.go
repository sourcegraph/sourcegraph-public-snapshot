package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type QueryRunnerStateStore struct {
	*basestore.Store
}

// QueryRunnerState instantiates and returns a new QueryRunnerStateStore with prepared statements.
func QueryRunnerState(db dbutil.DB) *QueryRunnerStateStore {
	return &QueryRunnerStateStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

// NewQueryRunnerStateStoreWithDB instantiates and returns a new QueryRunnerStateStore using the other store handle.
func QueryRunnerStateWith(other basestore.ShareableStore) *QueryRunnerStateStore {
	return &QueryRunnerStateStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *QueryRunnerStateStore) With(other basestore.ShareableStore) *QueryRunnerStateStore {
	return &QueryRunnerStateStore{Store: s.Store.With(other)}
}

func (s *QueryRunnerStateStore) Transact(ctx context.Context) (*QueryRunnerStateStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &QueryRunnerStateStore{Store: txBase}, err
}

type SavedQueryInfo struct {
	Query        string
	LastExecuted time.Time
	LatestResult time.Time
	ExecDuration time.Duration
}

// Get gets the saved query information for the given query. nil
// is returned if there is no existing saved query info.
func (s *QueryRunnerStateStore) Get(ctx context.Context, query string) (*SavedQueryInfo, error) {
	info := &SavedQueryInfo{
		Query: query,
	}
	var execDurationNs int64
	err := s.Handle().DB().QueryRowContext(
		ctx,
		"SELECT last_executed, latest_result, exec_duration_ns FROM query_runner_state WHERE query=$1",
		query,
	).Scan(&info.LastExecuted, &info.LatestResult, &execDurationNs)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "QueryRow")
	}
	info.ExecDuration = time.Duration(execDurationNs)
	return info, nil
}

// Set sets the saved query information for the given info.Query.
//
// It is not safe to call concurrently for the same info.Query, as it uses a
// poor man's upsert implementation.
func (s *QueryRunnerStateStore) Set(ctx context.Context, info *SavedQueryInfo) error {
	res, err := s.Handle().DB().ExecContext(
		ctx,
		"UPDATE query_runner_state SET last_executed=$1, latest_result=$2, exec_duration_ns=$3 WHERE query=$4",
		info.LastExecuted,
		info.LatestResult,
		int64(info.ExecDuration),
		info.Query,
	)
	if err != nil {
		return errors.Wrap(err, "UPDATE")
	}
	updated, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "RowsAffected")
	}
	if updated == 0 {
		// Didn't update any row, so insert a new one.
		_, err := s.Handle().DB().ExecContext(
			ctx,
			"INSERT INTO query_runner_state(query, last_executed, latest_result, exec_duration_ns) VALUES($1, $2, $3, $4)",
			info.Query,
			info.LastExecuted,
			info.LatestResult,
			int64(info.ExecDuration),
		)
		if err != nil {
			return errors.Wrap(err, "INSERT")
		}
	}
	return nil
}

func (s *QueryRunnerStateStore) Delete(ctx context.Context, query string) error {
	_, err := s.Handle().DB().ExecContext(
		ctx,
		"DELETE FROM query_runner_state WHERE query=$1",
		query,
	)
	return err
}
