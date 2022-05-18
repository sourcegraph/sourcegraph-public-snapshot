package migration

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type diagnosticsCountMigrator struct {
	serializer *lsifstore.Serializer
}

// NewDiagnosticsCountMigrator creates a new Migrator instance that reads records from
// the lsif_data_documents table with a schema version of 1 and populates that record's
// (new) num_diagnostics column. Updated records will have a schema version of 2.
func NewDiagnosticsCountMigrator(store *lsifstore.Store, batchSize int) oobmigration.Migrator {
	driver := &diagnosticsCountMigrator{
		serializer: lsifstore.NewSerializer(),
	}

	return newMigrator(store, driver, migratorOptions{
		tableName:     "lsif_data_documents",
		targetVersion: 2,
		batchSize:     batchSize,
		fields: []fieldSpec{
			{name: "path", postgresType: "text not null", primaryKey: true},
			{name: "data", postgresType: "bytea", readOnly: true},
			{name: "num_diagnostics", postgresType: "integer not null", updateOnly: true},
		},
	})
}

// MigrateRowUp reads the payload of the given row and returns an updateSpec on how to
// modify the record to conform to the new schema.
func (m *diagnosticsCountMigrator) MigrateRowUp(scanner scanner) ([]any, error) {
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
func (m *diagnosticsCountMigrator) MigrateRowDown(scanner scanner) ([]any, error) {
	var path string
	var rawData []byte

	if err := scanner.Scan(&path, &rawData); err != nil {
		return nil, err
	}

	return []any{path, 0}, nil
}
