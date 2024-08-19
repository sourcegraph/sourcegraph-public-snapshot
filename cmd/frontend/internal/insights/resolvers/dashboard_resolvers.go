package resolvers

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ graphqlbackend.InsightsDashboardConnectionResolver = &dashboardConnectionResolver{}
var _ graphqlbackend.InsightsDashboardResolver = &insightsDashboardResolver{}
var _ graphqlbackend.InsightViewConnectionResolver = &DashboardInsightViewConnectionResolver{}
var _ graphqlbackend.InsightsDashboardPayloadResolver = &insightsDashboardPayloadResolver{}
var _ graphqlbackend.InsightsPermissionGrantsResolver = &insightsPermissionGrantsResolver{}

type dashboardConnectionResolver struct {
	orgStore         database.OrgStore
	args             *graphqlbackend.InsightsDashboardsArgs
	withViewUniqueID *string

	baseInsightResolver

	// Cache results because they are used by multiple fields
	once       sync.Once
	dashboards []*types.Dashboard
	next       int64
	err        error
}

func (d *dashboardConnectionResolver) compute(ctx context.Context) ([]*types.Dashboard, error) {
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
		args.UserIDs, args.OrgIDs, err = getUserPermissions(ctx, d.orgStore)
		if err != nil {
			d.err = errors.Wrap(err, "getUserPermissions")
			return
		}

		if d.args.ID != nil {
			id, err := unmarshalDashboardID(*d.args.ID)
			if err != nil {
				d.err = errors.Wrap(err, "unmarshalDashboardID")
			}
			if !id.isVirtualized() {
				args.IDs = []int{int(id.Arg)}
			}
		}

		if d.withViewUniqueID != nil {
			args.WithViewUniqueID = d.withViewUniqueID
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
	return d.dashboards, d.err
}

func (d *dashboardConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.InsightsDashboardResolver, error) {
	dashboards, err := d.compute(ctx)
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

func (d *dashboardConnectionResolver) PageInfo(ctx context.Context) (*gqlutil.PageInfo, error) {
	_, err := d.compute(ctx)
	if err != nil {
		return nil, err
	}
	if d.next != 0 {
		return gqlutil.NextPageCursor(string(newRealDashboardID(d.next).marshal())), nil
	}
	return gqlutil.HasNextPage(false), nil
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

func (i *insightsDashboardResolver) Views(ctx context.Context, args graphqlbackend.DashboardInsightViewConnectionArgs) graphqlbackend.InsightViewConnectionResolver {
	return &DashboardInsightViewConnectionResolver{ids: i.dashboard.InsightIDs, dashboard: i.dashboard, baseInsightResolver: i.baseInsightResolver, args: args}
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

	args graphqlbackend.DashboardInsightViewConnectionArgs

	ids       []string
	dashboard *types.Dashboard

	once  sync.Once
	views []types.Insight
	next  string
	err   error
}

func (d *DashboardInsightViewConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.InsightViewResolver, error) {
	resolvers := make([]graphqlbackend.InsightViewResolver, 0, len(d.ids))
	views, _, err := d.computeConnectedViews(ctx)
	if err != nil {
		return nil, err
	}
	for i := range views {
		resolvers = append(resolvers, &insightViewResolver{view: &views[i], baseInsightResolver: d.baseInsightResolver})
	}
	return resolvers, nil
}

func (d *DashboardInsightViewConnectionResolver) PageInfo(ctx context.Context) (*gqlutil.PageInfo, error) {
	return gqlutil.HasNextPage(false), nil
}

func (d *DashboardInsightViewConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	args := store.InsightsOnDashboardQueryArgs{DashboardID: d.dashboard.ID}
	var err error
	viewSeries, err := d.insightStore.GetAllOnDashboard(ctx, args)
	if err != nil {
		return nil, err
	}
	views := d.insightStore.GroupByView(ctx, viewSeries)
	count := int32(len(views))
	return &count, nil
}

func (d *DashboardInsightViewConnectionResolver) computeConnectedViews(ctx context.Context) ([]types.Insight, string, error) {
	d.once.Do(func() {
		args := store.InsightsOnDashboardQueryArgs{DashboardID: d.dashboard.ID}
		if d.args.After != nil {
			var afterID string
			err := relay.UnmarshalSpec(graphql.ID(*d.args.After), &afterID)
			if err != nil {
				d.err = errors.Wrap(err, "unmarshalID")
				return
			}
			args.After = afterID
		}
		if d.args.First != nil {
			args.Limit = int(*d.args.First)
		}
		var err error

		viewSeries, err := d.insightStore.GetAllOnDashboard(ctx, args)
		if err != nil {
			d.err = err
			return
		}

		d.views = d.insightStore.GroupByView(ctx, viewSeries)
		sort.Slice(d.views, func(i, j int) bool {
			return d.views[i].DashboardViewId < d.views[j].DashboardViewId
		})

		if len(d.views) > 0 {
			d.next = fmt.Sprintf("%d", d.views[len(d.views)-1].DashboardViewId)
		}
	})
	return d.views, d.next, d.err
}

func (r *Resolver) CreateInsightsDashboard(ctx context.Context, args *graphqlbackend.CreateInsightsDashboardArgs) (graphqlbackend.InsightsDashboardPayloadResolver, error) {
	dashboardGrants, err := parseDashboardGrants(args.Input.Grants)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse dashboard grants")
	}
	if len(dashboardGrants) == 0 {
		return nil, errors.New("dashboard must be created with at least one grant")
	}

	userIds, orgIds, err := getUserPermissions(ctx, database.NewDBWith(r.logger, r.workerBaseStore).Orgs())
	if err != nil {
		return nil, errors.Wrap(err, "getUserPermissions")
	}
	hasPermissionToCreate := hasPermissionForGrants(dashboardGrants, userIds, orgIds)
	if !hasPermissionToCreate {
		return nil, errors.New("user does not have permission to create this dashboard")
	}

	dashboard, err := r.dashboardStore.CreateDashboard(ctx, store.CreateDashboardArgs{
		Dashboard: types.Dashboard{Title: args.Input.Title, Save: true},
		Grants:    dashboardGrants,
		UserIDs:   userIds,
		OrgIDs:    orgIds})
	if err != nil {
		return nil, err
	}
	if dashboard == nil {
		return nil, nil
	}
	return &insightsDashboardPayloadResolver{dashboard: dashboard, baseInsightResolver: r.baseInsightResolver}, nil
}

func (r *Resolver) UpdateInsightsDashboard(ctx context.Context, args *graphqlbackend.UpdateInsightsDashboardArgs) (graphqlbackend.InsightsDashboardPayloadResolver, error) {
	permissionsValidator := PermissionsValidatorFromBase(&r.baseInsightResolver)

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

	err = permissionsValidator.validateUserAccessForDashboard(ctx, int(dashboardID.Arg))
	if err != nil {
		return nil, err
	}

	dashboard, err := r.dashboardStore.UpdateDashboard(ctx, store.UpdateDashboardArgs{
		ID:      int(dashboardID.Arg),
		Title:   args.Input.Title,
		Grants:  dashboardGrants,
		UserIDs: permissionsValidator.userIds,
		OrgIDs:  permissionsValidator.orgIds})
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

// Checks that each grant is contained in the available user/org ids.
func hasPermissionForGrants(dashboardGrants []store.DashboardGrant, userIds []int, orgIds []int) bool {
	allowedUsers := make(map[int]bool)
	allowedOrgs := make(map[int]bool)

	for _, userId := range userIds {
		allowedUsers[userId] = true
	}
	for _, orgId := range orgIds {
		allowedOrgs[orgId] = true
	}

	for _, requestedGrant := range dashboardGrants {
		if requestedGrant.UserID != nil {
			if _, ok := allowedUsers[*requestedGrant.UserID]; !ok {
				return false
			}
		}
		if requestedGrant.OrgID != nil {
			if _, ok := allowedOrgs[*requestedGrant.OrgID]; !ok {
				return false
			}
		}
	}
	return true
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

	if licenseError := licensing.Check(licensing.FeatureCodeInsights); licenseError != nil {
		lamDashboardId, err := r.dashboardStore.EnsureLimitedAccessModeDashboard(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "EnsureLimitedAccessModeDashboard")
		}
		if lamDashboardId == int(dashboardID.Arg) {
			return nil, errors.New("Cannot delete this dashboard in Limited Access Mode")
		}
	}

	permissionsValidator := PermissionsValidatorFromBase(&r.baseInsightResolver)
	err = permissionsValidator.validateUserAccessForDashboard(ctx, int(dashboardID.Arg))
	if err != nil {
		return nil, err
	}

	err = r.dashboardStore.DeleteDashboard(ctx, int(dashboardID.Arg))
	if err != nil {
		return emptyResponse, err
	}
	return emptyResponse, nil
}

func (r *Resolver) AddInsightViewToDashboard(ctx context.Context, args *graphqlbackend.AddInsightViewToDashboardArgs) (_ graphqlbackend.InsightsDashboardPayloadResolver, err error) {
	var viewID string
	err = relay.UnmarshalSpec(args.Input.InsightViewID, &viewID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal insight view id")
	}
	dashboardID, err := unmarshalDashboardID(args.Input.DashboardID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal dashboard id")
	}

	tx, err := r.dashboardStore.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	if licenseError := licensing.Check(licensing.FeatureCodeInsights); licenseError != nil {
		lamDashboardId, err := tx.EnsureLimitedAccessModeDashboard(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "EnsureLimitedAccessModeDashboard")
		}
		if lamDashboardId == int(dashboardID.Arg) {
			return nil, errors.New("Cannot add insights to this dashboard while in Limited Access Mode")
		}
	}

	permissionsValidator := PermissionsValidatorFromBase(&r.baseInsightResolver)
	txValidator := permissionsValidator.WithBaseStore(tx.Store)
	err = txValidator.validateUserAccessForDashboard(ctx, int(dashboardID.Arg))
	if err != nil {
		return nil, err
	}
	err = txValidator.validateUserAccessForView(ctx, viewID)
	if err != nil {
		return nil, err
	}

	exists, err := tx.IsViewOnDashboard(ctx, int(dashboardID.Arg), viewID)
	if err != nil {
		return nil, errors.Wrap(err, "IsViewOnDashboard")
	}
	if !exists {
		r.logger.Debug("attempting to add insight view to dashboard", log.Int64("dashboardID", dashboardID.Arg), log.String("insightViewID", viewID))
		err = tx.AddViewsToDashboard(ctx, int(dashboardID.Arg), []string{viewID})
		if err != nil {
			return nil, errors.Wrap(err, "AddInsightViewToDashboard")
		}
	}

	dashboards, err := tx.GetDashboards(ctx, store.DashboardQueryArgs{IDs: []int{int(dashboardID.Arg)},
		UserIDs: txValidator.userIds, OrgIDs: txValidator.orgIds})
	if err != nil {
		return nil, errors.Wrap(err, "GetDashboards")
	} else if len(dashboards) < 1 {
		return nil, errors.New("dashboard not found")
	}

	return &insightsDashboardPayloadResolver{dashboard: dashboards[0], baseInsightResolver: r.baseInsightResolver}, nil
}

func (r *Resolver) RemoveInsightViewFromDashboard(ctx context.Context, args *graphqlbackend.RemoveInsightViewFromDashboardArgs) (_ graphqlbackend.InsightsDashboardPayloadResolver, err error) {
	var viewID string
	err = relay.UnmarshalSpec(args.Input.InsightViewID, &viewID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal insight view id")
	}
	dashboardID, err := unmarshalDashboardID(args.Input.DashboardID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal dashboard id")
	}

	tx, err := r.dashboardStore.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	if licenseError := licensing.Check(licensing.FeatureCodeInsights); licenseError != nil {
		lamDashboardId, err := tx.EnsureLimitedAccessModeDashboard(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "EnsureLimitedAccessModeDashboard")
		}
		if lamDashboardId == int(dashboardID.Arg) {
			return nil, errors.New("Cannot remove insights from this dashboard while in Limited Access Mode")
		}
	}

	permissionsValidator := PermissionsValidatorFromBase(&r.baseInsightResolver)
	txValidator := permissionsValidator.WithBaseStore(tx.Store)
	err = txValidator.validateUserAccessForDashboard(ctx, int(dashboardID.Arg))
	if err != nil {
		return nil, err
	}

	err = tx.RemoveViewsFromDashboard(ctx, int(dashboardID.Arg), []string{viewID})
	if err != nil {
		return nil, errors.Wrap(err, "RemoveViewsFromDashboard")
	}
	dashboards, err := tx.GetDashboards(ctx, store.DashboardQueryArgs{IDs: []int{int(dashboardID.Arg)},
		UserIDs: txValidator.userIds, OrgIDs: txValidator.orgIds})
	if err != nil {
		return nil, errors.Wrap(err, "GetDashboards")
	} else if len(dashboards) < 1 {
		return nil, errors.New("dashboard not found")
	}
	return &insightsDashboardPayloadResolver{dashboard: dashboards[0], baseInsightResolver: r.baseInsightResolver}, nil
}

type insightsDashboardPayloadResolver struct {
	dashboard *types.Dashboard

	baseInsightResolver
}

func (i *insightsDashboardPayloadResolver) Dashboard(ctx context.Context) (graphqlbackend.InsightsDashboardResolver, error) {
	id := newRealDashboardID(int64(i.dashboard.ID))
	return &insightsDashboardResolver{dashboard: i.dashboard, id: &id, baseInsightResolver: i.baseInsightResolver}, nil
}
