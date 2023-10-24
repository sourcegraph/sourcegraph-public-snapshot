package httpheader

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	sgactor "github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// SEE ALSO FOR MANUAL TESTING: See the Middleware docstring for information about the testproxy
// helper program, which helps with manual testing of the HTTP auth proxy behavior.
func TestMiddleware(t *testing.T) {
	defer licensing.TestingSkipFeatureChecks()()

	logger := logtest.Scoped(t)

	db := database.NewDB(logger, dbtest.NewDB(t))

	handler := middleware(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor := sgactor.FromContext(r.Context())
		if actor.IsAuthenticated() {
			fmt.Fprintf(w, "user %v", actor.UID)
		} else {
			fmt.Fprint(w, "no user")
		}
	}))

	const headerName = "x-sso-user-header"
	const emailHeaderName = "x-sso-email-header"
	providers.Update(pkgName, []providers.Provider{
		&provider{
			c: &schema.HTTPHeaderAuthProvider{
				EmailHeader:    emailHeaderName,
				UsernameHeader: headerName,
			},
		},
	})
	defer func() { providers.Update(pkgName, nil) }()

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
		req = req.WithContext(sgactor.WithActor(context.Background(), &sgactor.Actor{UID: 123}))
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
		auth.MockGetAndSaveUser = func(ctx context.Context, op auth.GetAndSaveUserOp) (newUserCreated bool, userID int32, safeErrMsg string, err error) {
			calledMock = true
			if op.ExternalAccount.ServiceType == "http-header" && op.ExternalAccount.ServiceID == "" && op.ExternalAccount.ClientID == "" && op.ExternalAccount.AccountID == "alice" {
				return false, 1, "", nil
			}
			return false, 0, "safeErr", errors.Errorf("account %v not found in mock", op.ExternalAccount)
		}
		defer func() { auth.MockGetAndSaveUser = nil }()
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
		req = req.WithContext(sgactor.WithActor(context.Background(), &sgactor.Actor{UID: 123}))
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "user 123"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("sent, with un-normalized username", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(headerName, "alice%zhao")
		const wantNormalizedUsername = "alice-zhao"
		var calledMock bool
		auth.MockGetAndSaveUser = func(ctx context.Context, op auth.GetAndSaveUserOp) (newUserCreated bool, userID int32, safeErrMsg string, err error) {
			calledMock = true
			if op.UserProps.Username != wantNormalizedUsername {
				t.Errorf("got %q, want %q", op.UserProps.Username, wantNormalizedUsername)
			}
			if op.ExternalAccount.ServiceType == "http-header" && op.ExternalAccount.ServiceID == "" && op.ExternalAccount.ClientID == "" && op.ExternalAccount.AccountID == "alice%zhao" {
				return false, 1, "", nil
			}
			return false, 0, "safeErr", errors.Errorf("account %v not found in mock", op.ExternalAccount)
		}
		defer func() { auth.MockGetAndSaveUser = nil }()
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "user 1"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
		if !calledMock {
			t.Error("!calledMock")
		}
	})

	t.Run("sent, email", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(emailHeaderName, "alice@example.com")
		var calledMock bool
		auth.MockGetAndSaveUser = func(ctx context.Context, op auth.GetAndSaveUserOp) (newUserCreated bool, userID int32, safeErrMsg string, err error) {
			calledMock = true
			if got, want := op.UserProps.Username, "alice"; got != want {
				t.Errorf("expected %v got %v", want, got)
			}
			if got, want := op.UserProps.Email, "alice@example.com"; got != want {
				t.Errorf("expected %v got %v", want, got)
			}
			if got, want := op.UserProps.EmailIsVerified, true; got != want {
				t.Errorf("expected %v got %v", want, got)
			}
			if op.ExternalAccount.ServiceType == "http-header" && op.ExternalAccount.ServiceID == "" && op.ExternalAccount.ClientID == "" && op.ExternalAccount.AccountID == "alice@example.com" {
				return false, 1, "", nil
			}
			t.Log(op.ExternalAccount)
			return false, 0, "safeErr", errors.Errorf("account %v not found in mock", op.ExternalAccount)
		}
		defer func() { auth.MockGetAndSaveUser = nil }()
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "user 1"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
		if !calledMock {
			t.Error("!calledMock")
		}
	})

	t.Run("sent, email & username", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(emailHeaderName, "alice@example.com")
		req.Header.Set(headerName, "bob")
		var calledMock bool
		auth.MockGetAndSaveUser = func(ctx context.Context, op auth.GetAndSaveUserOp) (newUserCreated bool, userID int32, safeErrMsg string, err error) {
			calledMock = true
			if got, want := op.UserProps.Username, "bob"; got != want {
				t.Errorf("expected %v got %v", want, got)
			}
			if got, want := op.UserProps.Email, "alice@example.com"; got != want {
				t.Errorf("expected %v got %v", want, got)
			}
			if got, want := op.UserProps.EmailIsVerified, true; got != want {
				t.Errorf("expected %v got %v", want, got)
			}
			if op.ExternalAccount.ServiceType == "http-header" && op.ExternalAccount.ServiceID == "" && op.ExternalAccount.ClientID == "" && op.ExternalAccount.AccountID == "alice@example.com" {
				return false, 1, "", nil
			}
			return false, 0, "safeErr", errors.Errorf("account %v not found in mock", op.ExternalAccount)
		}
		defer func() { auth.MockGetAndSaveUser = nil }()
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "user 1"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
		if !calledMock {
			t.Error("!calledMock")
		}
	})
}

func TestMiddleware_stripPrefix(t *testing.T) {
	defer licensing.TestingSkipFeatureChecks()()

	logger := logtest.Scoped(t)

	db := database.NewDB(logger, dbtest.NewDB(t))

	handler := middleware(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor := sgactor.FromContext(r.Context())
		if actor.IsAuthenticated() {
			fmt.Fprintf(w, "user %v", actor.UID)
		} else {
			fmt.Fprint(w, "no user")
		}
	}))

	const headerName = "x-sso-user-header"
	providers.Update(pkgName, []providers.Provider{
		&provider{
			c: &schema.HTTPHeaderAuthProvider{
				UsernameHeader:            headerName,
				StripUsernameHeaderPrefix: "accounts.google.com:",
			},
		},
	})
	defer func() { providers.Update(pkgName, nil) }()

	t.Run("sent, user", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(headerName, "accounts.google.com:alice")
		var calledMock bool
		auth.MockGetAndSaveUser = func(ctx context.Context, op auth.GetAndSaveUserOp) (newUserCreated bool, userID int32, safeErrMsg string, err error) {
			calledMock = true
			if op.ExternalAccount.ServiceType == "http-header" && op.ExternalAccount.ServiceID == "" && op.ExternalAccount.ClientID == "" && op.ExternalAccount.AccountID == "alice" {
				return false, 1, "", nil
			}
			return false, 0, "safeErr", errors.Errorf("account %v not found in mock", op.ExternalAccount)
		}
		defer func() { auth.MockGetAndSaveUser = nil }()
		handler.ServeHTTP(rr, req)
		if got, want := rr.Body.String(), "user 1"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
		if !calledMock {
			t.Error("!calledMock")
		}
	})
}
