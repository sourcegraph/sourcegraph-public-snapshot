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
	"github.com/sourcegraph/sourcegraph/internal/searcher/protocol"
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
			for range 10 {
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
		&protocol.PatternInfo{},
		[]query.BranchRepos{{Branch: "test", Repos: roaring.BitmapOf(1, 2, 3)}},
		0,
		time.Since,
		"",
		matchSender(nil),
	)
	require.Error(t, err)
}

func TestHandleSearchFilters(t *testing.T) {
	tests := []struct {
		name        string
		patternInfo *protocol.PatternInfo
		expectedQ   query.Q
		expectedErr error
	}{
		{
			name: "Include and exclude paths",
			patternInfo: &protocol.PatternInfo{
				IncludePaths:    []string{"\\.go", "cmd"},
				ExcludePaths:    "vendor/",
				IsCaseSensitive: false,
			},
			expectedQ: query.NewAnd(
				&query.Substring{Pattern: ".go", FileName: true},
				&query.Substring{Pattern: "cmd", FileName: true},
				&query.Not{Child: &query.Substring{Pattern: "vendor/", FileName: true}},
			),
			expectedErr: nil,
		},
		{
			name: "Include and exclude languages",
			patternInfo: &protocol.PatternInfo{
				IncludePaths: []string{"cmd"},
				IncludeLangs: []string{"go", "typescript"},
				ExcludeLangs: []string{"javascript"},
			},
			expectedQ: query.NewAnd(
				&query.Substring{Pattern: "cmd", FileName: true},
				&query.Language{Language: "go"},
				&query.Language{Language: "typescript"},
				&query.Not{Child: &query.Language{Language: "javascript"}},
			),
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := handleSearchFilters(tt.patternInfo)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedQ, q)
			}
		})
	}
}
