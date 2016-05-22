package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
)

func TestCookieMiddleware(t *testing.T) {
	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpctx.SetForRequest(r, context.Background())
		CookieMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			ctx := httpctx.FromRequest(r)

			creds := sourcegraph.CredentialsFromContext(ctx)
			tok, err := creds.(oauth2.TokenSource).Token()
			if err != nil {
				t.Fatal(err)
			}
			if want := "mytoken"; tok.AccessToken != want {
				t.Errorf("got token %q, want %q", tok.AccessToken, want)
			}
		})).ServeHTTP(w, r)
	}))
	defer server.Close()

	sc, err := NewSessionCookie(Session{AccessToken: "mytoken"})
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("GET", server.URL, nil)
	req.AddCookie(sc)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if !called {
		t.Error("!called")
	}
}
