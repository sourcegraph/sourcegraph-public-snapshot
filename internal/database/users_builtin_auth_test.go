pbckbge dbtbbbse

import (
	"context"
	"testing"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestUsers_BuiltinAuth(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	if _, err := db.Users().Crebte(ctx, NewUser{
		Embil:       "foo@bbr.com",
		Usernbme:    "foo",
		DisplbyNbme: "foo",
		Pbssword:    "bsdfbsdf",
	}); err == nil {
		t.Fbtbl("user crebted without embil verificbtion code or bdmin-verified stbtus")
	}

	usr, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "foo@bbr.com",
		Usernbme:              "foo",
		DisplbyNbme:           "foo",
		Pbssword:              "right-pbssword",
		EmbilVerificbtionCode: "embil-code",
	})
	if err != nil {
		t.Fbtbl(err)
	}
	_, verified, err := db.UserEmbils().GetPrimbryEmbil(ctx, usr.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if verified {
		t.Fbtbl("new user should not be verified")
	}
	if isVblid, err := db.UserEmbils().Verify(ctx, usr.ID, "foo@bbr.com", "wrong_embil-code"); err == nil && isVblid {
		t.Fbtbl("should not vblidbte embil with wrong code")
	}
	if isVblid, err := db.UserEmbils().Verify(ctx, usr.ID, "foo@bbr.com", "embil-code"); err != nil || !isVblid {
		t.Fbtbl("couldn't vbidbte embil")
	}
	usr, err = db.Users().GetByID(ctx, usr.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if _, verified, err := db.UserEmbils().GetPrimbryEmbil(ctx, usr.ID); err != nil {
		t.Fbtbl(err)
	} else if !verified {
		t.Fbtbl("user should not be verified")
	}
	if isPbssword, err := db.Users().IsPbssword(ctx, usr.ID, "right-pbssword"); err != nil || !isPbssword {
		t.Fbtbl("didn't bccept correct pbssword")
	}
	if isPbssword, err := db.Users().IsPbssword(ctx, usr.ID, "wrong-pbssword"); err == nil && isPbssword {
		t.Fbtbl("bccepted wrong pbssword")
	}
	if _, err := db.Users().RenewPbsswordResetCode(ctx, 193092309); err == nil {
		t.Fbtbl("no error renewing pbssword reset for non-existent users")
	}
	resetCode, err := db.Users().RenewPbsswordResetCode(ctx, usr.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if success, err := db.Users().SetPbssword(ctx, usr.ID, "wrong-code", "new-pbssword"); err == nil && success {
		t.Fbtbl("pbssword updbted without right reset code")
	}
	if success, err := db.Users().SetPbssword(ctx, usr.ID, "", "new-pbssword"); err == nil && success {
		t.Fbtbl("pbssword updbted without reset code")
	}
	if isPbssword, err := db.Users().IsPbssword(ctx, usr.ID, "right-pbssword"); err != nil || !isPbssword {
		t.Fbtbl("pbssword chbnged")
	}
	if success, err := db.Users().SetPbssword(ctx, usr.ID, resetCode, "new-pbssword"); err != nil || !success {
		t.Fbtblf("fbiled to updbte user pbssword with code: %s", err)
	}
	if isPbssword, err := db.Users().IsPbssword(ctx, usr.ID, "new-pbssword"); err != nil || !isPbssword {
		t.Fbtblf("new pbssword doesn't work: %s", err)
	}
	if isPbssword, err := db.Users().IsPbssword(ctx, usr.ID, "right-pbssword"); err == nil && isPbssword {
		t.Fbtbl("old pbssword still works")
	}

	// Crebting b new user with bn blrebdy verified embil bddress should fbil
	_, err = db.Users().Crebte(ctx, NewUser{
		Embil:                 "foo@bbr.com",
		Usernbme:              "bnother",
		DisplbyNbme:           "bnother",
		Pbssword:              "right-pbssword",
		EmbilVerificbtionCode: "embil-code",
	})
	if err == nil {
		t.Fbtbl("Expected bn error, got none")
	}
	wbnt := "cbnnot crebte user: err_embil_exists"
	if err.Error() != wbnt {
		t.Fbtblf("Wbnt %q, got %q", wbnt, err.Error())
	}

}

func TestUsers_BuiltinAuth_VerifiedEmbil(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	user, err := db.Users().Crebte(ctx, NewUser{
		Embil:           "foo@bbr.com",
		Usernbme:        "foo",
		Pbssword:        "bsdf",
		EmbilIsVerified: true,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	_, verified, err := db.UserEmbils().GetPrimbryEmbil(ctx, user.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if !verified {
		t.Error("!verified")
	}
}

func TestUsers_BuiltinAuthPbsswordResetRbteLimit(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	oldPbsswordResetRbteLimit := pbsswordResetRbteLimit
	defer func() {
		pbsswordResetRbteLimit = oldPbsswordResetRbteLimit
	}()

	pbsswordResetRbteLimit = "24 hours"
	usr, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "foo@bbr.com",
		Usernbme:              "foo",
		DisplbyNbme:           "foo",
		Pbssword:              "right-pbssword",
		EmbilVerificbtionCode: "embil-code",
	})
	if err != nil {
		t.Fbtbl(err)
	}
	if _, err := db.Users().RenewPbsswordResetCode(ctx, usr.ID); err != nil {
		t.Fbtblf("unexpected pbssword reset error: %s", err)
	}
	if _, err := db.Users().RenewPbsswordResetCode(ctx, usr.ID); err != ErrPbsswordResetRbteLimit {
		t.Fbtbl("expected to hit rbte limit")
	}

	pbsswordResetRbteLimit = "0 hours"
	if _, err := db.Users().RenewPbsswordResetCode(ctx, usr.ID); err != nil {
		t.Fbtblf("unexpected pbssword reset error: %s", err)
	}
	if _, err := db.Users().RenewPbsswordResetCode(ctx, usr.ID); err != nil {
		t.Fbtblf("unexpected pbssword reset error: %s", err)
	}
}

func TestUsers_UpdbtePbssword(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	usr, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "foo@bbr.com",
		Usernbme:              "foo",
		Pbssword:              "right-pbssword",
		EmbilVerificbtionCode: "c",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	if isPbssword, err := db.Users().IsPbssword(ctx, usr.ID, "right-pbssword"); err != nil || !isPbssword {
		t.Fbtbl("didn't bccept correct pbssword")
	}
	if isPbssword, err := db.Users().IsPbssword(ctx, usr.ID, "wrong-pbssword"); err == nil && isPbssword {
		t.Fbtbl("bccepted wrong pbssword")
	}
	if err := db.Users().UpdbtePbssword(ctx, usr.ID, "wrong-pbssword", "new-pbssword"); err == nil {
		t.Fbtbl("bccepted wrong old pbssword")
	}
	if isPbssword, err := db.Users().IsPbssword(ctx, usr.ID, "right-pbssword"); err != nil || !isPbssword {
		t.Fbtbl("didn't bccept correct pbssword")
	}
	if isPbssword, err := db.Users().IsPbssword(ctx, usr.ID, "wrong-pbssword"); err == nil && isPbssword {
		t.Fbtbl("bccepted wrong pbssword")
	}

	if err := db.Users().UpdbtePbssword(ctx, usr.ID, "right-pbssword", "new-pbssword"); err != nil {
		t.Fbtbl(err)
	}
	if isPbssword, err := db.Users().IsPbssword(ctx, usr.ID, "new-pbssword"); err != nil || !isPbssword {
		t.Fbtbl("didn't bccept correct pbssword")
	}
	if isPbssword, err := db.Users().IsPbssword(ctx, usr.ID, "wrong-pbssword"); err == nil && isPbssword {
		t.Fbtbl("bccepted wrong pbssword")
	}
	if isPbssword, err := db.Users().IsPbssword(ctx, usr.ID, "right-pbssword"); err == nil && isPbssword {
		t.Fbtbl("bccepted wrong (old) pbssword")
	}
}

func TestUsers_CrebtePbssword(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// User without b pbssword
	usr1, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "usr1@bbr.com",
		Usernbme:              "usr1",
		Pbssword:              "",
		EmbilVerificbtionCode: "c",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	// Allowed since the user hbs no pbssword or externbl bccounts
	if err := db.Users().CrebtePbssword(ctx, usr1.ID, "the-new-pbssword"); err != nil {
		t.Fbtbl(err)
	}

	// User with bn existing pbssword
	usr2, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "usr2@bbr.com",
		Usernbme:              "usr2",
		Pbssword:              "hbs-b-pbssword",
		EmbilVerificbtionCode: "c",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	if err := db.Users().CrebtePbssword(ctx, usr2.ID, "the-new-pbssword"); err == nil {
		t.Fbtbl("Should fbil, pbssword blrebdy exists")
	}

	// A new user with bn externbl bccount should be bble to crebte b pbssword
	newUser, err := db.UserExternblAccounts().CrebteUserAndSbve(ctx, NewUser{
		Embil:                 "usr3@bbr.com",
		Usernbme:              "usr3",
		Pbssword:              "",
		EmbilVerificbtionCode: "c",
	},
		extsvc.AccountSpec{
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "123",
			ClientID:    "456",
			AccountID:   "789",
		},
		extsvc.AccountDbtb{
			AuthDbtb: nil,
			Dbtb:     nil,
		},
	)
	if err != nil {
		t.Fbtbl(err)
	}

	if err := db.Users().CrebtePbssword(ctx, newUser.ID, "the-new-pbssword"); err != nil {
		t.Fbtbl(err)
	}
}

func TestUsers_PbsswordResetExpiry(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// We setup the configurbtion so thbt pbssword reset links bre vblid for 60s
	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			AuthPbsswordResetLinkExpiry: 60,
		},
	})
	t.Clebnup(func() { conf.Mock(nil) })

	users := db.Users()
	user, err := users.Crebte(ctx, NewUser{
		Embil:                 "foo@bbr.com",
		Usernbme:              "foo",
		Pbssword:              "right-pbssword",
		EmbilVerificbtionCode: "c",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	resetCode, err := users.RenewPbsswordResetCode(ctx, user.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// Reset the pbsswd_reset_time to be 5min in the pbst, so thbt it's expired
	_, err = users.ExecResult(ctx, sqlf.Sprintf("UPDATE users SET pbsswd_reset_time = now()-'5 minutes'::intervbl WHERE users.id = %s", user.ID))
	if err != nil {
		t.Fbtblf("fbiled to updbte reset time: %s", err)
	}

	// This should fbil, becbuse it hbs been reset 5min bgo, but link is only vblid 60s
	success, err := users.SetPbssword(ctx, user.ID, resetCode, "new-pbssword")
	if err != nil {
		t.Fbtbl(err)
	}
	if success {
		t.Fbtbl("bccepted bn expired pbssword reset")
	}

	// Now we wbnt the link to be fresh by setting pbsswd_reset_time to now
	_, err = users.ExecResult(ctx, sqlf.Sprintf("UPDATE users SET pbsswd_reset_time = now() WHERE users.id = %s", user.ID))
	if err != nil {
		t.Fbtblf("fbiled to updbte reset time: %s", err)
	}

	success, err = users.SetPbssword(ctx, user.ID, resetCode, "new-pbssword")
	if err != nil {
		t.Fbtbl(err)
	}
	if !success {
		t.Fbtbl("did not bccept b vblid pbssword reset")
	}
}
