package backend

import (
	"context"

	"github.com/google/zoekt"
	"github.com/google/zoekt/query"
)

// StreamSearcher is an optional interface which sends results over a channel
// as they are found.
//
// This is a Sourcegraph extension.
type StreamSearcher interface {
	zoekt.Searcher

	// StreamSearch returns a channel which needs to be read until closed.
	StreamSearch(ctx context.Context, q query.Q, opts *zoekt.SearchOptions) <-chan StreamSearchEvent
}

// StreamSearchEvent has fields optionally set representing events that happen
// during a search.
//
// This is a Sourcegraph extension.
type StreamSearchEvent struct {
	// SearchResult is non-nil if this event is a search result. These should be
	// combined with previous and later SearchResults.
	SearchResult *zoekt.SearchResult
	// Error indicates an error was encountered.
	Error error
}

// StreamSearchAdapter adapts a zoekt.Searcher to conform to the StreamSearch
// interface by calling zoekt.Searcher.Search.
type StreamSearchAdapter struct {
	zoekt.Searcher
}

func (s *StreamSearchAdapter) StreamSearch(ctx context.Context, q query.Q, opts *zoekt.SearchOptions) <-chan StreamSearchEvent {
	c := make(chan StreamSearchEvent)

	go func() {
		defer close(c)
		sr, err := s.Search(ctx, q, opts)
		c <- StreamSearchEvent{
			SearchResult: sr,
			Error:        err,
		}
	}()

	return c
}

func (s *StreamSearchAdapter) String() string {
	return "streamSearchAdapter{" + s.Searcher.String() + "}"
}
