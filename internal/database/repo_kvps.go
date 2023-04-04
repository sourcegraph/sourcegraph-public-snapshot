package database

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type RepoKVPStore interface {
	basestore.ShareableStore
	WithTransact(context.Context, func(RepoKVPStore) error) error
	With(basestore.ShareableStore) RepoKVPStore
	Get(context.Context, api.RepoID, string) (KeyValuePair, error)
	List(context.Context, api.RepoID) ([]KeyValuePair, error)
	Create(context.Context, api.RepoID, KeyValuePair) error
	Update(context.Context, api.RepoID, KeyValuePair) (KeyValuePair, error)
	Delete(context.Context, api.RepoID, string) error
}

type repoKVPStore struct {
	*basestore.Store
}

var _ RepoKVPStore = (*repoKVPStore)(nil)

func (s *repoKVPStore) WithTransact(ctx context.Context, f func(RepoKVPStore) error) error {
	return s.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		return f(&repoKVPStore{Store: tx})
	})
}

func (s *repoKVPStore) With(other basestore.ShareableStore) RepoKVPStore {
	return &repoKVPStore{Store: s.Store.With(other)}
}

type KeyValuePair struct {
	Key   string
	Value *string
}

func (s *repoKVPStore) Create(ctx context.Context, repoID api.RepoID, kvp KeyValuePair) error {
	q := `
	INSERT INTO repo_kvps (repo_id, key, value)
	VALUES (%s, %s, %s)
	`

	return s.Exec(ctx, sqlf.Sprintf(q, repoID, kvp.Key, kvp.Value))
}

func (s *repoKVPStore) Get(ctx context.Context, repoID api.RepoID, key string) (KeyValuePair, error) {
	q := `
	SELECT key, value
	FROM repo_kvps
	WHERE repo_id = %s
		AND key = %s
	`

	row := s.QueryRow(ctx, sqlf.Sprintf(q, repoID, key))

	var kvp KeyValuePair
	return kvp, row.Scan(&kvp.Key, &kvp.Value)
}

func (s *repoKVPStore) List(ctx context.Context, repoID api.RepoID) ([]KeyValuePair, error) {
	q := `
	SELECT key, value
	FROM repo_kvps
	WHERE repo_id = %s
	`

	scanKVPs := basestore.NewSliceScanner(func(scanner dbutil.Scanner) (KeyValuePair, error) {
		var kvp KeyValuePair
		return kvp, scanner.Scan(&kvp.Key, &kvp.Value)
	})

	return scanKVPs(s.Query(ctx, sqlf.Sprintf(q, repoID)))
}

func (s *repoKVPStore) Update(ctx context.Context, repoID api.RepoID, kvp KeyValuePair) (KeyValuePair, error) {
	q := `
	UPDATE repo_kvps
	SET value = %s
	WHERE repo_id = %s
		AND key = %s
	RETURNING key, value
	`

	row := s.QueryRow(ctx, sqlf.Sprintf(q, kvp.Value, repoID, kvp.Key))

	var updated KeyValuePair
	return updated, row.Scan(&updated.Key, &updated.Value)
}

func (s *repoKVPStore) Delete(ctx context.Context, repoID api.RepoID, key string) error {
	q := `
	DELETE FROM  repo_kvps
	WHERE repo_id = %s
		AND key = %s
	`

	return s.Exec(ctx, sqlf.Sprintf(q, repoID, key))
}
