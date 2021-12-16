package connections

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// RawNewFrontendDB creates a new connection to the frontend database. This method does not ensure that the schema
// matches any expected shape.
//
// This method should not be used outside of migration utilities.
func RawNewFrontendDB(dsn, appName string, observationContext *observation.Context) (*sql.DB, error) {
	return connectFrontendDB(dsn, appName, false, false, observationContext)
}

// EnsureNewFrontendDB creates a new connection to the frontend database. After successful connection, the schema
// version of the database will be compared against an expected version. If it is not up to date, an error will be
// returned.
//
// If the SG_DEV_MIGRATE_ON_APPLICATION_STARTUP environment variable is set, which it is during local development,
// then this call will behave equivalently to MigrateNewFrontendDB, which will attempt to  upgrade the database. We
// only do this in dev as we don't want to introduce the migrator into an otherwise fast feedback cycle for developers.
func EnsureNewFrontendDB(dsn, appName string, observationContext *observation.Context) (*sql.DB, error) {
	return connectFrontendDB(dsn, appName, true, false, observationContext)
}

// MigrateNewFrontendDB creates a new connection to the frontend database. After successful connection, the schema version
// of the database will be compared against an expected version. If it is not up to date, the most recent schema version will
// be applied.
func MigrateNewFrontendDB(dsn, appName string, observationContext *observation.Context) (*sql.DB, error) {
	return connectFrontendDB(dsn, appName, true, true, observationContext)
}

// RawNewCodeIntelDB creates a new connection to the codeintel database. This method does not ensure that the schema
// matches any expected shape.
//
// This method should not be used outside of migration utilities.
func RawNewCodeIntelDB(dsn, appName string, observationContext *observation.Context) (*sql.DB, error) {
	return connectCodeIntelDB(dsn, appName, false, false, observationContext)
}

// EnsureNewCodeIntelDB creates a new connection to the codeintel database. After successful connection, the schema
// version of the database will be compared against an expected version. If it is not up to date, an error will be
// returned.
//
// If the SG_DEV_MIGRATE_ON_APPLICATION_STARTUP environment variable is set, which it is during local development,
// then this call will behave equivalently to MigrateNewCodeIntelDB, which will attempt to  upgrade the database. We
// only do this in dev as we don't want to introduce the migrator into an otherwise fast feedback cycle for developers.
func EnsureNewCodeIntelDB(dsn, appName string, observationContext *observation.Context) (*sql.DB, error) {
	return connectCodeIntelDB(dsn, appName, true, false, observationContext)
}

// MigrateNewCodeIntelDB creates a new connection to the codeintel database. After successful connection, the schema version
// of the database will be compared against an expected version. If it is not up to date, the most recent schema version will
// be applied.
func MigrateNewCodeIntelDB(dsn, appName string, observationContext *observation.Context) (*sql.DB, error) {
	return connectCodeIntelDB(dsn, appName, true, true, observationContext)
}

// RawNewCodeInsightsDB creates a new connection to the codeinsights database. This method does not ensure that the schema
// matches any expected shape.
//
// This method should not be used outside of migration utilities.
func RawNewCodeInsightsDB(dsn, appName string, observationContext *observation.Context) (*sql.DB, error) {
	return connectCodeInsightsDB(dsn, appName, false, false, observationContext)
}

// EnsureNewCodeInsightsDB creates a new connection to the codeinsights database. After successful connection, the schema
// version of the database will be compared against an expected version. If it is not up to date, an error will be returned.
//
// If the SG_DEV_MIGRATE_ON_APPLICATION_STARTUP environment variable is set, which it is during local development,
// then this call will behave equivalently to MigrateNewCodeInsightsDB, which will attempt to  upgrade the database. We
// only do this in dev as we don't want to introduce the migrator into an otherwise fast feedback cycle for  developers.
func EnsureNewCodeInsightsDB(dsn, appName string, observationContext *observation.Context) (*sql.DB, error) {
	return connectCodeInsightsDB(dsn, appName, true, false, observationContext)
}

// MigrateNewCodeInsightsDB creates a new connection to the codeinsights database. After successful connection, the schema
// version of the database will be compared against an expected version. If it is not up to date, the most recent schema version
// will be applied.
func MigrateNewCodeInsightsDB(dsn, appName string, observationContext *observation.Context) (*sql.DB, error) {
	return connectCodeInsightsDB(dsn, appName, true, true, observationContext)
}
