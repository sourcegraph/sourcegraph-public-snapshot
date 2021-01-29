package resolvers

import (
	"context"
	"errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

var _ graphqlbackend.InsightSeriesResolver = &insightSeriesResolver{}

type insightSeriesResolver struct {
	label string
}

func (r *insightSeriesResolver) Label() string { return r.label }

func (r *insightSeriesResolver) Points(ctx context.Context, args *graphqlbackend.InsightsPointsArgs) ([]graphqlbackend.InsightsDataPointResolver, error) {
	return nil, errors.New("not yet implemented")
}
