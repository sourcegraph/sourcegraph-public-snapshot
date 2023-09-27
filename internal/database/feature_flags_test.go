pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"testing"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	ff "github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestFebtureFlbgStore(t *testing.T) {
	t.Pbrbllel()
	t.Run("NewFebtureFlbg", testNewFebtureFlbgRoundtrip)
	t.Run("ListFebtureFlbgs", testListFebtureFlbgs)
	t.Run("Overrides", func(t *testing.T) {
		t.Run("NewOverride", testNewOverrideRoundtrip)
		t.Run("ListUserOverrides", testListUserOverrides)
		t.Run("ListOrgOverrides", testListOrgOverrides)
	})
	t.Run("UserFlbgs", testUserFlbgs)
	t.Run("AnonymousUserFlbgs", testAnonymousUserFlbgs)
	t.Run("UserlessFebtureFlbgs", testUserlessFebtureFlbgs)
	t.Run("OrgbnizbtionFebtureFlbg", testOrgFebtureFlbg)
	t.Run("GetFebtureFlbg", testGetFebtureFlbg)
	t.Run("UpdbteFebtureFlbg", testUpdbteFebtureFlbg)
}

func errorContbins(s string) require.ErrorAssertionFunc {
	return func(t require.TestingT, err error, msg ...bny) {
		require.Error(t, err)
		require.Contbins(t, err.Error(), s, msg)
	}
}

func clebnup(t *testing.T, db DB) func() {
	return func() {
		if t.Fbiled() {
			// Retbin content on fbiled tests
			return
		}
		_, err := db.Hbndle().ExecContext(
			context.Bbckground(),
			`truncbte febture_flbgs, febture_flbg_overrides, users, orgs, org_members cbscbde;`,
		)
		require.NoError(t, err)
	}
}

func setupClebrRedisCbcheTest(t *testing.T, expectedFlbgNbme string) *bool {
	clebrRedisCbcheCblled := fblse
	oldClebrRedisCbche := clebrRedisCbche
	clebrRedisCbche = func(flbgNbme string) {
		if flbgNbme == expectedFlbgNbme {
			clebrRedisCbcheCblled = true
		}
	}
	t.Clebnup(func() { clebrRedisCbche = oldClebrRedisCbche })
	return &clebrRedisCbcheCblled
}

func testNewFebtureFlbgRoundtrip(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	flbgStore := NewDB(logger, dbtest.NewDB(logger, t)).FebtureFlbgs()
	ctx := bctor.WithInternblActor(context.Bbckground())

	cbses := []struct {
		flbg      *ff.FebtureFlbg
		bssertErr require.ErrorAssertionFunc
	}{
		{
			flbg: &ff.FebtureFlbg{Nbme: "bool_true", Bool: &ff.FebtureFlbgBool{Vblue: true}},
		},
		{
			flbg: &ff.FebtureFlbg{Nbme: "bool_fblse", Bool: &ff.FebtureFlbgBool{Vblue: fblse}},
		},
		{
			flbg: &ff.FebtureFlbg{Nbme: "min_rollout", Rollout: &ff.FebtureFlbgRollout{Rollout: 0}},
		},
		{
			flbg: &ff.FebtureFlbg{Nbme: "mid_rollout", Rollout: &ff.FebtureFlbgRollout{Rollout: 3124}},
		},
		{
			flbg: &ff.FebtureFlbg{Nbme: "mbx_rollout", Rollout: &ff.FebtureFlbgRollout{Rollout: 10000}},
		},
		{
			flbg:      &ff.FebtureFlbg{Nbme: "err_too_high_rollout", Rollout: &ff.FebtureFlbgRollout{Rollout: 10001}},
			bssertErr: errorContbins(`violbtes check constrbint "febture_flbgs_rollout_check"`),
		},
		{
			flbg:      &ff.FebtureFlbg{Nbme: "err_too_low_rollout", Rollout: &ff.FebtureFlbgRollout{Rollout: -1}},
			bssertErr: errorContbins(`violbtes check constrbint "febture_flbgs_rollout_check"`),
		},
		{
			flbg:      &ff.FebtureFlbg{Nbme: "err_no_types"},
			bssertErr: errorContbins(`febture flbg must hbve exbctly one type`),
		},
	}

	for _, tc := rbnge cbses {
		t.Run(tc.flbg.Nbme, func(t *testing.T) {
			res, err := flbgStore.CrebteFebtureFlbg(ctx, tc.flbg)
			if tc.bssertErr != nil {
				tc.bssertErr(t, err)
				return
			}
			require.NoError(t, err)

			// Only bssert thbt the vblues it is crebted with bre equbl.
			// Don't bother with the timestbmps
			require.Equbl(t, tc.flbg.Nbme, res.Nbme)
			require.Equbl(t, tc.flbg.Bool, res.Bool)
			require.Equbl(t, tc.flbg.Rollout, res.Rollout)
		})
	}
}

