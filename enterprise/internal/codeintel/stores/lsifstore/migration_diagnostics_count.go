package lsifstore

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type diagnosticsCountMigrator struct {
	store      *Store
	serializer *serializer
}

// DiagnosticsCountMigrationID is the primary key of the migration record an instance of
// diagnosticsCountMigrator handles. This is associated with the out-of-band migration
// record inserted in migrations/frontend/1528395786_diagnostic_counts_migration.up.sql.
const DiagnosticsCountMigrationID = 1

// NewDiagnosticsCountMigrator creates a new Migrator instance that reads the documents
// table and populates their num_diagnostics value based on their decoded payload. This
// will update rows with a schema_version of 1, and will set the row's schema version
// to 2 after processing.
func NewDiagnosticsCountMigrator(store *Store) oobmigration.Migrator {
	return &diagnosticsCountMigrator{
		store:      store,
		serializer: newSerializer(),
	}
}

// Progress returns the ratio of migrated records to total records. Any record with a
// schema version of two or greater is considered migrated.
func (m *diagnosticsCountMigrator) Progress(ctx context.Context) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(diagnosticsCountMigratorProgressQuery)))
	if err != nil {
		return 0, err
	}

	return progress, nil
}

const diagnosticsCountMigratorProgressQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/migration_diagnostics_count.go:Progress
SELECT CASE c2.count WHEN 0 THEN 1 ELSE cast(c1.count as float) / cast(c2.count as float) END FROM
	(SELECT COUNT(*) as count FROM lsif_data_documents WHERE schema_version >= 2) c1,
	(SELECT COUNT(*) as count FROM lsif_data_documents) c2
`

// DiagnosticCountMigrationBatchSize is the number of records that should be selected for
// update in a single invocation of Up.
const DiagnosticCountMigrationBatchSize = 1000

// Up reads records with a schema version of 1, decodes their data payload, then writes
// the number of diagnostic in the payload back to the record. The schema version of the
// modified row will be bumped to 2.
func (m *diagnosticsCountMigrator) Up(ctx context.Context) error {
	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	counts, err := m.scanDiagnosticCounts(tx.Query(ctx, sqlf.Sprintf(diagnosticsCountMigratorSelectQuery, DiagnosticCountMigrationBatchSize)))
	if err != nil {
		return err
	}

	for _, c := range counts {
		if _, err := basestore.ScanInts(tx.Query(ctx, sqlf.Sprintf(diagnosticsCountMigratorUpdateQuery, c.NumDiagnostics, c.DumpID, c.Path))); err != nil {
			return err
		}
	}

	return nil
}

const diagnosticsCountMigratorSelectQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/migration_diagnostics_count.go:Up
SELECT dump_id, path, data
FROM lsif_data_documents
WHERE schema_version = 1
LIMIT %s
FOR UPDATE SKIP LOCKED
`

const diagnosticsCountMigratorUpdateQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/migration_diagnostics_count.go:Up
UPDATE lsif_data_documents
SET num_diagnostics = %s, schema_version = 2
WHERE dump_id = %s AND path = %s
`

// Down is a no-op as the up migration is non-destructive and previous clients can still
// read. We do reset the schema_version, though, so we can have Progress report 0% in the
// event of a downgrade.
//
// This mainly exists to aid in UI/UX. This is not strictly necessary for this migration,
// as there are no previous migrations that depend on a schema version of one. This is
// mainly done to prevent copy and paste errors errors in the future when we have chains
// of schema versions depending on one another.
func (m *diagnosticsCountMigrator) Down(ctx context.Context) error {
	return m.store.Exec(ctx, sqlf.Sprintf(diagnosticsCountMigratorDownQuery, DiagnosticCountMigrationBatchSize))
}

const diagnosticsCountMigratorDownQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/migration_diagnostics_count.go:Down
WITH batch AS (
	SELECT dump_id, path
	FROM lsif_data_documents
	WHERE schema_version = 2
	LIMIT %d
	FOR UPDATE SKIP LOCKED
)
UPDATE lsif_data_documents SET schema_version = 1 WHERE (dump_id, path) IN (SELECT * FROM batch)
`

type diagnosticCount struct {
	DumpID         int
	Path           string
	NumDiagnostics int
}

// scanDiagnosticCounts scans a slice of diagnosticCount values from the return value
// of `*Store.query`.
func (m *diagnosticsCountMigrator) scanDiagnosticCounts(rows *sql.Rows, queryErr error) (_ []diagnosticCount, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var values []diagnosticCount
	for rows.Next() {
		var record diagnosticCount
		var rawData []byte
		if err := rows.Scan(&record.DumpID, &record.Path, &rawData); err != nil {
			return nil, err
		}

		data, err := m.serializer.UnmarshalDocumentData(rawData)
		if err != nil {
			return nil, err
		}
		record.NumDiagnostics = len(data.Diagnostics)

		values = append(values, record)
	}

	return values, nil
}
