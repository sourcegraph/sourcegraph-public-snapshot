package backend

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/rcache"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
)

func TestAsyncService_RefreshIndexes(t *testing.T) {
	rcache.SetupForTest("TestAsyncService_RefreshIndexes")

	s := &async{}
	w := &asyncWorker{}
	ctx := testContext()

	wantRepo := int32(10810)

	Mocks.Repos.MockResolveRev_NoCheck(t, vcs.CommitID("deadbeef"))
	Mocks.Repos.GetInventory = func(v0 context.Context, v1 *sourcegraph.RepoRevSpec) (*inventory.Inventory, error) {
		return &inventory.Inventory{Languages: []*inventory.Lang{{Name: "Go"}}}, nil
	}

	// Enqueue
	op := &sourcegraph.AsyncRefreshIndexesOp{
		Repo:   wantRepo,
		Source: "test",
	}
	job := &localstore.Job{
		Type: "RefreshIndexes",
		Args: mustMarshal(t, &sourcegraph.AsyncRefreshIndexesOp{
			Repo:   wantRepo,
			Source: "test (UID 1 test)",
		}),
	}
	calledEnqueue := localstore.Mocks.Queue.MockEnqueue(t, job)
	err := s.RefreshIndexes(ctx, op)
	if err != nil {
		t.Fatal(err)
	}
	if !*calledEnqueue {
		t.Fatal("async.RefreshIndexes did not call Enqueue")
	}

	// LockJob
	calledDefs := Mocks.Defs.MockRefreshIndex(t, &sourcegraph.DefsRefreshIndexOp{
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
	ctx := testContext()

	calledLockJob, _, _ := localstore.Mocks.Queue.MockLockJob_Return(t, nil)
	didWork := w.try(ctx)
	if didWork {
		t.Error("did work with an empty queue")
	}
	if !*calledLockJob {
		t.Error("did not call LockJob")
	}

	calledLockJob, calledSuccess, calledError := localstore.Mocks.Queue.MockLockJob_Return(t, &localstore.Job{Type: "NOOP"})
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

	calledLockJob, calledSuccess, calledError = localstore.Mocks.Queue.MockLockJob_Return(t, &localstore.Job{Type: "does not exist"})
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
	ctxParent := testContext()
	ctx1 := context.WithValue(ctxParent, "source", 1)
	ctx2 := context.WithValue(ctxParent, "source", 2)

	wantRepo := int32(10811)
	op := &sourcegraph.AsyncRefreshIndexesOp{
		Repo:   wantRepo,
		Source: "test",
	}

	// Do a call to RefreshIndexes, but that blocks until wait1 is
	// closed. That way we can do another call concurrently and ensure the
	// mutex blocks.
	called1 := make(chan interface{})
	called2 := false
	wait1 := make(chan interface{})
	done1 := make(chan interface{})
	Mocks.Defs.RefreshIndex = func(ctx context.Context, op *sourcegraph.DefsRefreshIndexOp) error {
		switch ctx.Value("source").(int) {
		case 1:
			close(called1)
			<-wait1
			return nil
		case 2:
			called2 = true
			wantOp := &sourcegraph.DefsRefreshIndexOp{
				Repo:                wantRepo,
				RefreshRefLocations: true,
			}
			if !reflect.DeepEqual(op, wantOp) {
				t.Fatalf("unexpected DefsRefreshIndexOp, got %+v != %+v", op, wantOp)
			}
			return nil
		default:
			t.Fatal("unexpected ctx")
		}
		return errors.New("unreachable")
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
	calledEnqueue := localstore.Mocks.Queue.MockEnqueue(t, &localstore.Job{
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
	if called2 {
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
