package resolvers

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/timeseries"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *Resolver) SearchInsightLivePreview(ctx context.Context, args graphqlbackend.SearchInsightLivePreviewArgs) ([]graphqlbackend.SearchInsightLivePreviewSeriesResolver, error) {
	if !args.Input.GeneratedFromCaptureGroups {
		return nil, errors.New("live preview is currently only supported for generated series from capture groups")
	}
	previewArgs := graphqlbackend.SearchInsightPreviewArgs{
		Input: graphqlbackend.SearchInsightPreviewInput{
			RepositoryScope: args.Input.RepositoryScope,
			TimeScope:       args.Input.TimeScope,
			Series: []graphqlbackend.SearchSeriesPreviewInput{
				{
					Query:                      args.Input.Query,
					Label:                      args.Input.Label,
					GeneratedFromCaptureGroups: args.Input.GeneratedFromCaptureGroups,
				},
			},
		},
	}
	return r.SearchInsightPreview(ctx, previewArgs)
}

func (r *Resolver) SearchInsightPreview(ctx context.Context, args graphqlbackend.SearchInsightPreviewArgs) ([]graphqlbackend.SearchInsightLivePreviewSeriesResolver, error) {
	if args.Input.TimeScope.StepInterval == nil {
		return nil, errors.New("live preview currently only supports a time interval time scope")
	}
	var resolvers []graphqlbackend.SearchInsightLivePreviewSeriesResolver
	var generatedSeries []query.GeneratedTimeSeries

	// get a consistent time to use across all preview series
	previewTime := time.Now().UTC()
	clock := func() time.Time {
		return previewTime
	}
	interval := timeseries.TimeInterval{
		Unit:  types.IntervalUnit(args.Input.TimeScope.StepInterval.Unit),
		Value: int(args.Input.TimeScope.StepInterval.Value),
	}
	repos := args.Input.RepositoryScope.Repositories
	for _, seriesArgs := range args.Input.Series {

		var series []query.GeneratedTimeSeries
		var err error

		if seriesArgs.GeneratedFromCaptureGroups {
			executor := query.NewCaptureGroupExecutor(r.postgresDB, r.insightsDB, clock)
			series, err = executor.Execute(ctx, seriesArgs.Query, repos, interval)
			if err != nil {
				return nil, err
			}
		} else {
			executor := query.NewStreamingExecutor(r.postgresDB, r.insightsDB, clock)
			series, err = executor.Execute(ctx, seriesArgs.Query, seriesArgs.Label, seriesArgs.Label, repos, interval)
			if err != nil {
				return nil, err
			}
		}
		generatedSeries = append(generatedSeries, series...)
	}

	for i := range generatedSeries {
		resolvers = append(resolvers, &searchInsightLivePreviewSeriesResolver{series: &generatedSeries[i]})
	}

	return resolvers, nil
}

type searchInsightLivePreviewSeriesResolver struct {
	series *query.GeneratedTimeSeries
}

func (s *searchInsightLivePreviewSeriesResolver) Points(ctx context.Context) ([]graphqlbackend.InsightsDataPointResolver, error) {
	var resolvers []graphqlbackend.InsightsDataPointResolver
	for _, point := range s.series.Points {
		resolvers = append(resolvers, &insightsDataPointResolver{store.SeriesPoint{
			SeriesID: s.series.SeriesId,
			Time:     point.Time,
			Value:    float64(point.Count),
		}})
	}
	return resolvers, nil
}

func (s *searchInsightLivePreviewSeriesResolver) Label(ctx context.Context) (string, error) {
	return s.series.Label, nil
}
