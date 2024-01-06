package streaming

import (
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
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
							MessagePreview: &result.MatchedString{MatchedRanges: make([]result.Range, 2)},
							Commit: gitdomain.Commit{
								Committer: &gitdomain.Signature{
									Name:  "test committer",
									Email: "test@committer.com",
									Date:  time.Now(),
								},
							},
						},
						// We prefer Committer, but it could be nil
						// so we fallback to Author which cannot be nil.
						// Author also has a Date property, but it is
						// less accurrate.
						&result.CommitMatch{
							Repo:           repo,
							MessagePreview: &result.MatchedString{MatchedRanges: make([]result.Range, 1)},
							Commit: gitdomain.Commit{
								Committer: nil,
								Author: gitdomain.Signature{
									Name:  "test author",
									Email: "test@author.com",
									Date:  time.Now(),
								},
							},
						},
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
					Value: "select:symbol.class",
					Label: "class",
					Count: 1,
					Kind:  "symbol type",
				},
				"select:symbol.variable": &Filter{
					Value: "select:symbol.variable",
					Label: "variable",
					Count: 2,
					Kind:  "symbol type",
				},
				"select:symbol.constant": &Filter{
					Value: "select:symbol.constant",
					Label: "constant",
					Count: 4,
					Kind:  "symbol type",
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
		want dateFilterInfo
	}{
		{now.Add(-13 * time.Hour), dateFilterInfo{AFTER, YESTERDAY, "Last 24 hours"}},
		{now.Add(-24 * time.Hour), dateFilterInfo{AFTER, YESTERDAY, "Last 24 hours"}},
		{now.Add(-2 * 24 * time.Hour), dateFilterInfo{BEFORE, ONE_WEEK_AGO, "Last week"}},
		{now.Add(-7 * 24 * time.Hour), dateFilterInfo{BEFORE, ONE_WEEK_AGO, "Last week"}},
		{now.Add(-27 * 24 * time.Hour), dateFilterInfo{BEFORE, ONE_MONTH_AGO, "Last month"}},
		{now.Add(-30 * 24 * time.Hour), dateFilterInfo{BEFORE, ONE_MONTH_AGO, "Last month"}},
	}

	for _, tc := range cases {
		got := determineTimeframe(tc.date)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("determineTimeframe(%v) = %v, want %v", tc.date, got, tc.want)
		}
	}
}
