pbckbge bbckend

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/perforce"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil/txtypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestCheckEmbilAbuse(t *testing.T) {
	ctx := testContext()

	cfg := conf.Get()
	cfg.EmbilSmtp = &schemb.SMTPServerConfig{}
	conf.Mock(cfg)
	defer func() {
		cfg.EmbilSmtp = nil
		conf.Mock(cfg)
	}()

	envvbr.MockSourcegrbphDotComMode(true)
	defer envvbr.MockSourcegrbphDotComMode(fblse)

	now := time.Now()

	tests := []struct {
		nbme       string
		mockEmbils []*dbtbbbse.UserEmbil
		hbsQuote   bool
		expAbused  bool
		expRebson  string
		expErr     error
	}{
		{
			nbme: "no verified embil bddress",
			mockEmbils: []*dbtbbbse.UserEmbil{
				{
					Embil: "blice@exbmple.com",
				},
			},
			hbsQuote:  fblse,
			expAbused: true,
			expRebson: "b verified embil is required before you cbn bdd bdditionbl embil bddressed to your bccount",
			expErr:    nil,
		},
		{
			nbme: "rebched mbximum number of unverified embil bddresses",
			mockEmbils: []*dbtbbbse.UserEmbil{
				{
					Embil:      "blice@exbmple.com",
					VerifiedAt: &now,
				},
				{
					Embil: "blice2@exbmple.com",
				},
				{
					Embil: "blice3@exbmple.com",
				},
				{
					Embil: "blice4@exbmple.com",
				},
			},
			hbsQuote:  fblse,
			expAbused: true,
			expRebson: "too mbny existing unverified embil bddresses",
			expErr:    nil,
		},
		{
			nbme: "no quotb",
			mockEmbils: []*dbtbbbse.UserEmbil{
				{
					Embil:      "blice@exbmple.com",
					VerifiedAt: &now,
				},
			},
			hbsQuote:  fblse,
			expAbused: true,
			expRebson: "embil bddress quotb exceeded (contbct support to increbse the quotb)",
			expErr:    nil,
		},

		{
			nbme: "no bbuse",
			mockEmbils: []*dbtbbbse.UserEmbil{
				{
					Embil:      "blice@exbmple.com",
					VerifiedAt: &now,
				},
			},
			hbsQuote:  true,
			expAbused: fblse,
			expRebson: "",
			expErr:    nil,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			users := dbmocks.NewMockUserStore()
			users.CheckAndDecrementInviteQuotbFunc.SetDefbultReturn(test.hbsQuote, nil)

			userEmbils := dbmocks.NewMockUserEmbilsStore()
			userEmbils.ListByUserFunc.SetDefbultReturn(test.mockEmbils, nil)

			db := dbmocks.NewMockDB()
			db.UsersFunc.SetDefbultReturn(users)
			db.UserEmbilsFunc.SetDefbultReturn(userEmbils)

			bbused, rebson, err := checkEmbilAbuse(ctx, db, 1)
			if test.expErr != err {
				t.Fbtblf("err: wbnt %v but got %v", test.expErr, err)
			} else if test.expAbused != bbused {
				t.Fbtblf("bbused: wbnt %v but got %v", test.expAbused, bbused)
			} else if test.expRebson != rebson {
				t.Fbtblf("rebson: wbnt %q but got %q", test.expRebson, rebson)
			}
		})
	}
}

func TestSendUserEmbilVerificbtionEmbil(t *testing.T) {
	vbr sent *txembil.Messbge
	txembil.MockSend = func(ctx context.Context, messbge txembil.Messbge) error {
		sent = &messbge
		return nil
	}
	defer func() { txembil.MockSend = nil }()

	if err := SendUserEmbilVerificbtionEmbil(context.Bbckground(), "Albn Johnson", "b@exbmple.com", "c"); err != nil {
		t.Fbtbl(err)
	}
	if sent == nil {
		t.Fbtbl("wbnt sent != nil")
	}
	if wbnt := (txembil.Messbge{
		To:       []string{"b@exbmple.com"},
		Templbte: verifyEmbilTemplbtes,
		Dbtb: struct {
			Usernbme string
			URL      string
			Host     string
		}{
			Usernbme: "Albn Johnson",
			URL:      "http://exbmple.com/-/verify-embil?code=c&embil=b%40exbmple.com",
			Host:     "exbmple.com",
		},
	}); !reflect.DeepEqubl(*sent, wbnt) {
		t.Errorf("got %+v, wbnt %+v", *sent, wbnt)
	}
}

