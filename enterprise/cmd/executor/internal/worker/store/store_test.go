package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestQueueShim_QueuedCount(t *testing.T) {
	queueStore := new(queueStoreMock)
	shim := store.QueueShim{
		Name:  "test-queue",
		Store: queueStore,
	}

	count, err := shim.QueuedCount(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "unimplemented", err.Error())
	assert.Zero(t, count)

	mock.AssertExpectationsForObjects(t, queueStore)
}

func TestQueueShim_Dequeue(t *testing.T) {
	queueStore := new(queueStoreMock)
	shim := store.QueueShim{
		Name:  "test-queue",
		Store: queueStore,
	}

	queueStore.On("Dequeue", mock.Anything, "test-queue", mock.Anything).
		Return(true, nil)

	record, dequeued, err := shim.Dequeue(context.Background(), "host-name", "foo")
	assert.NoError(t, err)
	assert.True(t, dequeued)
	assert.NotNil(t, record)

	mock.AssertExpectationsForObjects(t, queueStore)
}

func TestQueueShim_Dequeue_Error(t *testing.T) {
	queueStore := new(queueStoreMock)
	shim := store.QueueShim{
		Name:  "test-queue",
		Store: queueStore,
	}

	queueStore.On("Dequeue", mock.Anything, "test-queue", mock.Anything).
		Return(false, errors.New("failed"))

	record, dequeued, err := shim.Dequeue(context.Background(), "host-name", "foo")
	assert.Error(t, err)
	assert.Equal(t, "failed", err.Error())
	assert.False(t, dequeued)
	assert.Nil(t, record)

	mock.AssertExpectationsForObjects(t, queueStore)
}

func TestQueueShim_Heartbeat(t *testing.T) {
	queueStore := new(queueStoreMock)
	shim := store.QueueShim{
		Name:  "test-queue",
		Store: queueStore,
	}

	queueStore.On("Heartbeat", mock.Anything, "test-queue", []int{1, 2, 3}).
		Return([]int{4}, nil)

	ids, err := shim.Heartbeat(context.Background(), []int{1, 2, 3})
	assert.NoError(t, err)
	assert.Len(t, ids, 1)
	assert.Equal(t, []int{4}, ids)

	mock.AssertExpectationsForObjects(t, queueStore)
}

func TestQueueShim_AddExecutionLogEntry(t *testing.T) {
	queueStore := new(queueStoreMock)
	shim := store.QueueShim{
		Name:  "test-queue",
		Store: queueStore,
	}

	exitCode := 0
	duration := 10
	entry := workerutil.ExecutionLogEntry{
		Key:        "abc",
		Command:    []string{"foo", "bar"},
		StartTime:  time.Now(),
		ExitCode:   &exitCode,
		Out:        "faz baz",
		DurationMs: &duration,
	}
	queueStore.On("AddExecutionLogEntry", mock.Anything, "test-queue", 1, entry).
		Return(2, nil)

	id, err := shim.AddExecutionLogEntry(context.Background(), 1, entry)
	assert.NoError(t, err)
	assert.Equal(t, 2, id)

	mock.AssertExpectationsForObjects(t, queueStore)
}

func TestQueueShim_UpdateExecutionLogEntry(t *testing.T) {
	queueStore := new(queueStoreMock)
	shim := store.QueueShim{
		Name:  "test-queue",
		Store: queueStore,
	}

	exitCode := 0
	duration := 10
	entry := workerutil.ExecutionLogEntry{
		Key:        "abc",
		Command:    []string{"foo", "bar"},
		StartTime:  time.Now(),
		ExitCode:   &exitCode,
		Out:        "faz baz",
		DurationMs: &duration,
	}
	queueStore.On("UpdateExecutionLogEntry", mock.Anything, "test-queue", 1, 2, entry).
		Return(nil)

	err := shim.UpdateExecutionLogEntry(context.Background(), 1, 2, entry)
	assert.NoError(t, err)

	mock.AssertExpectationsForObjects(t, queueStore)
}

func TestQueueShim_MarkComplete(t *testing.T) {
	queueStore := new(queueStoreMock)
	shim := store.QueueShim{
		Name:  "test-queue",
		Store: queueStore,
	}

	queueStore.On("MarkComplete", mock.Anything, "test-queue", 1).
		Return(nil)

	marked, err := shim.MarkComplete(context.Background(), 1)
	assert.NoError(t, err)
	assert.True(t, marked)

	mock.AssertExpectationsForObjects(t, queueStore)
}

func TestQueueShim_MarkComplete_Error(t *testing.T) {
	queueStore := new(queueStoreMock)
	shim := store.QueueShim{
		Name:  "test-queue",
		Store: queueStore,
	}

	queueStore.On("MarkComplete", mock.Anything, "test-queue", 1).
		Return(errors.New("failed"))

	marked, err := shim.MarkComplete(context.Background(), 1)
	assert.Error(t, err)
	assert.Equal(t, "failed", err.Error())
	assert.True(t, marked)

	mock.AssertExpectationsForObjects(t, queueStore)
}

