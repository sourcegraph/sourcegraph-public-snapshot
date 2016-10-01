package localstore

import (
	"reflect"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store/mockstore"
)

func TestInstrumentQueue(t *testing.T) {
	m := &mockstore.Queue{}
	q := instrumentedQueue{m}
	ctx := context.Background()

	// Enqueue
	want := &store.Job{Type: "test"}
	called := m.MockEnqueue(t, want)
	if err := q.Enqueue(ctx, want); err != nil {
		t.Fatal(err)
	}
	if !*called {
		t.Error("Did not call underlying Enqueue")
	}

	// LockJob
	called, calledSuccess, calledError := m.MockLockJob_Return(t, want)
	got, err := q.LockJob(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Job, want) {
		t.Errorf("unexpected LockedJob, got %+v, wanted %+v", got.Job, want)
	}
	if !*called {
		t.Error("Did not call underlying LockJob")
	}

	// LockJob.MarkSuccess
	*calledSuccess, *calledError = false, false
	err = got.MarkSuccess()
	if err != nil {
		t.Fatal(err)
	}
	if !*calledSuccess {
		t.Error("Did not call underlying LockJob.MarkSuccess")
	}
	if *calledError {
		t.Error("Called underlying LockJob.MarkError")
	}

	// LockJob.MarkError
	*calledSuccess, *calledError = false, false
	err = got.MarkError("test")
	if err != nil {
		t.Fatal(err)
	}
	if !*calledError {
		t.Error("Did not call underlying LockJob.MarkError")
	}
	if *calledSuccess {
		t.Error("Called underlying LockJob.MarkSuccess")
	}
}

func TestQueueStatsCollector(t *testing.T) {
	m := &mockstore.Queue{}
	stats := map[string]store.QueueStats{
		"a": store.QueueStats{
			NumJobs:          3,
			NumJobsWithError: 1,
		},
		"b": store.QueueStats{
			NumJobs: 1,
		},
	}
	m.Stats_ = func(_ context.Context) (map[string]store.QueueStats, error) {
		return stats, nil
	}

	// We just check that we collect 4 stats, and don't actually check we
	// collect legit values.
	var (
		c     = newQueueStatsCollector(context.Background(), m)
		ch    = make(chan prometheus.Metric)
		count = 0
	)
	go func() {
		c.Collect(ch)
		close(ch)
	}()
	for {
		select {
		case m := <-ch:
			if m == nil && count == 4 {
				return
			}
			if count > 4 || (m == nil && count != 4) {
				t.Fatalf("collected %d metrics, wanted 4", count)
			}
			count++
		case <-time.After(1 * time.Second):
			t.Fatal("expected collect timed out")
		}
	}
}
