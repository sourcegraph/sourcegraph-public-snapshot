package resolvers

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/schema"
)

var _ graphqlbackend.InsightResolver = &insightResolver{}

type insightResolver struct {
	store   *store.Store
	insight *schema.Insight
}

func (r *insightResolver) Title() string { return r.insight.Title }

func (r *insightResolver) Description() string { return r.insight.Description }

func (r *insightResolver) Series() []graphqlbackend.InsightSeriesResolver {
	series := r.insight.Series
	resolvers := make([]graphqlbackend.InsightSeriesResolver, 0, len(series))
	for _, series := range series {
		resolvers = append(resolvers, &insightSeriesResolver{store: r.store, series: series})
	}
	return resolvers
}
