package store

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

func scanIndexJobs(rows *sql.Rows, queryErr error) (indexJobs []shared.IndexJob, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var indexJob shared.IndexJob

		if err = rows.Scan(
			&indexJob.Indexer,
		); err != nil {
			return nil, err
		}

		indexJobs = append(indexJobs, indexJob)
	}

	return indexJobs, nil
}
