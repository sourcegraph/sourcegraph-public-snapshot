package resolvers

import (
	"context"
	"sync"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/state"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
)

var _ graphqlbackend.CampaignResolver = &campaignResolver{}

type campaignResolver struct {
	store    *store.Store
	campaign *campaigns.Campaign

	// Cache the namespace on the resolver, since it's accessed more than once.
	namespaceOnce sync.Once
	namespace     graphqlbackend.NamespaceResolver
	namespaceErr  error
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
	return marshalCampaignID(r.campaign.ID)
}

func (r *campaignResolver) Name() string {
	return r.campaign.Name
}

func (r *campaignResolver) Description() *string {
	if r.campaign.Description == "" {
		return nil
	}
	return &r.campaign.Description
}

func (r *campaignResolver) InitialApplier(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	user, err := graphqlbackend.UserByIDInt32(ctx, r.campaign.InitialApplierID)
	if errcode.IsNotFound(err) {
		return nil, nil
	}
	return user, err
}

func (r *campaignResolver) LastApplier(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	user, err := graphqlbackend.UserByIDInt32(ctx, r.campaign.LastApplierID)
	if errcode.IsNotFound(err) {
		return nil, nil
	}
	return user, err
}

func (r *campaignResolver) LastAppliedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.campaign.LastAppliedAt}
}

func (r *campaignResolver) SpecCreator(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	spec, err := r.store.GetCampaignSpec(ctx, store.GetCampaignSpecOpts{
		ID: r.campaign.CampaignSpecID,
	})
	if err != nil {
		return nil, err
	}
	user, err := graphqlbackend.UserByIDInt32(ctx, spec.UserID)
	if errcode.IsNotFound(err) {
		return nil, nil
	}
	return user, err
}

func (r *campaignResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	return checkSiteAdminOrSameUser(ctx, r.campaign.InitialApplierID)
}

func (r *campaignResolver) URL(ctx context.Context) (string, error) {
	n, err := r.Namespace(ctx)
	if err != nil {
		return "", err
	}
	return campaignURL(n, r), nil
}

func (r *campaignResolver) Namespace(ctx context.Context) (graphqlbackend.NamespaceResolver, error) {
	return r.computeNamespace(ctx)
}

func (r *campaignResolver) computeNamespace(ctx context.Context) (graphqlbackend.NamespaceResolver, error) {
	r.namespaceOnce.Do(func() {
		if r.campaign.NamespaceUserID != 0 {
			r.namespace.Namespace, r.namespaceErr = graphqlbackend.UserByIDInt32(
				ctx,
				r.campaign.NamespaceUserID,
			)
		} else {
			r.namespace.Namespace, r.namespaceErr = graphqlbackend.OrgByIDInt32(
				ctx,
				r.campaign.NamespaceOrgID,
			)
		}
		if errcode.IsNotFound(r.namespaceErr) {
			r.namespace.Namespace = nil
			r.namespaceErr = errors.New("namespace of campaign has been deleted")
		}
	})

	return r.namespace, r.namespaceErr
}

func (r *campaignResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.campaign.CreatedAt}
}

func (r *campaignResolver) UpdatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.campaign.UpdatedAt}
}

func (r *campaignResolver) ClosedAt() *graphqlbackend.DateTime {
	if !r.campaign.Closed() {
		return nil
	}
	return &graphqlbackend.DateTime{Time: r.campaign.ClosedAt}
}

func (r *campaignResolver) ChangesetsStats(ctx context.Context) (graphqlbackend.ChangesetsStatsResolver, error) {
	stats, err := r.store.GetChangesetsStats(ctx, store.GetChangesetsStatsOpts{
		CampaignID: r.campaign.ID,
	})
	if err != nil {
		return nil, err
	}
	return &changesetsStatsResolver{stats: stats}, nil
}

func (r *campaignResolver) Changesets(
	ctx context.Context,
	args *graphqlbackend.ListChangesetsArgs,
) (graphqlbackend.ChangesetsConnectionResolver, error) {
	opts, safe, err := listChangesetOptsFromArgs(args, r.campaign.ID)
	if err != nil {
		return nil, err
	}
	opts.CampaignID = r.campaign.ID
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
	resolvers := []graphqlbackend.ChangesetCountsResolver{}

	publishedState := campaigns.ChangesetPublicationStatePublished
	opts := store.ListChangesetsOpts{
		CampaignID: r.campaign.ID,
		// Only load fully-synced changesets, so that the data we use for computing the changeset counts is complete.
		PublicationState: &publishedState,
	}
	cs, _, err := r.store.ListChangesets(ctx, opts)
	if err != nil {
		return resolvers, err
	}

	now := r.store.Clock()()

	weekAgo := now.Add(-7 * 24 * time.Hour)
	start := r.campaign.CreatedAt.UTC()
	if start.After(weekAgo) {
		start = weekAgo
	}
	if args.From != nil {
		start = args.From.Time.UTC()
	}

	end := now.UTC()
	if args.To != nil && args.To.Time.Before(end) {
		end = args.To.Time.UTC()
	}

	var es []*campaigns.ChangesetEvent
	changesetIDs := cs.IDs()
	if len(changesetIDs) > 0 {
		eventsOpts := store.ListChangesetEventsOpts{ChangesetIDs: changesetIDs, Kinds: state.RequiredEventTypesForHistory}
		es, _, err = r.store.ListChangesetEvents(ctx, eventsOpts)
		if err != nil {
			return resolvers, err
		}
	}

	counts, err := state.CalcCounts(start, end, cs, es...)
	if err != nil {
		return resolvers, err
	}

	for _, c := range counts {
		resolvers = append(resolvers, &changesetCountsResolver{counts: c})
	}

	return resolvers, nil
}

func (r *campaignResolver) DiffStat(ctx context.Context) (*graphqlbackend.DiffStat, error) {
	diffStat, err := r.store.GetCampaignDiffStat(ctx, store.GetCampaignDiffStatOpts{CampaignID: r.campaign.ID})
	if err != nil {
		return nil, err
	}
	return graphqlbackend.NewDiffStat(*diffStat), nil
}

func (r *campaignResolver) CurrentSpec(ctx context.Context) (graphqlbackend.CampaignSpecResolver, error) {
	campaignSpec, err := r.store.GetCampaignSpec(ctx, store.GetCampaignSpecOpts{ID: r.campaign.CampaignSpecID})
	if err != nil {
		// This spec should always exist, so fail hard on not found errors as well.
		return nil, err
	}

	return &campaignSpecResolver{store: r.store, campaignSpec: campaignSpec}, nil
}
