package jobutil

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/regexp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func Test_descriptionMatchRanges(t *testing.T) {
	repoIDsToDescriptions := map[api.RepoID]string{
		1: "this is a go package",
		2: "description for tests and validating input, among other things",
		3: "description containing go but also\na newline",
		4: "---zb bz zb bz---",
		5: "this description has unicode 🙈 characters",
	}

	// NOTE: Any pattern passed into repo:has.description() is converted to the format `(?:*).*?(?:*)` when the predicate
	// is parsed (see `internal/search/query/types.RepoHasDescription()`). The resulting value(s) are then compiled into
	// regex during job construction.
	tests := []struct {
		name          string
		inputPatterns []*regexp.Regexp
		want          map[api.RepoID][]result.Range
	}{
		{
			name:          "string literal match",
			inputPatterns: []*regexp.Regexp{regexp.MustCompile(`(?is)(?:go).*?(?:package)`)},
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
			inputPatterns: []*regexp.Regexp{regexp.MustCompile(`(?is)(?:test).*?(?:input)`)},
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
			inputPatterns: []*regexp.Regexp{regexp.MustCompile(`(?is)(?:containing).*?(?:newline)`)},
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
			inputPatterns: []*regexp.Regexp{regexp.MustCompile(`(?is)(?:this).*?(?:matches).*?(?:nothing)`)},
			want:          map[api.RepoID][]result.Range{},
		},
		{
			name:          "matches same pattern multiple times",
			inputPatterns: []*regexp.Regexp{regexp.MustCompile(`(?is)(?:zb)`)},
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
		{
			name:          "counts unicode characters correctly",
			inputPatterns: []*regexp.Regexp{regexp.MustCompile(`(?is)(?:unicode).*?(?:🙈).*?(?:char)`)},
			want: map[api.RepoID][]result.Range{
				5: {
					result.Range{
						Start: result.Location{
							Offset: 21,
							Line:   0,
							Column: 21,
						},
						End: result.Location{
							Offset: 38,
							Line:   0,
							Column: 35,
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			job := &RepoSearchJob{
				DescriptionPatterns: tc.inputPatterns,
			}

			got := job.descriptionMatchRanges(repoIDsToDescriptions)

			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("unexpected results (-want +got)\n%s", diff)
			}
		})
	}
}

func TestRepoMatchRanges(t *testing.T) {
	repoNameRegexps := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:misery).*?(?:business)`),
		regexp.MustCompile(`(?i)(?:brick).*?(?:by).*?(?:boring).*?(?:brick)`),
	}

	tests := []struct {
		name  string
		input string
		want  []result.Range
	}{
		{
			name:  "single repo name match range",
			input: "2007/riot/misery-business",
			want: []result.Range{
				{
					Start: result.Location{
						Offset: 10,
						Line:   0,
						Column: 10,
					},
					End: result.Location{
						Offset: 25,
						Line:   0,
						Column: 25,
					},
				},
			},
		},
		{
			name:  "multiple match ranges",
			input: "greatest-hits/miseryBusiness/crushcrushcrush/brickByBoringBrick",
			want: []result.Range{
				{
					Start: result.Location{
						Offset: 14,
						Line:   0,
						Column: 14,
					},
					End: result.Location{
						Offset: 28,
						Line:   0,
						Column: 28,
					},
				},
				{
					Start: result.Location{
						Offset: 45,
						Line:   0,
						Column: 45,
					},
					End: result.Location{
						Offset: 63,
						Line:   0,
						Column: 63,
					},
				},
			},
		},
		{
			name:  "strips code host from repo name before matching",
			input: "github.com/paramore/riot/tracklist/4-misery-business/5-when-it-rains",
			want: []result.Range{
				{
					Start: result.Location{
						Offset: 26,
						Line:   0,
						Column: 26,
					},
					End: result.Location{
						Offset: 41,
						Line:   0,
						Column: 41,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := repoMatchRanges(tc.input, repoNameRegexps)
			require.Equal(t, tc.want, got)
		})
	}
}
