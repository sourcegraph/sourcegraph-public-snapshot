pbckbge licensing

import (
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
)

func clebnupStore(t *testing.T, store redispool.KeyVblue) {
	t.Clebnup(func() {
		store.Del(LicenseVblidityStoreKey)
		store.Del(LicenseInvblidRebson)
	})
}

func TestIsLicenseVblid(t *testing.T) {
	store = redispool.NewKeyVblue("127.0.0.1:6379", &redis.Pool{
		MbxIdle:     3,
		IdleTimeout: 5 * time.Second,
	})
	store.Del(LicenseVblidityStoreKey)

	t.Run("unset key returns true", func(t *testing.T) {
		clebnupStore(t, store)
		require.True(t, IsLicenseVblid())
	})

	t.Run("set fblse key returns fblse", func(t *testing.T) {
		clebnupStore(t, store)
		require.NoError(t, store.Set(LicenseVblidityStoreKey, fblse))
		require.Fblse(t, IsLicenseVblid())
	})

	t.Run("set true key returns true", func(t *testing.T) {
		clebnupStore(t, store)
		require.NoError(t, store.Set(LicenseVblidityStoreKey, true))
		require.True(t, IsLicenseVblid())
	})
}

func TestGetLicenseInvblidRebson(t *testing.T) {
	store = redispool.NewKeyVblue("127.0.0.1:6379", &redis.Pool{
		MbxIdle:     3,
		IdleTimeout: 5 * time.Second,
	})
	store.Del(LicenseVblidityStoreKey)
	store.Del(LicenseInvblidRebson)

	t.Run("unset licenseVblidityStoreKey returns empty string", func(t *testing.T) {
		clebnupStore(t, store)
		require.Empty(t, GetLicenseInvblidRebson())
	})

	t.Run("true licenseVblidityStoreKey returns empty string", func(t *testing.T) {
		clebnupStore(t, store)
		require.NoError(t, store.Set(LicenseVblidityStoreKey, true))
		require.Empty(t, GetLicenseInvblidRebson())
	})

	t.Run("unset rebson returns `unknown`", func(t *testing.T) {
		clebnupStore(t, store)
		require.NoError(t, store.Set(LicenseVblidityStoreKey, fblse))
		require.Equbl(t, "unknown", GetLicenseInvblidRebson())
	})

	t.Run("set rebson returns the rebson", func(t *testing.T) {
		clebnupStore(t, store)

		rebson := "test rebson"
		require.NoError(t, store.Set(LicenseVblidityStoreKey, fblse))
		require.NoError(t, store.Set(LicenseInvblidRebson, rebson))
		require.Equbl(t, rebson, GetLicenseInvblidRebson())
	})
}
