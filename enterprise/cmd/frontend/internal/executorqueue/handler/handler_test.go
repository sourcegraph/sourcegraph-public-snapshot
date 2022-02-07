package handler

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	workerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	workerstoremocks "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store/mocks"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestDequeue(t *testing.T) {
	transformedJob := apiclient.Job{
		ID: 42,
		DockerSteps: []apiclient.DockerStep{
			{
				Image:    "alpine:latest",
				Commands: []string{"ls", "-a"},
			},
		},
	}

	store := workerstoremocks.NewMockStore()
	store.DequeueFunc.SetDefaultReturn(testRecord{ID: 42, Payload: "secret"}, true, nil)
	recordTransformer := func(ctx context.Context, record workerutil.Record) (apiclient.Job, error) {
		if tr, ok := record.(testRecord); !ok {
			t.Errorf("mismatched record type.")
		} else if tr.Payload != "secret" {
			t.Errorf("unexpected payload. want=%q have=%q", "secret", tr.Payload)
		}

		return transformedJob, nil
	}

	executorStore := NewMockStore()

	handler := newHandler(executorStore, QueueOptions{Store: store, RecordTransformer: recordTransformer})

	job, dequeued, err := handler.dequeue(context.Background(), "deadbeef")
	if err != nil {
		t.Fatalf("unexpected error dequeueing job: %s", err)
	}
	if !dequeued {
		t.Fatalf("expected job to be dequeued")
	}
	if job.ID != 42 {
		t.Errorf("unexpected id. want=%d have=%d", 42, job.ID)
	}
	if diff := cmp.Diff(transformedJob, job); diff != "" {
		t.Errorf("unexpected job (-want +got):\n%s", diff)
	}
}

func TestDequeueNoRecord(t *testing.T) {
	handler := newHandler(NewMockStore(), QueueOptions{Store: workerstoremocks.NewMockStore()})

	_, dequeued, err := handler.dequeue(context.Background(), "deadbeef")
	if err != nil {
		t.Fatalf("unexpected error dequeueing job: %s", err)
	}
	if dequeued {
		t.Fatalf("did not expect a job to be dequeued")
	}
}

func TestAddExecutionLogEntry(t *testing.T) {
	store := workerstoremocks.NewMockStore()
	store.DequeueFunc.SetDefaultReturn(testRecord{ID: 42}, true, nil)
	recordTransformer := func(ctx context.Context, record workerutil.Record) (apiclient.Job, error) {
		return apiclient.Job{ID: 42}, nil
	}
	fakeEntryID := 99
	store.AddExecutionLogEntryFunc.SetDefaultReturn(fakeEntryID, nil)

	executorStore := NewMockStore()

	handler := newHandler(executorStore, QueueOptions{Store: store, RecordTransformer: recordTransformer})

	job, dequeued, err := handler.dequeue(context.Background(), "deadbeef")
	if err != nil {
		t.Fatalf("unexpected error dequeueing job: %s", err)
	}
	if !dequeued {
		t.Fatalf("expected a job to be dequeued")
	}

	entry := workerutil.ExecutionLogEntry{
		Command: []string{"ls", "-a"},
		Out:     "<log payload>",
	}
	haveEntryID, err := handler.addExecutionLogEntry(context.Background(), "deadbeef", job.ID, entry)
	if err != nil {
		t.Fatalf("unexpected error updating log contents: %s", err)
	}
	if haveEntryID != fakeEntryID {
		t.Fatalf("unexpected entry ID returned. want=%d, have=%d", fakeEntryID, haveEntryID)
	}

	if value := len(store.AddExecutionLogEntryFunc.History()); value != 1 {
		t.Fatalf("unexpected number of calls to AddExecutionLogEntry. want=%d have=%d", 1, value)
	}
	call := store.AddExecutionLogEntryFunc.History()[0]
	if call.Arg1 != 42 {
		t.Errorf("unexpected job identifier. want=%d have=%d", 42, call.Arg1)
	}
	if diff := cmp.Diff(entry, call.Arg2); diff != "" {
		t.Errorf("unexpected entry (-want +got):\n%s", diff)
	}
}

