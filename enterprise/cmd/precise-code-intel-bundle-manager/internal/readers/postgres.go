package readers

import (
	"context"
	"database/sql"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/cache"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/postgres"
	sqlitestore "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/sqlite/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/sqlite/util"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
)

func migrateToPostgres(bundleDir string, storeCache cache.StoreCache, db *sql.DB, bundleFilenames []string) error {
	// TODO - filter out what already exists
	// TODO - move bundle files to another directory for explicit deletion

	log15.Info(
		"Migrating bundle data to Postgres in background",
		"numBundles", len(bundleFilenames),
	)

	for _, filename := range bundleFilenames {
		bundleID, err := strconv.Atoi(filepath.Base(filepath.Dir(filename)))
		if err != nil {
			log15.Error("Failed to extract bundle id from filename", "err", err, "filename", filename)
			continue
		}

		if err := migrateBundleToPostgres(context.Background(), bundleID, filename, db); err != nil {
			log15.Error("Failed to migrate bundle", "err", err, "filename", filename)
		}
	}

	log15.Info("Finished migration to Postgres")
	return nil
}

func migrateBundleToPostgres(ctx context.Context, dumpID int, filename string, to dbutil.DB) (err error) {
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
	values, err := sqlitestore.ScanInts(from.Query(ctx, sqlf.Sprintf("SELECT num_result_chunks FROM meta")))
	if err != nil {
		return err
	}

	return withBatchInserter(ctx, to, "lsif_data_metadata", []string{"dump_id", "num_result_chunks"}, func(inserter *postgres.BatchInserter) error {
		for _, value := range values {
			if err := inserter.Insert(ctx, dumpID, value); err != nil {
				return err
			}
		}

		return nil
	})
}

func migrateDocuments(ctx context.Context, from *sqlitestore.Store, to dbutil.DB, dumpID int) (err error) {
	documents, err := scanDocuments(from.Query(ctx, sqlf.Sprintf("SELECT path, data FROM documents")))
	if err != nil {
		return err
	}

	return withBatchInserter(ctx, to, "lsif_data_documents", []string{"dump_id", "path", "data"}, func(inserter *postgres.BatchInserter) error {
		for document := range documents {
			if err := inserter.Insert(ctx, dumpID, document.path, document.data); err != nil {
				return err
			}
		}

		return nil
	})
}

func migrateResultChunks(ctx context.Context, from *sqlitestore.Store, to dbutil.DB, dumpID int) (err error) {
	resultChunks, err := scanResultChunks(from.Query(ctx, sqlf.Sprintf("SELECT id, data FROM result_chunks")))
	if err != nil {
		return err
	}

	return withBatchInserter(ctx, to, "lsif_data_result_chunks", []string{"dump_id", "idx", "data"}, func(inserter *postgres.BatchInserter) error {
		for resultChunk := range resultChunks {
			if err := inserter.Insert(ctx, dumpID, resultChunk.index, resultChunk.data); err != nil {
				return err
			}
		}

		return nil
	})
}

func migrateDefinitions(ctx context.Context, from *sqlitestore.Store, to dbutil.DB, dumpID int) (err error) {
	locations, err := scanLocations(from.Query(ctx, sqlf.Sprintf("SELECT scheme, identifier, data FROM definitions")))
	if err != nil {
		return err
	}

	return withBatchInserter(ctx, to, "lsif_data_definitions", []string{"dump_id", "scheme", "identifier", "data"}, func(inserter *postgres.BatchInserter) error {
		for location := range locations {
			if err := inserter.Insert(ctx, dumpID, location.scheme, location.identifier, location.data); err != nil {
				return err
			}
		}

		return nil
	})
}

func migrateReferences(ctx context.Context, from *sqlitestore.Store, to dbutil.DB, dumpID int) (err error) {
	locations, err := scanLocations(from.Query(ctx, sqlf.Sprintf("SELECT scheme, identifier, data FROM references")))
	if err != nil {
		return err
	}

	return withBatchInserter(ctx, to, "lsif_data_references", []string{"dump_id", "scheme", "identifier", "data"}, func(inserter *postgres.BatchInserter) error {
		for location := range locations {
			if err := inserter.Insert(ctx, dumpID, location.scheme, location.identifier, location.data); err != nil {
				return err
			}
		}

		return nil
	})
}

var numWriterRoutines = runtime.GOMAXPROCS(0)

func withBatchInserter(ctx context.Context, db dbutil.DB, tableName string, columns []string, f func(inserter *postgres.BatchInserter) error) (err error) {
	return util.InvokeN(numWriterRoutines, func() (err error) {
		inserter := postgres.NewBatchInserter(ctx, db, tableName, columns...)
		defer func() {
			if flushErr := inserter.Flush(ctx); flushErr != nil {
				err = multierror.Append(err, errors.Wrap(flushErr, "inserter.Flush"))
			}
		}()

		return f(inserter)
	})
}
