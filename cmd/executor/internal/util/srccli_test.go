package util_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/util"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestLatestSrcCLIVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		writeResponse   func(w http.ResponseWriter)
		expectedVersion string
		expectedError   error
	}{
		{
			name: "Got latest version",
			writeResponse: func(w http.ResponseWriter) {
				w.Write([]byte(`{"version": "1.2.3"}`))
			},
			expectedVersion: "1.2.3",
		},
		{
			name: "Failed to get version",
			writeResponse: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedError: errors.New("unexpected status code 500"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				test.writeResponse(w)
			}))
			defer server.Close()

			client, err := apiclient.NewBaseClient(logtest.Scoped(t), apiclient.BaseClientOptions{EndpointOptions: apiclient.EndpointOptions{URL: server.URL}})
			require.NoError(t, err)

			version, err := util.LatestSrcCLIVersion(context.Background(), client)
			if test.expectedError != nil {
				require.Error(t, err)
				require.EqualError(t, err, test.expectedError.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedVersion, version)
			}
		})
	}
}
