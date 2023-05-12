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
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	executorstore "github.com/sourcegraph/sourcegraph/enterprise/internal/executor/store"
	executortypes "github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	dbworkerstoremocks "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store/mocks"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type dequeueEvent[T workerutil.Record] struct {
	queueName            string
	expectedStatusCode   int
	expectedResponseBody string
	mockFunc             func(mockStore *dbworkerstoremocks.MockStore[T], jobTokenStore *executorstore.MockJobTokenStore)
	transformerFunc      handler.TransformerFunc[T]
	assertionFunc        func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[T], jobTokenStore *executorstore.MockJobTokenStore)
}

func transformerFunc[T workerutil.Record](ctx context.Context, version string, t T, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
	return executortypes.Job{ID: t.RecordID()}, nil
}

type noDequeueEvent struct {
	mockFunc           func(codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore)
	expectedStatusCode int
	assertionFunc      func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore)
}

type dequeueTestCase struct {
	name string
	body string

	noDequeueEvent noDequeueEvent

	// TODO: due to generics, I'm not sure how to provide a single list of sequential dequeues.
	// Without fairness, these could just be evaluated in the order of the queues as provided in the POST body,
	// but when fairness comes into play, that will no longer apply. To circumvent this, I add the events with an ID
	// to determine in which order they should be evaluated. Should be revisited
	codeintelDequeueEvents map[int]dequeueEvent[uploadsshared.Index]
	batchesDequeueEvents   map[int]dequeueEvent[*btypes.BatchSpecWorkspaceExecutionJob]
	totalEvents            int
}

