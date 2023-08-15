package executorqueue

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	executorstore "github.com/sourcegraph/sourcegraph/internal/executor/store"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAuthMiddleware(t *testing.T) {
	logger := logtest.Scoped(t)
	accessToken := "hunter2"

	accessTokenFunc := func() string { return accessToken }

	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	router.Use(executorAuthMiddleware(logger, accessTokenFunc))

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
			expectedStatusCode: http.StatusUnauthorized,
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

func TestJobAuthMiddleware(t *testing.T) {
	logger := logtest.Scoped(t)
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{ExecutorsAccessToken: "hunter2"}})

	tests := []struct {
		name                 string
		routeName            routeName
		header               map[string]string
		mockFunc             func(executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore)
		expectedStatusCode   int
		expectedResponseBody string
		assertionFunc        func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore)
	}{
		{
			name:      "Queue Authorized",
			routeName: routeQueue,
			header: map[string]string{
				"Authorization":               "Bearer somejobtoken",
				"X-Sourcegraph-Job-ID":        "42",
				"X-Sourcegraph-Executor-Name": "test-executor",
			},
			mockFunc: func(executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executorstore.JobToken{JobID: 42, Queue: "test"}, nil)
				executorStore.GetByHostnameFunc.PushReturn(types.Executor{}, true, nil)
			},
			expectedStatusCode: http.StatusTeapot,
			assertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
				assert.Equal(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
				require.Len(t, executorStore.GetByHostnameFunc.History(), 1)
				assert.Equal(t, executorStore.GetByHostnameFunc.History()[0].Arg1, "test-executor")
			},
		},
		{
			name:      "Queue Authorized general access token",
			routeName: routeQueue,
			header: map[string]string{
				"Authorization": "token-executor hunter2",
			},
			expectedStatusCode: http.StatusTeapot,
			assertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
			},
		},
		{
			name:      "Git Authorized",
			routeName: routeGit,
			header: map[string]string{
				"Authorization":               "Bearer somejobtoken",
				"X-Sourcegraph-Job-ID":        "42",
				"X-Sourcegraph-Executor-Name": "test-executor",
			},
			mockFunc: func(executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executorstore.JobToken{JobID: 42, Repo: "test"}, nil)
				executorStore.GetByHostnameFunc.PushReturn(types.Executor{}, true, nil)
			},
			expectedStatusCode: http.StatusTeapot,
			assertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
				assert.Equal(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
				require.Len(t, executorStore.GetByHostnameFunc.History(), 1)
				assert.Equal(t, executorStore.GetByHostnameFunc.History()[0].Arg1, "test-executor")
			},
		},
		{
			name:      "Git Authorized general access token",
			routeName: routeGit,
			header: map[string]string{
				"Authorization": "token-executor hunter2",
			},
			expectedStatusCode: http.StatusTeapot,
			assertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
			},
		},
		{
			name:      "Files Authorized",
			routeName: routeFiles,
			header: map[string]string{
				"Authorization":               "Bearer somejobtoken",
				"X-Sourcegraph-Job-ID":        "42",
				"X-Sourcegraph-Executor-Name": "test-executor",
			},
			mockFunc: func(executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executorstore.JobToken{JobID: 42, Queue: "batches"}, nil)
				executorStore.GetByHostnameFunc.PushReturn(types.Executor{}, true, nil)
			},
			expectedStatusCode: http.StatusTeapot,
			assertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
				assert.Equal(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
				require.Len(t, executorStore.GetByHostnameFunc.History(), 1)
				assert.Equal(t, executorStore.GetByHostnameFunc.History()[0].Arg1, "test-executor")
			},
		},
		{
			name:      "Files Authorized general access token",
			routeName: routeFiles,
			header: map[string]string{
				"Authorization": "token-executor hunter2",
			},
			expectedStatusCode: http.StatusTeapot,
			assertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
			},
		},
		{
			name:      "No worker hostname provided",
			routeName: routeQueue,
			header: map[string]string{
				"Authorization":        "Bearer somejobtoken",
				"X-Sourcegraph-Job-ID": "42",
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "worker hostname cannot be empty\n",
			assertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
			},
		},
		{
			name:      "No job id header",
			routeName: routeQueue,
			header: map[string]string{
				"Authorization":               "Bearer somejobtoken",
				"X-Sourcegraph-Executor-Name": "test-executor",
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "job ID not provided in header 'X-Sourcegraph-Job-ID'\n",
			assertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
			},
		},
		{
			name:      "Invalid job id header",
			routeName: routeQueue,
			header: map[string]string{
				"Authorization":               "Bearer somejobtoken",
				"X-Sourcegraph-Executor-Name": "test-executor",
				"X-Sourcegraph-Job-ID":        "abc",
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "failed to parse Job ID: strconv.Atoi: parsing \"abc\": invalid syntax\n",
			assertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
			},
		},
		{
			name:                 "No Authorized header",
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: "no token value in the HTTP Authorization request header\n",
			assertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
			},
		},
		{
			name: "Invalid Authorized header parts",
			header: map[string]string{
				"Authorization": "somejobtoken",
			},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: "HTTP Authorization request header value must be of the following form: 'Bearer \"TOKEN\"' or 'token-executor TOKEN'\n",
			assertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
			},
		},
		{
			name: "Invalid Authorized header prefix",
			header: map[string]string{
				"Authorization": "Foo bar",
			},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: "unrecognized HTTP Authorization request header scheme (supported values: \"Bearer\", \"token-executor\")\n",
			assertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
			},
		},
		{
			name:      "Invalid general access token",
			routeName: routeQueue,
			header: map[string]string{
				"Authorization": "token-executor hunter3",
			},
			expectedStatusCode: http.StatusForbidden,
			assertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
			},
		},
		{
			name: "Unsupported route",
			header: map[string]string{
				"Authorization":               "Bearer somejobtoken",
				"X-Sourcegraph-Job-ID":        "42",
				"X-Sourcegraph-Executor-Name": "test-executor",
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "unsupported route\n",
			assertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
			},
		},
		{
			name:      "Failed to retrieve job token",
			routeName: routeQueue,
			header: map[string]string{
				"Authorization":               "Bearer somejobtoken",
				"X-Sourcegraph-Job-ID":        "42",
				"X-Sourcegraph-Executor-Name": "test-executor",
			},
			mockFunc: func(executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executorstore.JobToken{}, errors.New("failed to find job token"))
			},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: "invalid token\n",
			assertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
				assert.Equal(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
			},
		},
		{
			name:      "Job ID does not match",
			routeName: routeQueue,
			header: map[string]string{
				"Authorization":               "Bearer somejobtoken",
				"X-Sourcegraph-Job-ID":        "42",
				"X-Sourcegraph-Executor-Name": "test-executor",
			},
			mockFunc: func(executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executorstore.JobToken{JobID: 7, Queue: "test"}, nil)
			},
			expectedStatusCode:   http.StatusForbidden,
			expectedResponseBody: "invalid token\n",
			assertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
				assert.Equal(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
			},
		},
		{
			name:      "Queue does not match",
			routeName: routeQueue,
			header: map[string]string{
				"Authorization":               "Bearer somejobtoken",
				"X-Sourcegraph-Job-ID":        "42",
				"X-Sourcegraph-Executor-Name": "test-executor",
			},
			mockFunc: func(executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executorstore.JobToken{JobID: 42, Queue: "test1"}, nil)
			},
			expectedStatusCode:   http.StatusForbidden,
			expectedResponseBody: "invalid token\n",
			assertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
				assert.Equal(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
			},
		},
		{
			name:      "Executor host does not exist",
			routeName: routeQueue,
			header: map[string]string{
				"Authorization":               "Bearer somejobtoken",
				"X-Sourcegraph-Job-ID":        "42",
				"X-Sourcegraph-Executor-Name": "test-executor",
			},
			mockFunc: func(executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executorstore.JobToken{JobID: 42, Queue: "test"}, nil)
				executorStore.GetByHostnameFunc.PushReturn(types.Executor{}, false, errors.New("executor does not exist"))
			},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: "invalid token\n",
			assertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
				assert.Equal(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
				require.Len(t, executorStore.GetByHostnameFunc.History(), 1)
				assert.Equal(t, executorStore.GetByHostnameFunc.History()[0].Arg1, "test-executor")
			},
		},
		{
			name:      "Repo does not exist",
			routeName: routeGit,
			header: map[string]string{
				"Authorization":               "Bearer somejobtoken",
				"X-Sourcegraph-Job-ID":        "42",
				"X-Sourcegraph-Executor-Name": "test-executor",
			},
			mockFunc: func(executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executorstore.JobToken{JobID: 42, Repo: "test1"}, nil)
			},
			expectedStatusCode:   http.StatusForbidden,
			expectedResponseBody: "invalid token\n",
			assertionFunc: func(t *testing.T, executorStore *dbmocks.MockExecutorStore, jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
				assert.Equal(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executorStore := dbmocks.NewMockExecutorStore()
			jobTokenStore := executorstore.NewMockJobTokenStore()

			router := mux.NewRouter()
			if test.routeName == routeGit {
				router.HandleFunc("/{RepoName}", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusTeapot)
				})
			} else {
				router.HandleFunc("/{queueName}", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusTeapot)
				})
			}
			router.Use(jobAuthMiddleware(logger, test.routeName, jobTokenStore, executorStore))

			req, err := http.NewRequest("GET", "/test", nil)
			require.NoError(t, err)
			for k, v := range test.header {
				req.Header.Add(k, v)
			}

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(executorStore, jobTokenStore)
			}

			router.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatusCode, rw.Code)

			b, err := io.ReadAll(rw.Body)
			require.NoError(t, err)
			assert.Equal(t, test.expectedResponseBody, string(b))

			if test.assertionFunc != nil {
				test.assertionFunc(t, executorStore, jobTokenStore)
			}
		})
	}
}
