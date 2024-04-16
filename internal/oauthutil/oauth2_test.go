package oauthutil

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

func TestDoRequest(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)

	authToken := "original-auth"
	unauthedToken := "unauthed-token"
	refreshToken := "refresh-token"
	refreshedAuthToken := "refreshed-auth"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Write([]byte("unauthed"))
			return
		}

		if strings.HasSuffix(authHeader, unauthedToken) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if strings.HasPrefix(authHeader, "Bearer ") {
			w.Write([]byte(fmt.Sprintf("authed %s", strings.TrimPrefix(authHeader, "Bearer "))))
			return
		}
	}))

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		wantBody     string
		authToken    string
		refreshToken string
		expiresIn    int // minutes til expiry
	}{
		{
			name:     "unauthed request",
			wantBody: "unauthed",
		},
		{
			name:      "authed request no refresher",
			wantBody:  fmt.Sprintf("authed %s", authToken),
			authToken: authToken,
		},
		{
			name:         "expired auth with refresher",
			wantBody:     fmt.Sprintf("authed %s", refreshedAuthToken),
			authToken:    authToken,
			refreshToken: refreshToken,
			expiresIn:    -20, // Expired 20 minutes ago
		},
		{
			name:         "not expired but unauthed with refresher",
			wantBody:     fmt.Sprintf("authed %s", refreshedAuthToken),
			authToken:    unauthedToken,
			refreshToken: refreshToken,
			expiresIn:    20, // Expires in 20 minutes
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var auther auth.Authenticator

			if test.authToken != "" {
				auther = &auth.OAuthBearerToken{
					Token:        test.authToken,
					RefreshToken: test.refreshToken,
					Expiry:       time.Now().Add(time.Duration(test.expiresIn) * time.Minute),
					RefreshFunc: func(_ context.Context, _ httpcli.Doer, _ *auth.OAuthBearerToken) (string, string, time.Time, error) {
						return refreshedAuthToken, "", time.Time{}, nil
					},
				}
			}

			resp, err := DoRequest(ctx, logger, http.DefaultClient, req, auther)
			if err != nil {
				t.Fatal(err)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if string(body) != test.wantBody {
				t.Fatalf("expected %q, got %q", test.wantBody, string(body))
			}
		})
	}
}
