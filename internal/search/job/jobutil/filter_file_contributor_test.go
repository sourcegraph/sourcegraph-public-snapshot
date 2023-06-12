package jobutil

import (
	"context"
	"testing"

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

	ccs := func(nameAndEmail ...[]string) (contributorCounts []*gitdomain.ContributorCount) {
		cc := gitdomain.ContributorCount{}
		for _, nae := range nameAndEmail {
			if len(nameAndEmail) == 2 {
				cc.Name = nae[0]
				cc.Email = nae[1]
			}
			contributorCounts = append(contributorCounts, &cc)
		}
		return contributorCounts
	}

	tests := []struct {
		name                string
		ignoreCase          bool
		includeContributors []string
		excludeContributors []string
		matches             result.Match
		contributors        []*gitdomain.ContributorCount
		outputEvent         streaming.SearchEvent
	}{{
		name:                "include contributor matches name",
		ignoreCase:          true,
		includeContributors: []string{"Author"},
		excludeContributors: []string{},
		matches:             fm(),
		contributors:        ccs([]string{"Author", ""}),
		outputEvent:         streaming.SearchEvent{Results: result.Matches{}},
	}, {
		name:                "include contributor matches email",
		ignoreCase:          true,
		includeContributors: []string{"Author"},
		excludeContributors: []string{},
		matches:             fm(),
		contributors:        ccs([]string{"", "Author"}),
		outputEvent:         streaming.SearchEvent{Results: result.Matches{}},
	}, {
		name:                "no matches are files",
		ignoreCase:          true,
		includeContributors: []string{},
		excludeContributors: []string{},
		matches:             &result.CommitMatch{},
		contributors:        ccs(),
		outputEvent:         streaming.SearchEvent{Results: result.Matches{}},
	}, {
		name:                "not all matches are files",
		ignoreCase:          true,
		includeContributors: []string{},
		excludeContributors: []string{},
		matches:             fm(),
		outputEvent: streaming.SearchEvent{
			Results: r(&result.FileMatch{
				File: result.File{
					Path: "path",
				},
			}),
		},
	}, {
		name:                "not all matches are files",
		ignoreCase:          true,
		includeContributors: []string{},
		excludeContributors: []string{},
		matches:             fm(),
		outputEvent: streaming.SearchEvent{
			Results: r(&result.FileMatch{
				File: result.File{
					Path: "path",
				},
			}),
		},
	}}
	for _, tc := range tests {
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

		j := NewFileHasContributorsJob(childJob, tc.ignoreCase, tc.includeContributors, tc.excludeContributors)
		alert, err := j.Run(context.Background(), job.RuntimeClients{Gitserver: gitServerClient}, streamCollector)
		require.Nil(t, alert)
		require.NoError(t, err)
		require.Equal(t, tc.outputEvent, resultEvent)
	}
}
