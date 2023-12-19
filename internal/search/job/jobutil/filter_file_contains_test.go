package jobutil

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/mockjob"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	proto "github.com/sourcegraph/sourcegraph/internal/searcher/v1"
)

func TestFileContainsFilterJob(t *testing.T) {
	cm := func(matchedStrings ...string) result.ChunkMatch {
		ranges := make([]result.Range, 0, len(matchedStrings))
		currOffset := 0
		for _, matchedString := range matchedStrings {
			ranges = append(ranges, result.Range{
				Start: result.Location{Offset: currOffset},
				End:   result.Location{Offset: currOffset + len(matchedString)},
			})
			currOffset += len(matchedString)
		}
		return result.ChunkMatch{
			Content: strings.Join(matchedStrings, ""),
			Ranges:  ranges,
		}
	}
	fm := func(cms ...result.ChunkMatch) *result.FileMatch {
		if len(cms) == 0 {
			cms = result.ChunkMatches{}
		}
		return &result.FileMatch{
			ChunkMatches: cms,
		}
	}
	r := func(ms ...result.Match) (res result.Matches) {
		for _, m := range ms {
			res = append(res, m)
		}
		return res
	}
	cases := []struct {
		name            string
		includePatterns []string
		originalPattern query.Node
		caseSensitive   bool
		inputEvent      streaming.SearchEvent
		outputEvent     streaming.SearchEvent
	}{{
		name:            "no matches in streamed event",
		includePatterns: []string{"unused"},
		originalPattern: nil,
		caseSensitive:   false,
		inputEvent: streaming.SearchEvent{
			Stats: streaming.Stats{
				IsLimitHit: true,
			},
		},
		outputEvent: streaming.SearchEvent{
			Stats: streaming.Stats{
				IsLimitHit: true,
			},
		},
	}, {
		name:            "no original pattern",
		includePatterns: []string{"needle"},
		originalPattern: nil,
		caseSensitive:   false,
		inputEvent: streaming.SearchEvent{
			Results: r(fm(cm("needle"))),
		},
		outputEvent: streaming.SearchEvent{
			Results: r(fm()),
		},
	}, {
		name:            "overlapping original pattern",
		includePatterns: []string{"needle"},
		originalPattern: query.Pattern{Value: "needle"},
		caseSensitive:   false,
		inputEvent: streaming.SearchEvent{
			Results: r(fm(cm("needle"))),
		},
		outputEvent: streaming.SearchEvent{
			Results: r(fm(cm("needle"))),
		},
	}, {
		name:            "nonoverlapping original pattern",
		includePatterns: []string{"needle"},
		originalPattern: query.Pattern{Value: "pin"},
		caseSensitive:   false,
		inputEvent: streaming.SearchEvent{
			Results: r(fm(cm("needle", "pin"))),
		},
		outputEvent: streaming.SearchEvent{
			Results: r(fm(result.ChunkMatch{
				Content: "needlepin",
				Ranges: result.Ranges{{
					Start: result.Location{Offset: len("needle")},
					End:   result.Location{Offset: len("needlepin")},
				}},
			})),
		},
	}, {
		name:            "multiple include patterns",
		includePatterns: []string{"minimum", "viable", "product"},
		originalPattern: query.Pattern{Value: "predicates"},
		caseSensitive:   false,
		inputEvent: streaming.SearchEvent{
			Results: r(fm(cm("minimum", "viable"), cm("predicates", "product"))),
		},
		outputEvent: streaming.SearchEvent{
			Results: r(fm(result.ChunkMatch{
				Content: "predicatesproduct",
				Ranges: result.Ranges{{
					Start: result.Location{Offset: 0},
					End:   result.Location{Offset: len("predicates")},
				}},
			})),
		},
	}, {
		name:            "match that is not a file match",
		includePatterns: []string{"minimum", "viable", "product"},
		originalPattern: query.Pattern{Value: "predicates"},
		caseSensitive:   false,
		inputEvent: streaming.SearchEvent{
			Results: result.Matches{&result.RepoMatch{Name: "test"}},
		},
		outputEvent: streaming.SearchEvent{
			Results: result.Matches{},
		},
	}, {
		name:            "tree shaped pattern",
		includePatterns: []string{"predicate"},
		originalPattern: query.Operator{
			Kind: query.Or,
			Operands: []query.Node{
				query.Pattern{Value: "outer"},
				query.Operator{
					Kind: query.And,
					Operands: []query.Node{
						query.Pattern{Value: "inner1"},
						query.Pattern{Value: "inner2"},
					},
				},
			},
		},
		caseSensitive: false,
		inputEvent: streaming.SearchEvent{
			Results: r(fm(cm("inner1", "inner2", "predicate", "outer"))),
		},
		outputEvent: streaming.SearchEvent{
			Results: r(fm(result.ChunkMatch{
				Content: "inner1inner2predicateouter",
				Ranges: result.Ranges{{
					Start: result.Location{Offset: 0},
					End:   result.Location{Offset: len("inner1")},
				}, {
					Start: result.Location{Offset: len("inner1")},
					End:   result.Location{Offset: len("inner1inner2")},
				}, {
					Start: result.Location{Offset: len("inner1inner2predicate")},
					End:   result.Location{Offset: len("inner1inner2predicateouter")},
				}},
			})),
		},
	}, {
		name:            "diff search",
		includePatterns: []string{"predicate"},
		originalPattern: query.Pattern{Value: "needle"},
		caseSensitive:   false,
		inputEvent: streaming.SearchEvent{
			Results: r(&result.CommitMatch{
				DiffPreview: &result.MatchedString{
					Content: "file1 file2\n@@ -1,2 +1,6 @@\n+needle\n-needle\nfile3 file4\n@@ -3,4 +1,6 @@\n+needle\n-needle\n",
					MatchedRanges: result.Ranges{{
						Start: result.Location{Offset: 29, Line: 2, Column: 1},
						End:   result.Location{Offset: 35, Line: 2, Column: 7},
					}, {
						Start: result.Location{Offset: 37, Line: 3, Column: 1},
						End:   result.Location{Offset: 43, Line: 3, Column: 7},
					}, {
						Start: result.Location{Offset: 73, Line: 6, Column: 1},
						End:   result.Location{Offset: 79, Line: 6, Column: 7},
					}, {
						Start: result.Location{Offset: 81, Line: 7, Column: 1},
						End:   result.Location{Offset: 87, Line: 7, Column: 7},
					}},
				},
				Diff: []result.DiffFile{{
					OrigName: "file1",
					NewName:  "file2",
					Hunks: []result.Hunk{{
						OldStart: 1,
						NewStart: 1,
						OldCount: 2,
						NewCount: 6,
						Header:   "",
						Lines:    []string{"+needle", "-needle"},
					}},
				}, {
					OrigName: "file3",
					NewName:  "file4",
					Hunks: []result.Hunk{{
						OldStart: 3,
						NewStart: 1,
						OldCount: 4,
						NewCount: 6,
						Header:   "",
						Lines:    []string{"+needle", "-needle"},
					}},
				}},
			}),
		},
		outputEvent: streaming.SearchEvent{
			Results: r(&result.CommitMatch{
				DiffPreview: &result.MatchedString{
					Content: "file3 file4\n@@ -3,4 +1,6 @@\n+needle\n-needle\n",
					MatchedRanges: result.Ranges{{
						Start: result.Location{Offset: 29, Line: 2, Column: 1},
						End:   result.Location{Offset: 35, Line: 2, Column: 7},
					}, {
						Start: result.Location{Offset: 37, Line: 3, Column: 1},
						End:   result.Location{Offset: 43, Line: 3, Column: 7},
					}},
				},
				Diff: []result.DiffFile{{
					OrigName: "file3",
					NewName:  "file4",
					Hunks: []result.Hunk{{
						OldStart: 3,
						NewStart: 1,
						OldCount: 4,
						NewCount: 6,
						Header:   "",
						Lines:    []string{"+needle", "-needle"},
					}},
				}},
			}),
		},
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			childJob := mockjob.NewMockJob()
			childJob.RunFunc.SetDefaultHook(func(_ context.Context, _ job.RuntimeClients, s streaming.Sender) (*search.Alert, error) {
				s.Send(tc.inputEvent)
				return nil, nil
			})
			searcher.MockSearch = func(_ context.Context, _ api.RepoName, _ api.RepoID, _ api.CommitID, p *search.TextPatternInfo, _ time.Duration, onMatch func(*proto.FileMatch)) (limitHit bool, err error) {
				if len(p.IncludePatterns) > 0 {
					onMatch(&proto.FileMatch{Path: []byte("file4")})
				}
				return false, nil
			}
			var resultEvent streaming.SearchEvent
			streamCollector := streaming.StreamFunc(func(ev streaming.SearchEvent) {
				resultEvent = ev
			})
			j, err := NewFileContainsFilterJob(tc.includePatterns, tc.originalPattern, tc.caseSensitive, childJob)
			require.NoError(t, err)
			alert, err := j.Run(context.Background(), job.RuntimeClients{}, streamCollector)
			require.Nil(t, alert)
			require.NoError(t, err)
			require.Equal(t, tc.outputEvent, resultEvent)
		})
	}
}
