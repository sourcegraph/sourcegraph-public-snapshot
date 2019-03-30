package db

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

type recentSearches struct{}

// Add inserts the query q to the recentSearches table in the db.
func (*recentSearches) Add(ctx context.Context, q string) error {
	insert := `INSERT INTO recent_searches (query) VALUES ($1)`
	if dbconn.Global == nil {
		return errors.New("db connection is nil")
	}
	res, err := dbconn.Global.ExecContext(ctx, insert, q)
	if err != nil {
		return errors.Errorf("inserting %q into recentSearches table: %v", q, err)
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

// DeleteExcessRows keeps the row count in the recentSearches table below limit.
func (*recentSearches) DeleteExcessRows(ctx context.Context, limit int) error {
	enforceLimit := `
DELETE FROM recent_searches
	WHERE id <
		(SELECT id FROM recent_searches
		 ORDER BY id
		 OFFSET GREATEST(0, (SELECT (SELECT COUNT(*) FROM recent_searches) - $1))
		 LIMIT 1)
`
	if dbconn.Global == nil {
		return errors.New("db connection is nil")
	}
	if _, err := dbconn.Global.ExecContext(ctx, enforceLimit, limit); err != nil {
		return errors.Errorf("deleting excess rows in recentSearches table: %v", err)
	}
	return nil
}

// Get returns all the search queries in the recentSearches table.
func (*recentSearches) Get(ctx context.Context) ([]string, error) {
	sel := `SELECT query FROM recent_searches`
	rows, err := dbconn.Global.QueryContext(ctx, sel)
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
