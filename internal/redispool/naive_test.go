package redispool_test

import (
	"context"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestInMemoryKeyValue(t *testing.T) {
	testKeyValue(t, redispool.MemoryKeyValue())
}

func TestDBKeyValue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping DB test since -short is specified")
	}
	t.Parallel()

	require := require{TB: t}
	db := redispool.DBKeyValue("test")

	require.Equal(db.Get("db_test"), errors.New("redispool.DBRegisterStore has not been called"))

	// Now register and check if db starts to work
	if err := redispool.DBRegisterStore(dbStoreTransact(t)); err != nil {
		t.Fatal(err)
	}

	t.Run("integration", func(t *testing.T) {
		testKeyValue(t, db)
	})

	// Ensure we can't register twice
	if err := redispool.DBRegisterStore(dbStoreTransact(t)); err == nil {
		t.Fatal("expected second call to DBRegisterStore to fail")
	}
	if err := redispool.DBRegisterStore(nil); err == nil {
		t.Fatal("expected third call to DBRegisterStore to fail")
	}
	// Ensure we are still working
	require.Equal(db.Get("db_test"), redis.ErrNil)

	// Check that namespacing works. Intentionally use same namespace as db
	// for db1.
	db1 := redispool.DBKeyValue("test")
	db2 := redispool.DBKeyValue("test2")
	require.Works(db1.Set("db_test", "1"))
	require.Works(db2.Set("db_test", "2"))
	require.Equal(db1.Get("db_test"), "1")
	require.Equal(db2.Get("db_test"), "2")
}

func dbStoreTransact(t *testing.T) redispool.DBStoreTransact {
	logger := logtest.Scoped(t)
	kvNoTX := database.NewDB(logger, dbtest.NewDB(t)).RedisKeyValue()

	return func(ctx context.Context, f func(redispool.DBStore) error) error {
		return kvNoTX.WithTransact(ctx, func(tx database.RedisKeyValueStore) error {
			return f(tx)
		})
	}
}