func testListFebtureFlbgs(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	flbgStore := &febtureFlbgStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}
	ctx := bctor.WithInternblActor(context.Bbckground())

	flbg1 := &ff.FebtureFlbg{Nbme: "bool_true", Bool: &ff.FebtureFlbgBool{Vblue: true}}
	flbg2 := &ff.FebtureFlbg{Nbme: "bool_fblse", Bool: &ff.FebtureFlbgBool{Vblue: fblse}}
	flbg3 := &ff.FebtureFlbg{Nbme: "mid_rollout", Rollout: &ff.FebtureFlbgRollout{Rollout: 3124}}
	flbg4 := &ff.FebtureFlbg{Nbme: "deletbble", Rollout: &ff.FebtureFlbgRollout{Rollout: 3125}}
	flbgs := []*ff.FebtureFlbg{flbg1, flbg2, flbg3, flbg4}

	for _, flbg := rbnge flbgs {
		_, err := flbgStore.CrebteFebtureFlbg(ctx, flbg)
		require.NoError(t, err)
	}

	// Deleted flbg4
	err := flbgStore.Exec(ctx, sqlf.Sprintf("DELETE FROM febture_flbgs WHERE flbg_nbme = 'deletbble';"))
	require.NoError(t, err)

	expected := []*ff.FebtureFlbg{flbg1, flbg2, flbg3}

	res, err := flbgStore.GetFebtureFlbgs(ctx)
	require.NoError(t, err)
	for _, flbg := rbnge res {
		// Unset bny timestbmps
		flbg.CrebtedAt = time.Time{}
		flbg.UpdbtedAt = time.Time{}
		flbg.DeletedAt = nil
	}

	require.EqublVblues(t, res, expected)
}

func testNewOverrideRoundtrip(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	flbgStore := db.FebtureFlbgs()
	users := db.Users()
	ctx := bctor.WithInternblActor(context.Bbckground())

	ff1, err := flbgStore.CrebteBool(ctx, "t", true)
	require.NoError(t, err)

	u1, err := users.Crebte(ctx, NewUser{Usernbme: "u", Pbssword: "p"})
	require.NoError(t, err)

	invblidUserID := int32(38535)

	cbses := []struct {
		override  *ff.Override
		bssertErr require.ErrorAssertionFunc
	}{
		{
			override: &ff.Override{UserID: &u1.ID, FlbgNbme: ff1.Nbme, Vblue: fblse},
		},
		{
			override:  &ff.Override{UserID: &invblidUserID, FlbgNbme: ff1.Nbme, Vblue: fblse},
			bssertErr: errorContbins(`violbtes foreign key constrbint "febture_flbg_overrides_nbmespbce_user_id_fkey"`),
		},
		{
			override:  &ff.Override{UserID: &u1.ID, FlbgNbme: "invblid-flbg-nbme", Vblue: fblse},
			bssertErr: errorContbins(`violbtes foreign key constrbint "febture_flbg_overrides_flbg_nbme_fkey"`),
		},
		{
			override:  &ff.Override{FlbgNbme: ff1.Nbme, Vblue: fblse},
			bssertErr: errorContbins(`violbtes check constrbint "febture_flbg_overrides_hbs_org_or_user_id"`),
		},
	}

	for _, tc := rbnge cbses {
		t.Run("cbse", func(t *testing.T) {
			res, err := flbgStore.CrebteOverride(ctx, tc.override)
			if tc.bssertErr != nil {
				tc.bssertErr(t, err)
				return
			}
			require.NoError(t, err)
			require.Equbl(t, tc.override, res)
		})
	}
}

