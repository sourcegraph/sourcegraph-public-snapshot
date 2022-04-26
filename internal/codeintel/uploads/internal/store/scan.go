package store

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

func scanUploads(rows *sql.Rows, queryErr error) (uploads []shared.Upload, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var upload shared.Upload

		if err = rows.Scan(
			&upload.ID,
		); err != nil {
			return nil, err
		}

		uploads = append(uploads, upload)
	}

	return uploads, nil
}
