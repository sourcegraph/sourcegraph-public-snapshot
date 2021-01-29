package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/schema"
)

var _ graphqlbackend.InsightSeriesResolver = &insightSeriesResolver{}

type insightSeriesResolver struct {
	store  *store.Store
	series *schema.InsightSeries
}

func (r *insightSeriesResolver) Label() string { return r.series.Label }

func (r *insightSeriesResolver) Points(ctx context.Context, args *graphqlbackend.InsightsPointsArgs) ([]graphqlbackend.InsightsDataPointResolver, error) {
	var opts store.SeriesPointsOpts
	opts.SeriesID = nil // FUTURE: TODO: set opts.SeriesID to effective hash of r.series
	if args.From != nil {
		opts.From = &args.From.Time
	}
	if args.To != nil {
		opts.To = &args.To.Time
	}
	// FUTURE: Pass through opts.Limit

	points, err := r.store.SeriesPoints(ctx, opts)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.InsightsDataPointResolver, 0, len(points))
	for _, point := range points {
		resolvers = append(resolvers, insightsDataPointResolver{point})
	}
	return resolvers, nil
}

var _ graphqlbackend.InsightsDataPointResolver = insightsDataPointResolver{}

type insightsDataPointResolver struct{ p store.SeriesPoint }

func (i insightsDataPointResolver) DateTime() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: i.p.Time}
}

func (i insightsDataPointResolver) Value() float64 { return i.p.Value }
