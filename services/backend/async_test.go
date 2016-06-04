package backend

import (
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
)

func TestAsyncService_RefreshIndexes(t *testing.T) {
	s := &async{}
	ctx, mock := testContext()

	wantRepo := int32(10810)
	calledDefs := make(chan bool, 1)
	calledSearch := make(chan bool, 1)
	mock.servers.Defs.RefreshIndex_ = func(ctx context.Context, op *sourcegraph.DefsRefreshIndexOp) (*pbtypes.Void, error) {
		if op.Repo != wantRepo {
			t.Errorf("unexpected def repo, got %v != %v", op.Repo, wantRepo)
		}
		calledDefs <- true
		return nil, nil
	}
	mock.servers.Search.RefreshIndex_ = func(ctx context.Context, op *sourcegraph.SearchRefreshIndexOp) (*pbtypes.Void, error) {
		if !reflect.DeepEqual(op.Repos, []int32{wantRepo}) {
			t.Errorf("unexpected def repo, got %v != []int32{%v}", op.Repos, wantRepo)
		}
		calledSearch <- true
		return nil, nil
	}

	_, err := s.RefreshIndexes(ctx, &sourcegraph.AsyncRefreshIndexesOp{Repo: wantRepo})
	if err != nil {
		t.Fatal(err)
	}

	timeout := time.After(100 * time.Millisecond)
	select {
	case <-calledDefs:
		break
	case <-timeout:
		t.Fatal("waiting for calledDefs timed out")
	}

	select {
	case <-calledSearch:
		break
	case <-timeout:
		t.Fatal("waiting for calledSearch timed out")
	}
}