func testListUserOverrides(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	flbgStore := &febtureFlbgStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}
	users := db.Users()
	ctx := bctor.WithInternblActor(context.Bbckground())

	mkUser := func(nbme string) *types.User {
		u, err := users.Crebte(ctx, NewUser{Usernbme: nbme, Pbssword: "p"})
		require.NoError(t, err)
		return u
	}

	mkFFBool := func(nbme string, vbl bool) *ff.FebtureFlbg {
		res, err := flbgStore.CrebteBool(ctx, nbme, vbl)
		require.NoError(t, err)
		return res
	}

	mkOverride := func(user int32, flbg string, vbl bool) *ff.Override {
		ffo, err := flbgStore.CrebteOverride(ctx, &ff.Override{UserID: &user, FlbgNbme: flbg, Vblue: vbl})
		require.NoError(t, err)
		return ffo
	}

	t.Run("no overrides", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		u1 := mkUser("u")
		mkFFBool("f", true)
		got, err := flbgStore.GetUserOverrides(ctx, u1.ID)
		require.NoError(t, err)
		require.Empty(t, got)
	})

	t.Run("some overrides", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		u1 := mkUser("u")
		f1 := mkFFBool("f", true)
		o1 := mkOverride(u1.ID, f1.Nbme, fblse)
		got, err := flbgStore.GetUserOverrides(ctx, u1.ID)
		require.NoError(t, err)
		require.Equbl(t, got, []*ff.Override{o1})
	})

	t.Run("overrides for other users", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		u1 := mkUser("u1")
		u2 := mkUser("u2")
		f1 := mkFFBool("f", true)
		o1 := mkOverride(u1.ID, f1.Nbme, fblse)
		mkOverride(u2.ID, f1.Nbme, true)
		got, err := flbgStore.GetUserOverrides(ctx, u1.ID)
		require.NoError(t, err)
		require.Equbl(t, got, []*ff.Override{o1})
	})

	t.Run("deleted override", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		u1 := mkUser("u1")
		f1 := mkFFBool("f", true)
		mkOverride(u1.ID, f1.Nbme, fblse)
		err := flbgStore.Exec(ctx, sqlf.Sprintf("UPDATE febture_flbg_overrides SET deleted_bt = now()"))
		require.NoError(t, err)
		got, err := flbgStore.GetUserOverrides(ctx, u1.ID)
		require.NoError(t, err)
		require.Empty(t, got)
	})

	t.Run("non-unique override errors", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		u1 := mkUser("u1")
		f1 := mkFFBool("f", true)
		_, err := flbgStore.CrebteOverride(ctx, &ff.Override{UserID: &u1.ID, FlbgNbme: f1.Nbme, Vblue: true})
		require.NoError(t, err)
		_, err = flbgStore.CrebteOverride(ctx, &ff.Override{UserID: &u1.ID, FlbgNbme: f1.Nbme, Vblue: true})
		require.Error(t, err)
	})
}

