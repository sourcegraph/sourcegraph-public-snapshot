package resolvers

import (
	"context"
	"errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// Resolver is the GraphQL resolver of all things related to Insights.
type Resolver struct{}

// New returns a new Resolver whose store uses the given database
func New() graphqlbackend.InsightsResolver {
	return &Resolver{}
}

func (r *Resolver) Insights(ctx context.Context) (graphqlbackend.InsightsResolver, error) {
	return r, nil
}

func (r *Resolver) Points(ctx context.Context, args *graphqlbackend.InsightsPointsArgs) ([]graphqlbackend.InsightsDataPointResolver, error) {
	return nil, errors.New("not yet implemented")
}
