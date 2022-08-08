package database

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type RepoMetadataStore interface {
	basestore.ShareableStore
	Transact(context.Context) (RepoMetadataStore, error)
	With(basestore.ShareableStore) RepoMetadataStore
}

type repoMetadataStore struct {
	*basestore.Store
}

var _ RepoMetadataStore = (*repoMetadataStore)(nil)

func (s *repoMetadataStore) Transact(ctx context.Context) (RepoMetadataStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &repoMetadataStore{Store: txBase}, err
}

func (s *repoMetadataStore) With(other basestore.ShareableStore) RepoMetadataStore {
	return &repoMetadataStore{Store: s.Store.With(other)}
}

type KeyValuePair struct {
	Key   string
	Value *string
}

func (s *repoMetadataStore) Get(ctx context.Context, repoID api.RepoID, key string) (KeyValuePair, error) {
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

func (s *repoMetadataStore) List(ctx context.Context, repoID api.RepoID) ([]KeyValuePair, error) {
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

func (s *repoMetadataStore) Create(ctx context.Context, repoID api.RepoID, kvp KeyValuePair) error {
	q := `
	INSERT INTO repo_kvps (repo_id, key, value)
	VALUES (%s, %s, %s)
	`

	return s.Exec(ctx, sqlf.Sprintf(q, repoID, kvp.Key, kvp.Value))
}

func (s *repoMetadataStore) Update(ctx context.Context, repoID api.RepoID, kvp KeyValuePair) (KeyValuePair, error) {
	q := `
	UPDATE repo_kvps
	SET value = %s
	WHERE repo_id = %s
		AND key = %s
	RETURNING (key, value)
	`

	row := s.QueryRow(ctx, sqlf.Sprintf(q, kvp.Value, repoID, kvp.Key))

	var updated KeyValuePair
	return updated, row.Scan(&updated.Key, &updated.Value)
}

func (s *repoMetadataStore) Delete(ctx context.Context, repoID api.RepoID, key string) error {
	q := `
	DELETE FROM  repo_kvps
	WHERE repo_id = %s
		AND key = %s
	`

	return s.Exec(ctx, sqlf.Sprintf(q, repoID, key))
}
