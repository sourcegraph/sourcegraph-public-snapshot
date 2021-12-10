package connections

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// EnsureNewFrontendDB creates a new connection to the frontend database. After successful connection, the schema
// version of the database will be compared against an expected version. If it is not up to date, an error will be
// returned.
//
// If the SG_DEV_MIGRATE_ON_APPLICATION_STARTUP environment variable is set, which it is during local development,
// then this call will behave equivalently to MigrateNewFrontendDB, which will attempt to  upgrade the database. We
// only do this in dev as we don't want to introduce the migrator into an otherwise fast feedback cycle for developers.
func EnsureNewFrontendDB(dsn, appName string, observationContext *observation.Context) (*sql.DB, error) {
	return connectFrontendDB(dsn, appName, false, observationContext)
}

// MigrateNewFrontendDB creates a new connection to the frontend database. After successful connection, the schema version
// of the database will be compared against an expected version. If it is not up to date, the most recent schema version will
// be applied.
func MigrateNewFrontendDB(dsn, appName string, observationContext *observation.Context) (*sql.DB, error) {
	return connectFrontendDB(dsn, appName, true, observationContext)
}

// EnsureNewCodeIntelDB creates a new connection to the codeintel database. After successful connection, the schema
// version of the database will be compared against an expected version. If it is not up to date, an error will be
// returned.
//
// If the SG_DEV_MIGRATE_ON_APPLICATION_STARTUP environment variable is set, which it is during local development,
// then this call will behave equivalently to MigrateNewCodeIntelDB, which will attempt to  upgrade the database. We
// only do this in dev as we don't want to introduce the migrator into an otherwise fast feedback cycle for developers.
func EnsureNewCodeIntelDB(dsn, appName string, observationContext *observation.Context) (*sql.DB, error) {
	return connectCodeIntelDB(dsn, appName, false, observationContext)
}

// MigrateNewCodeIntelDB creates a new connection to the codeintel database. After successful connection, the schema version
// of the database will be compared against an expected version. If it is not up to date, the most recent schema version will
// be applied.
func MigrateNewCodeIntelDB(dsn, appName string, observationContext *observation.Context) (*sql.DB, error) {
	return connectCodeIntelDB(dsn, appName, true, observationContext)
}

// EnsureNewCodeInsightsDB creates a new connection to the codeinsights database. After successful connection, the schema
// version of the database will be compared against an expected version. If it is not up to date, an error will be returned.
//
// If the SG_DEV_MIGRATE_ON_APPLICATION_STARTUP environment variable is set, which it is during local development,
// then this call will behave equivalently to MigrateNewCodeInsightsDB, which will attempt to  upgrade the database. We
// only do this in dev as we don't want to introduce the migrator into an otherwise fast feedback cycle for  developers.
func EnsureNewCodeInsightsDB(dsn, appName string, observationContext *observation.Context) (*sql.DB, error) {
	return connectCodeInsightsDB(dsn, appName, false, observationContext)
}

// MigrateNewCodeInsightsDB creates a new connection to the codeinsights database. After successful connection, the schema
// version of the database will be compared against an expected version. If it is not up to date, the most recent schema version
// will be applied.
func MigrateNewCodeInsightsDB(dsn, appName string, observationContext *observation.Context) (*sql.DB, error) {
	return connectCodeInsightsDB(dsn, appName, true, observationContext)
}
