package db

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
)

// GitserverStore exposes methods to read and write to the set of tables that
// gitserver owns.
type GitserverStore struct {
	*basestore.Store
}

// NewGitserverStoreWithDB instantiates and returns a new GitserverStore.
func NewGitserverStoreWithDB(db dbutil.DB) *GitserverStore {
	return &GitserverStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

func (s *GitserverStore) With(other basestore.ShareableStore) *GitserverStore {
	return &GitserverStore{Store: s.Store.With(other)}
}

func (s *GitserverStore) Transact(ctx context.Context) (*GitserverStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &GitserverStore{Store: txBase}, err
}
