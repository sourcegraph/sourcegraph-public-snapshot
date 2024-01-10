package backend

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/zoekt"
	zoektquery "github.com/sourcegraph/zoekt/query"
)

func TestCachedSearcher(t *testing.T) {
	ms := &mockUncachedSearcher{
		FakeStreamer: FakeStreamer{Repos: []*zoekt.RepoListEntry{
			{Repository: zoekt.Repository{ID: 1, Name: "foo"}},
			{Repository: zoekt.Repository{ID: 2, Name: "bar", HasSymbols: true}},
		}},
	}

	ttl := 30 * time.Second
	s := NewCachedSearcher(ttl, ms).(*cachedSearcher)

	now := time.Now()
	s.now = func() time.Time { return now }

	ctx := context.Background()

	// RepoListFieldReposMap
	{
		s.List(ctx, &zoektquery.Const{Value: true}, &zoekt.ListOptions{Field: zoekt.RepoListFieldReposMap})

		have, _ := s.List(ctx, &zoektquery.Const{Value: true}, &zoekt.ListOptions{Field: zoekt.RepoListFieldReposMap})
		want := &zoekt.RepoList{
			ReposMap: zoekt.ReposMap{
				1: {},
				2: {HasSymbols: true},
			},
			Stats: zoekt.RepoStats{
				Repos: 2,
			},
		}

		if !cmp.Equal(have, want) {
			t.Fatalf("list mismatch: %s", cmp.Diff(have, want))
		}

		if have, want := atomic.LoadInt64(&ms.ListCalls), int64(1); have != want {
			t.Fatalf("have ListCalls %d, want %d", have, want)
		}

		atomic.StoreInt64(&ms.ListCalls, 0)
	}

	diffOpts := cmpopts.IgnoreUnexported(zoekt.Repository{})

	// RepoListFieldRepos
	{
		s.List(ctx, &zoektquery.Const{Value: true}, nil)

		have, _ := s.List(ctx, &zoektquery.Const{Value: true}, nil)
		want := &zoekt.RepoList{
			Repos: ms.FakeStreamer.Repos,
			Stats: zoekt.RepoStats{
				Repos: len(ms.FakeStreamer.Repos),
			},
		}

		if d := cmp.Diff(want, have, diffOpts); d != "" {
			t.Fatalf("list mismatch: %s", d)
		}

		if have, want := atomic.LoadInt64(&ms.ListCalls), int64(1); have != want {
			t.Fatalf("have ListCalls %d, want %d", have, want)
		}

		atomic.StoreInt64(&ms.ListCalls, 0)
	}

	// Now test the cache does invalidate. We only do this for one type of
	// field since it should cover all field types.
	now = now.Add(ttl)
	ms.FakeStreamer.Repos = append(ms.FakeStreamer.Repos, &zoekt.RepoListEntry{Repository: zoekt.Repository{ID: 3, Name: "baz"}})

	for {
		have, _ := s.List(ctx, &zoektquery.Const{Value: true}, nil)
		want := &zoekt.RepoList{
			Repos: ms.FakeStreamer.Repos,
			Stats: zoekt.RepoStats{
				Repos: len(ms.FakeStreamer.Repos),
			},
		}

		if !cmp.Equal(have, want, diffOpts) {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		break
	}

	if have, want := atomic.LoadInt64(&ms.ListCalls), int64(1); have != want {
		t.Fatalf("have ListCalls %d, want %d", have, want)
	}
}

type mockUncachedSearcher struct {
	testing.TB
	FakeStreamer
	ListCalls int64
}

func (s *mockUncachedSearcher) List(ctx context.Context, q zoektquery.Q, opts *zoekt.ListOptions) (*zoekt.RepoList, error) {
	atomic.AddInt64(&s.ListCalls, 1)
	return s.FakeStreamer.List(ctx, q, opts)
}
