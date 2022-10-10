package pipeline

import (
	"context"
	"fmt"
	"testing"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func makeJobGenerator(numJobs int) SearchJobGenerator {
	return func(ctx context.Context, req BackfillRequest) <-chan SearchJobGeneratorOutput {
		output := make(chan SearchJobGeneratorOutput)
		goroutine.Go(func() {
			defer close(output)
			for i := 0; i < numJobs; i++ {
				output <- SearchJobGeneratorOutput{
					BackfillRequest: &req,
					Job: &queryrunner.Job{
						SeriesID:    req.Series.SeriesID,
						SearchQuery: fmt.Sprintf("%d", i),
					},
				}
			}
		})

		return output
	}
}

func searchRunner(ctx context.Context, input <-chan SearchJobGeneratorOutput) <-chan SearchResultOutput {
	output := make(chan SearchResultOutput)
	goroutine.Go(func() {
		defer close(output)
		for job := range input {
			output <- SearchResultOutput{
				result:          searchResult{count: 10, capture: "", repo: job.Repo},
				BackfillRequest: job.BackfillRequest,
				err:             nil}
		}
	})

	return output
}

type runCounts struct {
	err         error
	resultCount int
	totalCount  int
}

func TestBackfillStepsConnected(t *testing.T) {

	testCases := []struct {
		numJobs int
		want    autogold.Value
	}{
		{10, autogold.Want("With Jobs", runCounts{resultCount: 10, totalCount: 100})},
		{0, autogold.Want("No Jobs", runCounts{})},
	}

	for _, tc := range testCases {
		got := runCounts{}
		countingPersister := func(ctx context.Context, input <-chan SearchResultOutput) error {
			for r := range input {
				got.resultCount++
				got.totalCount += r.result.count
			}
			return nil
		}

		backfiller := NewBackfiller(makeJobGenerator(tc.numJobs), searchRunner, countingPersister)
		got.err = backfiller.Run(context.Background(), BackfillRequest{Series: &types.InsightSeries{SeriesID: "1"}})
		tc.want.Equal(t, got)
	}

}
