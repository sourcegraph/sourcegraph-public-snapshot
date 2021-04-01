package migration

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type diagnosticsCountMigrator struct {
	serializer *lsifstore.Serializer
}

// NewDiagnosticsCountMigrator creates a new Migrator instance that reads records from
// the lsif_data_document table with a schema version of 1 and populates that record's
// (new) num_diagnostics column. Updated records will have a schema version of 2.
func NewDiagnosticsCountMigrator(store *lsifstore.Store, batchSize int) oobmigration.Migrator {
	driver := &diagnosticsCountMigrator{
		serializer: lsifstore.NewSerializer(),
	}

	return newMigrator(store, driver, migratorOptions{
		tableName:       "lsif_data_documents",
		selectionFields: []string{"path", "data"},
		targetVersion:   2,
		batchSize:       batchSize,
	})
}

// MigrateRowUp reads the payload of the given row and returns an updateSpec on how to
// modify the record to conform to the new schema.
func (m *diagnosticsCountMigrator) MigrateRowUp(scanner scanner) (updateSpec, error) {
	var dumpID int
	var path string
	var rawData []byte

	if err := scanner.Scan(&dumpID, &path, &rawData); err != nil {
		return updateSpec{}, err
	}

	data, err := m.serializer.UnmarshalDocumentData(rawData)
	if err != nil {
		return updateSpec{}, err
	}

	return updateSpec{
		DumpID:      dumpID,
		Conditions:  map[string]interface{}{"path": path},
		Assignments: map[string]interface{}{"num_diagnostics": len(data.Diagnostics)},
	}, nil
}

// MigrateRowDown sets num_diagnostics back to zero to undo the migration up direction.
func (m *diagnosticsCountMigrator) MigrateRowDown(scanner scanner) (updateSpec, error) {
	var dumpID int
	var path string
	var rawData []byte

	if err := scanner.Scan(&dumpID, &path, &rawData); err != nil {
		return updateSpec{}, err
	}

	return updateSpec{
		DumpID:      dumpID,
		Conditions:  map[string]interface{}{"path": path},
		Assignments: map[string]interface{}{"num_diagnostics": 0},
	}, nil
}
