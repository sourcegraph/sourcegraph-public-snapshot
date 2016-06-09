package middleware

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store/mockstore"
)

func TestInstrumentQueue(t *testing.T) {
	m := &mockstore.Queue{}
	q := InstrumentedQueue{m}
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
