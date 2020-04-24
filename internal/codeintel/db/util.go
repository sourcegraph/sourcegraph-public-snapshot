package db

import (
	"database/sql"

	"github.com/keegancsmith/sqlf"
)

// ignoreErrNoRows returns the given error if it's not sql.ErrNoRows.
func ignoreErrNoRows(err error) error {
	if err == sql.ErrNoRows {
		return nil
	}
	return err
}

// intsToQueries converts a slice of ints into a slice of queries.
func intsToQueries(values []int) []*sqlf.Query {
	var queries []*sqlf.Query
	for _, value := range values {
		queries = append(queries, sqlf.Sprintf("%d", value))
	}

	return queries
}
