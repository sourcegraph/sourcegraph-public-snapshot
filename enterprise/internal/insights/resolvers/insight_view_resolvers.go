package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

var _ graphqlbackend.InsightViewResolver = &insightViewResolver{}

// type stubInsightViewResolver struct {
// 	id string
// }
//
// func (s *stubInsightViewResolver) ID() graphql.ID {
// 	return relay.MarshalID("insight_view", s.id)
// }
//
// func (s *stubInsightViewResolver) VeryUniqueResolver() bool {
// 	return true
// }

type insightViewResolver struct {
}

func (i *insightViewResolver) ID() graphql.ID {
	panic("implement me")
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
	panic("implement me")
}

func (i *insightViewResolver) DataSeriesDefinitions(ctx context.Context) ([]graphqlbackend.SearchInsightDataSeriesDefinitionResolver, error) {
	panic("implement me")
}

func (r *Resolver) CreateLineChartSearchInsight(ctx context.Context, args *graphqlbackend.CreateLineChartSearchInsightArgs) (graphqlbackend.CreateInsightResultResolver, error) {
	panic("implement me")
}
