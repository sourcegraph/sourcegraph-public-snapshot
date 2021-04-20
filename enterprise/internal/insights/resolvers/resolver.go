package resolvers

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

var _ graphqlbackend.InsightsResolver = &Resolver{}

// Resolver is the GraphQL resolver of all things related to Insights.
type Resolver struct {
	insightsStore   store.Interface
	workerBaseStore *basestore.Store
	settingStore    *database.SettingStore
}

// New returns a new Resolver whose store uses the given Timescale and Postgres DBs.
func New(timescale, postgres dbutil.DB) graphqlbackend.InsightsResolver {
	return newWithClock(timescale, postgres, timeutil.Now)
}

// newWithClock returns a new Resolver whose store uses the given Timescale and Postgres DBs, and the given
// clock for timestamps.
func newWithClock(timescale, postgres dbutil.DB, clock func() time.Time) *Resolver {
	return &Resolver{
		insightsStore:   store.NewWithClock(timescale, clock),
		workerBaseStore: basestore.NewWithDB(postgres, sql.TxOptions{}),
		settingStore:    database.Settings(postgres),
	}
}

func (r *Resolver) Insights(ctx context.Context) (graphqlbackend.InsightConnectionResolver, error) {
	return &insightConnectionResolver{
		insightsStore:   r.insightsStore,
		workerBaseStore: r.workerBaseStore,
		settingStore:    r.settingStore,
	}, nil
}

type disabledResolver struct{}

func NewDisabledResolver() graphqlbackend.InsightsResolver {
	return &disabledResolver{}
}

var ErrCodeInsightsDisabled = errors.New("code insights has been disabled")

func (r *disabledResolver) Insights(ctx context.Context) (graphqlbackend.InsightConnectionResolver, error) {
	return nil, ErrCodeInsightsDisabled
}
