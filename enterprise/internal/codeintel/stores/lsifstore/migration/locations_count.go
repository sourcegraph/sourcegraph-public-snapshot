package migration

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type locationsCountMigrator struct {
	serializer *lsifstore.Serializer
}

// NewLocationsCountMigrator creates a new Migrator instance that reads records from
// the given table with a schema version of 1 and populates that record's (new) num_locations
// column. Updated records will have a schema version of 2.
func NewLocationsCountMigrator(store *lsifstore.Store, tableName string, batchSize int) oobmigration.Migrator {
	driver := &locationsCountMigrator{
		serializer: lsifstore.NewSerializer(),
	}

	return newMigrator(store, driver, migratorOptions{
		tableName:       tableName,
		selectionFields: []string{"scheme", "identifier", "data"},
		targetVersion:   2,
		batchSize:       batchSize,
	})
}

// MigrateRowUp reads the payload of the given row and returns an updateSpec on how to
// modify the record to conform to the new schema.
func (m *locationsCountMigrator) MigrateRowUp(scanner scanner) (updateSpec, error) {
	var dumpID int
	var scheme, identifier string
	var rawData []byte

	if err := scanner.Scan(&dumpID, &scheme, &identifier, &rawData); err != nil {
		return updateSpec{}, err
	}

	data, err := m.serializer.UnmarshalLocations(rawData)
	if err != nil {
		return updateSpec{}, err
	}

	return updateSpec{
		DumpID:      dumpID,
		Conditions:  map[string]interface{}{"scheme": scheme, "identifier": identifier},
		Assignments: map[string]interface{}{"num_locations": len(data)},
	}, nil
}

// MigrateRowDown sets num_locations back to zero to undo the migration up direction.
func (m *locationsCountMigrator) MigrateRowDown(scanner scanner) (updateSpec, error) {
	var dumpID int
	var scheme, identifier string
	var rawData []byte

	if err := scanner.Scan(&dumpID, &scheme, &identifier, &rawData); err != nil {
		return updateSpec{}, err
	}

	return updateSpec{
		DumpID:      dumpID,
		Conditions:  map[string]interface{}{"scheme": scheme, "identifier": identifier},
		Assignments: map[string]interface{}{"num_locations": 0},
	}, nil
}
