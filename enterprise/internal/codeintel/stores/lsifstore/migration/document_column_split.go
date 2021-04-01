package migration

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type documentColumnSplitMigrator struct {
	serializer *lsifstore.Serializer
}

// NewDocumentColumnSplitMigrator creates a new Migrator instance that reads records from
// the lsif_data_documents table with a schema version of 2 and unsets the payload in favor
// of populating the new ranges, hovers, monikers, packages, and diagnostics columns. Updated
// records will have a schema version of 3.
func NewDocumentColumnSplitMigrator(store *lsifstore.Store, batchSize int) oobmigration.Migrator {
	driver := &documentColumnSplitMigrator{
		serializer: lsifstore.NewSerializer(),
	}

	return newMigrator(store, driver, migratorOptions{
		tableName: "lsif_data_documents",
		selectionFields: []string{
			"path",
			"data",
			"ranges",
			"hovers",
			"monikers",
			"packages",
			"diagnostics",
		},
		targetVersion: 3,
		batchSize:     batchSize,
	})
}

// MigrateRowUp reads the payload of the given row and returns an updateSpec on how to
// modify the record to conform to the new schema.
func (m *documentColumnSplitMigrator) MigrateRowUp(scanner scanner) (updateSpec, error) {
	var dumpID int
	var path string
	var rawData, ignored []byte

	if err := scanner.Scan(
		&dumpID,
		&path,
		&rawData,
		&ignored,
		&ignored,
		&ignored,
		&ignored,
		&ignored,
	); err != nil {
		return updateSpec{}, err
	}

	decoded, err := m.serializer.UnmarshalLegacyDocumentData(rawData)
	if err != nil {
		return updateSpec{}, err
	}
	encoded, err := m.serializer.MarshalDocumentData(decoded)
	if err != nil {
		return updateSpec{}, err
	}

	return updateSpec{
		DumpID:     dumpID,
		Conditions: map[string]interface{}{"path": path},
		Assignments: map[string]interface{}{
			"data":        nil,
			"ranges":      encoded.Ranges,
			"hovers":      encoded.HoverResults,
			"monikers":    encoded.Monikers,
			"packages":    encoded.PackageInformation,
			"diagnostics": encoded.Diagnostics,
		},
	}, nil
}

// MigrateRowDown sets num_diagnostics back to zero to undo the migration up direction.
func (m *documentColumnSplitMigrator) MigrateRowDown(scanner scanner) (updateSpec, error) {
	var dumpID int
	var path string
	var rawData []byte
	var encoded lsifstore.MarshalledDocumentData

	if err := scanner.Scan(
		&dumpID,
		&path,
		&rawData,
		&encoded.Ranges,
		&encoded.HoverResults,
		&encoded.Monikers,
		&encoded.PackageInformation,
		&encoded.Diagnostics,
	); err != nil {
		return updateSpec{}, err
	}

	decoded, err := m.serializer.UnmarshalDocumentData(encoded)
	if err != nil {
		return updateSpec{}, err
	}
	reencoded, err := m.serializer.MarshalLegacyDocumentData(decoded)
	if err != nil {
		return updateSpec{}, err
	}

	return updateSpec{
		DumpID:     dumpID,
		Conditions: map[string]interface{}{"path": path},
		Assignments: map[string]interface{}{
			"data":        reencoded,
			"ranges":      nil,
			"hovers":      nil,
			"monikers":    nil,
			"packages":    nil,
			"diagnostics": nil,
		},
	}, nil
}
