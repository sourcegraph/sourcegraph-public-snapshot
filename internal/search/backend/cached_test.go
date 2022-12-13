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
		FakeSearcher: FakeSearcher{Repos: []*zoekt.RepoListEntry{
			{Repository: zoekt.Repository{ID: 1, Name: "foo"}},
			{Repository: zoekt.Repository{ID: 2, Name: "bar", HasSymbols: true}},
		}},
	}

	ttl := 30 * time.Second
	s := NewCachedSearcher(ttl, ms).(*cachedSearcher)

	now := time.Now()
	s.now = func() time.Time { return now }

	ctx := context.Background()

	s.List(ctx, &zoektquery.Const{Value: true}, &zoekt.ListOptions{Minimal: true})

	have, _ := s.List(ctx, &zoektquery.Const{Value: true}, &zoekt.ListOptions{Minimal: true})
	want := &zoekt.RepoList{
		Minimal: map[uint32]*zoekt.MinimalRepoListEntry{
			1: {},
			2: {HasSymbols: true},
		},
	}

	if !cmp.Equal(have, want) {
		t.Fatalf("list mismatch: %s", cmp.Diff(have, want))
	}

	if have, want := atomic.LoadInt64(&ms.ListCalls), int64(1); have != want {
		t.Fatalf("have ListCalls %d, want %d", have, want)
	}

	atomic.StoreInt64(&ms.ListCalls, 0)

	s.List(ctx, &zoektquery.Const{Value: true}, nil)

	have, _ = s.List(ctx, &zoektquery.Const{Value: true}, nil)
	want = &zoekt.RepoList{Repos: ms.FakeSearcher.Repos}

	diffOpts := cmpopts.IgnoreUnexported(zoekt.Repository{})
	if d := cmp.Diff(want, have, diffOpts); d != "" {
		t.Fatalf("list mismatch: %s", d)
	}

	if have, want := atomic.LoadInt64(&ms.ListCalls), int64(1); have != want {
		t.Fatalf("have ListCalls %d, want %d", have, want)
	}

	atomic.StoreInt64(&ms.ListCalls, 0)
	now = now.Add(ttl)
	ms.FakeSearcher.Repos = append(ms.FakeSearcher.Repos, &zoekt.RepoListEntry{Repository: zoekt.Repository{ID: 3, Name: "baz"}})

	for {
		have, _ = s.List(ctx, &zoektquery.Const{Value: true}, nil)
		want = &zoekt.RepoList{Repos: ms.FakeSearcher.Repos}

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
	FakeSearcher
	ListCalls int64
}

func (s *mockUncachedSearcher) List(ctx context.Context, q zoektquery.Q, opts *zoekt.ListOptions) (*zoekt.RepoList, error) {
	atomic.AddInt64(&s.ListCalls, 1)
	return s.FakeSearcher.List(ctx, q, opts)
}
