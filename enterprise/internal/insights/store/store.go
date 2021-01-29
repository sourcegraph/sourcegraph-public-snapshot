package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

// Store exposes methods to read and write code insights domain models from
// persistent storage.
type Store struct {
	*basestore.Store
	now func() time.Time
}

// New returns a new Store backed by the given Timescale db.
func New(db dbutil.DB) *Store {
	return NewWithClock(db, timeutil.Now)
}

// NewWithClock returns a new Store backed by the given db and
// clock for timestamps.
func NewWithClock(db dbutil.DB, clock func() time.Time) *Store {
	return &Store{Store: basestore.NewWithDB(db, sql.TxOptions{}), now: clock}
}

var _ basestore.ShareableStore = &Store{}

// Handle returns the underlying transactable database handle.
// Needed to implement the ShareableStore interface.
func (s *Store) Handle() *basestore.TransactableHandle { return s.Store.Handle() }

// With creates a new Store with the given basestore.Shareable store as the
// underlying basestore.Store.
// Needed to implement the basestore.Store interface
func (s *Store) With(other basestore.ShareableStore) *Store {
	return &Store{Store: s.Store.With(other), now: s.now}
}

type SeriesPointsOpts struct {
	SeriesID *int32
	From, To *time.Time
	Limit    int
}

type SeriesPoint struct {
	Time  time.Time
	Value float64
}

func (s *Store) SeriesPoints(ctx context.Context, opts SeriesPointsOpts) ([]SeriesPoint, error) {
	points := make([]SeriesPoint, 0, opts.Limit)
	err := s.query(ctx, seriesPointsQuery(opts), func(sc scanner) error {
		var point SeriesPoint
		err := sc.Scan(
			&point.Time,
			&point.Value,
		)
		if err != nil {
			return err
		}
		points = append(points, point)
		return nil
	})
	return points, err
}

var seriesPointsQueryFmtstr = `
-- source: enterprise/internal/insights/store/series_points.go
SELECT time, value FROM series_points
WHERE %s
ORDER BY time DESC
`

func seriesPointsQuery(opts SeriesPointsOpts) *sqlf.Query {
	preds := []*sqlf.Query{}

	if opts.SeriesID != nil {
		preds = append(preds, sqlf.Sprintf("series_id = %s", *opts.SeriesID))
	}
	if opts.From != nil {
		preds = append(preds, sqlf.Sprintf("time > %s", *opts.From))
	}
	if opts.To != nil {
		preds = append(preds, sqlf.Sprintf("time < %s", *opts.To))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}
	limitClause := ""
	if opts.Limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", opts.Limit)
	}
	return sqlf.Sprintf(
		seriesPointsQueryFmtstr+limitClause,
		sqlf.Join(preds, "\n AND "),
	)
}

func (s *Store) query(ctx context.Context, q *sqlf.Query, sc scanFunc) error {
	rows, err := s.Store.Query(ctx, q)
	if err != nil {
		return err
	}
	return scanAll(rows, sc)
}

// scanner captures the Scan method of sql.Rows and sql.Row
type scanner interface {
	Scan(dst ...interface{}) error
}

// a scanFunc scans one or more rows from a scanner, returning
// the last id column scanned and the count of scanned rows.
type scanFunc func(scanner) (err error)

func scanAll(rows *sql.Rows, scan scanFunc) (err error) {
	defer func() { err = basestore.CloseRows(rows, err) }()
	for rows.Next() {
		if err = scan(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}