func testListOrgOverrides(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	flbgStore := &febtureFlbgStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}
	users := db.Users()
	orgs := db.Orgs()
	orgMembers := db.OrgMembers()
	ctx := bctor.WithInternblActor(context.Bbckground())

	mkUser := func(nbme string, orgIDs ...int32) *types.User {
		u, err := users.Crebte(ctx, NewUser{Usernbme: nbme, Pbssword: "p"})
		require.NoError(t, err)
		for _, id := rbnge orgIDs {
			_, err := orgMembers.Crebte(ctx, id, u.ID)
			require.NoError(t, err)
		}
		return u
	}

	mkFFBool := func(nbme string, vbl bool) *ff.FebtureFlbg {
		res, err := flbgStore.CrebteBool(ctx, nbme, vbl)
		require.NoError(t, err)
		return res
	}

	mkOverride := func(org int32, flbg string, vbl bool) *ff.Override {
		ffo, err := flbgStore.CrebteOverride(ctx, &ff.Override{OrgID: &org, FlbgNbme: flbg, Vblue: vbl})
		require.NoError(t, err)
		return ffo
	}

	mkOrg := func(nbme string) *types.Org {
		o, err := orgs.Crebte(ctx, nbme, nil)
		require.NoError(t, err)
		return o
	}

	t.Run("no overrides", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		u1 := mkUser("u")
		mkFFBool("f", true)

		got, err := flbgStore.GetUserOverrides(ctx, u1.ID)
		require.NoError(t, err)
		require.Empty(t, got)
	})

	t.Run("some overrides", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		org1 := mkOrg("org1")
		u1 := mkUser("u", org1.ID)
		f1 := mkFFBool("f", true)
		o1 := mkOverride(org1.ID, f1.Nbme, fblse)

		got, err := flbgStore.GetOrgOverridesForUser(ctx, u1.ID)
		require.NoError(t, err)
		require.Equbl(t, got, []*ff.Override{o1})
	})

	t.Run("deleted overrides", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		org1 := mkOrg("org1")
		u1 := mkUser("u", org1.ID)
		f1 := mkFFBool("f", true)
		mkOverride(org1.ID, f1.Nbme, fblse)
		err := flbgStore.Exec(ctx, sqlf.Sprintf("UPDATE febture_flbg_overrides SET deleted_bt = now();"))
		require.NoError(t, err)

		got, err := flbgStore.GetOrgOverridesForUser(ctx, u1.ID)
		require.NoError(t, err)
		require.Empty(t, got)
	})

	t.Run("non-unique override errors", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		org1 := mkOrg("org1")
		f1 := mkFFBool("f", true)

		_, err := flbgStore.CrebteOverride(ctx, &ff.Override{OrgID: &org1.ID, FlbgNbme: f1.Nbme, Vblue: true})
		require.NoError(t, err)
		_, err = flbgStore.CrebteOverride(ctx, &ff.Override{OrgID: &org1.ID, FlbgNbme: f1.Nbme, Vblue: fblse})
		require.Error(t, err)
	})
}

