package database

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RepoKVPStore interface {
	basestore.ShareableStore
	WithTransact(context.Context, func(RepoKVPStore) error) error
	With(basestore.ShareableStore) RepoKVPStore
	Get(context.Context, api.RepoID, string) (KeyValuePair, error)
	CountKeys(context.Context, RepoKVPListKeysOptions) (int, error)
	ListKeys(context.Context, RepoKVPListKeysOptions, PaginationArgs) ([]string, error)
	CountValues(context.Context, RepoKVPListValuesOptions) (int, error)
	ListValues(context.Context, RepoKVPListValuesOptions, PaginationArgs) ([]string, error)
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

var (
	RepoKVPListKeyColumn   = "key"
	RepoKVPListValueColumn = "value"
)

type KeyValuePair struct {
	Key   string
	Value *string
}

func (s *repoKVPStore) Create(ctx context.Context, repoID api.RepoID, kvp KeyValuePair) error {
	q := `
	INSERT INTO repo_kvps (repo_id, key, value)
	VALUES (%s, %s, %s)
	`

	if err := s.Exec(ctx, sqlf.Sprintf(q, repoID, kvp.Key, kvp.Value)); err != nil {
		if dbutil.IsPostgresError(err, "23505") {
			return errors.Newf(`metadata key %q already exists for the given repository`, kvp.Key)
		}
		return err
	}

	return nil
}

func (s *repoKVPStore) Get(ctx context.Context, repoID api.RepoID, key string) (KeyValuePair, error) {
	q := `
	SELECT key, value
	FROM repo_kvps
	WHERE repo_id = %s
		AND key = %s
	`

	return scanKVP(s.QueryRow(ctx, sqlf.Sprintf(q, repoID, key)))
}

type RepoKVPListKeysOptions struct {
	Query *string
}

func (r *RepoKVPListKeysOptions) SQL() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if r.Query != nil {
		conds = append(conds, sqlf.Sprintf("key ILIKE %s", "%"+*r.Query+"%"))
	}
	return conds
}

func (s *repoKVPStore) CountKeys(ctx context.Context, options RepoKVPListKeysOptions) (int, error) {
	q := `
	WITH kvps AS (
		SELECT COUNT(*) FROM repo_kvps WHERE (%s) GROUP BY key
	)
	SELECT COUNT(*) FROM kvps
	`
	where := options.SQL()
	return basestore.ScanInt(s.QueryRow(ctx, sqlf.Sprintf(q, sqlf.Join(where, ") AND ("))))
}

func (s *repoKVPStore) ListKeys(ctx context.Context, options RepoKVPListKeysOptions, orderOptions PaginationArgs) ([]string, error) {
	where := options.SQL()
	p := orderOptions.SQL()
	if p.Where != nil {
		where = append(where, p.Where)
	}
	q := sqlf.Sprintf(`SELECT key FROM repo_kvps WHERE (%s) GROUP BY key`, sqlf.Join(where, ") AND ("))
	q = p.AppendOrderToQuery(q)
	q = p.AppendLimitToQuery(q)
	return basestore.ScanStrings(s.Query(ctx, q))
}

type RepoKVPListValuesOptions struct {
	Key   string
	Query *string
}

func (r *RepoKVPListValuesOptions) SQL() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("key = %s", r.Key), sqlf.Sprintf("value IS NOT NULL")}
	if r.Query != nil {
		conds = append(conds, sqlf.Sprintf("value ILIKE %s", "%"+*r.Query+"%"))
	}
	return conds
}

func (s *repoKVPStore) CountValues(ctx context.Context, options RepoKVPListValuesOptions) (int, error) {
	q := `SELECT COUNT(DISTINCT value) FROM repo_kvps WHERE (%s)`
	where := options.SQL()
	return basestore.ScanInt(s.QueryRow(ctx, sqlf.Sprintf(q, sqlf.Join(where, ") AND ("))))
}

func (s *repoKVPStore) ListValues(ctx context.Context, options RepoKVPListValuesOptions, orderOptions PaginationArgs) ([]string, error) {
	where := options.SQL()
	p := orderOptions.SQL()
	if p.Where != nil {
		where = append(where, p.Where)
	}
	q := sqlf.Sprintf(`SELECT DISTINCT value FROM repo_kvps WHERE (%s)`, sqlf.Join(where, ") AND ("))
	q = p.AppendOrderToQuery(q)
	q = p.AppendLimitToQuery(q)
	return basestore.ScanStrings(s.Query(ctx, q))
}

func (s *repoKVPStore) Update(ctx context.Context, repoID api.RepoID, kvp KeyValuePair) (KeyValuePair, error) {
	q := `
	UPDATE repo_kvps
	SET value = %s
	WHERE repo_id = %s
		AND key = %s
	RETURNING key, value
	`

	kvp, err := scanKVP(s.QueryRow(ctx, sqlf.Sprintf(q, kvp.Value, repoID, kvp.Key)))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return kvp, errors.Newf(`metadata key %q does not exist for the given repository`, kvp.Key)
		}
		return kvp, errors.Wrap(err, "scanning role")
	}
	return kvp, nil
}

func (s *repoKVPStore) Delete(ctx context.Context, repoID api.RepoID, key string) error {
	q := `
	DELETE FROM repo_kvps
	WHERE repo_id = %s
		AND key = %s
	`

	return s.Exec(ctx, sqlf.Sprintf(q, repoID, key))
}

func scanKVP(scanner dbutil.Scanner) (KeyValuePair, error) {
	var kvp KeyValuePair
	return kvp, scanner.Scan(&kvp.Key, &kvp.Value)
}
