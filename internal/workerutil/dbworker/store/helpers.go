package store

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

// BuildWorkerScan builds a callback that can be used as a `Scan` field in an
// `Options` struct. It must be given a function that can take a scanner and
// return a type that implements `workerutil.Record`.
func BuildWorkerScan[T workerutil.Record](scan func(dbutil.Scanner) (T, error)) func(*sql.Rows, error) ([]workerutil.Record, error) {
	return func(rows *sql.Rows, err error) ([]workerutil.Record, error) {
		if err != nil {
			return nil, err
		}

		defer func() { err = basestore.CloseRows(rows, err) }()

		records := []workerutil.Record{}
		for rows.Next() {
			record, err := scan(rows)
			if err != nil {
				return nil, err
			}

			records = append(records, record)
		}

		return records, nil
	}
}
