package streaming

import (
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/require"
)

func TestSearchFiltersUpdate(t *testing.T) {
	repo := types.MinimalRepo{
		Name: "foo",
	}

	cases := []struct {
		name            string
		events          []SearchEvent
		wantFilterValue string
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
			wantFilterValue: "repo:^foo$",
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
			wantFilterValue: "repo:^foo$",
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
			wantFilterValue: "repo:^foo$",
			wantFilterKind:  "repo",
			wantFilterCount: 2,
		},
		{
			name: "SymbolMatch",
			events: []SearchEvent{
				{
					Results: []result.Match{
						&result.FileMatch{
							Symbols: []*result.SymbolMatch{
								{
									Symbol: result.Symbol{
										Kind: "class",
									},
								},
								{
									Symbol: result.Symbol{
										Kind: "class",
									},
								},
								{
									Symbol: result.Symbol{
										Kind: "class",
									},
								},
							},
						},
					},
				},
			},
			wantFilterValue: "select:symbol.class",
			wantFilterKind:  "symbol type",
			wantFilterCount: 3,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			s := &SearchFilters{}
			for _, event := range c.events {
				s.Update(event)
			}

			f, ok := s.filters[c.wantFilterValue]
			if !ok {
				t.Fatalf("expected %s", c.wantFilterValue)
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

func TestSymbolCounts(t *testing.T) {
	cases := []struct {
		name        string
		events      []SearchEvent
		wantFilters map[string]*Filter
	}{
		{
			name: "return different counts for different symbol types",
			events: []SearchEvent{
				{
					Results: []result.Match{
						&result.FileMatch{
							Symbols: []*result.SymbolMatch{
								{
									Symbol: result.Symbol{
										Kind: "class",
									},
								},
								{
									Symbol: result.Symbol{
										Kind: "variable",
									},
								},
								{
									Symbol: result.Symbol{
										Kind: "variable",
									},
								},
								{
									Symbol: result.Symbol{
										Kind: "constant",
									},
								},
								{
									Symbol: result.Symbol{
										Kind: "constant",
									},
								},
								{
									Symbol: result.Symbol{
										Kind: "constant",
									},
								},
								{
									Symbol: result.Symbol{
										Kind: "constant",
									},
								},
							},
						},
					},
				},
			},
			wantFilters: map[string]*Filter{
				"select:symbol.class": &Filter{
					Value:      "select:symbol.class",
					Label:      "class",
					Count:      1,
					IsLimitHit: false,
					Kind:       "symbol type",
				},
				"select:symbol.variable": &Filter{
					Value:      "select:symbol.variable",
					Label:      "variable",
					Count:      2,
					IsLimitHit: false,
					Kind:       "symbol type",
				},
				"select:symbol.constant": &Filter{
					Value:      "select:symbol.constant",
					Label:      "constant",
					Count:      4,
					IsLimitHit: false,
					Kind:       "symbol type",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &SearchFilters{}
			for _, event := range tc.events {
				s.Update(event)
			}

			for key, filter := range tc.wantFilters {
				require.Equal(t, filter, s.filters[key])
			}
		})
	}
}

// This should be used to generate a large enough set to test IsLimitHit property.
func generateLargeResultSet(symbolKind string) []SearchEvent {
	symbolMatches := []*result.SymbolMatch{}
	for i := 0; i <= 500; i++ {
		symbolMatches = append(symbolMatches,
			&result.SymbolMatch{
				Symbol: result.Symbol{
					Kind: symbolKind,
				},
			},
		)
	}

	return []SearchEvent{
		{
			Results: result.Matches{
				&result.FileMatch{
					Symbols: symbolMatches,
				},
			},
		},
	}
}

func TestDetermineTimeframe(t *testing.T) {
	now := time.Now()
	cases := []struct {
		date time.Time
		want DateFilterInfo
	}{
		{now.Add(-13 * time.Hour), DateFilterInfo{AFTER, TODAY, "today"}},
		{now.Add(-24 * time.Hour), DateFilterInfo{AFTER, TODAY, "today"}},
		{now.Add(-2 * 24 * time.Hour), DateFilterInfo{AFTER, ONE_WEEK_AGO, "this week"}},
		{now.Add(-7 * 24 * time.Hour), DateFilterInfo{AFTER, ONE_WEEK_AGO, "this week"}},
		{now.Add(-10 * 24 * time.Hour), DateFilterInfo{AFTER, TWO_WEEKS_AGO, "since last week"}},
		{now.Add(-14 * 24 * time.Hour), DateFilterInfo{AFTER, TWO_WEEKS_AGO, "since last week"}},
		{now.Add(-27 * 24 * time.Hour), DateFilterInfo{AFTER, ONE_MONTH_AGO, "this month"}},
		{now.Add(-30 * 24 * time.Hour), DateFilterInfo{AFTER, ONE_MONTH_AGO, "this month"}},
		{now.Add(-44 * 24 * time.Hour), DateFilterInfo{AFTER, TWO_MONTHS_AGO, "since two months ago"}},
		{now.Add(-60 * 24 * time.Hour), DateFilterInfo{AFTER, TWO_MONTHS_AGO, "since two months ago"}},
		{now.Add(-72 * 24 * time.Hour), DateFilterInfo{AFTER, THREE_MONTHS_AGO, "since three months ago"}},
		{now.Add(-90 * 24 * time.Hour), DateFilterInfo{AFTER, THREE_MONTHS_AGO, "since three months ago"}},
		{now.Add(-288 * 24 * time.Hour), DateFilterInfo{AFTER, ONE_YEAR_AGO, "since one year ago"}},
		{now.Add(-365 * 24 * time.Hour), DateFilterInfo{AFTER, ONE_YEAR_AGO, "since one year ago"}},
		{now.Add(-400 * 24 * time.Hour), DateFilterInfo{BEFORE, ONE_YEAR_AGO, "before one year ago"}},
	}

	for _, tc := range cases {
		got := determineTimeframe(tc.date)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("determineTimeframe(%v) = %v, want %v", tc.date, got, tc.want)
		}
	}
}
