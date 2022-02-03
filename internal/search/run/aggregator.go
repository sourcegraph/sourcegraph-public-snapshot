package run

import (
	"sync"

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
	stats      streaming.Stats
	matchCount int
}

// Get finalises aggregation over the stream and returns the aggregated
// result. It should only be called once each do* function is finished
// running.
func (a *Aggregator) Get() (streaming.Stats, int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.stats, a.matchCount
}

// Send propagates the given event to the Aggregator's parent stream, or
// aggregates it within results.
//
// It currently also applies sub-repo permissions filtering (see inline docs).
func (a *Aggregator) Send(event streaming.SearchEvent) {
	a.parentStream.Send(event)

	a.mu.Lock()
	a.matchCount += len(event.Results)
	a.stats.Update(&event.Stats)
	a.mu.Unlock()
}
