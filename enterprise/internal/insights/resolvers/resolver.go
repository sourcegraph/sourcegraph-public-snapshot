package resolvers

import (
	"context"
	"database/sql"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

var _ graphqlbackend.InsightsResolver = &Resolver{}

// Resolver is the GraphQL resolver of all things related to Insights.
type Resolver struct {
	timeSeriesStore      store.Interface
	workerBaseStore      *basestore.Store
	insightMetadataStore store.InsightMetadataStore
	dataSeriesStore      store.DataSeriesStore
	dashboardStore       *store.DBDashboardStore
	insightsDatabase     dbutil.DB
	postgresDatabase     dbutil.DB
}

// New returns a new Resolver whose store uses the given Timescale and Postgres DBs.
func New(timescale, postgres dbutil.DB) graphqlbackend.InsightsResolver {
	return newWithClock(timescale, postgres, timeutil.Now)
}

// newWithClock returns a new Resolver whose store uses the given Timescale and Postgres DBs, and the given
// clock for timestamps.
func newWithClock(timescale, postgres dbutil.DB, clock func() time.Time) *Resolver {
	insightStore := store.NewInsightStore(timescale)
	return &Resolver{
		timeSeriesStore:      store.NewWithClock(timescale, store.NewInsightPermissionStore(postgres), clock),
		workerBaseStore:      basestore.NewWithDB(postgres, sql.TxOptions{}),
		insightMetadataStore: insightStore,
		dataSeriesStore:      insightStore,
		dashboardStore:       store.NewDashboardStore(timescale),
		insightsDatabase:     timescale,
		postgresDatabase:     postgres,
	}
}

func (r *Resolver) Insights(ctx context.Context, args *graphqlbackend.InsightsArgs) (graphqlbackend.InsightConnectionResolver, error) {
	var idList []string
	if args != nil && args.Ids != nil {
		idList = make([]string, len(*args.Ids))
		for i, id := range *args.Ids {
			idList[i] = string(id)
		}
	}
	return &insightConnectionResolver{
		insightsStore:        r.timeSeriesStore,
		workerBaseStore:      r.workerBaseStore,
		insightMetadataStore: r.insightMetadataStore,
		ids:                  idList,
		orgStore:             database.Orgs(r.workerBaseStore.Handle().DB()),
	}, nil
}

func (r *Resolver) InsightsDashboards(ctx context.Context, args *graphqlbackend.InsightsDashboardsArgs) (graphqlbackend.InsightsDashboardConnectionResolver, error) {
	return &dashboardConnectionResolver{
		dashboardStore: store.NewDashboardStore(r.insightsDatabase),
		orgStore:       database.Orgs(r.workerBaseStore.Handle().DB()),
		insightStore:   store.NewInsightStore(r.insightsDatabase),
		args:           args,
	}, nil
}

// 🚨 SECURITY
// only add users / orgs if the user is non-anonymous. This will restrict anonymous users to only see
// dashboards with a global grant.
func getUserPermissions(ctx context.Context, orgStore *database.OrgStore) (userIds []int, orgIds []int, err error) {
	userId := actor.FromContext(ctx).UID
	if userId != 0 {
		var orgs []*types.Org
		orgs, err = orgStore.GetByUserID(ctx, userId)
		if err != nil {
			return
		}
		userIds = []int{int(userId)}
		orgIds = make([]int, 0, len(orgs))
		for _, org := range orgs {
			orgIds = append(orgIds, int(org.ID))
		}
	}
	return
}
