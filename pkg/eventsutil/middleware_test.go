package eventsutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httputil/httpctx"
)

func TestAgentMiddleware(t *testing.T) {
	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpctx.SetForRequest(r, context.Background())
		AgentMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			ctx := httpctx.FromRequest(r)

			userAgent := UserAgentFromContext(ctx)
			if want := "sourcegraphbot"; userAgent != want {
				t.Errorf("got User-Agent %q, want %q", userAgent, want)
			}
		})).ServeHTTP(w, r)
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("User-Agent", "sourcegraphbot")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if !called {
		t.Error("!called")
	}
}
