package connections

import (
	"context"
	"database/sql"

	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Store interface {
	runner.Store
	EnsureSchemaTable(ctx context.Context) error
}

type StoreFactory func(db *sql.DB, migrationsTable string) Store

func newStoreFactory(observationContext *observation.Context) func(db *sql.DB, migrationsTable string) Store {
	return func(db *sql.DB, migrationsTable string) Store {
		return store.NewWithDB(db, migrationsTable, store.NewOperations(observationContext))
	}
}

func initStore(ctx context.Context, newStore StoreFactory, db *sql.DB, schema *schemas.Schema) (Store, error) {
	store := newStore(db, schema.MigrationsTableName)

	if err := store.EnsureSchemaTable(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}

		return nil, err
	}

	return store, nil
}
