package apiserver

import (
	"context"
	"testing"
	"time"

	"github.com/efritz/glock"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/apiworker/apiclient"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	workerstoremocks "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store/mocks"
)

func TestHeartbeat(t *testing.T) {
	store1 := workerstoremocks.NewMockStore()
	store2 := workerstoremocks.NewMockStore()
	recordTransformer := func(record workerutil.Record) (apiclient.Job, error) {
		return apiclient.Job{ID: record.RecordID()}, nil
	}

	store1.DequeueWithIndependentTransactionContextFunc.PushReturn(testRecord{ID: 41}, store1, true, nil)
	store1.DequeueWithIndependentTransactionContextFunc.PushReturn(testRecord{ID: 42}, store1, true, nil)
	store2.DequeueWithIndependentTransactionContextFunc.PushReturn(testRecord{ID: 42}, store2, true, nil)
	store2.DequeueWithIndependentTransactionContextFunc.PushReturn(testRecord{ID: 43}, store2, true, nil)

	options := Options{
		QueueOptions: map[string]QueueOptions{
			"q1": {Store: store1, RecordTransformer: recordTransformer},
			"q2": {Store: store2, RecordTransformer: recordTransformer},
		},
		MaximumNumTransactions: 10,
		UnreportedMaxAge:       time.Second,
	}
	clock := glock.NewMockClock()
	handler := newHandler(options, clock)

	_, dequeued1, _ := handler.dequeue(context.Background(), "q1", "deadbeef")
	_, dequeued2, _ := handler.dequeue(context.Background(), "q1", "deadveal")
	_, dequeued3, _ := handler.dequeue(context.Background(), "q2", "deadbeef")
	_, dequeued4, _ := handler.dequeue(context.Background(), "q2", "deadveal")
	if !dequeued1 || !dequeued2 || !dequeued3 || !dequeued4 {
		t.Fatalf("failed to dequeue records")
	}

	assertDoneCounts := func(c1, c2 int) {
		if value := len(store1.DoneFunc.History()); value != c1 {
			t.Fatalf("unexpected number of calls to Done. want=%d have=%d", c1, value)
		}
		if value := len(store2.DoneFunc.History()); value != c2 {
			t.Fatalf("unexpected number of calls to Done. want=%d have=%d", c2, value)
		}
	}

	// missing all jobs, but they're less than UnreportedMaxAge
	clock.Advance(time.Second / 2)
	if err := handler.heartbeat(context.Background(), "deadbeef", []int{}); err != nil {
		t.Fatalf("unexpected error performing heartbeat: %s", err)
	}
	if err := handler.heartbeat(context.Background(), "deadveal", []int{}); err != nil {
		t.Fatalf("unexpected error performing heartbeat: %s", err)
	}
	assertDoneCounts(0, 0)

	// missing no jobs
	clock.Advance(time.Minute * 2)
	if err := handler.heartbeat(context.Background(), "deadbeef", []int{41, 42}); err != nil {
		t.Fatalf("unexpected error performing heartbeat: %s", err)
	}
	if err := handler.heartbeat(context.Background(), "deadveal", []int{42, 43}); err != nil {
		t.Fatalf("unexpected error performing heartbeat: %s", err)
	}
	assertDoneCounts(0, 0)

	// missing one deadbeef jobs
	clock.Advance(time.Minute * 2)
	if err := handler.heartbeat(context.Background(), "deadbeef", []int{41}); err != nil {
		t.Fatalf("unexpected error performing heartbeat: %s", err)
	}
	if err := handler.heartbeat(context.Background(), "deadveal", []int{42, 43}); err != nil {
		t.Fatalf("unexpected error performing heartbeat: %s", err)
	}
	assertDoneCounts(0, 1)

	// missing two deadveal jobs
	clock.Advance(time.Minute * 2)
	if err := handler.heartbeat(context.Background(), "deadbeef", []int{41}); err != nil {
		t.Fatalf("unexpected error performing heartbeat: %s", err)
	}
	if err := handler.heartbeat(context.Background(), "deadveal", []int{}); err != nil {
		t.Fatalf("unexpected error performing heartbeat: %s", err)
	}
	assertDoneCounts(1, 2)
}

func TestCleanup(t *testing.T) {
	store1 := workerstoremocks.NewMockStore()
	store2 := workerstoremocks.NewMockStore()
	recordTransformer := func(record workerutil.Record) (apiclient.Job, error) {
		return apiclient.Job{ID: record.RecordID()}, nil
	}

	store1.DequeueWithIndependentTransactionContextFunc.PushReturn(testRecord{ID: 41}, store1, true, nil)
	store1.DequeueWithIndependentTransactionContextFunc.PushReturn(testRecord{ID: 42}, store1, true, nil)
	store2.DequeueWithIndependentTransactionContextFunc.PushReturn(testRecord{ID: 42}, store2, true, nil)
	store2.DequeueWithIndependentTransactionContextFunc.PushReturn(testRecord{ID: 43}, store2, true, nil)

	options := Options{
		QueueOptions: map[string]QueueOptions{
			"q1": {Store: store1, RecordTransformer: recordTransformer},
			"q2": {Store: store2, RecordTransformer: recordTransformer},
		},
		MaximumNumTransactions: 10,
		DeathThreshold:         time.Minute * 5,
	}
	clock := glock.NewMockClock()
	handler := newHandler(options, clock)

	_, dequeued1, _ := handler.dequeue(context.Background(), "q1", "deadbeef")
	_, dequeued2, _ := handler.dequeue(context.Background(), "q1", "deadveal")
	_, dequeued3, _ := handler.dequeue(context.Background(), "q2", "deadbeef")
	_, dequeued4, _ := handler.dequeue(context.Background(), "q2", "deadveal")
	if !dequeued1 || !dequeued2 || !dequeued3 || !dequeued4 {
		t.Fatalf("failed to dequeue records")
	}

	for i := 0; i < 6; i++ {
		clock.Advance(time.Minute)

		if err := handler.heartbeat(context.Background(), "deadbeef", []int{41, 42}); err != nil {
			t.Fatalf("unexpected error performing heartbeat: %s", err)
		}
	}

	if err := handler.cleanup(context.Background()); err != nil {
		t.Fatalf("unexpected error performing cleanup: %s", err)
	}

	if value := len(store1.DoneFunc.History()); value != 1 {
		t.Fatalf("unexpected number of calls to Done. want=%d have=%d", 1, value)
	}
	if value := len(store2.DoneFunc.History()); value != 1 {
		t.Fatalf("unexpected number of calls to Done. want=%d have=%d", 1, value)
	}
}
