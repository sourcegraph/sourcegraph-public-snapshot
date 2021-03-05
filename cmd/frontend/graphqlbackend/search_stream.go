package graphqlbackend

import (
	"context"
	"sync"

	"go.uber.org/atomic"

	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

// SearchEvent is an event on a search stream. It contains fields which can be
// aggregated up into a final result.
type SearchEvent struct {
	Results []SearchResultResolver
	Stats   streaming.Stats
}

// Sender is the interface that wraps the basic Send method. Send must not
// mutate the event.
type Sender interface {
	Send(SearchEvent)
}

type limitStream struct {
	s         Sender
	cancel    context.CancelFunc
	remaining atomic.Int64
}

func (s *limitStream) Send(event SearchEvent) {
	s.s.Send(event)

	var count int64
	for _, r := range event.Results {
		count += int64(r.ResultCount())
	}

	// Avoid limit checks if no change to result count.
	if count == 0 {
		return
	}

	old := s.remaining.Load()
	s.remaining.Sub(count)

	// Only send IsLimitHit once. Can race with other sends and be sent
	// multiple times, but this is fine. Want to avoid lots of noop events
	// after the first IsLimitHit.
	if old >= 0 && s.remaining.Load() < 0 {
		s.s.Send(SearchEvent{Stats: streaming.Stats{IsLimitHit: true}})
		s.cancel()
	}
}

// WithLimit returns a child Stream of parent as well as a child Context of
// ctx. The child stream passes on all events to parent. Once more than limit
// ResultCount are sent on the child stream the context is canceled and an
// IsLimitHit event is sent.
//
// Canceling this context releases resources associated with it, so code
// should call cancel as soon as the operations running in this Context and
// Stream are complete.
func WithLimit(ctx context.Context, parent Sender, limit int) (context.Context, Sender, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	stream := &limitStream{cancel: cancel, s: parent}
	stream.remaining.Store(int64(limit))
	return ctx, stream, cancel
}

// WithSelect returns a child Stream of parent that runs the select operation
// on each event, deduplicating where possible.
func WithSelect(parent Sender, s filter.SelectPath) Sender {
	var mux sync.Mutex
	dedup := NewDeduper()

	return StreamFunc(func(e SearchEvent) {
		mux.Lock()

		selected := e.Results[:0]
		for _, result := range e.Results {
			var current SearchResultResolver
			switch v := result.(type) {
			case *FileMatchResolver:
				current = v.Select(s)
			case *RepositoryResolver:
				current = v.Select(s)
			case *CommitSearchResultResolver:
				current = v.Select(s)
			default:
				current = result
			}

			if current == nil {
				continue
			}

			// If the selected file is a file match, send it unconditionally
			// to ensure we get all line matches for a file.
			_, isFileMatch := current.(*FileMatchResolver)
			seen := dedup.Seen(current)
			if seen && !isFileMatch {
				continue
			}

			dedup.Add(current)
			selected = append(selected, current)
		}
		e.Results = selected

		mux.Unlock()
		if parent != nil {
			parent.Send(e)
		}
	})
}

// StreamFunc is a convenience function to create a stream receiver from a
// function.
type StreamFunc func(SearchEvent)

func (f StreamFunc) Send(event SearchEvent) {
	f(event)
}

// collectStream will call search and aggregates all events it sends. It then
// returns the aggregate event and any error it returns.
func collectStream(search func(Sender) error) ([]SearchResultResolver, streaming.Stats, error) {
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
