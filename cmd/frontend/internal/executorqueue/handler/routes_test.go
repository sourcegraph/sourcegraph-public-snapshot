package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/executorqueue/handler"
)

func TestSetupRoutes(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		path               string
		expectedStatusCode int
		expectationsFunc   func(h *testExecutorHandler)
	}{
		{
			name:               "Dequeue",
			method:             http.MethodPost,
			path:               "/test/dequeue",
			expectedStatusCode: http.StatusOK,
			expectationsFunc: func(h *testExecutorHandler) {
				h.On("HandleDequeue").Once()
			},
		},
		{
			name:               "Heartbeat",
			method:             http.MethodPost,
			path:               "/test/heartbeat",
			expectedStatusCode: http.StatusOK,
			expectationsFunc: func(h *testExecutorHandler) {
				h.On("HandleHeartbeat").Once()
			},
		},
		{
			name:               "Invalid root",
			method:             http.MethodPost,
			path:               "/test1/dequeue",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "Invalid path",
			method:             http.MethodPost,
			path:               "/test/foo",
			expectedStatusCode: http.StatusNotFound,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := mux.NewRouter()
			h := new(testExecutorHandler)
			handler.SetupRoutes(h, router)

			req, err := http.NewRequest(test.method, test.path, nil)
			require.NoError(t, err)
			responseRecorder := httptest.NewRecorder()

			if test.expectationsFunc != nil {
				test.expectationsFunc(h)
			}
			router.ServeHTTP(responseRecorder, req)

			assert.Equal(t, test.expectedStatusCode, responseRecorder.Code)

			h.AssertExpectations(t)
		})
	}
}

func TestSetupJobRoutes(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		path               string
		expectedStatusCode int
		expectationsFunc   func(h *testExecutorHandler)
	}{
		{
			name:               "AddExecutionLogEntry",
			method:             http.MethodPost,
			path:               "/test/addExecutionLogEntry",
			expectedStatusCode: http.StatusOK,
			expectationsFunc: func(h *testExecutorHandler) {
				h.On("HandleAddExecutionLogEntry").Once()
			},
		},
		{
			name:               "UpdateExecutionLogEntry",
			method:             http.MethodPost,
			path:               "/test/updateExecutionLogEntry",
			expectedStatusCode: http.StatusOK,
			expectationsFunc: func(h *testExecutorHandler) {
				h.On("HandleUpdateExecutionLogEntry").Once()
			},
		},
		{
			name:               "MarkComplete",
			method:             http.MethodPost,
			path:               "/test/markComplete",
			expectedStatusCode: http.StatusOK,
			expectationsFunc: func(h *testExecutorHandler) {
				h.On("HandleMarkComplete").Once()
			},
		},
		{
			name:               "MarkErrored",
			method:             http.MethodPost,
			path:               "/test/markErrored",
			expectedStatusCode: http.StatusOK,
			expectationsFunc: func(h *testExecutorHandler) {
				h.On("HandleMarkErrored").Once()
			},
		},
		{
			name:               "MarkFailed",
			method:             http.MethodPost,
			path:               "/test/markFailed",
			expectedStatusCode: http.StatusOK,
			expectationsFunc: func(h *testExecutorHandler) {
				h.On("HandleMarkFailed").Once()
			},
		},
		{
			name:               "Invalid root",
			method:             http.MethodPost,
			path:               "/test1/addExecutionLogEntry",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "Invalid path",
			method:             http.MethodPost,
			path:               "/test/foo",
			expectedStatusCode: http.StatusNotFound,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := mux.NewRouter()
			h := new(testExecutorHandler)
			handler.SetupJobRoutes(h, router)

			req, err := http.NewRequest(test.method, test.path, nil)
			require.NoError(t, err)
			responseRecorder := httptest.NewRecorder()

			if test.expectationsFunc != nil {
				test.expectationsFunc(h)
			}
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
