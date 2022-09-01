package codeintel

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type documentColumnSplitMigrator struct {
	serializer *serializer
}

// NewDocumentColumnSplitMigrator creates a new Migrator instance that reads records from
// the lsif_data_documents table with a schema version of 2 and unsets the payload in favor
// of populating the new ranges, hovers, monikers, packages, and diagnostics columns. Updated
// records will have a schema version of 3.
func NewDocumentColumnSplitMigrator(store *basestore.Store, batchSize int) *migrator {
	driver := &documentColumnSplitMigrator{
		serializer: newSerializer(),
	}

	return newMigrator(store, driver, migratorOptions{
		tableName:     "lsif_data_documents",
		targetVersion: 3,
		batchSize:     batchSize,
		fields: []fieldSpec{
			{name: "path", postgresType: "text not null", primaryKey: true},
			{name: "data", postgresType: "bytea"},
			{name: "ranges", postgresType: "bytea"},
			{name: "hovers", postgresType: "bytea"},
			{name: "monikers", postgresType: "bytea"},
			{name: "packages", postgresType: "bytea"},
			{name: "diagnostics", postgresType: "bytea"},
		},
	})
}

func (m *documentColumnSplitMigrator) ID() int                 { return 7 }
func (m *documentColumnSplitMigrator) Interval() time.Duration { return time.Second }

// MigrateRowUp reads the payload of the given row and returns an updateSpec on how to
// modify the record to conform to the new schema.
func (m *documentColumnSplitMigrator) MigrateRowUp(scanner scanner) ([]any, error) {
	var path string
	var rawData, ignored []byte

	if err := scanner.Scan(
		&path,
		&rawData,
		&ignored, // ranges
		&ignored, // hovers
		&ignored, // monikers
		&ignored, // packages
		&ignored, // diagnostics
	); err != nil {
		return nil, err
	}

	decoded, err := m.serializer.UnmarshalLegacyDocumentData(rawData)
	if err != nil {
		return nil, err
	}
	encoded, err := m.serializer.MarshalDocumentData(decoded)
	if err != nil {
		return nil, err
	}

	return []any{
		path,
		nil,                        // data
		encoded.Ranges,             // ranges
		encoded.HoverResults,       // hovers
		encoded.Monikers,           // monikers
		encoded.PackageInformation, // packages
		encoded.Diagnostics,        // diagnostics
	}, nil
}

// MigrateRowDown recombines the split payloads into a single column to undo the migration
// up direction.
func (m *documentColumnSplitMigrator) MigrateRowDown(scanner scanner) ([]any, error) {
	var path string
	var rawData []byte
	var encoded MarshalledDocumentData

	if err := scanner.Scan(
		&path,
		&rawData,
		&encoded.Ranges,
		&encoded.HoverResults,
		&encoded.Monikers,
		&encoded.PackageInformation,
		&encoded.Diagnostics,
	); err != nil {
		return nil, err
	}

	decoded, err := m.serializer.UnmarshalDocumentData(encoded)
	if err != nil {
		return nil, err
	}
	reencoded, err := m.serializer.MarshalLegacyDocumentData(decoded)
	if err != nil {
		return nil, err
	}

	return []any{
		path,
		reencoded, // data
		nil,       // ranges
		nil,       // hovers
		nil,       // monikers
		nil,       // packages
		nil,       // diagnostics
	}, nil
}
