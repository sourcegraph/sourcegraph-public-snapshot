package graphqlbackend

import (
	"context"

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

// collectStream is a helper for batch interfaces calling stream based
// functions. It returns a context, stream and cleanup/get function. The
// cleanup/get function will return the aggregated event and must be called
// once you have stopped sending to stream.
func collectStream(ctx context.Context) (context.Context, Streamer, func() SearchEvent) {
	var agg SearchEvent

	ctx, cancel := context.WithCancel(ctx)

	done := make(chan struct{})
	stream := make(chan SearchEvent)
	go func() {
		defer close(done)
		for event := range stream {
			agg.Results = append(agg.Results, event.Results...)
			agg.Stats.Update(&event.Stats)
		}
	}()

	return ctx, SearchStream(stream), func() SearchEvent {
		cancel()
		close(stream)
		<-done
		return agg
	}
}
