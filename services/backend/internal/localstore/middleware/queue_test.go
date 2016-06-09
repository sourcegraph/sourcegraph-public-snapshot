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
	called, _, _ = m.MockLockJob_Return(t, want)
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
}
