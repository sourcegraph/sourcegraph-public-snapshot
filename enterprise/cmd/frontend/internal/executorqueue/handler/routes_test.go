package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
)

func TestSetupRoutes(t *testing.T) {
	router := mux.NewRouter()
	h := new(testExecutorHandler)
	handler.SetupRoutes(h, router)

	tests := []struct {
		name               string
		method             string
		path               string
		expectedStatusCode int
		expectationsFunc   func()
	}{
		{
			name:               "Dequeue",
			method:             http.MethodPost,
			path:               "/test/dequeue",
			expectedStatusCode: http.StatusOK,
			expectationsFunc: func() {
				h.On("HandleDequeue").Once()
			},
		},
		{
			name:               "Heartbeat",
			method:             http.MethodPost,
			path:               "/test/heartbeat",
			expectedStatusCode: http.StatusOK,
			expectationsFunc: func() {
				h.On("HandleHeartbeat").Once()
			},
		},
		{
			name:               "CanceledJobs",
			method:             http.MethodPost,
			path:               "/test/canceledJobs",
			expectedStatusCode: http.StatusOK,
			expectationsFunc: func() {
				h.On("HandleCanceledJobs").Once()
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(test.method, test.path, nil)
			require.NoError(t, err)
			responseRecorder := httptest.NewRecorder()

			h.On("AuthMiddleware").Times(0)
			test.expectationsFunc()
			router.ServeHTTP(responseRecorder, req)

			assert.Equal(t, test.expectedStatusCode, responseRecorder.Code)

			h.AssertExpectations(t)
		})
	}
}

type testExecutorHandler struct {
	mock.Mock
}

func (t *testExecutorHandler) Name() string {
	return "test"
}

func (t *testExecutorHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func (t *testExecutorHandler) HandleDequeue(w http.ResponseWriter, r *http.Request) {
	t.Called()
}

func (t *testExecutorHandler) HandleAddExecutionLogEntry(w http.ResponseWriter, r *http.Request) {
	t.Called()
}

func (t *testExecutorHandler) HandleUpdateExecutionLogEntry(w http.ResponseWriter, r *http.Request) {
	t.Called()
}

func (t *testExecutorHandler) HandleMarkComplete(w http.ResponseWriter, r *http.Request) {
	t.Called()
}

func (t *testExecutorHandler) HandleMarkErrored(w http.ResponseWriter, r *http.Request) {
	t.Called()
}

func (t *testExecutorHandler) HandleMarkFailed(w http.ResponseWriter, r *http.Request) {
	t.Called()
}

func (t *testExecutorHandler) HandleHeartbeat(w http.ResponseWriter, r *http.Request) {
	t.Called()
}

func (t *testExecutorHandler) HandleCanceledJobs(w http.ResponseWriter, r *http.Request) {
	t.Called()
}
