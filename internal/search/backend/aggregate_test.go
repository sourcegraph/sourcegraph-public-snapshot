package backend

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/zoekt"
	"github.com/google/zoekt/query"
)

func TestAggregateSearcher(t *testing.T) {
	var endpoints atomicMap
	endpoints.Store(prefixMap{})

	searcher := &AggregateSearcher{
		Map: &endpoints,
		Dial: func(endpoint string) zoekt.Searcher {
			var rle zoekt.RepoListEntry
			rle.Repository.Name = endpoint
			return &mockSearcher{
				searchResult: &zoekt.SearchResult{
					Files: []zoekt.FileMatch{{
						Repository: endpoint,
					}},
				},
				listResult: &zoekt.RepoList{Repos: []*zoekt.RepoListEntry{&rle}},
			}
		},
	}
	defer searcher.Close()

	// Start up background goroutines which continously hit the searcher
	// methods to ensure we are safe under concurrency.
	var wg sync.WaitGroup
	defer wg.Wait()
	done := make(chan struct{})
	defer close(done)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				_, err := searcher.Search(context.Background(), nil, nil)
				if err != nil {
					t.Fatal(err)
				}
				_, err = searcher.List(context.Background(), nil)
				if err != nil {
					t.Fatal(err)
				}

				select {
				case <-done:
					return
				default:
				}
			}
		}()
	}

	// each map is the set of servers at a point in time. This is to mainly
	// stress the management code.
	maps := []prefixMap{
		// Start with a normal config of two replicas
		{"1", "2"},

		// Add two
		{"1", "2", "3", "4"},

		// Lose two
		{"2", "4"},

		// Lose and add
		{"1", "2"},

		// Lose all
		{},

		// Lots
		{"1", "2", "3", "4", "5", "6", "7", "8", "9"},
	}

	for _, m := range maps {
		t.Log("current", searcher.String(), "next", m)
		endpoints.Store(m)

		// Our search results should be one per server
		sr, err := searcher.Search(context.Background(), nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, fm := range sr.Files {
			got = append(got, fm.Repository)
		}
		sort.Strings(got)
		want := []string(m)
		if !cmp.Equal(want, got, cmpopts.EquateEmpty()) {
			t.Errorf("search mismatch (-want +got):\n%s", cmp.Diff(want, got))
		}

		// Our list results should be one per server
		rle, err := searcher.List(context.Background(), nil)
		if err != nil {
			t.Fatal(err)
		}
		got = []string{}
		for _, r := range rle.Repos {
			got = append(got, r.Repository.Name)
		}
		sort.Strings(got)
		if !cmp.Equal(want, got, cmpopts.EquateEmpty()) {
			t.Errorf("list mismatch (-want +got):\n%s", cmp.Diff(want, got))
		}
	}

	searcher.Close()
}

type mockSearcher struct {
	searchResult *zoekt.SearchResult
	searchError  error
	listResult   *zoekt.RepoList
	listError    error
}

func (s *mockSearcher) Search(context.Context, query.Q, *zoekt.SearchOptions) (*zoekt.SearchResult, error) {
	return s.searchResult, s.searchError
}

func (s *mockSearcher) List(context.Context, query.Q) (*zoekt.RepoList, error) {
	return s.listResult, s.listError
}

func (*mockSearcher) Close() {}

func (*mockSearcher) String() string {
	return "mockSearcher"
}

type atomicMap struct {
	atomic.Value
}

func (m *atomicMap) Endpoints() (map[string]struct{}, error) {
	return m.Value.Load().(EndpointMap).Endpoints()
}

func (m *atomicMap) GetMany(keys ...string) ([]string, error) {
	return m.Value.Load().(EndpointMap).GetMany(keys...)
}
