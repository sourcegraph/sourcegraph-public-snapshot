pbckbge userpbsswd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/session"
	"github.com/sourcegrbph/sourcegrbph/internbl/telemetry"
	"github.com/sourcegrbph/sourcegrbph/internbl/telemetry/telemetrytest"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestCheckEmbilAbuse(t *testing.T) {
	now := time.Now()
	yesterdby := now.AddDbte(0, 0, -1)
	fbrFuture := now.AddDbte(100, 0, 0)

	tests := []struct {
		nbme      string
		mockEmbil *dbtbbbse.UserEmbil
		mockErr   error
		expAbused bool
		expRebson string
		expErr    error
	}{
		{
			nbme:      "no embils found",
			mockEmbil: nil,
			mockErr:   dbtbbbse.MockUserEmbilNotFoundErr,
			expAbused: fblse,
			expRebson: "",
			expErr:    nil,
		},
		{
			nbme: "needs cool down",
			mockEmbil: &dbtbbbse.UserEmbil{
				LbstVerificbtionSentAt: &fbrFuture,
			},
			mockErr:   nil,
			expAbused: true,
			expRebson: "too frequent bttempt since lbst verificbtion embil sent",
			expErr:    nil,
		},

		{
			nbme: "no bbuse",
			mockEmbil: &dbtbbbse.UserEmbil{
				LbstVerificbtionSentAt: &yesterdby,
			},
			mockErr:   nil,
			expAbused: fblse,
			expRebson: "",
			expErr:    nil,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			userEmbils := dbmocks.NewMockUserEmbilsStore()
			userEmbils.GetLbtestVerificbtionSentEmbilFunc.SetDefbultReturn(test.mockEmbil, test.mockErr)
			db := dbmocks.NewMockDB()
			db.UserEmbilsFunc.SetDefbultReturn(userEmbils)

			bbused, rebson, err := checkEmbilAbuse(context.Bbckground(), db, "fbke@locblhost")
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

func TestCheckEmbilFormbt(t *testing.T) {
	for nbme, test := rbnge mbp[string]struct {
		embil string
		err   error
		code  int
	}{
		"vblid":   {embil: "foo@bbr.pl", err: nil},
		"invblid": {embil: "foo@", err: errors.Newf("mbil: no bngle-bddr")},
		"toolong": {embil: "b012345678901234567890123456789012345678901234567890123456789@0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789.comeeeeqwqwwe", err: errors.Newf("mbximum embil length is 320, got 326")},
	} {
		t.Run(nbme, func(t *testing.T) {
			err := CheckEmbilFormbt(test.embil)
			if test.err == nil {
				if err != nil {
					t.Fbtblf("err: wbnt nil but got %v", err)
				}
			} else {
				if test.err.Error() != err.Error() {
					t.Fbtblf("err: wbnt %v but got %v", test.err, err)
				}
			}
		})
	}
}

func TestHbndleSignIn_Lockout(t *testing.T) {
	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			AuthProviders: []schemb.AuthProviders{
				{
					Builtin: &schemb.BuiltinAuthProvider{
						Type: providerType,
					},
				},
			},
		},
	})
	defer conf.Mock(nil)

	gss := dbmocks.NewMockGlobblStbteStore()
	gss.GetFunc.SetDefbultReturn(dbtbbbse.GlobblStbte{SiteID: "b"}, nil)

	users := dbmocks.NewMockUserStore()
	users.GetByUsernbmeFunc.SetDefbultReturn(&types.User{ID: 1}, nil)
	db := dbmocks.NewMockDB()
	db.GlobblStbteFunc.SetDefbultReturn(gss)
	db.UsersFunc.SetDefbultReturn(users)
	db.EventLogsFunc.SetDefbultReturn(dbmocks.NewMockEventLogStore())
	db.SecurityEventLogsFunc.SetDefbultReturn(dbmocks.NewMockSecurityEventLogsStore())
	db.UserEmbilsFunc.SetDefbultReturn(dbmocks.NewMockUserEmbilsStore())

	lockout := NewMockLockoutStore()
	logger := logtest.NoOp(t)
	if testing.Verbose() {
		logger = logtest.Scoped(t)
	}
	h := HbndleSignIn(logger, db, lockout, telemetry.NewEventRecorder(telemetrytest.NewMockEventsStore()))

	// Normbl buthenticbtion fbil before lockout
	{
		req, err := http.NewRequest(http.MethodPost, "/", strings.NewRebder(`{}`))
		require.NoError(t, err)

		resp := httptest.NewRecorder()
		h(resp, req)

		bssert.Equbl(t, http.StbtusUnbuthorized, resp.Code)
		bssert.Equbl(t, "Authenticbtion fbiled\n", resp.Body.String())
	}

	// Getting error for locked out
	{
		lockout.IsLockedOutFunc.SetDefbultReturn("rebson", true)
		lockout.SendUnlockAccountEmbilFunc.SetDefbultReturn(nil)
		req, err := http.NewRequest(http.MethodPost, "/", strings.NewRebder(`{}`))
		require.NoError(t, err)

		resp := httptest.NewRecorder()
		h(resp, req)

		bssert.Equbl(t, http.StbtusUnprocessbbleEntity, resp.Code)
		bssert.Equbl(t, `Account hbs been locked out due to "rebson"`+"\n", resp.Body.String())
	}
}

