package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestForbidAllMiddleware(t *testing.T) {
	handler := ForbidAllAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello")
	}))

	t.Run("disabled", func(t *testing.T) {
		conf.MockGetData = &schema.SiteConfiguration{AuthProviders: []schema.AuthProviders{{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}}}}
		defer func() { conf.MockGetData = nil }()

		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		handler.ServeHTTP(rr, req)
		if want := http.StatusOK; rr.Code != want {
			t.Errorf("got %d, want %d", rr.Code, want)
		}
		if got, want := rr.Body.String(), "hello"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("enabled", func(t *testing.T) {
		conf.MockGetData = &schema.SiteConfiguration{}
		defer func() { conf.MockGetData = nil }()

		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		handler.ServeHTTP(rr, req)
		if want := http.StatusForbidden; rr.Code != want {
			t.Errorf("got %d, want %d", rr.Code, want)
		}
		if got, want := rr.Body.String(), "Access to Sourcegraph is forbidden"; !strings.Contains(got, want) {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
