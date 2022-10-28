package resolvers

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query"
)

var _ graphqlbackend.InsightSeriesResolver = &dynamicInsightSeriesResolver{}
var _ graphqlbackend.InsightStatusResolver = &emptyInsightStatusResolver{}

// dynamicInsightSeriesResolver is a series resolver that expands based on matches from a search query.
type dynamicInsightSeriesResolver struct {
	generated *query.GeneratedTimeSeries
}

func (d *dynamicInsightSeriesResolver) SeriesId() string {
	return d.generated.SeriesId
}

func (d *dynamicInsightSeriesResolver) Label() string {
	return d.generated.Label
}

func (d *dynamicInsightSeriesResolver) Points(ctx context.Context, _ *graphqlbackend.InsightsPointsArgs) ([]graphqlbackend.InsightsDataPointResolver, error) {
	var resolvers []graphqlbackend.InsightsDataPointResolver
	for _, point := range d.generated.Points {
		resolvers = append(resolvers, &insightsDataPointResolver{store.SeriesPoint{
			SeriesID: d.generated.SeriesId,
			Time:     point.Time,
			Value:    float64(point.Count),
		}})
	}
	return resolvers, nil
}

func (d *dynamicInsightSeriesResolver) Status(ctx context.Context) (graphqlbackend.InsightStatusResolver, error) {
	return &emptyInsightStatusResolver{}, nil
}

func (d *dynamicInsightSeriesResolver) DirtyMetadata(ctx context.Context) ([]graphqlbackend.InsightDirtyQueryResolver, error) {
	return nil, nil
}

type emptyInsightStatusResolver struct{}

func (e emptyInsightStatusResolver) TotalPoints(ctx context.Context) (int32, error) {
	return 0, nil
}

func (e emptyInsightStatusResolver) PendingJobs(ctx context.Context) (int32, error) {
	return 0, nil
}

func (e emptyInsightStatusResolver) CompletedJobs(ctx context.Context) (int32, error) {
	return 0, nil
}

func (e emptyInsightStatusResolver) FailedJobs(ctx context.Context) (int32, error) {
	return 0, nil
}

func (e emptyInsightStatusResolver) IsLoadingData(ctx context.Context) (*bool, error) {
	// beacuse this resolver is created when dynamic data exists
	// it means it's not loading data.
	loading := false
	return &loading, nil
}

func (e emptyInsightStatusResolver) BackfillQueuedAt(ctx context.Context) *gqlutil.DateTime {
	current := time.Now().AddDate(-1, 0, 0)
	return gqlutil.DateTimeOrNil(&current)
}
