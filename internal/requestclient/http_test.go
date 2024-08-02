package requestclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"
)

type noopRoundTripper struct{ gotRequest *http.Request }

func (n *noopRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	n.gotRequest = req
	return nil, nil
}

func TestHTTP(t *testing.T) {
	tests := []struct {
		name string

		requestClient *Client

		wantClient autogold.Value
	}{
		{
			name:       "nil client",
			wantClient: autogold.Expect(&Client{IP: "192.0.2.1"}),
		},
		{
			name:          "non-nil empty client",
			requestClient: &Client{},
			wantClient:    autogold.Expect(&Client{IP: "192.0.2.1"}),
		},
		{
			name: "forwarded-for",
			requestClient: &Client{
				ForwardedFor: "192.168.1.2",
			},
			wantClient: autogold.Expect(&Client{IP: "192.0.2.1", ForwardedFor: "192.168.1.2"}),
		},
		{
			name: "client with user-agent sets forwarded-for-user-agent",
			requestClient: &Client{
				UserAgent: "Sourcegraph-Bot",
			},
			wantClient: autogold.Expect(&Client{IP: "192.0.2.1", ForwardedForUserAgent: "Sourcegraph-Bot"}),
		},
		{
			name: "client with forwarded-for-user-agent drops the current user-agent",
			requestClient: &Client{
				UserAgent:             "Not-Sourcegraph-Bot",
				ForwardedForUserAgent: "Sourcegraph-Bot",
			},
			wantClient: autogold.Expect(&Client{IP: "192.0.2.1", ForwardedForUserAgent: "Sourcegraph-Bot"}),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requestCtx := context.Background()
			if test.requestClient != nil {
				requestCtx = WithClient(requestCtx, test.requestClient)
			}

			rt := &noopRoundTripper{}
			_, err := (&HTTPTransport{RoundTripper: rt}).
				RoundTrip(
					httptest.NewRequest(http.MethodGet, "/", nil).
						WithContext(requestCtx),
				)
			require.NoError(t, err)

			var rc *Client
			httpMiddleware(
				http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
					rc = FromContext(r.Context())
				}),
				false,
			).ServeHTTP(httptest.NewRecorder(), rt.gotRequest)

			require.NotNil(t, rc)
			test.wantClient.Equal(t, rc, autogold.ExportedOnly())
		})
	}
}

