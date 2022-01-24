package resolvers

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

var _ graphqlbackend.CompareTwoInsightSeriesDataPointResolver = &compareTwoInsightSeriesDataPointResolver{}

type compareTwoInsightSeriesDataPointResolver struct {
	dataPoint compareTwoInsightSeriesDataPoint
}

func (c *compareTwoInsightSeriesDataPointResolver) FirstSeriesValue() *int32 {
	return c.dataPoint.FirstSeriesValue
}

func (c *compareTwoInsightSeriesDataPointResolver) SecondSeriesValue() *int32 {
	return c.dataPoint.SecondSeriesValue
}

func (c *compareTwoInsightSeriesDataPointResolver) Diff() int32 {
	return c.dataPoint.Diff
}

func (c *compareTwoInsightSeriesDataPointResolver) RepoName() string {
	return c.dataPoint.RepoName
}

func (c *compareTwoInsightSeriesDataPointResolver) Time() graphqlbackend.DateTime {
	return c.dataPoint.Time
}

type compareTwoInsightSeriesDataPoint struct {
	FirstSeriesValue  *int32
	SecondSeriesValue *int32
	Diff              int32
	RepoName          string
	Time              graphqlbackend.DateTime
}

func (r *Resolver) CompareTwoInsightSeries(ctx context.Context, args graphqlbackend.CompareTwoInsightSeriesArgs) ([]graphqlbackend.CompareTwoInsightSeriesDataPointResolver, error) {
	var resolvers []graphqlbackend.CompareTwoInsightSeriesDataPointResolver

	resolvers = append(resolvers, &compareTwoInsightSeriesDataPointResolver{dataPoint: compareTwoInsightSeriesDataPoint{
		FirstSeriesValue:  nilInt(5),
		SecondSeriesValue: nilInt(8),
		Diff:              3,
		RepoName:          "github.com/sourcegraph/sourcegraph",
		Time:              graphqlbackend.DateTime{Time: time.Now()},
	}})

	resolvers = append(resolvers, &compareTwoInsightSeriesDataPointResolver{dataPoint: compareTwoInsightSeriesDataPoint{
		FirstSeriesValue:  nilInt(2),
		SecondSeriesValue: nilInt(2),
		Diff:              0,
		RepoName:          "github.com/sourcegraph/handbook",
		Time:              graphqlbackend.DateTime{Time: time.Now()},
	}})
	return resolvers, nil
}

func nilInt(val int32) *int32 {
	return &val
}
