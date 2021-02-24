package backend

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/zoekt"
	"github.com/google/zoekt/query"
	"github.com/google/zoekt/rpc"
	zoektstream "github.com/google/zoekt/stream"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

var zoektHTTPClient = &http.Client{
	Transport: &ot.Transport{
		RoundTripper: http.DefaultTransport,
	},
}

// ZoektStreamFunc is a convenience function to create a stream receiver from a
// function.
type ZoektStreamFunc func(*zoekt.SearchResult)

func (f ZoektStreamFunc) Send(event *zoekt.SearchResult) {
	f(event)
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

func (s *StreamSearchAdapter) StreamSearch(ctx context.Context, q query.Q, opts *zoekt.SearchOptions, c zoekt.Sender) error {
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

// Temporarily store if we are on sourcegraph.com mode. We don't want to
// introduce a dependency on frontend/envvar so just duplicating how we do
// this.
var sourcegraphDotComMode, _ = strconv.ParseBool(os.Getenv("SOURCEGRAPHDOTCOM_MODE"))

// ZoektDial connects to a Searcher HTTP RPC server at address (host:port).
func ZoektDial(endpoint string) zoekt.Streamer {
	client := rpc.Client(endpoint)

	batchClient := &StreamSearchAdapter{client}
	streamClient := &zoektStream{
		Searcher: client,
		Client:   zoektstream.NewClient("http://"+endpoint, zoektHTTPClient),
	}

	if !sourcegraphDotComMode {
		return NewMeteredSearcher(endpoint, batchClient)
	}

	// Temporary Sourcegraph.com only mode. For Stefan and Keegan use our new
	// streaming mode.
	ds := &dynamicStreamer{
		Streamers: []zoekt.Streamer{batchClient, streamClient},
		Pick: func(ctx context.Context) zoekt.Streamer {
			uid := actor.FromContext(ctx).UID
			if uid == 7 || uid == 23082 {
				return streamClient
			}
			return batchClient
		},
	}

	return NewMeteredSearcher(endpoint, ds)
}

type zoektStream struct {
	zoekt.Searcher
	*zoektstream.Client
}

// dynamicStreamer picks a Streamer to use at search time based on the
// context.
type dynamicStreamer struct {
	// Streamers is the list of Streamers Pick can return. This is stored here
	// so we can propagate Close.
	Streamers []zoekt.Streamer

	// Pick returns a Streamer to run a search against.
	Pick func(context.Context) zoekt.Streamer
}

func (ds *dynamicStreamer) Search(ctx context.Context, q query.Q, opts *zoekt.SearchOptions) (*zoekt.SearchResult, error) {
	s := ds.Pick(ctx)
	return s.Search(ctx, q, opts)
}

func (ds *dynamicStreamer) List(ctx context.Context, q query.Q) (*zoekt.RepoList, error) {
	s := ds.Pick(ctx)
	return s.List(ctx, q)
}

func (ds *dynamicStreamer) StreamSearch(ctx context.Context, q query.Q, opts *zoekt.SearchOptions, sender zoekt.Sender) (err error) {
	s := ds.Pick(ctx)
	return s.StreamSearch(ctx, q, opts, sender)
}

func (ds *dynamicStreamer) Close() {
	for _, s := range ds.Streamers {
		s.Close()
	}
}

func (ds *dynamicStreamer) String() string {
	var streamers []string
	for _, s := range ds.Streamers {
		streamers = append(streamers, s.String())
	}

	return "DynamicSearcher{" + strings.Join(streamers, " ") + "}"
}