func TestHTTPMiddleware_E2E(t *testing.T) {
	tests := []struct {
		name                 string
		remoteAddr           string
		useCloudflareHeaders bool
		headers              map[string][]string
		expectedIP           string
		expectedForwardedFor string
		expectedUserAgent    string
	}{
		{
			name:       "Basic request",
			remoteAddr: "192.168.1.1:12345",
			headers: map[string][]string{
				"X-Forwarded-For": {"10.0.0.1, 20.0.0.1"},
				"User-Agent":      {"Test-Agent/1.0"},
			},
			expectedIP:           "192.168.1.1",
			expectedForwardedFor: "10.0.0.1, 20.0.0.1",
			expectedUserAgent:    "Test-Agent/1.0",
		},
		{
			name:       "multiple ForwardedFor headers",
			remoteAddr: "192.168.1.1:12345",
			headers: map[string][]string{
				"X-Forwarded-For": {"10.0.0.1, 20.0.0.1", "30.0.0.1"},
				"User-Agent":      {"Test-Agent/1.0"},
			},
			expectedIP:           "192.168.1.1",
			expectedForwardedFor: "10.0.0.1, 20.0.0.1,30.0.0.1",
			expectedUserAgent:    "Test-Agent/1.0",
		},
		{
			name:                 "x-forwarded-for uses the original header if the cloudflare option is false",
			useCloudflareHeaders: false,
			remoteAddr:           "192.168.1.1:12345",
			headers: map[string][]string{
				"X-Forwarded-For":  {"10.0.0.1, 20.0.0.1"},
				"User-Agent":       {"Test-Agent/1.0"},
				"Cf-Connecting-Ip": {"192.168.1.2"},
			},
			expectedIP:           "192.168.1.1",
			expectedForwardedFor: "10.0.0.1, 20.0.0.1",
			expectedUserAgent:    "Test-Agent/1.0",
		},
		{
			name:                 "x-forwarded-for prefers the cloudflare header if the option is set",
			useCloudflareHeaders: true,
			remoteAddr:           "192.168.1.1:12345",
			headers: map[string][]string{
				"X-Forwarded-For":  {"10.0.0.1, 20.0.0.1"},
				"User-Agent":       {"Test-Agent/1.0"},
				"Cf-Connecting-Ip": {"192.168.1.2"},
			},
			expectedIP:           "192.168.1.1",
			expectedForwardedFor: "192.168.1.2",
			expectedUserAgent:    "Test-Agent/1.0",
		},
		{
			name:                 "x-forwarded-for uses the original header if the cloudflare header is empty even if the option is set",
			useCloudflareHeaders: true,
			remoteAddr:           "192.168.1.1:12345",
			headers: map[string][]string{
				"X-Forwarded-For":  {"10.0.0.1, 20.0.0.1"},
				"User-Agent":       {"Test-Agent/1.0"},
				"Cf-Connecting-Ip": {""},
			},
			expectedIP:           "192.168.1.1",
			expectedForwardedFor: "10.0.0.1, 20.0.0.1",
			expectedUserAgent:    "Test-Agent/1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req, err := http.NewRequest("GET", "http://example.com", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.RemoteAddr = tt.remoteAddr
			for k, values := range tt.headers {
				for _, v := range values {
					req.Header.Add(k, v)
				}
			}
			var clientData *Client
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				clientData = FromContext(r.Context())
			})

			handler := httpMiddleware(next, tt.useCloudflareHeaders)

			handler.ServeHTTP(httptest.NewRecorder(), req)

			if clientData == nil {
				t.Fatal("Client data not set in context")
			}

			if diff := cmp.Diff(tt.expectedIP, clientData.IP); diff != "" {
				t.Errorf("Unexpected IP (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.expectedForwardedFor, clientData.ForwardedFor); diff != "" {
				t.Errorf("Unexpected ForwardedFor (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.expectedUserAgent, clientData.UserAgent); diff != "" {
				t.Errorf("Unexpected UserAgent (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetForwardedFor(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string][]string
		expected string
	}{
		{
			name:     "No X-Forwarded-For header",
			headers:  map[string][]string{},
			expected: "",
		},
		{
			name: "Single X-Forwarded-For header",
			headers: map[string][]string{
				"X-Forwarded-For": {"192.168.1.1"},
			},
			expected: "192.168.1.1",
		},
		{
			name: "Multiple X-Forwarded-For headers",
			headers: map[string][]string{
				"X-Forwarded-For": {"192.168.1.1", "10.0.0.1"},
			},
			expected: "192.168.1.1,10.0.0.1",
		},
		{
			name: "Multiple X-Forwarded-For headers with commas",
			headers: map[string][]string{
				"X-Forwarded-For": {"192.168.1.1, 10.0.0.1", "172.16.0.1"},
			},
			expected: "192.168.1.1, 10.0.0.1,172.16.0.1",
		},
		{
			name: "Mixed case header name",
			headers: map[string][]string{
				"x-ForWarded-FOR": {"192.168.1.1"},
			},
			expected: "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
			if err != nil {
				t.Fatal(err)
			}

			for k, values := range tt.headers {
				for _, v := range values {
					req.Header.Add(k, v)
				}
			}

			result := getForwardedFor(req)
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("Unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetCloudFlareIP(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string][]string
		expected string
	}{
		{
			name:     "No Cf-Connecting-Ip header",
			headers:  map[string][]string{},
			expected: "",
		},
		{
			name: "Single Cf-Connecting-Ip header",
			headers: map[string][]string{
				"Cf-Connecting-Ip": {"192.168.1.1"},
			},
			expected: "192.168.1.1",
		},
		{
			name: "Multiple Cf-Connecting-Ip headers (should use first)",
			headers: map[string][]string{
				"Cf-Connecting-Ip": {"192.168.1.1", "10.0.0.1"},
			},
			expected: "192.168.1.1",
		},
		{
			name: "Mixed case header name",
			headers: map[string][]string{
				"cF-ConNecting-IP": {"192.168.1.1"},
			},
			expected: "192.168.1.1",
		},
		{
			name: "IPv6 address",
			headers: map[string][]string{
				"Cf-Connecting-Ip": {"2001:db8::1"},
			},
			expected: "2001:db8::1",
		},
		{
			name: "Empty Cf-Connecting-Ip header",
			headers: map[string][]string{
				"Cf-Connecting-Ip": {""},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
			if err != nil {
				t.Fatal(err)
			}

			for k, values := range tt.headers {
				for _, v := range values {
					req.Header.Add(k, v)
				}
			}

			result := getCloudFlareIP(req)
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("Unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}
