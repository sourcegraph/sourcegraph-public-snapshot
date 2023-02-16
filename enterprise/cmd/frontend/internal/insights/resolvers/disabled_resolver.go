package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type disabledResolver struct {
	reason string
}

func NewDisabledResolver(reason string) graphqlbackend.InsightsResolver {
	return &disabledResolver{reason}
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

func (r *disabledResolver) CreateLineChartSearchInsight(ctx context.Context, args *graphqlbackend.CreateLineChartSearchInsightArgs) (graphqlbackend.InsightViewPayloadResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) UpdateLineChartSearchInsight(ctx context.Context, args *graphqlbackend.UpdateLineChartSearchInsightArgs) (graphqlbackend.InsightViewPayloadResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) CreatePieChartSearchInsight(ctx context.Context, args *graphqlbackend.CreatePieChartSearchInsightArgs) (graphqlbackend.InsightViewPayloadResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) UpdatePieChartSearchInsight(ctx context.Context, args *graphqlbackend.UpdatePieChartSearchInsightArgs) (graphqlbackend.InsightViewPayloadResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) InsightViews(ctx context.Context, args *graphqlbackend.InsightViewQueryArgs) (graphqlbackend.InsightViewConnectionResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) DeleteInsightView(ctx context.Context, args *graphqlbackend.DeleteInsightViewArgs) (*graphqlbackend.EmptyResponse, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) SearchInsightLivePreview(ctx context.Context, args graphqlbackend.SearchInsightLivePreviewArgs) ([]graphqlbackend.SearchInsightLivePreviewSeriesResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) SearchInsightPreview(ctx context.Context, args graphqlbackend.SearchInsightPreviewArgs) ([]graphqlbackend.SearchInsightLivePreviewSeriesResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) SearchQueryAggregate(ctx context.Context, args graphqlbackend.SearchQueryArgs) (graphqlbackend.SearchQueryAggregateResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) InsightViewDebug(ctx context.Context, args graphqlbackend.InsightViewDebugArgs) (graphqlbackend.InsightViewDebugResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) SaveInsightAsNewView(ctx context.Context, args graphqlbackend.SaveInsightAsNewViewArgs) (graphqlbackend.InsightViewPayloadResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) ValidateScopedInsightQuery(ctx context.Context, args graphqlbackend.ValidateScopedInsightQueryArgs) (graphqlbackend.ScopedInsightQueryPayloadResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) PreviewRepositoriesFromQuery(ctx context.Context, args graphqlbackend.PreviewRepositoriesFromQueryArgs) (graphqlbackend.RepositoryPreviewPayloadResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) InsightAdminBackfillQueue(ctx context.Context, args *graphqlbackend.AdminBackfillQueueArgs) (*graphqlutil.ConnectionResolver[*graphqlbackend.BackfillQueueItemResolver], error) {
	return nil, errors.New(r.reason)
}
func (r *disabledResolver) RetryInsightSeriesBackfill(ctx context.Context, args *graphqlbackend.BackfillArgs) (*graphqlbackend.BackfillQueueItemResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) MoveInsightSeriesBackfillToFrontOfQueue(ctx context.Context, args *graphqlbackend.BackfillArgs) (*graphqlbackend.BackfillQueueItemResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) MoveInsightSeriesBackfillToBackOfQueue(ctx context.Context, args *graphqlbackend.BackfillArgs) (*graphqlbackend.BackfillQueueItemResolver, error) {
	return nil, errors.New(r.reason)
}
