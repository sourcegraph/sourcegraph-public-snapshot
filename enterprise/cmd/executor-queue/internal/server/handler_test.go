package server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	workerstoremocks "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store/mocks"
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

	handler := newHandler(QueueOptions{Store: store, RecordTransformer: recordTransformer})

	job, dequeued, err := handler.dequeue(context.Background(), "deadbeef", "test")
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
	handler := newHandler(QueueOptions{Store: workerstoremocks.NewMockStore()})

	_, dequeued, err := handler.dequeue(context.Background(), "deadbeef", "test")
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

	handler := newHandler(QueueOptions{Store: store, RecordTransformer: recordTransformer})

	job, dequeued, err := handler.dequeue(context.Background(), "deadbeef", "test")
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
	if err := handler.addExecutionLogEntry(context.Background(), "deadbeef", job.ID, entry); err != nil {
		t.Fatalf("unexpected error updating log contents: %s", err)
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
	store.AddExecutionLogEntryFunc.SetDefaultReturn(ErrUnknownJob)
	handler := newHandler(QueueOptions{Store: store})

	entry := workerutil.ExecutionLogEntry{
		Command: []string{"ls", "-a"},
		Out:     "<log payload>",
	}
	if err := handler.addExecutionLogEntry(context.Background(), "deadbeef", 42, entry); err != ErrUnknownJob {
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

	handler := newHandler(QueueOptions{Store: store, RecordTransformer: recordTransformer})

	job, dequeued, err := handler.dequeue(context.Background(), "deadbeef", "test")
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
	handler := newHandler(QueueOptions{Store: store})

	if err := handler.markComplete(context.Background(), "deadbeef", 42); err != ErrUnknownJob {
		t.Fatalf("unexpected error. want=%q have=%q", ErrUnknownJob, err)
	}
}

func TestMarkErrored(t *testing.T) {
	store := workerstoremocks.NewMockStore()
	store.DequeueFunc.SetDefaultReturn(testRecord{ID: 42}, true, nil)
	store.MarkErroredFunc.SetDefaultReturn(true, nil)
	recordTransformer := func(ctx context.Context, record workerutil.Record) (apiclient.Job, error) {
		return apiclient.Job{ID: 42}, nil
	}

	handler := newHandler(QueueOptions{Store: store, RecordTransformer: recordTransformer})

	job, dequeued, err := handler.dequeue(context.Background(), "deadbeef", "test")
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
	handler := newHandler(QueueOptions{Store: store})

	if err := handler.markErrored(context.Background(), "deadbeef", 42, "OH NO"); err != ErrUnknownJob {
		t.Fatalf("unexpected error. want=%q have=%q", ErrUnknownJob, err)
	}
}

func TestMarkFailed(t *testing.T) {
	store := workerstoremocks.NewMockStore()
	store.DequeueFunc.SetDefaultReturn(testRecord{ID: 42}, true, nil)
	store.MarkFailedFunc.SetDefaultReturn(true, nil)
	recordTransformer := func(ctx context.Context, record workerutil.Record) (apiclient.Job, error) {
		return apiclient.Job{ID: 42}, nil
	}

	handler := newHandler(QueueOptions{Store: store, RecordTransformer: recordTransformer})

	job, dequeued, err := handler.dequeue(context.Background(), "deadbeef", "test")
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
	handler := newHandler(QueueOptions{Store: store})

	if err := handler.markFailed(context.Background(), "deadbeef", 42, "OH NO"); err != ErrUnknownJob {
		t.Fatalf("unexpected error. want=%q have=%q", ErrUnknownJob, err)
	}
}

type testRecord struct {
	ID      int
	Payload string
}

func (r testRecord) RecordID() int { return r.ID }
