package db

import (
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/services/executors/store"
)

// ensure that all the needed methods are implemented.
var _ store.Store = (*ExecutorStore)(nil)

type ExecutorStore struct {
	db *basestore.Store
}

// New instantiates and returns a new ExecutorStore with prepared statements.
func New(db database.DB) *ExecutorStore {
	return &ExecutorStore{db: basestore.NewWithHandle(db.Handle())}
}
