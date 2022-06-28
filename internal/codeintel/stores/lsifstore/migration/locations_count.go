package migration

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
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
		tableName:     tableName,
		targetVersion: 2,
		batchSize:     batchSize,
		fields: []fieldSpec{
			{name: "scheme", postgresType: "text not null", primaryKey: true},
			{name: "identifier", postgresType: "text not null", primaryKey: true},
			{name: "data", postgresType: "bytea", readOnly: true},
			{name: "num_locations", postgresType: "integer not null", updateOnly: true},
		},
	})
}

// MigrateRowUp reads the payload of the given row and returns an updateSpec on how to
// modify the record to conform to the new schema.
func (m *locationsCountMigrator) MigrateRowUp(scanner scanner) ([]any, error) {
	var scheme, identifier string
	var rawData []byte

	if err := scanner.Scan(&scheme, &identifier, &rawData); err != nil {
		return nil, err
	}

	data, err := m.serializer.UnmarshalLocations(rawData)
	if err != nil {
		return nil, err
	}

	return []any{scheme, identifier, len(data)}, nil
}

// MigrateRowDown sets num_locations back to zero to undo the migration up direction.
func (m *locationsCountMigrator) MigrateRowDown(scanner scanner) ([]any, error) {
	var scheme, identifier string
	var rawData []byte

	if err := scanner.Scan(&scheme, &identifier, &rawData); err != nil {
		return nil, err
	}

	return []any{scheme, identifier, 0}, nil
}
