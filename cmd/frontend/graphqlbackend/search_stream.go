package graphqlbackend

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"go.uber.org/atomic"
)

// SearchEvent is an event on a search stream. It contains fields which can be
// aggregated up into a final result.
type SearchEvent struct {
	Results []SearchResultResolver
	Stats   streaming.Stats
}

// Streamer is the interface that wraps the basic Send method. Send must not
// mutate the event.
type Streamer interface {
	Send(SearchEvent)
}

type SearchStream chan<- SearchEvent

func (s SearchStream) Send(event SearchEvent) {
	s <- event
}

type limitStream struct {
	s         Streamer
	cancel    context.CancelFunc
	remaining atomic.Int64
}

// Send sends an event on the stream. If the limit is reached, a final event with
// IsLimitHit = true is sent.
func (s *limitStream) Send(event SearchEvent) {
	s.s.Send(event)

	// Avoid limit checks if no change to result count.
	if len(event.Results) == 0 {
		return
	}

	old := s.remaining.Load()
	s.remaining.Sub(int64(len(event.Results)))

	// Only send IsLimitHit once. Can race with other sends and be sent
	// multiple times, but this is fine. Want to avoid lots of noop events
	// after the first IsLimitHit.
	if old >= 0 && s.remaining.Load() < 0 {
		s.s.Send(SearchEvent{Stats: streaming.Stats{IsLimitHit: true}})
		s.cancel()
	}
}

// WithLimit returns a Streamer and a context. The Streamer is limited to `limit`
// results. Once more than `limit` results have been sent on the stream, the
// returned context is canceled.
func WithLimit(ctx context.Context, stream Streamer, limit int) (context.Context, Streamer, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	newLimitStream := &limitStream{cancel: cancel, s: stream}
	newLimitStream.remaining.Store(int64(limit))
	return ctx, newLimitStream, cancel
}

// StreamFunc is a convenience function to create a stream receiver from a
// function.
type StreamFunc func(SearchEvent)

func (f StreamFunc) Send(event SearchEvent) {
	f(event)
}

// collectStream will call search and aggregates all events it sends. It then
// returns the aggregate event and any error it returns.
func collectStream(search func(Streamer) error) ([]SearchResultResolver, streaming.Stats, error) {
	var (
		mu      sync.Mutex
		results []SearchResultResolver
		stats   streaming.Stats
	)

	err := search(StreamFunc(func(event SearchEvent) {
		mu.Lock()
		results = append(results, event.Results...)
		stats.Update(&event.Stats)
		mu.Unlock()
	}))

	return results, stats, err
}
