package backend

import (
	"encoding/json"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
)

func TestAsyncService_RefreshIndexes(t *testing.T) {
	s := &async{}
	w := &asyncWorker{}
	ctx, mock := testContext()

	wantRepo := int32(10810)

	// Enqueue
	op := &sourcegraph.AsyncRefreshIndexesOp{
		Repo:   wantRepo,
		Source: "test",
	}
	job := &store.Job{
		Type: "RefreshIndexes",
		Args: mustMarshal(t, &sourcegraph.AsyncRefreshIndexesOp{
			Repo:   wantRepo,
			Source: "test (UID 1 test)",
		}),
	}
	calledEnqueue := mock.stores.Queue.MockEnqueue(t, job)
	_, err := s.RefreshIndexes(ctx, op)
	if err != nil {
		t.Fatal(err)
	}
	if !*calledEnqueue {
		t.Fatal("async.RefreshIndexes did not call Enqueue")
	}

	// LockJob
	calledDefs := mock.servers.Defs.MockRefreshIndex(t, &sourcegraph.DefsRefreshIndexOp{
		Repo:                wantRepo,
		RefreshRefLocations: true,
	})
	calledSearch := mock.servers.Search.MockRefreshIndex(t, &sourcegraph.SearchRefreshIndexOp{
		Repos:         []int32{wantRepo},
		RefreshCounts: true,
		RefreshSearch: true,
	})
	err = w.do(ctx, job)
	if err != nil {
		t.Fatal(err)
	}
	if !*calledDefs {
		t.Error("Did not call Defs.RefreshIndex")
	}
	if !*calledSearch {
		t.Error("Did not call Search.RefreshIndex")
	}
}

func TestAsyncWorker(t *testing.T) {
	w := &asyncWorker{}
	ctx, mock := testContext()

	calledLockJob, _, _ := mock.stores.Queue.MockLockJob_Return(t, nil)
	didWork := w.try(ctx)
	if didWork {
		t.Error("did work with an empty queue")
	}
	if !*calledLockJob {
		t.Error("did not call LockJob")
	}

	calledLockJob, calledSuccess, calledError := mock.stores.Queue.MockLockJob_Return(t, &store.Job{Type: "NOOP"})
	didWork = w.try(ctx)
	if !didWork {
		t.Error("did not do work")
	}
	if !*calledLockJob {
		t.Error("did not call LockJob")
	}
	if !*calledSuccess {
		t.Error("job should of succeeded")
	}
	if *calledError {
		t.Error("job should of succeeded")
	}

	calledLockJob, calledSuccess, calledError = mock.stores.Queue.MockLockJob_Return(t, &store.Job{Type: "does not exist"})
	didWork = w.try(ctx)
	if !didWork {
		t.Error("did not do work")
	}
	if !*calledLockJob {
		t.Error("did not call LockJob")
	}
	if *calledSuccess {
		t.Error("job should of failed")
	}
	if !*calledError {
		t.Error("job should of failed")
	}
}

func mustMarshal(t *testing.T, m interface{}) []byte {
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("Failed to json.Marshal %v: %s", m, err)
	}
	return b
}
