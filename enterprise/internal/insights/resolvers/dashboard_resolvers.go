package resolvers

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/graphql-go/relay"

	"github.com/cockroachdb/errors"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

var _ graphqlbackend.InsightsDashboardConnectionResolver = &dashboardConnectionResolver{}
var _ graphqlbackend.InsightsDashboardResolver = &insightsDashboardResolver{}
var _ graphqlbackend.InsightViewConnectionResolver = &stubDashboardInsightViewConnectionResolver{}
var _ graphqlbackend.InsightViewResolver = &stubInsightViewResolver{}
var _ graphqlbackend.InsightsDashboardPayloadResolver = &insightsDashboardPayloadResolver{}

type dashboardConnectionResolver struct {
	insightsDatabase dbutil.DB
	dashboardStore   store.DashboardStore
	args             *graphqlbackend.InsightsDashboardsArgs

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
			afterID, err := unmarshalDashboardID(graphql.ID(*d.args.After))
			if err != nil {
				d.err = errors.Wrap(err, "unmarshalID")
				return
			}
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
			if int64(dashboard.ID) > d.next {
				d.next = int64(dashboard.ID)
			}
		}
	})
	return d.dashboards, d.next, d.err
}

func (d *dashboardConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.InsightsDashboardResolver, error) {
	dashboards, _, err := d.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.InsightsDashboardResolver, 0, len(dashboards))
	for _, dashboard := range dashboards {
		id := newRealDashboardID(int64(dashboard.ID))
		resolvers = append(resolvers, &insightsDashboardResolver{dashboard: dashboard, id: &id})
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

type insightsDashboardResolver struct {
	dashboard *types.Dashboard
	id        *dashboardID
}

func (i *insightsDashboardResolver) Title() string {
	return i.dashboard.Title
}

func (i *insightsDashboardResolver) ID() graphql.ID {
	return i.id.marshal()
}

func (i *insightsDashboardResolver) Views() graphqlbackend.InsightViewConnectionResolver {
	return &stubDashboardInsightViewConnectionResolver{ids: i.dashboard.InsightIDs}
}

type stubDashboardInsightViewConnectionResolver struct {
	ids []string
}

func (d *stubDashboardInsightViewConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.InsightViewResolver, error) {
	resolvers := make([]graphqlbackend.InsightViewResolver, 0, len(d.ids))
	for _, id := range d.ids {
		resolvers = append(resolvers, &stubInsightViewResolver{id: id})
	}
	return resolvers, nil
}

func (d *stubDashboardInsightViewConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}

func (r *Resolver) CreateInsightsDashboard(ctx context.Context, args *graphqlbackend.CreateInsightsDashboardArgs) (graphqlbackend.InsightsDashboardPayloadResolver, error) {
	dashboardGrants, err := parseDashboardGrants(args.Input.Grants)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse dashboard grants")
	}

	dashboard, err := r.dashboardStore.CreateDashboard(ctx, types.Dashboard{Title: args.Input.Title, Save: true}, dashboardGrants)
	if err != nil {
		return nil, err
	}
	return &insightsDashboardPayloadResolver{&dashboard}, nil
}

func (r *Resolver) UpdateInsightsDashboard(ctx context.Context, args *graphqlbackend.UpdateInsightsDashboardArgs) (graphqlbackend.InsightsDashboardPayloadResolver, error) {
	var dashboardGrants []store.DashboardGrant
	if args.Input.Grants != nil {
		parsedGrants, err := parseDashboardGrants(*args.Input.Grants)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse dashboard grants")
		}
		dashboardGrants = parsedGrants
	}
	dashboardID, err := unmarshalDashboardID(args.Id)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal dashboard id")
	}
	if dashboardID.isVirtualized() {
		return nil, errors.New("unable to update a virtualized dashboard")
	}
	dashboard, err := r.dashboardStore.UpdateDashboard(ctx, store.UpdateDashboardArgs{ID: int(dashboardID.Arg), Title: args.Input.Title, Grants: dashboardGrants})
	if err != nil {
		return nil, err
	}
	return &insightsDashboardPayloadResolver{&dashboard}, nil
}

