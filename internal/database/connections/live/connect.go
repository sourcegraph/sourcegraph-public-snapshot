package connections

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// NewFrontendDB creates a new connection to the frontend database. After successful connection,
// the schema version of the database will be compared against an expected version and migrations
// may be run (taking an advisory lock to ensure exclusive access).
//
// TEMPORARY: The migrate flag controls whether or not migrations/version checks are performed on
// the version. When false, we give back a handle without running any migrations and assume that
// the database schema is up to date.
func NewFrontendDB(dsn, appName string, migrate bool, observationContext *observation.Context) (*sql.DB, error) {
	schema := schemas.Frontend
	if !migrate {
		schema = nil
	}

	return connect(dsn, appName, "frontend", schema, false, observationContext)
}

// NewCodeIntelDB creates a new connection to the codeintel database. After successful connection,
// the schema version of the database will be compared against an expected version and migrations
// may be run (taking an advisory lock to ensure exclusive access).
//
// TEMPORARY: The migrate flag controls whether or not migrations/version checks are performed on
// the version. When false, we give back a handle without running any migrations and assume that
// the database schema is up to date.
func NewCodeIntelDB(dsn, appName string, migrate bool, observationContext *observation.Context) (*sql.DB, error) {
	schema := schemas.CodeIntel
	if !migrate {
		schema = nil
	}

	return connect(dsn, appName, "codeintel", schema, false, observationContext)
}

// NewCodeInsightsDB creates a new connection to the codeinsights database. After successful
// connection, the schema version of the database will be compared against an expected version and
// migrations may be run (taking an advisory lock to ensure exclusive access).
//
// TEMPORARY: The migrate flag controls whether or not migrations/version checks are performed on
// the version. When false, we give back a handle without running any migrations and assume that
// the database schema is up to date.
func NewCodeInsightsDB(dsn, appName string, migrate bool, observationContext *observation.Context) (*sql.DB, error) {
	schema := schemas.CodeInsights
	if !migrate {
		schema = nil
	}

	return connect(dsn, appName, "codeinsight", schema, false, observationContext)
}

func connect(dsn, appName, dbName string, schema *schemas.Schema, validateOnly bool, observationContext *observation.Context) (*sql.DB, error) {
	db, err := dbconn.ConnectInternal(dsn, appName, dbName)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if closeErr := db.Close(); closeErr != nil {
				err = multierror.Append(err, closeErr)
			}
		}
	}()

	if schema != nil {
		if err := validateSchema(db, schema, validateOnly, observationContext); err != nil {
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
			return fmt.Errorf("database schema out of date")
		}

		options := runner.Options{
			Up:          true,
			SchemaNames: []string{schema.Name},
		}
		return migrationRunner.Run(ctx, options)
	}

	return nil
}

func shouldMigrate(validateOnly bool) bool {
	return !validateOnly || os.Getenv("SG_DEV_MIGRATE_ON_APPLICATION_STARTUP") != ""
}
