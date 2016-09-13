package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func TestCookieMiddleware(t *testing.T) {
	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		CookieMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			ctx := r.Context()

			tok := sourcegraph.AccessTokenFromContext(ctx)
			if want := "mytoken"; tok != want {
				t.Errorf("got token %q, want %q", tok, want)
			}
		})).ServeHTTP(w, r)
	}))
	defer server.Close()

	sc, err := NewSessionCookie(Session{AccessToken: "mytoken"}, false)
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
