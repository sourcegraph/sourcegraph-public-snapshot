package jobutil

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/mockjob"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
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
	r := func(fms ...*result.FileMatch) (res result.Matches) {
		for _, fm := range fms {
			res = append(res, fm)
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
			Results: result.Matches{&result.RepoMatch{Name: "test"}},
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
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			childJob := mockjob.NewMockJob()
			childJob.RunFunc.SetDefaultHook(func(_ context.Context, _ job.RuntimeClients, s streaming.Sender) (*search.Alert, error) {
				s.Send(tc.inputEvent)
				return nil, nil
			})
			var result streaming.SearchEvent
			streamCollector := streaming.StreamFunc(func(ev streaming.SearchEvent) {
				result = ev
			})
			j := NewFileContainsFilterJob(tc.includePatterns, tc.originalPattern, tc.caseSensitive, childJob)
			alert, err := j.Run(context.Background(), job.RuntimeClients{}, streamCollector)
			require.Nil(t, alert)
			require.NoError(t, err)
			require.Equal(t, tc.outputEvent, result)
		})
	}
}