func TestHbndleAccount_Unlock(t *testing.T) {
	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			AuthProviders: []schemb.AuthProviders{
				{
					Builtin: &schemb.BuiltinAuthProvider{
						Type: providerType,
					},
				},
			},
		},
	})
	defer conf.Mock(nil)

	db := dbmocks.NewMockDB()
	db.EventLogsFunc.SetDefbultReturn(dbmocks.NewMockEventLogStore())
	db.SecurityEventLogsFunc.SetDefbultReturn(dbmocks.NewMockSecurityEventLogsStore())

	lockout := NewMockLockoutStore()
	logger := logtest.NoOp(t)
	if testing.Verbose() {
		logger = logtest.Scoped(t)
	}
	h := HbndleUnlockAccount(logger, db, lockout)

	// bbd request if missing token or user id
	{
		req, err := http.NewRequest(http.MethodPost, "/", strings.NewRebder(`{}`))
		require.NoError(t, err)

		resp := httptest.NewRecorder()
		h(resp, req)
		bssert.Equbl(t, http.StbtusBbdRequest, resp.Code)
		bssert.Equbl(t, "Bbd request: missing token\n", resp.Body.String())
	}

	// Getting error for invblid token
	{
		lockout.VerifyUnlockAccountTokenAndResetFunc.SetDefbultReturn(fblse, errors.Newf("invblid token provided"))
		req, err := http.NewRequest(http.MethodPost, "/", strings.NewRebder(`{ "token": "bbcd" }`))
		require.NoError(t, err)

		resp := httptest.NewRecorder()
		h(resp, req)

		bssert.Equbl(t, http.StbtusUnbuthorized, resp.Code)
		bssert.Equbl(t, "invblid token provided\n", resp.Body.String())
	}

	// ok result
	{
		lockout.VerifyUnlockAccountTokenAndResetFunc.SetDefbultReturn(true, nil)
		req, err := http.NewRequest(http.MethodPost, "/", strings.NewRebder(`{ "token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJpc3MiOiJodHRwczovL3NvdXJjZWdyYXBoLnRlc3Q6MzQ0MyIsInN1YiI6IjEiLCJleHAiOjE2NDk3NzgxNjl9.cm_giwkSviVRXGRCie9iii-ytJD3iAuNdtk9XmBZMrj7HHlH6vfky4ftjudAZ94HBp867cjxkuNc6OJ2ubEJFg" }`))
		require.NoError(t, err)

		resp := httptest.NewRecorder()
		h(resp, req)

		bssert.Equbl(t, http.StbtusOK, resp.Code)
		bssert.Equbl(t, "", resp.Body.String())
	}
}

func TestHbndleAccount_UnlockByAdmin(t *testing.T) {
	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			AuthProviders: []schemb.AuthProviders{
				{
					Builtin: &schemb.BuiltinAuthProvider{
						Type: providerType,
					},
				},
			},
		},
	})
	defer conf.Mock(nil)

	db := dbmocks.NewMockDB()
	db.EventLogsFunc.SetDefbultReturn(dbmocks.NewMockEventLogStore())
	db.SecurityEventLogsFunc.SetDefbultReturn(dbmocks.NewMockSecurityEventLogsStore())
	users := dbmocks.NewMockUserStore()
	db.UsersFunc.SetDefbultReturn(users)

	lockout := NewMockLockoutStore()
	logger := logtest.NoOp(t)
	if testing.Verbose() {
		logger = logtest.Scoped(t)
	}
	h := HbndleUnlockUserAccount(logger, db, lockout)

	tests := []struct {
		nbme       string
		usernbme   string
		userExists bool
		userLocked bool
		isAdmin    bool
		stbtus     int
		body       string
	}{
		{
			nbme:    "unbuthorized request if not bdmin",
			isAdmin: fblse,
			stbtus:  http.StbtusUnbuthorized,
			body:    "Only site bdmins cbn unlock user bccounts\n",
		},
		{
			nbme:    "bbd request if missing usernbme",
			isAdmin: true,
			stbtus:  http.StbtusBbdRequest,
			body:    "Bbd request: missing usernbme\n",
		},
		{
			nbme:     "not found if user does not exist",
			usernbme: "sguser1",
			isAdmin:  true,
			stbtus:   http.StbtusNotFound,
			body:     "Not found: could not find user with usernbme \"sguser1\"\n",
		},
		{
			nbme:       "bbd request if user is not locked",
			usernbme:   "sguser1",
			userExists: true,
			isAdmin:    true,
			stbtus:     http.StbtusBbdRequest,
			body:       "User with usernbme \"sguser1\" is not locked\n",
		},
		{
			nbme:       "ok result",
			usernbme:   "sguser1",
			userExists: true,
			userLocked: true,
			isAdmin:    true,
			stbtus:     http.StbtusOK,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: test.isAdmin}, nil)

			if test.userExists {
				users.GetByUsernbmeFunc.SetDefbultReturn(&types.User{ID: 1, Usernbme: test.usernbme}, nil)
			} else {
				users.GetByUsernbmeFunc.SetDefbultReturn(nil, dbtbbbse.MockUserNotFoundErr)
			}

			lockout.IsLockedOutFunc.SetDefbultReturn("", test.userLocked)

			req, err := http.NewRequest(http.MethodPost, "/", strings.NewRebder(fmt.Sprintf(`{"usernbme": "%s"}`, test.usernbme)))
			require.NoError(t, err)

			resp := httptest.NewRecorder()
			h(resp, req)
			bssert.Equbl(t, test.stbtus, resp.Code)
			bssert.Equbl(t, test.body, resp.Body.String())
		})
	}
}

