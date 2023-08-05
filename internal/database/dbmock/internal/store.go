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

// baseStore is the base struct that holds all of the store dependencies.
// It allows us to transfer dependencies between WithDB calls without having
// to recreate the struct every time.
type baseStore struct{}

// store is ment to be a short-lived store that wraps the base store.
// It's the struct that actually implements the Store interface.
// This way the Store methods don't all have to take a db connection, but we
// also don't have to create an entirely new &store{...} instance every time
// we call .WithDB.
//
// Once it's initialised with a .WithDB call on the baseStore class, the
// underlying DB connection cannot be changed again.
type store struct {
	base *baseStore
	db   *basestore.Store
}

// NewStore creates a new BaseStore.
// The Store cannot be used until initialized with .WithDB, at which
// point it turns into the desired interface.
// The first generic parameter is the interface that the second parameter
// will be implementing.
// All dependencies of store can be initialized here, except for the
// *basestore.Store, which is configured by WithDB.
func NewStore( /* dependencies */ ) dbmock.BaseStore[Store] {
	return dbmock.NewBaseStore[Store](&baseStore{ /* initialize dependencies */ })
}

// WithDB sets the store's *basestore.Store and converts it to its
// corresponding interface.
func (s *baseStore) WithDB(db database.DB) Store {
	return &store{base: s, db: basestore.NewWithHandle(db.Handle())}
}

// Create is a Store method for demonstration purposes.
func (s *store) Create(ctx context.Context) error {
	return s.db.Exec(ctx, sqlf.Sprintf("SELECT * FROM test;"))
}
