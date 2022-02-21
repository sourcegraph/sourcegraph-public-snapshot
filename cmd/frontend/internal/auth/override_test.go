package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestOverrideAuthMiddleware(t *testing.T) {
	cleanup := session.ResetMockSessionStore(t)
	defer cleanup()

	newHandler := func(db database.DB) http.Handler {
		return OverrideAuthMiddleware(db, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor := actor.FromContext(r.Context())
			if actor.IsAuthenticated() {
				fmt.Fprintf(w, "user %v", actor.UID)
			} else {
				fmt.Fprint(w, "no user")
			}
		}))
	}

	const overrideSecret = "s"

	t.Run("disabled, not sent", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		handler := newHandler(database.NewMockDB())
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
		handler := newHandler(database.NewMockDB())
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
		handler := newHandler(database.NewMockDB())
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "user 2"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("sent, actor not set", func(t *testing.T) {
		envOverrideAuthSecret = overrideSecret
		defer func() { envOverrideAuthSecret = "" }()
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(overrideSecretHeader, overrideSecret)
		var calledMock bool
		auth.MockGetAndSaveUser = func(ctx context.Context, op auth.GetAndSaveUserOp) (userID int32, safeErrMsg string, err error) {
			calledMock = true
			if want := defaultUsername; op.UserProps.Username != want {
				t.Errorf("got %q, want %q", op.UserProps.Username, want)
			}
			return 1, "", nil
		}
		defer func() { auth.MockGetAndSaveUser = nil }()

		users := database.NewMockUserStore()
		users.SetIsSiteAdminFunc.SetDefaultReturn(nil)
		users.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int32) (*types.User, error) {
			return &types.User{ID: id, CreatedAt: time.Now()}, nil
		})
		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)

		handler := newHandler(db)
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "user 1"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
		if !calledMock {
			t.Error("!calledMock")
		}
	})

	t.Run("sent, actor already set", func(t *testing.T) {
		envOverrideAuthSecret = overrideSecret
		defer func() { envOverrideAuthSecret = "" }()
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(overrideSecretHeader, overrideSecret)
		req = req.WithContext(actor.WithActor(context.Background(), &actor.Actor{UID: 123}))
		var calledMock bool
		auth.MockGetAndSaveUser = func(ctx context.Context, op auth.GetAndSaveUserOp) (userID int32, safeErrMsg string, err error) {
			calledMock = true
			if op.ExternalAccount.ServiceType == "override" && op.ExternalAccount.AccountID == defaultUsername {
				return 1, "", nil
			}
			return 0, "safeErr", errors.New("x")
		}
		defer func() { auth.MockGetAndSaveUser = nil }()

		users := database.NewMockUserStore()
		users.SetIsSiteAdminFunc.SetDefaultReturn(nil)
		users.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int32) (*types.User, error) {
			return &types.User{ID: id, CreatedAt: time.Now()}, nil
		})
		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)

		handler := newHandler(db)
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "user 1"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
		if !calledMock {
			t.Error("!calledMock")
		}
	})

	t.Run("sent, wrong secret", func(t *testing.T) {
		envOverrideAuthSecret = overrideSecret
		defer func() { envOverrideAuthSecret = "" }()
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(overrideSecretHeader, "bad")
		handler := newHandler(database.NewMockDB())
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "no user"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
