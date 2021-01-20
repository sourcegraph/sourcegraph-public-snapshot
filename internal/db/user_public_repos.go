package db

import (
	"context"
	"database/sql"
	"sync"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
)

type UserPublicRepoStore struct {
	store *basestore.Store

	once sync.Once
}

// NewUserPublicRepoStoreWithDB instantiates and returns a new RepoStore with prepared statements.
func NewUserPublicRepoStoreWithDB(db dbutil.DB) *UserPublicRepoStore {
	return &UserPublicRepoStore{store: basestore.NewWithDB(db, sql.TxOptions{})}
}

// ensureStore instantiates a basestore.Store if necessary, using the dbconn.Global handle.
// This function ensures access to dbconn happens after the rest of the code or tests have
// initialized it.
func (s *UserPublicRepoStore) ensureStore() {
	s.once.Do(func() {
		if s.store == nil {
			s.store = basestore.NewWithDB(dbconn.Global, sql.TxOptions{})
		}
	})
}

func (s *UserPublicRepoStore) SetUserRepo(ctx context.Context, userID int32, repoID api.RepoID) error {
	s.ensureStore()

	return s.store.Exec(ctx, sqlf.Sprintf("INSERT INTO user_public_repos(user_id, repo_id) VALUES (%v, %v)", userID, repoID))
}

func (s *UserPublicRepoStore) ListByUser(ctx context.Context, userID int32) ([]int32, error) {
	s.ensureStore()
	return basestore.ScanInt32s(s.store.Query(ctx, sqlf.Sprintf("SELECT repo_id FROM user_public_repos WHERE user_id = %v", userID)))
}
