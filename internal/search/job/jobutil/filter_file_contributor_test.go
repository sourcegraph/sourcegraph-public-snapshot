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
	type ContributorSet struct {
		contributors []*gitdomain.ContributorCount
	}
	r := func(ms ...result.Match) (res result.Matches) {
		for _, m := range ms {
			res = append(res, m)
		}
		return res
	}

	tests := []struct {
		name                string
		ignoreCase          bool
		includeContributors []string
		excludeContributors []string
		matches             []result.Match
		contributorSets        []ContributorSet
		outputEvent         streaming.SearchEvent
	}{{
		name:                "no matches are files",
		ignoreCase:          true,
		includeContributors: []string{},
		excludeContributors: []string{},
		matches: []result.Match{
			&result.CommitMatch{},
		}
		contributorSets: []ContributorSet{
			{contributors: []*gitdomain.ContributorCount{}},
		},
		outputEvent: streaming.SearchEvent{Results: result.Matches{}},
	}, {
		name:                "not all matches are files",
		ignoreCase:          true,
		includeContributors: []string{},
		excludeContributors: []string{},
		matches: []result.Match{
			&result.FileMatch{
				File: result.File{
					Path: "path",
				},
			},
			&result.CommitMatch{},
		},
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
		matches: []result.Match{
			&result.FileMatch{
				File: result.File{
					Path: "path",
				},
			},
			&result.CommitMatch{},
		},
		outputEvent: streaming.SearchEvent{
			Results: r(&result.FileMatch{
				File: result.File{
					Path: "path",
				},
			}),
		},
	}}
	//}, {}}

	for _, tc := range tests {
		childJob := mockjob.NewMockJob()
		childJob.RunFunc.SetDefaultHook(func(_ context.Context, _ job.RuntimeClients, s streaming.Sender) (*search.Alert, error) {
			s.Send(streaming.SearchEvent{Results: tc.matches})
			return nil, nil
		})

		gitserverClient := gitserver.NewMockClient()
		for _, contributorsForMatch := range tc.contributors {
			gitserverClient.ContributorCountFunc.PushReturn(contributorsForMatch.contributors, nil)
		}

		var resultEvent streaming.SearchEvent
		streamCollector := streaming.StreamFunc(func(ev streaming.SearchEvent) {
			resultEvent = ev
		})

		j := NewFileHasContributorsJob(childJob, tc.ignoreCase, tc.includeContributors, tc.excludeContributors)
		alert, err := j.Run(context.Background(), job.RuntimeClients{Gitserver: gitserverClient}, streamCollector)
		require.Nil(t, alert)
		require.NoError(t, err)
		require.Equal(t, tc.outputEvent, resultEvent)
	}
}
