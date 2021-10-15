package resolvers

import (
	"context"
	"database/sql"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/actor"

	"github.com/sourcegraph/sourcegraph/internal/database"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

var _ graphqlbackend.InsightsResolver = &Resolver{}
var _ graphqlbackend.InsightSeriesMetadataPayloadResolver = &insightSeriesMetadataPayloadResolver{}
var _ graphqlbackend.InsightSeriesMetadataResolver = &insightSeriesMetadataResolver{}

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
		insightsDatabase: r.insightsDatabase,
		dashboardStore:   store.NewDashboardStore(r.insightsDatabase),
		args:             args,
	}, nil
}

type insightSeriesMetadataPayloadResolver struct {
	series *types.InsightSeries
}

func (i *insightSeriesMetadataPayloadResolver) Series(ctx context.Context) graphqlbackend.InsightSeriesMetadataResolver {
	return &insightSeriesMetadataResolver{series: i.series}
}

type insightSeriesMetadataResolver struct {
	series *types.InsightSeries
}

func (i *insightSeriesMetadataResolver) SeriesId(ctx context.Context) (string, error) {
	return i.series.SeriesID, nil
}

func (i *insightSeriesMetadataResolver) Query(ctx context.Context) (string, error) {
	return i.series.Query, nil
}

func (i *insightSeriesMetadataResolver) Enabled(ctx context.Context) (bool, error) {
	return i.series.Enabled, nil
}

func (r *Resolver) UpdateInsightSeries(ctx context.Context, args *graphqlbackend.UpdateInsightSeriesArgs) (graphqlbackend.InsightSeriesMetadataPayloadResolver, error) {
	actr := actor.FromContext(ctx)
	if err := backend.CheckUserIsSiteAdmin(ctx, r.postgresDatabase, actr.UID); err != nil {
		return nil, err
	}

	if args.Input.Enabled != nil {
		err := r.dataSeriesStore.SetSeriesEnabled(ctx, args.Input.SeriesId, *args.Input.Enabled)
		if err != nil {
			return nil, err
		}
	}

	series, err := r.dataSeriesStore.GetDataSeries(ctx, store.GetDataSeriesArgs{IncludeDeleted: true, SeriesID: args.Input.SeriesId})
	if err != nil {
		return nil, err
	}
	if len(series) == 0 {
		return nil, errors.Newf("unable to fetch series with series_id: %v", args.Input.SeriesId)
	}
	return &insightSeriesMetadataPayloadResolver{series: &series[0]}, nil
}

type disabledResolver struct {
	reason string
}

func NewDisabledResolver(reason string) graphqlbackend.InsightsResolver {
	return &disabledResolver{reason}
}

func (r *disabledResolver) Insights(ctx context.Context, args *graphqlbackend.InsightsArgs) (graphqlbackend.InsightConnectionResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) InsightsDashboards(ctx context.Context, args *graphqlbackend.InsightsDashboardsArgs) (graphqlbackend.InsightsDashboardConnectionResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) CreateInsightsDashboard(ctx context.Context, args *graphqlbackend.CreateInsightsDashboardArgs) (graphqlbackend.InsightsDashboardPayloadResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) UpdateInsightsDashboard(ctx context.Context, args *graphqlbackend.UpdateInsightsDashboardArgs) (graphqlbackend.InsightsDashboardPayloadResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) DeleteInsightsDashboard(ctx context.Context, args *graphqlbackend.DeleteInsightsDashboardArgs) (*graphqlbackend.EmptyResponse, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) AddInsightViewToDashboard(ctx context.Context, args *graphqlbackend.AddInsightViewToDashboardArgs) (graphqlbackend.InsightsDashboardPayloadResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) RemoveInsightViewFromDashboard(ctx context.Context, args *graphqlbackend.RemoveInsightViewFromDashboardArgs) (graphqlbackend.InsightsDashboardPayloadResolver, error) {
	return nil, errors.New(r.reason)
}

func (r *disabledResolver) UpdateInsightSeries(ctx context.Context, args *graphqlbackend.UpdateInsightSeriesArgs) (graphqlbackend.InsightSeriesMetadataPayloadResolver, error) {
	return nil, errors.New(r.reason)
}
