package licensing

import (
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/stretchr/testify/require"
)

func cleanupStore(t *testing.T, store redispool.KeyValue) {
	t.Cleanup(func() {
		store.Del(licenseValidityStoreKey)
	})
}

func TestIsLicenseValid(t *testing.T) {
	store = redispool.NewKeyValue("127.0.0.1:6379", &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 5 * time.Second,
	})
	store.Del(licenseValidityStoreKey)

	t.Run("unset key returns true", func(t *testing.T) {
		cleanupStore(t, store)
		require.True(t, IsLicenseValid())
	})

	t.Run("set false key returns false", func(t *testing.T) {
		cleanupStore(t, store)
		require.NoError(t, store.Set(licenseValidityStoreKey, false))
		require.False(t, IsLicenseValid())
	})

	t.Run("set true key returns true", func(t *testing.T) {
		cleanupStore(t, store)
		require.NoError(t, store.Set(licenseValidityStoreKey, true))
		require.True(t, IsLicenseValid())
	})
}
