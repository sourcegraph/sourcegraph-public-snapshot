package local

import (
	"errors"
	"sync"
	"testing"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/pkg/vcs"
)

// Test that DeltasService.Get returns partial info even if a call
// fails (e.g., it returns the base commit even if the head commit doesn't
// exist).
func TestDeltasService_Get_returnsPartialInfo(t *testing.T) {
	var s deltas
	ctx, mock := testContext()

	wantErr := errors.New("foo")

	var calledGetLock sync.Mutex
	var calledGet int
	mock.servers.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	mock.servers.Builds.MockGetRepoBuild(t, &sourcegraph.Build{})
	mock.servers.Repos.GetCommit_ = func(ctx context.Context, repoRevSpec *sourcegraph.RepoRevSpec) (*vcs.Commit, error) {
		calledGetLock.Lock()
		calledGet++
		calledGetLock.Unlock()
		if repoRevSpec != nil && repoRevSpec.URI == "head" {
			return nil, wantErr
		}
		return &vcs.Commit{}, nil
	}
	ds := new(sourcegraph.DeltaSpec)
	ds.Head.URI = "head"
	delta, err := s.Get(ctx, ds)
	if err.Error() != wantErr.Error() {
		t.Errorf("got error %v, want %v", err, wantErr)
	}
	if delta == nil || delta.BaseCommit == nil {
		t.Errorf("delta.BaseCommit==nil, want non-nil (partial result despite error)")
	}
	if want := 2; calledGet != want {
		t.Errorf("called get %d times, want %d times", calledGet, want)
	}
}
