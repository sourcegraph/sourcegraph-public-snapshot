package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/actor"
)

// unexpiredCookies returns the list of unexpired cookies set by the response
func unexpiredCookies(resp *http.Response) (cookies []*http.Cookie) {
	for _, cookie := range resp.Cookies() {
		if cookie.RawExpires == "" || cookie.Expires.After(time.Now()) {
			cookies = append(cookies, cookie)
		}
	}
	return
}

func requireAuthenticatedActor(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !actor.FromContext(r.Context()).IsAuthenticated() {
			t.Errorf("unauthenticated actor requested %s", r.URL)
		}
	})
}
