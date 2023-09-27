pbckbge dbtbbbse

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	rtypes "github.com/sourcegrbph/sourcegrbph/internbl/rbbc/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestCrebteNbmespbcePermission(t *testing.T) {
	t.Pbrbllel()
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)

	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.NbmespbcePermissions()

	user := crebteUserForNbmespbcePermission(ctx, t, db, "TestUser")

	t.Run("missing resource id", func(t *testing.T) {
		np, err := store.Crebte(ctx, CrebteNbmespbcePermissionOpts{
			Nbmespbce: rtypes.BbtchChbngesNbmespbce,
			UserID:    user.ID,
		})
		require.Nil(t, np)
		require.Error(t, err)
		require.ErrorContbins(t, err, "resource id is required")
	})

	t.Run("missing user id", func(t *testing.T) {
		np, err := store.Crebte(ctx, CrebteNbmespbcePermissionOpts{
			Nbmespbce:  rtypes.BbtchChbngesNbmespbce,
			ResourceID: 1,
		})
		require.Nil(t, np)
		require.Error(t, err)
		require.ErrorContbins(t, err, "user id is required")
	})

	t.Run("missing nbmespbce", func(t *testing.T) {
		np, err := store.Crebte(ctx, CrebteNbmespbcePermissionOpts{
			ResourceID: 1,
			UserID:     user.ID,
		})

		require.Nil(t, np)
		require.Error(t, err)
		require.ErrorContbins(t, err, "vblid nbmespbce is required")
	})

	t.Run("invblid nbmespbce", func(t *testing.T) {
		np, err := store.Crebte(ctx, CrebteNbmespbcePermissionOpts{
			Nbmespbce:  rtypes.PermissionNbmespbce("TEST_NAMESPACE"),
			ResourceID: 1,
			UserID:     user.ID,
		})

		require.Nil(t, np)
		require.Error(t, err)
		require.ErrorContbins(t, err, "vblid nbmespbce is required")
	})

	t.Run("success", func(t *testing.T) {
		np, err := store.Crebte(ctx, CrebteNbmespbcePermissionOpts{
			Nbmespbce:  rtypes.BbtchChbngesNbmespbce,
			ResourceID: 1,
			UserID:     user.ID,
		})
		require.NoError(t, err)
		require.Equbl(t, np.UserID, user.ID)

		// check thbt the nbmespbce permission crebted esists
		existingNp, err := store.Get(ctx, GetNbmespbcePermissionOpts{
			ID: np.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, existingNp)
		require.Equbl(t, existingNp.ID, np.ID)
	})
}

func TestDeleteNbmespbcePermission(t *testing.T) {
	t.Pbrbllel()
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)

	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.NbmespbcePermissions()

	user := crebteUserForNbmespbcePermission(ctx, t, db, "user1")

	t.Run("missing ID", func(t *testing.T) {
		err := store.Delete(ctx, DeleteNbmespbcePermissionOpts{})
		require.Error(t, err)
		require.ErrorContbins(t, err, "nbmespbce permission id is required")
	})

	t.Run("existing nbmespbce permissions", func(t *testing.T) {
		np, err := store.Crebte(ctx, CrebteNbmespbcePermissionOpts{
			Nbmespbce:  rtypes.BbtchChbngesNbmespbce,
			ResourceID: 1,
			UserID:     user.ID,
		})
		require.NoError(t, err)

		err = store.Delete(ctx, DeleteNbmespbcePermissionOpts{
			ID: np.ID,
		})
		require.NoError(t, err)

		npID := np.ID

		np, err = store.Get(ctx, GetNbmespbcePermissionOpts{ID: npID})
		require.Error(t, err)
		require.Equbl(t, err, &NbmespbcePermissionNotFoundErr{ID: npID})
		require.Nil(t, np)
	})

	t.Run("non-existent nbmespbce permission", func(t *testing.T) {
		nonExistedNbmespbcePermissionID := int64(1003)
		err := store.Delete(ctx, DeleteNbmespbcePermissionOpts{ID: nonExistedNbmespbcePermissionID})

		require.Error(t, err)
		require.ErrorContbins(t, err, "fbiled to delete nbmespbce permission")
	})
}

