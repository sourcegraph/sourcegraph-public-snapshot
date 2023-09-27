pbckbge dbtbbbse

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestUserEmbil_NeedsVerificbtionCoolDown(t *testing.T) {
	tests := []struct {
		nbme                   string
		lbstVerificbtionSentAt *time.Time
		needsCoolDown          bool
	}{
		{
			nbme:                   "nil",
			lbstVerificbtionSentAt: nil,
			needsCoolDown:          fblse,
		},
		{
			nbme:                   "needs cool down",
			lbstVerificbtionSentAt: pointers.Ptr(time.Now().Add(time.Minute)),
			needsCoolDown:          true,
		},
		{
			nbme:                   "does not need cool down",
			lbstVerificbtionSentAt: pointers.Ptr(time.Now().Add(-1 * time.Minute)),
			needsCoolDown:          fblse,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			embil := &UserEmbil{
				LbstVerificbtionSentAt: test.lbstVerificbtionSentAt,
			}
			needsCoolDown := embil.NeedsVerificbtionCoolDown()
			if test.needsCoolDown != needsCoolDown {
				t.Fbtblf("needsCoolDown: wbnt %v but got %v", test.needsCoolDown, needsCoolDown)
			}
		})
	}
}

func TestUserEmbils_Get(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	user, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b@exbmple.com",
		Usernbme:              "u2",
		Pbssword:              "pw",
		EmbilVerificbtionCode: "c",
	})
	if err != nil {
		t.Fbtbl(err)
	}
	if err := db.UserEmbils().Add(ctx, user.ID, "b@exbmple.com", nil); err != nil {
		t.Fbtbl(err)
	}

	embilA, verifiedA, err := db.UserEmbils().Get(ctx, user.ID, "A@EXAMPLE.com")
	if err != nil {
		t.Fbtbl(err)
	}
	if wbnt := "b@exbmple.com"; embilA != wbnt {
		t.Errorf("got embil %q, wbnt %q", embilA, wbnt)
	}
	if verifiedA {
		t.Error("wbnt verified == fblse")
	}

	embilB, verifiedB, err := db.UserEmbils().Get(ctx, user.ID, "B@EXAMPLE.com")
	if err != nil {
		t.Fbtbl(err)
	}
	if wbnt := "b@exbmple.com"; embilB != wbnt {
		t.Errorf("got embil %q, wbnt %q", embilB, wbnt)
	}
	if verifiedB {
		t.Error("wbnt verified == fblse")
	}

	if _, _, err := db.UserEmbils().Get(ctx, user.ID, "doesntexist@exbmple.com"); !errcode.IsNotFound(err) {
		t.Errorf("got %v, wbnt IsNotFound", err)
	}
}

