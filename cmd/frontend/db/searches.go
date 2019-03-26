package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

type searches struct{}

// Add adds the query q to the searches table in the db, deleting old rows in the case
// where the number of rows exceeds limit.
func (*searches) Add(ctx context.Context, q string, limit int) error {
	insert := `INSERT INTO searches (query) VALUES ($1)`
	if dbconn.Global == nil {
		return errors.New("db connection is nil")
	}
	res, err := dbconn.Global.ExecContext(ctx, insert, q)
	if err != nil {
		return fmt.Errorf("inserting '%s' into searches table: %v", q, err)
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting number of affected rows: %v", err)
	}
	if nrows == 0 {
		return fmt.Errorf("failed to insert row for query '%s'", q)
	}
	// Keep the row count to no more than limit.
	enforceLimit := `DELETE FROM searches WHERE id <= (SELECT MAX(id) FROM SEARCHES) - $1`
	if _, err = dbconn.Global.ExecContext(ctx, enforceLimit, limit); err != nil {
		return fmt.Errorf("enforcing limit on number of rows in searches table: %v", err)
	}
	return nil
}

// Get returns all the search queries in the searches table.
func (*searches) Get(ctx context.Context) ([]string, error) {
	sel := `SELECT query FROM searches`
	rows, err := dbconn.Global.QueryContext(ctx, sel)
	var qs []string
	if err != nil {
		return nil, fmt.Errorf("running SELECT query: %v", err)
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
