package database

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

func UserPublicReposWith(other basestore.ShareableStore) UserPublicRepoStore {
	return &userPublicRepoStore{Store: basestore.NewWithHandle(other.Handle())}
}

type UserPublicRepoStore interface {
	basestore.ShareableStore
	SetUserRepos(ctx context.Context, userID int32, repos []UserPublicRepo) error
	SetUserRepo(context.Context, UserPublicRepo) error
	ListByUser(context.Context, int32) ([]UserPublicRepo, error)
}

type userPublicRepoStore struct {
	*basestore.Store
}

// SetUserRepos replaces all the repos in user_public_repos for the passed user ID
func (s *userPublicRepoStore) SetUserRepos(ctx context.Context, userID int32, repos []UserPublicRepo) (err error) {
	var tx *basestore.Store
	tx, err = s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Done(err)
	}()
	// clear existing repos for this user
	err = tx.Exec(ctx, sqlf.Sprintf(
		"DELETE FROM user_public_repos WHERE user_id = %v",
		userID,
	))
	if err != nil {
		return err
	}
	if len(repos) == 0 {
		return nil
	}
	values := make([]*sqlf.Query, 0, len(repos))
	for _, repo := range repos {
		values = append(values, sqlf.Sprintf(
			"(%s, %s, %s)",
			userID, repo.RepoURI, repo.RepoID,
		))
	}
	return tx.Exec(ctx, sqlf.Sprintf(
		"INSERT INTO user_public_repos(user_id, repo_uri, repo_id) VALUES %s",
		sqlf.Join(values, ","),
	))
}

// SetUserRepo stores a UserPublicRepo record, if a record already exists for the same user_id & repo_id combo, the
// repo_uri is updated
func (s *userPublicRepoStore) SetUserRepo(ctx context.Context, upr UserPublicRepo) error {
	return s.Exec(ctx, sqlf.Sprintf(
		`INSERT INTO
			user_public_repos(user_id, repo_uri, repo_id)
		VALUES (%s, %s,  %s)
		ON CONFLICT(user_id, repo_id) DO UPDATE
		SET
			repo_uri = excluded.repo_uri`,
		upr.UserID, upr.RepoURI, upr.RepoID,
	))
}

func (s *userPublicRepoStore) ListByUser(ctx context.Context, userID int32) ([]UserPublicRepo, error) {
	rows, err := s.Query(ctx, sqlf.Sprintf(
		"SELECT user_id, repo_uri, repo_id FROM user_public_repos WHERE user_id = %v",
		userID,
	))
	if err != nil {
		return nil, err
	}

	var out []UserPublicRepo
	for rows.Next() {
		v := UserPublicRepo{}
		err := rows.Scan(&v.UserID, &v.RepoURI, &v.RepoID)
		if err != nil {
			return out, err
		}
		out = append(out, v)
	}

	return out, nil
}

type UserPublicRepo struct {
	UserID  int32
	RepoURI string
	RepoID  api.RepoID
}
