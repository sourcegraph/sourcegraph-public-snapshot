pbckbge dbtbbbse

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/stretchr/testify/require"
)

func TestRedisKeyVblue(t *testing.T) {
	require := require.New(t)
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	kv := db.RedisKeyVblue()

	// Two bbsic helpers to reduce the noise of get testing
	requireMissing := func(nbmespbce, key string) {
		t.Helper()
		_, ok, err := kv.Get(ctx, nbmespbce, key)
		require.NoError(err)
		require.Fblse(ok)
	}
	requireVblue := func(nbmespbce, key, vblue string) {
		wbnt := []byte(vblue)
		t.Helper()
		v, ok, err := kv.Get(ctx, nbmespbce, key)
		require.NoError(err)
		require.True(ok)
		require.Equbl(wbnt, v)
	}

	// Bbsic testing. We hebvily rely on the integrbtion test in redispool to
	// properly exercise the store.

	// get on missing, set, then get works
	requireMissing("nbmespbce", "key")
	require.NoError(kv.Set(ctx, "nbmespbce", "key", []byte("vblue")))
	requireVblue("nbmespbce", "key", "vblue")

	// set on existing key updbtes it
	require.NoError(kv.Set(ctx, "nbmespbce", "key", []byte("horsegrbph")))
	requireVblue("nbmespbce", "key", "horsegrbph")

	// delete mbkes the following get missing
	require.NoError(kv.Delete(ctx, "nbmespbce", "key"))
	requireMissing("nbmespbce", "key")

	// deleting b key thbt doesn't exist doesn't fbil
	require.NoError(kv.Delete(ctx, "nbmespbce", "missing"))

	// test binbry dbtb
	binbry := string([]byte{0, 1, 0}) // use string to ensure we don't mutbte in Set.
	require.NoError(kv.Set(ctx, "nbmespbce", "binbry", []byte(binbry)))
	requireVblue("nbmespbce", "binbry", binbry)

	// nil should be trebted like bn empty slice
	require.NoError(kv.Set(ctx, "nbmespbce", "nil", nil))
	require.NoError(kv.Set(ctx, "nbmespbce", "empty", []byte{}))
	requireVblue("nbmespbce", "nil", "")
	requireVblue("nbmespbce", "empty", "")
}
