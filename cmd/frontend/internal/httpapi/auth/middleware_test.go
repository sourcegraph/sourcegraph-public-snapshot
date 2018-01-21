package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
)

func init() {
	db.InitTest("frontend-internal-httpapi-auth")
}

func TestAuthorizationMiddleware(t *testing.T) {
	handler := AuthorizationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor := actor.FromContext(r.Context())
		if actor.IsAuthenticated() {
			fmt.Fprintf(w, "user %v", actor.UID)
		} else {
			fmt.Fprint(w, "no user")
		}
	}))

	t.Run("SSO HTTP header", func(t *testing.T) {
		ssoUserHeader = "x-sso-user-header"
		defer func() { ssoUserHeader = "" }()
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		t.Run("not sent", func(t *testing.T) {
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if got, want := rr.Body.String(), "no user"; got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})

		t.Run("sent, new user", func(t *testing.T) {
			rr := httptest.NewRecorder()
			req.Header.Set(ssoUserHeader, "alice")
			var calledGetByUsername, calledCreate bool
			db.Mocks.Users.GetByUsername = func(ctx context.Context, username string) (*types.User, error) {
				calledGetByUsername = true
				return nil, db.ErrUserNotFound{}
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

		t.Run("sent, new user (integration)", func(t *testing.T) {
			if testing.Short() {
				t.Skip()
			}
			rr := httptest.NewRecorder()
			req.Header.Set(ssoUserHeader, "alice")
			handler.ServeHTTP(rr, req)
			if got, want := rr.Body.String(), "user 1"; got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})

		t.Run("sent, existing user", func(t *testing.T) {
			rr := httptest.NewRecorder()
			req.Header.Set(ssoUserHeader, "bob")
			var calledGetByUsername bool
			db.Mocks.Users.GetByUsername = func(ctx context.Context, username string) (*types.User, error) {
				calledGetByUsername = true
				extID := "http-header:" + username
				return &types.User{ID: 1, ExternalID: &extID, Username: username}, nil
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

		t.Run("sent, existing user (integration)", func(t *testing.T) {
			rr := httptest.NewRecorder()
			req.Header.Set(ssoUserHeader, "bob")
			handler.ServeHTTP(rr, req)
			if got, want := rr.Body.String(), "user 2"; got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	})
}
