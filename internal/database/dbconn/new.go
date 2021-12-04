package dbconn

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/migrations"
)

// NewFrontendDB creates a new connection to the frontend database. After successful connection,
// the schema version of the database will be compared against an expected version and migrations
// may be run (taking an advisory lock to ensure exclusive access).
//
// TEMPORARY: The migrate flag controls whether or not migrations/version checks are performed on
// the version. When false, we give back a handle without running any migrations and assume that
// the database schema is up to date.
//
// This connection is not expected to be closed but last the life of the calling application.
func NewFrontendDB(dsn, appName string, migrate bool) (*sql.DB, error) {
	migrations := []*migrations.Schema{migrations.Frontend}
	if !migrate {
		migrations = nil
	}

	db, _, err := connect(dsn, appName, "frontend", migrations)
	return db, err
}

// NewCodeIntelDB creates a new connection to the codeintel database. After successful connection,
// the schema version of the database will be compared against an expected version and migrations
// may be run (taking an advisory lock to ensure exclusive access).
//
// TEMPORARY: The migrate flag controls whether or not migrations/version checks are performed on
// the version. When false, we give back a handle without running any migrations and assume that
// the database schema is up to date.
//
// This connection is not expected to be closed but last the life of the calling application.
func NewCodeIntelDB(dsn, appName string, migrate bool) (*sql.DB, error) {
	migrations := []*migrations.Schema{migrations.CodeIntel}
	if !migrate {
		migrations = nil
	}

	db, _, err := connect(dsn, appName, "codeintel", migrations)
	return db, err
}

// NewCodeInsightsDB creates a new connection to the codeinsights database. After successful
// connection, the schema version of the database will be compared against an expected version and
// migrations may be run (taking an advisory lock to ensure exclusive access).
//
// TEMPORARY: The migrate flag controls whether or not migrations/version checks are performed on
// the version. When false, we give back a handle without running any migrations and assume that
// the database schema is up to date.
//
// This connection is not expected to be closed but last the life of the calling application.
func NewCodeInsightsDB(dsn, appName string, migrate bool) (*sql.DB, error) {
	migrations := []*migrations.Schema{migrations.CodeInsights}
	if !migrate {
		migrations = nil
	}

	db, _, err := connect(dsn, appName, "codeinsight", migrations)
	return db, err
}
