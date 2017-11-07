package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"

	oidc "github.com/coreos/go-oidc"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
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

var testUser = "bob-test-user"

// new OIDCIDServer returns a new running mock OIDC ID Provider service. It is the caller's
// responsibility to call Close().
func newOIDCIDServer(t *testing.T, code string) *httptest.Server {
	idBearerToken := "test_id_token_f4bdefbd77f"
	s := http.NewServeMux()

	s.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(providerJSON{
			Issuer:      oidcIDProvider,
			AuthURL:     oidcIDProvider + "/oauth2/v1/authorize",
			TokenURL:    oidcIDProvider + "/oauth2/v1/token",
			UserInfoURL: oidcIDProvider + "/oauth2/v1/userinfo",
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
		w.Write([]byte(fmt.Sprintf(`{
			"access_token": "aaaaa",
			"token_type": "Bearer",
			"expires_in": 3600,
			"scope": "openid",
			"id_token": %q
		}`, idBearerToken)))
	})
	s.HandleFunc("/oauth2/v1/userinfo", func(w http.ResponseWriter, r *http.Request) {
		authzHeader := r.Header.Get("Authorization")
		authzParts := strings.Split(authzHeader, " ")
		if len(authzParts) != 2 {
			t.Fatalf("Expected 2 parts to authz header, instead got %d: %q", len(authzParts), authzHeader)
		}
		if authzParts[0] != "Bearer" {
			t.Fatalf("No bearer token found in authz header %q", authzHeader)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{
			"sub": %q,
			"profile": "This is a profile",
			"email": "bob@foo.com",
			"email_verified": true
		}`, testUser)))
	})

	srv := httptest.NewServer(s)

	// Mock user
	localstore.Mocks.Users.GetByAuth0ID = func(ctx context.Context, uid string) (*sourcegraph.User, error) {
		if uid == srv.URL+":"+testUser {
			return &sourcegraph.User{ID: 123, Auth0ID: uid, Username: uid}, nil
		}
		return nil, fmt.Errorf("user %q not found in mock", uid)
	}

	return srv
}

func Test_newOIDCAuthHandler(t *testing.T) {
	cleanup := session.ResetMockSessionStore(t)
	defer cleanup()

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

	validState := (&authnState{CSRFToken: "THE_CSRF_TOKEN", Redirect: "/redirect"}).Encode()
	mockVerifyIDToken = func(rawIDToken string) *oidc.IDToken {
		if rawIDToken != "test_id_token_f4bdefbd77f" {
			t.Fatalf("unexpected raw ID token: %s", rawIDToken)
		}
		return &oidc.IDToken{
			Issuer:  oidcIDServer.URL,
			Subject: testUser,
			Expiry:  time.Now().Add(time.Hour),
			Nonce:   validState, // we re-use the state param as the nonce
		}
	}

	authedHandler, err := newOIDCAuthHandler(context.Background(), appHandler, appURL)
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

	{
		t.Logf("unauthenticated homepage visit -> login redirect")
		resp := doRequest("GET", appURL, "", nil)
		checkEq(t, http.StatusFound, resp.StatusCode, "wrong response code")
		checkEq(t, "/.auth/login?redirect=", resp.Header.Get("Location"), "wrong redirect URL")
	}
	{
		t.Logf("unauthenticated subpage visit -> login redirect")
		resp := doRequest("GET", appURL+"/page", "", nil)
		checkEq(t, http.StatusFound, resp.StatusCode, "wrong response code")
		checkEq(t, "/.auth/login?redirect=%2Fpage", resp.Header.Get("Location"), "wrong redirect URL")
	}
	{
		t.Logf("unauthenticated non-existent page visit -> login redirect")
		resp := doRequest("GET", appURL+"/nonexistent", "", nil)
		checkEq(t, http.StatusFound, resp.StatusCode, "wrong response code")
		checkEq(t, "/.auth/login?redirect=%2Fnonexistent", resp.Header.Get("Location"), "wrong redirect URL")
	}
	{
		t.Logf("login redirect -> sso login")
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
		checkEq(t, "openid profile email", idpLoginURL.Query().Get("scope"), "scope was not \"openid\"")
	}
	{
		t.Logf("OIDC callback without CSRF token -> error")
		resp := doRequest("GET", appURL+"/.auth/callback?code=THECODE&state=ASDF", "", nil)
		checkEq(t, http.StatusBadRequest, resp.StatusCode, "wrong status code")
	}
	var authCookies []*http.Cookie
	{
		t.Logf("OIDC callback with CSRF token -> set auth cookies")
		resp := doRequest("GET", appURL+"/.auth/callback?code=THECODE&state="+url.PathEscape(validState), "", []*http.Cookie{{Name: oidcStateCookieName, Value: validState}})
		checkEq(t, http.StatusFound, resp.StatusCode, "wrong status code")
		checkEq(t, "/redirect", resp.Header.Get("Location"), "wrong redirect URL")
		authCookies = unexpiredCookies(resp)
	}
	{
		t.Logf("authenticated homepage visit")
		resp := doRequest("GET", appURL, "", authCookies)
		checkEq(t, http.StatusOK, resp.StatusCode, "wrong response code")
		respBody, _ := ioutil.ReadAll(resp.Body)
		checkEq(t, "This is the home", string(respBody), "wrong response body")
	}
	{
		t.Logf("authenticated subpage visit")
		resp := doRequest("GET", appURL+"/page", "", authCookies)
		checkEq(t, http.StatusOK, resp.StatusCode, "wrong response code")
		respBody, _ := ioutil.ReadAll(resp.Body)
		checkEq(t, "This is a page", string(respBody), "wrong response body")
	}
	{
		t.Logf("authenticated non-existent page visit -> 404")
		resp := doRequest("GET", appURL+"/nonexistent", "", authCookies)
		checkEq(t, http.StatusNotFound, resp.StatusCode, "wrong response code")
	}
}
