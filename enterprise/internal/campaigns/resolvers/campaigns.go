package resolvers

import (
	"context"
	"path"
	"sync"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

var _ graphqlbackend.CampaignsConnectionResolver = &campaignsConnectionResolver{}

type campaignsConnectionResolver struct {
	store       *ee.Store
	httpFactory *httpcli.Factory
	opts        ee.ListCampaignsOpts

	// cache results because they are used by multiple fields
	once      sync.Once
	campaigns []*campaigns.Campaign
	next      int64
	err       error
}

func (r *campaignsConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.CampaignResolver, error) {
	nodes, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.CampaignResolver, 0, len(nodes))
	for _, c := range nodes {
		resolvers = append(resolvers, &campaignResolver{store: r.store, httpFactory: r.httpFactory, Campaign: c})
	}
	return resolvers, nil
}

func (r *campaignsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	opts := ee.CountCampaignsOpts{ChangesetID: r.opts.ChangesetID, State: r.opts.State, OnlyForAuthor: r.opts.OnlyForAuthor}
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
	store       *ee.Store
	httpFactory *httpcli.Factory
	*campaigns.Campaign
}

func (r *campaignResolver) ID() graphql.ID {
	return campaigns.MarshalCampaignID(r.Campaign.ID)
}

func (r *campaignResolver) Name() string {
	return r.Campaign.Name
}

func (r *campaignResolver) Description() *string {
	if r.Campaign.Description == "" {
		return nil
	}
	return &r.Campaign.Description
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
	return checkSiteAdminOrSameUser(ctx, r.Campaign.AuthorID)
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

func (r *campaignResolver) Changesets(
	ctx context.Context,
	args *graphqlbackend.ListChangesetsArgs,
) (graphqlbackend.ChangesetsConnectionResolver, error) {
	opts, safe, err := listChangesetOptsFromArgs(args)
	if err != nil {
		return nil, err
	}
	opts.CampaignID = r.Campaign.ID
	return &changesetsConnectionResolver{
		store:    r.store,
		opts:     opts,
		optsSafe: safe,
	}, nil
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

	opts := ee.ListChangesetsOpts{CampaignID: r.Campaign.ID, Limit: -1}
	cs, _, err := r.store.ListChangesets(ctx, opts)
	if err != nil {
		return resolvers, err
	}

	weekAgo := time.Now().Add(-7 * 24 * time.Hour)
	start := r.Campaign.CreatedAt.UTC()
	if start.After(weekAgo) {
		start = weekAgo
	}
	if args.From != nil {
		start = args.From.Time.UTC()
	}

	end := time.Now().UTC()
	if args.To != nil && args.To.Time.Before(end) {
		end = args.To.Time.UTC()
	}

	eventsOpts := ee.ListChangesetEventsOpts{ChangesetIDs: cs.IDs(), Limit: -1}
	es, _, err := r.store.ListChangesetEvents(ctx, eventsOpts)
	if err != nil {
		return resolvers, err
	}

	counts, err := ee.CalcCounts(start, end, cs, es...)
	if err != nil {
		return resolvers, err
	}

	for _, c := range counts {
		resolvers = append(resolvers, &changesetCountsResolver{counts: c})
	}

	return resolvers, nil
}

func (r *campaignResolver) DiffStat(ctx context.Context) (*graphqlbackend.DiffStat, error) {
	changesetsConnection := &changesetsConnectionResolver{
		store: r.store,
		opts: ee.ListChangesetsOpts{
			CampaignID: r.Campaign.ID,
			Limit:      -1, // Get all changesets
		},
		optsSafe: true,
	}

	changesets, err := changesetsConnection.Nodes(ctx)
	if err != nil {
		return nil, err
	}

	totalStat := &graphqlbackend.DiffStat{}
	for _, cs := range changesets {
		// Not being able to convert is OK; it just means there's a hidden
		// changeset that we can't use the stats from.
		if external, ok := cs.ToExternalChangeset(); ok && external != nil {
			stat, err := external.DiffStat(ctx)
			if err != nil {
				return nil, err
			}
			if stat != nil {
				totalStat.AddDiffStat(stat)
			}
		}
	}

	return totalStat, nil
}
