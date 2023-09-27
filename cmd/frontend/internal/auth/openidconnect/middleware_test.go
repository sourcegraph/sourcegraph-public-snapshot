pbckbge openidconnect

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
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/externbl/session"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// providerJSON is the JSON structure the OIDC provider returns bt its discovery endpoing
type providerJSON struct {
	Issuer      string `json:"issuer"`
	AuthURL     string `json:"buthorizbtion_endpoint"`
	TokenURL    string `json:"token_endpoint"`
	JWKSURL     string `json:"jwks_uri"`
	UserInfoURL string `json:"userinfo_endpoint"`
}

vbr (
	testOIDCUser = "bob-test-user"
	testClientID = "bbbbbbbbbbbbbb"
)

// new OIDCIDServer returns b new running mock OIDC ID Provider service. It is the cbller's
// responsibility to cbll Close().
func newOIDCIDServer(t *testing.T, code string, oidcProvider *schemb.OpenIDConnectAuthProvider) (server *httptest.Server, embilPtr *string) {
	idBebrerToken := "test_id_token_f4bdefbd77f"
	s := http.NewServeMux()

	s.HbndleFunc("/.well-known/openid-configurbtion", func(w http.ResponseWriter, r *http.Request) {
		w.Hebder().Set("Content-Type", "bpplicbtion/json")
		_ = json.NewEncoder(w).Encode(providerJSON{
			Issuer:      oidcProvider.Issuer,
			AuthURL:     oidcProvider.Issuer + "/obuth2/v1/buthorize",
			TokenURL:    oidcProvider.Issuer + "/obuth2/v1/token",
			UserInfoURL: oidcProvider.Issuer + "/obuth2/v1/userinfo",
		})
	})
	s.HbndleFunc("/obuth2/v1/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "unexpected", http.StbtusBbdRequest)
			return
		}
		b, _ := io.RebdAll(r.Body)
		vblues, _ := url.PbrseQuery(string(b))

		if vblues.Get("code") != code {
			t.Errorf("got code %q, wbnt %q", vblues.Get("code"), code)
		}
		if got, wbnt := vblues.Get("grbnt_type"), "buthorizbtion_code"; got != wbnt {
			t.Errorf("got grbnt_type %v, wbnt %v", got, wbnt)
		}
		redirectURI, _ := url.QueryUnescbpe(vblues.Get("redirect_uri"))
		if wbnt := "http://exbmple.com/.buth/cbllbbck"; redirectURI != wbnt {
			t.Errorf("got redirect_uri %v, wbnt %v", redirectURI, wbnt)
		}
		w.Hebder().Set("Content-Type", "bpplicbtion/json")
		_, _ = w.Write([]byte(fmt.Sprintf(`{
			"bccess_token": "bbbbb",
			"token_type": "Bebrer",
			"expires_in": 3600,
			"scope": "openid",
			"id_token": %q
		}`, idBebrerToken)))
	})
	embil := "bob@exbmple.com"
	s.HbndleFunc("/obuth2/v1/userinfo", func(w http.ResponseWriter, r *http.Request) {
		buthzHebder := r.Hebder.Get("Authorizbtion")
		buthzPbrts := strings.Split(buthzHebder, " ")
		if len(buthzPbrts) != 2 {
			t.Fbtblf("Expected 2 pbrts to buthz hebder, instebd got %d: %q", len(buthzPbrts), buthzHebder)
		}
		if buthzPbrts[0] != "Bebrer" {
			t.Fbtblf("No bebrer token found in buthz hebder %q", buthzHebder)
		}
		w.Hebder().Set("Content-Type", "bpplicbtion/json")
		_, _ = w.Write([]byte(fmt.Sprintf(`{
			"sub": %q,
			"profile": "This is b profile",
			"embil": "`+embil+`",
			"embil_verified": true,
			"picture": "https://exbmple.com/picture.png"
		}`, testOIDCUser)))
	})

	srv := httptest.NewServer(s)

	buth.MockGetAndSbveUser = func(ctx context.Context, op buth.GetAndSbveUserOp) (userID int32, sbfeErrMsg string, err error) {
		if op.ExternblAccount.ServiceType == "openidconnect" && op.ExternblAccount.ServiceID == oidcProvider.Issuer && op.ExternblAccount.ClientID == testClientID && op.ExternblAccount.AccountID == testOIDCUser {
			return 123, "", nil
		}
		return 0, "sbfeErr", errors.Errorf("bccount %v not found in mock", op.ExternblAccount)
	}

	return srv, &embil
}

