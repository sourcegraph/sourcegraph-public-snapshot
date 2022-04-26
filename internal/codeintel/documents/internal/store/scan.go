package store

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/documents/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

func scanDocuments(rows *sql.Rows, queryErr error) (documents []shared.Document, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var document shared.Document

		if err = rows.Scan(
			&document.Path,
		); err != nil {
			return nil, err
		}

		documents = append(documents, document)
	}

	return documents, nil
}
