package database

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// RedisKeyValueStore is a store that exists to satisfy the interface
// redispool.DBStore. This is the interface that is needed to replace redis
// with postgres.
//
// We do not directly implement the interface since that introduces
// complications around dependency graphs.
type RedisKeyValueStore interface {
	basestore.ShareableStore
	WithTransact(context.Context, func(RedisKeyValueStore) error) error
	Get(ctx context.Context, namespace, key string) (value []byte, ok bool, err error)
	Set(ctx context.Context, namespace, key string, value []byte) (err error)
	Delete(ctx context.Context, namespace, key string) (err error)
}

type redisKeyValueStore struct {
	*basestore.Store
}

var _ RedisKeyValueStore = (*redisKeyValueStore)(nil)

func (f *redisKeyValueStore) WithTransact(ctx context.Context, fn func(RedisKeyValueStore) error) error {
	return f.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		return fn(&redisKeyValueStore{Store: tx})
	})
}

func (s *redisKeyValueStore) Get(ctx context.Context, namespace, key string) ([]byte, bool, error) {
	// redispool will often follow up a Get with a Set (eg for implementing
	// redis INCR). As such we need to lock the row with FOR UPDATE.
	q := sqlf.Sprintf(`
	SELECT value FROM redis_key_value
	WHERE namespace = %s AND key = %s
	FOR UPDATE
	`, namespace, key)
	row := s.QueryRow(ctx, q)

	var value []byte
	err := row.Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	} else {
		return value, true, nil
	}
}

func (s *redisKeyValueStore) Set(ctx context.Context, namespace, key string, value []byte) error {
	// value schema does not allow null, nor do we need to preserve nil. So
	// convert to empty string for robustness. This invariant is documented in
	// redispool.DBStore and enforced by tests.
	if value == nil {
		value = []byte{}
	}

	q := sqlf.Sprintf(`
	INSERT INTO redis_key_value (namespace, key, value)
	VALUES (%s, %s, %s)
	ON CONFLICT (namespace, key) DO UPDATE SET value = EXCLUDED.value
	`, namespace, key, value)
	return s.Exec(ctx, q)
}

func (s *redisKeyValueStore) Delete(ctx context.Context, namespace, key string) error {
	q := sqlf.Sprintf(`
	DELETE FROM redis_key_value
	WHERE namespace = %s AND key = %s
	`, namespace, key)
	return s.Exec(ctx, q)
}
