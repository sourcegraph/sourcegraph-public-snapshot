package internal

import (
	"context"
	"database/sql"
	"io"
	"time"

	"github.com/keegancsmith/sqlf"
)

// a paginatedQuery returns a query with the given pagination
// parameters
type paginatedQuery func(cursor, limit int64) *sqlf.Query

func (s *Store) paginate(ctx context.Context, limit, perPage int64, cursor int64, q paginatedQuery, scan scanFunc) (err error) {
	const defaultPerPageLimit = 10000

	if perPage <= 0 {
		perPage = defaultPerPageLimit
	}

	if limit > 0 && perPage > limit {
		perPage = limit
	}

	var (
		remaining   = limit
		next, count int64
	)

	// We need this so that we enter the loop below for the first iteration
	// since cursor will be < next
	next = cursor
	cursor--

	for cursor < next && err == nil && (limit <= 0 || remaining > 0) {
		cursor = next
		next, count, err = s.list(ctx, q(cursor, perPage), scan)
		if limit > 0 {
			if remaining -= count; perPage > remaining {
				perPage = remaining
			}
		}
	}

	return err
}

// scanner captures the Scan method of sql.Rows and sql.Row
type scanner interface {
	Scan(dst ...interface{}) error
}

// a scanFunc scans one or more rows from a scanner, returning
// the last id column scanned and the count of scanned rows.
type scanFunc func(scanner) (last, count int64, err error)

func scanAll(rows *sql.Rows, scan scanFunc) (last, count int64, err error) {
	defer closeErr(rows, &err)

	last = -1
	for rows.Next() {
		var n int64
		if last, n, err = scan(rows); err != nil {
			return last, count, err
		}
		count += n
	}

	return last, count, rows.Err()
}

func closeErr(c io.Closer, err *error) {
	if e := c.Close(); err != nil && *err == nil {
		*err = e
	}
}

func nullTimeColumn(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func nullStringColumn(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullInt32Column(i int32) *int32 {
	if i == 0 {
		return nil
	}
	return &i
}
