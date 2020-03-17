package resolvers

import (
	"context"
	"path"
	"sync"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

var _ graphqlbackend.CampaignsConnectionResolver = &campaignsConnectionResolver{}

type campaignsConnectionResolver struct {
	store *ee.Store
	opts  ee.ListCampaignsOpts

	// cache results because they are used by multiple fields
	once      sync.Once
	campaigns []*campaigns.Campaign
	next      int64
	err       error
}

func (r *campaignsConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.CampaignResolver, error) {
	campaigns, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.CampaignResolver, 0, len(campaigns))
	for _, c := range campaigns {
		resolvers = append(resolvers, &campaignResolver{store: r.store, Campaign: c})
	}
	return resolvers, nil
}

func (r *campaignsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	opts := ee.CountCampaignsOpts{ChangesetID: r.opts.ChangesetID, State: r.opts.State}
	count, err := r.store.CountCampaigns(ctx, opts)
	return int32(count), err
}

func (r *campaignsConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(next != 0), nil
}

func (r *campaignsConnectionResolver) compute(ctx context.Context) ([]*campaigns.Campaign, int64, error) {
	r.once.Do(func() {
		r.campaigns, r.next, r.err = r.store.ListCampaigns(ctx, r.opts)
	})
	return r.campaigns, r.next, r.err
}

var _ graphqlbackend.CampaignResolver = &campaignResolver{}

type campaignResolver struct {
	store *ee.Store
	*campaigns.Campaign
}

const campaignIDKind = "Campaign"

func marshalCampaignID(id int64) graphql.ID {
	return relay.MarshalID(campaignIDKind, id)
}

func unmarshalCampaignID(id graphql.ID) (campaignID int64, err error) {
	err = relay.UnmarshalSpec(id, &campaignID)
	return
}

func (r *campaignResolver) ID() graphql.ID {
	return marshalCampaignID(r.Campaign.ID)
}

func (r *campaignResolver) Name() string {
	return r.Campaign.Name
}

func (r *campaignResolver) Description() string {
	return r.Campaign.Description
}

func (r *campaignResolver) Branch() *string {
	if r.Campaign.Branch == "" {
		return nil
	}
	return &r.Campaign.Branch
}

func (r *campaignResolver) Author(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	return graphqlbackend.UserByIDInt32(ctx, r.AuthorID)
}

func (r *campaignResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	currentUser, err := backend.CurrentUser(ctx)
	if err != nil {
		return false, err
	}
	return currentUser.SiteAdmin, nil
}

func (r *campaignResolver) URL(ctx context.Context) (string, error) {
	return path.Join("/campaigns", string(r.ID())), nil
}

func (r *campaignResolver) Namespace(ctx context.Context) (n graphqlbackend.NamespaceResolver, err error) {
	if r.NamespaceUserID != 0 {
		n.Namespace, err = graphqlbackend.UserByIDInt32(ctx, r.NamespaceUserID)
	} else {
		n.Namespace, err = graphqlbackend.OrgByIDInt32(ctx, r.NamespaceOrgID)
	}

	return n, err
}

func (r *campaignResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.Campaign.CreatedAt}
}

func (r *campaignResolver) UpdatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.Campaign.UpdatedAt}
}

func (r *campaignResolver) ClosedAt() *graphqlbackend.DateTime {
	if r.Campaign.ClosedAt.IsZero() {
		return nil
	}
	return &graphqlbackend.DateTime{Time: r.Campaign.ClosedAt}
}

func (r *campaignResolver) PublishedAt(ctx context.Context) (*graphqlbackend.DateTime, error) {
	if r.Campaign.CampaignPlanID == 0 {
		return &graphqlbackend.DateTime{Time: r.Campaign.CreatedAt}, nil
	}

	createdAt, err := r.store.GetLatestChangesetJobCreatedAt(ctx, r.Campaign.ID)
	if err != nil {
		return nil, err
	}
	if createdAt.IsZero() {
		return nil, nil
	}
	return &graphqlbackend.DateTime{Time: createdAt}, nil
}

func (r *campaignResolver) Changesets(
	ctx context.Context,
	args *graphqlbackend.ListChangesetsArgs,
) (graphqlbackend.ExternalChangesetsConnectionResolver, error) {
	opts, err := listChangesetOptsFromArgs(args)
	if err != nil {
		return nil, err
	}
	opts.CampaignID = r.Campaign.ID
	return &changesetsConnectionResolver{
		store: r.store,
		opts:  opts,
	}, nil
}

