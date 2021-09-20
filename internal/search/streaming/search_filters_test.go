package streaming

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestSearchFiltersUpdate(t *testing.T) {
	repo := types.RepoName{
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
							Repo: repo,
							Body: result.HighlightedString{Highlights: make([]result.HighlightedRange, 2)}},
						&result.CommitMatch{
							Repo: repo,
							Body: result.HighlightedString{Highlights: make([]result.HighlightedRange, 1)}},
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
							LineMatches: []*result.LineMatch{{
								OffsetAndLengths: make([][2]int32, 2),
							}},
						},
					},
				},
			},
			wantFilterName:  "repo:^foo$",
			wantFilterKind:  "repo",
			wantFilterCount: 2,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			s := &SearchFilters{}
			for _, event := range c.events {
				s.Update(event)
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
