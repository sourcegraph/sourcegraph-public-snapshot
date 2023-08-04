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
type Store interface {
	// Create is a Store method for demonstration purposes.
	Create(context.Context) error
}

// store is the actual implementation of the Store interface.
// Note that the *basestore.Store is private, and the caller
// needs to call WithDB to set the store.
type store struct {
	store *basestore.Store
}

// NewStore creates a new BaseStore.
// The Store cannot be used until initialized with .WithDB, at which
// point it turns into the desired interface.
// The first generic parameter is the interface that the second parameter
// will be implementing.
// All dependencies of store can be initialized here, except for the
// *basestore.Store, which is configured by WithDB.
func NewStore( /* dependencies */ ) dbmock.BaseStore[Store] {
	return dbmock.NewBaseStore[Store](&store{ /* initialize dependencies */ })
}

// WithDB sets the store's *basestore.Store and converts it to its
// corresponding interface.
func (s *store) WithDB(db database.DB) Store {
	return &store{store: basestore.NewWithHandle(db.Handle())}
}

// Create is a Store method for demonstration purposes.
func (s *store) Create(ctx context.Context) error {
	return s.store.Exec(ctx, sqlf.Sprintf("SELECT * FROM test;"))
}
