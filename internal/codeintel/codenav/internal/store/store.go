package store

import (
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Store provides the interface for codenav storage.
type Store interface{}

// store manages the codenav store.
type store struct {
	db         *basestore.Store
	operations *operations
}

// New returns a new codenav store.
func New(db database.DB, observationContext *observation.Context) Store {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		operations: newOperations(observationContext),
	}
}
