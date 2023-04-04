package connections

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	migrationstore "github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Store interface {
	runner.Store
	EnsureSchemaTable(ctx context.Context) error
	BackfillSchemaVersions(ctx context.Context) error
}

type StoreFactory func(db *sql.DB, migrationsTable string) Store

func newStoreFactory(observationCtx *observation.Context) func(db *sql.DB, migrationsTable string) Store {
	return func(db *sql.DB, migrationsTable string) Store {
		return NewStoreShim(migrationstore.NewWithDB(observationCtx, db, migrationsTable))
	}
}

func initStore(ctx context.Context, newStore StoreFactory, db *sql.DB, schema *schemas.Schema) (Store, error) {
	store := newStore(db, schema.MigrationsTableName)

	if err := store.EnsureSchemaTable(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			err = errors.Append(err, closeErr)
		}

		return nil, err
	}

	if err := store.BackfillSchemaVersions(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			err = errors.Append(err, closeErr)
		}

		return nil, err
	}

	return store, nil
}

type storeShim struct {
	*migrationstore.Store
}

func NewStoreShim(s *migrationstore.Store) Store {
	return &storeShim{s}
}

func (s *storeShim) Transact(ctx context.Context) (runner.Store, error) {
	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &storeShim{tx}, nil
}
