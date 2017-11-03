package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	oidc "github.com/coreos/go-oidc"
)

var appHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "":
		w.Write([]byte("This is the home"))
	case "/page":
		w.Write([]byte("This is a page"))
	default:
		http.Error(w, "", http.StatusNotFound)
	}
})

// providerJSON is the JSON structure the OIDC provider returns at its discovery endpoing
type providerJSON struct {
	Issuer      string `json:"issuer"`
	AuthURL     string `json:"authorization_endpoint"`
	TokenURL    string `json:"token_endpoint"`
	JWKSURL     string `json:"jwks_uri"`
	UserInfoURL string `json:"userinfo_endpoint"`
}

// new OIDCIDServer returns a new running mock OIDC ID Provider service. It is the caller's
// responsibility to call Close().
func newOIDCIDServer(t *testing.T, code string) *httptest.Server {
	s := http.NewServeMux()
	s.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(providerJSON{
			Issuer:   oidcIDProvider,
			AuthURL:  oidcIDProvider + "/oauth2/v1/authorize",
			TokenURL: oidcIDProvider + "/oauth2/v1/token",
		})
	})
	s.HandleFunc("/oauth2/v1/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "unexpected", http.StatusBadRequest)
			return
		}
		b, _ := ioutil.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(b))

		check(t, code == values.Get("code"), "code did not match expected")
		checkEq(t, "authorization_code", values.Get("grant_type"), "wrong grant_type")
		redirectURI, _ := url.QueryUnescape(values.Get("redirect_uri"))
		checkEq(t, appURL+"/.auth/callback", redirectURI, "wrong redirect_uri")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"access_token": "aaaaa",
			"token_type": "Bearer",
			"expires_in": 3600,
			"scope": "openid",
			"id_token": "test_id_token_f4bdefbd77f"
		}`))
	})
	return httptest.NewServer(s)
}

func Test_newOIDCAuthHandler(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "sourcegraph-oidc-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempdir)

	oidcIDServer := newOIDCIDServer(t, "THECODE")
	defer oidcIDServer.Close()

	oidcIDProvider = oidcIDServer.URL
	oidcClientID = "aaaaaaaaaaaaaa"
	oidcClientSecret = "aaaaaaaaaaaaaaaaaaaaaaaaa"

	mockVerifyIDToken = func(rawIDToken string) *oidc.IDToken {
		if rawIDToken != "test_id_token_f4bdefbd77f" {
			t.Fatalf("unexpected raw ID token: %s", rawIDToken)
		}
		return &oidc.IDToken{
			Issuer:  oidcIDServer.URL,
			Subject: "test-subject",
			Expiry:  time.Now().Add(time.Hour),
			Nonce:   "THESTATE", // we re-use the state CSRF token as the nonce
		}
	}

	authedHandler, err := newOIDCAuthHandler(context.Background(), appHandler, false, appURL)
	if err != nil {
		t.Fatal(err)
	}

	doRequest := func(method, urlStr, body string, cookies []*http.Cookie) *http.Response {
		req := httptest.NewRequest(method, urlStr, bytes.NewBufferString(body))
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
		respRecorder := httptest.NewRecorder()
		authedHandler.ServeHTTP(respRecorder, req)
		return respRecorder.Result()
	}

	{ // unauthenticated homepage visit -> login redirect
		resp := doRequest("GET", appURL, "", nil)
		checkEq(t, http.StatusFound, resp.StatusCode, "wrong response code")
		checkEq(t, "/.auth/login", resp.Header.Get("Location"), "wrong redirect URL")
	}
	{ // unauthenticated subpage visit -> login redirect
		resp := doRequest("GET", appURL+"/page", "", nil)
		checkEq(t, http.StatusFound, resp.StatusCode, "wrong response code")
		checkEq(t, "/.auth/login", resp.Header.Get("Location"), "wrong redirect URL")
	}
	{ // unauthenticated non-existent page visit -> login redirect
		resp := doRequest("GET", appURL+"/nonexistent", "", nil)
		checkEq(t, http.StatusFound, resp.StatusCode, "wrong response code")
		checkEq(t, "/.auth/login", resp.Header.Get("Location"), "wrong redirect URL")
	}
	{ // login redirect -> sso login
		resp := doRequest("GET", appURL+"/.auth/login", "", nil)
		checkEq(t, http.StatusFound, resp.StatusCode, "wrong response code")
		locHeader := resp.Header.Get("Location")
		check(t, strings.HasPrefix(locHeader, oidcIDProvider+"/"), "did not redirect to OIDC Provider")
		idpLoginURL, err := url.Parse(locHeader)
		if err != nil {
			t.Fatal(err)
		}
		check(t, oidcClientID == idpLoginURL.Query().Get("client_id"), "client id didn't match")
		checkEq(t, appURL+"/.auth/callback", idpLoginURL.Query().Get("redirect_uri"), "wrong redirect_uri")
		checkEq(t, "code", idpLoginURL.Query().Get("response_type"), "response_type was not \"code\"")
		checkEq(t, "openid", idpLoginURL.Query().Get("scope"), "scope was not \"openid\"")
	}
	{ // OIDC callback without CSRF token -> error
		resp := doRequest("GET", appURL+"/.auth/callback?code=THECODE&state=ASDF", "", nil)
		checkEq(t, http.StatusBadRequest, resp.StatusCode, "wrong status code")
	}
	var authCookies []*http.Cookie
	{ // OIDC callback with CSRF token -> set auth cookies
		resp := doRequest("GET", appURL+"/.auth/callback?code=THECODE&state=THESTATE", "", []*http.Cookie{{Name: oidcStateCookieName, Value: "THESTATE"}})
		checkEq(t, http.StatusFound, resp.StatusCode, "wrong status code")
		checkEq(t, "/", resp.Header.Get("Location"), "wrong redirect URL")
		authCookies = unexpiredCookies(resp)
	}
	{ // authenticated homepage visit
		resp := doRequest("GET", appURL, "", authCookies)
		checkEq(t, http.StatusOK, resp.StatusCode, "wrong response code")
		respBody, _ := ioutil.ReadAll(resp.Body)
		checkEq(t, "This is the home", string(respBody), "wrong response body")
	}
	{ // authenticated subpage visit
		resp := doRequest("GET", appURL+"/page", "", authCookies)
		checkEq(t, http.StatusOK, resp.StatusCode, "wrong response code")
		respBody, _ := ioutil.ReadAll(resp.Body)
		checkEq(t, "This is a page", string(respBody), "wrong response body")
	}
	{ // authenticated non-existent page visit -> 404
		resp := doRequest("GET", appURL+"/nonexistent", "", authCookies)
		checkEq(t, http.StatusNotFound, resp.StatusCode, "wrong response code")
	}
}
