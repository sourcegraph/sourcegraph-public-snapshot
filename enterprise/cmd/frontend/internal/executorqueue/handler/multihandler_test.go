package handler_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/queues/multi"
	executorstore "github.com/sourcegraph/sourcegraph/enterprise/internal/executor/store"
	executortypes "github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	metricsstore "github.com/sourcegraph/sourcegraph/internal/metrics/store"
	dbworkerstoremocks "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store/mocks"
)

func TestMultiHandler_HandleDequeue(t *testing.T) {
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
			body: `{"queues": ["codeintel", "batches"], "workerHostName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
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
		//{
		//	name:                 "Invalid version",
		//	body:                 `{"executorName": "test-executor", "version":"\n1.2", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
		//	expectedStatusCode:   http.StatusInternalServerError,
		//	expectedResponseBody: `{"error":"Invalid Semantic Version"}`,
		//	assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
		//		require.Len(t, mockStore.DequeueFunc.History(), 0)
		//		require.Len(t, jobTokenStore.CreateFunc.History(), 0)
		//	},
		//},
		//{
		//	name: "Dequeue error",
		//	body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
		//	mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
		//		mockStore.DequeueFunc.PushReturn(testRecord{}, false, errors.New("failed to dequeue"))
		//	},
		//	expectedStatusCode:   http.StatusInternalServerError,
		//	expectedResponseBody: `{"error":"dbworkerstore.Dequeue: failed to dequeue"}`,
		//	assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
		//		require.Len(t, mockStore.DequeueFunc.History(), 1)
		//		require.Len(t, jobTokenStore.CreateFunc.History(), 0)
		//	},
		//},
		//{
		//	name: "Nothing to dequeue",
		//	body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
		//	mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
		//		mockStore.DequeueFunc.PushReturn(testRecord{}, false, nil)
		//	},
		//	expectedStatusCode: http.StatusNoContent,
		//	assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
		//		require.Len(t, mockStore.DequeueFunc.History(), 1)
		//		require.Len(t, jobTokenStore.CreateFunc.History(), 0)
		//	},
		//},
		//{
		//	name: "Failed to transform record",
		//	body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
		//	transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
		//		return executortypes.Job{}, errors.New("failed")
		//	},
		//	mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
		//		mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
		//		mockStore.MarkFailedFunc.PushReturn(true, nil)
		//	},
		//	expectedStatusCode:   http.StatusInternalServerError,
		//	expectedResponseBody: `{"error":"RecordTransformer: failed"}`,
		//	assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
		//		require.Len(t, mockStore.DequeueFunc.History(), 1)
		//		require.Len(t, mockStore.MarkFailedFunc.History(), 1)
		//		assert.Equal(t, 1, mockStore.MarkFailedFunc.History()[0].Arg1)
		//		assert.Equal(t, "failed to transform record: failed", mockStore.MarkFailedFunc.History()[0].Arg2)
		//		assert.Equal(t, dbworkerstore.MarkFinalOptions{}, mockStore.MarkFailedFunc.History()[0].Arg3)
		//		require.Len(t, jobTokenStore.CreateFunc.History(), 0)
		//	},
		//},
		//{
		//	name: "Failed to mark record as failed",
		//	body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
		//	transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
		//		return executortypes.Job{}, errors.New("failed")
		//	},
		//	mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
		//		mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
		//		mockStore.MarkFailedFunc.PushReturn(false, errors.New("failed to mark"))
		//	},
		//	expectedStatusCode:   http.StatusInternalServerError,
		//	expectedResponseBody: `{"error":"RecordTransformer: failed"}`,
		//	assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
		//		require.Len(t, mockStore.DequeueFunc.History(), 1)
		//		require.Len(t, mockStore.MarkFailedFunc.History(), 1)
		//		assert.Equal(t, 1, mockStore.MarkFailedFunc.History()[0].Arg1)
		//		assert.Equal(t, "failed to transform record: failed", mockStore.MarkFailedFunc.History()[0].Arg2)
		//		assert.Equal(t, dbworkerstore.MarkFinalOptions{}, mockStore.MarkFailedFunc.History()[0].Arg3)
		//		require.Len(t, jobTokenStore.CreateFunc.History(), 0)
		//	},
		//},
		//{
		//	name: "V2 job",
		//	body: `{"executorName": "test-executor", "version": "dev", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
		//	transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
		//		return executortypes.Job{ID: record.RecordID()}, nil
		//	},
		//	mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
		//		mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
		//		jobTokenStore.CreateFunc.PushReturn("sometoken", nil)
		//	},
		//	expectedStatusCode:   http.StatusOK,
		//	expectedResponseBody: `{"version":2,"id":1,"token":"sometoken","repositoryName":"","repositoryDirectory":"","commit":"","fetchTags":false,"shallowClone":false,"sparseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redactedValues":null,"dockerAuthConfig":{}}`,
		//	assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
		//		require.Len(t, mockStore.DequeueFunc.History(), 1)
		//		require.Len(t, jobTokenStore.CreateFunc.History(), 1)
		//	},
		//},
		//{
		//	name: "Failed to create job token",
		//	body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
		//	transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
		//		return executortypes.Job{ID: record.RecordID()}, nil
		//	},
		//	mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
		//		mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
		//		jobTokenStore.CreateFunc.PushReturn("", errors.New("failed to create token"))
		//	},
		//	expectedStatusCode:   http.StatusInternalServerError,
		//	expectedResponseBody: `{"error":"CreateToken: failed to create token"}`,
		//	assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
		//		require.Len(t, mockStore.DequeueFunc.History(), 1)
		//		require.Len(t, jobTokenStore.CreateFunc.History(), 1)
		//		require.Len(t, jobTokenStore.RegenerateFunc.History(), 0)
		//	},
		//},
		//{
		//	name: "Job token already exists",
		//	body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
		//	transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
		//		return executortypes.Job{ID: record.RecordID()}, nil
		//	},
		//	mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
		//		mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
		//		jobTokenStore.CreateFunc.PushReturn("", executorstore.ErrJobTokenAlreadyCreated)
		//		jobTokenStore.RegenerateFunc.PushReturn("somenewtoken", nil)
		//	},
		//	expectedStatusCode:   http.StatusOK,
		//	expectedResponseBody: `{"id":1,"token":"somenewtoken","repositoryName":"","repositoryDirectory":"","commit":"","fetchTags":false,"shallowClone":false,"sparseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redactedValues":null}`,
		//	assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
		//		require.Len(t, mockStore.DequeueFunc.History(), 1)
		//		require.Len(t, jobTokenStore.CreateFunc.History(), 1)
		//		require.Len(t, jobTokenStore.RegenerateFunc.History(), 1)
		//		assert.Equal(t, 1, jobTokenStore.RegenerateFunc.History()[0].Arg1)
		//		assert.Equal(t, "test", jobTokenStore.RegenerateFunc.History()[0].Arg2)
		//	},
		//},
		//{
		//	name: "Failed to regenerate token",
		//	body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
		//	transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
		//		return executortypes.Job{ID: record.RecordID()}, nil
		//	},
		//	mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
		//		mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
		//		jobTokenStore.CreateFunc.PushReturn("", executorstore.ErrJobTokenAlreadyCreated)
		//		jobTokenStore.RegenerateFunc.PushReturn("", errors.New("failed to regen token"))
		//	},
		//	expectedStatusCode:   http.StatusInternalServerError,
		//	expectedResponseBody: `{"error":"RegenerateToken: failed to regen token"}`,
		//	assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
		//		require.Len(t, mockStore.DequeueFunc.History(), 1)
		//		require.Len(t, jobTokenStore.CreateFunc.History(), 1)
		//		require.Len(t, jobTokenStore.RegenerateFunc.History(), 1)
		//	},
		//},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStore := dbworkerstoremocks.NewMockStore[testRecord]()
			jobTokenStore := executorstore.NewMockJobTokenStore()
			mockExecutorStore := database.NewMockExecutorStore()
			mockMetricsStore := metricsstore.NewMockDistributedStore()

			mockBatchesHandler := handler.NewHandler(
				mockExecutorStore,
				jobTokenStore,
				mockMetricsStore,
				handler.QueueHandler[testRecord]{Name: "batches", Store: mockStore, RecordTransformer: test.transformerFunc},
			)

			mockCodeIntelHandler := handler.NewHandler(
				mockExecutorStore,
				jobTokenStore,
				mockMetricsStore,
				handler.QueueHandler[testRecord]{Name: "codeintel", Store: mockStore, RecordTransformer: test.transformerFunc},
			)

			multiHandler := handler.NewMultiHandler(multi.QueueHandler(map[string]handler.ExecutorHandler{
				"batches":   mockBatchesHandler,
				"codeintel": mockCodeIntelHandler,
			}))

			router := mux.NewRouter()
			router.HandleFunc("/dequeue", multiHandler.HandleDequeue)

			req, err := http.NewRequest(http.MethodPost, "/dequeue", strings.NewReader(test.body))
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