func TestAddExecutionLogEntryUnknownJob(t *testing.T) {
	store := workerstoremocks.NewMockStore()
	store.AddExecutionLogEntryFunc.SetDefaultReturn(0, workerstore.ErrExecutionLogEntryNotUpdated)
	executorStore := NewMockStore()
	handler := newHandler(executorStore, QueueOptions{Store: store})

	entry := workerutil.ExecutionLogEntry{
		Command: []string{"ls", "-a"},
		Out:     "<log payload>",
	}
	if _, err := handler.addExecutionLogEntry(context.Background(), "deadbeef", 42, entry); err != ErrUnknownJob {
		t.Fatalf("unexpected error. want=%q have=%q", ErrUnknownJob, err)
	}
}

func TestUpdateExecutionLogEntry(t *testing.T) {
	store := workerstoremocks.NewMockStore()
	store.DequeueFunc.SetDefaultReturn(testRecord{ID: 42}, true, nil)
	recordTransformer := func(ctx context.Context, record workerutil.Record) (apiclient.Job, error) {
		return apiclient.Job{ID: 42}, nil
	}

	executorStore := NewMockStore()

	handler := newHandler(executorStore, QueueOptions{Store: store, RecordTransformer: recordTransformer})

	job, dequeued, err := handler.dequeue(context.Background(), "deadbeef")
	if err != nil {
		t.Fatalf("unexpected error dequeueing job: %s", err)
	}
	if !dequeued {
		t.Fatalf("expected a job to be dequeued")
	}

	entry := workerutil.ExecutionLogEntry{
		Command: []string{"ls", "-a"},
		Out:     "<log payload>",
	}

	if err := handler.updateExecutionLogEntry(context.Background(), "deadbeef", job.ID, 99, entry); err != nil {
		t.Fatalf("unexpected error updating log contents: %s", err)
	}

	if value := len(store.UpdateExecutionLogEntryFunc.History()); value != 1 {
		t.Fatalf("unexpected number of calls to UpdateExecutionLogEntry. want=%d have=%d", 1, value)
	}
	call := store.UpdateExecutionLogEntryFunc.History()[0]
	if call.Arg1 != 42 {
		t.Errorf("unexpected job identifier. want=%d have=%d", 42, call.Arg1)
	}
	if call.Arg2 != 99 {
		t.Errorf("unexpected entry ID. want=%d have=%d", 99, call.Arg1)
	}
	if diff := cmp.Diff(entry, call.Arg3); diff != "" {
		t.Errorf("unexpected entry (-want +got):\n%s", diff)
	}
}

func TestUpdateExecutionLogEntryUnknownJob(t *testing.T) {
	store := workerstoremocks.NewMockStore()
	store.UpdateExecutionLogEntryFunc.SetDefaultReturn(workerstore.ErrExecutionLogEntryNotUpdated)
	executorStore := NewMockStore()
	handler := newHandler(executorStore, QueueOptions{Store: store})

	entry := workerutil.ExecutionLogEntry{
		Command: []string{"ls", "-a"},
		Out:     "<log payload>",
	}
	if err := handler.updateExecutionLogEntry(context.Background(), "deadbeef", 42, 99, entry); err != ErrUnknownJob {
		t.Fatalf("unexpected error. want=%q have=%q", ErrUnknownJob, err)
	}
}

func TestMarkComplete(t *testing.T) {
	store := workerstoremocks.NewMockStore()
	store.DequeueFunc.SetDefaultReturn(testRecord{ID: 42}, true, nil)
	store.MarkCompleteFunc.SetDefaultReturn(true, nil)
	recordTransformer := func(ctx context.Context, record workerutil.Record) (apiclient.Job, error) {
		return apiclient.Job{ID: 42}, nil
	}

	executorStore := NewMockStore()

	handler := newHandler(executorStore, QueueOptions{Store: store, RecordTransformer: recordTransformer})

	job, dequeued, err := handler.dequeue(context.Background(), "deadbeef")
	if err != nil {
		t.Fatalf("unexpected error dequeueing job: %s", err)
	}
	if !dequeued {
		t.Fatalf("expected a job to be dequeued")
	}

	if err := handler.markComplete(context.Background(), "deadbeef", job.ID); err != nil {
		t.Fatalf("unexpected error completing job: %s", err)
	}

	if value := len(store.MarkCompleteFunc.History()); value != 1 {
		t.Fatalf("unexpected number of calls to MarkComplete. want=%d have=%d", 1, value)
	}
	call := store.MarkCompleteFunc.History()[0]
	if call.Arg1 != 42 {
		t.Errorf("unexpected job identifier. want=%d have=%d", 42, call.Arg1)
	}
}

