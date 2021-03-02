package migrations

import (
	"embed"
)

// MigrationsFS contains all the migrations in this directory in
// an embedded file system.
//go:embed *.sql
var MigrationsFS embed.FS
