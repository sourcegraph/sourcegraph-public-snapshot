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
	"github.com/sourcegraph/sourcegraph/internal/database"
	metricsstore "github.com/sourcegraph/sourcegraph/internal/metrics/store"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	dbworkerstoremocks "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store/mocks"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type MockRandom struct {
	IntnResults []int
}

func (r *MockRandom) Seed(seed int64) {}

func (r *MockRandom) Intn(n int) int {
	if len(r.IntnResults) == 0 {
		return 0
	}
	result := r.IntnResults[0]
	r.IntnResults = r.IntnResults[1:]
	return result
}

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

	mockFunc      func(codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore)
	assertionFunc func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore)

	// TODO: due to generics, I'm not sure how to provide a single list of sequential dequeues.
	// Without fairness, these could just be evaluated in the order of the queues as provided in the POST body,
	// but when fairness comes into play, that will no longer apply. To circumvent this, I add the events with an ID
	// to determine in which order they should be evaluated. Should be revisited
	codeintelDequeueEvents map[int]dequeueEvent[uploadsshared.Index]
	batchesDequeueEvents   map[int]dequeueEvent[*btypes.BatchSpecWorkspaceExecutionJob]
}

func TestMultiHandler_HandleDequeue(t *testing.T) {
	tests := []dequeueTestCase{
		{
			name: "Dequeue one record for each queue",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["codeintel", "batches"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				codeintelMockStore.DequeueFunc.PushReturn(uploadsshared.Index{ID: 1}, true, nil)
				jobTokenStore.CreateFunc.PushReturn("token1", nil)
				batchesMockStore.DequeueFunc.PushReturn(&btypes.BatchSpecWorkspaceExecutionJob{ID: 2}, true, nil)
				jobTokenStore.CreateFunc.PushReturn("token2", nil)
			},
			assertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.CreateFunc.History(), 2)

				require.Len(t, codeintelMockStore.DequeueFunc.History(), 1)
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
		},
		{
			name: "Dequeue only codeintel record when requesting codeintel queue and batches record exists",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["codeintel"]}`,
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
		},
		{
			name: "Dequeue only codeintel record when requesting both queues and batches record doesn't exists",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["codeintel", "batches"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				codeintelMockStore.DequeueFunc.PushReturn(uploadsshared.Index{ID: 1}, true, nil)
				batchesMockStore.DequeueFunc.PushReturn(&btypes.BatchSpecWorkspaceExecutionJob{}, false, nil)
				jobTokenStore.CreateFunc.PushReturn("token1", nil)
			},
			assertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, jobTokenStore.CreateFunc.History(), 1)

				// codeintel dequeue gets called twice: the second dequeue invocation mocks fairness to first dequeue the batches queue, which yields this sequence of events:
				//   1. Queue order after fairness: [codeintel, batches], dequeue codeintel -> successful dequeue
				//   2. Queue order after fairness: [batches, codeintel], dequeue batches -> no job to dequeue, dequeue codeintel -> no job to dequeue
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
		},
		{
			name: "Nothing to dequeue",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["codeintel","batches"]}`,
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
		},
		{
			name: "No queue names provided",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": []}`,
			assertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 0)
				require.Len(t, batchesMockStore.DequeueFunc.History(), 0)
				require.Len(t, jobTokenStore.CreateFunc.History(), 0)
			},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name: "Invalid queue name",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["invalidqueue"]}`,
			assertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 0)
				require.Len(t, jobTokenStore.CreateFunc.History(), 0)
			},
			codeintelDequeueEvents: map[int]dequeueEvent[uploadsshared.Index]{
				0: {
					expectedStatusCode:   http.StatusInternalServerError,
					expectedResponseBody: `{"error":"Invalid queue name(s) 'invalidqueue' found. Supported queue names are 'codeintel, batches'."}`,
				},
			},
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
		},
		{
			name: "Failed to transform record codeintel",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["codeintel"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				codeintelMockStore.DequeueFunc.PushReturn(uploadsshared.Index{ID: 1}, true, nil)
				codeintelMockStore.MarkFailedFunc.PushReturn(true, nil)
			},
			assertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 1)
				require.Len(t, codeintelMockStore.MarkFailedFunc.History(), 1)
				assert.Equal(t, 1, codeintelMockStore.MarkFailedFunc.History()[0].Arg1)
				assert.Equal(t, "failed to transform record: failed", codeintelMockStore.MarkFailedFunc.History()[0].Arg2)
				assert.Equal(t, dbworkerstore.MarkFinalOptions{}, codeintelMockStore.MarkFailedFunc.History()[0].Arg3)
				require.Len(t, jobTokenStore.CreateFunc.History(), 0)
			},
			codeintelDequeueEvents: map[int]dequeueEvent[uploadsshared.Index]{
				0: {
					expectedStatusCode:   http.StatusInternalServerError,
					expectedResponseBody: `{"error":"RecordTransformer codeintel: failed"}`,
					transformerFunc: func(ctx context.Context, version string, record uploadsshared.Index, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
						return executortypes.Job{}, errors.New("failed")
					},
				},
			},
		},
		{
			name: "Failed to transform record batches",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["batches"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				batchesMockStore.DequeueFunc.PushReturn(&btypes.BatchSpecWorkspaceExecutionJob{ID: 1}, true, nil)
				batchesMockStore.MarkFailedFunc.PushReturn(true, nil)
			},
			assertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, batchesMockStore.DequeueFunc.History(), 1)
				require.Len(t, batchesMockStore.MarkFailedFunc.History(), 1)
				assert.Equal(t, 1, batchesMockStore.MarkFailedFunc.History()[0].Arg1)
				assert.Equal(t, "failed to transform record: failed", batchesMockStore.MarkFailedFunc.History()[0].Arg2)
				assert.Equal(t, dbworkerstore.MarkFinalOptions{}, batchesMockStore.MarkFailedFunc.History()[0].Arg3)
				require.Len(t, jobTokenStore.CreateFunc.History(), 0)
			},
			batchesDequeueEvents: map[int]dequeueEvent[*btypes.BatchSpecWorkspaceExecutionJob]{
				0: {
					expectedStatusCode:   http.StatusInternalServerError,
					expectedResponseBody: `{"error":"RecordTransformer batches: failed"}`,
					transformerFunc: func(ctx context.Context, version string, record *btypes.BatchSpecWorkspaceExecutionJob, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
						return executortypes.Job{}, errors.New("failed")
					},
				},
			},
		},
		{
			name: "Failed to mark record as failed codeintel",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["codeintel"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				codeintelMockStore.DequeueFunc.PushReturn(uploadsshared.Index{ID: 1}, true, nil)
				codeintelMockStore.MarkFailedFunc.PushReturn(true, errors.New("failed to mark"))
			},
			assertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 1)
				require.Len(t, codeintelMockStore.MarkFailedFunc.History(), 1)
				assert.Equal(t, 1, codeintelMockStore.MarkFailedFunc.History()[0].Arg1)
				assert.Equal(t, "failed to transform record: failed", codeintelMockStore.MarkFailedFunc.History()[0].Arg2)
				assert.Equal(t, dbworkerstore.MarkFinalOptions{}, codeintelMockStore.MarkFailedFunc.History()[0].Arg3)
				require.Len(t, jobTokenStore.CreateFunc.History(), 0)
			},
			codeintelDequeueEvents: map[int]dequeueEvent[uploadsshared.Index]{
				0: {
					expectedStatusCode:   http.StatusInternalServerError,
					expectedResponseBody: `{"error":"RecordTransformer codeintel: 2 errors occurred:\n\t* failed\n\t* failed to mark"}`,
					transformerFunc: func(ctx context.Context, version string, record uploadsshared.Index, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
						return executortypes.Job{}, errors.New("failed")
					},
				},
			},
		},
		{
			name: "Failed to mark record as failed batches",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["batches"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				batchesMockStore.DequeueFunc.PushReturn(&btypes.BatchSpecWorkspaceExecutionJob{ID: 1}, true, nil)
				batchesMockStore.MarkFailedFunc.PushReturn(true, errors.New("failed to mark"))
			},
			assertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, batchesMockStore.DequeueFunc.History(), 1)
				require.Len(t, batchesMockStore.MarkFailedFunc.History(), 1)
				assert.Equal(t, 1, batchesMockStore.MarkFailedFunc.History()[0].Arg1)
				assert.Equal(t, "failed to transform record: failed", batchesMockStore.MarkFailedFunc.History()[0].Arg2)
				assert.Equal(t, dbworkerstore.MarkFinalOptions{}, batchesMockStore.MarkFailedFunc.History()[0].Arg3)
				require.Len(t, jobTokenStore.CreateFunc.History(), 0)
			},
			batchesDequeueEvents: map[int]dequeueEvent[*btypes.BatchSpecWorkspaceExecutionJob]{
				0: {
					expectedStatusCode:   http.StatusInternalServerError,
					expectedResponseBody: `{"error":"RecordTransformer batches: 2 errors occurred:\n\t* failed\n\t* failed to mark"}`,
					transformerFunc: func(ctx context.Context, version string, record *btypes.BatchSpecWorkspaceExecutionJob, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error) {
						return executortypes.Job{}, errors.New("failed")
					},
				},
			},
		},
		{
			name: "Failed to create job token",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["codeintel","batches"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				codeintelMockStore.DequeueFunc.PushReturn(uploadsshared.Index{ID: 1}, true, nil)
				jobTokenStore.CreateFunc.PushReturn("", errors.New("failed to create token"))
			},
			assertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CreateFunc.History(), 1)
				require.Len(t, jobTokenStore.RegenerateFunc.History(), 0)
			},
			codeintelDequeueEvents: map[int]dequeueEvent[uploadsshared.Index]{
				0: {
					expectedStatusCode:   http.StatusInternalServerError,
					expectedResponseBody: `{"error":"CreateToken: failed to create token"}`,
				},
			},
		},
		{
			name: "Job token already exists",
			body: `{"executorName": "test-executor","numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["codeintel","batches"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				codeintelMockStore.DequeueFunc.PushReturn(uploadsshared.Index{ID: 1}, true, nil)
				jobTokenStore.CreateFunc.PushReturn("", executorstore.ErrJobTokenAlreadyCreated)
				jobTokenStore.RegenerateFunc.PushReturn("somenewtoken", nil)
			},
			assertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CreateFunc.History(), 1)
				require.Len(t, jobTokenStore.RegenerateFunc.History(), 1)
				assert.Equal(t, 1, jobTokenStore.RegenerateFunc.History()[0].Arg1)
				assert.Equal(t, "codeintel", jobTokenStore.RegenerateFunc.History()[0].Arg2)
			},
			codeintelDequeueEvents: map[int]dequeueEvent[uploadsshared.Index]{
				0: {
					expectedStatusCode:   http.StatusOK,
					expectedResponseBody: `{"id":1,"token":"somenewtoken","queue":"codeintel", "repositoryName":"","repositoryDirectory":"","commit":"","fetchTags":false,"shallowClone":false,"sparseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redactedValues":null}`,
				},
			},
		},
		{
			name: "Failed to regenerate token",
			body: `{"executorName": "test-executor","numCPUs": 1, "memory": "1GB", "diskSpace": "10GB","queues": ["codeintel","batches"]}`,
			mockFunc: func(codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				codeintelMockStore.DequeueFunc.PushReturn(uploadsshared.Index{ID: 1}, true, nil)
				jobTokenStore.CreateFunc.PushReturn("", executorstore.ErrJobTokenAlreadyCreated)
				jobTokenStore.RegenerateFunc.PushReturn("", errors.New("failed to regen token"))
			},
			assertionFunc: func(t *testing.T, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob], jobTokenStore *executorstore.MockJobTokenStore) {
				require.Len(t, codeintelMockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CreateFunc.History(), 1)
				require.Len(t, jobTokenStore.RegenerateFunc.History(), 1)
			},
			codeintelDequeueEvents: map[int]dequeueEvent[uploadsshared.Index]{
				0: {
					expectedStatusCode:   http.StatusInternalServerError,
					expectedResponseBody: `{"error":"RegenerateToken: failed to regen token"}`,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jobTokenStore := executorstore.NewMockJobTokenStore()
			codeIntelMockStore := dbworkerstoremocks.NewMockStore[uploadsshared.Index]()
			batchesMockStore := dbworkerstoremocks.NewMockStore[*btypes.BatchSpecWorkspaceExecutionJob]()

			mh := handler.NewMultiHandler(
				database.NewMockExecutorStore(),
				jobTokenStore,
				metricsstore.NewMockDistributedStore(),
				handler.QueueHandler[uploadsshared.Index]{Name: "codeintel", Store: codeIntelMockStore, RecordTransformer: transformerFunc[uploadsshared.Index]},
				handler.QueueHandler[*btypes.BatchSpecWorkspaceExecutionJob]{Name: "batches", Store: batchesMockStore, RecordTransformer: transformerFunc[*btypes.BatchSpecWorkspaceExecutionJob]},
			)

			// Mock random fairness to alternate between how the queues are defined in the request and reversed between dequeues in a single test.
			// The first queue will be attempted to be dequeued. Without randomness, the first queue would be dequeued until no jobs are returned.
			//
			// Example with two dequeues and request ["codeintel", "batches"]: 1. codeintel, 2. batches (instead of 1. codeintel, 2. codeintel -> no job, then dequeue batches)
			// Example with one dequeue and request ["codeintel", "batches"]: 1. codeintel
			var intnResults []int
			for i := len(test.codeintelDequeueEvents) + len(test.batchesDequeueEvents); i > 0; i-- {
				intnResults = append(intnResults, i-1)
			}
			mh.RandomGenerator = &MockRandom{IntnResults: intnResults}

			router := mux.NewRouter()
			router.HandleFunc("/dequeue", mh.HandleDequeue)

			if test.mockFunc != nil {
				test.mockFunc(codeIntelMockStore, batchesMockStore, jobTokenStore)
			}

			if test.expectedStatusCode != 0 {
				evaluateEvent(test.body, test.expectedStatusCode, "", t, router)
			} else {
				for eventIndex := 0; eventIndex < len(test.batchesDequeueEvents)+len(test.codeintelDequeueEvents); eventIndex++ {
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
			}
		})
	}
}

