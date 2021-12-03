package migrations

import (
	"io/fs"

	"github.com/sourcegraph/sourcegraph/migrations"
)

// Schema describe a schema in one of our Postgres(-like) databases.
type Schema struct {
	// Name is the name of the schema.
	Name string

	// MigrationsTableName is the name of the table that tracks the schema version.
	MigrationsTableName string

	// FS describes the raw migration assets of the schema.
	FS fs.FS
}

var (
	Frontend = &Schema{
		Name:                "frontend",
		MigrationsTableName: "schema_migrations",
		FS:                  migrations.Frontend,
	}

	CodeIntel = &Schema{
		Name:                "codeintel",
		MigrationsTableName: "codeintel_schema_migrations",
		FS:                  migrations.CodeIntel,
	}

	CodeInsights = &Schema{
		Name:                "codeinsights",
		MigrationsTableName: "codeinsights_schema_migrations",
		FS:                  migrations.CodeInsights,
	}
)
