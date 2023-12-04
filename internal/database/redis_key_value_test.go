package database

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/stretchr/testify/require"
)

func TestRedisKeyValue(t *testing.T) {
	require := require.New(t)
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()
	kv := db.RedisKeyValue()

	// Two basic helpers to reduce the noise of get testing
	requireMissing := func(namespace, key string) {
		t.Helper()
		_, ok, err := kv.Get(ctx, namespace, key)
		require.NoError(err)
		require.False(ok)
	}
	requireValue := func(namespace, key, value string) {
		want := []byte(value)
		t.Helper()
		v, ok, err := kv.Get(ctx, namespace, key)
		require.NoError(err)
		require.True(ok)
		require.Equal(want, v)
	}

	// Basic testing. We heavily rely on the integration test in redispool to
	// properly exercise the store.

	// get on missing, set, then get works
	requireMissing("namespace", "key")
	require.NoError(kv.Set(ctx, "namespace", "key", []byte("value")))
	requireValue("namespace", "key", "value")

	// set on existing key updates it
	require.NoError(kv.Set(ctx, "namespace", "key", []byte("horsegraph")))
	requireValue("namespace", "key", "horsegraph")

	// delete makes the following get missing
	require.NoError(kv.Delete(ctx, "namespace", "key"))
	requireMissing("namespace", "key")

	// deleting a key that doesn't exist doesn't fail
	require.NoError(kv.Delete(ctx, "namespace", "missing"))

	// test binary data
	binary := string([]byte{0, 1, 0}) // use string to ensure we don't mutate in Set.
	require.NoError(kv.Set(ctx, "namespace", "binary", []byte(binary)))
	requireValue("namespace", "binary", binary)

	// nil should be treated like an empty slice
	require.NoError(kv.Set(ctx, "namespace", "nil", nil))
	require.NoError(kv.Set(ctx, "namespace", "empty", []byte{}))
	requireValue("namespace", "nil", "")
	requireValue("namespace", "empty", "")
}
