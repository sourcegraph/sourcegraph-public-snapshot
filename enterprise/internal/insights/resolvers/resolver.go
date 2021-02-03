package resolvers

import (
	"context"
	"errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

// Resolver is the GraphQL resolver of all things related to Insights.
type Resolver struct {
	Store        *store.Store
	settingStore *database.SettingStore
}

// New returns a new Resolver whose store uses the given Timescale and Postgres DBs.
func New(timescale, postgres dbutil.DB) *Resolver {
	return &Resolver{
		Store:        store.New(timescale),
		settingStore: database.Settings(postgres),
	}
}

func (r *Resolver) Insights(ctx context.Context) (graphqlbackend.InsightsResolver, error) {
	return r, nil
}

func (r *Resolver) Points(ctx context.Context, args *graphqlbackend.InsightsPointsArgs) ([]graphqlbackend.InsightsDataPointResolver, error) {
	return nil, errors.New("not yet implemented")
}
