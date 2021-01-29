package resolvers

import (
	"context"
	"errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

// Resolver is the GraphQL resolver of all things related to Insights.
type Resolver struct {
	store        *store.Store
	settingStore *database.SettingStore
}

// New returns a new Resolver whose store uses the given Timescale and Postgres DBs.
func New(timescale, postgres dbutil.DB) graphqlbackend.InsightsResolver {
	return &Resolver{
		store:        store.New(timescale),
		settingStore: database.Settings(postgres),
	}
}

func (r *Resolver) Insights(ctx context.Context) (graphqlbackend.InsightsResolver, error) {
	return r, nil
}

func (r *Resolver) Points(ctx context.Context, args *graphqlbackend.InsightsPointsArgs) ([]graphqlbackend.InsightsDataPointResolver, error) {
	return nil, errors.New("not yet implemented")
}
