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
	// TODO: locate time series from r.store DB.
	return nil
}
