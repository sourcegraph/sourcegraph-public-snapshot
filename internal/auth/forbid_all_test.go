package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestForbidAllMiddleware(t *testing.T) {
	handler := ForbidAllRequestsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello")
	}))

	t.Run("disabled", func(t *testing.T) {
		conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{AuthProviders: []schema.AuthProviders{{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}}}}})
		defer conf.Mock(nil)

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
		conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{}})
		defer conf.Mock(nil)

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
