package files_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient/files"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestNew(t *testing.T) {
	observationContext := &observation.TestContext

	tests := []struct {
		name    string
		baseURL string

		expectedErr error
	}{
		{
			name:    "Valid URL",
			baseURL: "http://some-url.foo",
		},
		{
			name:        "Invalid URL",
			baseURL:     ":foo",
			expectedErr: errors.New("parse \":foo\": missing protocol scheme"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := apiclient.BaseClientOptions{
				EndpointOptions: apiclient.EndpointOptions{
					URL: test.baseURL,
				},
			}

			_, err := files.New(observationContext, options)
			if test.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_Exists(t *testing.T) {
	observationContext := &observation.TestContext

	tests := []struct {
		name string

		handler func(t *testing.T) http.Handler

		expectedValue bool
		expectedErr   error
	}{
		{
			name: "File exists",
			handler: func(t *testing.T) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodHead, r.Method)
					assert.Contains(t, r.URL.Path, "some-bucket/foo/bar")
					assert.Equal(t, r.Header.Get("Authorization"), "token-executor hunter2")
					w.WriteHeader(http.StatusOK)
				})
			},
			expectedValue: true,
		},
		{
			name: "File does not exist",
			handler: func(t *testing.T) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				})
			},
			expectedValue: false,
		},
		{
			name: "Unexpected error",
			handler: func(t *testing.T) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				})
			},
			expectedValue: false,
			expectedErr:   errors.New("unexpected status code 500"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := httptest.NewServer(test.handler(t))
			defer srv.Close()
			options := apiclient.BaseClientOptions{
				EndpointOptions: apiclient.EndpointOptions{
					URL:        srv.URL,
					PathPrefix: "/.executors/files",
					Token:      "hunter2",
				},
			}

			client, err := files.New(observationContext, options)
			require.NoError(t, err)

			exists, err := client.Exists(context.Background(), "some-bucket", "foo/bar")

			if test.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr.Error(), err.Error())
				assert.False(t, exists)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedValue, exists)
			}
		})
	}
}

func TestClient_Get(t *testing.T) {
	observationContext := &observation.TestContext

	tests := []struct {
		name string

		handler func(t *testing.T) http.Handler

		expectedValue string
		expectedErr   error
	}{
		{
			name: "Get content",
			handler: func(t *testing.T) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodGet, r.Method)
					assert.Contains(t, r.URL.Path, "some-bucket/foo/bar")
					assert.Equal(t, r.Header.Get("Authorization"), "token-executor hunter2")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("hello world!"))
				})
			},
			expectedValue: "hello world!",
		},
		{
			name: "Failed to get content",
			handler: func(t *testing.T) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodGet, r.Method)
					assert.Contains(t, r.URL.Path, "some-bucket/foo/bar")
					assert.Equal(t, r.Header.Get("Authorization"), "token-executor hunter2")
					w.WriteHeader(http.StatusInternalServerError)
				})
			},
			expectedErr: errors.New("unexpected status code 500"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := httptest.NewServer(test.handler(t))
			defer srv.Close()
			options := apiclient.BaseClientOptions{
				EndpointOptions: apiclient.EndpointOptions{
					URL:        srv.URL,
					PathPrefix: "/.executors/files",
					Token:      "hunter2",
				},
			}

			client, err := files.New(observationContext, options)
			require.NoError(t, err)

			content, err := client.Get(context.Background(), "some-bucket", "foo/bar")
			if test.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr.Error(), err.Error())
				assert.Nil(t, content)
			} else {
				defer content.Close()
				assert.NoError(t, err)
				assert.NotNil(t, content)
				actualBytes, err := io.ReadAll(content)
				require.NoError(t, err)
				assert.Equal(t, []byte(test.expectedValue), actualBytes)
			}
		})
	}
}