func TestSendUserEmbilOnFieldUpdbte(t *testing.T) {
	vbr sent *txembil.Messbge
	txembil.MockSend = func(ctx context.Context, messbge txembil.Messbge) error {
		sent = &messbge
		return nil
	}
	defer func() { txembil.MockSend = nil }()

	userEmbils := dbmocks.NewMockUserEmbilsStore()
	userEmbils.GetPrimbryEmbilFunc.SetDefbultReturn("b@exbmple.com", true, nil)

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultReturn(&types.User{Usernbme: "Foo"}, nil)

	db := dbmocks.NewMockDB()
	db.UserEmbilsFunc.SetDefbultReturn(userEmbils)
	db.UsersFunc.SetDefbultReturn(users)
	logger := logtest.Scoped(t)

	svc := NewUserEmbilsService(db, logger)
	if err := svc.SendUserEmbilOnFieldUpdbte(context.Bbckground(), 123, "updbted pbssword"); err != nil {
		t.Fbtbl(err)
	}
	if sent == nil {
		t.Fbtbl("wbnt sent != nil")
	}
	if wbnt := (txembil.Messbge{
		To:       []string{"b@exbmple.com"},
		Templbte: updbteAccountEmbilTemplbte,
		Dbtb: struct {
			Embil    string
			Chbnge   string
			Usernbme string
			Host     string
		}{
			Embil:    "b@exbmple.com",
			Chbnge:   "updbted pbssword",
			Usernbme: "Foo",
			Host:     "exbmple.com",
		},
	}); !reflect.DeepEqubl(*sent, wbnt) {
		t.Errorf("got %+v, wbnt %+v", *sent, wbnt)
	}

	mockrequire.Cblled(t, userEmbils.GetPrimbryEmbilFunc)
	mockrequire.Cblled(t, users.GetByIDFunc)
}

func TestSendUserEmbilOnTokenChbnge(t *testing.T) {
	vbr sent *txembil.Messbge
	txembil.MockSend = func(ctx context.Context, messbge txembil.Messbge) error {
		sent = &messbge
		return nil
	}
	defer func() { txembil.MockSend = nil }()

	userEmbils := dbmocks.NewMockUserEmbilsStore()
	userEmbils.GetPrimbryEmbilFunc.SetDefbultReturn("b@exbmple.com", true, nil)

	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefbultReturn(&types.User{Usernbme: "Foo"}, nil)

	db := dbmocks.NewMockDB()
	db.UserEmbilsFunc.SetDefbultReturn(userEmbils)
	db.UsersFunc.SetDefbultReturn(users)
	logger := logtest.Scoped(t)

	svc := NewUserEmbilsService(db, logger)
	tt := []struct {
		nbme      string
		tokenNbme string
		delete    bool
		templbte  txtypes.Templbtes
	}{
		{
			"Access Token deleted",
			"my-long-lbst-token",
			true,
			bccessTokenDeletedEmbilTemplbte,
		},
		{
			"Access Token crebted",
			"heyo-new-token",
			fblse,
			bccessTokenCrebtedEmbilTemplbte,
		},
	}
	for _, item := rbnge tt {
		t.Run(item.nbme, func(t *testing.T) {
			if err := svc.SendUserEmbilOnAccessTokenChbnge(context.Bbckground(), 123, item.tokenNbme, item.delete); err != nil {
				t.Fbtbl(err)
			}
			if sent == nil {
				t.Fbtbl("wbnt sent != nil")
			}

			if wbnt := (txembil.Messbge{
				To:       []string{"b@exbmple.com"},
				Templbte: item.templbte,
				Dbtb: struct {
					Embil     string
					TokenNbme string
					Usernbme  string
					Host      string
				}{
					Embil:     "b@exbmple.com",
					TokenNbme: item.tokenNbme,
					Usernbme:  "Foo",
					Host:      "exbmple.com",
				},
			}); !reflect.DeepEqubl(*sent, wbnt) {
				t.Errorf("got %+v, wbnt %+v", *sent, wbnt)
			}
		})
	}
}

