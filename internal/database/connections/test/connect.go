package connections

import (
	"context"
	"database/sql"
	"testing"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewTestDB creates a new connection to the a database and applies the given migrations.
func NewTestDB(t testing.TB, logger log.Logger, dsn string, schemas ...*schemas.Schema) (_ *sql.DB, err error) {
	db, err := dbconn.ConnectInternal(logger, dsn, "", "")
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

	schemaNames := schemaNames(schemas)

	operations := make([]runner.MigrationOperation, 0, len(schemaNames))
	for _, schemaName := range schemaNames {
		operations = append(operations, runner.MigrationOperation{
			SchemaName: schemaName,
			Type:       runner.MigrationOperationTypeUpgrade,
		})
	}

	options := runner.Options{
		Operations: operations,
	}

	migrationLogger := logtest.ScopedWith(t, logtest.LoggerOptions{Level: log.LevelError})
	if err := runner.NewRunnerWithSchemas(migrationLogger, newStoreFactoryMap(db, schemas), schemas).Run(context.Background(), options); err != nil {
		return nil, err
	}

	return db, nil
}

func newStoreFactoryMap(db *sql.DB, schemas []*schemas.Schema) map[string]runner.StoreFactory {
	storeFactoryMap := make(map[string]runner.StoreFactory, len(schemas))
	for _, schema := range schemas {
		schema := schema

		storeFactoryMap[schema.Name] = func(ctx context.Context) (runner.Store, error) {
			return newMemoryStore(db), nil
		}
	}

	return storeFactoryMap
}

func schemaNames(schemas []*schemas.Schema) []string {
	names := make([]string, 0, len(schemas))
	for _, schema := range schemas {
		names = append(names, schema.Name)
	}

	return names
}
