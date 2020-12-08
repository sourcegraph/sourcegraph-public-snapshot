package resolvers

import (
	"context"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/service"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/syncer"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

var _ graphqlbackend.ChangesetApplyPreviewConnectionResolver = &changesetApplyPreviewConnectionResolver{}

type changesetApplyPreviewConnectionResolver struct {
	store *store.Store

	opts           store.GetRewirerMappingsOpts
	campaignSpecID int64

	once     sync.Once
	mappings store.RewirerMappings
	campaign *campaigns.Campaign
	err      error
}

func (r *changesetApplyPreviewConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	mappings, _, err := r.compute(ctx)
	if err != nil {
		return 0, err
	}
	return int32(len(mappings)), nil
}

func (r *changesetApplyPreviewConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}

func (r *changesetApplyPreviewConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ChangesetApplyPreviewResolver, error) {
	mappings, campaign, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	syncData, err := r.store.ListChangesetSyncData(ctx, store.ListChangesetSyncDataOpts{ChangesetIDs: mappings.ChangesetIDs()})
	if err != nil {
		return nil, err
	}
	scheduledSyncs := make(map[int64]time.Time)
	for _, d := range syncData {
		scheduledSyncs[d.ChangesetID] = syncer.NextSync(time.Now, d)
	}

	resolvers := make([]graphqlbackend.ChangesetApplyPreviewResolver, 0, len(mappings))
	for _, mapping := range mappings {
		resolvers = append(resolvers, &changesetApplyPreviewResolver{
			store:             r.store,
			mapping:           mapping,
			preloadedNextSync: scheduledSyncs[mapping.ChangesetID],
			preloadedCampaign: campaign,
		})
	}

	return resolvers, nil
}

func (r *changesetApplyPreviewConnectionResolver) compute(ctx context.Context) (store.RewirerMappings, *campaigns.Campaign, error) {
	r.once.Do(func() {
		opts := r.opts
		opts.CampaignSpecID = r.campaignSpecID

		svc := service.New(r.store)
		campaignSpec, err := r.store.GetCampaignSpec(ctx, store.GetCampaignSpecOpts{ID: r.campaignSpecID})
		if err != nil {
			r.err = err
			return
		}
		// Dry-run reconcile the campaign with the new campaign spec.
		r.campaign, _, err = svc.ReconcileCampaign(ctx, campaignSpec)
		if err != nil {
			r.err = err
			return
		}

		opts.CampaignID = r.campaign.ID

		r.mappings, r.err = r.store.GetRewirerMappings(ctx, opts)
		if r.err != nil {
			return
		}
		r.err = r.mappings.Hydrate(ctx, r.store)
	})

	return r.mappings, r.campaign, r.err
}
