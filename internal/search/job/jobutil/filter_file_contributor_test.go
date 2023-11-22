package jobutil

import (
	"context"
	"testing"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/mockjob"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/stretchr/testify/require"
)

func TestFileHasContributorsJob(t *testing.T) {
	r := func(ms ...result.Match) (res result.Matches) {
		for _, m := range ms {
			res = append(res, m)
		}
		return res
	}

	fm := func() *result.FileMatch {
		return &result.FileMatch{
			File: result.File{
				Path:     "path",
				CommitID: "commitID",
			},
		}
	}

	ccs := func(nameAndEmails ...[]string) (contributorCounts []*gitdomain.ContributorCount) {
		cc := gitdomain.ContributorCount{}
		for _, nae := range nameAndEmails {
			if len(nae) == 2 {
				cc.Name = nae[0]
				cc.Email = nae[1]
			}
			contributorCounts = append(contributorCounts, &cc)
		}
		return contributorCounts
	}

	tests := []struct {
		name          string
		caseSensitive bool
		include       []string
		exclude       []string
		matches       result.Match
		contributors  []*gitdomain.ContributorCount
		outputEvent   streaming.SearchEvent
	}{{
		name:         "include matches name",
		include:      []string{"Author"},
		matches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mail.com"}, []string{"author", "author@mail.com"}),
		outputEvent:  streaming.SearchEvent{Results: r(fm())},
	}, {
		name:         "include matches email",
		include:      []string{"Author@mail.com"},
		matches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mail.com"}, []string{"author", "author@mail.com"}),
		outputEvent:  streaming.SearchEvent{Results: r(fm())},
	}, {
		name:         "include has no matches",
		include:      []string{"Author"},
		matches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mail.com"}),
		outputEvent:  streaming.SearchEvent{Results: result.Matches{}},
	}, {
		name:         "exclude matches name",
		exclude:      []string{"Author"},
		matches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mail.com"}, []string{"author", "author@mail.com"}),
		outputEvent:  streaming.SearchEvent{Results: result.Matches{}},
	}, {
		name:         "exclude matches email",
		exclude:      []string{"Author@mail.com"},
		matches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mail.com"}, []string{"author", "author@mail.com"}),
		outputEvent:  streaming.SearchEvent{Results: result.Matches{}},
	}, {
		name:         "exclude has no matches",
		exclude:      []string{"Author"},
		matches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mail.com"}),
		outputEvent:  streaming.SearchEvent{Results: r(fm())},
	}, {
		name:         "exclude and include each match",
		include:      []string{"contributor"},
		exclude:      []string{"Author"},
		matches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mail.com"}, []string{"author", "author@mail.com"}),
		outputEvent:  streaming.SearchEvent{Results: result.Matches{}},
	}, {
		name:         "not every include matches",
		include:      []string{"contributor", "author"},
		matches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mail.com"}, []string{"editor", "editor@mail.com"}),
		outputEvent:  streaming.SearchEvent{Results: result.Matches{}},
	}, {
		name:         "not every exclude matches",
		exclude:      []string{"contributor", "author"},
		matches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mail.com"}, []string{"editor", "editor@mail.com"}),
		outputEvent:  streaming.SearchEvent{Results: r(fm())},
	}, {
		name:         "include regex matches",
		include:      []string{"Au.hor@mai.*"},
		matches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mail.com"}, []string{"author", "author@mail.com"}),
		outputEvent:  streaming.SearchEvent{Results: r(fm())},
	}, {
		name:         "exclude regex matches",
		exclude:      []string{"Au.hor@mai.*"},
		matches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mail.com"}, []string{"author", "author@mail.com"}),
		outputEvent:  streaming.SearchEvent{Results: result.Matches{}},
	}, {
		name:          "include case sensitive has matches",
		include:       []string{"Author"},
		caseSensitive: true,
		matches:       fm(),
		contributors:  ccs([]string{"Author", "author@mail.com"}),
		outputEvent:   streaming.SearchEvent{Results: r(fm())},
	}, {
		name:          "include case sensitive has no matches",
		include:       []string{"Author"},
		caseSensitive: true,
		matches:       fm(),
		contributors:  ccs([]string{"author", "author@mail.com"}),
		outputEvent:   streaming.SearchEvent{Results: result.Matches{}},
	}, {
		name:          "exclude case sensitive has matches",
		exclude:       []string{"Author"},
		caseSensitive: true,
		matches:       fm(),
		contributors:  ccs([]string{"Author", "author@mail.com"}),
		outputEvent:   streaming.SearchEvent{Results: result.Matches{}},
	}, {
		name:          "exclude case sensitive has no matches",
		exclude:       []string{"Author"},
		caseSensitive: true,
		matches:       fm(),
		contributors:  ccs([]string{"author", "author@mail.com"}),
		outputEvent:   streaming.SearchEvent{Results: r(fm())},
	}, {
		name:         "empty include and empty exclude always returns",
		matches:      fm(),
		contributors: ccs([]string{"author", "author@mail.com"}),
		outputEvent:  streaming.SearchEvent{Results: r(fm())},
	}, {
		name:        "not all matches are files",
		include:     []string{"Author"},
		matches:     &result.CommitMatch{},
		outputEvent: streaming.SearchEvent{Results: result.Matches{}},
	}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			childJob := mockjob.NewMockJob()
			childJob.RunFunc.SetDefaultHook(func(_ context.Context, _ job.RuntimeClients, s streaming.Sender) (*search.Alert, error) {
				s.Send(streaming.SearchEvent{Results: r(tc.matches)})
				return nil, nil
			})

			gitServerClient := gitserver.NewMockClient()
			gitServerClient.ContributorCountFunc.PushReturn(tc.contributors, nil)

			var resultEvent streaming.SearchEvent
			streamCollector := streaming.StreamFunc(func(ev streaming.SearchEvent) {
				resultEvent = ev
			})

			includeRegexp := toRe(tc.include, tc.caseSensitive)
			excludeRegexp := toRe(tc.exclude, tc.caseSensitive)
			j := NewFileHasContributorsJob(childJob, includeRegexp, excludeRegexp)
			alert, err := j.Run(context.Background(), job.RuntimeClients{Gitserver: gitServerClient}, streamCollector)
			require.Nil(t, alert)
			require.NoError(t, err)
			require.Equal(t, tc.outputEvent, resultEvent)
		})
	}
}

func toRe(contributors []string, isCaseSensitive bool) (res []*regexp.Regexp) {
	for _, pattern := range contributors {
		if isCaseSensitive {
			res = append(res, regexp.MustCompile(pattern))
		} else {
			res = append(res, regexp.MustCompile(`(?i)`+pattern))
		}
	}
	return res
}
