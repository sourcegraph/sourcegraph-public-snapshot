package backend

import (
	"encoding/json"
	"testing"
	"time"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/rcache"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sqs/pbtypes"
)

func TestAsyncService_RefreshIndexes(t *testing.T) {
	rcache.SetupForTest("TestAsyncService_RefreshIndexes")

	s := &async{}
	w := &asyncWorker{}
	ctx, mock := testContext()

	wantRepo := int32(10810)

	mock.servers.Repos.MockResolveRev_NoCheck(t, vcs.CommitID("deadbeef"))
	mock.servers.Repos.GetInventory_ = func(v0 context.Context, v1 *sourcegraph.RepoRevSpec) (*inventory.Inventory, error) {
		return &inventory.Inventory{Languages: []*inventory.Lang{{Name: "Go"}}}, nil
	}

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
	err = w.do(ctx, job)
	if err != nil {
		t.Fatal(err)
	}
	if !*calledDefs {
		t.Error("Did not call Defs.RefreshIndex")
	}
}

func TestAsyncWorker(t *testing.T) {
	rcache.SetupForTest("TestAsyncWorker")

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

func TestAsyncWorker_mutex(t *testing.T) {
	// TODO(keegancsmith) distributed locking should be a store, so we can
	// mock it
	rcache.SetupForTest("TestAsyncWorker_mutex")

	w := &asyncWorker{}
	ctx1, mock1 := testContext()
	ctx2, mock2 := testContext()

	wantRepo := int32(10811)
	op := &sourcegraph.AsyncRefreshIndexesOp{
		Repo:   wantRepo,
		Source: "test",
	}

	// Do a call to RefreshIndexes, but that blocks until wait1 is
	// closed. That way we can do another call concurrently and ensure the
	// mutex blocks.
	called1 := make(chan interface{})
	wait1 := make(chan interface{})
	done1 := make(chan interface{})
	mock1.servers.Defs.RefreshIndex_ = func(ctx context.Context, op *sourcegraph.DefsRefreshIndexOp) (*pbtypes.Void, error) {
		close(called1)
		<-wait1
		return nil, nil
	}
	go func() {
		err := w.refreshIndexes(ctx1, op)
		if err != nil {
			t.Fatal(err)
		}
		close(done1)
	}()
	<-called1

	// Now we should fail to acquire the mutex -> do not run Defs.RefreshIndex
	called2 := mock2.servers.Defs.MockRefreshIndex(t, &sourcegraph.DefsRefreshIndexOp{
		Repo:                wantRepo,
		RefreshRefLocations: true,
	})
	calledEnqueue := mock2.stores.Queue.MockEnqueue(t, &store.Job{
		Type: "RefreshIndexes",
		Args: mustMarshal(t, &sourcegraph.AsyncRefreshIndexesOp{
			Repo:   wantRepo,
			Source: "test (mutex)",
		}),
		Delay: 10 * time.Minute,
	})
	err := w.refreshIndexes(ctx2, op)
	if err != nil {
		t.Fatal(err)
	}
	if !*calledEnqueue {
		t.Error("second refreshIndexes did not enqueue job for later")
	}
	if *called2 {
		t.Error("second refreshIndexes should not run Defs.RefreshIndex")
	}

	close(wait1)
	<-done1
}

func mustMarshal(t *testing.T, m interface{}) []byte {
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("Failed to json.Marshal %v: %s", m, err)
	}
	return b
}
