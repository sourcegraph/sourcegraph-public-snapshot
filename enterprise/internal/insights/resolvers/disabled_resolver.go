package resolvers

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type disabledResolver struct {
	reason string
}

func NewDisabledResolver(reason string) graphqlbackend.InsightsResolver {
	return &disabledResolver{reason}
}

func (r *disabledResolver) Insights(ctx context.Context, args *graphqlbackend.InsightsArgs) (graphqlbackend.InsightConnectionResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) InsightsDashboards(ctx context.Context, args *graphqlbackend.InsightsDashboardsArgs) (graphqlbackend.InsightsDashboardConnectionResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) CreateInsightsDashboard(ctx context.Context, args *graphqlbackend.CreateInsightsDashboardArgs) (graphqlbackend.InsightsDashboardPayloadResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) UpdateInsightsDashboard(ctx context.Context, args *graphqlbackend.UpdateInsightsDashboardArgs) (graphqlbackend.InsightsDashboardPayloadResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) DeleteInsightsDashboard(ctx context.Context, args *graphqlbackend.DeleteInsightsDashboardArgs) (*graphqlbackend.EmptyResponse, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) AddInsightViewToDashboard(ctx context.Context, args *graphqlbackend.AddInsightViewToDashboardArgs) (graphqlbackend.InsightsDashboardPayloadResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) RemoveInsightViewFromDashboard(ctx context.Context, args *graphqlbackend.RemoveInsightViewFromDashboardArgs) (graphqlbackend.InsightsDashboardPayloadResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) UpdateInsightSeries(ctx context.Context, args *graphqlbackend.UpdateInsightSeriesArgs) (graphqlbackend.InsightSeriesMetadataPayloadResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) InsightSeriesQueryStatus(ctx context.Context) ([]graphqlbackend.InsightSeriesQueryStatusResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) CreateLineChartSearchInsight(ctx context.Context, args *graphqlbackend.CreateLineChartSearchInsightArgs) (graphqlbackend.CreateInsightResultResolver, error) {
	return nil, errors.New(r.reason)
}