func TestMultiHandler_HandleHeartbeat(t *testing.T) {
	tests := []struct {
		name                 string
		body                 string
		mockFunc             func(metricsStore *metricsstore.MockDistributedStore, executorStore *database.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob])
		expectedStatusCode   int
		expectedResponseBody string
		assertionFunc        func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *database.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob])
	}{
		{
			name: "Heartbeat for multiple queues",
			body: `{"executorName": "test-executor", "queueNames": ["codeintel", "batches"], "jobIdsByQueue": [{"queueName": "codeintel", "jobIds": ["42", "7"]}, {"queueName": "batches", "jobIds": ["43", "8"]}], "os": "test-os", "architecture": "test-arch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *database.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob]) {
				executorStore.UpsertHeartbeatFunc.PushReturn(nil)
				codeintelMockStore.HeartbeatFunc.PushReturn([]string{"42", "7"}, nil, nil)
				batchesMockStore.HeartbeatFunc.PushReturn([]string{"43", "8"}, nil, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"knownIds":["42-codeintel", "7-codeintel", "43-batches", "8-batches"],"cancelIds":null}`,
			assertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *database.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob]) {
				require.Len(t, executorStore.UpsertHeartbeatFunc.History(), 1)

				assert.Equal(
					t,
					types.Executor{
						Hostname:        "test-executor",
						QueueNames:      []string{"codeintel", "batches"},
						OS:              "test-os",
						Architecture:    "test-arch",
						DockerVersion:   "1.0",
						ExecutorVersion: "2.0",
						GitVersion:      "3.0",
						IgniteVersion:   "4.0",
						SrcCliVersion:   "5.0",
					},
					executorStore.UpsertHeartbeatFunc.History()[0].Arg1,
				)
				require.Len(t, codeintelMockStore.HeartbeatFunc.History(), 1)
				assert.Equal(t, []string{"42", "7"}, codeintelMockStore.HeartbeatFunc.History()[0].Arg1)
				assert.Equal(t, dbworkerstore.HeartbeatOptions{WorkerHostname: "test-executor"}, codeintelMockStore.HeartbeatFunc.History()[0].Arg2)

				require.Len(t, batchesMockStore.HeartbeatFunc.History(), 1)
				assert.Equal(t, []string{"43", "8"}, batchesMockStore.HeartbeatFunc.History()[0].Arg1)
				assert.Equal(t, dbworkerstore.HeartbeatOptions{WorkerHostname: "test-executor"}, batchesMockStore.HeartbeatFunc.History()[0].Arg2)
			},
		},
		{
			name: "Heartbeat for single queue",
			body: `{"executorName": "test-executor", "queueNames": ["codeintel"], "jobIdsByQueue": [{"queueName": "codeintel", "jobIds": ["42", "7"]}], "os": "test-os", "architecture": "test-arch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *database.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob]) {
				executorStore.UpsertHeartbeatFunc.PushReturn(nil)
				codeintelMockStore.HeartbeatFunc.PushReturn([]string{"42", "7"}, nil, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"knownIds":["42-codeintel", "7-codeintel"],"cancelIds":null}`,
			assertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *database.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob]) {
				require.Len(t, executorStore.UpsertHeartbeatFunc.History(), 1)

				assert.Equal(
					t,
					types.Executor{
						Hostname:        "test-executor",
						QueueNames:      []string{"codeintel"},
						OS:              "test-os",
						Architecture:    "test-arch",
						DockerVersion:   "1.0",
						ExecutorVersion: "2.0",
						GitVersion:      "3.0",
						IgniteVersion:   "4.0",
						SrcCliVersion:   "5.0",
					},
					executorStore.UpsertHeartbeatFunc.History()[0].Arg1,
				)
				require.Len(t, codeintelMockStore.HeartbeatFunc.History(), 1)
				assert.Equal(t, []string{"42", "7"}, codeintelMockStore.HeartbeatFunc.History()[0].Arg1)
				assert.Equal(t, dbworkerstore.HeartbeatOptions{WorkerHostname: "test-executor"}, codeintelMockStore.HeartbeatFunc.History()[0].Arg2)

				require.Len(t, batchesMockStore.HeartbeatFunc.History(), 0)
			},
		},
		{
			name: "No running jobs",
			body: `{"executorName": "test-executor", "queueNames": ["codeintel", "batches"], "jobIdsByQueue": [], "os": "test-os", "architecture": "test-arch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *database.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob]) {
				executorStore.UpsertHeartbeatFunc.PushReturn(nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"knownIds":null,"cancelIds":null}`,
			assertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *database.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob]) {
				require.Len(t, executorStore.UpsertHeartbeatFunc.History(), 1)

				assert.Equal(
					t,
					types.Executor{
						Hostname:        "test-executor",
						QueueNames:      []string{"codeintel", "batches"},
						OS:              "test-os",
						Architecture:    "test-arch",
						DockerVersion:   "1.0",
						ExecutorVersion: "2.0",
						GitVersion:      "3.0",
						IgniteVersion:   "4.0",
						SrcCliVersion:   "5.0",
					},
					executorStore.UpsertHeartbeatFunc.History()[0].Arg1,
				)
				require.Len(t, codeintelMockStore.HeartbeatFunc.History(), 0)
				require.Len(t, batchesMockStore.HeartbeatFunc.History(), 0)
			},
		},
		{
			name: "Known and canceled IDs",
			body: `{"executorName": "test-executor", "queueNames": ["codeintel", "batches"], "jobIdsByQueue": [{"queueName": "codeintel", "jobIds": ["42", "7"]}, {"queueName": "batches", "jobIds": ["43", "8"]}], "os": "test-os", "architecture": "test-arch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *database.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob]) {
				executorStore.UpsertHeartbeatFunc.PushReturn(nil)
				codeintelMockStore.HeartbeatFunc.PushReturn([]string{"42"}, []string{"7"}, nil)
				batchesMockStore.HeartbeatFunc.PushReturn([]string{"43"}, []string{"8"}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"knownIds":["42-codeintel", "43-batches"],"cancelIds":["7-codeintel", "8-batches"]}`,
			assertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *database.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob]) {
				require.Len(t, executorStore.UpsertHeartbeatFunc.History(), 1)

				assert.Equal(
					t,
					types.Executor{
						Hostname:        "test-executor",
						QueueNames:      []string{"codeintel", "batches"},
						OS:              "test-os",
						Architecture:    "test-arch",
						DockerVersion:   "1.0",
						ExecutorVersion: "2.0",
						GitVersion:      "3.0",
						IgniteVersion:   "4.0",
						SrcCliVersion:   "5.0",
					},
					executorStore.UpsertHeartbeatFunc.History()[0].Arg1,
				)
				require.Len(t, codeintelMockStore.HeartbeatFunc.History(), 1)
				assert.Equal(t, []string{"42", "7"}, codeintelMockStore.HeartbeatFunc.History()[0].Arg1)
				assert.Equal(t, dbworkerstore.HeartbeatOptions{WorkerHostname: "test-executor"}, codeintelMockStore.HeartbeatFunc.History()[0].Arg2)

				require.Len(t, batchesMockStore.HeartbeatFunc.History(), 1)
				assert.Equal(t, []string{"43", "8"}, batchesMockStore.HeartbeatFunc.History()[0].Arg1)
				assert.Equal(t, dbworkerstore.HeartbeatOptions{WorkerHostname: "test-executor"}, batchesMockStore.HeartbeatFunc.History()[0].Arg2)
			},
		},
		{
			name:                 "Invalid worker hostname",
			body:                 `{"executorName": "", "queueNames": ["codeintel", "batches"], "jobIdsByQueue": [{"queueName": "codeintel", "jobIds": ["42", "7"]}, {"queueName": "batches", "jobIds": ["43", "8"]}], "os": "test-os", "architecture": "test-arch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"worker hostname cannot be empty"}`,
			assertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *database.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob]) {
				require.Len(t, executorStore.UpsertHeartbeatFunc.History(), 0)
				require.Len(t, codeintelMockStore.HeartbeatFunc.History(), 0)
				require.Len(t, batchesMockStore.HeartbeatFunc.History(), 0)
			},
		},
		{
			name:                 "Job IDs by queue contains name not in queue names",
			body:                 `{"executorName": "test-executor", "queueNames": ["codeintel", "batches"], "jobIdsByQueue": [{"queueName": "foo", "jobIds": ["42"]}, {"queueName": "bar", "jobIds": ["43"]}], "os": "test-os", "architecture": "test-arch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"unsupported queue name(s) 'foo, bar' submitted in queueJobIds, executor is configured for queues 'codeintel, batches'"}`,
			assertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *database.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob]) {
				require.Len(t, executorStore.UpsertHeartbeatFunc.History(), 0)
				require.Len(t, codeintelMockStore.HeartbeatFunc.History(), 0)
				require.Len(t, batchesMockStore.HeartbeatFunc.History(), 0)
			},
		},
		{
			name:                 "Queue names missing",
			body:                 `{"executorName": "test-executor", "jobIdsByQueue": [{"queueName": "codeintel", "jobIds": ["42"]}, {"queueName": "batches", "jobIds": ["43"]}], "os": "test-os", "architecture": "test-arch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"queueNames must be set for multi-queue heartbeats"}`,
			assertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *database.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob]) {
				require.Len(t, executorStore.UpsertHeartbeatFunc.History(), 0)
				require.Len(t, codeintelMockStore.HeartbeatFunc.History(), 0)
				require.Len(t, batchesMockStore.HeartbeatFunc.History(), 0)
			},
		},
		{
			name: "Failed to upsert heartbeat",
			body: `{"executorName": "test-executor", "queueNames": ["codeintel", "batches"], "jobIdsByQueue": [{"queueName": "codeintel", "jobIds": ["42", "7"]}, {"queueName": "batches", "jobIds": ["43", "8"]}], "os": "test-os", "architecture": "test-arch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *database.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob]) {
				executorStore.UpsertHeartbeatFunc.PushReturn(errors.Newf("failed"))
				codeintelMockStore.HeartbeatFunc.PushReturn([]string{"42", "7"}, nil, nil)
				batchesMockStore.HeartbeatFunc.PushReturn([]string{"43", "8"}, nil, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"knownIds":["42-codeintel", "7-codeintel", "43-batches", "8-batches"],"cancelIds":null}`,
			assertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *database.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob]) {
				require.Len(t, executorStore.UpsertHeartbeatFunc.History(), 1)
				assert.Equal(
					t,
					types.Executor{
						Hostname:        "test-executor",
						QueueNames:      []string{"codeintel", "batches"},
						OS:              "test-os",
						Architecture:    "test-arch",
						DockerVersion:   "1.0",
						ExecutorVersion: "2.0",
						GitVersion:      "3.0",
						IgniteVersion:   "4.0",
						SrcCliVersion:   "5.0",
					},
					executorStore.UpsertHeartbeatFunc.History()[0].Arg1,
				)
				require.Len(t, codeintelMockStore.HeartbeatFunc.History(), 1)
				assert.Equal(t, []string{"42", "7"}, codeintelMockStore.HeartbeatFunc.History()[0].Arg1)
				assert.Equal(t, dbworkerstore.HeartbeatOptions{WorkerHostname: "test-executor"}, codeintelMockStore.HeartbeatFunc.History()[0].Arg2)

				require.Len(t, batchesMockStore.HeartbeatFunc.History(), 1)
				assert.Equal(t, []string{"43", "8"}, batchesMockStore.HeartbeatFunc.History()[0].Arg1)
				assert.Equal(t, dbworkerstore.HeartbeatOptions{WorkerHostname: "test-executor"}, batchesMockStore.HeartbeatFunc.History()[0].Arg2)
			},
		},
		{
			name: "Failed to heartbeat first queue, second is ignored",
			body: `{"executorName": "test-executor", "queueNames": ["codeintel", "batches"], "jobIdsByQueue": [{"queueName": "batches", "jobIds": ["43", "8"]}, {"queueName": "codeintel", "jobIds": ["42", "7"]}], "os": "test-os", "architecture": "test-arch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *database.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob]) {
				executorStore.UpsertHeartbeatFunc.PushReturn(nil)
				codeintelMockStore.HeartbeatFunc.PushReturn([]string{"42", "7"}, nil, nil)
				batchesMockStore.HeartbeatFunc.PushReturn(nil, nil, errors.New("failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"multiqueue.UpsertHeartbeat: failed"}`,
			assertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *database.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob]) {
				require.Len(t, executorStore.UpsertHeartbeatFunc.History(), 1)
				assert.Equal(
					t,
					types.Executor{
						Hostname:        "test-executor",
						QueueNames:      []string{"codeintel", "batches"},
						OS:              "test-os",
						Architecture:    "test-arch",
						DockerVersion:   "1.0",
						ExecutorVersion: "2.0",
						GitVersion:      "3.0",
						IgniteVersion:   "4.0",
						SrcCliVersion:   "5.0",
					},
					executorStore.UpsertHeartbeatFunc.History()[0].Arg1,
				)
				// switch statement in MultiHandler.heartbeat starts with batches, which is first in `jobIdsByQueue`, so not called
				require.Len(t, codeintelMockStore.HeartbeatFunc.History(), 0)

				require.Len(t, batchesMockStore.HeartbeatFunc.History(), 1)
				assert.Equal(t, []string{"43", "8"}, batchesMockStore.HeartbeatFunc.History()[0].Arg1)
				assert.Equal(t, dbworkerstore.HeartbeatOptions{WorkerHostname: "test-executor"}, batchesMockStore.HeartbeatFunc.History()[0].Arg2)
			},
		},
		{
			name: "First queue successful heartbeat, failed to heartbeat second queue",
			body: `{"executorName": "test-executor", "queueNames": ["codeintel", "batches"], "jobIdsByQueue": [{"queueName": "codeintel", "jobIds": ["42", "7"]}, {"queueName": "batches", "jobIds": ["43", "8"]}], "os": "test-os", "architecture": "test-arch", "dockerVersion": "1.0", "executorVersion": "2.0", "gitVersion": "3.0", "igniteVersion": "4.0", "srcCliVersion": "5.0", "prometheusMetrics": ""}`,
			mockFunc: func(metricsStore *metricsstore.MockDistributedStore, executorStore *database.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob]) {
				executorStore.UpsertHeartbeatFunc.PushReturn(nil)
				codeintelMockStore.HeartbeatFunc.PushReturn([]string{"42", "7"}, nil, nil)
				batchesMockStore.HeartbeatFunc.PushReturn(nil, nil, errors.New("failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"multiqueue.UpsertHeartbeat: failed"}`,
			assertionFunc: func(t *testing.T, metricsStore *metricsstore.MockDistributedStore, executorStore *database.MockExecutorStore, codeintelMockStore *dbworkerstoremocks.MockStore[uploadsshared.Index], batchesMockStore *dbworkerstoremocks.MockStore[*btypes.BatchSpecWorkspaceExecutionJob]) {
				require.Len(t, executorStore.UpsertHeartbeatFunc.History(), 1)
				assert.Equal(
					t,
					types.Executor{
						Hostname:        "test-executor",
						QueueNames:      []string{"codeintel", "batches"},
						OS:              "test-os",
						Architecture:    "test-arch",
						DockerVersion:   "1.0",
						ExecutorVersion: "2.0",
						GitVersion:      "3.0",
						IgniteVersion:   "4.0",
						SrcCliVersion:   "5.0",
					},
					executorStore.UpsertHeartbeatFunc.History()[0].Arg1,
				)
				// switch statement in MultiHandler.heartbeat starts with batches, which is first in `jobIdsByQueue`, so not called
				require.Len(t, codeintelMockStore.HeartbeatFunc.History(), 1)
				assert.Equal(t, []string{"42", "7"}, codeintelMockStore.HeartbeatFunc.History()[0].Arg1)
				assert.Equal(t, dbworkerstore.HeartbeatOptions{WorkerHostname: "test-executor"}, codeintelMockStore.HeartbeatFunc.History()[0].Arg2)

				require.Len(t, batchesMockStore.HeartbeatFunc.History(), 1)
				assert.Equal(t, []string{"43", "8"}, batchesMockStore.HeartbeatFunc.History()[0].Arg1)
				assert.Equal(t, dbworkerstore.HeartbeatOptions{WorkerHostname: "test-executor"}, batchesMockStore.HeartbeatFunc.History()[0].Arg2)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executorStore := database.NewMockExecutorStore()
			metricsStore := metricsstore.NewMockDistributedStore()
			codeIntelMockStore := dbworkerstoremocks.NewMockStore[uploadsshared.Index]()
			batchesMockStore := dbworkerstoremocks.NewMockStore[*btypes.BatchSpecWorkspaceExecutionJob]()

			mh := handler.NewMultiHandler(
				executorStore,
				executorstore.NewMockJobTokenStore(),
				metricsStore,
				handler.QueueHandler[uploadsshared.Index]{Name: "codeintel", Store: codeIntelMockStore},
				handler.QueueHandler[*btypes.BatchSpecWorkspaceExecutionJob]{Name: "batches", Store: batchesMockStore},
			)

			router := mux.NewRouter()
			router.HandleFunc("/heartbeat", mh.HandleHeartbeat)

			req, err := http.NewRequest(http.MethodPost, "/heartbeat", strings.NewReader(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(metricsStore, executorStore, codeIntelMockStore, batchesMockStore)
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
				test.assertionFunc(t, metricsStore, executorStore, codeIntelMockStore, batchesMockStore)
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
