package executorqueue

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	accessToken := "hunter2"

	accessTokenFunc := func() string { return accessToken }

	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	router.Use(authExecutorMiddleware(accessTokenFunc))

	tests := []struct {
		name                 string
		headers              http.Header
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:               "Authorized",
			headers:            http.Header{"Authorization": {"token-executor hunter2"}},
			expectedStatusCode: http.StatusTeapot,
		},
		{
			name:                 "Missing Authorization header",
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: "no token value in the HTTP Authorization request header (recommended) or basic auth (deprecated)\n",
		},
		{
			name:               "Wrong token",
			headers:            http.Header{"Authorization": {"token-executor foobar"}},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:                 "Invalid prefix",
			headers:              http.Header{"Authorization": {"foo hunter2"}},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: "unrecognized HTTP Authorization request header scheme (supported values: \"token-executor\")\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			require.NoError(t, err)
			req.Header = test.headers

			rw := httptest.NewRecorder()

			router.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatusCode, rw.Code)

			b, err := io.ReadAll(rw.Body)
			require.NoError(t, err)
			assert.Equal(t, test.expectedResponseBody, string(b))
		})
	}
}