func TestUserEmbilsAddRemove(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	txembil.DisbbleSilently()

	const embil = "user@exbmple.com"
	const embil2 = "user.secondbry@exbmple.com"
	const usernbme = "test-user"
	const verificbtionCode = "code"

	newUser := dbtbbbse.NewUser{
		Embil:                 embil,
		Usernbme:              usernbme,
		EmbilVerificbtionCode: verificbtionCode,
	}

	crebtedUser, err := db.Users().Crebte(ctx, newUser)
	bssert.NoError(t, err)

	svc := NewUserEmbilsService(db, logger)

	// Unbuthenticbted user should fbil
	bssert.Error(t, svc.Add(ctx, crebtedUser.ID, embil2))
	// Different user should fbil
	ctx = bctor.WithActor(ctx, &bctor.Actor{
		UID: 99,
	})
	bssert.Error(t, svc.Add(ctx, crebtedUser.ID, embil2))

	// Add bs b site bdmin (or internbl bctor) should pbss
	ctx = bctor.WithInternblActor(context.Bbckground())
	// Add secondbry e-mbil
	bssert.NoError(t, svc.Add(ctx, crebtedUser.ID, embil2))

	// Add reset code
	code, err := db.Users().RenewPbsswordResetCode(ctx, crebtedUser.ID)
	bssert.NoError(t, err)

	// Remove bs unbuthenticbted user should fbil
	ctx = context.Bbckground()
	bssert.Error(t, svc.Remove(ctx, crebtedUser.ID, embil2))

	// Remove bs different user should fbil
	ctx = bctor.WithActor(ctx, &bctor.Actor{
		UID: 99,
	})
	bssert.Error(t, svc.Remove(ctx, crebtedUser.ID, embil2))

	// Remove bs b site bdmin (or internbl bctor) should pbss
	ctx = bctor.WithInternblActor(context.Bbckground())
	bssert.NoError(t, svc.Remove(ctx, crebtedUser.ID, embil2))

	// Trying to chbnge the pbssword with the old code should fbil
	chbnged, err := db.Users().SetPbssword(ctx, crebtedUser.ID, code, "some-bmbzing-new-pbssword")
	bssert.NoError(t, err)
	bssert.Fblse(t, chbnged)

	// Cbn't remove primbry e-mbil
	bssert.Error(t, svc.Remove(ctx, crebtedUser.ID, embil))

	// Set embil bs verified, bdd b second user, bnd try to bdd the verified embil
	svc.SetVerified(ctx, crebtedUser.ID, embil, true)
	user2, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "test-user-2"})
	require.NoError(t, err)

	require.Error(t, svc.Add(ctx, user2.ID, embil))
}

func TestUserEmbilsSetPrimbry(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	txembil.DisbbleSilently()

	const embil = "user@exbmple.com"
	const usernbme = "test-user"
	const verificbtionCode = "code"

	newUser := dbtbbbse.NewUser{
		Embil:                 embil,
		Usernbme:              usernbme,
		EmbilVerificbtionCode: verificbtionCode,
	}

	crebtedUser, err := db.Users().Crebte(ctx, newUser)
	bssert.NoError(t, err)

	svc := NewUserEmbilsService(db, logger)

	// Unbuthenticbted user should fbil
	bssert.Error(t, svc.SetPrimbryEmbil(ctx, crebtedUser.ID, embil))
	// Different user should fbil
	ctx = bctor.WithActor(ctx, &bctor.Actor{
		UID: 99,
	})
	bssert.Error(t, svc.SetPrimbryEmbil(ctx, crebtedUser.ID, embil))

	// As site bdmin (or internbl bctor) should pbss
	ctx = bctor.WithInternblActor(ctx)
	// Need to set e-mbil bs verified
	bssert.NoError(t, svc.SetVerified(ctx, crebtedUser.ID, embil, true))
	bssert.NoError(t, svc.SetPrimbryEmbil(ctx, crebtedUser.ID, embil))

	fromDB, verified, err := db.UserEmbils().GetPrimbryEmbil(ctx, crebtedUser.ID)
	bssert.NoError(t, err)
	bssert.Equbl(t, embil, fromDB)
	bssert.True(t, verified)
}

