package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"

	"github.com/pkg/errors"
)

type RecentSearches struct{}

// Log inserts the query q into to the recent_searches table in the db.
func (rs *RecentSearches) Log(ctx context.Context, q string) error {
	insert := `INSERT INTO recent_searches (query) VALUES ($1)`
	res, err := dbconn.Global.ExecContext(ctx, insert, q)
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

// Cleanup keeps the row count in the recent_searches table below limit.
func (rs *RecentSearches) Cleanup(ctx context.Context, limit int) error {
	enforceLimit := `
DELETE FROM recent_searches
	WHERE id <
		(SELECT id FROM recent_searches
		 ORDER BY id
		 OFFSET GREATEST(0, (SELECT (SELECT COUNT(*) FROM recent_searches) - $1))
		 LIMIT 1)
`
	if _, err := dbconn.Global.ExecContext(ctx, enforceLimit, limit); err != nil {
		return errors.Errorf("deleting excess rows in recent_searches table: %v", err)
	}
	return nil
}

// List returns all the search queries in the recent_searches table.
func (rs *RecentSearches) List(ctx context.Context) ([]string, error) {
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

// Top returns the top n queries in the recent_searches table.
func (rs *RecentSearches) Top(ctx context.Context, n int32) ([]string, []int32, error) {
	sel := `SELECT query, COUNT(*) FROM recent_searches GROUP BY query ORDER BY count DESC, query ASC LIMIT $1`
	rows, err := dbconn.Global.QueryContext(ctx, sel, n)
	if err != nil {
		return nil, nil, errors.Wrap(err, "running db query to get top search queries")
	}
	var queries []string
	var counts []int32
	for rows.Next() {
		var query string
		var count int32
		if err := rows.Scan(&query, &count); err != nil {
			return nil, nil, errors.Wrap(err, "scanning row")
		}
		queries = append(queries, query)
		counts = append(counts, count)
	}
	return queries, counts, nil
}
