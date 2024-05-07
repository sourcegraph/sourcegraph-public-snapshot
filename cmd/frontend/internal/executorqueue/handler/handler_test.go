package handler_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	internalexecutor "github.com/sourcegraph/sourcegraph/internal/executor"
	executorstore "github.com/sourcegraph/sourcegraph/internal/executor/store"
	executortypes "github.com/sourcegraph/sourcegraph/internal/executor/types"
	metricsstore "github.com/sourcegraph/sourcegraph/internal/metrics/store"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	dbworkerstoremocks "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store/mocks"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestHandler_Name(t *testing.T) {
	queueHandler := handler.QueueHandler[testRecord]{Name: "test"}
	h := handler.NewHandler(
		dbmocks.NewMockExecutorStore(),
		executorstore.NewMockJobTokenStore(),
		metricsstore.NewMockDistributedStore(),
		queueHandler,
	)
	assert.Equal(t, "test", h.Name())
}

func TestHandler_HandleDequeue(t *testing.T) {
	tests := []struct {
		name                 string
		body                 string
		transformerFunc      handler.TransformerFunc[testRecord]
		mockFunc             func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore)
		expectedStatusCode   int
		expectedResponseBody string
		assertionFunc        func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore)
	}{
		{
			name: "Dequeue record",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
			transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
				return executortypes.Job{ID: record.RecordID()}, nil
			},
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				jobTokenStore.CreateFunc.PushReturn("sometoken", nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":1,"token":"sometoken","repositoryName":"","repositoryDirectory":"","commit":"","fetchTags":false,"shallowClone":false,"sparseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redactedValues":null}`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				assert.Equal(t, "test-executor", mockStore.DequeueFunc.History()[0].Arg1)
				assert.Nil(t, mockStore.DequeueFunc.History()[0].Arg2)
				require.Len(t, jobTokenStore.CreateFunc.History(), 1)
				assert.Equal(t, 1, jobTokenStore.CreateFunc.History()[0].Arg1)
				assert.Equal(t, "test", jobTokenStore.CreateFunc.History()[0].Arg2)
			},
		},
		{
			name:                 "Invalid version",
			body:                 `{"executorName": "test-executor", "version":"\n1.2", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"failed to check version \"\\n1.2\": Invalid Semantic Version"}`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 0)
				require.Len(t, jobTokenStore.CreateFunc.History(), 0)
			},
		},
		{
			name: "Dequeue error",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{}, false, errors.New("failed to dequeue"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"dbworkerstore.Dequeue: failed to dequeue"}`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CreateFunc.History(), 0)
			},
		},
		{
			name: "Nothing to dequeue",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{}, false, nil)
			},
			expectedStatusCode: http.StatusNoContent,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CreateFunc.History(), 0)
			},
		},
		{
			name: "Failed to transform record",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
			transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
				return executortypes.Job{}, errors.New("failed")
			},
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				mockStore.MarkFailedFunc.PushReturn(true, nil)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"RecordTransformer: failed"}`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, mockStore.MarkFailedFunc.History(), 1)
				assert.Equal(t, 1, mockStore.MarkFailedFunc.History()[0].Arg1)
				assert.Equal(t, "failed to transform record: failed", mockStore.MarkFailedFunc.History()[0].Arg2)
				assert.Equal(t, dbworkerstore.MarkFinalOptions{}, mockStore.MarkFailedFunc.History()[0].Arg3)
				require.Len(t, jobTokenStore.CreateFunc.History(), 0)
			},
		},
		{
			name: "Failed to mark record as failed",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
			transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
				return executortypes.Job{}, errors.New("failed")
			},
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				mockStore.MarkFailedFunc.PushReturn(false, errors.New("failed to mark"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"RecordTransformer: failed"}`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, mockStore.MarkFailedFunc.History(), 1)
				assert.Equal(t, 1, mockStore.MarkFailedFunc.History()[0].Arg1)
				assert.Equal(t, "failed to transform record: failed", mockStore.MarkFailedFunc.History()[0].Arg2)
				assert.Equal(t, dbworkerstore.MarkFinalOptions{}, mockStore.MarkFailedFunc.History()[0].Arg3)
				require.Len(t, jobTokenStore.CreateFunc.History(), 0)
			},
		},
		{
			name: "V2 job",
			body: `{"executorName": "test-executor", "version": "dev", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
			transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
				return executortypes.Job{ID: record.RecordID()}, nil
			},
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				jobTokenStore.CreateFunc.PushReturn("sometoken", nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"version":2,"id":1,"token":"sometoken","repositoryName":"","repositoryDirectory":"","commit":"","fetchTags":false,"shallowClone":false,"sparseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redactedValues":null,"dockerAuthConfig":{}}`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CreateFunc.History(), 1)
			},
		},
		{
			name: "Failed to create job token",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
			transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
				return executortypes.Job{ID: record.RecordID()}, nil
			},
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				jobTokenStore.CreateFunc.PushReturn("", errors.New("failed to create token"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"CreateToken: failed to create token"}`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CreateFunc.History(), 1)
				require.Len(t, jobTokenStore.RegenerateFunc.History(), 0)
			},
		},
		{
			name: "Job token already exists",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
			transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
				return executortypes.Job{ID: record.RecordID()}, nil
			},
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				jobTokenStore.CreateFunc.PushReturn("", executorstore.ErrJobTokenAlreadyCreated)
				jobTokenStore.RegenerateFunc.PushReturn("somenewtoken", nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":1,"token":"somenewtoken","repositoryName":"","repositoryDirectory":"","commit":"","fetchTags":false,"shallowClone":false,"sparseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redactedValues":null}`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CreateFunc.History(), 1)
				require.Len(t, jobTokenStore.RegenerateFunc.History(), 1)
				assert.Equal(t, 1, jobTokenStore.RegenerateFunc.History()[0].Arg1)
				assert.Equal(t, "test", jobTokenStore.RegenerateFunc.History()[0].Arg2)
			},
		},
		{
			name: "Failed to regenerate token",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
			transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
				return executortypes.Job{ID: record.RecordID()}, nil
			},
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				jobTokenStore.CreateFunc.PushReturn("", executorstore.ErrJobTokenAlreadyCreated)
				jobTokenStore.RegenerateFunc.PushReturn("", errors.New("failed to regen token"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"RegenerateToken: failed to regen token"}`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CreateFunc.History(), 1)
				require.Len(t, jobTokenStore.RegenerateFunc.History(), 1)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStore := dbworkerstoremocks.NewMockStore[testRecord]()
			jobTokenStore := executorstore.NewMockJobTokenStore()

			h := handler.NewHandler(
				dbmocks.NewMockExecutorStore(),
				jobTokenStore,
				metricsstore.NewMockDistributedStore(),
				handler.QueueHandler[testRecord]{Store: mockStore, RecordTransformer: test.transformerFunc},
			)

			router := mux.NewRouter()
			router.HandleFunc("/{queueName}", h.HandleDequeue)

			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(mockStore, jobTokenStore)
			}

			router.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatusCode, rw.Code)

			b, err := io.ReadAll(rw.Body)
			require.NoError(t, err)

			if len(test.expectedResponseBody) > 0 {
				assert.JSONEq(t, test.expectedResponseBody, string(b))
			} else {
				assert.Empty(t, string(b))
			}

			if test.assertionFunc != nil {
				test.assertionFunc(t, mockStore, jobTokenStore)
			}
		})
	}
}

func TestHandler_HandleAddExecutionLogEntry(t *testing.T) {
	startTime := time.Date(2023, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name                 string
		body                 string
		mockFunc             func(mockStore *dbworkerstoremocks.MockStore[testRecord])
		expectedStatusCode   int
		expectedResponseBody string
		assertionFunc        func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord])
	}{
		{
			name: "Add execution log entry",
			body: fmt.Sprintf(`{"executorName": "test-executor", "jobId": 42, "key": "foo", "command": ["faz", "baz"], "startTime": "%s", "exitCode": 0, "out": "done", "durationMs":100}`, startTime.Format(time.RFC3339)),
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				mockStore.AddExecutionLogEntryFunc.PushReturn(10, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `10`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, mockStore.AddExecutionLogEntryFunc.History(), 1)
				assert.Equal(t, 42, mockStore.AddExecutionLogEntryFunc.History()[0].Arg1)
				assert.Equal(
					t,
					internalexecutor.ExecutionLogEntry{
						Key:        "foo",
						Command:    []string{"faz", "baz"},
						StartTime:  startTime,
						ExitCode:   pointers.Ptr(0),
						Out:        "done",
						DurationMs: pointers.Ptr(100),
					},
					mockStore.AddExecutionLogEntryFunc.History()[0].Arg2,
				)
				assert.Equal(
					t,
					dbworkerstore.ExecutionLogEntryOptions{WorkerHostname: "test-executor", State: "processing"},
					mockStore.AddExecutionLogEntryFunc.History()[0].Arg3,
				)
			},
		},
		{
			name: "Log entry not added",
			body: fmt.Sprintf(`{"executorName": "test-executor", "jobId": 42, "key": "foo", "command": ["faz", "baz"], "startTime": "%s", "exitCode": 0, "out": "done", "durationMs":100}`, startTime.Format(time.RFC3339)),
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				mockStore.AddExecutionLogEntryFunc.PushReturn(0, errors.New("failed to add"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"dbworkerstore.AddExecutionLogEntry: failed to add"}`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, mockStore.AddExecutionLogEntryFunc.History(), 1)
			},
		},
		{
			name: "Unknown job",
			body: fmt.Sprintf(`{"executorName": "test-executor", "jobId": 42, "key": "foo", "command": ["faz", "baz"], "startTime": "%s", "exitCode": 0, "out": "done", "durationMs":100}`, startTime.Format(time.RFC3339)),
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				mockStore.AddExecutionLogEntryFunc.PushReturn(0, dbworkerstore.ErrExecutionLogEntryNotUpdated)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"unknown job"}`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, mockStore.AddExecutionLogEntryFunc.History(), 1)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStore := dbworkerstoremocks.NewMockStore[testRecord]()

			h := handler.NewHandler(
				dbmocks.NewMockExecutorStore(),
				executorstore.NewMockJobTokenStore(),
				metricsstore.NewMockDistributedStore(),
				handler.QueueHandler[testRecord]{Store: mockStore},
			)

			router := mux.NewRouter()
			router.HandleFunc("/{queueName}", h.HandleAddExecutionLogEntry)

			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(mockStore)
			}

			router.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatusCode, rw.Code)

			b, err := io.ReadAll(rw.Body)
			require.NoError(t, err)

			if len(test.expectedResponseBody) > 0 {
				assert.JSONEq(t, test.expectedResponseBody, string(b))
			} else {
				assert.Empty(t, string(b))
			}

			if test.assertionFunc != nil {
				test.assertionFunc(t, mockStore)
			}
		})
	}
}

func TestHandler_HandleUpdateExecutionLogEntry(t *testing.T) {
	startTime := time.Date(2023, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name                 string
		body                 string
		mockFunc             func(mockStore *dbworkerstoremocks.MockStore[testRecord])
		expectedStatusCode   int
		expectedResponseBody string
		assertionFunc        func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord])
	}{
		{
			name: "Update execution log entry",
			body: fmt.Sprintf(`{"entryId": 10, "executorName": "test-executor", "jobId": 42, "key": "foo", "command": ["faz", "baz"], "startTime": "%s", "exitCode": 0, "out": "done", "durationMs":100}`, startTime.Format(time.RFC3339)),
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				mockStore.UpdateExecutionLogEntryFunc.PushReturn(nil)
			},
			expectedStatusCode: http.StatusNoContent,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, mockStore.UpdateExecutionLogEntryFunc.History(), 1)
				assert.Equal(t, 42, mockStore.UpdateExecutionLogEntryFunc.History()[0].Arg1)
				assert.Equal(t, 10, mockStore.UpdateExecutionLogEntryFunc.History()[0].Arg2)
				assert.Equal(
					t,
					internalexecutor.ExecutionLogEntry{
						Key:        "foo",
						Command:    []string{"faz", "baz"},
						StartTime:  startTime,
						ExitCode:   pointers.Ptr(0),
						Out:        "done",
						DurationMs: pointers.Ptr(100),
					},
					mockStore.UpdateExecutionLogEntryFunc.History()[0].Arg3,
				)
				assert.Equal(
					t,
					dbworkerstore.ExecutionLogEntryOptions{WorkerHostname: "test-executor", State: "processing"},
					mockStore.UpdateExecutionLogEntryFunc.History()[0].Arg4,
				)
			},
		},
		{
			name: "Log entry not updated",
			body: fmt.Sprintf(`{"entryId": 10, "executorName": "test-executor", "jobId": 42, "key": "foo", "command": ["faz", "baz"], "startTime": "%s", "exitCode": 0, "out": "done", "durationMs":100}`, startTime.Format(time.RFC3339)),
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				mockStore.UpdateExecutionLogEntryFunc.PushReturn(errors.New("failed to update"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"dbworkerstore.UpdateExecutionLogEntry: failed to update"}`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, mockStore.UpdateExecutionLogEntryFunc.History(), 1)
			},
		},
		{
			name: "Unknown job",
			body: fmt.Sprintf(`{"entryId": 10, "executorName": "test-executor", "jobId": 42, "key": "foo", "command": ["faz", "baz"], "startTime": "%s", "exitCode": 0, "out": "done", "durationMs":100}`, startTime.Format(time.RFC3339)),
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				mockStore.UpdateExecutionLogEntryFunc.PushReturn(dbworkerstore.ErrExecutionLogEntryNotUpdated)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"unknown job"}`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, mockStore.UpdateExecutionLogEntryFunc.History(), 1)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStore := dbworkerstoremocks.NewMockStore[testRecord]()

			h := handler.NewHandler(
				dbmocks.NewMockExecutorStore(),
				executorstore.NewMockJobTokenStore(),
				metricsstore.NewMockDistributedStore(),
				handler.QueueHandler[testRecord]{Store: mockStore},
			)

			router := mux.NewRouter()
			router.HandleFunc("/{queueName}", h.HandleUpdateExecutionLogEntry)

			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(mockStore)
			}

			router.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatusCode, rw.Code)

			b, err := io.ReadAll(rw.Body)
			require.NoError(t, err)

			if len(test.expectedResponseBody) > 0 {
				assert.JSONEq(t, test.expectedResponseBody, string(b))
			} else {
				assert.Empty(t, string(b))
			}

			if test.assertionFunc != nil {
				test.assertionFunc(t, mockStore)
			}
		})
	}
}

func TestHandler_HandleMarkComplete(t *testing.T) {
	tests := []struct {
		name                 string
		body                 string
		mockFunc             func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore)
		expectedStatusCode   int
		expectedResponseBody string
		assertionFunc        func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore)
	}{
		{
			name: "Mark complete",
			body: `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MarkCompleteFunc.PushReturn(true, nil)
				tokenStore.DeleteFunc.PushReturn(nil)
			},
			expectedStatusCode: http.StatusNoContent,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MarkCompleteFunc.History(), 1)
				assert.Equal(t, 42, mockStore.MarkCompleteFunc.History()[0].Arg1)
				assert.Equal(t, dbworkerstore.MarkFinalOptions{WorkerHostname: "test-executor"}, mockStore.MarkCompleteFunc.History()[0].Arg2)
				require.Len(t, tokenStore.DeleteFunc.History(), 1)
				assert.Equal(t, 42, tokenStore.DeleteFunc.History()[0].Arg1)
				assert.Equal(t, "test", tokenStore.DeleteFunc.History()[0].Arg2)
			},
		},
		{
			name: "Failed to mark complete",
			body: `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MarkCompleteFunc.PushReturn(false, errors.New("failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"dbworkerstore.MarkComplete: failed"}`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MarkCompleteFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 0)
			},
		},
		{
			name: "Unknown job",
			body: `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MarkCompleteFunc.PushReturn(false, nil)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `null`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MarkCompleteFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 0)
			},
		},
		{
			name: "Failed to delete job token",
			body: `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MarkCompleteFunc.PushReturn(true, nil)
				tokenStore.DeleteFunc.PushReturn(errors.New("failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"jobTokenStore.Delete: failed"}`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MarkCompleteFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 1)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStore := dbworkerstoremocks.NewMockStore[testRecord]()
			tokenStore := executorstore.NewMockJobTokenStore()

			h := handler.NewHandler(
				dbmocks.NewMockExecutorStore(),
				tokenStore,
				metricsstore.NewMockDistributedStore(),
				handler.QueueHandler[testRecord]{Store: mockStore},
			)

			router := mux.NewRouter()
			router.HandleFunc("/{queueName}", h.HandleMarkComplete)

			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(mockStore, tokenStore)
			}

			router.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatusCode, rw.Code)

			b, err := io.ReadAll(rw.Body)
			require.NoError(t, err)

			if len(test.expectedResponseBody) > 0 {
				assert.JSONEq(t, test.expectedResponseBody, string(b))
			} else {
				assert.Empty(t, string(b))
			}

			if test.assertionFunc != nil {
				test.assertionFunc(t, mockStore, tokenStore)
			}
		})
	}
}