func TestMarkCompleteUnknownJob(t *testing.T) {
	store := workerstoremocks.NewMockStore()
	store.MarkCompleteFunc.SetDefaultReturn(false, nil)
	executorStore := NewMockStore()
	handler := newHandler(executorStore, QueueOptions{Store: store})

	if err := handler.markComplete(context.Background(), "deadbeef", 42); err != ErrUnknownJob {
		t.Fatalf("unexpected error. want=%q have=%q", ErrUnknownJob, err)
	}
}

func TestMarkCompleteStoreError(t *testing.T) {
	store := workerstoremocks.NewMockStore()
	internalErr := errors.New("something went wrong")
	store.MarkCompleteFunc.SetDefaultReturn(false, internalErr)
	executorStore := NewMockStore()
	handler := newHandler(executorStore, QueueOptions{Store: store})

	if err := handler.markComplete(context.Background(), "deadbeef", 42); err == nil || errors.UnwrapAll(err).Error() != internalErr.Error() {
		t.Fatalf("unexpected error. want=%q have=%q", internalErr, errors.UnwrapAll(err))
	}
}

func TestMarkErrored(t *testing.T) {
	store := workerstoremocks.NewMockStore()
	store.DequeueFunc.SetDefaultReturn(testRecord{ID: 42}, true, nil)
	store.MarkErroredFunc.SetDefaultReturn(true, nil)
	recordTransformer := func(ctx context.Context, record workerutil.Record) (apiclient.Job, error) {
		return apiclient.Job{ID: 42}, nil
	}

	executorStore := NewMockStore()

	handler := newHandler(executorStore, QueueOptions{Store: store, RecordTransformer: recordTransformer})

	job, dequeued, err := handler.dequeue(context.Background(), "deadbeef")
	if err != nil {
		t.Fatalf("unexpected error dequeueing job: %s", err)
	}
	if !dequeued {
		t.Fatalf("expected a job to be dequeued")
	}

	if err := handler.markErrored(context.Background(), "deadbeef", job.ID, "OH NO"); err != nil {
		t.Fatalf("unexpected error completing job: %s", err)
	}

	if value := len(store.MarkErroredFunc.History()); value != 1 {
		t.Fatalf("unexpected number of calls to MarkErrored. want=%d have=%d", 1, value)
	}
	call := store.MarkErroredFunc.History()[0]
	if call.Arg1 != 42 {
		t.Errorf("unexpected job identifier. want=%d have=%d", 42, call.Arg1)
	}
	if call.Arg2 != "OH NO" {
		t.Errorf("unexpected job error. want=%s have=%s", "OH NO", call.Arg2)
	}
}

func TestMarkErroredUnknownJob(t *testing.T) {
	store := workerstoremocks.NewMockStore()
	store.MarkErroredFunc.SetDefaultReturn(false, nil)
	executorStore := NewMockStore()
	handler := newHandler(executorStore, QueueOptions{Store: store})

	if err := handler.markErrored(context.Background(), "deadbeef", 42, "OH NO"); err != ErrUnknownJob {
		t.Fatalf("unexpected error. want=%q have=%q", ErrUnknownJob, err)
	}
}

func TestMarkErroredStoreError(t *testing.T) {
	store := workerstoremocks.NewMockStore()
	storeErr := errors.New("something went wrong")
	store.MarkErroredFunc.SetDefaultReturn(false, storeErr)
	executorStore := NewMockStore()
	handler := newHandler(executorStore, QueueOptions{Store: store})

	if err := handler.markErrored(context.Background(), "deadbeef", 42, "OH NO"); err == nil || errors.UnwrapAll(err).Error() != storeErr.Error() {
		t.Fatalf("unexpected error. want=%q have=%q", storeErr, errors.UnwrapAll(err))
	}
}

