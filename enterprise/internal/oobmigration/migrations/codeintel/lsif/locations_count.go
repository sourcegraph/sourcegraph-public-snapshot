package lsif

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

func NewDefinitionLocationsCountMigrator(store *basestore.Store, batchSize, numRoutines int) *migrator {
	return newLocationsCountMigrator(store, 4, time.Second, "lsif_data_definitions", batchSize, numRoutines)
}

func NewReferencesLocationsCountMigrator(store *basestore.Store, batchSize, numRoutines int) *migrator {
	return newLocationsCountMigrator(store, 5, time.Second, "lsif_data_references", batchSize, numRoutines)
}

type locationsCountMigrator struct {
	id         int
	interval   time.Duration
	serializer *serializer
}

// newLocationsCountMigrator creates a new Migrator instance that reads records from
// the given table with a schema version of 1 and populates that record's (new) num_locations
// column. Updated records will have a schema version of 2.
func newLocationsCountMigrator(store *basestore.Store, id int, interval time.Duration, tableName string, batchSize, numRoutines int) *migrator {
	driver := &locationsCountMigrator{
		id:         id,
		interval:   interval,
		serializer: newSerializer(),
	}

	return newMigrator(store, driver, migratorOptions{
		tableName:     tableName,
		targetVersion: 2,
		batchSize:     batchSize,
		numRoutines:   numRoutines,
		fields: []fieldSpec{
			{name: "scheme", postgresType: "text not null", primaryKey: true},
			{name: "identifier", postgresType: "text not null", primaryKey: true},
			{name: "data", postgresType: "bytea", readOnly: true},
			{name: "num_locations", postgresType: "integer not null", updateOnly: true},
		},
	})
}

func (m *locationsCountMigrator) ID() int                 { return m.id }
func (m *locationsCountMigrator) Interval() time.Duration { return m.interval }

// MigrateRowUp reads the payload of the given row and returns an updateSpec on how to
// modify the record to conform to the new schema.
func (m *locationsCountMigrator) MigrateRowUp(scanner dbutil.Scanner) ([]any, error) {
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
func (m *locationsCountMigrator) MigrateRowDown(scanner dbutil.Scanner) ([]any, error) {
	var scheme, identifier string
	var rawData []byte

	if err := scanner.Scan(&scheme, &identifier, &rawData); err != nil {
		return nil, err
	}

	return []any{scheme, identifier, 0}, nil
}
