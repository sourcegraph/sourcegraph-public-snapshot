package httpheader

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

// SEE ALSO FOR MANUAL TESTING: See the Middleware docstring for information about the testproxy
// helper program, which helps with manual testing of the HTTP auth proxy behavior.
func TestMiddleware(t *testing.T) {
	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor := actor.FromContext(r.Context())
		if actor.IsAuthenticated() {
			fmt.Fprintf(w, "user %v", actor.UID)
		} else {
			fmt.Fprint(w, "no user")
		}
	}))

	const headerName = "x-sso-user-header"
	conf.Mock(&schema.SiteConfiguration{AuthProvider: "http-header", AuthUserIdentityHTTPHeader: headerName})
	defer conf.Mock(nil)

	t.Run("not sent", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "no user"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("not sent, actor present", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req = req.WithContext(actor.WithActor(context.Background(), &actor.Actor{UID: 123}))
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "user 123"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("sent, user", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(headerName, "alice")
		var calledMock bool
		auth.MockCreateOrUpdateUser = func(u db.NewUser, a db.ExternalAccountSpec) (userID int32, err error) {
			calledMock = true
			if a.ServiceType == "http-header" && a.ServiceID == "" && a.ClientID == "" && a.AccountID == "alice" {
				return 1, nil
			}
			return 0, fmt.Errorf("account %v not found in mock", a)
		}
		defer func() { auth.MockCreateOrUpdateUser = nil }()
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "user 1"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
		if !calledMock {
			t.Error("!calledMock")
		}
	})

	t.Run("sent, actor already set", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(headerName, "alice")
		req = req.WithContext(actor.WithActor(context.Background(), &actor.Actor{UID: 123}))
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "user 123"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("sent, with un-normalized username", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(headerName, "alice.zhao")
		const wantNormalizedUsername = "alice-zhao"
		var calledMock bool
		auth.MockCreateOrUpdateUser = func(u db.NewUser, a db.ExternalAccountSpec) (userID int32, err error) {
			calledMock = true
			if u.Username != wantNormalizedUsername {
				t.Errorf("got %q, want %q", u.Username, wantNormalizedUsername)
			}
			if a.ServiceType == "http-header" && a.ServiceID == "" && a.ClientID == "" && a.AccountID == "alice.zhao" {
				return 1, nil
			}
			return 0, fmt.Errorf("account %v not found in mock", a)
		}
		defer func() { auth.MockCreateOrUpdateUser = nil }()
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "user 1"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
		if !calledMock {
			t.Error("!calledMock")
		}
	})
}
