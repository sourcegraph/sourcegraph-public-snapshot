package run

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestValidateAuthorizationToken(t *testing.T) {
	tests := []struct {
		name                string
		statusCode          int
		expectedErr         error
		isUnauthorizedError bool
	}{
		{
			name:       "Valid response",
			statusCode: http.StatusOK,
		},
		{
			name:                "Unauthorized",
			statusCode:          http.StatusUnauthorized,
			expectedErr:         authorizationFailedErr,
			isUnauthorizedError: true,
		},
		{
			name:                "Internal server error",
			statusCode:          http.StatusInternalServerError,
			expectedErr:         errors.New("failed to validate authorization token: unexpected status code 500"),
			isUnauthorizedError: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, client := newTestServerAndClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(test.statusCode)
			})
			defer server.Close()

			err := validateAuthorizationToken(context.Background(), client)
			if test.expectedErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, errors.Is(err, authorizationFailedErr), test.isUnauthorizedError)
				assert.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func newTestServerAndClient(t *testing.T, handlerFunc func(w http.ResponseWriter, r *http.Request)) (*httptest.Server, *apiclient.BaseClient) {
	server := httptest.NewServer(http.HandlerFunc(handlerFunc))
	testOpts := testOptions(&config.Config{FrontendURL: server.URL, FrontendAuthorizationToken: "hunter2hunter2"})
	client, err := apiclient.NewBaseClient(logtest.Scoped(t), testOpts)
	require.NoError(t, err)

	return server, client
}
