// Example package that demonstrates how to create a new MockableStore.
package internal

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmock"
)

// Store is the interface that will be mocked in tests.
// This has to be added to the mockgen.temp.yaml file so that the mock
// interface can be generated with `sg generate`.
type Store interface {
	// Create is a Store method for demonstration purposes.
	Create(context.Context) error
}

// ToEmbeddable has to be added to the _MockStore_ struct that is generated
// from the `sg generate` command. This is done outside the generated mock
// file so that it does not get overwritten on subsequent `sg generate` calls.
// This exact method can be copied.
func (s *MockStore) ToEmbeddable() dbmock.Embeddable {
	return dbmock.NewEmbeddable(s)
}

// store is the actual implementation of the Store interface.
// It has an embedded *dbmock.MockableStore[Store]. The generic parameter
// specifies which interface will be mockable.
type store struct {
	*dbmock.MockableStore[Store]
}

// NewStore creates a new store struct.
// Any dependencies of store should be passed to NewStore, except for a
// database.DB dependency. The embedded MockableStore takes care of the
// database.DB dependency.
//
// The self-referential initialization may seem a bit strange. MockableStore
// requires access to store so that it can be returned when the WithDB method
// is called.
func NewStore( /* dependencies */ ) *store {
	s := &store{ /* initialize dependencies */ }
	s.MockableStore = dbmock.NewMockableStore[Store](s)
	return s
}

// Create is a Store method for demonstration purposes.
func (s *store) Create(ctx context.Context) error {
	return s.Exec(ctx, sqlf.Sprintf("SELECT * FROM test;"))
}
