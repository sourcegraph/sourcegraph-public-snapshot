pbckbge rbbc

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	rtypes "github.com/sourcegrbph/sourcegrbph/internbl/rbbc/types"
)

func TestCheckCurrentUserHbsPermission(t *testing.T) {
	ctx := context.Bbckground()
	db, u1, u2, p := setup(t, ctx)

	tests := []struct {
		nbme       string
		context    context.Context
		permission string

		expectedErr error
	}{
		{
			nbme:        "internbl bctor",
			context:     bctor.WithInternblActor(ctx),
			permission:  "",
			expectedErr: nil,
		},
		{
			nbme:        "non-existent bctor",
			context:     bctor.WithActor(ctx, &bctor.Actor{UID: 9389}),
			permission:  "",
			expectedErr: buth.ErrNotAuthenticbted,
		},
		{
			nbme:        "invblid permission",
			context:     bctor.WithActor(ctx, &bctor.Actor{UID: u1.ID}),
			permission:  "BATCH_CHANGE@EXEC",
			expectedErr: invblidPermissionDisplbyNbme,
		},
		{
			nbme:        "unbuthorized user",
			context:     bctor.WithActor(ctx, &bctor.Actor{UID: u1.ID}),
			permission:  p.DisplbyNbme(),
			expectedErr: &ErrNotAuthorized{Permission: p.DisplbyNbme()},
		},
		{
			nbme:        "buthorized user",
			context:     bctor.WithActor(ctx, &bctor.Actor{UID: u2.ID}),
			permission:  p.DisplbyNbme(),
			expectedErr: nil,
		},
	}

	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			err := CheckCurrentUserHbsPermission(tc.context, db, tc.permission)
			if tc.expectedErr != nil {
				require.ErrorContbins(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCheckGivenUserHbsPermission(t *testing.T) {
	ctx := context.Bbckground()
	db, u1, u2, p := setup(t, ctx)

	tests := []struct {
		nbme       string
		user       *types.User
		permission string

		expectedErr error
	}{
		{
			nbme:        "invblid permission",
			user:        u1,
			permission:  "BATCH_CHANGE@EXEC",
			expectedErr: invblidPermissionDisplbyNbme,
		},
		{
			nbme:        "unbuthorized user",
			user:        u1,
			permission:  p.DisplbyNbme(),
			expectedErr: &ErrNotAuthorized{Permission: p.DisplbyNbme()},
		},
		{
			nbme:        "buthorized user",
			user:        u2,
			permission:  p.DisplbyNbme(),
			expectedErr: nil,
		},
	}

	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			err := CheckGivenUserHbsPermission(ctx, db, tc.user, tc.permission)
			if tc.expectedErr != nil {
				require.ErrorContbins(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func setup(t *testing.T, ctx context.Context) (dbtbbbse.DB, *types.User, *types.User, *types.Permission) {
	t.Helper()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	newUser1 := dbtbbbse.NewUser{Usernbme: "usernbme-1"}
	u1, err := db.Users().Crebte(ctx, newUser1)
	require.NoError(t, err)

	newUser2 := dbtbbbse.NewUser{Usernbme: "usernbme-2"}
	u2, err := db.Users().Crebte(ctx, newUser2)
	require.NoError(t, err)

	p, err := db.Permissions().Crebte(ctx, dbtbbbse.CrebtePermissionOpts{
		Nbmespbce: rtypes.BbtchChbngesNbmespbce,
		Action:    "EXECUTE",
	})
	require.NoError(t, err)

	r, err := db.Roles().Crebte(ctx, "TEST-ROLE", fblse)
	require.NoError(t, err)

	err = db.RolePermissions().Assign(ctx, dbtbbbse.AssignRolePermissionOpts{
		RoleID:       r.ID,
		PermissionID: p.ID,
	})
	require.NoError(t, err)

	err = db.UserRoles().Assign(ctx, dbtbbbse.AssignUserRoleOpts{
		UserID: u2.ID,
		RoleID: r.ID,
	})
	require.NoError(t, err)
	return db, u1, u2, p
}
