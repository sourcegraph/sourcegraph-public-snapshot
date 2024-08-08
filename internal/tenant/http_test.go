package tenant

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func TestInternalHTTPTransport(t *testing.T) {
	tests := []struct {
		name        string
		tenant      *Tenant
		wantHeaders map[string]string
	}{
		{
			name: "unauthenticated",
			wantHeaders: map[string]string{
				headerKeyTenantID: headerValueNoTenant,
			},
		},
		{
			name:   "with tenant",
			tenant: &Tenant{_id: 1234},
			wantHeaders: map[string]string{
				headerKeyTenantID: "1234",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &InternalHTTPTransport{
				RoundTripper: roundTripFunc(func(req *http.Request) *http.Response {
					for k, want := range tt.wantHeaders {
						if got := req.Header.Get(k); got == "" {
							t.Errorf("did not find expected header %q", k)
						} else if diff := cmp.Diff(want, got); diff != "" {
							t.Errorf("headers mismatch (-want +got):\n%s", diff)
						}
					}
					return &http.Response{StatusCode: http.StatusOK}
				}),
			}
			ctx := context.Background()
			if tt.tenant != nil {
				ctx = withTenant(ctx, tt.tenant.ID())
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/test", nil)
			if err != nil {
				t.Fatal(err)
			}
			got, err := transport.RoundTrip(req)
			if err != nil {
				t.Fatalf("Transport.RoundTrip() error = %v", err)
			}
			if got.StatusCode != http.StatusOK {
				t.Fatalf("Unexpected response: %+v", got)
			}
		})
	}
}

func TestInternalHTTPMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		headers        map[string]string
		wantTenant     *Tenant
		wantStatusCode int
	}{{
		name: "unauthenticated",
		headers: map[string]string{
			headerKeyTenantID: headerValueNoTenant,
		},
		wantTenant:     nil,
		wantStatusCode: http.StatusOK,
	}, {
		name: "invalid tenant",
		headers: map[string]string{
			headerKeyTenantID: "not-a-valid-id",
		},
		wantTenant:     nil,
		wantStatusCode: http.StatusForbidden,
	}, {
		name: "with tenant",
		headers: map[string]string{
			headerKeyTenantID: "1234",
		},
		wantTenant:     &Tenant{_id: 1234},
		wantStatusCode: http.StatusOK,
	}, {
		name: "empty tenant",
		headers: map[string]string{
			headerKeyTenantID: "",
		},
		wantTenant:     nil,
		wantStatusCode: http.StatusOK,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := InternalHTTPMiddleware(logtest.Scoped(t), http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				got, err := FromContext(r.Context())
				if tt.wantTenant == nil {
					require.Error(t, err)
				} else {
					require.Equal(t, tt.wantTenant, got)
				}
			}))
			req, err := http.NewRequest(http.MethodGet, "/test", nil)
			if err != nil {
				t.Fatal(err)
			}
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			require.Equal(t, tt.wantStatusCode, rec.Code)
		})
	}
}
