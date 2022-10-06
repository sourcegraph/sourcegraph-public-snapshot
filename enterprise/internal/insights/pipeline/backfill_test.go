package pipeline

import (
	"context"
	"fmt"
	"testing"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
)

// type SearchJobGenerator func(ctx context.Context, req requestContext) (context.Context, *requestContext, []*queryrunner.Job, error)
// type SearchRunner func(ctx context.Context, reqContext *requestContext, jobs []*queryrunner.Job, err error) (context.Context, *requestContext, []store.RecordSeriesPointArgs, error)
// type ResultsPersister func(ctx context.Context, reqContext *requestContext, points []store.RecordSeriesPointArgs, err error) (*requestContext, error)

func makeJobGenerator(numJobs int) SearchJobGenerator {
	return func(ctx context.Context, req requestContext) (context.Context, *requestContext, []*queryrunner.Job, error) {
		jobs := make([]*queryrunner.Job, 0, numJobs)
		for i := 0; i < numJobs; i++ {
			jobs = append(jobs, &queryrunner.Job{
				SeriesID:    req.backfillRequest.Series.SeriesID,
				SearchQuery: fmt.Sprintf("%d", i),
			})
		}
		return ctx, &req, jobs, nil
	}
}

func searchRunner(ctx context.Context, reqContext *requestContext, jobs []*queryrunner.Job, err error) (context.Context, *requestContext, []store.RecordSeriesPointArgs, error) {
	points := make([]store.RecordSeriesPointArgs, 0, len(jobs))
	for _, _ = range jobs {
		points = append(points, store.RecordSeriesPointArgs{Point: store.SeriesPoint{Value: 10}})
	}
	return ctx, reqContext, points, nil
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
		countingPersister := func(ctx context.Context, reqContext *requestContext, points []store.RecordSeriesPointArgs, err error) (*requestContext, error) {
			for _, p := range points {
				got.resultCount++
				got.totalCount += int(p.Point.Value)
			}
			return reqContext, nil
		}

		backfiller := NewBackfiller(makeJobGenerator(tc.numJobs), searchRunner, countingPersister)
		got.err = backfiller.Run(context.Background(), BackfillRequest{Series: &types.InsightSeries{SeriesID: "1"}})
		tc.want.Equal(t, got)
	}
}
