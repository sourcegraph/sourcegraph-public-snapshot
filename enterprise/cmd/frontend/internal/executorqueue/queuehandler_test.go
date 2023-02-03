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
	router.Use(executorAuthMiddleware(accessTokenFunc))

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

//func TestHandler_AuthMiddleware(t *testing.T) {
//	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{ExecutorsAccessToken: "hunter2"}})
//
//	tests := []struct {
//		name                 string
//		header               http.Header
//		body                 string
//		mockFunc             func(executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore)
//		expectedStatusCode   int
//		expectedResponseBody string
//		assertionFunc        func(t *testing.T, executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore)
//	}{
//		{
//			name:   "Authorized",
//			header: http.Header{"Authorization": []string{"Bearer somejobtoken"}},
//			body:   `{"executorName": "test-executor", "jobId": 42}`,
//			mockFunc: func(executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
//				jobTokenStore.GetByTokenFunc.PushReturn(executor.JobToken{JobID: 42, Queue: "test"}, nil)
//				executorStore.GetByHostnameFunc.PushReturn(types.Executor{}, true, nil)
//			},
//			expectedStatusCode: http.StatusTeapot,
//			assertionFunc: func(t *testing.T, executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
//				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
//				assert.Equal(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
//				require.Len(t, executorStore.GetByHostnameFunc.History(), 1)
//				assert.Equal(t, executorStore.GetByHostnameFunc.History()[0].Arg1, "test-executor")
//			},
//		},
//		{
//			name:   "Authorized general access token",
//			header: http.Header{"Authorization": []string{"token-executor hunter2"}},
//			body:   `{"executorName": "test-executor", "jobId": 42}`,
//			mockFunc: func(executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
//				jobTokenStore.GetByTokenFunc.PushReturn(executor.JobToken{JobID: 42, Queue: "test"}, nil)
//				executorStore.GetByHostnameFunc.PushReturn(types.Executor{}, true, nil)
//			},
//			expectedStatusCode: http.StatusTeapot,
//			assertionFunc: func(t *testing.T, executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
//				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
//				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
//			},
//		},
//		{
//			name:                 "No request body",
//			expectedStatusCode:   http.StatusBadRequest,
//			expectedResponseBody: "No request body provided\n",
//		},
//		{
//			name:                 "Malformed request body",
//			body:                 `{"executorName": "test-executor"`,
//			expectedStatusCode:   http.StatusBadRequest,
//			expectedResponseBody: "Failed to parse request body\n",
//		},
//		{
//			name:                 "No worker hostname provided",
//			body:                 `{"jobId": 42}`,
//			expectedStatusCode:   http.StatusBadRequest,
//			expectedResponseBody: "worker hostname cannot be empty\n",
//		},
//		{
//			name:                 "No Authorized header",
//			body:                 `{"executorName": "test-executor", "jobId": 42}`,
//			expectedStatusCode:   http.StatusUnauthorized,
//			expectedResponseBody: "no token value in the HTTP Authorization request header\n",
//		},
//		{
//			name:                 "Invalid Authorized header parts",
//			header:               http.Header{"Authorization": []string{"token-executor"}},
//			body:                 `{"executorName": "test-executor", "jobId": 42}`,
//			expectedStatusCode:   http.StatusUnauthorized,
//			expectedResponseBody: "HTTP Authorization request header value must be of the following form: 'Bearer \"TOKEN\"' or 'token-executor TOKEN'\n",
//		},
//		{
//			name:                 "Invalid Authorized header prefix",
//			header:               http.Header{"Authorization": []string{"Foo bar"}},
//			body:                 `{"executorName": "test-executor", "jobId": 42}`,
//			expectedStatusCode:   http.StatusUnauthorized,
//			expectedResponseBody: "unrecognized HTTP Authorization request header scheme (supported values: \"Bearer\", \"token-executor\")\n",
//		},
//		{
//			name:               "Invalid general access token",
//			header:             http.Header{"Authorization": []string{"token-executor hunter1"}},
//			body:               `{"executorName": "test-executor", "jobId": 42}`,
//			expectedStatusCode: http.StatusForbidden,
//		},
//		{
//			name:   "Failed to retrieve job token",
//			header: http.Header{"Authorization": []string{"Bearer somejobtoken"}},
//			body:   `{"executorName": "test-executor", "jobId": 42}`,
//			mockFunc: func(executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
//				jobTokenStore.GetByTokenFunc.PushReturn(executor.JobToken{}, errors.New("failed to find job token"))
//			},
//			expectedStatusCode:   http.StatusUnauthorized,
//			expectedResponseBody: "invalid token\n",
//			assertionFunc: func(t *testing.T, executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
//				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
//				assert.Equal(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
//				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
//			},
//		},
//		{
//			name:   "JobID does not match",
//			header: http.Header{"Authorization": []string{"Bearer somejobtoken"}},
//			body:   `{"executorName": "test-executor", "jobId": 42}`,
//			mockFunc: func(executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
//				jobTokenStore.GetByTokenFunc.PushReturn(executor.JobToken{JobID: 7, Queue: "test"}, nil)
//			},
//			expectedStatusCode:   http.StatusForbidden,
//			expectedResponseBody: "invalid token\n",
//			assertionFunc: func(t *testing.T, executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
//				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
//				assert.Equal(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
//				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
//			},
//		},
//		{
//			name:   "Queue does not match",
//			header: http.Header{"Authorization": []string{"Bearer somejobtoken"}},
//			body:   `{"executorName": "test-executor", "jobId": 42}`,
//			mockFunc: func(executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
//				jobTokenStore.GetByTokenFunc.PushReturn(executor.JobToken{JobID: 42, Queue: "test1"}, nil)
//			},
//			expectedStatusCode:   http.StatusForbidden,
//			expectedResponseBody: "invalid token\n",
//			assertionFunc: func(t *testing.T, executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
//				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
//				assert.Equal(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
//				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
//			},
//		},
//		{
//			name:   "Executor host does not exist",
//			header: http.Header{"Authorization": []string{"Bearer somejobtoken"}},
//			body:   `{"executorName": "test-executor", "jobId": 42}`,
//			mockFunc: func(executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
//				jobTokenStore.GetByTokenFunc.PushReturn(executor.JobToken{JobID: 42, Queue: "test"}, nil)
//				executorStore.GetByHostnameFunc.PushReturn(types.Executor{}, false, errors.New("executor does not exist"))
//			},
//			expectedStatusCode:   http.StatusUnauthorized,
//			expectedResponseBody: "invalid token\n",
//			assertionFunc: func(t *testing.T, executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
//				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
//				assert.Equal(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
//				require.Len(t, executorStore.GetByHostnameFunc.History(), 1)
//				assert.Equal(t, executorStore.GetByHostnameFunc.History()[0].Arg1, "test-executor")
//			},
//		},
//	}
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			executorStore := database.NewMockExecutorStore()
//			jobTokenStore := executor.NewMockJobTokenStore()
//
//			h := handler.NewHandler(
//				executorStore,
//				jobTokenStore,
//				metricsstore.NewMockDistributedStore(),
//				handler.QueueHandler[testRecord]{},
//			)
//
//			router := mux.NewRouter()
//			router.HandleFunc("/{queueName}", func(w http.ResponseWriter, r *http.Request) {
//				w.WriteHeader(http.StatusTeapot)
//			})
//			router.Use(h.AuthMiddleware)
//
//			req, err := http.NewRequest("GET", "/test", strings.NewReader(test.body))
//			require.NoError(t, err)
//			req.Header = test.header
//
//			rw := httptest.NewRecorder()
//
//			if test.mockFunc != nil {
//				test.mockFunc(executorStore, jobTokenStore)
//			}
//
//			router.ServeHTTP(rw, req)
//
//			assert.Equal(t, test.expectedStatusCode, rw.Code)
//
//			b, err := io.ReadAll(rw.Body)
//			require.NoError(t, err)
//			assert.Equal(t, test.expectedResponseBody, string(b))
//
//			if test.assertionFunc != nil {
//				test.assertionFunc(t, executorStore, jobTokenStore)
//			}
//		})
//	}
//}
