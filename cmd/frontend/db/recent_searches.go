package db

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

type RecentSearches struct {
	DB func() *sql.DB
}

// Add inserts the query q to the recent_searches table in the db.
func (rs *RecentSearches) Log(ctx context.Context, q string) error {
	insert := `INSERT INTO recent_searches (query) VALUES ($1)`
	res, err := rs.DB().ExecContext(ctx, insert, q)
	if err != nil {
		return errors.Errorf("inserting %q into recent_searches table: %v", q, err)
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return errors.Errorf("getting number of affected rows: %v", err)
	}
	if nrows == 0 {
		return errors.Errorf("failed to insert row for query %q", q)
	}
	return nil
}

// DeleteExcessRows keeps the row count in the recent_searches table below limit.
func (rs *RecentSearches) Cleanup(ctx context.Context, limit int) error {
	enforceLimit := `
DELETE FROM recent_searches
	WHERE id <
		(SELECT id FROM recent_searches
		 ORDER BY id
		 OFFSET GREATEST(0, (SELECT (SELECT COUNT(*) FROM recent_searches) - $1))
		 LIMIT 1)
`
	if _, err := rs.DB().ExecContext(ctx, enforceLimit, limit); err != nil {
		return errors.Errorf("deleting excess rows in recent_searches table: %v", err)
	}
	return nil
}

// Get returns all the search queries in the recent_searches table.
func (rs *RecentSearches) List(ctx context.Context) ([]string, error) {
	sel := `SELECT query FROM recent_searches`
	rows, err := rs.DB().QueryContext(ctx, sel)
	var qs []string
	if err != nil {
		return nil, errors.Errorf("running SELECT query: %v", err)
	}
	for rows.Next() {
		var q string
		if err := rows.Scan(&q); err != nil {
			return nil, err
		}
		qs = append(qs, q)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return qs, nil
}
