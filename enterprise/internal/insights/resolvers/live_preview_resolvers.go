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
	} else if args.Input.TimeScope.StepInterval == nil {
		return nil, errors.New("live preview currently only supports a time interval time scope")
	}

	executor := query.NewCaptureGroupExecutor(r.postgresDB, r.insightsDB, time.Now)
	interval := timeseries.TimeInterval{
		Unit:  types.IntervalUnit(args.Input.TimeScope.StepInterval.Unit),
		Value: int(args.Input.TimeScope.StepInterval.Value),
	}
	generatedSeries, err := executor.Execute(ctx, args.Input.Query, args.Input.RepositoryScope.Repositories, interval)
	if err != nil {
		return nil, err
	}

	var resolvers []graphqlbackend.SearchInsightLivePreviewSeriesResolver
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
