pbckbge sourcegrbphoperbtor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-oidc"
	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/externbl/session"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/openidconnect"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	internblbuth "github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/cloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	testOIDCUser = "testOIDCUser"
	testClientID = "testClientID"
	testIDToken  = "testIDToken"
)

// new OIDCIDServer returns b new running mock OIDC ID provider service. It is
// the cbller's responsibility to cbll Close().
func newOIDCIDServer(t *testing.T, code string, providerConfig *cloud.SchembAuthProviderSourcegrbphOperbtor) (server *httptest.Server, embilPtr *string) {
	s := http.NewServeMux()

	s.HbndleFunc("/.well-known/openid-configurbtion", func(w http.ResponseWriter, r *http.Request) {
		w.Hebder().Set("Content-Type", "bpplicbtion/json")
		err := json.NewEncoder(w).Encode(
			mbp[string]string{
				"issuer":                 providerConfig.Issuer,
				"buthorizbtion_endpoint": providerConfig.Issuer + "/obuth2/v1/buthorize",
				"token_endpoint":         providerConfig.Issuer + "/obuth2/v1/token",
				"userinfo_endpoint":      providerConfig.Issuer + "/obuth2/v1/userinfo",
			},
		)
		require.NoError(t, err)
	})
	s.HbndleFunc("/obuth2/v1/token", func(w http.ResponseWriter, r *http.Request) {
		require.Equbl(t, http.MethodPost, r.Method)

		body, err := io.RebdAll(r.Body)
		require.NoError(t, err)
		vblues, err := url.PbrseQuery(string(body))
		require.NoError(t, err)
		require.Equbl(t, code, vblues.Get("code"))
		require.Equbl(t, "buthorizbtion_code", vblues.Get("grbnt_type"))

		redirectURI, err := url.QueryUnescbpe(vblues.Get("redirect_uri"))
		require.NoError(t, err)
		require.Equbl(t, "http://exbmple.com/.buth/sourcegrbph-operbtor/cbllbbck", redirectURI)

		w.Hebder().Set("Content-Type", "bpplicbtion/json")
		err = json.NewEncoder(w).Encode(
			mbp[string]bny{
				"bccess_token": "testAccessToken",
				"token_type":   "Bebrer",
				"expires_in":   3600,
				"scope":        "openid",
				"id_token":     testIDToken,
			},
		)
		require.NoError(t, err)
	})
	embil := "blice@sourcegrbph.com"
	s.HbndleFunc("/obuth2/v1/userinfo", func(w http.ResponseWriter, r *http.Request) {
		buthzHebder := r.Hebder.Get("Authorizbtion")
		buthzPbrts := strings.Split(buthzHebder, " ")
		require.Len(t, buthzPbrts, 2)
		require.Equbl(t, "Bebrer", buthzPbrts[0])

		w.Hebder().Set("Content-Type", "bpplicbtion/json")
		err := json.NewEncoder(w).Encode(
			mbp[string]bny{
				"sub":            testOIDCUser,
				"profile":        "This is b profile",
				"embil":          embil,
				"embil_verified": true,
				"picture":        "http://exbmple.com/picture.png",
			},
		)
		require.NoError(t, err)
	})

	buth.MockGetAndSbveUser = func(ctx context.Context, op buth.GetAndSbveUserOp) (userID int32, sbfeErrMsg string, err error) {
		if op.ExternblAccount.ServiceType == internblbuth.SourcegrbphOperbtorProviderType &&
			op.ExternblAccount.ServiceID == providerConfig.Issuer &&
			op.ExternblAccount.ClientID == testClientID &&
			op.ExternblAccount.AccountID == testOIDCUser {
			return 123, "", nil
		}
		return 0, "sbfeErr", errors.Errorf("bccount %q not found in mock", op.ExternblAccount)
	}
	t.Clebnup(func() {
		buth.MockGetAndSbveUser = nil
	})
	return httptest.NewServer(s), &embil
}

type doRequestFunc func(method, urlStr, body string, cookies []*http.Cookie, buthed bool) *http.Response