func TestQueueShim_MarkErrored(t *testing.T) {
	queueStore := new(queueStoreMock)
	shim := store.QueueShim{
		Name:  "test-queue",
		Store: queueStore,
	}

	queueStore.On("MarkErrored", mock.Anything, "test-queue", 1, "failed to handle").
		Return(nil)

	marked, err := shim.MarkErrored(context.Background(), 1, "failed to handle")
	assert.NoError(t, err)
	assert.True(t, marked)

	mock.AssertExpectationsForObjects(t, queueStore)
}

func TestQueueShim_MarkErrored_Error(t *testing.T) {
	queueStore := new(queueStoreMock)
	shim := store.QueueShim{
		Name:  "test-queue",
		Store: queueStore,
	}

	queueStore.On("MarkErrored", mock.Anything, "test-queue", 1, "failed to handle").
		Return(errors.New("failed"))

	marked, err := shim.MarkErrored(context.Background(), 1, "failed to handle")
	assert.Error(t, err)
	assert.Equal(t, "failed", err.Error())
	assert.True(t, marked)

	mock.AssertExpectationsForObjects(t, queueStore)
}

func TestQueueShim_MarkFailed(t *testing.T) {
	queueStore := new(queueStoreMock)
	shim := store.QueueShim{
		Name:  "test-queue",
		Store: queueStore,
	}

	queueStore.On("MarkFailed", mock.Anything, "test-queue", 1, "failed to handle").
		Return(nil)

	marked, err := shim.MarkFailed(context.Background(), 1, "failed to handle")
	assert.NoError(t, err)
	assert.True(t, marked)

	mock.AssertExpectationsForObjects(t, queueStore)
}

func TestQueueShim_MarkFailed_Error(t *testing.T) {
	queueStore := new(queueStoreMock)
	shim := store.QueueShim{
		Name:  "test-queue",
		Store: queueStore,
	}

	queueStore.On("MarkFailed", mock.Anything, "test-queue", 1, "failed to handle").
		Return(errors.New("failed"))

	marked, err := shim.MarkFailed(context.Background(), 1, "failed to handle")
	assert.Error(t, err)
	assert.Equal(t, "failed", err.Error())
	assert.True(t, marked)

	mock.AssertExpectationsForObjects(t, queueStore)
}

func TestQueueShim_CanceledJobs(t *testing.T) {
	queueStore := new(queueStoreMock)
	shim := store.QueueShim{
		Name:  "test-queue",
		Store: queueStore,
	}

	queueStore.On("CanceledJobs", mock.Anything, "test-queue", []int{1, 2, 3}).
		Return([]int{4, 5}, nil)

	ids, err := shim.CanceledJobs(context.Background(), []int{1, 2, 3})
	assert.NoError(t, err)
	assert.Len(t, ids, 2)
	assert.Equal(t, []int{4, 5}, ids)

	mock.AssertExpectationsForObjects(t, queueStore)
}

type queueStoreMock struct {
	mock.Mock
}

func (m *queueStoreMock) Dequeue(ctx context.Context, queueName string, payload *executor.Job) (bool, error) {
	args := m.Called(ctx, queueName, payload)
	return args.Bool(0), args.Error(1)
}

func (m *queueStoreMock) AddExecutionLogEntry(ctx context.Context, queueName string, jobID int, entry workerutil.ExecutionLogEntry) (int, error) {
	args := m.Called(ctx, queueName, jobID, entry)
	return args.Int(0), args.Error(1)
}

func (m *queueStoreMock) UpdateExecutionLogEntry(ctx context.Context, queueName string, jobID, entryID int, entry workerutil.ExecutionLogEntry) error {
	args := m.Called(ctx, queueName, jobID, entryID, entry)
	return args.Error(0)
}

func (m *queueStoreMock) MarkComplete(ctx context.Context, queueName string, jobID int) error {
	args := m.Called(ctx, queueName, jobID)
	return args.Error(0)
}

func (m *queueStoreMock) MarkErrored(ctx context.Context, queueName string, jobID int, errorMessage string) error {
	args := m.Called(ctx, queueName, jobID, errorMessage)
	return args.Error(0)
}

func (m *queueStoreMock) MarkFailed(ctx context.Context, queueName string, jobID int, errorMessage string) error {
	args := m.Called(ctx, queueName, jobID, errorMessage)
	return args.Error(0)
}

func (m *queueStoreMock) Heartbeat(ctx context.Context, queueName string, jobIDs []int) (knownIDs []int, err error) {
	args := m.Called(ctx, queueName, jobIDs)
	return args.Get(0).([]int), args.Error(1)
}

func (m *queueStoreMock) CanceledJobs(ctx context.Context, queueName string, knownIDs []int) (canceledIDs []int, err error) {
	args := m.Called(ctx, queueName, knownIDs)
	return args.Get(0).([]int), args.Error(1)
}