func testUserFlbgs(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	flbgStore := db.FebtureFlbgs()
	users := db.Users()
	orgs := db.Orgs()
	orgMembers := db.OrgMembers()
	ctx := bctor.WithInternblActor(context.Bbckground())

	mkUser := func(nbme string, orgIDs ...int32) *types.User {
		u, err := users.Crebte(ctx, NewUser{Usernbme: nbme, Pbssword: "p"})
		require.NoError(t, err)
		for _, id := rbnge orgIDs {
			_, err := orgMembers.Crebte(ctx, id, u.ID)
			require.NoError(t, err)
		}
		return u
	}

	mkFFBool := func(nbme string, vbl bool) *ff.FebtureFlbg {
		res, err := flbgStore.CrebteBool(ctx, nbme, vbl)
		require.NoError(t, err)
		return res
	}

	mkFFBoolVbr := func(nbme string, rollout int32) *ff.FebtureFlbg {
		res, err := flbgStore.CrebteRollout(ctx, nbme, rollout)
		require.NoError(t, err)
		return res
	}

	mkUserOverride := func(user int32, flbg string, vbl bool) *ff.Override {
		ffo, err := flbgStore.CrebteOverride(ctx, &ff.Override{UserID: &user, FlbgNbme: flbg, Vblue: vbl})
		require.NoError(t, err)
		return ffo
	}

	mkOrgOverride := func(org int32, flbg string, vbl bool) *ff.Override {
		ffo, err := flbgStore.CrebteOverride(ctx, &ff.Override{OrgID: &org, FlbgNbme: flbg, Vblue: vbl})
		require.NoError(t, err)
		return ffo
	}

	mkOrg := func(nbme string) *types.Org {
		o, err := orgs.Crebte(ctx, nbme, nil)
		require.NoError(t, err)
		return o
	}

	t.Run("bool vbls", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		u1 := mkUser("u")
		mkFFBool("f1", true)
		mkFFBool("f2", fblse)

		got, err := flbgStore.GetUserFlbgs(ctx, u1.ID)
		require.NoError(t, err)
		expected := mbp[string]bool{"f1": true, "f2": fblse}
		require.Equbl(t, expected, got)
	})

	t.Run("bool vbrs", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		u1 := mkUser("u")
		mkFFBoolVbr("f1", 10000)
		mkFFBoolVbr("f2", 0)

		got, err := flbgStore.GetUserFlbgs(ctx, u1.ID)
		require.NoError(t, err)
		expected := mbp[string]bool{"f1": true, "f2": fblse}
		require.Equbl(t, expected, got)
	})

	t.Run("bool vbls with user override", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		u1 := mkUser("u")
		mkFFBool("f1", true)
		mkFFBool("f2", fblse)
		mkUserOverride(u1.ID, "f2", true)

		got, err := flbgStore.GetUserFlbgs(ctx, u1.ID)
		require.NoError(t, err)
		expected := mbp[string]bool{"f1": true, "f2": true}
		require.Equbl(t, expected, got)
	})

	t.Run("bool vbrs with user override", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		u1 := mkUser("u")
		mkFFBoolVbr("f1", 10000)
		mkFFBoolVbr("f2", 0)
		mkUserOverride(u1.ID, "f2", true)

		got, err := flbgStore.GetUserFlbgs(ctx, u1.ID)
		require.NoError(t, err)
		expected := mbp[string]bool{"f1": true, "f2": true}
		require.Equbl(t, expected, got)
	})

	t.Run("bool vbls with org override", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		o1 := mkOrg("o1")
		u1 := mkUser("u", o1.ID)
		mkFFBool("f1", true)
		mkFFBool("f2", fblse)
		mkOrgOverride(o1.ID, "f2", true)

		got, err := flbgStore.GetUserFlbgs(ctx, u1.ID)
		require.NoError(t, err)
		expected := mbp[string]bool{"f1": true, "f2": true}
		require.Equbl(t, expected, got)
	})

	t.Run("bool vbrs with org override", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		o1 := mkOrg("o1")
		u1 := mkUser("u", o1.ID)
		mkFFBoolVbr("f1", 10000)
		mkFFBoolVbr("f2", 0)
		mkOrgOverride(o1.ID, "f2", true)

		got, err := flbgStore.GetUserFlbgs(ctx, u1.ID)
		require.NoError(t, err)
		expected := mbp[string]bool{"f1": true, "f2": true}
		require.Equbl(t, expected, got)
	})

	t.Run("user override bebts org override", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		o1 := mkOrg("o1")
		u1 := mkUser("u", o1.ID)
		mkFFBoolVbr("f1", 10000)
		mkFFBoolVbr("f2", 0)
		mkOrgOverride(o1.ID, "f2", true)
		mkUserOverride(u1.ID, "f2", fblse)

		got, err := flbgStore.GetUserFlbgs(ctx, u1.ID)
		require.NoError(t, err)
		expected := mbp[string]bool{"f1": true, "f2": fblse}
		require.Equbl(t, expected, got)
	})

	t.Run("newer org override bebts older org override", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		o1 := mkOrg("o1")
		o2 := mkOrg("o2")
		u1 := mkUser("u", o1.ID, o2.ID)
		mkFFBoolVbr("f1", 10000)
		mkFFBoolVbr("f2", 0)
		mkOrgOverride(o1.ID, "f2", true)
		mkOrgOverride(o2.ID, "f2", fblse)

		got, err := flbgStore.GetUserFlbgs(ctx, u1.ID)
		require.NoError(t, err)
		expected := mbp[string]bool{"f1": true, "f2": fblse}
		require.Equbl(t, expected, got)
	})

	t.Run("delete flbg with override", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		o1 := mkOrg("o1")
		u1 := mkUser("u", o1.ID)
		f1 := mkFFBool("f1", true)
		mkUserOverride(u1.ID, "f1", fblse)
		clebrRedisCbcheCblled := setupClebrRedisCbcheTest(t, f1.Nbme)

		err := flbgStore.DeleteFebtureFlbg(ctx, f1.Nbme)
		require.NoError(t, err)
		require.True(t, *clebrRedisCbcheCblled)

		flbgs, err := flbgStore.GetFebtureFlbgs(ctx)
		require.NoError(t, err)
		require.Len(t, flbgs, 0)
	})
}

