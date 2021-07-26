package server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	workerstoremocks "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store/mocks"
)

func TestHeartbeat(t *testing.T) {
	store1 := workerstoremocks.NewMockStore()
	recordTransformer := func(ctx context.Context, record workerutil.Record) (apiclient.Job, error) {
		return apiclient.Job{ID: record.RecordID()}, nil
	}

	store1.DequeueFunc.PushReturn(testRecord{ID: 41}, true, nil)
	store1.DequeueFunc.PushReturn(testRecord{ID: 42}, true, nil)
	store1.DequeueFunc.PushReturn(testRecord{ID: 43}, true, nil)
	store1.DequeueFunc.PushReturn(testRecord{ID: 44}, true, nil)
	store1.HeartbeatFunc.SetDefaultHook(func(ctx context.Context, ids []int, options store.HeartbeatOptions) ([]int, error) {
		knownIDs := make([]int, 0)
		for _, id := range ids {
			if id >= 41 && id <= 44 {
				knownIDs = append(knownIDs, id)
			}
		}

		return knownIDs, nil
	})

	handler := newHandler(QueueOptions{Store: store1, RecordTransformer: recordTransformer})

	_, dequeued1, _ := handler.dequeue(context.Background(), "deadbeef", "test")
	_, dequeued2, _ := handler.dequeue(context.Background(), "deadveal", "test")
	_, dequeued3, _ := handler.dequeue(context.Background(), "deadbeef", "test")
	_, dequeued4, _ := handler.dequeue(context.Background(), "deadveal", "test")
	if !dequeued1 || !dequeued2 || !dequeued3 || !dequeued4 {
		t.Fatalf("failed to dequeue records")
	}

	// missing no jobs
	if ids, err := handler.heartbeat(context.Background(), "deadbeef", []int{41, 43}); err != nil {
		t.Fatalf("unexpected error performing heartbeat: %s", err)
	} else if diff := cmp.Diff(ids, []int{}); diff != "" {
		t.Fatalf("invalid unknownIDs returned diff=%s", diff)
	}
	if ids, err := handler.heartbeat(context.Background(), "deadveal", []int{42, 44}); err != nil {
		t.Fatalf("unexpected error performing heartbeat: %s", err)
	} else if diff := cmp.Diff(ids, []int{}); diff != "" {
		t.Fatalf("invalid unknownIDs returned diff=%s", diff)
	}

	// missing one deadbeef jobs
	if ids, err := handler.heartbeat(context.Background(), "deadbeef", []int{41}); err != nil {
		t.Fatalf("unexpected error performing heartbeat: %s", err)
	} else if diff := cmp.Diff(ids, []int{}); diff != "" {
		t.Fatalf("invalid unknownIDs returned diff=%s", diff)
	}

	// missing two deadveal jobs
	if ids, err := handler.heartbeat(context.Background(), "deadbeef", []int{}); err != nil {
		t.Fatalf("unexpected error performing heartbeat: %s", err)
	} else if diff := cmp.Diff(ids, []int{}); diff != "" {
		t.Fatalf("invalid unknownIDs returned diff=%s", diff)
	}

	// unknown jobs
	if unknownIDs, err := handler.heartbeat(context.Background(), "deadbeef", []int{41, 43, 45}); err != nil {
		t.Fatalf("unexpected error performing heartbeat: %s", err)
	} else if diff := cmp.Diff([]int{45}, unknownIDs); diff != "" {
		t.Errorf("unexpected unknown ids (-want +got):\n%s", diff)
	}
	if unknownIDs, err := handler.heartbeat(context.Background(), "deadveal", []int{42, 44, 45}); err != nil {
		t.Fatalf("unexpected error performing heartbeat: %s", err)
	} else if diff := cmp.Diff([]int{45}, unknownIDs); diff != "" {
		t.Errorf("unexpected unknown ids (-want +got):\n%s", diff)
	}
}