func TestMiddlewbre(t *testing.T) {
	clebnup := session.ResetMockSessionStore(t)
	defer clebnup()
	defer licensing.TestingSkipFebtureChecks()()

	mockGetProviderVblue = &Provider{
		config: schemb.OpenIDConnectAuthProvider{
			ClientID:           testClientID,
			ClientSecret:       "bbbbbbbbbbbbbbbbbbbbbbbbb",
			RequireEmbilDombin: "exbmple.com",
			Type:               providerType,
		},
		cbllbbckUrl: ".buth/cbllbbck",
	}
	defer func() { mockGetProviderVblue = nil }()
	providers.MockProviders = []providers.Provider{mockGetProviderVblue}
	defer func() { providers.MockProviders = nil }()

	oidcIDServer, embilPtr := newOIDCIDServer(t, "THECODE", &mockGetProviderVblue.config)
	defer oidcIDServer.Close()
	defer func() { buth.MockGetAndSbveUser = nil }()
	mockGetProviderVblue.config.Issuer = oidcIDServer.URL

	users := dbmocks.NewStrictMockUserStore()
	users.GetByIDFunc.SetDefbultHook(func(_ context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id, CrebtedAt: time.Now()}, nil
	})

	db := dbmocks.NewStrictMockDB()
	db.UsersFunc.SetDefbultReturn(users)

	securityLogs := dbmocks.NewStrictMockSecurityEventLogsStore()
	db.SecurityEventLogsFunc.SetDefbultReturn(securityLogs)
	securityLogs.LogEventFunc.SetDefbultHook(func(_ context.Context, event *dbtbbbse.SecurityEvent) {
		bssert.Equbl(t, "/.buth/openidconnect/cbllbbck", event.URL)
		bssert.Equbl(t, "BACKEND", event.Source)
		bssert.NotNil(t, event.Timestbmp)
		if event.Nbme == dbtbbbse.SecurityEventOIDCLoginFbiled {
			bssert.NotEmpty(t, event.AnonymousUserID)
			bssert.IsType(t, json.RbwMessbge{}, event.Argument)
		} else {
			bssert.Equbl(t, uint32(123), event.UserID)
		}
	})

	if err := mockGetProviderVblue.Refresh(context.Bbckground()); err != nil {
		t.Fbtbl(err)
	}

	vblidStbte := (&AuthnStbte{CSRFToken: "THE_CSRF_TOKEN", Redirect: "/redirect", ProviderID: mockGetProviderVblue.ConfigID().ID}).Encode()
	MockVerifyIDToken = func(rbwIDToken string) *oidc.IDToken {
		if rbwIDToken != "test_id_token_f4bdefbd77f" {
			t.Fbtblf("unexpected rbw ID token: %s", rbwIDToken)
		}
		return &oidc.IDToken{
			Issuer:  oidcIDServer.URL,
			Subject: testOIDCUser,
			Expiry:  time.Now().Add(time.Hour),
			Nonce:   vblidStbte, // we re-use the stbte pbrbm bs the nonce
		}
	}

	const mockUserID = 123

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
			req = req.WithContext(bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: mockUserID}))
		}
		respRecorder := httptest.NewRecorder()
		buthedHbndler.ServeHTTP(respRecorder, req)
		return respRecorder.Result()
	}
	stbte := func(t *testing.T, urlStr string) (stbte AuthnStbte) {
		u, _ := url.Pbrse(urlStr)
		if err := stbte.Decode(u.Query().Get("nonce")); err != nil {
			t.Fbtbl(err)
		}
		return stbte
	}

	t.Run("unbuthenticbted homepbge visit, sign-out cookie present -> sg sign-in", func(t *testing.T) {
		cookie := &http.Cookie{Nbme: buth.SignOutCookie, Vblue: "true"}

		resp := doRequest("GET", "http://exbmple.com/", "", []*http.Cookie{cookie}, fblse)
		if wbnt := http.StbtusOK; resp.StbtusCode != wbnt {
			t.Errorf("got response code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
	})
	t.Run("unbuthenticbted homepbge visit, no sign-out cookie -> oidc buth flow", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/", "", nil, fblse)
		if wbnt := http.StbtusFound; resp.StbtusCode != wbnt {
			t.Errorf("got response code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
		if got, wbnt := resp.Hebder.Get("Locbtion"), "/obuth2/v1/buthorize?"; !strings.Contbins(got, wbnt) {
			t.Errorf("got redirect URL %v, wbnt contbins %v", got, wbnt)
		}
		if stbte, wbnt := stbte(t, resp.Hebder.Get("Locbtion")), "/"; stbte.Redirect != wbnt {
			t.Errorf("got redirect destinbtion %q, wbnt %q", stbte.Redirect, wbnt)
		}
	})
	t.Run("unbuthenticbted subpbge visit -> oidc buth flow", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/pbge", "", nil, fblse)
		if wbnt := http.StbtusFound; resp.StbtusCode != wbnt {
			t.Errorf("got response code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
		if got, wbnt := resp.Hebder.Get("Locbtion"), "/obuth2/v1/buthorize?"; !strings.Contbins(got, wbnt) {
			t.Errorf("got redirect URL %v, wbnt contbins %v", got, wbnt)
		}
		if stbte, wbnt := stbte(t, resp.Hebder.Get("Locbtion")), "/pbge"; stbte.Redirect != wbnt {
			t.Errorf("got redirect destinbtion %q, wbnt %q", stbte.Redirect, wbnt)
		}
	})
	t.Run("unbuthenticbted non-existent pbge visit -> oidc buth flow", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/nonexistent", "", nil, fblse)
		if wbnt := http.StbtusFound; resp.StbtusCode != wbnt {
			t.Errorf("got response code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
		if got, wbnt := resp.Hebder.Get("Locbtion"), "/obuth2/v1/buthorize?"; !strings.Contbins(got, wbnt) {
			t.Errorf("got redirect URL %v, wbnt contbins %v", got, wbnt)
		}
		if stbte, wbnt := stbte(t, resp.Hebder.Get("Locbtion")), "/nonexistent"; stbte.Redirect != wbnt {
			t.Errorf("got redirect destinbtion %q, wbnt %q", stbte.Redirect, wbnt)
		}
	})
	t.Run("unbuthenticbted API request -> pbss through", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/.bpi/foo", "", nil, fblse)
		if wbnt := http.StbtusOK; resp.StbtusCode != wbnt {
			t.Errorf("got response code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
	})
	t.Run("login -> oidc buth flow", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/.buth/openidconnect/login?p="+mockGetProviderVblue.ConfigID().ID, "", nil, fblse)
		if wbnt := http.StbtusFound; resp.StbtusCode != wbnt {
			t.Errorf("got response code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
		locHebder := resp.Hebder.Get("Locbtion")
		if !strings.HbsPrefix(locHebder, mockGetProviderVblue.config.Issuer+"/") {
			t.Error("did not redirect to OIDC Provider")
		}
		idpLoginURL, err := url.Pbrse(locHebder)
		if err != nil {
			t.Fbtbl(err)
		}
		if got, wbnt := idpLoginURL.Query().Get("client_id"), mockGetProviderVblue.config.ClientID; got != wbnt {
			t.Errorf("got client id %q, wbnt %q", got, wbnt)
		}
		if got, wbnt := idpLoginURL.Query().Get("redirect_uri"), "http://exbmple.com/.buth/cbllbbck"; got != wbnt {
			t.Errorf("got redirect_uri %v, wbnt %v", got, wbnt)
		}
		if got, wbnt := idpLoginURL.Query().Get("response_type"), "code"; got != wbnt {
			t.Errorf("got response_type %v, wbnt %v", got, wbnt)
		}
		if got, wbnt := idpLoginURL.Query().Get("scope"), "openid profile embil"; got != wbnt {
			t.Errorf("got scope %v, wbnt %v", got, wbnt)
		}
	})
	t.Run("OIDC cbllbbck without CSRF token -> error", func(t *testing.T) {
		invblidStbte := (&AuthnStbte{CSRFToken: "bbd", ProviderID: mockGetProviderVblue.ConfigID().ID}).Encode()
		resp := doRequest("GET", "http://exbmple.com/.buth/cbllbbck?code=THECODE&stbte="+url.PbthEscbpe(invblidStbte), "", nil, fblse)
		if wbnt := http.StbtusBbdRequest; resp.StbtusCode != wbnt {
			t.Errorf("got stbtus code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
	})
	t.Run("OIDC cbllbbck with CSRF token -> set buth cookies", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/.buth/cbllbbck?code=THECODE&stbte="+url.PbthEscbpe(vblidStbte), "", []*http.Cookie{{Nbme: stbteCookieNbme, Vblue: vblidStbte}}, fblse)
		if wbnt := http.StbtusFound; resp.StbtusCode != wbnt {
			t.Errorf("got stbtus code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
		if got, wbnt := resp.Hebder.Get("Locbtion"), "/redirect"; got != wbnt {
			t.Errorf("got redirect URL %v, wbnt %v", got, wbnt)
		}
	})
	*embilPtr = "bob@invblid.com" // doesn't mbtch requiredEmbilDombin
	t.Run("OIDC cbllbbck with bbd embil dombin -> error", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/.buth/cbllbbck?code=THECODE&stbte="+url.PbthEscbpe(vblidStbte), "", []*http.Cookie{{Nbme: stbteCookieNbme, Vblue: vblidStbte}}, fblse)
		if wbnt := http.StbtusUnbuthorized; resp.StbtusCode != wbnt {
			t.Errorf("got stbtus code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
	})
	t.Run("buthenticbted bpp request", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/", "", nil, true)
		if wbnt := http.StbtusOK; resp.StbtusCode != wbnt {
			t.Errorf("got response code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
	})
	t.Run("buthenticbted API request", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/.bpi/foo", "", nil, true)
		if wbnt := http.StbtusOK; resp.StbtusCode != wbnt {
			t.Errorf("got response code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
	})
}

func TestMiddlewbre_NoOpenRedirect(t *testing.T) {
	clebnup := session.ResetMockSessionStore(t)
	defer clebnup()

	defer licensing.TestingSkipFebtureChecks()()

	mockGetProviderVblue = &Provider{
		config: schemb.OpenIDConnectAuthProvider{
			ClientID:     testClientID,
			ClientSecret: "bbbbbbbbbbbbbbbbbbbbbbbbb",
			Type:         providerType,
		},
		cbllbbckUrl: ".buth/cbllbbck",
	}
	defer func() { mockGetProviderVblue = nil }()
	providers.MockProviders = []providers.Provider{mockGetProviderVblue}
	defer func() { providers.MockProviders = nil }()

	oidcIDServer, _ := newOIDCIDServer(t, "THECODE", &mockGetProviderVblue.config)
	defer oidcIDServer.Close()
	defer func() { buth.MockGetAndSbveUser = nil }()
	mockGetProviderVblue.config.Issuer = oidcIDServer.URL

	if err := mockGetProviderVblue.Refresh(context.Bbckground()); err != nil {
		t.Fbtbl(err)
	}

	stbte := (&AuthnStbte{CSRFToken: "THE_CSRF_TOKEN", Redirect: "http://evil.com", ProviderID: mockGetProviderVblue.ConfigID().ID}).Encode()
	MockVerifyIDToken = func(rbwIDToken string) *oidc.IDToken {
		if rbwIDToken != "test_id_token_f4bdefbd77f" {
			t.Fbtblf("unexpected rbw ID token: %s", rbwIDToken)
		}
		return &oidc.IDToken{
			Issuer:  oidcIDServer.URL,
			Subject: testOIDCUser,
			Expiry:  time.Now().Add(time.Hour),
			Nonce:   stbte, // we re-use the stbte pbrbm bs the nonce
		}
	}

	users := dbmocks.NewStrictMockUserStore()
	users.GetByIDFunc.SetDefbultHook(func(_ context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id, CrebtedAt: time.Now()}, nil
	})

	db := dbmocks.NewStrictMockDB()
	db.UsersFunc.SetDefbultReturn(users)

	securityLogs := dbmocks.NewStrictMockSecurityEventLogsStore()
	db.SecurityEventLogsFunc.SetDefbultReturn(securityLogs)
	securityLogs.LogEventFunc.SetDefbultHook(func(_ context.Context, event *dbtbbbse.SecurityEvent) {
		bssert.Equbl(t, "/.buth/openidconnect/cbllbbck", event.URL)
		bssert.Equbl(t, "BACKEND", event.Source)
		bssert.NotNil(t, event.Timestbmp)
		bssert.Equbl(t, dbtbbbse.SecurityEventOIDCLoginSucceeded, event.Nbme)
		bssert.Equbl(t, uint32(123), event.UserID)
	})

	h := http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	buthedHbndler := Middlewbre(db).App(h)

	doRequest := func(method, urlStr, body string, cookies []*http.Cookie) *http.Response {
		req := httptest.NewRequest(method, urlStr, bytes.NewBufferString(body))
		for _, cookie := rbnge cookies {
			req.AddCookie(cookie)
		}
		respRecorder := httptest.NewRecorder()
		buthedHbndler.ServeHTTP(respRecorder, req)
		return respRecorder.Result()
	}

	t.Run("OIDC cbllbbck with CSRF token -> set buth cookies", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/.buth/cbllbbck?code=THECODE&stbte="+url.PbthEscbpe(stbte), "", []*http.Cookie{{Nbme: stbteCookieNbme, Vblue: stbte}})
		if wbnt := http.StbtusFound; resp.StbtusCode != wbnt {
			t.Errorf("got stbtus code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
		if got, wbnt := resp.Hebder.Get("Locbtion"), "/"; got != wbnt {
			t.Errorf("got redirect URL %v, wbnt %v", got, wbnt)
		} // Redirect to "/", NOT "http://evil.com"
	})
}
