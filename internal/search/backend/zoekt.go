package backend

import (
	"context"

	"github.com/google/zoekt"
	"github.com/google/zoekt/query"
)

type ZoektStreamer interface {
	Send(*zoekt.SearchResult)
}

type ZoektStreamerFunc chan<- *zoekt.SearchResult

func (c ZoektStreamerFunc) Send(res *zoekt.SearchResult) {
	c <- res
}

type ZoektStreamObserver func(*zoekt.SearchResult)

type ZoektStreamerWithObserver struct {
	s ZoektStreamer
	o ZoektStreamObserver
}

func (s *ZoektStreamerWithObserver) Send(event *zoekt.SearchResult) {
	s.o(event)
	s.s.Send(event)
}

func WithObserver(s ZoektStreamer, o ZoektStreamObserver) ZoektStreamer {
	return &ZoektStreamerWithObserver{s, o}
}

// StreamSearcher is an optional interface which sends results over a channel
// as they are found.
//
// This is a Sourcegraph extension.
type StreamSearcher interface {
	zoekt.Searcher

	// StreamSearch returns a channel which needs to be read until closed.
	StreamSearch(ctx context.Context, q query.Q, opts *zoekt.SearchOptions, c ZoektStreamer) error
}

// StreamSearchEvent has fields optionally set representing events that happen
// during a search.
//
// This is a Sourcegraph extension.
type StreamSearchEvent struct {
	// SearchResult is non-nil if this event is a search result. These should be
	// combined with previous and later SearchResults.
	SearchResult *zoekt.SearchResult
}

// StreamSearchAdapter adapts a zoekt.Searcher to conform to the StreamSearch
// interface by calling zoekt.Searcher.Search.
type StreamSearchAdapter struct {
	zoekt.Searcher
}

func (s *StreamSearchAdapter) StreamSearch(ctx context.Context, q query.Q, opts *zoekt.SearchOptions, c ZoektStreamer) error {
	sr, err := s.Search(ctx, q, opts)
	if err != nil {
		return err
	}
	c.Send(sr)
	return nil
}

func (s *StreamSearchAdapter) String() string {
	return "streamSearchAdapter{" + s.Searcher.String() + "}"
}
