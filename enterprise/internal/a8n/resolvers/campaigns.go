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
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
)

type campaignsConnectionResolver struct {
	store *ee.Store
	opts  ee.ListCampaignsOpts

	// cache results because they are used by multiple fields
	once      sync.Once
	campaigns []*a8n.Campaign
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
	opts := ee.CountCampaignsOpts{ChangesetID: r.opts.ChangesetID}
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

func (r *campaignsConnectionResolver) compute(ctx context.Context) ([]*a8n.Campaign, int64, error) {
	r.once.Do(func() {
		r.campaigns, r.next, r.err = r.store.ListCampaigns(ctx, r.opts)
	})
	return r.campaigns, r.next, r.err
}

type campaignResolver struct {
	store *ee.Store
	*a8n.Campaign
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

func (r *campaignResolver) Author(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	return graphqlbackend.UserByIDInt32(ctx, r.AuthorID)
}

func (r *campaignResolver) URL(ctx context.Context) (string, error) {
	// TODO(tsenart): Query for namespace only once
	ns, err := r.Namespace(ctx)
	if err != nil {
		return "", err
	}
	return path.Join(ns.URL(), "campaigns", string(r.ID())), nil
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

func (r *campaignResolver) Changesets(ctx context.Context, args struct {
	graphqlutil.ConnectionArgs
}) graphqlbackend.ExternalChangesetsConnectionResolver {
	return &changesetsConnectionResolver{
		store: r.store,
		opts: ee.ListChangesetsOpts{
			CampaignID: r.Campaign.ID,
			Limit:      int(args.ConnectionArgs.GetFirst()),
		},
	}
}

func (r *campaignResolver) ChangesetCountsOverTime(
	ctx context.Context,
	args *graphqlbackend.ChangesetCountsArgs,
) ([]graphqlbackend.ChangesetCountsResolver, error) {
	// 🚨 SECURITY: Only site admins may access the counts for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
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

func (r *campaignResolver) ChangesetCreationStatus(ctx context.Context) (graphqlbackend.BackgroundProcessStatus, error) {
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
