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
	Count(context.Context, RepoKVPListOptions) (int, error)
	List(context.Context, RepoKVPListOptions, PaginationArgs) ([]KeyValuePair, error)
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

	if err := s.Exec(ctx, sqlf.Sprintf(q, repoID, kvp.Key, kvp.Value)); err != nil {
		if dbutil.IsPostgresError(err, "23505") {
			return errors.Newf(`metadata key "%q" already exists for the given repository`, kvp.Key)
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

type RepoKVPListOptions struct {
	QueryKey   *string
	QueryValue *string
}

func (r *RepoKVPListOptions) SQL() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if r.QueryKey != nil {
		conds = append(conds, sqlf.Sprintf("key ILIKE %s", "%"+*r.QueryKey+"%"))
	}
	if r.QueryValue != nil {
		conds = append(conds, sqlf.Sprintf("value ILIKE %s", "%"+*r.QueryValue+"%"))
	}
	return conds
}

func (s *repoKVPStore) Count(ctx context.Context, options RepoKVPListOptions) (int, error) {
	q := `
	WITH kvps AS (
		SELECT COUNT(*) FROM repo_kvps WHERE (%s) GROUP BY key, value
	)
	SELECT COUNT(*) FROM kvps
	`
	where := options.SQL()
	return basestore.ScanInt(s.QueryRow(ctx, sqlf.Sprintf(q, sqlf.Join(where, ") AND ("))))
}

func (s *repoKVPStore) List(ctx context.Context, options RepoKVPListOptions, orderOptions PaginationArgs) ([]KeyValuePair, error) {
	where := options.SQL()
	p := orderOptions.SQL()
	if p.Where != nil {
		where = append(where, p.Where)
	}
	if p.Order != nil {
		p.Order = sqlf.Sprintf("COUNT(*) DESC")
	}

	q := sqlf.Sprintf(`SELECT key, value FROM repo_kvps WHERE (%s) GROUP BY key, value`, sqlf.Join(where, ") AND ("))
	q = p.AppendOrderToQuery(q)
	q = p.AppendLimitToQuery(q)

	return scanKVPs(s.Query(ctx, q))
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

var scanKVPs = basestore.NewSliceScanner(scanKVP)
