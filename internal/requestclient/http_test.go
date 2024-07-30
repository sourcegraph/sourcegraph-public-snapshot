package requestclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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
