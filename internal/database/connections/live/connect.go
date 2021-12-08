package connections

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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
func NewFrontendDB(dsn, appName string, migrate bool, observationContext *observation.Context) (*sql.DB, error) {
	migrations := []*schemas.Schema{schemas.Frontend}
	if !migrate {
		migrations = nil
	}

	db, _, err := dbconn.ConnectInternal(dsn, appName, "frontend", migrations)
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
func NewCodeIntelDB(dsn, appName string, migrate bool, observationContext *observation.Context) (*sql.DB, error) {
	migrations := []*schemas.Schema{schemas.CodeIntel}
	if !migrate {
		migrations = nil
	}

	db, _, err := dbconn.ConnectInternal(dsn, appName, "codeintel", migrations)
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
func NewCodeInsightsDB(dsn, appName string, migrate bool, observationContext *observation.Context) (*sql.DB, error) {
	migrations := []*schemas.Schema{schemas.CodeInsights}
	if !migrate {
		migrations = nil
	}

	db, _, err := dbconn.ConnectInternal(dsn, appName, "codeinsight", migrations)
	return db, err
}

func newStoreFactory(observationContext *observation.Context) func(db *sql.DB, migrationsTable string) Store {
	return func(db *sql.DB, migrationsTable string) Store {
		return store.NewWithDB(db, migrationsTable, store.NewOperations(observationContext))
	}
}
