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

type dequeueTestCase struct {
	name string
	body string

	// If this is set, we expect all queues to be empty - otherwise, it's configured on a specific dequeue event
	// only valid status code for this field is http.StatusNoContent
	expectedStatusCode int

	// TODO: due to generics, I'm not sure how to provide a single list of sequential dequeues.
	// Without fairness, these could just be evaluated in the order of the queues as provided in the POST body,
	// but when fairness comes into play, that will no longer apply. To circumvent this, I add the events with an ID
	// to determine in which order they should be evaluated. Should be revisited
	mockFunc      func(codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore)
	assertionFunc func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore)

	codeintelDequeueEvents map[int]dequeueEvent[uploadsshared.Index]
	batchesDequeueEvents   map[int]dequeueEvent[*btypes.BatchSpecWorkspaceExecutionJob]
	totalEvents            int
}

func TestMultiHandler_HandleDequeue(t *testing.T) {
	tests := []dequeueTestCase{
		{
			name: "Dequeue one record for each queue",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB", "queues": ["codeintel", "batches"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				codeintelMockStore.DequeueFunc.PushReturn(uploadsshared.Index{ID: 1}, true, nil)
				jobTokenStore.CreateFunc.PushReturn("token1", nil)
				batchesMockStore.DequeueFunc.PushReturn(&btypes.BatchSpecWorkspaceExecutionJob{ID: 2}, true, nil)
				jobTokenStore.CreateFunc.PushReturn("token2", nil)
			},
			assertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.CreateFunc.History(), 2)

				// codeintel dequeue gets called twice: the queues are handled in order of the value in the request body.
				// Once a queue is empty, the next in line gets dequeued. The second call returns an empty job.
				// TODO: this could break when fairness is implemented.
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 2)
				assert.Equal(t, "test-executor", codeintelMockStore.DequeueFunc.History()[0].Arg1)
				assert.Nil(t, codeintelMockStore.DequeueFunc.History()[0].Arg2)
				assert.Equal(t, 1, jobTokenStore.CreateFunc.History()[0].Arg1)
				assert.Equal(t, "codeintel", jobTokenStore.CreateFunc.History()[0].Arg2)

				require.Len(t, batchesMockStore.DequeueFunc.History(), 1)
				assert.Equal(t, "test-executor", batchesMockStore.DequeueFunc.History()[0].Arg1)
				assert.Nil(t, batchesMockStore.DequeueFunc.History()[0].Arg2)
				assert.Equal(t, 2, jobTokenStore.CreateFunc.History()[1].Arg1)
				assert.Equal(t, "batches", jobTokenStore.CreateFunc.History()[1].Arg2)
			},
			codeintelDequeueEvents: map[int]dequeueEvent[uploadsshared.Index]{
				0: {
					queueName:            "codeintel",
					expectedStatusCode:   http.StatusOK,
					expectedResponseBody: `{"id":1,"token":"token1","queue":"codeintel","repositoryName":"","repositoryDirectory":"","commit":"","fetchTags":false,"shallowClone":false,"sparseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redactedValues":null}`,
				},
			},
			batchesDequeueEvents: map[int]dequeueEvent[*btypes.BatchSpecWorkspaceExecutionJob]{
				1: {
					queueName:            "batches",
					expectedStatusCode:   http.StatusOK,
					expectedResponseBody: `{"id":2,"token":"token2","queue":"batches","repositoryName":"","repositoryDirectory":"","commit":"","fetchTags":false,"shallowClone":false,"sparseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redactedValues":null}`,
				},
			},
			totalEvents: 2,
		},
		{
			name: "Dequeue only codeintel record when requesting codeintel queue and batches record exists",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB", "queues": ["codeintel"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				codeintelMockStore.DequeueFunc.PushReturn(uploadsshared.Index{ID: 1}, true, nil)
				batchesMockStore.DequeueFunc.PushReturn(&btypes.BatchSpecWorkspaceExecutionJob{ID: 2}, true, nil)
				jobTokenStore.CreateFunc.PushReturn("token1", nil)
			},
			assertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.CreateFunc.History(), 1)

				// The batches dequeue call is made to verify that a batch job isn't returned when the queue is not provided in the request body.
				// This yields another invocation of dequeue on codeintel, hence a length of 2. The last returns an empty job as the queue is empty.
				// TODO: this could break when fairness is implemented.
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 2)
				assert.Equal(t, "test-executor", codeintelMockStore.DequeueFunc.History()[0].Arg1)
				assert.Nil(t, codeintelMockStore.DequeueFunc.History()[0].Arg2)
				assert.Equal(t, 1, jobTokenStore.CreateFunc.History()[0].Arg1)
				assert.Equal(t, "codeintel", jobTokenStore.CreateFunc.History()[0].Arg2)

				require.Len(t, batchesMockStore.DequeueFunc.History(), 0)
			},
			codeintelDequeueEvents: map[int]dequeueEvent[uploadsshared.Index]{
				0: {
					queueName:            "codeintel",
					expectedStatusCode:   http.StatusOK,
					expectedResponseBody: `{"id":1,"token":"token1","queue":"codeintel","repositoryName":"","repositoryDirectory":"","commit":"","fetchTags":false,"shallowClone":false,"sparseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redactedValues":null}`,
				},
			},
			batchesDequeueEvents: map[int]dequeueEvent[*btypes.BatchSpecWorkspaceExecutionJob]{
				1: {
					queueName:          "batches",
					expectedStatusCode: http.StatusNoContent,
				},
			},
			totalEvents: 2,
		},
		{
			name: "Dequeue only codeintel record when requesting both queues and batches record doesn't exists",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB", "queues": ["codeintel", "batches"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				codeintelMockStore.DequeueFunc.PushReturn(uploadsshared.Index{ID: 1}, true, nil)
				batchesMockStore.DequeueFunc.PushReturn(&btypes.BatchSpecWorkspaceExecutionJob{}, false, nil)
				jobTokenStore.CreateFunc.PushReturn("token1", nil)
			},
			assertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.CreateFunc.History(), 1)

				// codeintel dequeue gets called twice: the queues are handled in order of the value in the request body.
				// Once a queue is empty, the next in line gets dequeued. The second call returns an empty job.
				// TODO: this could break when fairness is implemented.
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 2)
				assert.Equal(t, "test-executor", codeintelMockStore.DequeueFunc.History()[0].Arg1)
				assert.Nil(t, codeintelMockStore.DequeueFunc.History()[0].Arg2)
				assert.Equal(t, 1, jobTokenStore.CreateFunc.History()[0].Arg1)
				assert.Equal(t, "codeintel", jobTokenStore.CreateFunc.History()[0].Arg2)

				require.Len(t, batchesMockStore.DequeueFunc.History(), 1)
			},
			codeintelDequeueEvents: map[int]dequeueEvent[uploadsshared.Index]{
				0: {
					queueName:            "codeintel",
					expectedStatusCode:   http.StatusOK,
					expectedResponseBody: `{"id":1,"token":"token1","queue":"codeintel","repositoryName":"","repositoryDirectory":"","commit":"","fetchTags":false,"shallowClone":false,"sparseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redactedValues":null}`,
				},
			},
			batchesDequeueEvents: map[int]dequeueEvent[*btypes.BatchSpecWorkspaceExecutionJob]{
				1: {
					queueName:          "batches",
					expectedStatusCode: http.StatusNoContent,
				},
			},
			totalEvents: 2,
		},
		{
			name: "Nothing to dequeue",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB", "queues": ["codeintel","batches"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				codeintelMockStore.DequeueFunc.PushReturn(uploadsshared.Index{}, false, nil)
				batchesMockStore.DequeueFunc.PushReturn(&btypes.BatchSpecWorkspaceExecutionJob{}, false, nil)
			},
			assertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 1)
				require.Len(t, batchesMockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CreateFunc.History(), 0)
			},
			expectedStatusCode: http.StatusNoContent,
			totalEvents:        0,
		},
		{
			name: "Invalid version",
			body: `{"executorName": "test-executor", "version":"\n1.2", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["codeintel","batches"]}`,
			assertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 0)
				require.Len(t, jobTokenStore.CreateFunc.History(), 0)
			},
			codeintelDequeueEvents: map[int]dequeueEvent[uploadsshared.Index]{
				0: {
					expectedStatusCode:   http.StatusInternalServerError,
					expectedResponseBody: `{"error":"Invalid Semantic Version"}`,
				},
			},
			totalEvents: 1,
		},
		{
			name: "Dequeue error codeintel",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["codeintel"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				codeintelMockStore.DequeueFunc.PushReturn(uploadsshared.Index{}, false, errors.New("failed to dequeue"))
			},
			assertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 1)
				require.Len(t, batchesMockStore.DequeueFunc.History(), 0)
				require.Len(t, jobTokenStore.CreateFunc.History(), 0)
			},
			codeintelDequeueEvents: map[int]dequeueEvent[uploadsshared.Index]{
				0: {
					expectedStatusCode:   http.StatusInternalServerError,
					expectedResponseBody: `{"error":"dbworkerstore.Dequeue codeintel: failed to dequeue"}`,
				},
			},
			totalEvents: 1,
		},
		{
			name: "Dequeue error batches",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["batches"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				batchesMockStore.DequeueFunc.PushReturn(&btypes.BatchSpecWorkspaceExecutionJob{}, false, errors.New("failed to dequeue"))
			},
			assertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 0)
				require.Len(t, batchesMockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CreateFunc.History(), 0)
			},
			batchesDequeueEvents: map[int]dequeueEvent[*btypes.BatchSpecWorkspaceExecutionJob]{
				0: {
					expectedStatusCode:   http.StatusInternalServerError,
					expectedResponseBody: `{"error":"dbworkerstore.Dequeue batches: failed to dequeue"}`,
				},
			},
			totalEvents: 1,
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
		{
			name: "Failed to mark record as failed",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["codeintel","batches"]}`,
			codeintelDequeueEvents: map[int]dequeueEvent[uploadsshared.Index]{
				0: {
					expectedStatusCode:   http.StatusInternalServerError,
					expectedResponseBody: `{"error":"RecordTransformer codeintel: 2 errors occurred:\n\t* failed\n\t* failed to mark"}`,
					mockFunc: func(mockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], jobTokenStore *executorstore.MockJobTokenStore) {
						mockStore.DequeueFunc.PushReturn(uploadsshared.Index{ID: 1}, true, nil)
						mockStore.MarkFailedFunc.PushReturn(true, errors.New("failed to mark"))
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
					expectedResponseBody: `{"error":"RecordTransformer batches: 2 errors occurred:\n\t* failed\n\t* failed to mark"}`,
					mockFunc: func(mockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
						mockStore.DequeueFunc.PushReturn(&btypes.BatchSpecWorkspaceExecutionJob{ID: 1}, true, nil)
						mockStore.MarkFailedFunc.PushReturn(true, errors.New("failed to mark"))
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
		{
			name: "Failed to create job token",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["codeintel","batches"]}`,
			codeintelDequeueEvents: map[int]dequeueEvent[uploadsshared.Index]{
				0: {
					expectedStatusCode:   http.StatusInternalServerError,
					expectedResponseBody: `{"error":"CreateToken: failed to create token"}`,
					mockFunc: func(mockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], jobTokenStore *executorstore.MockJobTokenStore) {
						mockStore.DequeueFunc.PushReturn(uploadsshared.Index{ID: 1}, true, nil)
						jobTokenStore.CreateFunc.PushReturn("", errors.New("failed to create token"))
					},
					transformerFunc: func(ctx context.Context, version string, record uploadsshared.Index, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
						return executortypes.Job{ID: record.RecordID()}, nil
					},
					assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], jobTokenStore *executorstore.MockJobTokenStore) {
						require.Len(t, mockStore.DequeueFunc.History(), 1)
						require.Len(t, jobTokenStore.CreateFunc.History(), 1)
						require.Len(t, jobTokenStore.RegenerateFunc.History(), 0)
					},
				},
			},
			totalEvents: 1,
		},
		{
			name: "Job token already exists",
			body: `{"executorName": "test-executor","numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["codeintel","batches"]}`,
			codeintelDequeueEvents: map[int]dequeueEvent[uploadsshared.Index]{
				0: {
					expectedStatusCode:   http.StatusOK,
					expectedResponseBody: `{"id":1,"token":"somenewtoken","queue":"codeintel", "repositoryName":"","repositoryDirectory":"","commit":"","fetchTags":false,"shallowClone":false,"sparseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redactedValues":null}`,
					mockFunc: func(mockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], jobTokenStore *executorstore.MockJobTokenStore) {
						mockStore.DequeueFunc.PushReturn(uploadsshared.Index{ID: 1}, true, nil)
						jobTokenStore.CreateFunc.PushReturn("", executorstore.ErrJobTokenAlreadyCreated)
						jobTokenStore.RegenerateFunc.PushReturn("somenewtoken", nil)
					},
					transformerFunc: func(ctx context.Context, version string, record uploadsshared.Index, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
						return executortypes.Job{ID: record.RecordID()}, nil
					},
					assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], jobTokenStore *executorstore.MockJobTokenStore) {
						require.Len(t, mockStore.DequeueFunc.History(), 1)
						require.Len(t, jobTokenStore.CreateFunc.History(), 1)
						require.Len(t, jobTokenStore.RegenerateFunc.History(), 1)
						assert.Equal(t, 1, jobTokenStore.RegenerateFunc.History()[0].Arg1)
						assert.Equal(t, "codeintel", jobTokenStore.RegenerateFunc.History()[0].Arg2)
					},
				},
			},
			totalEvents: 1,
		},
		{
			name: "Failed to regenerate token",
			body: `{"executorName": "test-executor","numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["codeintel","batches"]}`,
			codeintelDequeueEvents: map[int]dequeueEvent[uploadsshared.Index]{
				0: {
					expectedStatusCode:   http.StatusInternalServerError,
					expectedResponseBody: `{"error":"RegenerateToken: failed to regen token"}`,
					mockFunc: func(mockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], jobTokenStore *executorstore.MockJobTokenStore) {
						mockStore.DequeueFunc.PushReturn(uploadsshared.Index{ID: 1}, true, nil)
						jobTokenStore.CreateFunc.PushReturn("", executorstore.ErrJobTokenAlreadyCreated)
						jobTokenStore.RegenerateFunc.PushReturn("", errors.New("failed to regen token"))
					},
					transformerFunc: func(ctx context.Context, version string, record uploadsshared.Index, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
						return executortypes.Job{ID: record.RecordID()}, nil
					},
					assertionFunc: func(t *testing.T, mockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], jobTokenStore *executorstore.MockJobTokenStore) {
						require.Len(t, mockStore.DequeueFunc.History(), 1)
						require.Len(t, jobTokenStore.CreateFunc.History(), 1)
						require.Len(t, jobTokenStore.RegenerateFunc.History(), 1)
					},
				},
			},
			totalEvents: 1,
		},
	}
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

			if test.mockFunc != nil {
				test.mockFunc(codeIntelMockStore, batchesMockStore, jobTokenStore)
			}

			if test.expectedStatusCode != 0 {
				evaluateEvent(test.body, test.expectedStatusCode, "", t, router)
			}

			for eventIndex := 0; eventIndex < test.totalEvents; eventIndex++ {
				if _, ok := test.codeintelDequeueEvents[eventIndex]; ok {
					event := test.codeintelDequeueEvents[eventIndex]
					if event.transformerFunc != nil {
						mh.CodeIntelQueueHandler.RecordTransformer = event.transformerFunc
					}
					evaluateEvent(test.body, event.expectedStatusCode, event.expectedResponseBody, t, router)
				} else {
					event := test.batchesDequeueEvents[eventIndex]
					if event.transformerFunc != nil {
						mh.BatchesQueueHandler.RecordTransformer = event.transformerFunc
					}
					evaluateEvent(test.body, event.expectedStatusCode, event.expectedResponseBody, t, router)
				}
			}

			if test.assertionFunc != nil {
				test.assertionFunc(t, codeIntelMockStore, batchesMockStore, jobTokenStore)
			}
		})
	}
}

func evaluateEvent(
	requestBody string,
	expectedStatusCode int,
	expectedResponseBody string,
	t *testing.T,
	router *mux.Router,
) {
	req, err := http.NewRequest(http.MethodPost, "/dequeue", strings.NewReader(requestBody))
	require.NoError(t, err)

	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)
	assert.Equal(t, expectedStatusCode, rw.Code)

	b, err := io.ReadAll(rw.Body)
	require.NoError(t, err)

	if len(expectedResponseBody) > 0 {
		assert.JSONEq(t, expectedResponseBody, string(b))
	} else {
		assert.Empty(t, string(b))
	}
}
