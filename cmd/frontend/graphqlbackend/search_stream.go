package graphqlbackend

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
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
	s      Streamer
	cancel func()

	mu    sync.Mutex
	limit int
	count int
}

// Send sends an event on the stream. If the limit is reached, a final event with
// IsLimitHit = true is sent.
func (s *limitStream) Send(event SearchEvent) {
	s.s.Send(event)

	// Avoid limit checks if no change to result count.
	if len(event.Results) == 0 {
		return
	}

	s.mu.Lock()
	oldCount := s.count
	s.count += len(event.Results)
	newCount := s.count
	s.mu.Unlock()

	// Only send IsLimitHit once
	if newCount > s.limit && oldCount <= s.limit {
		s.s.Send(SearchEvent{Stats: streaming.Stats{IsLimitHit: true}})
		s.cancel()
	}
}

// WithLimit returns a Streamer and a context. The Streamer is limited to `limit`
// results. Once more than `limit` results have been sent on the stream, the
// returned context is canceled.
func WithLimit(ctx context.Context, stream Streamer, limit int) (Streamer, context.Context, func()) {
	cancelContext, cancel := context.WithCancel(ctx)
	cleanup := func() {
		cancel()
	}
	newLimitStream := &limitStream{cancel: cancel, s: stream, limit: limit}
	return newLimitStream, cancelContext, cleanup
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
