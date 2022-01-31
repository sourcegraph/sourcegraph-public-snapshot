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

	// eventTransformer can be applied to transform each SearchEvent before it gets
	// propagated.
	//
	// It is currently used to provide sub-repo permissions filtering.
	//
	// SearchEvent is still propagated even in an error case - eventTransformer
	// should make sure the appropriate transformations are made before returning an
	// error.
	eventTransformer EventTransformer

	mu         sync.Mutex
	results    []result.Match
	stats      streaming.Stats
	matchCount int
}

// EventTransformer is a function that is expected to transform search events
type EventTransformer func(event streaming.SearchEvent) (streaming.SearchEvent, error)

// SetEventTransformer sets the event transformer for the aggregator. It is not
// safe for concurrent use and is expected to be called once before the
// aggregator is used.
func (a *Aggregator) SetEventTransformer(et EventTransformer) {
	a.eventTransformer = et
}

// Get finalises aggregation over the stream and returns the aggregated
// result. It should only be called once each do* function is finished
// running.
func (a *Aggregator) Get() ([]result.Match, streaming.Stats, int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.results, a.stats, a.matchCount
}

// Send propagates the given event to the Aggregator's parent stream, or
// aggregates it within results.
//
// It currently also applies sub-repo permissions filtering (see inline docs).
func (a *Aggregator) Send(event streaming.SearchEvent) {
	if a.eventTransformer != nil {
		// TODO: a.Error has been removed, how to proceed?
		event, _ = a.eventTransformer(event)
	}

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
