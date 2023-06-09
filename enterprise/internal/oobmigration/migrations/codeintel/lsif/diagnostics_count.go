package lsif

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type diagnosticsCountMigrator struct {
	serializer *serializer
}

// NewDiagnosticsCountMigrator creates a new Migrator instance that reads records from
// the lsif_data_documents table with a schema version of 1 and populates that record's
// (new) num_diagnostics column. Updated records will have a schema version of 2.
func NewDiagnosticsCountMigrator(store *basestore.Store, batchSize, numRoutines int) *migrator {
	driver := &diagnosticsCountMigrator{
		serializer: newSerializer(),
	}

	return newMigrator(store, driver, migratorOptions{
		tableName:     "lsif_data_documents",
		targetVersion: 2,
		batchSize:     batchSize,
		numRoutines:   numRoutines,
		fields: []fieldSpec{
			{name: "path", postgresType: "text not null", primaryKey: true},
			{name: "data", postgresType: "bytea", readOnly: true},
			{name: "num_diagnostics", postgresType: "integer not null", updateOnly: true},
		},
	})
}

func (m *diagnosticsCountMigrator) ID() int                 { return 1 }
func (m *diagnosticsCountMigrator) Interval() time.Duration { return time.Second }

// MigrateRowUp reads the payload of the given row and returns an updateSpec on how to
// modify the record to conform to the new schema.
func (m *diagnosticsCountMigrator) MigrateRowUp(scanner dbutil.Scanner) ([]any, error) {
	var path string
	var rawData []byte

	if err := scanner.Scan(&path, &rawData); err != nil {
		return nil, err
	}

	data, err := m.serializer.UnmarshalLegacyDocumentData(rawData)
	if err != nil {
		return nil, err
	}

	return []any{path, len(data.Diagnostics)}, nil
}

// MigrateRowDown sets num_diagnostics back to zero to undo the migration up direction.
func (m *diagnosticsCountMigrator) MigrateRowDown(scanner dbutil.Scanner) ([]any, error) {
	var path string
	var rawData []byte

	if err := scanner.Scan(&path, &rawData); err != nil {
		return nil, err
	}

	return []any{path, 0}, nil
}
