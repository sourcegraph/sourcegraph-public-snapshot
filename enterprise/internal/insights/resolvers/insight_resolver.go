package resolvers

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
)

var _ graphqlbackend.InsightResolver = &insightResolver{}

type insightResolver struct {
	store   *store.Store
	insight graphqlbackend.InsightResolver
}

func (r *insightResolver) Title() string { return r.insight.Title() }

func (r *insightResolver) Description() string { return r.insight.Description() }

func (r *insightResolver) Series() []graphqlbackend.InsightSeriesResolver {
	// TODO: locate time series from r.store DB.
	return nil
}
