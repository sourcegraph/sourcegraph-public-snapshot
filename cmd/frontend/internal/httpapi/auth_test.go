package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
)

func TestAccessTokenAuthMiddleware(t *testing.T) {
	handler := AccessTokenAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor := actor.FromContext(r.Context())
		if actor.IsAuthenticated() {
			fmt.Fprintf(w, "user %v", actor.UID)
		} else {
			fmt.Fprint(w, "no user")
		}
	}))
	checkHTTPResponse := func(t *testing.T, req *http.Request, wantStatusCode int, wantBody string) {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != wantStatusCode {
			t.Errorf("got response status %d, want %d", rr.Code, wantStatusCode)
		}
		if got := rr.Body.String(); got != wantBody {
			t.Errorf("got response body %q, want %q", got, wantBody)
		}
	}

	t.Run("no header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		checkHTTPResponse(t, req, http.StatusOK, "no user")
	})

	// Test that the absence of an Authorization header doesn't unset the actor provided by a prior
	// auth middleware.
	t.Run("no header, actor present", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req = req.WithContext(actor.WithActor(context.Background(), &actor.Actor{UID: 123}))
		checkHTTPResponse(t, req, http.StatusOK, "user 123")
	})

	for _, invalidHeaderValue := range []string{"x", "x y", "token "} {
		t.Run("invalid header "+invalidHeaderValue, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", invalidHeaderValue)
			checkHTTPResponse(t, req, http.StatusUnauthorized, "Invalid Authorization header.\n")
		})
	}

	t.Run("valid header with invalid token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "token badbad")
		var calledAccessTokensLookup bool
		db.Mocks.AccessTokens.Lookup = func(tokenHexEncoded, requiredScope string) (subjectUserID int32, err error) {
			calledAccessTokensLookup = true
			return 0, errors.New("x")
		}
		defer func() { db.Mocks = db.MockStores{} }()
		checkHTTPResponse(t, req, http.StatusUnauthorized, "Invalid access token.\n")
		if !calledAccessTokensLookup {
			t.Error("!calledAccessTokensLookup")
		}
	})

	for _, headerValue := range []string{"token abcdef", `token token="abcdef"`} {
		t.Run("valid token: "+headerValue, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", headerValue)
			var calledAccessTokensLookup bool
			db.Mocks.AccessTokens.Lookup = func(tokenHexEncoded, requiredScope string) (subjectUserID int32, err error) {
				if want := "abcdef"; tokenHexEncoded != want {
					t.Errorf("got %q, want %q", tokenHexEncoded, want)
				}
				if want := authz.ScopeUserAll; requiredScope != want {
					t.Errorf("got %q, want %q", requiredScope, want)
				}
				calledAccessTokensLookup = true
				return 123, nil
			}
			defer func() { db.Mocks = db.MockStores{} }()
			checkHTTPResponse(t, req, http.StatusOK, "user 123")
			if !calledAccessTokensLookup {
				t.Error("!calledAccessTokensLookup")
			}
		})
	}

	// Test that an access token overwrites the actor set by a prior auth middleware.
	t.Run("actor present, valid token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "token abcdef")
		req = req.WithContext(actor.WithActor(context.Background(), &actor.Actor{UID: 456}))
		var calledAccessTokensLookup bool
		db.Mocks.AccessTokens.Lookup = func(tokenHexEncoded, requiredScope string) (subjectUserID int32, err error) {
			if want := "abcdef"; tokenHexEncoded != want {
				t.Errorf("got %q, want %q", tokenHexEncoded, want)
			}
			if want := authz.ScopeUserAll; requiredScope != want {
				t.Errorf("got %q, want %q", requiredScope, want)
			}
			calledAccessTokensLookup = true
			return 123, nil
		}
		defer func() { db.Mocks = db.MockStores{} }()
		checkHTTPResponse(t, req, http.StatusOK, "user 123")
		if !calledAccessTokensLookup {
			t.Error("!calledAccessTokensLookup")
		}
	})
}
