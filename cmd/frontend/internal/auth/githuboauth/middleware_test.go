package githuboauth

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/oauth2"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/oauth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/providers"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/session"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/schema"
)

// TestMiddleware exercises the Middleware with requests that simulate the OAuth 2 login flow on
// GitHub. This tests the logic between the client-issued HTTP requests and the responses from the
// various endpoints, but does NOT cover the logic that is contained within `golang.org/x/oauth2`
// and `github.com/dghubble/gologin` which ensures the correctness of the `/callback` handler.
func TestMiddleware(t *testing.T) {
	logger := logtest.Scoped(t)
	session.ResetMockSessionStore(t)

	db := database.NewDB(logger, dbtest.NewDB(t))

	const mockUserID = 123

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("got through"))
	})
	authedHandler := http.NewServeMux()
	authedHandler.Handle("/.api/", Middleware(nil).API(h))
	authedHandler.Handle("/", Middleware(nil).App(h))

	mockGitHubCom := newMockProvider(t, db, "githubcomclient", "githubcomsecret", "https://github.com/")
	mockGHE := newMockProvider(t, db, "githubenterpriseclient", "githubenterprisesecret", "https://mycompany.com/")
	providers.MockProviders = []providers.Provider{mockGitHubCom.Provider}
	defer func() { providers.MockProviders = nil }()

	doRequest := func(method, urlStr, body string, state string, cookies []*http.Cookie, authed bool) *http.Response {
		req := httptest.NewRequest(method, urlStr, bytes.NewBufferString(body))
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
		req.Header.Set("User-Agent", "Mozilla")
		if authed {
			req = req.WithContext(actor.WithActor(context.Background(), &actor.Actor{UID: mockUserID}))
		}
		respRecorder := httptest.NewRecorder()
		session.SetData(respRecorder, req, "oauthState", state)
		authedHandler.ServeHTTP(respRecorder, req)
		return respRecorder.Result()
	}

	t.Run("unauthenticated homepage visit, sign-out cookie present, access requests enabled -> sg sign-in", func(t *testing.T) {
		cookie := &http.Cookie{Name: session.SignOutCookie, Value: "true"}
		resp := doRequest("GET", "http://example.com/", "", "", []*http.Cookie{cookie}, false)
		if want := http.StatusOK; resp.StatusCode != want {
			t.Errorf("got response code %v, want %v", resp.StatusCode, want)
		}
	})
	t.Run("unauthenticated homepage visit, sign-out cookie present, access requests disabled -> sg sign-in", func(t *testing.T) {
		falseVal := false
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				AuthAccessRequest: &schema.AuthAccessRequest{
					Enabled: &falseVal,
				},
			},
		})
		t.Cleanup(func() { conf.Mock(nil) })

		cookie := &http.Cookie{Name: session.SignOutCookie, Value: "true"}

		resp := doRequest("GET", "http://example.com/", "", "", []*http.Cookie{cookie}, false)
		if want := http.StatusOK; resp.StatusCode != want {
			t.Errorf("got response code %v, want %v", resp.StatusCode, want)
		}
	})
	t.Run("unauthenticated homepage visit, no sign-out cookie, access requests disabled -> github oauth flow", func(t *testing.T) {
		falseVal := false
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				AuthAccessRequest: &schema.AuthAccessRequest{
					Enabled: &falseVal,
				},
			},
		})
		t.Cleanup(func() { conf.Mock(nil) })

		resp := doRequest("GET", "http://example.com/", "", "", nil, false)
		if want := http.StatusFound; resp.StatusCode != want {
			t.Errorf("got response code %v, want %v", resp.StatusCode, want)
		}
		if got, want := resp.Header.Get("Location"), "/.auth/github/login?"; !strings.Contains(got, want) {
			t.Errorf("got redirect URL %v, want contains %v", got, want)
		}
		redirectURL, err := url.Parse(resp.Header.Get("Location"))
		if err != nil {
			t.Fatal(err)
		}
		if got, want := redirectURL.Query().Get("redirect"), "/"; got != want {
			t.Errorf("got return-to URL %v, want %v", got, want)
		}
	})
	t.Run("unauthenticated homepage visit, no sign-out cookie, access requests enabled -> sg sign-in", func(t *testing.T) {
		resp := doRequest("GET", "http://example.com/", "", "", nil, false)
		if want := http.StatusOK; resp.StatusCode != want {
			t.Errorf("got response code %v, want %v", resp.StatusCode, want)
		}
	})
	t.Run("unauthenticated subpage visit, access requests disabled -> github oauth flow", func(t *testing.T) {
		falseVal := false
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				AuthAccessRequest: &schema.AuthAccessRequest{
					Enabled: &falseVal,
				},
			},
		})
		t.Cleanup(func() { conf.Mock(nil) })
		resp := doRequest("GET", "http://example.com/page", "", "", nil, false)
		if want := http.StatusFound; resp.StatusCode != want {
			t.Errorf("got response code %v, want %v", resp.StatusCode, want)
		}

		if got, want := resp.Header.Get("Location"), "/.auth/github/login?"; !strings.Contains(got, want) {
			t.Errorf("got redirect URL %v, want contains %v", got, want)
		}
		redirectURL, err := url.Parse(resp.Header.Get("Location"))
		if err != nil {
			t.Fatal(err)
		}
		if got, want := redirectURL.Query().Get("redirect"), "/page"; got != want {
			t.Errorf("got return-to URL %v, want %v", got, want)
		}
	})
	t.Run("unauthenticated subpage visit, access requests enabled -> sg sign-in", func(t *testing.T) {
		resp := doRequest("GET", "http://example.com/page", "", "", nil, false)
		if want := http.StatusOK; resp.StatusCode != want {
			t.Errorf("got response code %v, want %v", resp.StatusCode, want)
		}
	})

	// Add 2 GitHub auth providers
	providers.MockProviders = []providers.Provider{mockGHE.Provider, mockGitHubCom.Provider}

	t.Run("unauthenticated API request -> pass through", func(t *testing.T) {
		resp := doRequest("GET", "http://example.com/.api/foo", "", "", nil, false)
		if want := http.StatusOK; resp.StatusCode != want {
			t.Errorf("got response code %v, want %v", resp.StatusCode, want)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := string(body), "got through"; got != want {
			t.Errorf("got response body %v, want %v", got, want)
		}
	})
	t.Run("login -> github auth flow (github.com)", func(t *testing.T) {
		resp := doRequest("GET", "http://example.com/.auth/github/login?pc="+mockGitHubCom.Provider.ConfigID().ID, "", "", nil, false)
		if want := http.StatusFound; resp.StatusCode != want {
			t.Errorf("got response code %v, want %v", resp.StatusCode, want)
		}
		redirect := resp.Header.Get("Location")
		if got, want := redirect, "https://github.com/login/oauth/authorize?"; !strings.HasPrefix(got, want) {
			t.Errorf("got redirect URL %v, want contains %v", got, want)
		}
		uredirect, err := url.Parse(redirect)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := uredirect.Query().Get("client_id"), mockGitHubCom.Provider.CachedInfo().ClientID; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := uredirect.Query().Get("scope"), "user:email repo read:org"; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := uredirect.Query().Get("response_type"), "code"; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		state, err := oauth.DecodeState(uredirect.Query().Get("state"))
		if err != nil {
			t.Fatalf("could not decode state: %v", err)
		}
		if got, want := state.ProviderID, mockGitHubCom.Provider.ConfigID().ID; got != want {
			t.Fatalf("got state provider ID %v, want %v", got, want)
		}
		if got, want := state.Redirect, ""; got != want {
			t.Fatalf("got state redirect %v, want %v", got, want)
		}
	})
	t.Run("login -> github auth flow (GitHub enterprise)", func(t *testing.T) {
		resp := doRequest("GET", "http://example.com/.auth/github/login?pc="+mockGHE.Provider.ConfigID().ID, "", "", nil, false)
		if want := http.StatusFound; resp.StatusCode != want {
			t.Errorf("got response code %v, want %v", resp.StatusCode, want)
		}
		redirect := resp.Header.Get("Location")
		if got, want := redirect, "https://mycompany.com/login/oauth/authorize?"; !strings.HasPrefix(got, want) {
			t.Errorf("got redirect URL %v, want contains %v", got, want)
		}
		uredirect, err := url.Parse(redirect)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := uredirect.Query().Get("client_id"), mockGHE.Provider.CachedInfo().ClientID; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := uredirect.Query().Get("scope"), "user:email repo read:org"; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := uredirect.Query().Get("response_type"), "code"; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		state, err := oauth.DecodeState(uredirect.Query().Get("state"))
		if err != nil {
			t.Fatalf("could not decode state: %v", err)
		}
		if got, want := state.ProviderID, mockGHE.Provider.ConfigID().ID; got != want {
			t.Fatalf("got state provider ID %v, want %v", got, want)
		}
		if got, want := state.Redirect, ""; got != want {
			t.Fatalf("got state redirect %v, want %v", got, want)
		}
	})
	t.Run("login -> github auth flow with redirect param", func(t *testing.T) {
		resp := doRequest("GET", "http://example.com/.auth/github/login?pc="+mockGitHubCom.Provider.ConfigID().ID+"&redirect=%2Fpage", "", "", nil, false)
		if want := http.StatusFound; resp.StatusCode != want {
			t.Errorf("got response code %v, want %v", resp.StatusCode, want)
		}
		redirect := resp.Header.Get("Location")
		if got, want := redirect, "https://github.com/login/oauth/authorize?"; !strings.HasPrefix(got, want) {
			t.Errorf("got redirect URL %v, want contains %v", got, want)
		}
		uredirect, err := url.Parse(redirect)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := uredirect.Query().Get("client_id"), mockGitHubCom.Provider.CachedInfo().ClientID; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := uredirect.Query().Get("scope"), "user:email repo read:org"; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := uredirect.Query().Get("response_type"), "code"; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		state, err := oauth.DecodeState(uredirect.Query().Get("state"))
		if err != nil {
			t.Fatalf("could not decode state: %v", err)
		}
		if got, want := state.ProviderID, mockGitHubCom.Provider.ConfigID().ID; got != want {
			t.Fatalf("got state provider ID %v, want %v", got, want)
		}
		if got, want := state.Redirect, "/page"; got != want {
			t.Fatalf("got state redirect %v, want %v", got, want)
		}
	})
	t.Run("GitHub OAuth callback with valid state param", func(t *testing.T) {
		encodedState, err := oauth.LoginState{
			Redirect:   "/return-to-url",
			ProviderID: mockGitHubCom.Provider.ConfigID().ID,
			CSRF:       "csrf-code",
		}.Encode()
		if err != nil {
			t.Fatal(err)
		}
		resp := doRequest("GET", "http://example.com/.auth/github/callback?code=the-oauth-code&state="+encodedState, "", encodedState, nil, false)
		if want := http.StatusFound; resp.StatusCode != want {
			t.Errorf("got response code %v, want %v", resp.StatusCode, want)
		}
		if got, want := mockGitHubCom.lastCallbackRequestURL, "http://example.com/callback?code=the-oauth-code&state="+encodedState; got == nil || got.String() != want {
			t.Errorf("got last githubcom callback request url %v, want %v", got, want)
		}
		mockGitHubCom.lastCallbackRequestURL = nil
	})
	t.Run("GitHub OAuth callback with state with unknown provider", func(t *testing.T) {
		encodedState, err := oauth.LoginState{
			Redirect:   "/return-to-url",
			ProviderID: "unknown",
			CSRF:       "csrf-code",
		}.Encode()
		if err != nil {
			t.Fatal(err)
		}
		resp := doRequest("GET", "http://example.com/.auth/github/callback?code=the-oauth-code&state="+encodedState, "", encodedState, nil, false)
		if want := http.StatusBadRequest; resp.StatusCode != want {
			t.Errorf("got response code %v, want %v", resp.StatusCode, want)
		}
		if mockGitHubCom.lastCallbackRequestURL != nil {
			t.Errorf("got last github.com callback request url was non-nil: %v", mockGitHubCom.lastCallbackRequestURL)
		}
		mockGitHubCom.lastCallbackRequestURL = nil
	})
	t.Run("authenticated app request", func(t *testing.T) {
		resp := doRequest("GET", "http://example.com/", "", "", nil, true)
		if want := http.StatusOK; resp.StatusCode != want {
			t.Errorf("got response code %v, want %v", resp.StatusCode, want)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := string(body), "got through"; got != want {
			t.Errorf("got response body %v, want %v", got, want)
		}
	})
	t.Run("authenticated API request", func(t *testing.T) {
		resp := doRequest("GET", "http://example.com/.api/foo", "", "", nil, true)
		if want := http.StatusOK; resp.StatusCode != want {
			t.Errorf("got response code %v, want %v", resp.StatusCode, want)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := string(body), "got through"; got != want {
			t.Errorf("got response body %v, want %v", got, want)
		}
	})
}

type MockProvider struct {
	*oauth.Provider
	lastCallbackRequestURL *url.URL
}

func newMockProvider(t *testing.T, db database.DB, clientID, clientSecret, baseURL string) *MockProvider {
	var (
		mp       MockProvider
		problems []string
	)
	cfg := schema.AuthProviders{Github: &schema.GitHubAuthProvider{
		Url:          baseURL,
		ClientSecret: clientSecret,
		ClientID:     clientID,
		AllowOrgs:    []string{"myorg"},
	}}
	mp.Provider, problems = parseProvider(logtest.Scoped(t), cfg.Github, db, cfg)
	if len(problems) > 0 {
		t.Fatalf("Expected 0 problems, but got %d: %+v", len(problems), problems)
	}
	if mp.Provider == nil {
		t.Fatalf("Expected provider")
	}
	mp.Provider.Callback = func(oauth2.Config) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got, want := r.Method, "GET"; got != want {
				t.Errorf("In OAuth callback handler got %q request, wanted %q", got, want)
			}
			w.WriteHeader(http.StatusFound)
			mp.lastCallbackRequestURL = r.URL
		})
	}
	return &mp
}