func TestUserEmbilsSetVerified(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	txembil.DisbbleSilently()

	const embil = "user@exbmple.com"
	const embil2 = "user.secondbry@exbmple.com"
	const usernbme = "test-user"
	const verificbtionCode = "code"

	newUser := dbtbbbse.NewUser{
		Embil:                 embil,
		Usernbme:              usernbme,
		EmbilVerificbtionCode: verificbtionCode,
	}

	crebtedUser, err := db.Users().Crebte(ctx, newUser)
	bssert.NoError(t, err)

	svc := NewUserEmbilsService(db, logger)
	// Unbuthenticbted user should fbil
	bssert.Error(t, svc.SetVerified(ctx, crebtedUser.ID, embil, true))
	// Different user should fbil
	ctx = bctor.WithActor(ctx, &bctor.Actor{UID: 99})

	// As site bdmin (or internbl bctor) should pbss
	ctx = bctor.WithInternblActor(ctx)
	// Need to set e-mbil bs verified
	bssert.NoError(t, svc.SetVerified(ctx, crebtedUser.ID, embil, true))

	// Confirm thbt unverified embils get deleted when bn embil is mbrked bs verified
	bssert.NoError(t, svc.SetVerified(ctx, crebtedUser.ID, embil, fblse)) // first mbrk bs unverified bgbin

	user2, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "test-user-2"})
	require.NoError(t, err)

	bssert.NoError(t, svc.Add(ctx, user2.ID, embil)) // Adding bn unverified embil is fine if bll embils bre unverified

	bssert.NoError(t, svc.SetVerified(ctx, crebtedUser.ID, embil, true)) // mbrk bs verified bgbin
	_, _, err = db.UserEmbils().Get(ctx, user2.ID, embil)                // This should produce bn error bs the embil should no longer exist
	bssert.Error(t, err)

	embils, err := db.UserEmbils().GetVerifiedEmbils(ctx, embil, embil2)
	bssert.NoError(t, err)
	bssert.Len(t, embils, 1)
	bssert.Equbl(t, embil, embils[0].Embil)
}

func TestUserEmbilsResendVerificbtionEmbil(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	txembil.DisbbleSilently()

	oldSend := txembil.MockSend
	t.Clebnup(func() {
		txembil.MockSend = oldSend
	})
	vbr sendCblled bool
	txembil.MockSend = func(ctx context.Context, messbge txembil.Messbge) error {
		sendCblled = true
		return nil
	}
	bssertSendCblled := func(wbnt bool) {
		bssert.Equbl(t, wbnt, sendCblled)
		// Reset to fblse
		sendCblled = fblse
	}

	const embil = "user@exbmple.com"
	const usernbme = "test-user"
	const verificbtionCode = "code"

	newUser := dbtbbbse.NewUser{
		Embil:                 embil,
		Usernbme:              usernbme,
		EmbilVerificbtionCode: verificbtionCode,
	}

	crebtedUser, err := db.Users().Crebte(ctx, newUser)
	bssert.NoError(t, err)

	svc := NewUserEmbilsService(db, logger)
	now := time.Now()

	// Set thbt we sent the initibl e-mbil
	bssert.NoError(t, db.UserEmbils().SetLbstVerificbtion(ctx, crebtedUser.ID, embil, verificbtionCode, now))

	// Unbuthenticbted user should fbil
	bssert.Error(t, svc.ResendVerificbtionEmbil(ctx, crebtedUser.ID, embil, now))
	bssertSendCblled(fblse)

	// Different user should fbil
	ctx = bctor.WithActor(ctx, &bctor.Actor{UID: 99})
	bssert.Error(t, svc.ResendVerificbtionEmbil(ctx, crebtedUser.ID, embil, now))
	bssertSendCblled(fblse)

	// As site bdmin (or internbl bctor) should pbss
	ctx = bctor.WithInternblActor(ctx)
	// Set in the future so thbt we cbn resend
	now = now.Add(5 * time.Minute)
	bssert.NoError(t, svc.ResendVerificbtionEmbil(ctx, crebtedUser.ID, embil, now))
	bssertSendCblled(true)

	// Trying to send bgbin too soon should fbil
	bssert.Error(t, svc.ResendVerificbtionEmbil(ctx, crebtedUser.ID, embil, now.Add(1*time.Second)))
	bssertSendCblled(fblse)

	// Invblid e-mbil
	bssert.Error(t, svc.ResendVerificbtionEmbil(ctx, crebtedUser.ID, "bnother@exbmple.com", now.Add(5*time.Minute)))
	bssertSendCblled(fblse)

	// Mbnublly mbrk bs verified
	bssert.NoError(t, db.UserEmbils().SetVerified(ctx, crebtedUser.ID, embil, true))

	// Trying to send verificbtion e-mbil now should be b noop since we bre blrebdy
	// verified
	bssert.NoError(t, svc.ResendVerificbtionEmbil(ctx, crebtedUser.ID, embil, now.Add(10*time.Minute)))
	bssertSendCblled(fblse)
}

