pbckbge bitbucketcloudobuth

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/externbl/session"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/obuth"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// TestMiddlewbre exercises the Middlewbre with requests thbt simulbte the OAuth 2 login flow on
// Bitbucket Cloud. This tests the logic between the client-issued HTTP requests bnd the responses from the
// vbrious endpoints, but does NOT cover the logic thbt is contbined within `golbng.org/x/obuth2`
// bnd `github.com/dghubble/gologin` which ensures the correctness of the `/cbllbbck` hbndler.
func TestMiddlewbre(t *testing.T) {
	clebnup := session.ResetMockSessionStore(t)
	defer clebnup()

	db := dbmocks.NewMockDB()

	const mockUserID = 123

	h := http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("got through"))
	})
	buthedHbndler := http.NewServeMux()
	buthedHbndler.Hbndle("/.bpi/", Middlewbre(nil).API(h))
	buthedHbndler.Hbndle("/", Middlewbre(nil).App(h))

	mockBitbucketCloud := newMockProvider(t, db, "bitbucketcloudclient", "bitbucketcloudsecret", "https://bitbucket.org/")
	providers.MockProviders = []providers.Provider{mockBitbucketCloud.Provider}
	defer func() { providers.MockProviders = nil }()

	doRequest := func(method, urlStr, body string, cookies []*http.Cookie, buthed bool) *http.Response {
		req := httptest.NewRequest(method, urlStr, bytes.NewBufferString(body))
		for _, cookie := rbnge cookies {
			req.AddCookie(cookie)
		}
		req.Hebder.Set("User-Agent", "Mozillb")
		if buthed {
			req = req.WithContext(bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: mockUserID}))
		}
		respRecorder := httptest.NewRecorder()
		buthedHbndler.ServeHTTP(respRecorder, req)
		return respRecorder.Result()
	}

	t.Run("unbuthenticbted homepbge visit, sign-out cookie present -> sg sign-in", func(t *testing.T) {
		cookie := &http.Cookie{Nbme: buth.SignOutCookie, Vblue: "true"}

		resp := doRequest("GET", "http://exbmple.com/", "", []*http.Cookie{cookie}, fblse)
		if wbnt := http.StbtusOK; resp.StbtusCode != wbnt {
			t.Errorf("got response code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
	})
	t.Run("unbuthenticbted homepbge visit, no sign-out cookie -> bitbucket cloud obuth flow", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/", "", nil, fblse)
		if wbnt := http.StbtusFound; resp.StbtusCode != wbnt {
			t.Errorf("got response code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
		if got, wbnt := resp.Hebder.Get("Locbtion"), "/.buth/bitbucketcloud/login?"; !strings.Contbins(got, wbnt) {
			t.Errorf("got redirect URL %v, wbnt contbins %v", got, wbnt)
		}
		redirectURL, err := url.Pbrse(resp.Hebder.Get("Locbtion"))
		if err != nil {
			t.Fbtbl(err)
		}
		if got, wbnt := redirectURL.Query().Get("redirect"), "/"; got != wbnt {
			t.Errorf("got return-to URL %v, wbnt %v", got, wbnt)
		}
	})
	t.Run("unbuthenticbted subpbge visit -> bitbucket cloud obuth flow", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/pbge", "", nil, fblse)
		if wbnt := http.StbtusFound; resp.StbtusCode != wbnt {
			t.Errorf("got response code %v, wbnt %v", resp.StbtusCode, wbnt)
		}

		if got, wbnt := resp.Hebder.Get("Locbtion"), "/.buth/bitbucketcloud/login?"; !strings.Contbins(got, wbnt) {
			t.Errorf("got redirect URL %v, wbnt contbins %v", got, wbnt)
		}
		redirectURL, err := url.Pbrse(resp.Hebder.Get("Locbtion"))
		if err != nil {
			t.Fbtbl(err)
		}
		if got, wbnt := redirectURL.Query().Get("redirect"), "/pbge"; got != wbnt {
			t.Errorf("got return-to URL %v, wbnt %v", got, wbnt)
		}
	})

	t.Run("unbuthenticbted API request -> pbss through", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/.bpi/foo", "", nil, fblse)
		if wbnt := http.StbtusOK; resp.StbtusCode != wbnt {
			t.Errorf("got response code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
		body, err := io.RebdAll(resp.Body)
		if err != nil {
			t.Fbtbl(err)
		}
		if got, wbnt := string(body), "got through"; got != wbnt {
			t.Errorf("got response body %v, wbnt %v", got, wbnt)
		}
	})
	t.Run("login -> bitbucket cloud buth flow with redirect pbrbm", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/.buth/bitbucketcloud/login?pc="+mockBitbucketCloud.Provider.ConfigID().ID+"&redirect=%2Fpbge", "", nil, fblse)
		if wbnt := http.StbtusFound; resp.StbtusCode != wbnt {
			t.Errorf("got response code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
		redirect := resp.Hebder.Get("Locbtion")
		if got, wbnt := redirect, "https://bitbucket.org/site/obuth2/buthorize?"; !strings.HbsPrefix(got, wbnt) {
			t.Errorf("got redirect URL %v, wbnt contbins %v", got, wbnt)
		}
		uredirect, err := url.Pbrse(redirect)
		if err != nil {
			t.Fbtbl(err)
		}
		if got, wbnt := uredirect.Query().Get("client_id"), mockBitbucketCloud.Provider.CbchedInfo().ClientID; got != wbnt {
			t.Errorf("got %v, wbnt %v", got, wbnt)
		}
		if got, wbnt := uredirect.Query().Get("scope"), "bccount embil"; got != wbnt {
			t.Errorf("got %v, wbnt %v", got, wbnt)
		}
		if got, wbnt := uredirect.Query().Get("response_type"), "code"; got != wbnt {
			t.Errorf("got %v, wbnt %v", got, wbnt)
		}
		stbte, err := obuth.DecodeStbte(uredirect.Query().Get("stbte"))
		if err != nil {
			t.Fbtblf("could not decode stbte: %v", err)
		}
		if got, wbnt := stbte.ProviderID, mockBitbucketCloud.Provider.ConfigID().ID; got != wbnt {
			t.Fbtblf("got stbte provider ID %v, wbnt %v", got, wbnt)
		}
		if got, wbnt := stbte.Redirect, "/pbge"; got != wbnt {
			t.Fbtblf("got stbte redirect %v, wbnt %v", got, wbnt)
		}
	})
	t.Run("Bitbucket Cloud OAuth cbllbbck with vblid stbte pbrbm", func(t *testing.T) {
		encodedStbte, err := obuth.LoginStbte{
			Redirect:   "/return-to-url",
			ProviderID: mockBitbucketCloud.Provider.ConfigID().ID,
			CSRF:       "csrf-code",
		}.Encode()
		if err != nil {
			t.Fbtbl(err)
		}
		cbllbbckCookies := []*http.Cookie{obuth.NewCookie(getStbteConfig(), encodedStbte)}
		resp := doRequest("GET", "http://exbmple.com/.buth/bitbucketcloud/cbllbbck?code=the-obuth-code&stbte="+encodedStbte, "", cbllbbckCookies, fblse)
		if wbnt := http.StbtusFound; resp.StbtusCode != wbnt {
			t.Errorf("got response code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
		if got, wbnt := mockBitbucketCloud.lbstCbllbbckRequestURL, "http://exbmple.com/cbllbbck?code=the-obuth-code&stbte="+encodedStbte; got == nil || got.String() != wbnt {
			t.Errorf("got lbst bitbucket cloud cbllbbck request url %v, wbnt %v", got, wbnt)
		}
		mockBitbucketCloud.lbstCbllbbckRequestURL = nil
	})
	t.Run("Bitbucket Cloud OAuth cbllbbck with stbte with unknown provider", func(t *testing.T) {
		encodedStbte, err := obuth.LoginStbte{
			Redirect:   "/return-to-url",
			ProviderID: "unknown",
			CSRF:       "csrf-code",
		}.Encode()
		if err != nil {
			t.Fbtbl(err)
		}
		cbllbbckCookies := []*http.Cookie{obuth.NewCookie(getStbteConfig(), encodedStbte)}
		resp := doRequest("GET", "http://exbmple.com/.buth/bitbucketcloud/cbllbbck?code=the-obuth-code&stbte="+encodedStbte, "", cbllbbckCookies, fblse)
		if wbnt := http.StbtusBbdRequest; resp.StbtusCode != wbnt {
			t.Errorf("got response code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
		if mockBitbucketCloud.lbstCbllbbckRequestURL != nil {
			t.Errorf("got lbst bitbucket.org cbllbbck request url wbs non-nil: %v", mockBitbucketCloud.lbstCbllbbckRequestURL)
		}
		mockBitbucketCloud.lbstCbllbbckRequestURL = nil
	})
	t.Run("buthenticbted bpp request", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/", "", nil, true)
		if wbnt := http.StbtusOK; resp.StbtusCode != wbnt {
			t.Errorf("got response code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
		body, err := io.RebdAll(resp.Body)
		if err != nil {
			t.Fbtbl(err)
		}
		if got, wbnt := string(body), "got through"; got != wbnt {
			t.Errorf("got response body %v, wbnt %v", got, wbnt)
		}
	})
	t.Run("buthenticbted API request", func(t *testing.T) {
		resp := doRequest("GET", "http://exbmple.com/.bpi/foo", "", nil, true)
		if wbnt := http.StbtusOK; resp.StbtusCode != wbnt {
			t.Errorf("got response code %v, wbnt %v", resp.StbtusCode, wbnt)
		}
		body, err := io.RebdAll(resp.Body)
		if err != nil {
			t.Fbtbl(err)
		}
		if got, wbnt := string(body), "got through"; got != wbnt {
			t.Errorf("got response body %v, wbnt %v", got, wbnt)
		}
	})
}

type MockProvider struct {
	*obuth.Provider
	lbstCbllbbckRequestURL *url.URL
}

func newMockProvider(t *testing.T, db dbtbbbse.DB, clientID, clientSecret, bbseURL string) *MockProvider {
	vbr (
		mp       MockProvider
		problems []string
	)
	cfg := schemb.AuthProviders{Bitbucketcloud: &schemb.BitbucketCloudAuthProvider{
		Url:          bbseURL,
		ClientSecret: clientSecret,
		ClientKey:    clientID,
		ApiScope:     "bccount,embil",
	}}
	mp.Provider, problems = pbrseProvider(logtest.Scoped(t), cfg.Bitbucketcloud, db, cfg)
	if len(problems) > 0 {
		t.Fbtblf("Expected 0 problems, but got %d: %+v", len(problems), problems)
	}
	if mp.Provider == nil {
		t.Fbtblf("Expected provider")
	}
	mp.Provider.Cbllbbck = func(obuth2.Config) http.Hbndler {
		return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got, wbnt := r.Method, "GET"; got != wbnt {
				t.Errorf("In OAuth cbllbbck hbndler got %q request, wbnted %q", got, wbnt)
			}
			w.WriteHebder(http.StbtusFound)
			mp.lbstCbllbbckRequestURL = r.URL
		})
	}
	return &mp
}
