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

var _ graphqlbackend.BatchChangeResolver = &batchChangeResolver{}

type batchChangeResolver struct {
	store *store.Store
	*campaigns.Campaign

	// Cache the namespace on the resolver, since it's accessed more than once.
	namespaceOnce sync.Once
	namespace     graphqlbackend.NamespaceResolver
	namespaceErr  error
}

// TODO: We need to marshal the ID depending on how we landed here
const batchChangeIDKind = "BatchChange"

func marshalCampaignID(id int64) graphql.ID {
	return relay.MarshalID(batchChangeIDKind, id)
}

func unmarshalCampaignID(id graphql.ID) (campaignID int64, err error) {
	err = relay.UnmarshalSpec(id, &campaignID)
	return
}

func (r *batchChangeResolver) ID() graphql.ID {
	return marshalCampaignID(r.Campaign.ID)
}

func (r *batchChangeResolver) Name() string {
	return r.Campaign.Name
}

func (r *batchChangeResolver) Description() *string {
	if r.Campaign.Description == "" {
		return nil
	}
	return &r.Campaign.Description
}

func (r *batchChangeResolver) InitialApplier(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	user, err := graphqlbackend.UserByIDInt32(ctx, r.store.DB(), r.Campaign.InitialApplierID)
	if errcode.IsNotFound(err) {
		return nil, nil
	}
	return user, err
}

func (r *batchChangeResolver) LastApplier(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	user, err := graphqlbackend.UserByIDInt32(ctx, r.store.DB(), r.Campaign.LastApplierID)
	if errcode.IsNotFound(err) {
		return nil, nil
	}
	return user, err
}

func (r *batchChangeResolver) LastAppliedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.Campaign.LastAppliedAt}
}

func (r *batchChangeResolver) SpecCreator(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	spec, err := r.store.GetCampaignSpec(ctx, store.GetCampaignSpecOpts{
		ID: r.Campaign.CampaignSpecID,
	})
	if err != nil {
		return nil, err
	}
	user, err := graphqlbackend.UserByIDInt32(ctx, r.store.DB(), spec.UserID)
	if errcode.IsNotFound(err) {
		return nil, nil
	}
	return user, err
}

func (r *batchChangeResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	return checkSiteAdminOrSameUser(ctx, r.Campaign.InitialApplierID)
}

func (r *batchChangeResolver) URL(ctx context.Context) (string, error) {
	n, err := r.Namespace(ctx)
	if err != nil {
		return "", err
	}
	return campaignURL(n, r), nil
}

func (r *batchChangeResolver) Namespace(ctx context.Context) (graphqlbackend.NamespaceResolver, error) {
	return r.computeNamespace(ctx)
}

func (r *batchChangeResolver) computeNamespace(ctx context.Context) (graphqlbackend.NamespaceResolver, error) {
	r.namespaceOnce.Do(func() {
		if r.Campaign.NamespaceUserID != 0 {
			r.namespace.Namespace, r.namespaceErr = graphqlbackend.UserByIDInt32(
				ctx,
				r.store.DB(),
				r.Campaign.NamespaceUserID,
			)
		} else {
			r.namespace.Namespace, r.namespaceErr = graphqlbackend.OrgByIDInt32(
				ctx,
				r.store.DB(),
				r.Campaign.NamespaceOrgID,
			)
		}
		if errcode.IsNotFound(r.namespaceErr) {
			r.namespace.Namespace = nil
			r.namespaceErr = errors.New("namespace of campaign has been deleted")
		}
	})

	return r.namespace, r.namespaceErr
}

func (r *batchChangeResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.Campaign.CreatedAt}
}

func (r *batchChangeResolver) UpdatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.Campaign.UpdatedAt}
}

func (r *batchChangeResolver) ClosedAt() *graphqlbackend.DateTime {
	if !r.Campaign.Closed() {
		return nil
	}
	return &graphqlbackend.DateTime{Time: r.Campaign.ClosedAt}
}

func (r *batchChangeResolver) ChangesetsStats(ctx context.Context) (graphqlbackend.ChangesetsStatsResolver, error) {
	stats, err := r.store.GetChangesetsStats(ctx, store.GetChangesetsStatsOpts{
		CampaignID: r.Campaign.ID,
	})
	if err != nil {
		return nil, err
	}
	return &changesetsStatsResolver{stats: stats}, nil
}

func (r *batchChangeResolver) Changesets(
	ctx context.Context,
	args *graphqlbackend.ListChangesetsArgs,
) (graphqlbackend.ChangesetsConnectionResolver, error) {
	opts, safe, err := listChangesetOptsFromArgs(args, r.Campaign.ID)
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

func (r *batchChangeResolver) ChangesetCountsOverTime(
	ctx context.Context,
	args *graphqlbackend.ChangesetCountsArgs,
) ([]graphqlbackend.ChangesetCountsResolver, error) {
	resolvers := []graphqlbackend.ChangesetCountsResolver{}

	publishedState := campaigns.ChangesetPublicationStatePublished
	opts := store.ListChangesetsOpts{
		CampaignID: r.Campaign.ID,
		// Only load fully-synced changesets, so that the data we use for computing the changeset counts is complete.
		PublicationState: &publishedState,
	}
	cs, _, err := r.store.ListChangesets(ctx, opts)
	if err != nil {
		return resolvers, err
	}

	now := r.store.Clock()()

	weekAgo := now.Add(-7 * 24 * time.Hour)
	start := r.Campaign.CreatedAt.UTC()
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

func (r *batchChangeResolver) DiffStat(ctx context.Context) (*graphqlbackend.DiffStat, error) {
	diffStat, err := r.store.GetCampaignDiffStat(ctx, store.GetCampaignDiffStatOpts{CampaignID: r.Campaign.ID})
	if err != nil {
		return nil, err
	}
	return graphqlbackend.NewDiffStat(*diffStat), nil
}

func (r *batchChangeResolver) CurrentSpec(ctx context.Context) (graphqlbackend.CampaignSpecResolver, error) {
	campaignSpec, err := r.store.GetCampaignSpec(ctx, store.GetCampaignSpecOpts{ID: r.Campaign.CampaignSpecID})
	if err != nil {
		// This spec should always exist, so fail hard on not found errors as well.
		return nil, err
	}

	return &campaignSpecResolver{store: r.store, campaignSpec: campaignSpec}, nil
}