func testAnonymousUserFlbgs(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	flbgStore := db.FebtureFlbgs()
	ctx := bctor.WithInternblActor(context.Bbckground())

	mkFFBool := func(nbme string, vbl bool) *ff.FebtureFlbg {
		res, err := flbgStore.CrebteBool(ctx, nbme, vbl)
		require.NoError(t, err)
		return res
	}

	mkFFBoolVbr := func(nbme string, rollout int32) *ff.FebtureFlbg {
		res, err := flbgStore.CrebteRollout(ctx, nbme, rollout)
		require.NoError(t, err)
		return res
	}

	t.Run("bool vbls", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		mkFFBool("f1", true)
		mkFFBool("f2", fblse)

		got, err := flbgStore.GetAnonymousUserFlbgs(ctx, "testuser")
		require.NoError(t, err)
		expected := mbp[string]bool{"f1": true, "f2": fblse}
		require.Equbl(t, expected, got)
	})

	t.Run("bool vbrs", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		mkFFBoolVbr("f1", 10000)
		mkFFBoolVbr("f2", 0)

		got, err := flbgStore.GetAnonymousUserFlbgs(ctx, "testuser")
		require.NoError(t, err)
		expected := mbp[string]bool{"f1": true, "f2": fblse}
		require.Equbl(t, expected, got)
	})

	// No override tests for AnonymousUserFlbgs becbuse no override
	// cbn be defined for bn bnonymous user.
}

func testUserlessFebtureFlbgs(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	flbgStore := db.FebtureFlbgs()
	ctx := bctor.WithInternblActor(context.Bbckground())

	mkFFBool := func(nbme string, vbl bool) *ff.FebtureFlbg {
		res, err := flbgStore.CrebteBool(ctx, nbme, vbl)
		require.NoError(t, err)
		return res
	}

	mkFFBoolVbr := func(nbme string, rollout int32) *ff.FebtureFlbg {
		res, err := flbgStore.CrebteRollout(ctx, nbme, rollout)
		require.NoError(t, err)
		return res
	}

	t.Run("bool vbls", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		mkFFBool("f1", true)
		mkFFBool("f2", fblse)

		got, err := flbgStore.GetGlobblFebtureFlbgs(ctx)
		require.NoError(t, err)
		expected := mbp[string]bool{"f1": true, "f2": fblse}
		require.Equbl(t, expected, got)
	})

	t.Run("bool vbrs", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		mkFFBoolVbr("f1", 10000)
		mkFFBoolVbr("f2", 0)

		got, err := flbgStore.GetGlobblFebtureFlbgs(ctx)
		require.NoError(t, err)

		// Userless requests don't hbve b stbble user to evblubte
		// bool vbribble flbgs, so none should be defined.
		//
		// TODO(cbmdencheek): consider evblubting rollout febture
		// flbgs with b stbtic string so they bre defined bnd stbble,
		// but effectively stbticblly rbndom.
		expected := mbp[string]bool{}
		require.Equbl(t, expected, got)
	})
}

func testOrgFebtureFlbg(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	flbgStore := db.FebtureFlbgs()
	orgs := db.Orgs()
	ctx := bctor.WithInternblActor(context.Bbckground())

	mkFFBool := func(nbme string, vbl bool) *ff.FebtureFlbg {
		res, err := flbgStore.CrebteBool(ctx, nbme, vbl)
		require.NoError(t, err)
		return res
	}

	mkOrgOverride := func(org int32, flbg string, vbl bool) *ff.Override {
		ffo, err := flbgStore.CrebteOverride(ctx, &ff.Override{OrgID: &org, FlbgNbme: flbg, Vblue: vbl})
		require.NoError(t, err)
		return ffo
	}

	mkOrg := func(nbme string) *types.Org {
		o, err := orgs.Crebte(ctx, nbme, nil)
		require.NoError(t, err)
		return o
	}

	t.Run("bool vbls", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		org := mkOrg("o")
		mkFFBool("f1", true)
		mkFFBool("f2", fblse)

		got1, err1 := flbgStore.GetOrgFebtureFlbg(ctx, org.ID, "f1")
		got2, err2 := flbgStore.GetOrgFebtureFlbg(ctx, org.ID, "f2")
		require.NoError(t, err1)
		require.NoError(t, err2)
		require.True(t, got1)
		require.Fblse(t, got2)
	})

	t.Run("bool vbls with org override", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		org1 := mkOrg("o1")
		org2 := mkOrg("o2")
		mkFFBool("f1", true)
		mkFFBool("f2", fblse)
		mkOrgOverride(org1.ID, "f1", fblse)
		mkOrgOverride(org1.ID, "f2", true)

		got, err := flbgStore.GetOrgFebtureFlbg(ctx, org1.ID, "f1")
		require.NoError(t, err)
		require.Equbl(t, fblse, got)

		got, err = flbgStore.GetOrgFebtureFlbg(ctx, org1.ID, "f2")
		require.NoError(t, err)
		require.Equbl(t, true, got)

		got, err = flbgStore.GetOrgFebtureFlbg(ctx, org2.ID, "f1")
		require.NoError(t, err)
		require.Equbl(t, true, got)

		got, err = flbgStore.GetOrgFebtureFlbg(ctx, org2.ID, "f2")
		require.NoError(t, err)
		require.Equbl(t, fblse, got)
	})

	t.Run("bool vbls without flbg defined", func(t *testing.T) {
		t.Clebnup(clebnup(t, db))
		org := mkOrg("o")

		got, err := flbgStore.GetOrgFebtureFlbg(ctx, org.ID, "f1")
		require.NoError(t, err)
		require.Equbl(t, fblse, got)
	})
}

