package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

// recordScanner is a callback that knows how to scan the results of the query
// that has been executed into a target object.
type recordScanner[T any] func(target *T, sc dbutil.Scanner) error

// createOrUpdateRecord executes the given query, scans the results back into
// the given record, and returns any error.
func createOrUpdateRecord[T any](ctx context.Context, s *Store, q *sqlf.Query, rs recordScanner[T], record *T) error {
	return s.query(ctx, q, func(sc dbutil.Scanner) error {
		return rs(record, sc)
	})
}

// getRecord returns a single record, if any, from the given query, and return
// ErrNoResults if no record was found.
func getRecord[T any](ctx context.Context, s *Store, q *sqlf.Query, rs recordScanner[T]) (*T, error) {
	var (
		record T
		exists bool
	)

	if err := s.query(ctx, q, func(sc dbutil.Scanner) error {
		exists = true
		return rs(&record, sc)
	}); err != nil {
		return nil, err
	}

	if !exists {
		return nil, ErrNoResults
	}

	return &record, nil
}
