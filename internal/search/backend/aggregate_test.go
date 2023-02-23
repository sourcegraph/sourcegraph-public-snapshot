package backend

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/zoekt"
)

func TestFlushCollectSender(t *testing.T) {
	replicas := prefixMap{"1", "2", "3", "4", "5", "6"}
	nonemptyEndpoints := 4

	var endpoints atomicMap
	endpoints.Store(replicas)

	searcher := &HorizontalSearcher{
		Map: &endpoints,
		Dial: func(endpoint string) zoekt.Streamer {
			endpointID, _ := strconv.Atoi(endpoint)
			if endpointID > nonemptyEndpoints {
				return &FakeStreamer{}
			}

			repoList := make([]*zoekt.RepoListEntry, 3)
			results := make([]*zoekt.SearchResult, 3)

			for i := 0; i < len(results); i++ {
				repoID := 100*endpointID + i
				repoName := strconv.Itoa(repoID)

				results[i] = &zoekt.SearchResult{
					Files: []zoekt.FileMatch{{
						Score:              float64(repoID),
						RepositoryPriority: float64(repoID),
						Repository:         repoName,
					},
					}}

				repoList[i] = &zoekt.RepoListEntry{
					Repository: zoekt.Repository{
						Name: repoName,
						ID:   uint32(repoID),
					},
				}
			}

			return &FakeStreamer{
				Results: results,
				Repos:   repoList,
			}
		},
	}
	defer searcher.Close()

	// Start up background goroutines which continuously hit the searcher
	// methods to ensure we are safe under concurrency.
	for i := 0; i < 5; i++ {
		cleanup := backgroundSearch(searcher)
		defer cleanup(t)
	}

	opts := zoekt.SearchOptions{
		UseDocumentRanks: true,
		FlushWallTime:    100 * time.Millisecond,
	}

	// Collect all search results in order they were sent to stream
	var results []*zoekt.SearchResult
	err := searcher.StreamSearch(context.Background(), nil, &opts,
		ZoektStreamFunc(func(event *zoekt.SearchResult) { results = append(results, event) }))
	if err != nil {
		t.Fatal(err)
	}

	// Check the aggregated result was flushed early
	if len(results) == 0 {
		t.Fatal("no results returned from search")
	}
	if results[0].Stats.FlushReason != zoekt.FlushReasonTimerExpired {
		t.Fatalf("expected flush reason %s but got %s", zoekt.FlushReasonTimerExpired, results[0].Stats.FlushReason)
	}

	// CHeck that the results were streamed in the expected order
	var repos []string
	for _, r := range results {
		if r.Files != nil {
			for _, f := range r.Files {
				repos = append(repos, f.Repository)
			}
		}
	}

	expectedRepos := nonemptyEndpoints * 3
	if len(repos) != expectedRepos {
		t.Fatalf("expected %d results but got %d", expectedRepos, len(repos))
	}

	// The first results should always include one result per endpoint, ordered by score
	want := []string{"400", "300", "200", "100"}
	if !cmp.Equal(want, repos[:nonemptyEndpoints]) {
		t.Errorf("search mismatch (-want +got):\n%s", cmp.Diff(want, repos))
	}
}