func TestMultiHandler_HandleDequeue(t *testing.T) {
	tests := []dequeueTestCase{
		{
			name: "Dequeue one record for each queue",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB", "queues": ["codeintel", "batches"]}`,
			codeintelDequeueEvents: map[int]dequeueEvent[uploadsshared.Index]{
				0: {
					queueName:            "codeintel",
					expectedStatusCode:   http.StatusOK,
					expectedResponseBody: `{"id":1,"token":"token1","queue":"codeintel","repositoryName":"","repositoryDirectory":"","commit":"","fetchTags":false,"shallowClone":false,"sparseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redactedValues":null}`,
					mockFunc: func(mockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], jobTokenStore *executorstore.MockJobTokenStore) {
						mockStore.DequeueFunc.PushReturn(uploadsshared.Index{ID: 1}, true, nil)
						jobTokenStore.CreateFunc.PushReturn("token1", nil)
					},
					assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], jobTokenStore *executorstore.MockJobTokenStore) {
						require.Len(t, mockStore.DequeueFunc.History(), 1)
						assert.Equal(t, "test-executor", mockStore.DequeueFunc.History()[0].Arg1)
						assert.Nil(t, mockStore.DequeueFunc.History()[0].Arg2)
						require.Len(t, jobTokenStore.CreateFunc.History(), 1)
						assert.Equal(t, 1, jobTokenStore.CreateFunc.History()[0].Arg1)
						assert.Equal(t, "codeintel", jobTokenStore.CreateFunc.History()[0].Arg2)
					},
				},
			},
			batchesDequeueEvents: map[int]dequeueEvent[*btypes.BatchSpecWorkspaceExecutionJob]{
				1: {
					queueName: "batches",
					mockFunc: func(mockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
						mockStore.DequeueFunc.PushReturn(&btypes.BatchSpecWorkspaceExecutionJob{ID: 2}, true, nil)
						jobTokenStore.CreateFunc.PushReturn("token2", nil)
					},
					expectedStatusCode:   http.StatusOK,
					expectedResponseBody: `{"id":2,"token":"token2","queue":"batches","repositoryName":"","repositoryDirectory":"","commit":"","fetchTags":false,"shallowClone":false,"sparseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redactedValues":null}`,
					assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
						require.Len(t, mockStore.DequeueFunc.History(), 1)
						assert.Equal(t, "test-executor", mockStore.DequeueFunc.History()[0].Arg1)
						assert.Nil(t, mockStore.DequeueFunc.History()[0].Arg2)
						// shared job token store, so this is the second create invocation
						require.Len(t, jobTokenStore.CreateFunc.History(), 2)
						assert.Equal(t, 2, jobTokenStore.CreateFunc.History()[1].Arg1)
						assert.Equal(t, "batches", jobTokenStore.CreateFunc.History()[1].Arg2)
					},
				},
			},
			totalEvents: 2,
		},
		{
			name: "Nothing to dequeue",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB", "queues": ["codeintel","batches"]}`,
			noDequeueEvent: noDequeueEvent{
				mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
					codeintelMockStore.DequeueFunc.PushReturn(uploadsshared.Index{}, false, nil)
					batchesMockStore.DequeueFunc.PushReturn(&btypes.BatchSpecWorkspaceExecutionJob{}, false, nil)
				},
				expectedStatusCode: http.StatusNoContent,
				assertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
					require.Len(t, codeintelMockStore.DequeueFunc.History(), 1)
					require.Len(t, batchesMockStore.DequeueFunc.History(), 1)
					require.Len(t, jobTokenStore.CreateFunc.History(), 0)
				},
			},
			totalEvents: 0,
		},
		{
			name: "Invalid version",
			body: `{"executorName": "test-executor", "version":"\n1.2", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["codeintel","batches"]}`,
			codeintelDequeueEvents: map[int]dequeueEvent[uploadsshared.Index]{
				0: {
					expectedStatusCode:   http.StatusInternalServerError,
					expectedResponseBody: `{"error":"Invalid Semantic Version"}`,
					assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], jobTokenStore *executorstore.MockJobTokenStore) {
						require.Len(t, mockStore.DequeueFunc.History(), 0)
						require.Len(t, jobTokenStore.CreateFunc.History(), 0)
					},
				},
			},
			totalEvents: 1,
		},
		{
			name: "Dequeue error",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["codeintel","batches"]}`,
			codeintelDequeueEvents: map[int]dequeueEvent[uploadsshared.Index]{
				0: {
					expectedStatusCode:   http.StatusInternalServerError,
					expectedResponseBody: `{"error":"dbworkerstore.Dequeue codeintel: failed to dequeue"}`,
					mockFunc: func(mockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], jobTokenStore *executorstore.MockJobTokenStore) {
						mockStore.DequeueFunc.PushReturn(uploadsshared.Index{}, false, errors.New("failed to dequeue"))
					},
					assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], jobTokenStore *executorstore.MockJobTokenStore) {
						require.Len(t, mockStore.DequeueFunc.History(), 1)
						require.Len(t, jobTokenStore.CreateFunc.History(), 0)
					},
				},
			},
			batchesDequeueEvents: map[int]dequeueEvent[*btypes.BatchSpecWorkspaceExecutionJob]{
				1: {
					expectedStatusCode:   http.StatusInternalServerError,
					expectedResponseBody: `{"error":"dbworkerstore.Dequeue batches: failed to dequeue"}`,
					mockFunc: func(mockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
						mockStore.DequeueFunc.PushReturn(&btypes.BatchSpecWorkspaceExecutionJob{}, false, errors.New("failed to dequeue"))
					},
					assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
						require.Len(t, mockStore.DequeueFunc.History(), 1)
						require.Len(t, jobTokenStore.CreateFunc.History(), 0)
					},
				},
			},
			totalEvents: 2,
		},
		{
			name: "Failed to transform record",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["codeintel","batches"]}`,
			codeintelDequeueEvents: map[int]dequeueEvent[uploadsshared.Index]{
				0: {
					expectedStatusCode:   http.StatusInternalServerError,
					expectedResponseBody: `{"error":"RecordTransformer codeintel: failed"}`,
					mockFunc: func(mockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], jobTokenStore *executorstore.MockJobTokenStore) {
						mockStore.DequeueFunc.PushReturn(uploadsshared.Index{ID: 1}, true, nil)
						mockStore.MarkFailedFunc.PushReturn(true, nil)
					},
					transformerFunc: func(ctx context.Context, version string, record uploadsshared.Index, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
						return executortypes.Job{}, errors.New("failed")
					},
					assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], jobTokenStore *executorstore.MockJobTokenStore) {
						require.Len(t, mockStore.DequeueFunc.History(), 1)
						require.Len(t, mockStore.MarkFailedFunc.History(), 1)
						assert.Equal(t, 1, mockStore.MarkFailedFunc.History()[0].Arg1)
						assert.Equal(t, "failed to transform record: failed", mockStore.MarkFailedFunc.History()[0].Arg2)
						assert.Equal(t, dbworkerstore.MarkFinalOptions{}, mockStore.MarkFailedFunc.History()[0].Arg3)
						require.Len(t, jobTokenStore.CreateFunc.History(), 0)
					},
				},
			},
			batchesDequeueEvents: map[int]dequeueEvent[*btypes.BatchSpecWorkspaceExecutionJob]{
				1: {
					expectedStatusCode:   http.StatusInternalServerError,
					expectedResponseBody: `{"error":"RecordTransformer batches: failed"}`,
					mockFunc: func(mockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
						mockStore.DequeueFunc.PushReturn(&btypes.BatchSpecWorkspaceExecutionJob{ID: 1}, true, nil)
						mockStore.MarkFailedFunc.PushReturn(true, nil)
					},
					transformerFunc: func(ctx context.Context, version string, record *btypes.BatchSpecWorkspaceExecutionJob, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
						return executortypes.Job{}, errors.New("failed")
					},
					assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
						require.Len(t, mockStore.DequeueFunc.History(), 1)
						require.Len(t, mockStore.MarkFailedFunc.History(), 1)
						assert.Equal(t, 1, mockStore.MarkFailedFunc.History()[0].Arg1)
						assert.Equal(t, "failed to transform record: failed", mockStore.MarkFailedFunc.History()[0].Arg2)
						assert.Equal(t, dbworkerstore.MarkFinalOptions{}, mockStore.MarkFailedFunc.History()[0].Arg3)
						require.Len(t, jobTokenStore.CreateFunc.History(), 0)
					},
				},
			},
			totalEvents: 2,
		},
	}
	//	{
	//		name: "Failed to mark record as failed",
	//		body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
	//		transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
	//			return executortypes.Job{}, errors.New("failed")
	//		},
	//		mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
	//			mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
	//			mockStore.MarkFailedFunc.PushReturn(false, errors.New("failed to mark"))
	//		},
	//		expectedStatusCode:   http.StatusInternalServerError,
	//		expectedResponseBody: `{"error":"RecordTransformer: failed"}`,
	//		assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
	//			require.Len(t, mockStore.DequeueFunc.History(), 1)
	//			require.Len(t, mockStore.MarkFailedFunc.History(), 1)
	//			assert.Equal(t, 1, mockStore.MarkFailedFunc.History()[0].Arg1)
	//			assert.Equal(t, "failed to transform record: failed", mockStore.MarkFailedFunc.History()[0].Arg2)
	//			assert.Equal(t, dbworkerstore.MarkFinalOptions{}, mockStore.MarkFailedFunc.History()[0].Arg3)
	//			require.Len(t, jobTokenStore.CreateFunc.History(), 0)
	//		},
	//	},
	//	{
	//		name: "V2 job",
	//		body: `{"executorName": "test-executor", "version": "dev", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
	//		transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
	//			return executortypes.Job{ID: record.RecordID()}, nil
	//		},
	//		mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
	//			mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
	//			jobTokenStore.CreateFunc.PushReturn("sometoken", nil)
	//		},
	//		expectedStatusCode:   http.StatusOK,
	//		expectedResponseBody: `{"version":2,"id":1,"token":"sometoken","repositoryName":"","repositoryDirectory":"","commit":"","fetchTags":false,"shallowClone":false,"sparseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redactedValues":null,"dockerAuthConfig":{}}`,
	//		assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
	//			require.Len(t, mockStore.DequeueFunc.History(), 1)
	//			require.Len(t, jobTokenStore.CreateFunc.History(), 1)
	//		},
	//	},
	//	{
	//		name: "Failed to create job token",
	//		body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
	//		transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
	//			return executortypes.Job{ID: record.RecordID()}, nil
	//		},
	//		mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
	//			mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
	//			jobTokenStore.CreateFunc.PushReturn("", errors.New("failed to create token"))
	//		},
	//		expectedStatusCode:   http.StatusInternalServerError,
	//		expectedResponseBody: `{"error":"CreateToken: failed to create token"}`,
	//		assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
	//			require.Len(t, mockStore.DequeueFunc.History(), 1)
	//			require.Len(t, jobTokenStore.CreateFunc.History(), 1)
	//			require.Len(t, jobTokenStore.RegenerateFunc.History(), 0)
	//		},
	//	},
	//	{
	//		name: "Job token already exists",
	//		body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
	//		transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
	//			return executortypes.Job{ID: record.RecordID()}, nil
	//		},
	//		mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
	//			mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
	//			jobTokenStore.CreateFunc.PushReturn("", executorstore.ErrJobTokenAlreadyCreated)
	//			jobTokenStore.RegenerateFunc.PushReturn("somenewtoken", nil)
	//		},
	//		expectedStatusCode:   http.StatusOK,
	//		expectedResponseBody: `{"id":1,"token":"somenewtoken","repositoryName":"","repositoryDirectory":"","commit":"","fetchTags":false,"shallowClone":false,"sparseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redactedValues":null}`,
	//		assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
	//			require.Len(t, mockStore.DequeueFunc.History(), 1)
	//			require.Len(t, jobTokenStore.CreateFunc.History(), 1)
	//			require.Len(t, jobTokenStore.RegenerateFunc.History(), 1)
	//			assert.Equal(t, 1, jobTokenStore.RegenerateFunc.History()[0].Arg1)
	//			assert.Equal(t, "test", jobTokenStore.RegenerateFunc.History()[0].Arg2)
	//		},
	//	},
	//	{
	//		name: "Failed to regenerate token",
	//		body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
	//		transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
	//			return executortypes.Job{ID: record.RecordID()}, nil
	//		},
	//		mockFunc: func(mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
	//			mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
	//			jobTokenStore.CreateFunc.PushReturn("", executorstore.ErrJobTokenAlreadyCreated)
	//			jobTokenStore.RegenerateFunc.PushReturn("", errors.New("failed to regen token"))
	//		},
	//		expectedStatusCode:   http.StatusInternalServerError,
	//		expectedResponseBody: `{"error":"RegenerateToken: failed to regen token"}`,
	//		assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[testRecord], jobTokenStore *executorstore.MockJobTokenStore) {
	//			require.Len(t, mockStore.DequeueFunc.History(), 1)
	//			require.Len(t, jobTokenStore.CreateFunc.History(), 1)
	//			require.Len(t, jobTokenStore.RegenerateFunc.History(), 1)
	//		},
	//	},
	//}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jobTokenStore := executorstore.NewMockJobTokenStore()
			codeIntelMockStore := dbworkerstoremocks.NewMockStore[uploadsshared.Index]()
			batchesMockStore := dbworkerstoremocks.NewMockStore[*btypes.BatchSpecWorkspaceExecutionJob]()

			mh := handler.NewMultiHandler(
				jobTokenStore,
				handler.QueueHandler[uploadsshared.Index]{Store: codeIntelMockStore, RecordTransformer: transformerFunc[uploadsshared.Index]},
				handler.QueueHandler[*btypes.BatchSpecWorkspaceExecutionJob]{Store: batchesMockStore, RecordTransformer: transformerFunc[*btypes.BatchSpecWorkspaceExecutionJob]},
			)

			router := mux.NewRouter()
			router.HandleFunc("/dequeue", mh.ServeHTTP)

			if test.totalEvents == 0 {
				req, err := http.NewRequest(http.MethodPost, "/dequeue", strings.NewReader(test.body))
				require.NoError(t, err)

				rw := httptest.NewRecorder()

				if test.noDequeueEvent.mockFunc != nil {
					test.noDequeueEvent.mockFunc(codeIntelMockStore, batchesMockStore, jobTokenStore)
				}

				router.ServeHTTP(rw, req)
				assert.Equal(t, test.noDequeueEvent.expectedStatusCode, rw.Code)

				b, err := io.ReadAll(rw.Body)
				require.NoError(t, err)
				assert.Empty(t, string(b))

				if test.noDequeueEvent.assertionFunc != nil {
					test.noDequeueEvent.assertionFunc(t, codeIntelMockStore, batchesMockStore, jobTokenStore)
				}
			} else {
				for eventIndex := 0; eventIndex < test.totalEvents; eventIndex++ {
					if _, ok := test.codeintelDequeueEvents[eventIndex]; ok {
						event := test.codeintelDequeueEvents[eventIndex]
						if event.transformerFunc != nil {
							mh.CodeIntelQueueHandler.RecordTransformer = event.transformerFunc
						}
						evaluateEvent(test, event, codeIntelMockStore, jobTokenStore, t, router)
					} else {
						event := test.batchesDequeueEvents[eventIndex]
						if event.transformerFunc != nil {
							mh.BatchesQueueHandler.RecordTransformer = event.transformerFunc
						}
						evaluateEvent(test, event, batchesMockStore, jobTokenStore, t, router)
					}
				}
			}
		})
	}
}

func evaluateEvent[T workerutil.Record](
	test dequeueTestCase,
	event dequeueEvent[T],
	store *dbworkerstoremocks.MockStore[T],
	jobTokenStore *executorstore.MockJobTokenStore,
	t *testing.T,
	router *mux.Router,
) {
	req, err := http.NewRequest(http.MethodPost, "/dequeue", strings.NewReader(test.body))
	require.NoError(t, err)

	rw := httptest.NewRecorder()

	if event.mockFunc != nil {
		event.mockFunc(store, jobTokenStore)
	}

	router.ServeHTTP(rw, req)
	assert.Equal(t, event.expectedStatusCode, rw.Code)

	b, err := io.ReadAll(rw.Body)
	require.NoError(t, err)

	if len(event.expectedResponseBody) > 0 {
		assert.JSONEq(t, event.expectedResponseBody, string(b))
	} else {
		assert.Empty(t, string(b))
	}

	if event.assertionFunc != nil {
		event.assertionFunc(t, store, jobTokenStore)
	}
}