func TestHandler_HandleMarkErrored(t *testing.T) {
	tests := []struct {
		name                 string
		body                 string
		mockFunc             func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore)
		expectedStatusCode   int
		expectedResponseBody string
		assertionFunc        func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore)
	}{
		{
			name: "Mark errored",
			body: `{"executorName": "test-executor", "jobId": 42, "errorMessage": "it failed"}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MarkErroredFunc.PushReturn(true, nil)
				tokenStore.DeleteFunc.PushReturn(nil)
			},
			expectedStatusCode: http.StatusNoContent,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MarkErroredFunc.History(), 1)
				assert.Equal(t, 42, mockStore.MarkErroredFunc.History()[0].Arg1)
				assert.Equal(t, "it failed", mockStore.MarkErroredFunc.History()[0].Arg2)
				assert.Equal(t, dbworkerstore.MarkFinalOptions{WorkerHostname: "test-executor"}, mockStore.MarkErroredFunc.History()[0].Arg3)
				require.Len(t, tokenStore.DeleteFunc.History(), 1)
				assert.Equal(t, 42, tokenStore.DeleteFunc.History()[0].Arg1)
				assert.Equal(t, "test", tokenStore.DeleteFunc.History()[0].Arg2)
			},
		},
		{
			name: "Failed to mark errored",
			body: `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MarkErroredFunc.PushReturn(false, errors.New("failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"dbworkerstore.MarkErrored: failed"}`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MarkErroredFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 0)
			},
		},
		{
			name: "Unknown job",
			body: `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MarkErroredFunc.PushReturn(false, nil)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `null`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MarkErroredFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 0)
			},
		},
		{
			name: "Failed to delete job token",
			body: `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MarkErroredFunc.PushReturn(true, nil)
				tokenStore.DeleteFunc.PushReturn(errors.New("failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"jobTokenStore.Delete: failed"}`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MarkErroredFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 1)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStore := dbworkerstoremocks.NewMockStore[testRecord]()
			tokenStore := executorstore.NewMockJobTokenStore()

			h := handler.NewHandler(
				dbmocks.NewMockExecutorStore(),
				tokenStore,
				metricsstore.NewMockDistributedStore(),
				handler.QueueHandler[testRecord]{Store: mockStore},
			)

			router := mux.NewRouter()
			router.HandleFunc("/{queueName}", h.HandleMarkErrored)

			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(mockStore, tokenStore)
			}

			router.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatusCode, rw.Code)

			b, err := io.ReadAll(rw.Body)
			require.NoError(t, err)

			if len(test.expectedResponseBody) > 0 {
				assert.JSONEq(t, test.expectedResponseBody, string(b))
			} else {
				assert.Empty(t, string(b))
			}

			if test.assertionFunc != nil {
				test.assertionFunc(t, mockStore, tokenStore)
			}
		})
	}
}

