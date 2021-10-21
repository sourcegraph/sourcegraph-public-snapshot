package resolvers

import (
	"context"
	"fmt"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/database"

	"github.com/graph-gophers/graphql-go/relay"

	"github.com/cockroachdb/errors"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

var _ graphqlbackend.InsightsDashboardConnectionResolver = &dashboardConnectionResolver{}
var _ graphqlbackend.InsightsDashboardResolver = &insightsDashboardResolver{}
var _ graphqlbackend.InsightViewConnectionResolver = &DashboardInsightViewConnectionResolver{}
var _ graphqlbackend.InsightsDashboardPayloadResolver = &insightsDashboardPayloadResolver{}
var _ graphqlbackend.InsightsPermissionGrantsResolver = &insightsPermissionGrantsResolver{}

type dashboardConnectionResolver struct {
	orgStore *database.OrgStore
	args     *graphqlbackend.InsightsDashboardsArgs

	baseInsightResolver

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
		var err error
		args.UserID, args.OrgID, err = getUserPermissions(ctx, d.orgStore)
		if err != nil {
			d.err = errors.Wrap(err, "getUserPermissions")
			return
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
		resolvers = append(resolvers, &insightsDashboardResolver{dashboard: dashboard, id: &id, baseInsightResolver: d.baseInsightResolver})
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

	baseInsightResolver
}

func (i *insightsDashboardResolver) Title() string {
	return i.dashboard.Title
}

func (i *insightsDashboardResolver) ID() graphql.ID {
	return i.id.marshal()
}

func (i *insightsDashboardResolver) Views() graphqlbackend.InsightViewConnectionResolver {
	return &DashboardInsightViewConnectionResolver{ids: i.dashboard.InsightIDs, dashboard: i.dashboard, baseInsightResolver: i.baseInsightResolver}
}

func (i *insightsDashboardResolver) Grants() graphqlbackend.InsightsPermissionGrantsResolver {
	return &insightsPermissionGrantsResolver{
		UserIdGrants: i.dashboard.UserIdGrants,
		OrgIdGrants:  i.dashboard.OrgIdGrants,
		GlobalGrant:  i.dashboard.GlobalGrant,
	}
}

type insightsPermissionGrantsResolver struct {
	UserIdGrants []int64
	OrgIdGrants  []int64
	GlobalGrant  bool
}

func (i *insightsPermissionGrantsResolver) Users() []graphql.ID {
	var marshalledUserIds []graphql.ID
	for _, userIdGrant := range i.UserIdGrants {
		marshalledUserIds = append(marshalledUserIds, graphqlbackend.MarshalUserID(int32(userIdGrant)))
	}
	return marshalledUserIds
}

func (i *insightsPermissionGrantsResolver) Organizations() []graphql.ID {
	var marshalledOrgIds []graphql.ID
	for _, orgIdGrant := range i.OrgIdGrants {
		marshalledOrgIds = append(marshalledOrgIds, graphqlbackend.MarshalOrgID(int32(orgIdGrant)))
	}
	return marshalledOrgIds
}

func (i *insightsPermissionGrantsResolver) Global() bool {
	return i.GlobalGrant
}

type DashboardInsightViewConnectionResolver struct {
	baseInsightResolver

	ids       []string
	dashboard *types.Dashboard
}

func (d *DashboardInsightViewConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.InsightViewResolver, error) {
	resolvers := make([]graphqlbackend.InsightViewResolver, 0, len(d.ids))
	views, err := d.insightStore.GetMapped(ctx, store.InsightQueryArgs{UniqueIDs: d.ids, WithoutAuthorization: true})
	if err != nil {
		return nil, err
	}
	for i := range views {
		resolvers = append(resolvers, &insightViewResolver{view: &views[i], baseInsightResolver: d.baseInsightResolver})
	}
	return resolvers, nil
}

func (d *DashboardInsightViewConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}

func (r *Resolver) CreateInsightsDashboard(ctx context.Context, args *graphqlbackend.CreateInsightsDashboardArgs) (graphqlbackend.InsightsDashboardPayloadResolver, error) {
	dashboardGrants, err := parseDashboardGrants(args.Input.Grants)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse dashboard grants")
	}

	userIds, orgIds, err := getUserPermissions(ctx, database.Orgs(r.workerBaseStore.Handle().DB()))
	if err != nil {
		return nil, errors.Wrap(err, "getUserPermissions")
	}

	dashboard, err := r.dashboardStore.CreateDashboard(ctx, store.CreateDashboardArgs{
		Dashboard: types.Dashboard{Title: args.Input.Title, Save: true},
		Grants:    dashboardGrants,
		UserID:    userIds,
		OrgID:     orgIds})
	if err != nil {
		return nil, err
	}
	if dashboard == nil {
		return nil, nil
	}
	return &insightsDashboardPayloadResolver{dashboard: dashboard, baseInsightResolver: r.baseInsightResolver}, nil
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
	userIds, orgIds, err := r.ensureDashboardPermission(ctx, int(dashboardID.Arg))
	if err != nil {
		return nil, err
	}
	dashboard, err := r.dashboardStore.UpdateDashboard(ctx, store.UpdateDashboardArgs{
		ID:     int(dashboardID.Arg),
		Title:  args.Input.Title,
		Grants: dashboardGrants,
		UserID: userIds,
		OrgID:  orgIds})
	if err != nil {
		return nil, err
	}
	if dashboard == nil {
		return nil, nil
	}
	return &insightsDashboardPayloadResolver{dashboard: dashboard, baseInsightResolver: r.baseInsightResolver}, nil
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

func (r *Resolver) ensureDashboardPermission(ctx context.Context, dashboardId int) (userIds []int, orgIds []int, err error) {
	userIds, orgIds, err = getUserPermissions(ctx, database.Orgs(r.workerBaseStore.Handle().DB()))
	if err != nil {
		errors.Wrap(err, "getUserPermissions")
		return
	}
	hasPermissionToUpdate, err := r.dashboardStore.HasDashboardPermission(ctx, dashboardId, userIds, orgIds)
	if err != nil {
		errors.Wrap(err, "HasDashboardPermission")
		return
	}
	if !hasPermissionToUpdate {
		err = errors.New("this user does not have permission to modify this dashboard")
		return
	}
	return
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
	_, _, err = r.ensureDashboardPermission(ctx, int(dashboardID.Arg))
	if err != nil {
		return emptyResponse, err
	}

	err = r.dashboardStore.DeleteDashboard(ctx, dashboardID.Arg)
	if err != nil {
		return emptyResponse, err
	}
	return emptyResponse, nil
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
	userIds, orgIds, err := r.ensureDashboardPermission(ctx, int(dashboardID.Arg))
	if err != nil {
		return nil, err
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
	dashboards, err := r.dashboardStore.GetDashboards(ctx, store.DashboardQueryArgs{ID: int(dashboardID.Arg), UserID: userIds, OrgID: orgIds})
	if err != nil || len(dashboards) < 1 {
		return nil, errors.Wrap(err, "GetDashboards")
	}
	return &insightsDashboardPayloadResolver{dashboard: dashboards[0], baseInsightResolver: r.baseInsightResolver}, nil
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
	userIds, orgIds, err := r.ensureDashboardPermission(ctx, int(dashboardID.Arg))
	if err != nil {
		return nil, err
	}

	err = r.dashboardStore.RemoveViewsFromDashboard(ctx, int(dashboardID.Arg), []string{viewID})
	if err != nil {
		return nil, errors.Wrap(err, "RemoveViewsFromDashboard")
	}
	dashboards, err := r.dashboardStore.GetDashboards(ctx, store.DashboardQueryArgs{ID: int(dashboardID.Arg), UserID: userIds, OrgID: orgIds})
	if err != nil || len(dashboards) < 1 {
		return nil, errors.Wrap(err, "GetDashboards")
	}
	return &insightsDashboardPayloadResolver{dashboard: dashboards[0], baseInsightResolver: r.baseInsightResolver}, nil
}

type insightsDashboardPayloadResolver struct {
	dashboard *types.Dashboard

	baseInsightResolver
}

func (i *insightsDashboardPayloadResolver) Dashboard(ctx context.Context) (graphqlbackend.InsightsDashboardResolver, error) {
	id := newRealDashboardID(int64(i.dashboard.ID))
	return &insightsDashboardResolver{dashboard: i.dashboard, id: &id}, nil
}
