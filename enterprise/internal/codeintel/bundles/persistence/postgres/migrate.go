package postgres

import (
	"context"
	"database/sql"

	"github.com/hashicorp/go-multierror"
	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	sqlitestore "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/sqlite/store"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
)

func MigrateBundleToPostgres(ctx context.Context, dumpID int, filename string, to dbutil.DB) (err error) {
	from, closer, err := sqlitestore.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := closer(); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}
	}()

	if err := migrateMeta(ctx, from, to, dumpID); err != nil {
		return errors.Wrap(err, "migrating meta")
	}
	if err := migrateDocuments(ctx, from, to, dumpID); err != nil {
		return errors.Wrap(err, "migrating documents")
	}
	if err := migrateResultChunks(ctx, from, to, dumpID); err != nil {
		return errors.Wrap(err, "migrating result chunks")
	}
	if err := migrateDefinitions(ctx, from, to, dumpID); err != nil {
		return errors.Wrap(err, "migrating definitions")
	}
	if err := migrateReferences(ctx, from, to, dumpID); err != nil {
		return errors.Wrap(err, "migrating references")
	}

	return nil
}

func migrateMeta(ctx context.Context, from *sqlitestore.Store, to dbutil.DB, dumpID int) (err error) {
	values, err := sqlitestore.ScanInts(from.Query(ctx, sqlf.Sprintf(`SELECT num_result_chunks FROM meta`)))
	if err != nil {
		return err
	}

	return withBatchInserter(ctx, to, "lsif_data_metadata", []string{"dump_id", "num_result_chunks"}, func(inserter *BatchInserter) error {
		for _, value := range values {
			if err := inserter.Insert(ctx, dumpID, value); err != nil {
				return err
			}
		}

		return nil
	})
}

func migrateDocuments(ctx context.Context, from *sqlitestore.Store, to dbutil.DB, dumpID int) (err error) {
	documents, err := scanDocuments(from.Query(ctx, sqlf.Sprintf(`SELECT path, data FROM documents`)))
	if err != nil {
		return err
	}

	return withBatchInserter(ctx, to, "lsif_data_documents", []string{"dump_id", "path", "data"}, func(inserter *BatchInserter) error {
		for document := range documents {
			if err := inserter.Insert(ctx, dumpID, document.path, document.data); err != nil {
				return err
			}
		}

		return nil
	})
}

func migrateResultChunks(ctx context.Context, from *sqlitestore.Store, to dbutil.DB, dumpID int) (err error) {
	resultChunks, err := scanResultChunks(from.Query(ctx, sqlf.Sprintf(`SELECT id, data FROM result_chunks`)))
	if err != nil {
		return err
	}

	return withBatchInserter(ctx, to, "lsif_data_result_chunks", []string{"dump_id", "idx", "data"}, func(inserter *BatchInserter) error {
		for resultChunk := range resultChunks {
			if err := inserter.Insert(ctx, dumpID, resultChunk.index, resultChunk.data); err != nil {
				return err
			}
		}

		return nil
	})
}

func migrateDefinitions(ctx context.Context, from *sqlitestore.Store, to dbutil.DB, dumpID int) (err error) {
	locations, err := scanLocations(from.Query(ctx, sqlf.Sprintf(`SELECT scheme, identifier, data FROM definitions`)))
	if err != nil {
		return err
	}

	return withBatchInserter(ctx, to, "lsif_data_definitions", []string{"dump_id", "scheme", "identifier", "data"}, func(inserter *BatchInserter) error {
		for location := range locations {
			if err := inserter.Insert(ctx, dumpID, location.scheme, location.identifier, location.data); err != nil {
				return err
			}
		}

		return nil
	})
}

func migrateReferences(ctx context.Context, from *sqlitestore.Store, to dbutil.DB, dumpID int) (err error) {
	locations, err := scanLocations(from.Query(ctx, sqlf.Sprintf(`SELECT scheme, identifier, data FROM "references"`)))
	if err != nil {
		return err
	}

	return withBatchInserter(ctx, to, "lsif_data_references", []string{"dump_id", "scheme", "identifier", "data"}, func(inserter *BatchInserter) error {
		for location := range locations {
			if err := inserter.Insert(ctx, dumpID, location.scheme, location.identifier, location.data); err != nil {
				return err
			}
		}

		return nil
	})
}

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
