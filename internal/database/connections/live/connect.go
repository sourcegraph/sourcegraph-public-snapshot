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

func connectFrontendDB(dsn, appName string, validate, migrate bool, observationContext *observation.Context) (*sql.DB, error) {
	schema := schemas.FrontendDefinition
	if !validate {
		schema = nil
	}

	return connect(dsn, appName, "frontend", schema, migrate, observationContext)
}

func connectCodeIntelDB(dsn, appName string, validate, migrate bool, observationContext *observation.Context) (*sql.DB, error) {
	schema := schemas.CodeIntelDefinition
	if !validate {
		schema = nil
	}

	return connect(dsn, appName, "codeintel", schema, migrate, observationContext)
}

func connectCodeInsightsDB(dsn, appName string, validate, migrate bool, observationContext *observation.Context) (*sql.DB, error) {
	schema := schemas.CodeInsightsDefinition
	if !validate {
		schema = nil
	}

	return connect(dsn, appName, "codeinsights", schema, migrate, observationContext)
}

func connect(dsn, appName, dbName string, schema *schemas.Schema, migrate bool, observationContext *observation.Context) (*sql.DB, error) {
	db, err := dbconn.ConnectInternal(dsn, appName, dbName)
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
		if err := validateSchema(db, schema, !migrate, observationContext); err != nil {
			return nil, err
		}
	}

	return db, nil
}

func validateSchema(db *sql.DB, schema *schemas.Schema, validateOnly bool, observationContext *observation.Context) error {
	ctx := context.Background()
	storeFactory := newStoreFactory(observationContext)
	migrationRunner := runnerFromDB(storeFactory, db, schema)

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