func TestHandler_HandleMarkFailed(t *testing.T) {
	tests := []struct {
		name                 string
		body                 string
		mockFunc             func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore)
		expectedStatusCode   int
		expectedResponseBody string
		assertionFunc        func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore)
	}{
		{
			name: "Mark failed",
			body: `{"executorName": "test-executor", "jobId": 42, "errorMessage": "it failed"}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MarkFailedFunc.PushReturn(true, nil)
				tokenStore.DeleteFunc.PushReturn(nil)
			},
			expectedStatusCode: http.StatusNoContent,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MarkFailedFunc.History(), 1)
				assert.Equal(t, 42, mockStore.MarkFailedFunc.History()[0].Arg1)
				assert.Equal(t, "it failed", mockStore.MarkFailedFunc.History()[0].Arg2)
				assert.Equal(t, dbworkerstore.MarkFinalOptions{WorkerHostname: "test-executor"}, mockStore.MarkFailedFunc.History()[0].Arg3)
				require.Len(t, tokenStore.DeleteFunc.History(), 1)
				assert.Equal(t, 42, tokenStore.DeleteFunc.History()[0].Arg1)
				assert.Equal(t, "test", tokenStore.DeleteFunc.History()[0].Arg2)
			},
		},
		{
			name: "Failed to mark failed",
			body: `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MarkFailedFunc.PushReturn(false, errors.New("failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"dbworkerstore.MarkFailed: failed"}`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MarkFailedFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 0)
			},
		},
		{
			name: "Unknown job",
			body: `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MarkErroredFunc.PushReturn(false, nil)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `null`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MarkFailedFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 0)
			},
		},
		{
			name: "Failed to delete job token",
			body: `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				mockStore.MarkFailedFunc.PushReturn(true, nil)
				tokenStore.DeleteFunc.PushReturn(errors.New("failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"jobTokenStore.Delete: failed"}`,
			assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], tokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, mockStore.MarkFailedFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 1)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStore := dbworkerstoremocks.NewMockStore[testRecord]()
			tokenStore := executorstore.NewMockJobTokenStore()

			h := handler.NewHandler(
				dbmocks.NewMockExecutorStore(),
				tokenStore,
				metricsstore.NewMockDistributedStore(),
				handler.QueueHandler[testRecord]{Store: mockStore},
			)

			router := mux.NewRouter()
			router.HandleFunc("/{queueName}", h.HandleMarkFailed)

			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(mockStore, tokenStore)
			}

			router.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatusCode, rw.Code)

			b, err := io.ReadAll(rw.Body)
			require.NoError(t, err)

			if len(test.expectedResponseBody) > 0 {
				assert.JSONEq(t, test.expectedResponseBody, string(b))
			} else {
				assert.Empty(t, string(b))
			}

			if test.assertionFunc != nil {
				test.assertionFunc(t, mockStore, tokenStore)
			}
		})
	}
}

