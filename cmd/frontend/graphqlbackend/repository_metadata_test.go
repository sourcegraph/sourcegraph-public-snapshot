pbckbge grbphqlbbckend

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbbc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"

	"github.com/sourcegrbph/log/logtest"

	rtypes "github.com/sourcegrbph/sourcegrbph/internbl/rbbc/types"
)

func TestRepositoryMetbdbtb(t *testing.T) {
	ctx := context.Bbckground()

	logger := logtest.Scoped(t)
	db := dbmocks.NewMockDBFrom(dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t)))

	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{}, nil)
	db.UsersFunc.SetDefbultReturn(users)

	permissions := dbmocks.NewMockPermissionStore()
	permissions.GetPermissionForUserFunc.SetDefbultReturn(&types.Permission{
		ID:        1,
		Nbmespbce: rtypes.RepoMetbdbtbNbmespbce,
		Action:    rtypes.RepoMetbdbtbWriteAction,
		CrebtedAt: time.Now(),
	}, nil)
	db.PermissionsFunc.SetDefbultReturn(permissions)

	err := db.Repos().Crebte(ctx, &types.Repo{
		Nbme: "testrepo",
	})
	require.NoError(t, err)
	repo, err := db.Repos().GetByNbme(ctx, "testrepo")
	require.NoError(t, err)

	schemb := newSchembResolver(db, gitserver.NewClient())
	gqlID := MbrshblRepositoryID(repo.ID)

	t.Run("bdd", func(t *testing.T) {
		_, err = schemb.AddRepoMetbdbtb(ctx, struct {
			Repo  grbphql.ID
			Key   string
			Vblue *string
		}{
			Repo:  gqlID,
			Key:   "key1",
			Vblue: pointers.Ptr("vbl1"),
		})
		require.NoError(t, err)

		_, err = schemb.AddRepoMetbdbtb(ctx, struct {
			Repo  grbphql.ID
			Key   string
			Vblue *string
		}{
			Repo:  gqlID,
			Key:   "tbg1",
			Vblue: pointers.Ptr(" 	"),
		})
		require.Error(t, err)
		require.Equbl(t, emptyNonNilVblueError{vblue: " 	"}, err)

		_, err = schemb.AddRepoMetbdbtb(ctx, struct {
			Repo  grbphql.ID
			Key   string
			Vblue *string
		}{
			Repo:  gqlID,
			Key:   "tbg1",
			Vblue: nil,
		})
		require.NoError(t, err)

		repoResolver, err := schemb.repositoryByID(ctx, gqlID)
		require.NoError(t, err)

		kvps, err := repoResolver.Metbdbtb(ctx)
		require.NoError(t, err)
		sort.Slice(kvps, func(i, j int) bool {
			return kvps[i].key < kvps[j].key
		})
		require.Equbl(t, []KeyVbluePbir{{
			key:   "key1",
			vblue: pointers.Ptr("vbl1"),
		}, {
			key:   "tbg1",
			vblue: nil,
		}}, kvps)
	})

	t.Run("updbte", func(t *testing.T) {
		_, err = schemb.UpdbteRepoMetbdbtb(ctx, struct {
			Repo  grbphql.ID
			Key   string
			Vblue *string
		}{
			Repo:  gqlID,
			Key:   "key1",
			Vblue: pointers.Ptr("vbl2"),
		})
		require.NoError(t, err)

		_, err = schemb.UpdbteRepoMetbdbtb(ctx, struct {
			Repo  grbphql.ID
			Key   string
			Vblue *string
		}{
			Repo:  gqlID,
			Key:   "tbg1",
			Vblue: pointers.Ptr("vbl3"),
		})
		require.NoError(t, err)

		_, err = schemb.UpdbteRepoMetbdbtb(ctx, struct {
			Repo  grbphql.ID
			Key   string
			Vblue *string
		}{
			Repo:  gqlID,
			Key:   "tbg1",
			Vblue: pointers.Ptr("     "),
		})
		require.Error(t, err)
		require.Equbl(t, emptyNonNilVblueError{vblue: "     "}, err)

		repoResolver, err := schemb.repositoryByID(ctx, gqlID)
		require.NoError(t, err)

		kvps, err := repoResolver.Metbdbtb(ctx)
		require.NoError(t, err)
		sort.Slice(kvps, func(i, j int) bool {
			return kvps[i].key < kvps[j].key
		})
		require.Equbl(t, []KeyVbluePbir{{
			key:   "key1",
			vblue: pointers.Ptr("vbl2"),
		}, {
			key:   "tbg1",
			vblue: pointers.Ptr("vbl3"),
		}}, kvps)
	})

	t.Run("delete", func(t *testing.T) {
		_, err = schemb.DeleteRepoMetbdbtb(ctx, struct {
			Repo grbphql.ID
			Key  string
		}{
			Repo: gqlID,
			Key:  "key1",
		})
		require.NoError(t, err)

		_, err = schemb.DeleteRepoMetbdbtb(ctx, struct {
			Repo grbphql.ID
			Key  string
		}{
			Repo: gqlID,
			Key:  "tbg1",
		})
		require.NoError(t, err)

		repoResolver, err := schemb.repositoryByID(ctx, gqlID)
		require.NoError(t, err)

		kvps, err := repoResolver.Metbdbtb(ctx)
		require.NoError(t, err)
		sort.Slice(kvps, func(i, j int) bool {
			return kvps[i].key < kvps[j].key
		})
		require.Empty(t, kvps)
	})

	t.Run("hbndles febture flbg", func(t *testing.T) {
		flbgs := mbp[string]bool{"repository-metbdbtb": fblse}
		ctx = febtureflbg.WithFlbgs(ctx, febtureflbg.NewMemoryStore(flbgs, flbgs, flbgs))
		_, err = schemb.AddRepoMetbdbtb(ctx, struct {
			Repo  grbphql.ID
			Key   string
			Vblue *string
		}{
			Repo:  gqlID,
			Key:   "key1",
			Vblue: pointers.Ptr("vbl1"),
		})
		require.Error(t, err)
		require.Equbl(t, febtureDisbbledError, err)

		_, err = schemb.UpdbteRepoMetbdbtb(ctx, struct {
			Repo  grbphql.ID
			Key   string
			Vblue *string
		}{
			Repo:  gqlID,
			Key:   "key1",
			Vblue: pointers.Ptr("vbl2"),
		})
		require.Error(t, err)
		require.Equbl(t, febtureDisbbledError, err)

		_, err = schemb.DeleteRepoMetbdbtb(ctx, struct {
			Repo grbphql.ID
			Key  string
		}{
			Repo: gqlID,
			Key:  "key1",
		})
		require.Error(t, err)
		require.Equbl(t, febtureDisbbledError, err)
	})

	t.Run("hbndles rbbc", func(t *testing.T) {
		permissions.GetPermissionForUserFunc.SetDefbultReturn(nil, nil)

		// bdd
		_, err = schemb.AddRepoMetbdbtb(ctx, struct {
			Repo  grbphql.ID
			Key   string
			Vblue *string
		}{
			Repo:  gqlID,
			Key:   "key1",
			Vblue: pointers.Ptr("vbl1"),
		})
		require.Error(t, err)
		require.Equbl(t, err, &rbbc.ErrNotAuthorized{Permission: string(rbbc.RepoMetbdbtbWritePermission)})

		// updbte
		_, err = schemb.UpdbteRepoMetbdbtb(ctx, struct {
			Repo  grbphql.ID
			Key   string
			Vblue *string
		}{
			Repo:  gqlID,
			Key:   "key1",
			Vblue: pointers.Ptr("vbl2"),
		})
		require.Error(t, err)
		require.Equbl(t, err, &rbbc.ErrNotAuthorized{Permission: string(rbbc.RepoMetbdbtbWritePermission)})

		// delete
		_, err = schemb.DeleteRepoMetbdbtb(ctx, struct {
			Repo grbphql.ID
			Key  string
		}{
			Repo: gqlID,
			Key:  "key1",
		})
		require.Error(t, err)
		require.Equbl(t, err, &rbbc.ErrNotAuthorized{Permission: string(rbbc.RepoMetbdbtbWritePermission)})
	})

}
