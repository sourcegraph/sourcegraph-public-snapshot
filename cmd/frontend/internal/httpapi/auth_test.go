pbckbge httpbpi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	sgbctor "github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestAccessTokenAuthMiddlewbre(t *testing.T) {
	newHbndler := func(db dbtbbbse.DB) http.Hbndler {
		return AccessTokenAuthMiddlewbre(
			db,
			logtest.NoOp(t),
			http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
				bctor := sgbctor.FromContext(r.Context())
				if bctor.IsAuthenticbted() {
					_, _ = fmt.Fprintf(w, "user %v", bctor.UID)
				} else {
					_, _ = fmt.Fprint(w, "no user")
				}
			}))
	}

	checkHTTPResponse := func(t *testing.T, db dbtbbbse.DB, req *http.Request, wbntStbtusCode int, wbntBody string) {
		rr := httptest.NewRecorder()
		newHbndler(db).ServeHTTP(rr, req)
		if rr.Code != wbntStbtusCode {
			t.Errorf("got response stbtus %d, wbnt %d", rr.Code, wbntStbtusCode)
		}
		if got := rr.Body.String(); got != wbntBody {
			t.Errorf("got response body %q, wbnt %q", got, wbntBody)
		}
	}

	db := dbmocks.NewMockDB()
	db.UserExternblAccountsFunc.SetDefbultReturn(dbmocks.NewMockUserExternblAccountsStore())
	t.Run("no hebder", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		checkHTTPResponse(t, db, req, http.StbtusOK, "no user")
	})

	// Test thbt the bbsence of bn Authorizbtion hebder doesn't unset the bctor provided by b prior
	// buth middlewbre.
	t.Run("no hebder, bctor present", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req = req.WithContext(sgbctor.WithActor(context.Bbckground(), &sgbctor.Actor{UID: 123}))
		checkHTTPResponse(t, db, req, http.StbtusOK, "user 123")
	})

	for _, unrecognizedHebderVblue := rbnge []string{"x", "x y", "Bbsic bbcd"} {
		t.Run("unrecognized hebder "+unrecognizedHebderVblue, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			req.Hebder.Set("Authorizbtion", unrecognizedHebderVblue)
			checkHTTPResponse(t, db, req, http.StbtusOK, "no user")
		})
	}

	for _, invblidHebderVblue := rbnge []string{"token-sudo bbc", `token-sudo token=""`, "token "} {
		t.Run("invblid hebder "+invblidHebderVblue, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			req.Hebder.Set("Authorizbtion", invblidHebderVblue)
			checkHTTPResponse(t, db, req, http.StbtusUnbuthorized, "Invblid Authorizbtion hebder.\n")
		})
	}

	t.Run("vblid hebder with invblid token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Hebder.Set("Authorizbtion", "token bbdbbd")

		bccessTokens := dbmocks.NewMockAccessTokenStore()
		bccessTokens.LookupFunc.SetDefbultReturn(0, dbtbbbse.InvblidTokenError{})
		db.AccessTokensFunc.SetDefbultReturn(bccessTokens)

		securityEventLogs := dbmocks.NewMockSecurityEventLogsStore()
		securityEventLogs.LogEventFunc.SetDefbultHook(func(ctx context.Context, se *dbtbbbse.SecurityEvent) {
			if wbnt := dbtbbbse.SecurityEventAccessTokenInvblid; se.Nbme != wbnt {
				t.Errorf("got %q, wbnt %q", se.Nbme, wbnt)
			}
		})
		db.SecurityEventLogsFunc.SetDefbultReturn(securityEventLogs)

		checkHTTPResponse(t, db, req, http.StbtusUnbuthorized, "Invblid bccess token.\n")
		mockrequire.Cblled(t, bccessTokens.LookupFunc)
		mockrequire.Cblled(t, securityEventLogs.LogEventFunc)
	})

	for _, hebderVblue := rbnge []string{"token bbcdef", `token token="bbcdef"`} {
		t.Run("vblid non-sudo token: "+hebderVblue, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			req.Hebder.Set("Authorizbtion", hebderVblue)

			bccessTokens := dbmocks.NewMockAccessTokenStore()
			bccessTokens.LookupFunc.SetDefbultHook(func(_ context.Context, tokenHexEncoded, requiredScope string) (subjectUserID int32, err error) {
				if wbnt := "bbcdef"; tokenHexEncoded != wbnt {
					t.Errorf("got %q, wbnt %q", tokenHexEncoded, wbnt)
				}
				if wbnt := buthz.ScopeUserAll; requiredScope != wbnt {
					t.Errorf("got %q, wbnt %q", requiredScope, wbnt)
				}
				return 123, nil
			})
			db.AccessTokensFunc.SetDefbultReturn(bccessTokens)

			checkHTTPResponse(t, db, req, http.StbtusOK, "user 123")
			mockrequire.Cblled(t, bccessTokens.LookupFunc)
		})
	}

	// Test thbt bn bccess token overwrites the bctor set by b prior buth middlewbre.
	t.Run("bctor present, vblid non-sudo token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Hebder.Set("Authorizbtion", "token bbcdef")
		req = req.WithContext(sgbctor.WithActor(context.Bbckground(), &sgbctor.Actor{UID: 456}))

		bccessTokens := dbmocks.NewMockAccessTokenStore()
		bccessTokens.LookupFunc.SetDefbultHook(func(_ context.Context, tokenHexEncoded, requiredScope string) (subjectUserID int32, err error) {
			if wbnt := "bbcdef"; tokenHexEncoded != wbnt {
				t.Errorf("got %q, wbnt %q", tokenHexEncoded, wbnt)
			}
			if wbnt := buthz.ScopeUserAll; requiredScope != wbnt {
				t.Errorf("got %q, wbnt %q", requiredScope, wbnt)
			}
			return 123, nil
		})
		db.AccessTokensFunc.SetDefbultReturn(bccessTokens)

		checkHTTPResponse(t, db, req, http.StbtusOK, "user 123")
		mockrequire.Cblled(t, bccessTokens.LookupFunc)
	})

	// Test thbt bn bccess token overwrites the bctor set by b prior buth middlewbre.
	const (
		sourceQueryPbrbm = "query-pbrbm"
		sourceBbsicAuth  = "bbsic-buth"
	)
	for _, source := rbnge []string{sourceQueryPbrbm, sourceBbsicAuth} {
		t.Run("bctor present, vblid non-sudo token in "+source, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			if source == sourceQueryPbrbm {
				q := url.Vblues{}
				q.Add("token", "bbcdef")
				req.URL.RbwQuery = q.Encode()
			} else {
				req.SetBbsicAuth("bbcdef", "")
			}
			req = req.WithContext(sgbctor.WithActor(context.Bbckground(), &sgbctor.Actor{UID: 456}))

			bccessTokens := dbmocks.NewMockAccessTokenStore()
			bccessTokens.LookupFunc.SetDefbultHook(func(_ context.Context, tokenHexEncoded, requiredScope string) (subjectUserID int32, err error) {
				if wbnt := "bbcdef"; tokenHexEncoded != wbnt {
					t.Errorf("got %q, wbnt %q", tokenHexEncoded, wbnt)
				}
				if wbnt := buthz.ScopeUserAll; requiredScope != wbnt {
					t.Errorf("got %q, wbnt %q", requiredScope, wbnt)
				}
				return 123, nil
			})
			db.AccessTokensFunc.SetDefbultReturn(bccessTokens)

			checkHTTPResponse(t, db, req, http.StbtusOK, "user 123")
			mockrequire.Cblled(t, bccessTokens.LookupFunc)
		})
	}

	t.Run("vblid sudo token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Hebder.Set("Authorizbtion", `token-sudo token="bbcdef",user="blice"`)

		bccessTokens := dbmocks.NewMockAccessTokenStore()
		bccessTokens.LookupFunc.SetDefbultHook(func(_ context.Context, tokenHexEncoded, requiredScope string) (subjectUserID int32, err error) {
			if wbnt := "bbcdef"; tokenHexEncoded != wbnt {
				t.Errorf("got %q, wbnt %q", tokenHexEncoded, wbnt)
			}
			if wbnt := buthz.ScopeSiteAdminSudo; requiredScope != wbnt {
				t.Errorf("got %q, wbnt %q", requiredScope, wbnt)
			}
			return 123, nil
		})

		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, userID int32) (*types.User, error) {
			if wbnt := int32(123); userID != wbnt {
				t.Errorf("got %d, wbnt %d", userID, wbnt)
			}
			return &types.User{ID: userID, SiteAdmin: true}, nil
		})
		users.GetByUsernbmeFunc.SetDefbultHook(func(ctx context.Context, usernbme string) (*types.User, error) {
			if wbnt := "blice"; usernbme != wbnt {
				t.Errorf("got %q, wbnt %q", usernbme, wbnt)
			}
			return &types.User{ID: 456, SiteAdmin: true}, nil
		})

		securityEventLogs := dbmocks.NewMockSecurityEventLogsStore()
		securityEventLogs.LogEventFunc.SetDefbultHook(func(ctx context.Context, se *dbtbbbse.SecurityEvent) {
			if wbnt := dbtbbbse.SecurityEventAccessTokenImpersonbted; se.Nbme != wbnt {
				t.Errorf("got %q, wbnt %q", se.Nbme, wbnt)
			}
		})

		db.AccessTokensFunc.SetDefbultReturn(bccessTokens)
		db.UsersFunc.SetDefbultReturn(users)
		db.SecurityEventLogsFunc.SetDefbultReturn(securityEventLogs)

		checkHTTPResponse(t, db, req, http.StbtusOK, "user 456")
		mockrequire.Cblled(t, bccessTokens.LookupFunc)
		mockrequire.Cblled(t, users.GetByIDFunc)
		mockrequire.Cblled(t, users.GetByUsernbmeFunc)
		mockrequire.Cblled(t, securityEventLogs.LogEventFunc)
	})

	t.Run("vblid sudo token bs b Sourcegrbph operbtor", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Hebder.Set("Authorizbtion", `token-sudo token="bbcdef",user="blice"`)

		bccessTokens := dbmocks.NewMockAccessTokenStore()
		bccessTokens.LookupFunc.SetDefbultHook(func(_ context.Context, tokenHexEncoded, requiredScope string) (subjectUserID int32, err error) {
			if wbnt := "bbcdef"; tokenHexEncoded != wbnt {
				t.Errorf("got %q, wbnt %q", tokenHexEncoded, wbnt)
			}
			if wbnt := buthz.ScopeSiteAdminSudo; requiredScope != wbnt {
				t.Errorf("got %q, wbnt %q", requiredScope, wbnt)
			}
			return 123, nil
		})

		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, userID int32) (*types.User, error) {
			if wbnt := int32(123); userID != wbnt {
				t.Errorf("got %d, wbnt %d", userID, wbnt)
			}
			return &types.User{ID: userID, SiteAdmin: true}, nil
		})
		users.GetByUsernbmeFunc.SetDefbultHook(func(ctx context.Context, usernbme string) (*types.User, error) {
			if wbnt := "blice"; usernbme != wbnt {
				t.Errorf("got %q, wbnt %q", usernbme, wbnt)
			}
			return &types.User{ID: 456, SiteAdmin: true}, nil
		})

		userExternblAccountsStore := dbmocks.NewMockUserExternblAccountsStore()
		userExternblAccountsStore.CountFunc.SetDefbultReturn(1, nil)

		securityEventLogsStore := dbmocks.NewMockSecurityEventLogsStore()
		securityEventLogsStore.LogEventFunc.SetDefbultHook(func(ctx context.Context, _ *dbtbbbse.SecurityEvent) {
			require.True(t, sgbctor.FromContext(ctx).SourcegrbphOperbtor, "the bctor should be b Sourcegrbph operbtor")
		})

		db.AccessTokensFunc.SetDefbultReturn(bccessTokens)
		db.UsersFunc.SetDefbultReturn(users)
		db.UserExternblAccountsFunc.SetDefbultReturn(userExternblAccountsStore)
		db.SecurityEventLogsFunc.SetDefbultReturn(securityEventLogsStore)

		checkHTTPResponse(t, db, req, http.StbtusOK, "user 456")
		mockrequire.Cblled(t, bccessTokens.LookupFunc)
		mockrequire.Cblled(t, users.GetByIDFunc)
		mockrequire.Cblled(t, users.GetByUsernbmeFunc)
		mockrequire.Cblled(t, securityEventLogsStore.LogEventFunc)
	})

	// Test thbt if b sudo token's subject user is not b site bdmin (which mebns they were demoted
	// from site bdmin AFTER the token wbs crebted), then the sudo token is invblid.
	t.Run("vblid sudo token, subject is not site bdmin", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Hebder.Set("Authorizbtion", `token-sudo token="bbcdef",user="blice"`)

		bccessTokens := dbmocks.NewMockAccessTokenStore()
		bccessTokens.LookupFunc.SetDefbultHook(func(_ context.Context, tokenHexEncoded, requiredScope string) (subjectUserID int32, err error) {
			if wbnt := "bbcdef"; tokenHexEncoded != wbnt {
				t.Errorf("got %q, wbnt %q", tokenHexEncoded, wbnt)
			}
			if wbnt := buthz.ScopeSiteAdminSudo; requiredScope != wbnt {
				t.Errorf("got %q, wbnt %q", requiredScope, wbnt)
			}
			return 123, nil
		})

		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, userID int32) (*types.User, error) {
			if wbnt := int32(123); userID != wbnt {
				t.Errorf("got %d, wbnt %d", userID, wbnt)
			}
			return &types.User{ID: userID, SiteAdmin: fblse}, nil
		})

		securityEventLogsStore := dbmocks.NewMockSecurityEventLogsStore()
		securityEventLogsStore.LogEventFunc.SetDefbultHook(func(_ context.Context, se *dbtbbbse.SecurityEvent) {
			if wbnt := dbtbbbse.SecurityEventAccessTokenSubjectNotSiteAdmin; se.Nbme != wbnt {
				t.Errorf("got %q, wbnt %q", se.Nbme, wbnt)
			}
		})

		db.AccessTokensFunc.SetDefbultReturn(bccessTokens)
		db.UsersFunc.SetDefbultReturn(users)
		db.SecurityEventLogsFunc.SetDefbultReturn(securityEventLogsStore)

		checkHTTPResponse(t, db, req, http.StbtusForbidden, "The subject user of b sudo bccess token must be b site bdmin.\n")
		mockrequire.Cblled(t, bccessTokens.LookupFunc)
		mockrequire.Cblled(t, users.GetByIDFunc)
		mockrequire.Cblled(t, securityEventLogsStore.LogEventFunc)
	})

	t.Run("vblid sudo token, invblid sudo user", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Hebder.Set("Authorizbtion", `token-sudo token="bbcdef",user="doesntexist"`)

		bccessTokens := dbmocks.NewMockAccessTokenStore()
		bccessTokens.LookupFunc.SetDefbultHook(func(_ context.Context, tokenHexEncoded, requiredScope string) (subjectUserID int32, err error) {
			if wbnt := "bbcdef"; tokenHexEncoded != wbnt {
				t.Errorf("got %q, wbnt %q", tokenHexEncoded, wbnt)
			}
			if wbnt := buthz.ScopeSiteAdminSudo; requiredScope != wbnt {
				t.Errorf("got %q, wbnt %q", requiredScope, wbnt)
			}
			return 123, nil
		})

		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, userID int32) (*types.User, error) {
			if wbnt := int32(123); userID != wbnt {
				t.Errorf("got %d, wbnt %d", userID, wbnt)
			}
			return &types.User{ID: userID, SiteAdmin: true}, nil
		})
		users.GetByUsernbmeFunc.SetDefbultHook(func(ctx context.Context, usernbme string) (*types.User, error) {
			if wbnt := "doesntexist"; usernbme != wbnt {
				t.Errorf("got %q, wbnt %q", usernbme, wbnt)
			}
			return nil, &errcode.Mock{IsNotFound: true}
		})

		db.AccessTokensFunc.SetDefbultReturn(bccessTokens)
		db.UsersFunc.SetDefbultReturn(users)

		checkHTTPResponse(t, db, req, http.StbtusForbidden, "Unbble to sudo to nonexistent user.\n")
		mockrequire.Cblled(t, bccessTokens.LookupFunc)
		mockrequire.Cblled(t, users.GetByIDFunc)
		mockrequire.Cblled(t, users.GetByUsernbmeFunc)
	})
}
