package jobutil

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func Test_descriptionMatchRanges(t *testing.T) {
	repoIDsToDescriptions := map[api.RepoID]string{
		1: "this is a go package",
		2: "description for tests and validating input, among other things",
		3: "description containing go but also\na newline",
		4: "---zb bz zb bz---",
	}

	// NOTE: Any pattern passed into repo:has.description() is converted to the format `(?:*).*?(?:*)` when the predicate
	// is parsed (see `internal/search/query/types.RepoHasDescription()`). The resulting value(s) are then used to populate
	// the DescriptionPatterns field on RepoOptions. As of now, descriptionMatchRanges() should never receive a []string
	// argument containing a pattern that has not already been formatted in this manner.
	// If you're wondering why all the test cases have inputPatterns' elements formatted this way, that's the reason.
	tests := []struct {
		name          string
		inputPatterns []string
		want          map[api.RepoID][]result.Range
	}{
		{
			name:          "string literal match",
			inputPatterns: []string{"(?:go).*?(?:package)"},
			want: map[api.RepoID][]result.Range{
				1: {
					result.Range{
						Start: result.Location{
							Offset: 10,
							Line:   0,
							Column: 10,
						},
						End: result.Location{
							Offset: 20,
							Line:   0,
							Column: 20,
						},
					},
				},
			},
		},
		{
			name:          "wildcard match",
			inputPatterns: []string{"(?:test).*?(?:input)"},
			want: map[api.RepoID][]result.Range{
				2: {
					result.Range{
						Start: result.Location{
							Offset: 16,
							Line:   0,
							Column: 16,
						},
						End: result.Location{
							Offset: 42,
							Line:   0,
							Column: 42,
						},
					},
				},
			},
		},
		{
			name:          "match across newline",
			inputPatterns: []string{"(?:containing).*?(?:newline)"},
			want: map[api.RepoID][]result.Range{
				3: {
					result.Range{
						Start: result.Location{
							Offset: 12,
							Line:   0,
							Column: 12,
						},
						End: result.Location{
							Offset: 44,
							Line:   0,
							Column: 44,
						},
					},
				},
			},
		},
		{
			name:          "no matches",
			inputPatterns: []string{"(?:this).*?(?:matches).*?(?:nothing)"},
			want:          map[api.RepoID][]result.Range{},
		},
		{
			name:          "matches same pattern multiple times",
			inputPatterns: []string{"(?:zb)"},
			want: map[api.RepoID][]result.Range{
				4: {
					result.Range{
						Start: result.Location{
							Offset: 3,
							Line:   0,
							Column: 3,
						},
						End: result.Location{
							Offset: 5,
							Line:   0,
							Column: 5,
						},
					},
					result.Range{
						Start: result.Location{
							Offset: 9,
							Line:   0,
							Column: 9,
						},
						End: result.Location{
							Offset: 11,
							Line:   0,
							Column: 11,
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := descriptionMatchRanges(repoIDsToDescriptions, tc.inputPatterns)

			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("unexpected results (-want +got)\n%s", diff)
			}
		})
	}
}