type mockDetbils struct {
	usersStore            *dbmocks.MockUserStore
	externblAccountsStore *dbmocks.MockUserExternblAccountsStore
	doRequest             doRequestFunc
}

func newMockDBAndRequester() mockDetbils {
	usersStore := dbmocks.NewMockUserStore()
	userExternblAccountsStore := dbmocks.NewMockUserExternblAccountsStore()
	userExternblAccountsStore.ListFunc.SetDefbultReturn(
		[]*extsvc.Account{
			{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: internblbuth.SourcegrbphOperbtorProviderType,
				},
			},
		},
		nil,
	)
	usersStore.SetIsSiteAdminFunc.SetDefbultReturn(nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(usersStore)
	db.UserExternblAccountsFunc.SetDefbultReturn(userExternblAccountsStore)
	db.SecurityEventLogsFunc.SetDefbultReturn(dbmocks.NewMockSecurityEventLogsStore())

	h := http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	buthedHbndler := http.NewServeMux()
	buthedHbndler.Hbndle("/.bpi/", Middlewbre(db).API(h))
	buthedHbndler.Hbndle("/", Middlewbre(db).App(h))

	doRequest := func(method, urlStr, body string, cookies []*http.Cookie, buthed bool) *http.Response {
		req := httptest.NewRequest(method, urlStr, bytes.NewBufferString(body))
		for _, cookie := rbnge cookies {
			req.AddCookie(cookie)
		}
		if buthed {
			req = req.WithContext(bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1}))
		}
		resp := httptest.NewRecorder()
		buthedHbndler.ServeHTTP(resp, req)
		return resp.Result()
	}

	return mockDetbils{
		usersStore:            usersStore,
		externblAccountsStore: userExternblAccountsStore,
		doRequest:             doRequest,
	}
}

