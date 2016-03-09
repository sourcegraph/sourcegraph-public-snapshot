package local

import (
	"errors"
	"sync"
	"testing"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// Test that DeltasService.Get returns partial info even if a call
// fails (e.g., it returns the base repo even if the head repo doesn't
// exist).
func TestDeltasService_Get_returnsPartialInfo(t *testing.T) {
	var s deltas
	ctx, mock := testContext()

	wantErr := errors.New("foo")

	var calledGetLock sync.Mutex
	var calledGet int
	mock.servers.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	mock.servers.Builds.MockGetRepoBuild(t, &sourcegraph.Build{})
	mock.servers.Repos.Get_ = func(ctx context.Context, repoSpec *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		calledGetLock.Lock()
		calledGet++
		calledGetLock.Unlock()
		if repoSpec != nil && repoSpec.URI == "head" {
			return nil, wantErr
		}
		return &sourcegraph.Repo{URI: "x", DefaultBranch: "b"}, nil
	}
	ds := new(sourcegraph.DeltaSpec)
	ds.Head.URI = "head"
	delta, err := s.Get(ctx, ds)
	if err.Error() != wantErr.Error() {
		t.Errorf("got error %v, want %v", err, wantErr)
	}
	if delta == nil || delta.BaseRepo == nil {
		t.Errorf("delta.BaseRepo==nil, want non-nil (partial result despite error)")
	}
	if want := 2; calledGet != want {
		t.Errorf("called get %d times, want %d times", calledGet, want)
	}
}
