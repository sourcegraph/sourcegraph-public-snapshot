package handler_test

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestJobAuthMiddleware(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	router.Use(handler.JobAuthMiddleware)

	accessToken := "hunter2"
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			ExecutorsAccessToken: accessToken,
			Executors: &schema.Executors{
				JobAccessToken: &schema.JobAccessToken{
					SigningKey: "ZXhlY3V0b3JzLnRlc3Quc2lnbmluZ0tleQo=",
				},
			},
		},
	})

	tests := []struct {
		name                 string
		headers              http.Header
		body                 string
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:               "Authorized",
			headers:            http.Header{"Authorization": {fmt.Sprintf("Bearer %s", newJobToken(t, "test-executor", 42, accessToken, 10))}},
			body:               `{"executorName": "test-executor", "jobId": 42}`,
			expectedStatusCode: http.StatusTeapot,
		},
		{
			name:               "Authorized with general access token",
			headers:            http.Header{"Authorization": {"token-executor hunter2"}},
			body:               `{"executorName": "test-executor", "jobId": 42}`,
			expectedStatusCode: http.StatusTeapot,
		},
		{
			name:                 "No request body",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Failed to parse request body\n",
		},
		{
			name:                 "Malformed request body",
			body:                 `{"executorName": "test-executor", "jobId": 42`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Failed to parse request body\n",
		},
		{
			name:                 "No authorization header",
			body:                 `{"executorName": "test-executor", "jobId": 42}`,
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: "no token value in the HTTP Authorization request header\n",
		},
		{
			name:                 "Invalid authorization header parts",
			headers:              http.Header{"Authorization": {accessToken}},
			body:                 `{"executorName": "test-executor", "jobId": 42}`,
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: "HTTP Authorization request header value must be of the following form: 'Bearer \"TOKEN\"' or 'token-executor TOKEN'\n",
		},
		{
			name:                 "Invalid authorization header prefix",
			headers:              http.Header{"Authorization": {"foo hunter2"}},
			body:                 `{"executorName": "test-executor", "jobId": 42}`,
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: "unrecognized HTTP Authorization request header scheme (supported values: \"Bearer\", \"token-executor\")\n",
		},
		{
			name:               "Invalid general access token",
			headers:            http.Header{"Authorization": {"token-executor hunter3"}},
			body:               `{"executorName": "test-executor", "jobId": 42}`,
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:                 "Invalid job token",
			headers:              http.Header{"Authorization": {"Bearer foobar"}},
			body:                 `{"executorName": "test-executor", "jobId": 42}`,
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: "invalid token\n",
		},
		{
			name:               "Job token does not match general access token",
			headers:            http.Header{"Authorization": {fmt.Sprintf("Bearer %s", newJobToken(t, "test-executor", 42, "hunter3", 10))}},
			body:               `{"executorName": "test-executor", "jobId": 42}`,
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:               "Job token does not match executor name",
			headers:            http.Header{"Authorization": {fmt.Sprintf("Bearer %s", newJobToken(t, "test-executor", 42, accessToken, 10))}},
			body:               `{"executorName": "test-executor1", "jobId": 42}`,
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:               "Job token does not match job id",
			headers:            http.Header{"Authorization": {fmt.Sprintf("Bearer %s", newJobToken(t, "test-executor", 42, accessToken, 10))}},
			body:               `{"executorName": "test-executor", "jobId": 7}`,
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:                 "Job token expired",
			headers:              http.Header{"Authorization": {fmt.Sprintf("Bearer %s", newJobToken(t, "test-executor", 42, accessToken, -10))}},
			body:                 `{"executorName": "test-executor", "jobId": 7}`,
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: "invalid token\n",
		},
		{
			name:                 "Invalid worker hostname",
			headers:              http.Header{"Authorization": {fmt.Sprintf("Bearer %s", newJobToken(t, "", 42, accessToken, 10))}},
			body:                 `{"executorName": "", "jobId": 42}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "worker hostname cannot be empty\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", strings.NewReader(test.body))
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

func newJobToken(t *testing.T, executorName string, jobId int, accessToken string, expiryInt int) string {
	expiry := time.Now().Add(time.Minute * time.Duration(expiryInt))
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, struct {
		jwt.RegisteredClaims
		AccessToken string `json:"accessToken"`
	}{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    executorName,
			ExpiresAt: jwt.NewNumericDate(expiry),
			Subject:   strconv.FormatInt(int64(jobId), 10),
		},
		AccessToken: accessToken,
	})
	decodedSigningKey, err := base64.StdEncoding.DecodeString(conf.SiteConfig().Executors.JobAccessToken.SigningKey)
	require.NoError(t, err)
	tokenString, err := token.SignedString(decodedSigningKey)
	require.NoError(t, err)
	return tokenString
}
