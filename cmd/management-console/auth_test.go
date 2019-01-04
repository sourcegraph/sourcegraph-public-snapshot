package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/db/globalstatedb"
)

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name                                      string
		setHeaders                                func(r *http.Request)
		wantAuthenticated                         bool
		authenticateManagementConsoleMockPassword string
	}{
		{
			name:              "no_authorization_header",
			setHeaders:        nil,
			wantAuthenticated: false,
		},
		{
			name: "invalid_authorization_header",
			setHeaders: func(r *http.Request) {
				r.Header.Set("Authorization", "9lives")
			},
			wantAuthenticated: false,
		},
		{
			name: "valid_authorization_header_invalid_credentials",
			setHeaders: func(r *http.Request) {
				r.SetBasicAuth("foo", "bar")
			},
			authenticateManagementConsoleMockPassword: "123",
			wantAuthenticated:                         false,
		},
		{
			name: "valid_authorization_header",
			setHeaders: func(r *http.Request) {
				r.SetBasicAuth("anything123", "baz")
			},
			authenticateManagementConsoleMockPassword: "baz",
			wantAuthenticated:                         true,
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			called := false
			h := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
			}))

			if tst.authenticateManagementConsoleMockPassword != "" {
				passwordCheckCalled := false
				globalstatedb.Mock.AuthenticateManagementConsole = func(ctx context.Context, password string) error {
					passwordCheckCalled = true
					if password != tst.authenticateManagementConsoleMockPassword {
						if tst.wantAuthenticated {
							t.Error("invalid password")
						}
						return errors.New("invalid password")
					}
					return nil
				}
				defer func() {
					globalstatedb.Mock.AuthenticateManagementConsole = nil
					if !passwordCheckCalled {
						t.Error("expected password check to be called")
					}
				}()
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)
			if tst.setHeaders != nil {
				tst.setHeaders(req)
			}
			h.ServeHTTP(rec, req)
			res := rec.Result()
			if tst.wantAuthenticated {
				if !called {
					t.Fatal("authenticated requests should pass through")
				}
				if res.StatusCode != http.StatusOK {
					t.Fatalf("expected 200, got %v", res.StatusCode)
				}
				return
			}

			if called {
				t.Fatal("unauthenticated requests should NEVER pass through")
			}
			if res.StatusCode != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %v", res.StatusCode)
			}
		})
	}
}
