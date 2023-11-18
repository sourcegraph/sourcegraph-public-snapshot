package confauth

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/session"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
)

func TestMiddleware(t *testing.T) {
	cleanup := session.ResetMockSessionStore(t)
	defer cleanup()

	value := false
	ok := false
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value, ok = r.Context().Value(auth.AllowAnonymousRequestContextKey).(bool)
	})
	handler := http.NewServeMux()
	handler.Handle("/.api/", Middleware().API(h))
	handler.Handle("/", Middleware().API(h))

	doRequest := func(method, urlStr, body string) *http.Response {
		req := httptest.NewRequest(method, urlStr, bytes.NewBufferString(body))
		respRecorder := httptest.NewRecorder()
		handler.ServeHTTP(respRecorder, req)
		return respRecorder.Result()
	}

	expiresAt := time.Now().Add(time.Hour)

	tests := []struct {
		name      string
		license   *license.Info
		wantOk    bool
		wantValue bool
	}{
		{
			name:      "no license",
			license:   nil,
			wantOk:    false,
			wantValue: false,
		},
		{
			name:      "with license, no special tag",
			license:   &license.Info{UserCount: 10, ExpiresAt: expiresAt},
			wantOk:    true,
			wantValue: false,
		},
		{
			name:      "with license, with special tag",
			license:   &license.Info{Tags: []string{licensing.AllowAnonymousUsageTag}, UserCount: 10, ExpiresAt: expiresAt},
			wantOk:    true,
			wantValue: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
				return test.license, "test-signature", nil
			}
			defer func() { licensing.MockGetConfiguredProductLicenseInfo = nil }()

			resp := doRequest("GET", "/", "")
			if want := http.StatusOK; resp.StatusCode != want {
				t.Errorf("got response code %v, want %v", resp.StatusCode, want)
			}
			require.Equal(t, test.wantOk, ok)
			require.Equal(t, test.wantValue, value)
		})
	}
}
