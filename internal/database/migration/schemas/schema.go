package schemas

import (
	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
)

// Schema describes a schema in one of our Postgres(-like) databases.
type Schema struct {
	// Name is the name of the schema.
	Name string

	// MigrationsTableName is the name of the table that tracks the schema version.
	MigrationsTableName string

	// Definitions describes the parsed migration assets of the schema.
	Definitions *definition.Definitions
}