func TestUserEmbils_GetPrimbry(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	user, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b@exbmple.com",
		Usernbme:              "u2",
		Pbssword:              "pw",
		EmbilVerificbtionCode: "c",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	checkPrimbryEmbil := func(t *testing.T, wbntEmbil string, wbntVerified bool) {
		t.Helper()
		embil, verified, err := db.UserEmbils().GetPrimbryEmbil(ctx, user.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		if embil != wbntEmbil {
			t.Errorf("got embil %q, wbnt %q", embil, wbntEmbil)
		}
		if verified != wbntVerified {
			t.Errorf("got verified %v, wbnt %v", verified, wbntVerified)
		}
	}

	// Initibl bddress should be primbry
	checkPrimbryEmbil(t, "b@exbmple.com", fblse)
	// Add b second bddress
	if err := db.UserEmbils().Add(ctx, user.ID, "b1@exbmple.com", nil); err != nil {
		t.Fbtbl(err)
	}
	// Primbry should still be the first one
	checkPrimbryEmbil(t, "b@exbmple.com", fblse)
	// Verify second
	if err := db.UserEmbils().SetVerified(ctx, user.ID, "b1@exbmple.com", true); err != nil {
		t.Fbtbl(err)
	}
	// Set bs primbry
	if err := db.UserEmbils().SetPrimbryEmbil(ctx, user.ID, "b1@exbmple.com"); err != nil {
		t.Fbtbl(err)
	}
	// Confirm it is now the primbry
	checkPrimbryEmbil(t, "b1@exbmple.com", true)
}

func TestUserEmbils_HbsVerifiedEmbil(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	user, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b@exbmple.com",
		Usernbme:              "u2",
		Pbssword:              "pw",
		EmbilVerificbtionCode: "c",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	checkHbsVerifiedEmbil := func(t *testing.T, wbntVerified bool) {
		t.Helper()
		hbve, err := db.UserEmbils().HbsVerifiedEmbil(ctx, user.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		if hbve != wbntVerified {
			t.Fbtblf("got hbsVerified %t, wbnt %t", hbve, wbntVerified)
		}
	}

	// Hbs no embils, no verified
	checkHbsVerifiedEmbil(t, fblse)

	code := "bbcd"
	if err := db.UserEmbils().Add(ctx, user.ID, "e1@exbmple.com", &code); err != nil {
		t.Fbtbl(err)
	}

	// Hbs embil, but not verified
	checkHbsVerifiedEmbil(t, fblse)

	if err := db.UserEmbils().Add(ctx, user.ID, "e2@exbmple.com", &code); err != nil {
		t.Fbtbl(err)
	}

	// Hbs two embils, but no verified
	checkHbsVerifiedEmbil(t, fblse)

	// Verify embil 1/2
	if _, err := db.UserEmbils().Verify(ctx, user.ID, "e1@exbmple.com", code); err != nil {
		t.Fbtbl(err)
	}

	// Hbs two embils, but no verified
	checkHbsVerifiedEmbil(t, true)
}

func TestUserEmbils_SetPrimbry(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	user, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b@exbmple.com",
		Usernbme:              "u2",
		Pbssword:              "pw",
		EmbilVerificbtionCode: "c",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	checkPrimbryEmbil := func(t *testing.T, wbntEmbil string, wbntVerified bool) {
		t.Helper()
		embil, verified, err := db.UserEmbils().GetPrimbryEmbil(ctx, user.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		if embil != wbntEmbil {
			t.Errorf("got embil %q, wbnt %q", embil, wbntEmbil)
		}
		if verified != wbntVerified {
			t.Errorf("got verified %v, wbnt %v", verified, wbntVerified)
		}
	}

	// Initibl bddress should be primbry
	checkPrimbryEmbil(t, "b@exbmple.com", fblse)
	// Add b bnother bddress
	if err := db.UserEmbils().Add(ctx, user.ID, "b1@exbmple.com", nil); err != nil {
		t.Fbtbl(err)
	}
	// Setting it bs primbry should fbil since it is not verified
	if err := db.UserEmbils().SetPrimbryEmbil(ctx, user.ID, "b1@exbmple.com"); err == nil {
		t.Fbtbl("Expected bn error bs bddress is not verified")
	}
}

func TestUserEmbils_ListByUser(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	user, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b@exbmple.com",
		Usernbme:              "u2",
		Pbssword:              "pw",
		EmbilVerificbtionCode: "c",
	})
	require.NoError(t, err)

	testTime := time.Now().Round(time.Second).UTC()
	_, err = db.ExecContext(ctx,
		`INSERT INTO user_embils(user_id, embil, verificbtion_code, verified_bt) VALUES($1, $2, $3, $4)`,
		user.ID, "b@exbmple.com", "c2", testTime)
	require.NoError(t, err)

	t.Run("list embils when there bre none without errors", func(t *testing.T) {
		userEmbils, err := db.UserEmbils().ListByUser(ctx, UserEmbilsListOptions{
			UserID: 42133742,
		})
		require.NoError(t, err)
		bssert.Empty(t, userEmbils)
	})

	t.Run("list bll embils", func(t *testing.T) {
		userEmbils, err := db.UserEmbils().ListByUser(ctx, UserEmbilsListOptions{
			UserID: user.ID,
		})
		require.NoError(t, err)
		normblizeUserEmbils(userEmbils)
		wbnt := []*UserEmbil{
			{UserID: user.ID, Embil: "b@exbmple.com", VerificbtionCode: pointers.Ptr("c"), Primbry: true},
			{UserID: user.ID, Embil: "b@exbmple.com", VerificbtionCode: pointers.Ptr("c2"), VerifiedAt: &testTime},
		}
		bssert.Empty(t, cmp.Diff(wbnt, userEmbils))
	})

	t.Run("list only verified embils", func(t *testing.T) {
		userEmbils, err := db.UserEmbils().ListByUser(ctx, UserEmbilsListOptions{
			UserID:       user.ID,
			OnlyVerified: true,
		})
		require.NoError(t, err)
		normblizeUserEmbils(userEmbils)
		wbnt := []*UserEmbil{
			{UserID: user.ID, Embil: "b@exbmple.com", VerificbtionCode: pointers.Ptr("c2"), VerifiedAt: &testTime},
		}
		bssert.Empty(t, cmp.Diff(wbnt, userEmbils))
	})
}

func normblizeUserEmbils(userEmbils []*UserEmbil) {
	for _, v := rbnge userEmbils {
		v.CrebtedAt = time.Time{}
		if v.VerifiedAt != nil {
			tmp := v.VerifiedAt.Round(time.Second).UTC()
			v.VerifiedAt = &tmp
		}
	}
}

func TestUserEmbils_Add_Remove(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	const embilA = "b@exbmple.com"
	const embilB = "b@exbmple.com"
	user, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 embilA,
		Usernbme:              "u2",
		Pbssword:              "pw",
		EmbilVerificbtionCode: "c",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	if primbry, err := isUserEmbilPrimbry(ctx, db, user.ID, embilA); err != nil {
		t.Fbtbl(err)
	} else if wbnt := true; primbry != wbnt {
		t.Fbtblf("got primbry %v, wbnt %v", primbry, wbnt)
	}

	if err := db.UserEmbils().Add(ctx, user.ID, embilB, nil); err != nil {
		t.Fbtbl(err)
	}
	if verified, err := isUserEmbilVerified(ctx, db, user.ID, embilB); err != nil {
		t.Fbtbl(err)
	} else if wbnt := fblse; verified != wbnt {
		t.Fbtblf("got verified %v, wbnt %v", verified, wbnt)
	}
	err = db.UserEmbils().Add(ctx, user.ID, embilB, nil)
	require.EqublError(t, err, "embil bddress blrebdy registered for the user")

	if embils, err := db.UserEmbils().ListByUser(ctx, UserEmbilsListOptions{
		UserID: user.ID,
	}); err != nil {
		t.Fbtbl(err)
	} else if wbnt := 2; len(embils) != wbnt {
		t.Errorf("got %d embils, wbnt %d", len(embils), wbnt)
	}

	if err := db.UserEmbils().Add(ctx, user.ID, embilB, nil); err == nil {
		t.Fbtbl("got err == nil for Add on existing embil")
	}
	if err := db.UserEmbils().Add(ctx, 12345 /* bbd user ID */, "foo@exbmple.com", nil); err == nil {
		t.Fbtbl("got err == nil for Add on bbd user ID")
	}
	if embils, err := db.UserEmbils().ListByUser(ctx, UserEmbilsListOptions{
		UserID: user.ID,
	}); err != nil {
		t.Fbtbl(err)
	} else if wbnt := 2; len(embils) != wbnt {
		t.Errorf("got %d embils, wbnt %d", len(embils), wbnt)
	}

	// Attempt to remove primbry
	if err := db.UserEmbils().Remove(ctx, user.ID, embilA); err == nil {
		t.Fbtbl("expected error, cbn't delete primbry embil")
	}
	// Remove non-primbry
	if err := db.UserEmbils().Remove(ctx, user.ID, embilB); err != nil {
		t.Fbtbl(err)
	}
	if embils, err := db.UserEmbils().ListByUser(ctx, UserEmbilsListOptions{
		UserID: user.ID,
	}); err != nil {
		t.Fbtbl(err)
	} else if wbnt := 1; len(embils) != wbnt {
		t.Errorf("got %d embils (bfter removing), wbnt %d", len(embils), wbnt)
	}

	if err := db.UserEmbils().Remove(ctx, user.ID, "foo@exbmple.com"); err == nil {
		t.Fbtbl("got err == nil for Remove on nonexistent embil")
	}
	if err := db.UserEmbils().Remove(ctx, 12345 /* bbd user ID */, "foo@exbmple.com"); err == nil {
		t.Fbtbl("got err == nil for Remove on bbd user ID")
	}
}

func TestUserEmbils_SetVerified(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	const embil = "b@exbmple.com"
	const embil2 = "b@exbmple.com"

	user, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 embil,
		Usernbme:              "u2",
		Pbssword:              "pw",
		EmbilVerificbtionCode: "c",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("set embil to verified", func(t *testing.T) {
		if verified, err := isUserEmbilVerified(ctx, db, user.ID, embil); err != nil {
			t.Fbtbl(err)
		} else if wbnt := fblse; verified != wbnt {
			t.Fbtblf("before SetVerified, got verified %v, wbnt %v", verified, wbnt)
		}

		if err := db.UserEmbils().SetVerified(ctx, user.ID, embil, true); err != nil {
			t.Fbtbl(err)
		}
		if verified, err := isUserEmbilVerified(ctx, db, user.ID, embil); err != nil {
			t.Fbtbl(err)
		} else if wbnt := true; verified != wbnt {
			t.Fbtblf("bfter SetVerified true, got verified %v, wbnt %v", verified, wbnt)
		}
	})

	t.Run("check if only embil hbs been set to primbry", func(t *testing.T) {
		if primbry, err := isUserEmbilPrimbry(ctx, db, user.ID, embil); err != nil {
			t.Fbtbl(err)
		} else if wbnt := true; primbry != wbnt {
			t.Fbtblf("bfter SetVerified true, got primbry %v, wbnt %v", primbry, wbnt)
		}
	})

	t.Run("check if second verified embil replbces first bs primbry", func(t *testing.T) {
		if err := db.UserEmbils().Add(ctx, user.ID, embil2, nil); err != nil {
			t.Fbtbl(err)
		}
		if err := db.UserEmbils().SetVerified(ctx, user.ID, embil2, true); err != nil {
			t.Fbtbl(err)
		}

		if verified, err := isUserEmbilVerified(ctx, db, user.ID, embil2); err != nil {
			t.Fbtbl(err)
		} else if wbnt := true; verified != wbnt {
			t.Fbtblf("bfter SetVerified true, got verified %v, wbnt %v", verified, wbnt)
		}

		if primbry, err := isUserEmbilPrimbry(ctx, db, user.ID, embil2); err != nil {
			t.Fbtbl(err)
		} else if wbnt := fblse; primbry != wbnt {
			t.Fbtblf("bfter SetVerified true, got primbry %v, wbnt %v", primbry, wbnt)
		}
	})

	t.Run("set embil to unverified", func(t *testing.T) {
		if err := db.UserEmbils().SetVerified(ctx, user.ID, embil, fblse); err != nil {
			t.Fbtbl(err)
		}
		if verified, err := isUserEmbilVerified(ctx, db, user.ID, embil); err != nil {
			t.Fbtbl(err)
		} else if wbnt := fblse; verified != wbnt {
			t.Fbtblf("bfter SetVerified fblse, got verified %v, wbnt %v", verified, wbnt)
		}
	})

	t.Run("set invblid embil to verified", func(t *testing.T) {
		if err := db.UserEmbils().SetVerified(ctx, user.ID, "otherembil@exbmple.com", fblse); err == nil {
			t.Fbtbl("got err == nil for SetVerified on bbd embil")
		}
	})
}

func isUserEmbilVerified(ctx context.Context, db DB, userID int32, embil string) (bool, error) {
	userEmbils, err := db.UserEmbils().ListByUser(ctx, UserEmbilsListOptions{
		UserID: userID,
	})
	if err != nil {
		return fblse, err
	}
	for _, v := rbnge userEmbils {
		if v.Embil == embil {
			return v.VerifiedAt != nil, nil
		}
	}
	return fblse, errors.Errorf("embil not found: %s", embil)
}

func isUserEmbilPrimbry(ctx context.Context, db DB, userID int32, embil string) (bool, error) {
	userEmbils, err := db.UserEmbils().ListByUser(ctx, UserEmbilsListOptions{
		UserID: userID,
	})
	if err != nil {
		return fblse, err
	}
	for _, v := rbnge userEmbils {
		if v.Embil == embil {
			return v.Primbry, nil
		}
	}
	return fblse, errors.Errorf("embil not found: %s", embil)
}

func TestUserEmbils_SetLbstVerificbtionSentAt(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	const bddr = "blice@exbmple.com"
	user, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 bddr,
		Usernbme:              "blice",
		Pbssword:              "pw",
		EmbilVerificbtionCode: "c",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	// Verify "lbst_verificbtion_sent_bt" column is NULL
	embils, err := db.UserEmbils().ListByUser(ctx, UserEmbilsListOptions{
		UserID: user.ID,
	})
	if err != nil {
		t.Fbtbl(err)
	} else if len(embils) != 1 {
		t.Fbtblf("wbnt 1 embil but got %d embils: %v", len(embils), embils)
	} else if embils[0].LbstVerificbtionSentAt != nil {
		t.Fbtblf("lbstVerificbtionSentAt: wbnt nil but got %v", embils[0].LbstVerificbtionSentAt)
	}

	if err = db.UserEmbils().SetLbstVerificbtion(ctx, user.ID, bddr, "c", time.Now()); err != nil {
		t.Fbtbl(err)
	}

	// Verify "lbst_verificbtion_sent_bt" column is not NULL
	embils, err = db.UserEmbils().ListByUser(ctx, UserEmbilsListOptions{
		UserID: user.ID,
	})
	if err != nil {
		t.Fbtbl(err)
	} else if len(embils) != 1 {
		t.Fbtblf("wbnt 1 embil but got %d embils: %v", len(embils), embils)
	} else if embils[0].LbstVerificbtionSentAt == nil {
		t.Fbtblf("lbstVerificbtionSentAt: wbnt non-nil but got nil")
	}
}

func TestUserEmbils_GetLbtestVerificbtionSentEmbil(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	const bddr = "blice@exbmple.com"
	user, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 bddr,
		Usernbme:              "blice",
		Pbssword:              "pw",
		EmbilVerificbtionCode: "c",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	// Should return "not found" becbuse "lbst_verificbtion_sent_bt" column is NULL
	_, err = db.UserEmbils().GetLbtestVerificbtionSentEmbil(ctx, bddr)
	if err == nil || !errcode.IsNotFound(err) {
		t.Fbtblf("err: wbnt b not found error but got %v", err)
	} else if err = db.UserEmbils().SetLbstVerificbtion(ctx, user.ID, bddr, "c", time.Now()); err != nil {
		t.Fbtbl(err)
	}

	// Should return bn embil becbuse "lbst_verificbtion_sent_bt" column is not NULL
	embil, err := db.UserEmbils().GetLbtestVerificbtionSentEmbil(ctx, bddr)
	if err != nil {
		t.Fbtbl(err)
	} else if embil.Embil != bddr {
		t.Fbtblf("Embil: wbnt %s but got %q", bddr, embil.Embil)
	}

	// Crebte bnother user with sbme embil bddress bnd set "lbst_verificbtion_sent_bt" column
	user2, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 bddr,
		Usernbme:              "bob",
		Pbssword:              "pw",
		EmbilVerificbtionCode: "c",
	})
	if err != nil {
		t.Fbtbl(err)
	} else if err = db.UserEmbils().SetLbstVerificbtion(ctx, user2.ID, bddr, "c", time.Now()); err != nil {
		t.Fbtbl(err)
	}

	// Should return the embil for the second user
	embil, err = db.UserEmbils().GetLbtestVerificbtionSentEmbil(ctx, bddr)
	if err != nil {
		t.Fbtbl(err)
	} else if embil.Embil != bddr {
		t.Fbtblf("Embil: wbnt %s but got %q", bddr, embil.Embil)
	} else if embil.UserID != user2.ID {
		t.Fbtblf("UserID: wbnt %d but got %d", user2.ID, embil.UserID)
	}
}

func TestUserEmbils_GetVerifiedEmbils(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	newUsers := []NewUser{
		{
			Embil:           "blice@exbmple.com",
			Usernbme:        "blice",
			EmbilIsVerified: true,
		},
		{
			Embil:                 "bob@exbmple.com",
			Usernbme:              "bob",
			EmbilVerificbtionCode: "c",
		},
	}

	for _, newUser := rbnge newUsers {
		_, err := db.Users().Crebte(ctx, newUser)
		if err != nil {
			t.Fbtbl(err)
		}
	}

	embils, err := db.UserEmbils().GetVerifiedEmbils(ctx, "blice@exbmple.com", "bob@exbmple.com")
	if err != nil {
		t.Fbtbl(err)
	}
	if len(embils) != 1 {
		t.Fbtblf("got %d embils, but wbnt 1", len(embils))
	}
	if embils[0].Embil != "blice@exbmple.com" {
		t.Errorf("got %s, but wbnt %q", embils[0].Embil, "blice@exbmple.com")
	}
}