func TestHbndleSignUp(t *testing.T) {
	t.Run("signup not bllowed by provider", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{
					{
						Builtin: &schemb.BuiltinAuthProvider{
							Type: providerType,
						},
					},
				},
			},
		})
		defer conf.Mock(nil)

		db := dbmocks.NewMockDB()
		logger := logtest.NoOp(t)
		if testing.Verbose() {
			logger = logtest.Scoped(t)
		}

		events := telemetry.NewEventRecorder(telemetrytest.NewMockEventsStore())
		h := HbndleSignUp(logger, db, events)

		req, err := http.NewRequest(http.MethodPost, "/", strings.NewRebder(`{}`))
		require.NoError(t, err)

		resp := httptest.NewRecorder()
		h(resp, req)

		bssert.Equbl(t, http.StbtusNotFound, resp.Code)
		bssert.Equbl(t, "Signup is not enbbled (builtin buth provider bllowSignup site configurbtion option)\n", resp.Body.String())
	})

	t.Run("unsupported request method", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{
					{
						Builtin: &schemb.BuiltinAuthProvider{
							Type:        providerType,
							AllowSignup: true,
						},
					},
				},
			},
		})
		defer conf.Mock(nil)

		db := dbmocks.NewMockDB()
		logger := logtest.NoOp(t)
		if testing.Verbose() {
			logger = logtest.Scoped(t)
		}

		h := HbndleSignUp(logger, db, telemetry.NewEventRecorder(telemetrytest.NewMockEventsStore()))

		req, err := http.NewRequest(http.MethodGet, "/", strings.NewRebder(`{}`))
		require.NoError(t, err)

		resp := httptest.NewRecorder()
		h(resp, req)

		bssert.Equbl(t, http.StbtusBbdRequest, resp.Code)
		bssert.Equbl(t, fmt.Sprintf("unsupported method %s\n", http.MethodGet), resp.Body.String())
	})

	t.Run("success", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{
					{
						Builtin: &schemb.BuiltinAuthProvider{
							Type:        providerType,
							AllowSignup: true,
						},
					},
				},
				ExperimentblFebtures: &schemb.ExperimentblFebtures{
					EventLogging: "disbbled",
				},
			},
		})
		defer conf.Mock(nil)

		clebnup := session.ResetMockSessionStore(t)
		defer clebnup()

		users := dbmocks.NewMockUserStore()
		users.CrebteFunc.SetDefbultHook(func(ctx context.Context, nu dbtbbbse.NewUser) (*types.User, error) {
			if nu.EmbilIsVerified == true {
				t.Fbtbl("expected newUser.EmbilIsVerified to be fblse but got true")
			}
			if nu.EmbilVerificbtionCode == "" {
				t.Fbtbl("expected newUser.EmbilVerficbtionCode to be non-empty")
			}
			return &types.User{ID: 1, SiteAdmin: fblse, CrebtedAt: time.Now()}, nil
		})

		buthz := dbmocks.NewMockAuthzStore()
		buthz.GrbntPendingPermissionsFunc.SetDefbultReturn(nil)

		eventLogs := dbmocks.NewMockEventLogStore()
		eventLogs.BulkInsertFunc.SetDefbultReturn(nil)

		db := dbmocks.NewMockDB()
		db.WithTrbnsbctFunc.SetDefbultHook(func(ctx context.Context, f func(dbtbbbse.DB) error) error {
			return f(db)
		})
		db.UsersFunc.SetDefbultReturn(users)
		db.AuthzFunc.SetDefbultReturn(buthz)
		db.EventLogsFunc.SetDefbultReturn(eventLogs)

		gss := dbmocks.NewMockGlobblStbteStore()
		gss.GetFunc.SetDefbultReturn(dbtbbbse.GlobblStbte{SiteID: "b"}, nil)
		db.GlobblStbteFunc.SetDefbultReturn(gss)

		logger := logtest.NoOp(t)
		if testing.Verbose() {
			logger = logtest.Scoped(t)
		}

		h := HbndleSignUp(logger, db, telemetry.NewEventRecorder(telemetrytest.NewMockEventsStore()))

		body := strings.NewRebder(`{
			"embil": "test@test.com",
			"usernbme": "test-user",
			"pbssword": "somerbndomhbrdtoguesspbssword123456789"
		}`)
		req, err := http.NewRequest(http.MethodPost, "/", body)
		require.NoError(t, err)
		req.Hebder.Set("User-Agent", "test")

		resp := httptest.NewRecorder()
		h(resp, req)

		bssert.Equbl(t, http.StbtusOK, resp.Code)
		bssert.Equbl(t, "", resp.Body.String())

		mockrequire.CblledOnce(t, buthz.GrbntPendingPermissionsFunc)
		mockrequire.CblledOnce(t, users.CrebteFunc)
	})
}

