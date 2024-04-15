package codycontext

import (
	"context"
	"errors"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/mockjob"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"

	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	codeJob := mockjob.NewMockJob()
	textJob := mockjob.NewMockJob()

	// Create the job
	searchJob := &searchJob{
		codeCount: 4,
		textCount: 2,
		codeJob:   codeJob,
		textJob:   textJob,
	}

	{
		// Test single error
		codeJob.RunFunc.SetDefaultReturn(nil, errors.New("code job failed"))
		alert, err := searchJob.Run(context.Background(), job.RuntimeClients{}, streaming.NewNullStream())
		require.Nil(t, alert)
		require.NotNil(t, err)
	}

	{
		// Test both jobs error
		codeJob.RunFunc.SetDefaultReturn(nil, errors.New("code job failed"))
		textJob.RunFunc.SetDefaultReturn(nil, errors.New("text job failed"))
		_, err := searchJob.Run(context.Background(), job.RuntimeClients{}, streaming.NewNullStream())
		require.NotNil(t, err)
	}

	{
		// Test we select max priority alert
		codeJob.RunFunc.SetDefaultReturn(&search.Alert{Priority: 1}, nil)
		textJob.RunFunc.SetDefaultReturn(&search.Alert{Priority: 2}, nil)
		alert, _ := searchJob.Run(context.Background(), job.RuntimeClients{}, streaming.NewNullStream())
		require.NotNil(t, alert)
		require.Equal(t, 2, alert.Priority)
	}

	{
		// Test that results are limited and combined as expected
		codeMatches := result.Matches{
			&result.FileMatch{File: result.File{Path: "file1.go"}},
			&result.FileMatch{File: result.File{Path: "file2.go"}},
			&result.FileMatch{File: result.File{Path: "file3.go"}},
			&result.FileMatch{File: result.File{Path: "file4.go"}},
			&result.FileMatch{File: result.File{Path: "file5.go"}},
		}

		textMatches := result.Matches{
			&result.FileMatch{File: result.File{Path: "file1.md"}},
			&result.FileMatch{File: result.File{Path: "file2.md"}},
			&result.FileMatch{File: result.File{Path: "file3.md"}},
			&result.FileMatch{File: result.File{Path: "file4.md"}},
			&result.FileMatch{File: result.File{Path: "file5.md"}},
		}

		codeJob.RunFunc.SetDefaultHook(
			func(ctx context.Context, clients job.RuntimeClients, sender streaming.Sender) (*search.Alert, error) {
				sender.Send(streaming.SearchEvent{Results: codeMatches})
				return &search.Alert{}, nil
			})
		textJob.RunFunc.SetDefaultHook(
			func(ctx context.Context, clients job.RuntimeClients, sender streaming.Sender) (*search.Alert, error) {
				sender.Send(streaming.SearchEvent{Results: textMatches})
				return nil, nil
			})

		stream := streaming.NewAggregatingStream()
		alert, err := searchJob.Run(context.Background(), job.RuntimeClients{}, stream)
		require.NotNil(t, alert)
		require.Nil(t, err)

		require.Equal(t, append(codeMatches[:4], textMatches[:2]...), stream.Results)
	}
}
