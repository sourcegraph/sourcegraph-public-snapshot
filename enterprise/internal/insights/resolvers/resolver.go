package resolvers

import (
	"context"
	"errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
)

// Resolver is the GraphQL resolver of all things related to Insights.
type Resolver struct {
	store *store.Store
}

// New returns a new Resolver whose store uses the given db
func New(db dbutil.DB) graphqlbackend.InsightsResolver {
	return &Resolver{store: store.New(db)}
}

func (r *Resolver) Insights(ctx context.Context) (graphqlbackend.InsightsResolver, error) {
	return r, nil
}

func (r *Resolver) Points(ctx context.Context, args *graphqlbackend.InsightsPointsArgs) ([]graphqlbackend.InsightsDataPointResolver, error) {
	return nil, errors.New("not yet implemented")
}
