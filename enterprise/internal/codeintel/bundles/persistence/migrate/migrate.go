package migrate

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/postgres"
	sqlitestore "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/sqlite/store"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
)

func Migrate(ctx context.Context, dumpID int, filename string, to dbutil.DB) (err error) {
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

func migrateMeta(ctx context.Context, from *sqlitestore.Store, to dbutil.DB, dumpID int) error {
	numResultChunks, _, err := sqlitestore.ScanFirstInt(from.Query(ctx, sqlf.Sprintf("SELECT num_result_chunks FROM meta LIMIT 1")))
	if err != nil {
		return err
	}

	inserter := postgres.NewBatchInserter(to, "lsif_data_metadata", "dump_id", "num_result_chunks")

	if err := inserter.Insert(ctx, dumpID, numResultChunks); err != nil {
		return err
	}

	if err := inserter.Flush(ctx); err != nil {
		return errors.Wrap(err, "inserter.Flush")
	}

	return nil
}

func migrateDocuments(ctx context.Context, from *sqlitestore.Store, to dbutil.DB, dumpID int) error {
	rows, err := from.Query(ctx, sqlf.Sprintf("SELECT path, data FROM documents"))
	if err != nil {
		return err
	}
	defer func() { err = sqlitestore.CloseRows(rows, err) }()

	inserter := postgres.NewBatchInserter(to, "lsif_data_documents", "dump_id", "path", "data")

	for rows.Next() {
		var path string
		var data []byte
		if err := rows.Scan(&path, &data); err != nil {
			return err
		}

		if err := inserter.Insert(ctx, dumpID, path, data); err != nil {
			return err
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		return errors.Wrap(err, "inserter.Flush")
	}

	return nil
}

func migrateResultChunks(ctx context.Context, from *sqlitestore.Store, to dbutil.DB, dumpID int) error {
	rows, err := from.Query(ctx, sqlf.Sprintf("SELECT id, data FROM result_chunks"))
	if err != nil {
		return err
	}
	defer func() { err = sqlitestore.CloseRows(rows, err) }()

	inserter := postgres.NewBatchInserter(to, "lsif_data_result_chunks", "dump_id", "idx", "data")

	for rows.Next() {
		var index int
		var data []byte
		if err := rows.Scan(&index, &data); err != nil {
			return err
		}

		if err := inserter.Insert(ctx, dumpID, index, data); err != nil {
			return err
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		return errors.Wrap(err, "inserter.Flush")
	}

	return nil
}

func migrateDefinitions(ctx context.Context, from *sqlitestore.Store, to dbutil.DB, dumpID int) error {
	rows, err := from.Query(ctx, sqlf.Sprintf("SELECT scheme, identifier, data FROM definitions"))
	if err != nil {
		return err
	}
	defer func() { err = sqlitestore.CloseRows(rows, err) }()

	inserter := postgres.NewBatchInserter(to, "lsif_data_definitions", "dump_id", "scheme", "identifier", "data")

	for rows.Next() {
		var scheme string
		var identifier string
		var data []byte
		if err := rows.Scan(&scheme, &identifier, &data); err != nil {
			return err
		}

		if err := inserter.Insert(ctx, dumpID, scheme, identifier, data); err != nil {
			return err
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		return errors.Wrap(err, "inserter.Flush")
	}

	return nil
}

func migrateReferences(ctx context.Context, from *sqlitestore.Store, to dbutil.DB, dumpID int) error {
	rows, err := from.Query(ctx, sqlf.Sprintf(`SELECT scheme, identifier, data FROM "references"`))
	if err != nil {
		return err
	}
	defer func() { err = sqlitestore.CloseRows(rows, err) }()

	inserter := postgres.NewBatchInserter(to, "lsif_data_references", "dump_id", "scheme", "identifier", "data")

	for rows.Next() {
		var scheme string
		var identifier string
		var data []byte
		if err := rows.Scan(&scheme, &identifier, &data); err != nil {
			return err
		}

		if err := inserter.Insert(ctx, dumpID, scheme, identifier, data); err != nil {
			return err
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		return errors.Wrap(err, "inserter.Flush")
	}

	return nil
}
