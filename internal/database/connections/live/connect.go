package connections

import (
	"context"
	"database/sql"

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
	schemas := []*schemas.Schema{schemas.Frontend}
	if !migrate {
		schemas = nil
	}

	return connect(dsn, appName, "frontend", observationContext, schemas...)
}

// NewCodeIntelDB creates a new connection to the codeintel database. After successful connection,
// the schema version of the database will be compared against an expected version and migrations
// may be run (taking an advisory lock to ensure exclusive access).
//
// TEMPORARY: The migrate flag controls whether or not migrations/version checks are performed on
// the version. When false, we give back a handle without running any migrations and assume that
// the database schema is up to date.
func NewCodeIntelDB(dsn, appName string, migrate bool, observationContext *observation.Context) (*sql.DB, error) {
	schemas := []*schemas.Schema{schemas.CodeIntel}
	if !migrate {
		schemas = nil
	}

	return connect(dsn, appName, "codeintel", observationContext, schemas...)
}

// NewCodeInsightsDB creates a new connection to the codeinsights database. After successful
// connection, the schema version of the database will be compared against an expected version and
// migrations may be run (taking an advisory lock to ensure exclusive access).
//
// TEMPORARY: The migrate flag controls whether or not migrations/version checks are performed on
// the version. When false, we give back a handle without running any migrations and assume that
// the database schema is up to date.
func NewCodeInsightsDB(dsn, appName string, migrate bool, observationContext *observation.Context) (*sql.DB, error) {
	schemas := []*schemas.Schema{schemas.CodeInsights}
	if !migrate {
		schemas = nil
	}

	return connect(dsn, appName, "codeinsight", observationContext, schemas...)
}

func connect(dsn, appName, dbName string, observationContext *observation.Context, schemas ...*schemas.Schema) (*sql.DB, error) {
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

	options := runner.Options{
		Up:          true,
		SchemaNames: schemaNames(schemas),
	}
	if err := runnerFromDB(newStoreFactory(observationContext), db, schemas...).Run(context.Background(), options); err != nil {
		return nil, err
	}

	return db, nil
}

func schemaNames(schemas []*schemas.Schema) []string {
	names := make([]string, 0, len(schemas))
	for _, schema := range schemas {
		names = append(names, schema.Name)
	}

	return names
}
