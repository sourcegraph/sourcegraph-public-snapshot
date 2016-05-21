package local

import (
	"errors"
	"sync"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
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
	mock.servers.Repos.GetCommit_ = func(ctx context.Context, repoRevSpec *sourcegraph.RepoRevSpec) (*vcs.Commit, error) {
		calledGetLock.Lock()
		calledGet++
		calledGetLock.Unlock()
		if repoRevSpec != nil && repoRevSpec.CommitID == "head" {
			return nil, wantErr
		}
		return &vcs.Commit{}, nil
	}
	ds := new(sourcegraph.DeltaSpec)
	ds.Head.CommitID = "head"
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

func TestDeltasCacheKeyDeterministic(t *testing.T) {
	spec := new(sourcegraph.DeltaSpec)
	spec.Base.URI = "base"
	spec.Base.CommitID = "base-commit"
	spec.Head.URI = "head"
	spec.Head.CommitID = "head-commit"
	k := deltasCacheKey(spec)
	for i := 0; i < 100; i++ {
		x := deltasCacheKey(spec)
		if k != x {
			t.Errorf("Deltas cache key is not determinstic got %q, expected %q", x, k)
		}
	}
}

func TestDeltasCacheInvalid(t *testing.T) {
	cache := newDeltasCache(1)
	delta := new(sourcegraph.Delta)

	spec := new(sourcegraph.DeltaSpec)
	spec.Base.CommitID = "base-commit"
	cache.Add(spec, delta)
	_, ok := cache.Get(spec)
	if ok {
		t.Error("Invalid spec cached head commit id not set")
	}

	spec.Base.CommitID = ""
	spec.Head.CommitID = "head-commit"
	cache.Add(spec, delta)
	_, ok = cache.Get(spec)
	if ok {
		t.Error("Invalid spec cached base commit id not set")
	}
}

func TestDeltasCache(t *testing.T) {
	cache := newDeltasCache(1)
	delta := new(sourcegraph.Delta)
	delta.Head.URI = "thedelta"

	spec := new(sourcegraph.DeltaSpec)
	spec.Base.CommitID = "base-commit"
	spec.Head.CommitID = "head-commit"
	cache.Add(spec, delta)
	hit, ok := cache.Get(spec)
	if !ok {
		t.Error("Delta not cached")
	}
	if hit.Head.URI != "thedelta" {
		t.Errorf("Cache hit was %q", hit.Head.URI)
	}

	// Test eviction.
	otherspec := new(sourcegraph.DeltaSpec)
	otherspec.Base.CommitID = "base-other-commit"
	otherspec.Head.CommitID = "head-other-commit"
	cache.Add(otherspec, delta)
	_, ok = cache.Get(spec)
	if ok {
		t.Error("Delta should have been evicted")
	}
}