func TestMarkFailed(t *testing.T) {
	store := workerstoremocks.NewMockStore()
	store.DequeueFunc.SetDefaultReturn(testRecord{ID: 42}, true, nil)
	store.MarkFailedFunc.SetDefaultReturn(true, nil)
	recordTransformer := func(ctx context.Context, record workerutil.Record) (apiclient.Job, error) {
		return apiclient.Job{ID: 42}, nil
	}

	executorStore := NewMockStore()

	handler := newHandler(executorStore, QueueOptions{Store: store, RecordTransformer: recordTransformer})

	job, dequeued, err := handler.dequeue(context.Background(), "deadbeef")
	if err != nil {
		t.Fatalf("unexpected error dequeueing job: %s", err)
	}
	if !dequeued {
		t.Fatalf("expected a job to be dequeued")
	}

	if err := handler.markFailed(context.Background(), "deadbeef", job.ID, "OH NO"); err != nil {
		t.Fatalf("unexpected error completing job: %s", err)
	}

	if value := len(store.MarkFailedFunc.History()); value != 1 {
		t.Fatalf("unexpected number of calls to MarkFailed. want=%d have=%d", 1, value)
	}
	call := store.MarkFailedFunc.History()[0]
	if call.Arg1 != 42 {
		t.Errorf("unexpected job identifier. want=%d have=%d", 42, call.Arg1)
	}
	if call.Arg2 != "OH NO" {
		t.Errorf("unexpected job error. want=%s have=%s", "OH NO", call.Arg2)
	}
}

func TestMarkFailedUnknownJob(t *testing.T) {
	store := workerstoremocks.NewMockStore()
	store.MarkFailedFunc.SetDefaultReturn(false, nil)
	executorStore := NewMockStore()
	handler := newHandler(executorStore, QueueOptions{Store: store})

	if err := handler.markFailed(context.Background(), "deadbeef", 42, "OH NO"); err != ErrUnknownJob {
		t.Fatalf("unexpected error. want=%q have=%q", ErrUnknownJob, err)
	}
}

func TestMarkFailedStoreError(t *testing.T) {
	store := workerstoremocks.NewMockStore()
	storeErr := errors.New("something went wrong")
	store.MarkFailedFunc.SetDefaultReturn(false, storeErr)
	executorStore := NewMockStore()
	handler := newHandler(executorStore, QueueOptions{Store: store})

	if err := handler.markFailed(context.Background(), "deadbeef", 42, "OH NO"); err == nil || errors.UnwrapAll(err).Error() != storeErr.Error() {
		t.Fatalf("unexpected error. want=%q have=%q", storeErr, errors.UnwrapAll(err))
	}
}

func TestHeartbeat(t *testing.T) {
	s := workerstoremocks.NewMockStore()
	recordTransformer := func(ctx context.Context, record workerutil.Record) (apiclient.Job, error) {
		return apiclient.Job{ID: record.RecordID()}, nil
	}
	testKnownID := 10
	s.HeartbeatFunc.SetDefaultHook(func(ctx context.Context, ids []int, options store.HeartbeatOptions) ([]int, error) {
		return []int{testKnownID}, nil
	})

	executorStore := NewMockStore()

	executor := types.Executor{
		Hostname:        "test-hostname",
		QueueName:       "test-queue-name",
		OS:              "test-os",
		Architecture:    "test-architecture",
		DockerVersion:   "test-docker-version",
		ExecutorVersion: "test-executor-version",
		GitVersion:      "test-git-version",
		IgniteVersion:   "test-ignite-version",
		SrcCliVersion:   "test-src-cli-version",
	}

	handler := newHandler(executorStore, QueueOptions{Store: s, RecordTransformer: recordTransformer})

	if knownIDs, err := handler.heartbeat(context.Background(), executor, []int{testKnownID, 10}); err != nil {
		t.Fatalf("unexpected error performing heartbeat: %s", err)
	} else if diff := cmp.Diff([]int{testKnownID}, knownIDs); diff != "" {
		t.Errorf("unexpected unknown ids (-want +got):\n%s", diff)
	}

	if callCount := len(executorStore.UpsertHeartbeatFunc.History()); callCount != 1 {
		t.Errorf("unexpected heartbeat upsert count. want=%d have=%d", 1, callCount)
	} else if name := executorStore.UpsertHeartbeatFunc.History()[0].Arg1; name != executor {
		t.Errorf("unexpected heartbeat name. want=%q have=%q", "deadbeef", name)
	}
}

type testRecord struct {
	ID      int
	Payload string
}

func (r testRecord) RecordID() int { return r.ID }