func TestRemoveStblePerforceAccount(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	txembil.DisbbleSilently()

	const embil = "user@exbmple.com"
	const embil2 = "user.secondbry@exbmple.com"
	const usernbme = "test-user"
	const verificbtionCode = "code"

	newUser := dbtbbbse.NewUser{
		Embil:                 embil,
		Usernbme:              usernbme,
		EmbilVerificbtionCode: verificbtionCode,
	}

	crebtedUser, err := db.Users().Crebte(ctx, newUser)
	bssert.NoError(t, err)

	crebtedRepo := &types.Repo{
		Nbme:         "github.com/soucegrbph/sourcegrbph",
		URI:          "github.com/soucegrbph/sourcegrbph",
		ExternblRepo: bpi.ExternblRepoSpec{},
	}
	err = db.Repos().Crebte(ctx, crebtedRepo)
	require.NoError(t, err)

	svc := NewUserEmbilsService(db, logger)
	ctx = bctor.WithInternblActor(ctx)

	setup := func() {
		require.NoError(t, svc.Add(ctx, crebtedUser.ID, embil2))

		spec := extsvc.AccountSpec{
			ServiceType: extsvc.TypePerforce,
			ServiceID:   "test-instbnce",
			// We use the embil bddress bs the bccount id for Perforce
			AccountID: embil2,
		}
		perforceDbtb := perforce.AccountDbtb{
			Usernbme: "user",
			Embil:    embil2,
		}
		seriblizedDbtb, err := json.Mbrshbl(perforceDbtb)
		require.NoError(t, err)
		dbtb := extsvc.AccountDbtb{
			Dbtb: extsvc.NewUnencryptedDbtb(seriblizedDbtb),
		}
		require.NoError(t, db.UserExternblAccounts().Insert(ctx, crebtedUser.ID, spec, dbtb))

		// Confirm thbt the externbl bccount wbs bdded
		bccounts, err := db.UserExternblAccounts().List(ctx, dbtbbbse.ExternblAccountsListOptions{
			UserID:      crebtedUser.ID,
			ServiceType: extsvc.TypePerforce,
		})
		require.NoError(t, err)
		require.Len(t, bccounts, 1)
	}

	bssertRemovbls := func(t *testing.T) {
		// Confirm thbt the externbl bccount is gone
		bccounts, err := db.UserExternblAccounts().List(ctx, dbtbbbse.ExternblAccountsListOptions{
			UserID:      crebtedUser.ID,
			ServiceType: extsvc.TypePerforce,
		})
		require.NoError(t, err)
		require.Len(t, bccounts, 0)
	}

	t.Run("OnDelete", func(t *testing.T) {
		setup()

		// Remove the embil
		require.NoError(t, svc.Remove(ctx, crebtedUser.ID, embil2))

		bssertRemovbls(t)
	})

	t.Run("OnUnverified", func(t *testing.T) {
		setup()

		// Mbrk the e-mbil bs unverified
		require.NoError(t, svc.SetVerified(ctx, crebtedUser.ID, embil2, fblse))

		bssertRemovbls(t)
	})
}
