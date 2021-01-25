package database

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type UserPublicRepoStore struct {
	*basestore.Store
}

// UserPublicRepos instantiates and returns a new RepoStore with prepared statements.
func UserPublicRepos(db dbutil.DB) *UserPublicRepoStore {
	return &UserPublicRepoStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

// NewUserPublicRepoStoreWithDB instantiates and returns a new UserPublicRepoStore using the other store handle.
func UserPublicReposWith(other basestore.ShareableStore) *UserPublicRepoStore {
	return &UserPublicRepoStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *UserPublicRepoStore) With(other basestore.ShareableStore) *UserPublicRepoStore {
	return &UserPublicRepoStore{Store: s.Store.With(other)}
}

func (s *UserPublicRepoStore) Transact(ctx context.Context) (*UserPublicRepoStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &UserPublicRepoStore{Store: txBase}, err
}

func (s *UserPublicRepoStore) SetUserRepo(ctx context.Context, userID int32, repoID api.RepoID) error {
	return s.Store.Exec(ctx, sqlf.Sprintf("INSERT INTO user_public_repos(user_id, repo_id) VALUES (%v, %v)", userID, repoID))
}

func (s *UserPublicRepoStore) ListByUser(ctx context.Context, userID int32) ([]int32, error) {
	rows, err := basestore.ScanInt32s(s.Store.Query(ctx, sqlf.Sprintf("SELECT repo_id FROM user_public_repos WHERE user_id = %v", userID)))
	if err != nil {
		return nil, err
	}
	return rows, nil
}