func parseDashboardGrants(inputGrants graphqlbackend.InsightsPermissionGrants) ([]store.DashboardGrant, error) {
	dashboardGrants := []store.DashboardGrant{}
	if inputGrants.Users != nil {
		for _, userGrant := range *inputGrants.Users {
			userID, err := graphqlbackend.UnmarshalUserID(userGrant)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("unable to unmarshal user id: %s", userGrant))
			}
			dashboardGrants = append(dashboardGrants, store.UserDashboardGrant(int(userID)))
		}
	}
	if inputGrants.Organizations != nil {
		for _, orgGrant := range *inputGrants.Organizations {
			orgID, err := graphqlbackend.UnmarshalOrgID(orgGrant)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("unable to unmarshal org id: %s", orgGrant))
			}
			dashboardGrants = append(dashboardGrants, store.OrgDashboardGrant(int(orgID)))
		}
	}
	if inputGrants.Global != nil && *inputGrants.Global {
		dashboardGrants = append(dashboardGrants, store.GlobalDashboardGrant())
	}
	return dashboardGrants, nil
}

func (r *Resolver) DeleteInsightsDashboard(ctx context.Context, args *graphqlbackend.DeleteInsightsDashboardArgs) (*graphqlbackend.EmptyResponse, error) {
	emptyResponse := &graphqlbackend.EmptyResponse{}

	dashboardID, err := unmarshalDashboardID(args.Id)
	if err != nil {
		return emptyResponse, err
	}
	if dashboardID.isVirtualized() {
		return emptyResponse, nil
	}

	err = r.dashboardStore.DeleteDashboard(ctx, dashboardID.Arg)
	if err != nil {
		return emptyResponse, err
	}
	return emptyResponse, nil
}

type stubInsightViewResolver struct {
	id string
}

func (s *stubInsightViewResolver) ID() graphql.ID {
	return relay.MarshalID("insight_view", s.id)
}

func (s *stubInsightViewResolver) VeryUniqueResolver() bool {
	return true
}

func (r *Resolver) AddInsightViewToDashboard(ctx context.Context, args *graphqlbackend.AddInsightViewToDashboardArgs) (graphqlbackend.InsightsDashboardPayloadResolver, error) {
	var viewID string
	err := relay.UnmarshalSpec(args.Input.InsightViewID, &viewID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal insight view id")
	}
	dashboardID, err := unmarshalDashboardID(args.Input.DashboardID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal dashboard id")
	}

	exists, err := r.dashboardStore.IsViewOnDashboard(ctx, int(dashboardID.Arg), viewID)
	if err != nil {
		return nil, errors.Wrap(err, "IsViewOnDashboard")
	}
	if exists {
		return nil, errors.New("this insight view is already attached to this dashboard")
	}

	err = r.dashboardStore.AddViewsToDashboard(ctx, int(dashboardID.Arg), []string{viewID})
	if err != nil {
		return nil, errors.Wrap(err, "AddInsightViewToDashboard")
	}
	dashboards, err := r.dashboardStore.GetDashboards(ctx, store.DashboardQueryArgs{ID: int(dashboardID.Arg)})
	if err != nil || len(dashboards) < 1 {
		return nil, errors.Wrap(err, "GetDashboards")
	}
	return &insightsDashboardPayloadResolver{dashboard: dashboards[0]}, nil
}

func (r *Resolver) RemoveInsightViewFromDashboard(ctx context.Context, args *graphqlbackend.RemoveInsightViewFromDashboardArgs) (graphqlbackend.InsightsDashboardPayloadResolver, error) {
	var viewID string
	err := relay.UnmarshalSpec(args.Input.InsightViewID, &viewID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal insight view id")
	}
	dashboardID, err := unmarshalDashboardID(args.Input.DashboardID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal dashboard id")
	}

	err = r.dashboardStore.RemoveViewsFromDashboard(ctx, int(dashboardID.Arg), []string{viewID})
	if err != nil {
		return nil, errors.Wrap(err, "RemoveViewsFromDashboard")
	}
	dashboards, err := r.dashboardStore.GetDashboards(ctx, store.DashboardQueryArgs{ID: int(dashboardID.Arg)})
	if err != nil || len(dashboards) < 1 {
		return nil, errors.Wrap(err, "GetDashboards")
	}
	return &insightsDashboardPayloadResolver{dashboard: dashboards[0]}, nil
}

type insightsDashboardPayloadResolver struct {
	dashboard *types.Dashboard
}

func (i *insightsDashboardPayloadResolver) Dashboard(ctx context.Context) (graphqlbackend.InsightsDashboardResolver, error) {
	id := newRealDashboardID(int64(i.dashboard.ID))
	return &insightsDashboardResolver{dashboard: i.dashboard, id: &id}, nil
}
