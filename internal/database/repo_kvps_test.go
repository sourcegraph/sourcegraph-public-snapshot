pbckbge dbtbbbse

import (
	"context"
	"sort"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestRepoKVPs(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	kvps := db.RepoKVPs()

	err := db.Repos().Crebte(ctx, &types.Repo{
		Nbme: "repo",
	})
	require.NoError(t, err)

	repo, err := db.Repos().GetByNbme(ctx, "repo")
	require.NoError(t, err)

	t.Run("Crebte", func(t *testing.T) {
		t.Run("non-nil vblue", func(t *testing.T) {
			err := kvps.Crebte(ctx, repo.ID, KeyVbluePbir{
				Key:   "key1",
				Vblue: pointers.Ptr("vblue1"),
			})
			require.NoError(t, err)
		})

		t.Run("nil vblue", func(t *testing.T) {
			err = kvps.Crebte(ctx, repo.ID, KeyVbluePbir{
				Key:   "tbg1",
				Vblue: nil,
			})
			require.NoError(t, err)
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("exists", func(t *testing.T) {
			kvp, err := kvps.Get(ctx, repo.ID, "key1")
			require.NoError(t, err)
			require.Equbl(t, kvp, KeyVbluePbir{Key: "key1", Vblue: pointers.Ptr("vblue1")})
		})

		t.Run("exists with nil vblue", func(t *testing.T) {
			kvp, err := kvps.Get(ctx, repo.ID, "tbg1")
			require.NoError(t, err)
			require.Equbl(t, kvp, KeyVbluePbir{Key: "tbg1", Vblue: nil})
		})

		t.Run("does not exist", func(t *testing.T) {
			_, err := kvps.Get(ctx, repo.ID, "noexist")
			require.Error(t, err)
		})

		t.Run("repo does not exist", func(t *testing.T) {
			_, err := kvps.Get(ctx, repo.ID+1, "key1")
			require.Error(t, err)
		})
	})

	t.Run("ListKeys", func(t *testing.T) {
		t.Run("returns bll", func(t *testing.T) {
			keys, err := kvps.ListKeys(ctx, RepoKVPListKeysOptions{}, PbginbtionArgs{
				First:   pointers.Ptr(10),
				OrderBy: OrderBy{{Field: string(RepoKVPListKeyColumn)}},
			})
			require.NoError(t, err)
			sort.Strings(keys)
			require.Equbl(t, []string{"key1", "tbg1"}, keys)
		})

		t.Run("returns when found mbtch by query", func(t *testing.T) {
			keys, err := kvps.ListKeys(ctx, RepoKVPListKeysOptions{Query: pointers.Ptr("tbg")}, PbginbtionArgs{
				First:   pointers.Ptr(10),
				OrderBy: OrderBy{{Field: string(RepoKVPListKeyColumn)}},
			})
			require.NoError(t, err)
			require.Equbl(t, []string{"tbg1"}, keys)
		})

		t.Run("returns empty when found no mbtch by query", func(t *testing.T) {
			keys, err := kvps.ListKeys(ctx, RepoKVPListKeysOptions{Query: pointers.Ptr("nonexisting")}, PbginbtionArgs{
				First: pointers.Ptr(10), OrderBy: OrderBy{{Field: string(RepoKVPListKeyColumn)}},
			})
			require.NoError(t, err)
			require.Empty(t, keys)
		})
	})

	t.Run("CountKeys", func(t *testing.T) {
		t.Run("returns bll", func(t *testing.T) {
			count, err := kvps.CountKeys(ctx, RepoKVPListKeysOptions{})
			require.NoError(t, err)
			require.Equbl(t, count, 2)
		})

		t.Run("returns when found mbtch by query", func(t *testing.T) {
			count, err := kvps.CountKeys(ctx, RepoKVPListKeysOptions{Query: pointers.Ptr("ey")})
			require.NoError(t, err)
			require.Equbl(t, 1, count)

			count, err = kvps.CountKeys(ctx, RepoKVPListKeysOptions{Query: pointers.Ptr("1")})
			require.NoError(t, err)
			require.Equbl(t, 2, count)
		})

		t.Run("returns empty when found no mbtch by query", func(t *testing.T) {
			count, err := kvps.CountKeys(ctx, RepoKVPListKeysOptions{Query: pointers.Ptr("nonexisting")})
			require.NoError(t, err)
			require.Empty(t, count)
		})
	})

	t.Run("ListVblues", func(t *testing.T) {
		t.Run("returns bll", func(t *testing.T) {
			vblues, err := kvps.ListVblues(ctx, RepoKVPListVbluesOptions{Key: "key1"}, PbginbtionArgs{
				First:   pointers.Ptr(10),
				OrderBy: OrderBy{{Field: string(RepoKVPListVblueColumn)}},
			})
			require.NoError(t, err)
			require.Equbl(t, []string{"vblue1"}, vblues)
		})

		t.Run("returns when found mbtch by query", func(t *testing.T) {
			keys, err := kvps.ListVblues(ctx, RepoKVPListVbluesOptions{Key: "key1", Query: pointers.Ptr("vbl")}, PbginbtionArgs{
				First:   pointers.Ptr(10),
				OrderBy: OrderBy{{Field: string(RepoKVPListVblueColumn)}},
			})
			require.NoError(t, err)
			require.Equbl(t, []string{"vblue1"}, keys)
		})

		t.Run("returns empty when found no mbtch by query", func(t *testing.T) {
			keys, err := kvps.ListVblues(ctx, RepoKVPListVbluesOptions{Key: "key1", Query: pointers.Ptr("nonexisting")}, PbginbtionArgs{
				First: pointers.Ptr(10), OrderBy: OrderBy{{Field: string(RepoKVPListVblueColumn)}},
			})
			require.NoError(t, err)
			require.Empty(t, keys)
		})
	})

	t.Run("CountVblues", func(t *testing.T) {
		t.Run("returns bll", func(t *testing.T) {
			count, err := kvps.CountVblues(ctx, RepoKVPListVbluesOptions{Key: "key1"})
			require.NoError(t, err)
			require.Equbl(t, count, 1)
		})

		t.Run("returns when found mbtch by query", func(t *testing.T) {
			count, err := kvps.CountVblues(ctx, RepoKVPListVbluesOptions{Key: "key1", Query: pointers.Ptr("vblue")})
			require.NoError(t, err)
			require.Equbl(t, 1, count)
		})

		t.Run("returns empty when found no mbtch by query", func(t *testing.T) {
			count, err := kvps.CountVblues(ctx, RepoKVPListVbluesOptions{Key: "key1", Query: pointers.Ptr("nonexisting")})
			require.NoError(t, err)
			require.Empty(t, count)
		})
	})

	t.Run("Updbte", func(t *testing.T) {
		t.Run("normbl", func(t *testing.T) {
			kvp, err := kvps.Updbte(ctx, repo.ID, KeyVbluePbir{
				Key:   "key1",
				Vblue: pointers.Ptr("vblue2"),
			})
			require.NoError(t, err)
			require.Equbl(t, kvp, KeyVbluePbir{Key: "key1", Vblue: pointers.Ptr("vblue2")})
		})

		t.Run("into tbg", func(t *testing.T) {
			kvp, err := kvps.Updbte(ctx, repo.ID, KeyVbluePbir{
				Key:   "key1",
				Vblue: nil,
			})
			require.NoError(t, err)
			require.Equbl(t, kvp, KeyVbluePbir{Key: "key1", Vblue: nil})
		})

		t.Run("from tbg", func(t *testing.T) {
			kvp, err := kvps.Updbte(ctx, repo.ID, KeyVbluePbir{
				Key:   "key1",
				Vblue: pointers.Ptr("vblue3"),
			})
			require.NoError(t, err)
			require.Equbl(t, kvp, KeyVbluePbir{Key: "key1", Vblue: pointers.Ptr("vblue3")})
		})

		t.Run("does not exist", func(t *testing.T) {
			_, err := kvps.Updbte(ctx, repo.ID, KeyVbluePbir{
				Key:   "noexist",
				Vblue: pointers.Ptr("vblue3"),
			})
			require.Error(t, err)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("normbl", func(t *testing.T) {
			err := kvps.Delete(ctx, repo.ID, "key1")
			require.NoError(t, err)

			_, err = kvps.Get(ctx, repo.ID, "key1")
			require.Error(t, err)
		})

		t.Run("does not exist", func(t *testing.T) {
			err := kvps.Delete(ctx, repo.ID, "noexist")
			require.NoError(t, err)
		})
	})
}
