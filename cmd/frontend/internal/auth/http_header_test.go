package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
)

// SEE ALSO FOR MANUAL TESTING: See the httpHeaderAuthMiddleware docstring for information about the
// testproxy helper program, which helps with manual testing of the HTTP auth proxy behavior.
func Test_httpHeaderAuthMiddleware(t *testing.T) {
	handler := httpHeaderAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor := actor.FromContext(r.Context())
		if actor.IsAuthenticated() {
			fmt.Fprintf(w, "user %v", actor.UID)
		} else {
			fmt.Fprint(w, "no user")
		}
	}))

	ssoUserHeader = "x-sso-user-header"
	defer func() { ssoUserHeader = "" }()

	t.Run("not sent", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "must access via HTTP authentication proxy\n"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("not sent, actor present", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req = req.WithContext(actor.WithActor(context.Background(), &actor.Actor{UID: 123}))
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "must access via HTTP authentication proxy\n"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("sent, new user", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(ssoUserHeader, "alice")
		var calledGetByUsername, calledCreate bool
		db.Mocks.Users.GetByExternalID = func(ctx context.Context, provider, id string) (*types.User, error) {
			if want := "http-header:alice"; id != want {
				t.Errorf("got %q, want %q", id, want)
			}
			calledGetByUsername = true
			return nil, &errcode.Mock{Message: "user not found", IsNotFound: true}
		}
		db.Mocks.Users.Create = func(ctx context.Context, info db.NewUser) (*types.User, error) {
			calledCreate = true
			return &types.User{ID: 1, ExternalID: &info.ExternalID, Username: info.Username}, nil
		}
		defer func() { db.Mocks = db.MockStores{} }()
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "user 1"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
		if !calledGetByUsername {
			t.Error("!calledGetByUsername")
		}
		if !calledCreate {
			t.Error("!calledCreate")
		}
	})

	t.Run("sent, actor already set", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(ssoUserHeader, "alice")
		req = req.WithContext(actor.WithActor(context.Background(), &actor.Actor{UID: 123}))
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "user 123"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("sent, new user with un-normalized username", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(ssoUserHeader, "alice.zhao")
		const wantNormalizedUsername = "alice-zhao"
		var calledGetByUsername, calledCreate bool
		db.Mocks.Users.GetByExternalID = func(ctx context.Context, provider, id string) (*types.User, error) {
			if want := "http-header:alice.zhao"; /* pre-normalized */ id != want {
				t.Errorf("got %q, want %q", id, want)
			}
			calledGetByUsername = true
			return nil, &errcode.Mock{Message: "user not found", IsNotFound: true}
		}
		db.Mocks.Users.Create = func(ctx context.Context, info db.NewUser) (*types.User, error) {
			if info.Username != wantNormalizedUsername {
				t.Errorf("got %q, want %q", info.Username, wantNormalizedUsername)
			}
			calledCreate = true
			return &types.User{ID: 1, ExternalID: &info.ExternalID, Username: info.Username}, nil
		}
		defer func() { db.Mocks = db.MockStores{} }()
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "user 1"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
		if !calledGetByUsername {
			t.Error("!calledGetByUsername")
		}
		if !calledCreate {
			t.Error("!calledCreate")
		}
	})

	t.Run("sent, existing user", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(ssoUserHeader, "bob")
		var calledGetByUsername bool
		db.Mocks.Users.GetByExternalID = func(ctx context.Context, provider, id string) (*types.User, error) {
			if want := "http-header:bob"; id != want {
				t.Errorf("got %q, want %q", id, want)
			}
			calledGetByUsername = true
			return &types.User{ID: 1, ExternalID: &id, Username: "bob"}, nil
		}
		defer func() { db.Mocks = db.MockStores{} }()
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "user 1"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
		if !calledGetByUsername {
			t.Error("!calledGetByUsername")
		}
	})
}
