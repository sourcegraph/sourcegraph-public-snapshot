package repos

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestCachedList(t *testing.T) {
	want := []*types.RepoName{{
		ID:   api.RepoID(1),
		Name: "github.com/foo/bar",
	}, {
		ID:   api.RepoID(2),
		Name: "github.com/foo/baz",
	}}

	cache := &CachedList{}

	count := 0
	f := cache.Wrap(func(ctx context.Context) ([]*types.RepoName, error) {
		count++
		return want, nil
	})

	get := func() []*types.RepoName {
		t.Helper()
		got, err := f(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if d := cmp.Diff(want, got); d != "" {
			t.Fatalf("(-want, +got):\n%s", d)
		}
		return got
	}

	// We should only call underlying list once. Note: this could race with
	// write/reads on count, but that is only if we the underlying code is
	// broken.
	for i := 0; i < 100; i++ {
		_ = get()
	}

	// Check that mutating the array doesn't affect other callers.
	bad := get()
	bad[0], bad[1] = bad[1], bad[0]

	got := get()
	if cmp.Equal(bad, got) {
		t.Fatal("mutating response mutated cached value")
	}

	if count != 1 {
		t.Fatalf("expected to call fill function once, called %d times", count)
	}
}
