package licensing

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/stretchr/testify/require"
)

func TestIsLicenseValid(t *testing.T) {
	rcache.SetupForTest(t)

	t.Run("unset key returns true", func(t *testing.T) {
		require.True(t, isLicenseValid())
	})

	t.Run("set false key returns false", func(t *testing.T) {
		require.NoError(t, redispool.Store.Set("is_license_valid", false))
		require.False(t, isLicenseValid())
	})

	t.Run("set true key returns true", func(t *testing.T) {
		require.NoError(t, redispool.Store.Set("is_license_valid", true))
		require.True(t, isLicenseValid())
	})
}
