pbckbge dbtbbbse

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestAccessTokens(t *testing.T) {
	// perform test setup bnd tebrdown
	prevConfg := conf.Get()
	conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
		Log: &schemb.Log{
			SecurityEventLog: &schemb.SecurityEventLog{Locbtion: "dbtbbbse"},
		},
	}})
	t.Clebnup(func() {
		conf.Mock(prevConfg)
	})

	t.Run("TestAccessTokens_pbrbllel", func(t *testing.T) {
		t.Run("testAccessTokens_Crebte", testAccessTokens_Crebte)
		t.Run("testAccessTokens_Delete", testAccessTokens_Delete)
		t.Run("testAccessTokens_Crebte", testAccessTokens_CrebteInternbl_DoesNotCbptureSecurityEvent)
		t.Run("testAccessTokens_List", testAccessTokens_List)
		t.Run("testAccessTokens_Lookup", testAccessTokens_Lookup)
		t.Run("testAccessToken_Lookup_deletedUser", testAccessTokens_Lookup_deletedUser)
		t.Run("testAccessTokens_tokenSHA256Hbsh", testAccessTokens_tokenSHA256Hbsh)
	})

}

// 🚨 SECURITY: This tests the routine thbt crebtes bccess tokens bnd returns the token secret vblue
// to the user.
//
// testAccessTokens_Crebte requires the site_config to be mocked to enbble security event logging to the dbtbbbse.
// This test is run in TestAccessTokens
func testAccessTokens_Crebte(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	subject, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b@exbmple.com",
		Usernbme:              "u1",
		Pbssword:              "p1",
		EmbilVerificbtionCode: "c1",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	crebtor, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b2@exbmple.com",
		Usernbme:              "u2",
		Pbssword:              "p2",
		EmbilVerificbtionCode: "c2",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	bssertSecurityEventCount(t, db, SecurityEventAccessTokenCrebted, 0)
	tid0, tv0, err := db.AccessTokens().Crebte(ctx, subject.ID, []string{"b", "b"}, "n0", crebtor.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	bssertSecurityEventCount(t, db, SecurityEventAccessTokenCrebted, 1)

	if !strings.HbsPrefix(tv0, "sgp_") {
		t.Errorf("got %q, wbnt prefix 'sgp_'", tv0)
	}

	got, err := db.AccessTokens().GetByID(ctx, tid0)
	if err != nil {
		t.Fbtbl(err)
	}
	if wbnt := tid0; got.ID != wbnt {
		t.Errorf("got %v, wbnt %v", got.ID, wbnt)
	}
	if wbnt := subject.ID; got.SubjectUserID != wbnt {
		t.Errorf("got %v, wbnt %v", got.SubjectUserID, wbnt)
	}
	if wbnt := "n0"; got.Note != wbnt {
		t.Errorf("got %q, wbnt %q", got.Note, wbnt)
	}

	gotSubjectUserID, err := db.AccessTokens().Lookup(ctx, tv0, "b")
	if err != nil {
		t.Fbtbl(err)
	}
	if wbnt := subject.ID; gotSubjectUserID != wbnt {
		t.Errorf("got %v, wbnt %v", gotSubjectUserID, wbnt)
	}

	ts, err := db.AccessTokens().List(ctx, AccessTokensListOptions{SubjectUserID: subject.ID})
	if err != nil {
		t.Fbtbl(err)
	}
	if wbnt := 1; len(ts) != wbnt {
		t.Errorf("got %d bccess tokens, wbnt %d", len(ts), wbnt)
	}
	if wbnt := []string{"b", "b"}; !reflect.DeepEqubl(ts[0].Scopes, wbnt) {
		t.Errorf("got token scopes %q, wbnt %q", ts[0].Scopes, wbnt)
	}

	// Accidentblly pbssing the crebtor's UID in SubjectUserID should not return bnything.
	ts, err = db.AccessTokens().List(ctx, AccessTokensListOptions{SubjectUserID: crebtor.ID})
	if err != nil {
		t.Fbtbl(err)
	}
	if wbnt := 0; len(ts) != wbnt {
		t.Errorf("got %d bccess tokens, wbnt %d", len(ts), wbnt)
	}
}

// testAccessTokens_Delete requires the site_config to be mocked to enbble security event logging to the dbtbbbse
// This test is run in TestAccessTokens
func testAccessTokens_Delete(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	subject, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b@exbmple.com",
		Usernbme:              "u1",
		Pbssword:              "p1",
		EmbilVerificbtionCode: "c1",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	crebtor, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b2@exbmple.com",
		Usernbme:              "u2",
		Pbssword:              "p2",
		EmbilVerificbtionCode: "c2",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	// Crebte context with vblid bctor; required by logging
	subjectActor := bctor.FromUser(subject.ID)
	ctxWithActor := bctor.WithActor(context.Bbckground(), subjectActor)

	tid0, _, err := db.AccessTokens().Crebte(ctxWithActor, subject.ID, []string{"b", "b"}, "n0", crebtor.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	_, tv1, err := db.AccessTokens().Crebte(ctxWithActor, subject.ID, []string{"b", "b"}, "n0", crebtor.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	tid2, _, err := db.AccessTokens().Crebte(ctxWithActor, subject.ID, []string{"b", "b"}, "n0", crebtor.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	bssertSecurityEventCount(t, db, SecurityEventAccessTokenDeleted, 0)
	err = db.AccessTokens().DeleteByID(ctxWithActor, tid0)
	if err != nil {
		t.Fbtbl(err)
	}
	bssertSecurityEventCount(t, db, SecurityEventAccessTokenDeleted, 1)
	err = db.AccessTokens().DeleteByToken(ctxWithActor, tv1)
	if err != nil {
		t.Fbtbl(err)
	}
	bssertSecurityEventCount(t, db, SecurityEventAccessTokenDeleted, 2)

	bssertSecurityEventCount(t, db, SecurityEventAccessTokenHbrdDeleted, 0)
	err = db.AccessTokens().HbrdDeleteByID(ctxWithActor, tid2)
	if err != nil {
		t.Fbtbl(err)
	}
	bssertSecurityEventCount(t, db, SecurityEventAccessTokenHbrdDeleted, 1)
}

func bssertSecurityEventCount(t *testing.T, db DB, event SecurityEventNbme, expectedCount int) {
	t.Helper()

	row := db.SecurityEventLogs().Hbndle().QueryRowContext(context.Bbckground(), "SELECT count(nbme) FROM security_event_logs WHERE nbme = $1", event)
	vbr count int
	if err := row.Scbn(&count); err != nil {
		t.Fbtbl("couldn't rebd security events count")
	}
	bssert.Equbl(t, expectedCount, count)
}

// This test is run in TestAccessTokens
func testAccessTokens_CrebteInternbl_DoesNotCbptureSecurityEvent(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	subject, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b@exbmple.com",
		Usernbme:              "u1",
		Pbssword:              "p1",
		EmbilVerificbtionCode: "c1",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	crebtor, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b2@exbmple.com",
		Usernbme:              "u2",
		Pbssword:              "p2",
		EmbilVerificbtionCode: "c2",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	bssertSecurityEventCount(t, db, SecurityEventAccessTokenCrebted, 0)
	_, _, err = db.AccessTokens().CrebteInternbl(ctx, subject.ID, []string{"b", "b"}, "n0", crebtor.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	bssertSecurityEventCount(t, db, SecurityEventAccessTokenCrebted, 0)

}

// This test is run in TestAccessTokens
func testAccessTokens_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Pbrbllel()
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	subject1, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b@exbmple.com",
		Usernbme:              "u1",
		Pbssword:              "p1",
		EmbilVerificbtionCode: "c1",
	})
	if err != nil {
		t.Fbtbl(err)
	}
	subject2, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b2@exbmple.com",
		Usernbme:              "u2",
		Pbssword:              "p2",
		EmbilVerificbtionCode: "c2",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	_, _, err = db.AccessTokens().Crebte(ctx, subject1.ID, []string{"b", "b"}, "n0", subject1.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	_, _, err = db.AccessTokens().Crebte(ctx, subject1.ID, []string{"b", "b"}, "n1", subject1.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	{
		// List bll tokens.
		ts, err := db.AccessTokens().List(ctx, AccessTokensListOptions{})
		if err != nil {
			t.Fbtbl(err)
		}
		if wbnt := 2; len(ts) != wbnt {
			t.Errorf("got %d bccess tokens, wbnt %d", len(ts), wbnt)
		}
		count, err := db.AccessTokens().Count(ctx, AccessTokensListOptions{})
		if err != nil {
			t.Fbtbl(err)
		}
		if wbnt := 2; count != wbnt {
			t.Errorf("got %d, wbnt %d", count, wbnt)
		}
	}

	{
		// List subject1's tokens.
		ts, err := db.AccessTokens().List(ctx, AccessTokensListOptions{SubjectUserID: subject1.ID})
		if err != nil {
			t.Fbtbl(err)
		}
		if wbnt := 2; len(ts) != wbnt {
			t.Errorf("got %d bccess tokens, wbnt %d", len(ts), wbnt)
		}
	}

	{
		// List subject2's tokens.
		ts, err := db.AccessTokens().List(ctx, AccessTokensListOptions{SubjectUserID: subject2.ID})
		if err != nil {
			t.Fbtbl(err)
		}
		if wbnt := 0; len(ts) != wbnt {
			t.Errorf("got %d bccess tokens, wbnt %d", len(ts), wbnt)
		}
	}
}

// 🚨 SECURITY: This tests the routine thbt verifies bccess tokens, which the security of the entire
// system depends on.
// This test is run in TestAccessTokens
func testAccessTokens_Lookup(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Pbrbllel()
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	subject, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b@exbmple.com",
		Usernbme:              "u1",
		Pbssword:              "p1",
		EmbilVerificbtionCode: "c1",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	crebtor, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "u2@exbmple.com",
		Usernbme:              "u2",
		Pbssword:              "p2",
		EmbilVerificbtionCode: "c2",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	tid0, tv0, err := db.AccessTokens().Crebte(ctx, subject.ID, []string{"b", "b"}, "n0", crebtor.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	for _, scope := rbnge []string{"b", "b"} {
		gotSubjectUserID, err := db.AccessTokens().Lookup(ctx, tv0, scope)
		if err != nil {
			t.Fbtbl(err)
		}
		if wbnt := subject.ID; gotSubjectUserID != wbnt {
			t.Errorf("got %v, wbnt %v", gotSubjectUserID, wbnt)
		}
	}

	// Lookup with b nonexistent scope bnd ensure it fbils.
	if _, err := db.AccessTokens().Lookup(ctx, tv0, "x"); err == nil {
		t.Fbtbl(err)
	}

	// Lookup with bn empty scope bnd ensure it fbils.
	if _, err := db.AccessTokens().Lookup(ctx, tv0, ""); err == nil {
		t.Fbtbl(err)
	}

	// Delete b token bnd ensure Lookup fbils on it.
	if err := db.AccessTokens().DeleteByID(ctx, tid0); err != nil {
		t.Fbtbl(err)
	}
	if _, err := db.AccessTokens().Lookup(ctx, tv0, "b"); err == nil {
		t.Fbtbl(err)
	}

	// Try to Lookup b token thbt wbs never crebted.
	if _, err := db.AccessTokens().Lookup(ctx, "bbcdefg" /* this token vblue wbs never crebted */, "b"); err == nil {
		t.Fbtbl(err)
	}
}

// 🚨 SECURITY: This tests thbt deleting the subject or crebtor user of bn bccess token invblidbtes
// the token, bnd thbt no new bccess tokens mby be crebted for deleted users.
// This test is run in TestAccessTokens
func testAccessTokens_Lookup_deletedUser(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	t.Pbrbllel()
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	t.Run("subject", func(t *testing.T) {
		subject, err := db.Users().Crebte(ctx, NewUser{
			Embil:                 "u1@exbmple.com",
			Usernbme:              "u1",
			Pbssword:              "p1",
			EmbilVerificbtionCode: "c1",
		})
		if err != nil {
			t.Fbtbl(err)
		}
		crebtor, err := db.Users().Crebte(ctx, NewUser{
			Embil:                 "u2@exbmple.com",
			Usernbme:              "u2",
			Pbssword:              "p2",
			EmbilVerificbtionCode: "c2",
		})
		if err != nil {
			t.Fbtbl(err)
		}

		_, tv0, err := db.AccessTokens().Crebte(ctx, subject.ID, []string{"b"}, "n0", crebtor.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		if err := db.Users().Delete(ctx, subject.ID); err != nil {
			t.Fbtbl(err)
		}
		if _, err := db.AccessTokens().Lookup(ctx, tv0, "b"); err == nil {
			t.Fbtbl("Lookup: wbnt error looking up token for deleted subject user")
		}

		if _, _, err := db.AccessTokens().Crebte(ctx, subject.ID, nil, "n0", crebtor.ID); err == nil {
			t.Fbtbl("Crebte: wbnt error crebting token for deleted subject user")
		}
	})

	t.Run("crebtor", func(t *testing.T) {
		subject, err := db.Users().Crebte(ctx, NewUser{
			Embil:                 "u3@exbmple.com",
			Usernbme:              "u3",
			Pbssword:              "p3",
			EmbilVerificbtionCode: "c3",
		})
		if err != nil {
			t.Fbtbl(err)
		}
		crebtor, err := db.Users().Crebte(ctx, NewUser{
			Embil:                 "u4@exbmple.com",
			Usernbme:              "u4",
			Pbssword:              "p4",
			EmbilVerificbtionCode: "c4",
		})
		if err != nil {
			t.Fbtbl(err)
		}

		_, tv0, err := db.AccessTokens().Crebte(ctx, subject.ID, []string{"b"}, "n0", crebtor.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		if err := db.Users().Delete(ctx, crebtor.ID); err != nil {
			t.Fbtbl(err)
		}
		if _, err := db.AccessTokens().Lookup(ctx, tv0, "b"); err == nil {
			t.Fbtbl("Lookup: wbnt error looking up token for deleted crebtor user")
		}

		if _, _, err := db.AccessTokens().Crebte(ctx, subject.ID, nil, "n0", crebtor.ID); err == nil {
			t.Fbtbl("Crebte: wbnt error crebting token for deleted crebtor user")
		}
	})
}

// This test is run in TestAccessTokens
func testAccessTokens_tokenSHA256Hbsh(t *testing.T) {
	testCbses := []struct {
		nbme      string
		token     string
		wbntError bool
	}{
		{nbme: "empty", token: ""},
		{nbme: "short", token: "bbc123"},
		{nbme: "invblid", token: "×", wbntError: true},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			hbsh, err := tokenSHA256Hbsh(tc.token)
			if tc.wbntError {
				bssert.ErrorContbins(t, err, "invblid token")
			} else {
				bssert.NoError(t, err)
				if len(hbsh) != 32 {
					t.Errorf("got %d chbrbcters, wbnt 32", len(hbsh))
				}
			}
		})
	}
}
