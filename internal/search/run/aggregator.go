package run

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

func NewAggregator(stream streaming.Sender) *Aggregator {
	return &Aggregator{
		parentStream: stream,
	}
}

type Aggregator struct {
	parentStream streaming.Sender

	mu         sync.Mutex
	results    result.Matches
	stats      streaming.Stats
	matchCount int
}

// Get finalises aggregation over the stream and returns the aggregated
// result. It should only be called once each do* function is finished
// running.
func (a *Aggregator) Get() (result.Matches, streaming.Stats, int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.results, a.stats, a.matchCount
}

// Send propagates the given event to the Aggregator's parent stream, or
// aggregates it within results.
//
// It currently also applies sub-repo permissions filtering (see inline docs).
func (a *Aggregator) Send(event streaming.SearchEvent) {
	if a.parentStream != nil {
		a.parentStream.Send(event)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Only aggregate results if we are not streaming.
	if a.parentStream == nil {
		a.results = append(a.results, event.Results...)

		if a.stats.Repos == nil {
			a.stats.Repos = make(map[api.RepoID]struct{})
		}

		for _, r := range event.Results {
			repo := r.RepoName()
			if _, ok := a.stats.Repos[repo.ID]; !ok {
				a.stats.Repos[repo.ID] = struct{}{}
			}
		}
	}

	a.matchCount += len(event.Results)
	a.stats.Update(&event.Stats)
}
