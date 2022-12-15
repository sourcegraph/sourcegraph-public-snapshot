package connections

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// RawNewFrontendDB creates a new connection to the frontend database. This method does not ensure that the schema
// matches any expected shape.
//
// This method should not be used outside of migration utilities.
func RawNewFrontendDB(observationCtx *observation.Context, dsn, appName string) (*sql.DB, error) {
	return connectFrontendDB(observationCtx, dsn, appName, false, false)
}

// EnsureNewFrontendDB creates a new connection to the frontend database. After successful connection, the schema
// version of the database will be compared against an expected version. If it is not up to date, an error will be
// returned.
//
// If the SG_DEV_MIGRATE_ON_APPLICATION_STARTUP environment variable is set, which it is during local development,
// then this call will behave equivalently to MigrateNewFrontendDB, which will attempt to  upgrade the database. We
// only do this in dev as we don't want to introduce the migrator into an otherwise fast feedback cycle for developers.
func EnsureNewFrontendDB(observationCtx *observation.Context, dsn, appName string) (*sql.DB, error) {
	return connectFrontendDB(observationCtx, dsn, appName, true, false)
}

// MigrateNewFrontendDB creates a new connection to the frontend database. After successful connection, the schema version
// of the database will be compared against an expected version. If it is not up to date, the most recent schema version will
// be applied.
func MigrateNewFrontendDB(observationCtx *observation.Context, dsn, appName string) (*sql.DB, error) {
	return connectFrontendDB(observationCtx, dsn, appName, true, true)
}

// RawNewCodeIntelDB creates a new connection to the codeintel database. This method does not ensure that the schema
// matches any expected shape.
//
// This method should not be used outside of migration utilities.
func RawNewCodeIntelDB(observationCtx *observation.Context, dsn, appName string) (*sql.DB, error) {
	return connectCodeIntelDB(observationCtx, dsn, appName, false, false)
}

// EnsureNewCodeIntelDB creates a new connection to the codeintel database. After successful connection, the schema
// version of the database will be compared against an expected version. If it is not up to date, an error will be
// returned.
//
// If the SG_DEV_MIGRATE_ON_APPLICATION_STARTUP environment variable is set, which it is during local development,
// then this call will behave equivalently to MigrateNewCodeIntelDB, which will attempt to  upgrade the database. We
// only do this in dev as we don't want to introduce the migrator into an otherwise fast feedback cycle for developers.
func EnsureNewCodeIntelDB(observationCtx *observation.Context, dsn, appName string) (*sql.DB, error) {
	return connectCodeIntelDB(observationCtx, dsn, appName, true, false)
}

// MigrateNewCodeIntelDB creates a new connection to the codeintel database. After successful connection, the schema version
// of the database will be compared against an expected version. If it is not up to date, the most recent schema version will
// be applied.
func MigrateNewCodeIntelDB(observationCtx *observation.Context, dsn, appName string) (*sql.DB, error) {
	return connectCodeIntelDB(observationCtx, dsn, appName, true, true)
}

// RawNewCodeInsightsDB creates a new connection to the codeinsights database. This method does not ensure that the schema
// matches any expected shape.
//
// This method should not be used outside of migration utilities.
func RawNewCodeInsightsDB(observationCtx *observation.Context, dsn, appName string) (*sql.DB, error) {
	return connectCodeInsightsDB(observationCtx, dsn, appName, false, false)
}

// EnsureNewCodeInsightsDB creates a new connection to the codeinsights database. After successful connection, the schema
// version of the database will be compared against an expected version. If it is not up to date, an error will be returned.
//
// If the SG_DEV_MIGRATE_ON_APPLICATION_STARTUP environment variable is set, which it is during local development,
// then this call will behave equivalently to MigrateNewCodeInsightsDB, which will attempt to  upgrade the database. We
// only do this in dev as we don't want to introduce the migrator into an otherwise fast feedback cycle for  developers.
func EnsureNewCodeInsightsDB(observationCtx *observation.Context, dsn, appName string) (*sql.DB, error) {
	return connectCodeInsightsDB(observationCtx, dsn, appName, true, false)
}

// MigrateNewCodeInsightsDB creates a new connection to the codeinsights database. After successful connection, the schema
// version of the database will be compared against an expected version. If it is not up to date, the most recent schema version
// will be applied.
func MigrateNewCodeInsightsDB(observationCtx *observation.Context, dsn, appName string) (*sql.DB, error) {
	return connectCodeInsightsDB(observationCtx, dsn, appName, true, true)
}
