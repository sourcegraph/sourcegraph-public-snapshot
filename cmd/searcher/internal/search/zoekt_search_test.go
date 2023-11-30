package search

import (
	"context"
	"testing"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/zoekt"
	"github.com/sourcegraph/zoekt/query"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type mockClient struct {
	zoekt.Streamer
	mockStreamSearch func(context.Context, query.Q, *zoekt.SearchOptions, zoekt.Sender) error
}

func (mc *mockClient) StreamSearch(ctx context.Context, q query.Q, opts *zoekt.SearchOptions, sender zoekt.Sender) (err error) {
	return mc.mockStreamSearch(ctx, q, opts, sender)
}

func Test_zoektSearch(t *testing.T) {
	ctx := context.Background()

	// Create a mock client that will send a few files worth of matches
	client := &mockClient{
		mockStreamSearch: func(ctx context.Context, q query.Q, so *zoekt.SearchOptions, s zoekt.Sender) error {
			for i := 0; i < 10; i++ {
				s.Send(&zoekt.SearchResult{
					Files: []zoekt.FileMatch{{}, {}},
				})
			}
			return nil
		},
	}

	// Structural search fails immediately, so can't consume the events from the zoekt stream
	mockStructuralSearch = func(ctx context.Context, inputType comby.Input, paths filePatterns, extensionHint, pattern, rule string, languages []string, repo api.RepoName, sender matchSender) error {
		return errors.New("oops")
	}
	t.Cleanup(func() { mockStructuralSearch = nil })

	// Ensure that this returns an error from structuralSearch, and does not block
	// indefinitely because the reader returns early.
	err := zoektSearch(
		ctx,
		logtest.Scoped(t),
		client,
		&search.TextPatternInfo{},
		[]query.BranchRepos{{Branch: "test", Repos: roaring.BitmapOf(1, 2, 3)}},
		time.Since,
		"",
		matchSender(nil),
	)
	require.Error(t, err)
}