func (r *campaignResolver) ChangesetPlans(
	ctx context.Context,
	args *graphqlutil.ConnectionArgs,
) graphqlbackend.ChangesetPlansConnectionResolver {
	if r.Campaign.CampaignPlanID == 0 {
		return &emptyChangesetPlansConnectionsResolver{}
	}

	return &campaignJobsConnectionResolver{
		store: r.store,
		opts: ee.ListCampaignJobsOpts{
			CampaignPlanID:            r.Campaign.CampaignPlanID,
			Limit:                     int(args.GetFirst()),
			OnlyWithDiff:              true,
			OnlyUnpublishedInCampaign: r.Campaign.ID,
		},
	}
}

func (r *campaignResolver) ChangesetCountsOverTime(
	ctx context.Context,
	args *graphqlbackend.ChangesetCountsArgs,
) ([]graphqlbackend.ChangesetCountsResolver, error) {
	// 🚨 SECURITY: Only site admins or users when read-access is enabled may access changesets.
	if err := allowReadAccess(ctx); err != nil {
		return nil, err
	}

	resolvers := []graphqlbackend.ChangesetCountsResolver{}

	opts := ee.ListChangesetsOpts{CampaignID: r.Campaign.ID}
	cs, _, err := r.store.ListChangesets(ctx, opts)
	if err != nil {
		return resolvers, err
	}

	start := r.Campaign.CreatedAt.UTC()
	if args.From != nil {
		start = args.From.Time.UTC()
	}

	end := time.Now().UTC()
	if args.To != nil && args.To.Time.Before(end) {
		end = args.To.Time.UTC()
	}

	changesetIDs := make([]int64, len(cs))
	for i, c := range cs {
		changesetIDs[i] = c.ID
	}

	eventsOpts := ee.ListChangesetEventsOpts{
		ChangesetIDs: changesetIDs,
		Limit:        -1,
	}
	es, _, err := r.store.ListChangesetEvents(ctx, eventsOpts)
	if err != nil {
		return resolvers, err
	}

	events := make([]ee.Event, len(es))
	for i, e := range es {
		events[i] = e
	}

	counts, err := ee.CalcCounts(start, end, cs, events...)
	if err != nil {
		return resolvers, err
	}

	for _, c := range counts {
		resolvers = append(resolvers, &changesetCountsResolver{counts: c})
	}

	return resolvers, nil
}

func (r *campaignResolver) Plan(ctx context.Context) (graphqlbackend.CampaignPlanResolver, error) {
	if r.Campaign.CampaignPlanID == 0 {
		return nil, nil
	}

	plan, err := r.store.GetCampaignPlan(ctx, ee.GetCampaignPlanOpts{ID: r.Campaign.CampaignPlanID})
	if err != nil {
		return nil, err
	}

	return &campaignPlanResolver{store: r.store, campaignPlan: plan}, nil
}

func (r *campaignResolver) RepositoryDiffs(
	ctx context.Context,
	args *graphqlutil.ConnectionArgs,
) (graphqlbackend.RepositoryComparisonConnectionResolver, error) {
	changesetsConnection := &changesetsConnectionResolver{
		store: r.store,
		opts: ee.ListChangesetsOpts{
			CampaignID: r.Campaign.ID,
			Limit:      int(args.GetFirst()),
		},
	}
	return &changesetDiffsConnectionResolver{changesetsConnection}, nil
}

func (r *campaignResolver) Status(ctx context.Context) (graphqlbackend.BackgroundProcessStatus, error) {
	return r.store.GetCampaignStatus(ctx, r.Campaign.ID)
}

type changesetDiffsConnectionResolver struct {
	*changesetsConnectionResolver
}

func (r *changesetDiffsConnectionResolver) Nodes(ctx context.Context) ([]*graphqlbackend.RepositoryComparisonResolver, error) {
	changesets, err := r.changesetsConnectionResolver.Nodes(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*graphqlbackend.RepositoryComparisonResolver, 0, len(changesets))
	for _, c := range changesets {
		comp, err := c.Diff(ctx)
		if err != nil {
			return nil, err
		}
		if comp != nil {
			resolvers = append(resolvers, comp)
		}
	}
	return resolvers, nil
}

type emptyChangesetPlansConnectionsResolver struct{}

func (r *emptyChangesetPlansConnectionsResolver) Nodes(ctx context.Context) ([]graphqlbackend.ChangesetPlanResolver, error) {
	return []graphqlbackend.ChangesetPlanResolver{}, nil
}

func (r *emptyChangesetPlansConnectionsResolver) TotalCount(ctx context.Context) (int32, error) {
	return 0, nil
}

func (r *emptyChangesetPlansConnectionsResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}
