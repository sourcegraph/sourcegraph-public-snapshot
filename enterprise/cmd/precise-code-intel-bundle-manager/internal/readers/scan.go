package readers

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
)

type document struct {
	path string
	data string
	err  error
}

func scanDocuments(rows *sql.Rows, queryErr error) (<-chan document, error) {
	if queryErr != nil {
		return nil, queryErr
	}

	ch := make(chan document)
	go func() {
		defer close(ch)

		defer func() {
			if err := basestore.CloseRows(rows, nil); err != nil {
				ch <- document{err: err}
			}
		}()

		for rows.Next() {
			var value document
			if err := rows.Scan(&value.path, &value.data); err != nil {
				ch <- document{err: err}
				return
			}

			ch <- value
		}
	}()

	return ch, nil
}

type resultChunk struct {
	index int
	data  string
	err   error
}

func scanResultChunks(rows *sql.Rows, queryErr error) (<-chan resultChunk, error) {
	if queryErr != nil {
		return nil, queryErr
	}

	ch := make(chan resultChunk)
	go func() {
		defer close(ch)

		defer func() {
			if err := basestore.CloseRows(rows, nil); err != nil {
				ch <- resultChunk{err: err}
			}
		}()

		for rows.Next() {
			var value resultChunk
			if err := rows.Scan(&value.index, &value.data); err != nil {
				ch <- resultChunk{err: err}
				return
			}

			ch <- value
		}
	}()

	return ch, nil
}

type location struct {
	scheme     string
	identifier string
	data       string
	err        error
}

func scanLocations(rows *sql.Rows, queryErr error) (<-chan location, error) {
	if queryErr != nil {
		return nil, queryErr
	}

	ch := make(chan location)
	go func() {
		defer close(ch)

		defer func() {
			if err := basestore.CloseRows(rows, nil); err != nil {
				ch <- location{err: err}
			}
		}()

		for rows.Next() {
			var value location
			if err := rows.Scan(&value.scheme, &value.identifier, &value.data); err != nil {
				ch <- location{err: err}
				return
			}

			ch <- value
		}
	}()

	return ch, nil
}
