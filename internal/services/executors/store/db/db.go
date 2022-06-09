package db

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/services/executors/store"
)

// ensure that all the needed methods are implemented.
var _ store.Store = (*ExecutorStore)(nil)

type ExecutorStore struct {
	db *basestore.Store
}

// New instantiates and returns a new ExecutorStore with prepared statements.
func New(db dbutil.DB) *ExecutorStore {
	return &ExecutorStore{db: basestore.NewWithDB(db, sql.TxOptions{})}
}