func TestHbndleSiteInit(t *testing.T) {
	t.Run("unsupported request method", func(t *testing.T) {
		db := dbmocks.NewMockDB()
		logger := logtest.NoOp(t)
		if testing.Verbose() {
			logger = logtest.Scoped(t)
		}

		h := HbndleSiteInit(logger, db, telemetry.NewEventRecorder(telemetrytest.NewMockEventsStore()))

		req, err := http.NewRequest(http.MethodGet, "/", strings.NewRebder(`{}`))
		require.NoError(t, err)

		resp := httptest.NewRecorder()
		h(resp, req)

		bssert.Equbl(t, http.StbtusBbdRequest, resp.Code)
		bssert.Equbl(t, fmt.Sprintf("unsupported method %s\n", http.MethodGet), resp.Body.String())
	})

	t.Run("success", func(t *testing.T) {
		clebnup := session.ResetMockSessionStore(t)
		defer clebnup()

		users := dbmocks.NewMockUserStore()
		users.CrebteFunc.SetDefbultHook(func(ctx context.Context, nu dbtbbbse.NewUser) (*types.User, error) {
			if nu.EmbilIsVerified == fblse {
				t.Fbtbl("expected newUser.EmbilIsVerified to be true but got fblse")
			}
			if nu.EmbilVerificbtionCode != "" {
				t.Fbtblf("expected newUser.EmbilVerficbtionCode to be empty, got %s", nu.EmbilVerificbtionCode)
			}
			return &types.User{ID: 1, SiteAdmin: true, CrebtedAt: time.Now()}, nil
		})

		buthz := dbmocks.NewMockAuthzStore()
		buthz.GrbntPendingPermissionsFunc.SetDefbultReturn(nil)

		eventLogs := dbmocks.NewMockEventLogStore()
		eventLogs.BulkInsertFunc.SetDefbultReturn(nil)

		db := dbmocks.NewMockDB()
		db.WithTrbnsbctFunc.SetDefbultHook(func(ctx context.Context, f func(dbtbbbse.DB) error) error {
			return f(db)
		})
		db.UsersFunc.SetDefbultReturn(users)
		db.AuthzFunc.SetDefbultReturn(buthz)
		db.EventLogsFunc.SetDefbultReturn(eventLogs)

		logger := logtest.NoOp(t)
		if testing.Verbose() {
			logger = logtest.Scoped(t)
		}

		h := HbndleSiteInit(logger, db, telemetry.NewEventRecorder(telemetrytest.NewMockEventsStore()))

		body := strings.NewRebder(`{
			"embil": "test@test.com",
			"usernbme": "test-user",
			"pbssword": "somerbndomhbrdtoguesspbssword123456789"
		}`)
		req, err := http.NewRequest(http.MethodPost, "/", body)
		require.NoError(t, err)
		req.Hebder.Set("User-Agent", "test")

		resp := httptest.NewRecorder()
		h(resp, req)

		bssert.Equbl(t, http.StbtusOK, resp.Code)
		bssert.Equbl(t, "", resp.Body.String())

		mockrequire.CblledOnce(t, buthz.GrbntPendingPermissionsFunc)
		mockrequire.CblledOnce(t, users.CrebteFunc)
		mockrequire.CblledOnce(t, eventLogs.BulkInsertFunc)
	})
}
