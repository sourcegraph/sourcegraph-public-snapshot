package streaming

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestSearchFiltersUpdate(t *testing.T) {
	repo := types.MinimalRepo{
		Name: "foo",
	}

	cases := []struct {
		name            string
		events          []SearchEvent
		wantFilterName  string
		wantFilterCount int
		wantFilterKind  string
	}{
		{
			name: "CommitMatch",
			events: []SearchEvent{
				{
					Results: []result.Match{
						&result.CommitMatch{
							Repo:           repo,
							MessagePreview: &result.MatchedString{MatchedRanges: make([]result.Range, 2)}},
						&result.CommitMatch{
							Repo:           repo,
							MessagePreview: &result.MatchedString{MatchedRanges: make([]result.Range, 1)}},
					},
				}},
			wantFilterName:  "repo:^foo$",
			wantFilterKind:  "repo",
			wantFilterCount: 3,
		},
		{
			name: "RepoMatch",
			events: []SearchEvent{
				{
					Results: []result.Match{
						&result.RepoMatch{
							Name: "foo",
						},
					},
				},
			},
			wantFilterName:  "repo:^foo$",
			wantFilterKind:  "repo",
			wantFilterCount: 1,
		},
		{
			name: "FileMatch, repo: filter",
			events: []SearchEvent{
				{
					Results: []result.Match{
						&result.FileMatch{
							File: result.File{
								Repo: repo,
							},
							ChunkMatches: result.ChunkMatches{{Ranges: make(result.Ranges, 2)}},
						},
					},
				},
			},
			wantFilterName:  "repo:^foo$",
			wantFilterKind:  "repo",
			wantFilterCount: 2,
		},
	}

	// add mocks for the file metrics caching that is now involved in search filters
	// avoids NPEs because dbmocks does not instantiate any mock functions
	mockDB := dbmocks.NewMockDB()
	fileMetricsStore := dbmocks.NewMockFileMetricsStore()
	fileMetricsStore.GetFileMetricsFunc.SetDefaultHook(func(ctx context.Context, ri api.RepoID, ci api.CommitID, s string) *fileutil.FileMetrics {
		// no caching in tests
		return nil
	})
	fileMetricsStore.SetFileMetricsFunc.SetDefaultHook(func(ctx context.Context, ri api.RepoID, ci api.CommitID, s string, fm *fileutil.FileMetrics, b bool) error {
		// no caching in tests
		return nil
	})
	mockDB.FileMetricsFunc.SetDefaultReturn(fileMetricsStore)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			s := &SearchFilters{}
			for _, event := range c.events {
				s.Update(event, mockDB)
			}

			f, ok := s.filters[c.wantFilterName]
			if !ok {
				t.Fatalf("expected %s", c.wantFilterName)
			}

			if f.Kind != c.wantFilterKind {
				t.Fatalf("want %s, got %s", c.wantFilterKind, f.Kind)
			}

			if f.Count != c.wantFilterCount {
				t.Fatalf("want %d, got %d", c.wantFilterCount, f.Count)
			}
		})
	}
}
