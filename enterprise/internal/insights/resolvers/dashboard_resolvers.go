package resolvers

import (
	"context"
	"strings"
	"sync"

	"github.com/inconshreveable/log15"

	"github.com/cockroachdb/errors"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

var _ graphqlbackend.InsightsDashboardConnectionResolver = &dashboardConnectionResolver{}
var _ graphqlbackend.InsightDashboardResolver = &insightDashboardResolver{}

type dashboardConnectionResolver struct {
	insightsDatabase dbutil.DB
	dashboardStore   store.DashboardStore
	args             *graphqlbackend.InsightDashboardsArgs

	// Cache results because they are used by multiple fields
	once       sync.Once
	dashboards []*types.Dashboard
	next       int64
	err        error
}

func (d *dashboardConnectionResolver) compute(ctx context.Context) ([]*types.Dashboard, int64, error) {
	d.once.Do(func() {
		args := store.DashboardQueryArgs{}
		if d.args.After != nil {
			afterID, err := unmarshal(graphql.ID(*d.args.After))
			if err != nil {
				d.err = errors.Wrap(err, "unmarshalID")
				return
			}
			log15.Debug("DashboardCompute", "afterID", afterID)
			args.After = int(afterID.Arg)
		}
		if d.args.First != nil {
			args.Limit = int(*d.args.First)
		}
		dashboards, err := d.dashboardStore.GetDashboards(ctx, args)
		if err != nil {
			d.err = err
			return
		}
		d.dashboards = dashboards
		for _, dashboard := range dashboards {
			log15.Info("dashboard iterate", "dbid", dashboard.ID, "dnext", d.next)
			if int64(dashboard.ID) > d.next {
				log15.Info("dnext replace", "dbid", dashboard.ID, "dnext", d.next)

				d.next = int64(dashboard.ID)
			}
		}
	})
	return d.dashboards, d.next, d.err
}

func (d *dashboardConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.InsightDashboardResolver, error) {
	dashboards, _, err := d.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.InsightDashboardResolver, 0, len(dashboards))
	for _, dashboard := range dashboards {
		id := newRealDashboardID(int64(dashboard.ID))
		resolvers = append(resolvers, &insightDashboardResolver{dashboard: dashboard, id: &id})
	}
	return resolvers, nil
}

func (d *dashboardConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, _, err := d.compute(ctx)
	if err != nil {
		return nil, err
	}
	if d.next != 0 {
		return graphqlutil.NextPageCursor(string(newRealDashboardID(d.next).marshal())), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

type insightDashboardResolver struct {
	dashboard *types.Dashboard
	id        *dashboardID
}

func (i *insightDashboardResolver) Title() string {
	return i.dashboard.Title
}

func (i *insightDashboardResolver) ID() graphql.ID {
	return i.id.marshal()
}

func (i *insightDashboardResolver) Views() graphqlbackend.InsightViewConnectionResolver {
	panic("implement me")
}

func newRealDashboardID(arg int64) dashboardID {
	return newDashboardID("real", arg)
}
func newDashboardID(idType string, arg int64) dashboardID {
	return dashboardID{
		IdType: idType,
		Arg:    arg,
	}
}

const dashboardKind = "dashboard"

type dashboardID struct {
	IdType string
	Arg    int64
}

func (id dashboardID) isVirtualized() bool {
	return id.isUser() || id.isOrg()
}

func (id dashboardID) isUser() bool {
	return strings.EqualFold(id.IdType, "user")
}

func (id dashboardID) isOrg() bool {
	return strings.EqualFold(id.IdType, "organization")
}

func (id dashboardID) isReal() bool {
	return strings.EqualFold(id.IdType, "custom")
}

func unmarshal(id graphql.ID) (dashboardID, error) {
	var dbid dashboardID
	err := relay.UnmarshalSpec(id, &dbid)
	if err != nil {
		return dashboardID{}, errors.Wrap(err, "unmarshalDashboardID")
	}
	return dbid, nil
}

func (id dashboardID) marshal() graphql.ID {
	return relay.MarshalID(dashboardKind, id)
}
