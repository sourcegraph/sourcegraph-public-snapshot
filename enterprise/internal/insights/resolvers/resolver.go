package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

var _ graphqlbackend.InsightsResolver = &Resolver{}

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

func (r *Resolver) Insights(ctx context.Context) (graphqlbackend.InsightConnectionResolver, error) {
	return &insightConnectionResolver{
		store:        r.store,
		settingStore: r.settingStore,
	}, nil
}
