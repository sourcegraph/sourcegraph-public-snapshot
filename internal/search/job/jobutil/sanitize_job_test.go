package jobutil

import (
	"context"
	"strings"
	"testing"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/mockjob"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/stretchr/testify/require"
)

func TestSanitizeJob(t *testing.T) {
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
	cdm := func(matchedStrings ...string) *result.CommitMatch {
		ranges := make([]result.Range, 0, len(matchedStrings))
		currOffset := 0
		for _, matchedString := range matchedStrings {
			ranges = append(ranges, result.Range{
				Start: result.Location{Offset: currOffset},
				End:   result.Location{Offset: currOffset + len(matchedString)},
			})
			currOffset += len(matchedString)
		}
		return &result.CommitMatch{
			DiffPreview: &result.MatchedString{
				Content:       strings.Join(matchedStrings, ""),
				MatchedRanges: ranges,
			},
		}
	}
	r := func(ms ...result.Match) (res result.Matches) {
		for _, m := range ms {
			res = append(res, m)
		}
		return res
	}

	omitPatterns := []*regexp.Regexp{
		regexp.MustCompile("omitme[a-zA-z]{5}"),
		regexp.MustCompile("^pattern1$"),
		regexp.MustCompile("(?im)Pattern2[a-zA-Z]{3}"),
	}

	tests := []struct {
		name        string
		inputEvent  streaming.SearchEvent
		outputEvent streaming.SearchEvent
	}{
		{
			name: "no sanitize patterns apply",
			inputEvent: streaming.SearchEvent{
				Results: r(fm(cm("nothing to sanitize"))),
			},
			outputEvent: streaming.SearchEvent{
				Results: r(fm(cm("nothing to sanitize"))),
			},
		},
		{
			name: "sanitize chunk match",
			inputEvent: streaming.SearchEvent{
				Results: r(fm(cm("omitmeABcDe"), cm("don't omit me"))),
			},
			outputEvent: streaming.SearchEvent{
				Results: r(fm(cm("don't omit me"))),
			},
		},
		{
			name: "sanitize range within a chunk match",
			inputEvent: streaming.SearchEvent{
				Results: r(fm(cm("pattern1", " some other text"))),
			},
			outputEvent: streaming.SearchEvent{
				Results: r(fm(result.ChunkMatch{
					Content: "pattern1 some other text",
					Ranges: result.Ranges{
						{Start: result.Location{Offset: len("pattern1")}, End: result.Location{Offset: len("pattern1 some other text")}},
					},
				})),
			},
		},
		{
			name: "sanitize commit diff match",
			inputEvent: streaming.SearchEvent{
				Results: r(cdm("patTErn2ABC"), cdm("good diff")),
			},
			outputEvent: streaming.SearchEvent{
				Results: r(cdm("good diff")),
			},
		},
		{
			name: "no-op for commit match that is not a diff match",
			inputEvent: streaming.SearchEvent{
				Results: r(&result.CommitMatch{
					MessagePreview: &result.MatchedString{
						Content: "commit msg",
						MatchedRanges: []result.Range{
							{Start: result.Location{Offset: 0}, End: result.Location{Offset: len("commit")}},
						},
					},
				}),
			},
			outputEvent: streaming.SearchEvent{
				Results: r(&result.CommitMatch{
					MessagePreview: &result.MatchedString{
						Content: "commit msg",
						MatchedRanges: []result.Range{
							{Start: result.Location{Offset: 0}, End: result.Location{Offset: len("commit")}},
						},
					},
				}),
			},
		},
		{
			name: "no-op for result type other than FileMatch or CommitMatch",
			inputEvent: streaming.SearchEvent{
				Results: r(&result.RepoMatch{Name: "weird al greatest hits"}),
			},
			outputEvent: streaming.SearchEvent{
				Results: r(&result.RepoMatch{Name: "weird al greatest hits"}),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			childJob := mockjob.NewMockJob()
			childJob.RunFunc.SetDefaultHook(func(_ context.Context, _ job.RuntimeClients, s streaming.Sender) (*search.Alert, error) {
				s.Send(tc.inputEvent)
				return nil, nil
			})

			var searchEvent streaming.SearchEvent
			streamCollector := streaming.StreamFunc(func(event streaming.SearchEvent) {
				searchEvent = event
			})

			j := NewSanitizeJob(omitPatterns, childJob)
			alert, err := j.Run(context.Background(), job.RuntimeClients{}, streamCollector)
			require.Nil(t, alert)
			require.NoError(t, err)
			require.Equal(t, tc.outputEvent, searchEvent)
		})
	}
}
