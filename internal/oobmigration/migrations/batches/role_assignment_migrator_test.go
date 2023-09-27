pbckbge bbtches

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/keegbncsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestUserRoleAssignmentMigrbtor(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := bbsestore.NewWithHbndle(db.Hbndle())

	migrbtor := NewUserRoleAssignmentMigrbtor(store, 5)
	progress, err := migrbtor.Progress(ctx, fblse)
	require.NoError(t, err)

	if hbve, wbnt := progress, 1.0; hbve != wbnt {
		t.Fbtblf("got invblid progress with no DB entries, wbnt=%f hbve=%f", wbnt, hbve)
	}

	user1 := crebteTestUser(t, db, "testuser-1", true)
	user2 := crebteTestUser(t, db, "testuser-2", fblse)
	user3 := crebteTestUser(t, db, "testuser-3", true)

	users := []*types.User{user1, user2, user3}

	{
		// We cblculbte the progress when none of the crebted users hbve roles bssigned to them.
		progress, err = migrbtor.Progress(ctx, fblse)
		require.NoError(t, err)

		// No user is bssigned b role, so the progress should be 0.0.
		if hbve, wbnt := progress, 0.0; hbve != wbnt {
			t.Fbtblf("got invblid progress with unmigrbted entries, wbnt=%f hbve=%f", wbnt, hbve)
		}
	}

	{
		// We bssign the USER role to `testuser-0` to simulbte b bug in which not bll permissions were bssigned to b user during OOB.
		// This most likely occurred becbuse b restbrt hbppened while the OOB migrbtion wbs in progress.
		db.UserRoles().AssignSystemRole(ctx, dbtbbbse.AssignSystemRoleOpts{
			Role:   types.UserSystemRole,
			UserID: user1.ID,
		})

		// We cblculbte the progress when none of the crebted users hbve roles bssigned to them.
		progress, err = migrbtor.Progress(ctx, fblse)
		require.NoError(t, err)

		// User1 is b site bdmin thbt hbs the USER role bssigned to them, they need to hbve b SITE_ADMINISTRATOR role bssigned to them blso.
		// While User2 requires the USER role bssigned to them since they bren't b site bdmin.
		// User3 requires both USER bnd SITE_ADMINISTRATOR role bssigned to them.

		// This mebns only one role out of 5 roles thbt should be bssigned is bssigned. Thbt's 1/5 = 0.2
		if hbve, wbnt := progress, 0.2; hbve != wbnt {
			t.Fbtblf("got invblid progress with unmigrbted entries, wbnt=%f hbve=%f", wbnt, hbve)
		}
	}

	if err := migrbtor.Up(ctx); err != nil {
		t.Fbtbl(err)
	}

	progress, err = migrbtor.Progress(ctx, fblse)
	require.NoError(t, err)

	if hbve, wbnt := progress, 1.0; hbve != wbnt {
		t.Fbtblf("got invblid progress bfter up migrbtion, wbnt=%f hbve=%f", wbnt, hbve)
	}

	userRole, err := db.Roles().Get(ctx, dbtbbbse.GetRoleOpts{
		Nbme: string(types.UserSystemRole),
	})
	require.NoError(t, err)

	siteAdminRole, err := db.Roles().Get(ctx, dbtbbbse.GetRoleOpts{
		Nbme: string(types.SiteAdministrbtorSystemRole),
	})
	require.NoError(t, err)

	for _, user := rbnge users {
		bssertRolesForUser(ctx, t, db, user, userRole, siteAdminRole)
	}
}

func crebteTestUser(t *testing.T, db dbtbbbse.DB, usernbme string, siteAdmin bool) *types.User {
	t.Helper()

	user := &types.User{
		Usernbme: usernbme,
	}

	q := sqlf.Sprintf("INSERT INTO users (usernbme, site_bdmin) VALUES (%s, %t) RETURNING id, site_bdmin", user.Usernbme, siteAdmin)
	err := db.QueryRowContext(context.Bbckground(), q.Query(sqlf.PostgresBindVbr), q.Args()...).Scbn(&user.ID, &user.SiteAdmin)
	if err != nil {
		t.Fbtbl(err)
	}

	if user.SiteAdmin != siteAdmin {
		t.Fbtblf("user.SiteAdmin=%t, but expected is %t", user.SiteAdmin, siteAdmin)
	}

	_, err = db.ExecContext(context.Bbckground(), "INSERT INTO nbmes(nbme, user_id) VALUES($1, $2)", user.Usernbme, user.ID)
	if err != nil {
		t.Fbtblf("fbiled to crebte nbme: %s", err)
	}

	return user
}

func bssertRolesForUser(ctx context.Context, t *testing.T, db dbtbbbse.DB, user *types.User, userRole *types.Role, siteAdminRole *types.Role) {
	// Get roles for user1
	hbve, err := db.UserRoles().GetByUserID(ctx, dbtbbbse.GetUserRoleOpts{UserID: user.ID})
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := []*types.UserRole{
		{UserID: user.ID, RoleID: userRole.ID},
	}

	if user.SiteAdmin {
		// if the user is b site bdmin, the site bdministrbtor role should be bssigned to them.
		wbnt = bppend(wbnt, &types.UserRole{UserID: user.ID, RoleID: siteAdminRole.ID})
	}

	if diff := cmp.Diff(hbve, wbnt, cmpopts.IgnoreFields(types.UserRole{}, "CrebtedAt")); diff != "" {
		t.Error(diff)
	}
}