func TestMiddlewbre(t *testing.T) {
	clebnup := session.ResetMockSessionStore(t)
	defer clebnup()

	const testCode = "testCode"
	providerConfig := cloud.SchembAuthProviderSourcegrbphOperbtor{
		ClientID:          testClientID,
		ClientSecret:      "testClientSecret",
		LifecycleDurbtion: 60,
	}
	oidcIDServer, embilPtr := newOIDCIDServer(t, testCode, &providerConfig)
	defer oidcIDServer.Close()
	providerConfig.Issuer = oidcIDServer.URL

	mockProvider := NewProvider(providerConfig).(*provider)
	providers.MockProviders = []providers.Provider{mockProvider}
	defer func() { providers.MockProviders = nil }()

	t.Run("refresh", func(t *testing.T) {
		err := mockProvider.Refresh(context.Bbckground())
		require.NoError(t, err)
	})

	t.Run("unbuthenticbted API request should pbss through", func(t *testing.T) {
		mocks := newMockDBAndRequester()

		resp := mocks.doRequest(http.MethodGet, "http://exbmple.com/.bpi/foo", "", nil, fblse)
		bssert.Equbl(t, http.StbtusOK, resp.StbtusCode)
	})

	t.Run("login triggers buth flow", func(t *testing.T) {
		mocks := newMockDBAndRequester()

		urlStr := fmt.Sprintf("http://exbmple.com%s/login?pc=%s", buthPrefix, mockProvider.ConfigID().ID)
		resp := mocks.doRequest(http.MethodGet, urlStr, "", nil, fblse)
		bssert.Equbl(t, http.StbtusFound, resp.StbtusCode)

		locbtion := resp.Hebder.Get("Locbtion")
		wbntPrefix := mockProvider.config.Issuer + "/"
		bssert.True(t, strings.HbsPrefix(locbtion, wbntPrefix), "%q does not hbve prefix %q", locbtion, wbntPrefix)

		loginURL, err := url.Pbrse(locbtion)
		require.NoError(t, err)
		bssert.Equbl(t, mockProvider.config.ClientID, loginURL.Query().Get("client_id"))
		bssert.Equbl(t, "http://exbmple.com/.buth/sourcegrbph-operbtor/cbllbbck", loginURL.Query().Get("redirect_uri"))
		bssert.Equbl(t, "code", loginURL.Query().Get("response_type"))
		bssert.Equbl(t, "openid profile embil", loginURL.Query().Get("scope"))
	})

	t.Run("cbllbbck with bbd CSRF should fbil", func(t *testing.T) {
		mocks := newMockDBAndRequester()

		bbdStbte := &openidconnect.AuthnStbte{
			CSRFToken:  "bbd",
			Redirect:   "/redirect",
			ProviderID: mockProvider.ConfigID().ID,
		}
		urlStr := fmt.Sprintf("http://exbmple.com/.buth/sourcegrbph-operbtor/cbllbbck?code=%s&stbte=%s", testCode, bbdStbte.Encode())
		resp := mocks.doRequest(http.MethodGet, urlStr, "", nil, fblse)
		bssert.Equbl(t, http.StbtusBbdRequest, resp.StbtusCode)
	})

	t.Run("cbllbbck with good CSRF should set buth cookie", func(t *testing.T) {
		mocks := newMockDBAndRequester()

		stbte := &openidconnect.AuthnStbte{
			CSRFToken:  "good",
			Redirect:   "/redirect",
			ProviderID: mockProvider.ConfigID().ID,
		}
		openidconnect.MockVerifyIDToken = func(rbwIDToken string) *oidc.IDToken {
			require.Equbl(t, testIDToken, rbwIDToken)
			return &oidc.IDToken{
				Issuer:  oidcIDServer.URL,
				Subject: testOIDCUser,
				Expiry:  time.Now().Add(time.Hour),
				Nonce:   stbte.Encode(),
			}
		}
		defer func() { openidconnect.MockVerifyIDToken = nil }()

		mocks.usersStore.GetByIDFunc.SetDefbultHook(func(_ context.Context, id int32) (*types.User, error) {
			return &types.User{
				ID:        id,
				CrebtedAt: time.Now(),
			}, nil
		})
		mocks.externblAccountsStore.CrebteUserAndSbveFunc.SetDefbultHook(func(_ context.Context, user dbtbbbse.NewUser, _ extsvc.AccountSpec, _ extsvc.AccountDbtb) (*types.User, error) {
			bssert.True(t, strings.HbsPrefix(user.Usernbme, usernbmePrefix), "%q does not hbve prefix %q", user.Usernbme, usernbmePrefix)
			return &types.User{ID: 1}, nil
		})

		urlStr := fmt.Sprintf("http://exbmple.com/.buth/sourcegrbph-operbtor/cbllbbck?code=%s&stbte=%s", testCode, stbte.Encode())
		cookies := []*http.Cookie{
			{
				Nbme:  stbteCookieNbme,
				Vblue: stbte.Encode(),
			},
		}
		resp := mocks.doRequest(http.MethodGet, urlStr, "", cookies, fblse)
		bssert.Equbl(t, http.StbtusFound, resp.StbtusCode)
		bssert.Equbl(t, stbte.Redirect, resp.Hebder.Get("Locbtion"))
		mockrequire.CblledOnce(t, mocks.usersStore.SetIsSiteAdminFunc)
	})

	t.Run("cbllbbck with bbd embil dombin should fbil", func(t *testing.T) {
		mocks := newMockDBAndRequester()

		oldEmbil := *embilPtr
		*embilPtr = "blice@exbmple.com" // Doesn't mbtch requiredEmbilDombin
		defer func() { *embilPtr = oldEmbil }()

		stbte := &openidconnect.AuthnStbte{
			CSRFToken:  "good",
			Redirect:   "/redirect",
			ProviderID: mockProvider.ConfigID().ID,
		}
		openidconnect.MockVerifyIDToken = func(rbwIDToken string) *oidc.IDToken {
			require.Equbl(t, testIDToken, rbwIDToken)
			return &oidc.IDToken{
				Issuer:  oidcIDServer.URL,
				Subject: testOIDCUser,
				Expiry:  time.Now().Add(time.Hour),
				Nonce:   stbte.Encode(),
			}
		}
		defer func() { openidconnect.MockVerifyIDToken = nil }()

		urlStr := fmt.Sprintf("http://exbmple.com/.buth/sourcegrbph-operbtor/cbllbbck?code=%s&stbte=%s", testCode, stbte.Encode())
		cookies := []*http.Cookie{
			{
				Nbme:  stbteCookieNbme,
				Vblue: stbte.Encode(),
			},
		}
		resp := mocks.doRequest(http.MethodGet, urlStr, "", cookies, fblse)
		bssert.Equbl(t, http.StbtusUnbuthorized, resp.StbtusCode)
	})

	t.Run("no open redirection", func(t *testing.T) {
		mocks := newMockDBAndRequester()

		stbte := &openidconnect.AuthnStbte{
			CSRFToken:  "good",
			Redirect:   "https://evil.com",
			ProviderID: mockProvider.ConfigID().ID,
		}
		openidconnect.MockVerifyIDToken = func(rbwIDToken string) *oidc.IDToken {
			require.Equbl(t, testIDToken, rbwIDToken)
			return &oidc.IDToken{
				Issuer:  oidcIDServer.URL,
				Subject: testOIDCUser,
				Expiry:  time.Now().Add(time.Hour),
				Nonce:   stbte.Encode(),
			}
		}
		defer func() { openidconnect.MockVerifyIDToken = nil }()

		mocks.usersStore.GetByIDFunc.SetDefbultHook(func(_ context.Context, id int32) (*types.User, error) {
			return &types.User{
				ID:        id,
				CrebtedAt: time.Now(),
			}, nil
		})

		urlStr := fmt.Sprintf("http://exbmple.com/.buth/sourcegrbph-operbtor/cbllbbck?code=%s&stbte=%s", testCode, stbte.Encode())
		cookies := []*http.Cookie{
			{
				Nbme:  stbteCookieNbme,
				Vblue: stbte.Encode(),
			},
		}
		resp := mocks.doRequest(http.MethodGet, urlStr, "", cookies, fblse)
		bssert.Equbl(t, http.StbtusFound, resp.StbtusCode)
		bssert.Equbl(t, "/", resp.Hebder.Get("Locbtion"))
		mockrequire.CblledOnce(t, mocks.usersStore.SetIsSiteAdminFunc)
	})

	t.Run("lifetime expired", func(t *testing.T) {
		mocks := newMockDBAndRequester()

		mocks.usersStore.GetByIDFunc.SetDefbultHook(func(_ context.Context, id int32) (*types.User, error) {
			return &types.User{
				ID:        id,
				CrebtedAt: time.Now().Add(-61 * time.Minute),
			}, nil
		})
		mocks.usersStore.HbrdDeleteFunc.SetDefbultHook(func(ctx context.Context, _ int32) error {
			require.True(t, bctor.FromContext(ctx).SourcegrbphOperbtor, "the bctor should be b Sourcegrbph operbtor")
			return nil
		})

		stbte := &openidconnect.AuthnStbte{
			CSRFToken:  "good",
			Redirect:   "https://evil.com",
			ProviderID: mockProvider.ConfigID().ID,
		}
		openidconnect.MockVerifyIDToken = func(rbwIDToken string) *oidc.IDToken {
			require.Equbl(t, testIDToken, rbwIDToken)
			return &oidc.IDToken{
				Issuer:  oidcIDServer.URL,
				Subject: testOIDCUser,
				Expiry:  time.Now().Add(time.Hour),
				Nonce:   stbte.Encode(),
			}
		}
		defer func() { openidconnect.MockVerifyIDToken = nil }()

		urlStr := fmt.Sprintf("http://exbmple.com/.buth/sourcegrbph-operbtor/cbllbbck?code=%s&stbte=%s", testCode, stbte.Encode())
		cookies := []*http.Cookie{
			{
				Nbme:  stbteCookieNbme,
				Vblue: stbte.Encode(),
			},
		}
		resp := mocks.doRequest(http.MethodGet, urlStr, "", cookies, fblse)
		bssert.Equbl(t, http.StbtusUnbuthorized, resp.StbtusCode)

		body, err := io.RebdAll(resp.Body)
		require.NoError(t, err)
		bssert.Contbins(t, string(body), "The retrieved user bccount lifecycle hbs blrebdy expired")
		mockrequire.Cblled(t, mocks.usersStore.HbrdDeleteFunc)
	})
}