func TestHandler_HandleHeartbeat(t *testing.T) {
	tests := []struct {
		name                 string
		body                 string
		mockFunc             func(metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord])
		expectedStatusCode   int
		expectedResponseBody string
		assertionFunc        func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord])
	}{
		{
			name: "V2 Heartbeat number IDs",
			body: `{"version":"V2", "executorName": "test-executor", "jobIds": [42, 7], "os": "test-os", "architecture": "test-arch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				executorStore.UpsertHeartbeatFunc.PushReturn(nil)
				mockStore.HeartbeatFunc.PushReturn([]string{"42", "7"}, nil, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"knownIds":["42","7"],"cancelIds":null}`,
			assertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, executorStore.UpsertHeartbeatFunc.History(), 1)
				require.Len(t, mockStore.HeartbeatFunc.History(), 1)
			},
		},
		{
			name: "V2 Heartbeat",
			body: `{"version":"V2", "executorName": "test-executor", "jobIds": ["42", "7"], "os": "test-os", "architecture": "test-arch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				executorStore.UpsertHeartbeatFunc.PushReturn(nil)
				mockStore.HeartbeatFunc.PushReturn([]string{"42", "7"}, nil, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"knownIds":["42","7"],"cancelIds":null}`,
			assertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, executorStore.UpsertHeartbeatFunc.History(), 1)
				require.Len(t, mockStore.HeartbeatFunc.History(), 1)
			},
		},
		{
			name:                 "Invalid worker hostname",
			body:                 `{"executorName": "", "jobIds": ["42", "7"], "os": "test-os", "architecture": "test-arch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"worker hostname cannot be empty"}`,
			assertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, executorStore.UpsertHeartbeatFunc.History(), 0)
				require.Len(t, mockStore.HeartbeatFunc.History(), 0)
			},
		},
		{
			name: "Failed to upsert heartbeat",
			body: `{"executorName": "test-executor", "jobIds": ["42", "7"], "os": "test-os", "architecture": "test-arch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				executorStore.UpsertHeartbeatFunc.PushReturn(errors.New("failed"))
				mockStore.HeartbeatFunc.PushReturn([]string{"42", "7"}, nil, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"knownIds":["42","7"],"cancelIds":null}`,
			assertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, executorStore.UpsertHeartbeatFunc.History(), 1)
				require.Len(t, mockStore.HeartbeatFunc.History(), 1)
			},
		},
		{
			name: "Failed to heartbeat",
			body: `{"executorName": "test-executor", "jobIds": ["42", "7"], "os": "test-os", "architecture": "test-arch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				executorStore.UpsertHeartbeatFunc.PushReturn(nil)
				mockStore.HeartbeatFunc.PushReturn(nil, nil, errors.New("failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"dbworkerstore.UpsertHeartbeat: failed"}`,
			assertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, executorStore.UpsertHeartbeatFunc.History(), 1)
				require.Len(t, mockStore.HeartbeatFunc.History(), 1)
			},
		},
		{
			name: "V2 has cancelled ids",
			body: `{"version": "V2", "executorName": "test-executor", "jobIds": ["42", "7"], "os": "test-os", "architecture": "test-arch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				executorStore.UpsertHeartbeatFunc.PushReturn(nil)
				mockStore.HeartbeatFunc.PushReturn(nil, []string{"42", "7"}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"knownIds":null,"cancelIds":["42","7"]}`,
			assertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *dbmocks.MockExecutorStore, mockStore *dbworkerstoremocks.MockStore[testRecord]) {
				require.Len(t, executorStore.UpsertHeartbeatFunc.History(), 1)
				require.Len(t, mockStore.HeartbeatFunc.History(), 1)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStore := dbworkerstoremocks.NewMockStore[testRecord]()
			executorStore := dbmocks.NewMockExecutorStore()
			metricsStore := metricsstore.NewMockDistributedStore()

			h := handler.NewHandler(
				executorStore,
				executorstore.NewMockJobTokenStore(),
				metricsStore,
				handler.QueueHandler[testRecord]{Store: mockStore},
			)

			router := mux.NewRouter()
			router.HandleFunc("/{queueName}", h.HandleHeartbeat)

			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(metricsStore, executorStore, mockStore)
			}

			router.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatusCode, rw.Code)

			b, err := io.ReadAll(rw.Body)
			require.NoError(t, err)

			if len(test.expectedResponseBody) > 0 {
				assert.JSONEq(t, test.expectedResponseBody, string(b))
			} else {
				assert.Empty(t, string(b))
			}

			if test.assertionFunc != nil {
				test.assertionFunc(t, metricsStore, executorStore, mockStore)
			}
		})
	}
}

type testRecord struct {
	id int
}

func (r testRecord) RecordID() int { return r.id }

func (r testRecord) RecordUID() string {
	return strconv.Itoa(r.id)
}
