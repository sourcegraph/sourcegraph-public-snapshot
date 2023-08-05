// Example package that demonstrates how to create a new MockableStore.
package internal

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmock"
)

// Store is the interface that will be mocked in tests.
// This has to be added to the mockgen.temp.yaml file so that the mock
// interface can be generated with `sg generate`.
type DBStore interface {
	// Create is a Store method for demonstration purposes.
	Create(context.Context) error
}

// store is the base struct that holds all of the Store dependencies.
// It allows us to transfer dependencies between WithDB calls without having
// to recreate the struct every time.
type Store struct {
	/* Dependencies like loggers */
}

// dbStore is ment to be a short-lived dbStore that wraps the base store.
// It's the struct that actually implements the Store interface.
// This way the Store methods don't all have to take a db connection, but we
// also don't have to create an entirely new &dbStore{...} instance every time
// we call .WithDB.
//
// Once it's initialised with a .WithDB call on the store struct, the
// underlying DB connection should not be changed again.
type dbStore struct {
	// We don't embed *store directly because we
	// don't want WithDB to be accessible.
	base *Store
	db   *basestore.Store
}

// WithDB converts the store to a dbStore that implements the Store interface.
func (s *Store) WithDB(db database.DB) DBStore {
	// For a store to be mockable, this function has to be called when
	// adding the db.
	// It checks if a mock store is found within db.
	if s := dbmock.Get[DBStore](db); s != nil {
		return s
	}

	return &dbStore{base: s, db: basestore.NewWithHandle(db.Handle())}
}

// Create is a Store method for demonstration purposes.
func (s *dbStore) Create(ctx context.Context) error {
	return s.db.Exec(ctx, sqlf.Sprintf("SELECT * FROM test;"))
}
