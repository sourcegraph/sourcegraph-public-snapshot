package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

var _ graphqlbackend.InsightViewResolver = &insightViewResolver{}
var _ graphqlbackend.LineChartInsightViewPresentation = &lineChartInsightViewPresentation{}

type insightViewResolver struct {
	view *types.Insight
}

func (i *insightViewResolver) ID() graphql.ID {
	return relay.MarshalID("insight_view", i.view.UniqueID)
}

func (i *insightViewResolver) DefaultFilters(ctx context.Context) (graphqlbackend.InsightViewFiltersResolver, error) {
	panic("implement me")
}

func (i *insightViewResolver) AppliedFilters(ctx context.Context) (graphqlbackend.InsightViewFiltersResolver, error) {
	panic("implement me")
}

func (i *insightViewResolver) DataSeries(ctx context.Context) ([]graphqlbackend.InsightSeriesResolver, error) {
	panic("implement me")
}

func (i *insightViewResolver) Presentation(ctx context.Context) (graphqlbackend.LineChartInsightViewPresentation, error) {
	return &lineChartInsightViewPresentation{view: i.view}, nil
}

func (i *insightViewResolver) DataSeriesDefinitions(ctx context.Context) ([]graphqlbackend.SearchInsightDataSeriesDefinitionResolver, error) {
	panic("implement me")
}

func (r *Resolver) CreateLineChartSearchInsight(ctx context.Context, args *graphqlbackend.CreateLineChartSearchInsightArgs) (graphqlbackend.CreateInsightResultResolver, error) {
	panic("implement me")
}

type lineChartInsightViewPresentation struct {
	view *types.Insight
}

func (l *lineChartInsightViewPresentation) Title(ctx context.Context) (string, error) {
	return l.view.Title, nil
}

func (l *lineChartInsightViewPresentation) SeriesPresentation(ctx context.Context) ([]graphqlbackend.LineChartDataSeriesPresentationResolver, error) {
	panic("implement me")
}
