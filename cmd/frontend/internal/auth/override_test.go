package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
)

func TestOverrideAuthMiddleware(t *testing.T) {
	cleanup := session.ResetMockSessionStore(t)
	defer cleanup()

	handler := OverrideAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor := actor.FromContext(r.Context())
		if actor.IsAuthenticated() {
			fmt.Fprintf(w, "user %v", actor.UID)
		} else {
			fmt.Fprint(w, "no user")
		}
	}))

	const overrideSecret = "s"

	t.Run("disabled, not sent", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "no user"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("not sent", func(t *testing.T) {
		envOverrideAuthSecret = overrideSecret
		defer func() { envOverrideAuthSecret = "" }()
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "no user"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("not sent, actor present", func(t *testing.T) {
		envOverrideAuthSecret = overrideSecret
		defer func() { envOverrideAuthSecret = "" }()
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req = req.WithContext(actor.WithActor(context.Background(), &actor.Actor{UID: 2}))
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "user 2"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("sent, new user", func(t *testing.T) {
		envOverrideAuthSecret = overrideSecret
		defer func() { envOverrideAuthSecret = "" }()
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(overrideHeader, overrideSecret)
		var calledGetByExternalID, calledCreate bool
		db.Mocks.Users.GetByExternalID = func(ctx context.Context, provider, id string) (*types.User, error) {
			calledGetByExternalID = true
			return nil, &errcode.Mock{IsNotFound: true}
		}
		db.Mocks.Users.Create = func(ctx context.Context, info db.NewUser) (*types.User, error) {
			if want := "anon-user"; info.Username != want {
				t.Errorf("got %q, want %q", info.Username, want)
			}
			calledCreate = true
			return &types.User{ID: 1, ExternalID: &info.ExternalID, Username: info.Username}, nil
		}
		defer func() { db.Mocks = db.MockStores{} }()
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "user 1"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
		if !calledGetByExternalID {
			t.Error("!calledGetByExternalID")
		}
		if !calledCreate {
			t.Error("!calledCreate")
		}
	})

	t.Run("sent, actor already set", func(t *testing.T) {
		envOverrideAuthSecret = overrideSecret
		defer func() { envOverrideAuthSecret = "" }()
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(overrideHeader, overrideSecret)
		req = req.WithContext(actor.WithActor(context.Background(), &actor.Actor{UID: 123}))
		var calledGetByExternalID bool
		db.Mocks.Users.GetByExternalID = func(ctx context.Context, provider, id string) (*types.User, error) {
			calledGetByExternalID = true
			if provider == "override" && id == "anon-user" {
				return &types.User{ID: 1, ExternalID: &id, Username: "anon-user"}, nil
			}
			return nil, &errcode.Mock{IsNotFound: true}
		}
		defer func() { db.Mocks = db.MockStores{} }()
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "user 1"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
		if !calledGetByExternalID {
			t.Error("!calledGetByExternalID")
		}
	})

	t.Run("sent, existing user", func(t *testing.T) {
		envOverrideAuthSecret = overrideSecret
		defer func() { envOverrideAuthSecret = "" }()
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(overrideHeader, overrideSecret)
		var calledGetByExternalID bool
		db.Mocks.Users.GetByExternalID = func(ctx context.Context, provider, id string) (*types.User, error) {
			calledGetByExternalID = true
			if provider == "override" && id == "anon-user" {
				return &types.User{ID: 1, ExternalID: &id, Username: "anon-user"}, nil
			}
			return nil, errors.New("x")
		}
		defer func() { db.Mocks = db.MockStores{} }()
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "user 1"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
		if !calledGetByExternalID {
			t.Error("!calledGetByExternalID")
		}
	})

	t.Run("sent, wrong secret", func(t *testing.T) {
		envOverrideAuthSecret = overrideSecret
		defer func() { envOverrideAuthSecret = "" }()
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(overrideHeader, "bad")
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "no user"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