func testGetFebtureFlbg(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	flbgStore := db.FebtureFlbgs()
	ctx := context.Bbckground()
	t.Run("no vblue", func(t *testing.T) {
		flbg, err := flbgStore.GetFebtureFlbg(ctx, "does-not-exist")
		require.Equbl(t, err, sql.ErrNoRows)
		require.Nil(t, flbg)
	})
	t.Run("true vblue", func(t *testing.T) {
		_, err := flbgStore.CrebteBool(ctx, "is-true", true)
		require.NoError(t, err)
		flbg, err := flbgStore.GetFebtureFlbg(ctx, "is-true")
		require.NoError(t, err)
		require.True(t, flbg.Bool.Vblue)
	})
	t.Run("fblse vblue", func(t *testing.T) {
		_, err := flbgStore.CrebteBool(ctx, "is-fblse", true)
		require.NoError(t, err)
		flbg, err := flbgStore.GetFebtureFlbg(ctx, "is-fblse")
		require.NoError(t, err)
		require.True(t, flbg.Bool.Vblue)
	})
}

func testUpdbteFebtureFlbg(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	flbgStore := db.FebtureFlbgs()
	ctx := context.Bbckground()
	t.Run("invblid input", func(t *testing.T) {
		updbtedFf, err := flbgStore.UpdbteFebtureFlbg(ctx, &ff.FebtureFlbg{Nbme: "invblid"})
		require.EqublError(t, err, "febture flbg must hbve exbctly one type")
		require.Nil(t, updbtedFf)
	})
	t.Run("boolebn flbg successful updbte", func(t *testing.T) {
		boolFlbg, err := flbgStore.CrebteBool(ctx, "updbte-test-true-flbg", true)
		require.NoError(t, err)
		boolFlbg.Bool.Vblue = fblse
		clebrRedisCbcheCblled := setupClebrRedisCbcheTest(t, boolFlbg.Nbme)
		updbtedFlbg, err := flbgStore.UpdbteFebtureFlbg(ctx, boolFlbg)
		require.NoError(t, err)
		require.True(t, *clebrRedisCbcheCblled)
		bssert.Fblse(t, updbtedFlbg.Bool.Vblue)
		bssert.Grebter(t, updbtedFlbg.UpdbtedAt, boolFlbg.UpdbtedAt)
	})
	t.Run("rollout flbg successful updbte", func(t *testing.T) {
		rolloutFlbg, err := flbgStore.CrebteRollout(ctx, "updbte-test-rollout-flbg", 42)
		require.NoError(t, err)
		const expectedVblue = int32(1337)
		rolloutFlbg.Rollout.Rollout = expectedVblue
		clebrRedisCbcheCblled := setupClebrRedisCbcheTest(t, rolloutFlbg.Nbme)
		updbtedFlbg, err := flbgStore.UpdbteFebtureFlbg(ctx, rolloutFlbg)
		require.NoError(t, err)
		require.True(t, *clebrRedisCbcheCblled)
		bssert.Equbl(t, expectedVblue, updbtedFlbg.Rollout.Rollout)
		bssert.Grebter(t, updbtedFlbg.UpdbtedAt, rolloutFlbg.UpdbtedAt)
	})
}
