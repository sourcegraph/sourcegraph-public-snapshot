package connections

import (
	"context"
	"database/sql"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func connectFrontendDB(observationCtx *observation.Context, dsn, appName string, validate, migrate bool) (*sql.DB, error) {
	schema := schemas.Frontend
	if !validate {
		schema = nil
	}

	return connect(observationCtx, dsn, appName, "frontend", schema, migrate)
}

func connectCodeIntelDB(observationCtx *observation.Context, dsn, appName string, validate, migrate bool) (*sql.DB, error) {
	schema := schemas.CodeIntel
	if !validate {
		schema = nil
	}

	return connect(observationCtx, dsn, appName, "codeintel", schema, migrate)
}

func connectCodeInsightsDB(observationCtx *observation.Context, dsn, appName string, validate, migrate bool) (*sql.DB, error) {
	schema := schemas.CodeInsights
	if !validate {
		schema = nil
	}

	return connect(observationCtx, dsn, appName, "codeinsights", schema, migrate)
}

func connect(observationCtx *observation.Context, dsn, appName, dbName string, schema *schemas.Schema, migrate bool) (*sql.DB, error) {
	db, err := dbconn.ConnectInternal(observationCtx.Logger, dsn, appName, dbName)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if closeErr := db.Close(); closeErr != nil {
				err = errors.Append(err, closeErr)
			}
		}
	}()

	if schema != nil {
		if err := validateSchema(observationCtx, db, schema, !migrate); err != nil {
			return nil, err
		}
	}

	return db, nil
}

func validateSchema(observationCtx *observation.Context, db *sql.DB, schema *schemas.Schema, validateOnly bool) error {
	ctx := context.Background()
	storeFactory := newStoreFactory(observationCtx)
	migrationRunner := runnerFromDB(observationCtx.Logger, storeFactory, db, schema)

	if err := migrationRunner.Validate(ctx, schema.Name); err != nil {
		outOfDateError := new(runner.SchemaOutOfDateError)
		if !errors.As(err, &outOfDateError) {
			return err
		}
		if !shouldMigrate(validateOnly) {
			return errors.Newf("database schema out of date")
		}

		return migrationRunner.Run(ctx, runner.Options{
			Operations: []runner.MigrationOperation{
				{
					SchemaName: schema.Name,
					Type:       runner.MigrationOperationTypeUpgrade,
				},
			},
		})
	}

	return nil
}

func shouldMigrate(validateOnly bool) bool {
	return !validateOnly || os.Getenv("SG_DEV_MIGRATE_ON_APPLICATION_STARTUP") != ""
}