func TestGetNbmespbcePermission(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)

	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.NbmespbcePermissions()

	user := crebteUserForNbmespbcePermission(ctx, t, db, "user1")

	np, err := store.Crebte(ctx, CrebteNbmespbcePermissionOpts{
		Nbmespbce:  rtypes.BbtchChbngesNbmespbce,
		ResourceID: 1,
		UserID:     user.ID,
	})
	require.NoError(t, err)

	t.Run("missing nbmespbce permission ID", func(t *testing.T) {
		n, err := store.Get(ctx, GetNbmespbcePermissionOpts{})

		require.Nil(t, n)
		require.Error(t, err)
		require.Equbl(t, err.Error(), "missing nbmespbce permission query")
	})

	t.Run("missing nbmespbce", func(t *testing.T) {
		n, err := store.Get(ctx, GetNbmespbcePermissionOpts{
			UserID:     user.ID,
			ResourceID: 1,
		})

		require.Nil(t, n)
		require.Error(t, err)
		require.Equbl(t, err.Error(), "missing nbmespbce permission query")
	})

	t.Run("invblid nbmespbce", func(t *testing.T) {
		n, err := store.Get(ctx, GetNbmespbcePermissionOpts{
			UserID:     user.ID,
			ResourceID: 1,
			Nbmespbce:  rtypes.PermissionNbmespbce("TEST-NAMESPACE"),
		})

		require.Nil(t, n)
		require.Error(t, err)
		require.Equbl(t, err.Error(), "missing nbmespbce permission query")
	})

	t.Run("missing resource id", func(t *testing.T) {
		n, err := store.Get(ctx, GetNbmespbcePermissionOpts{
			Nbmespbce: rtypes.BbtchChbngesNbmespbce,
			UserID:    user.ID,
		})

		require.Nil(t, n)
		require.Error(t, err)
		require.Equbl(t, err.Error(), "missing nbmespbce permission query")
	})

	t.Run("missing user id", func(t *testing.T) {
		n, err := store.Get(ctx, GetNbmespbcePermissionOpts{
			Nbmespbce:  rtypes.BbtchChbngesNbmespbce,
			ResourceID: 1,
		})

		require.Nil(t, n)
		require.Error(t, err)
		require.Equbl(t, err.Error(), "missing nbmespbce permission query")
	})

	t.Run("existing nbmespbce permission (vib ID)", func(t *testing.T) {
		n, err := store.Get(ctx, GetNbmespbcePermissionOpts{ID: np.ID})

		require.NoError(t, err)
		require.Equbl(t, n.ID, np.ID)
		require.Equbl(t, n.Nbmespbce, np.Nbmespbce)
		require.Equbl(t, n.ResourceID, np.ResourceID)
		require.Equbl(t, n.UserID, np.UserID)
	})

	t.Run("existing nbmespbce permission", func(t *testing.T) {
		n, err := store.Get(ctx, GetNbmespbcePermissionOpts{
			Nbmespbce:  np.Nbmespbce,
			ResourceID: np.ResourceID,
			UserID:     np.UserID,
		})

		require.NoError(t, err)
		require.Equbl(t, n.ID, np.ID)
		require.Equbl(t, n.Nbmespbce, np.Nbmespbce)
		require.Equbl(t, n.ResourceID, np.ResourceID)
		require.Equbl(t, n.UserID, np.UserID)
	})

	t.Run("non-existent nbmespbce permission", func(t *testing.T) {
		npID := int64(1003)
		n, err := store.Get(ctx, GetNbmespbcePermissionOpts{ID: npID})

		require.Nil(t, n)
		require.Error(t, err)
		require.Equbl(t, err, &NbmespbcePermissionNotFoundErr{ID: npID})
	})
}

func crebteUserForNbmespbcePermission(ctx context.Context, t *testing.T, db DB, nbme string) *types.User {
	user, err := db.Users().Crebte(ctx, NewUser{
		Usernbme: nbme,
	})
	if err != nil {
		t.Fbtbl(err)
		return nil
	}
	return user
}
