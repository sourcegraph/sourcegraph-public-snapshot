package resolvers

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"

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

	once        sync.Once
	mappings    store.RewirerMappings
	allMappings store.RewirerMappings
	campaign    *campaigns.Campaign
	totalCount  int
	err         error
}

func (r *changesetApplyPreviewConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	_, _, _, totalCount, err := r.compute(ctx)
	if err != nil {
		return 0, err
	}
	return int32(totalCount), nil
}

func (r *changesetApplyPreviewConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if r.opts.LimitOffset == nil {
		return graphqlutil.HasNextPage(false), nil
	}
	_, _, _, totalCount, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if (r.opts.LimitOffset.Limit + r.opts.LimitOffset.Offset) >= totalCount {
		return graphqlutil.HasNextPage(false), nil
	}
	return graphqlutil.NextPageCursor(strconv.Itoa(r.opts.LimitOffset.Limit + r.opts.LimitOffset.Offset)), nil
}

func (r *changesetApplyPreviewConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ChangesetApplyPreviewResolver, error) {
	mappings, campaign, _, _, err := r.compute(ctx)
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

type changesetApplyPreviewConnectionStatsResolver struct {
	push         int32
	update       int32
	undraft      int32
	publish      int32
	publishDraft int32
	sync         int32
	_import      int32
	close        int32
	reopen       int32
	sleep        int32
	detach       int32
	unpublished  int32

	added    int32
	modified int32
	removed  int32
}

func (r *changesetApplyPreviewConnectionStatsResolver) Push() int32 {
	return r.push
}
func (r *changesetApplyPreviewConnectionStatsResolver) Update() int32 {
	return r.update
}
func (r *changesetApplyPreviewConnectionStatsResolver) Undraft() int32 {
	return r.undraft
}
func (r *changesetApplyPreviewConnectionStatsResolver) Publish() int32 {
	return r.publish
}
func (r *changesetApplyPreviewConnectionStatsResolver) PublishDraft() int32 {
	return r.publishDraft
}
func (r *changesetApplyPreviewConnectionStatsResolver) Sync() int32 {
	return r.sync
}
func (r *changesetApplyPreviewConnectionStatsResolver) Import() int32 {
	return r._import
}
func (r *changesetApplyPreviewConnectionStatsResolver) Close() int32 {
	return r.close
}
func (r *changesetApplyPreviewConnectionStatsResolver) Reopen() int32 {
	return r.reopen
}
func (r *changesetApplyPreviewConnectionStatsResolver) Sleep() int32 {
	return r.sleep
}
func (r *changesetApplyPreviewConnectionStatsResolver) Detach() int32 {
	return r.detach
}
func (r *changesetApplyPreviewConnectionStatsResolver) Added() int32 {
	return r.added
}
func (r *changesetApplyPreviewConnectionStatsResolver) Modified() int32 {
	return r.modified
}
func (r *changesetApplyPreviewConnectionStatsResolver) Removed() int32 {
	return r.removed
}

var _ graphqlbackend.ChangesetApplyPreviewConnectionStatsResolver = &changesetApplyPreviewConnectionStatsResolver{}

func (r *changesetApplyPreviewConnectionResolver) Stats(ctx context.Context) (graphqlbackend.ChangesetApplyPreviewConnectionStatsResolver, error) {
	_, campaign, mappings, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	stats := &changesetApplyPreviewConnectionStatsResolver{}
	for _, mapping := range mappings {
		res := &changesetApplyPreviewResolver{
			store:             r.store,
			mapping:           mapping,
			preloadedCampaign: campaign,
			campaignSpecID:    r.campaignSpecID,
		}
		var ops []campaigns.ReconcilerOperation
		if _, ok := res.ToHiddenChangesetApplyPreview(); ok {
			// Hidden ones never perform operations.
			continue
		}

		visRes, ok := res.ToVisibleChangesetApplyPreview()
		if !ok {
			return nil, errors.New("expected node to be a 'VisibleChangesetApplyPreview', but wasn't")
		}
		ops, err = visRes.Operations(ctx)
		if err != nil {
			return nil, err
		}
		targets := visRes.Targets()
		if _, ok := targets.ToVisibleApplyPreviewTargetsAttach(); ok {
			stats.added++
		}
		if _, ok := targets.ToVisibleApplyPreviewTargetsUpdate(); ok {
			if len(ops) > 0 {
				stats.modified++
			}
		}
		if _, ok := targets.ToVisibleApplyPreviewTargetsDetach(); ok {
			stats.removed++
		}
		for _, op := range ops {
			switch op {
			case campaigns.ReconcilerOperationPush:
				stats.push++
			case campaigns.ReconcilerOperationUpdate:
				stats.update++
			case campaigns.ReconcilerOperationUndraft:
				stats.undraft++
			case campaigns.ReconcilerOperationPublish:
				stats.publish++
			case campaigns.ReconcilerOperationPublishDraft:
				stats.publishDraft++
			case campaigns.ReconcilerOperationSync:
				stats.sync++
			case campaigns.ReconcilerOperationImport:
				stats._import++
			case campaigns.ReconcilerOperationClose:
				stats.close++
			case campaigns.ReconcilerOperationReopen:
				stats.reopen++
			case campaigns.ReconcilerOperationSleep:
				stats.sleep++
			case campaigns.ReconcilerOperationDetach:
				stats.detach++
			}
		}
	}

	return stats, nil
}

func (r *changesetApplyPreviewConnectionResolver) compute(ctx context.Context) (store.RewirerMappings, *campaigns.Campaign, store.RewirerMappings, int, error) {
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
		r.campaign, _, r.err = svc.ReconcileCampaign(ctx, campaignSpec)
		if r.err != nil {
			return
		}

		opts.CampaignID = r.campaign.ID

		r.mappings, r.err = r.store.GetRewirerMappings(ctx, opts)
		if r.err != nil {
			return
		}
		r.err = r.mappings.Hydrate(ctx, r.store)
		if r.err != nil {
			return
		}

		allOpts := store.GetRewirerMappingsOpts{
			CampaignSpecID: opts.CampaignSpecID,
			CampaignID:     opts.CampaignID,
			TextSearch:     opts.TextSearch,
		}
		r.allMappings, r.err = r.store.GetRewirerMappings(ctx, allOpts)
		if r.err != nil {
			return
		}
		r.err = r.allMappings.Hydrate(ctx, r.store)
		if r.err != nil {
			return
		}
		r.totalCount = len(r.allMappings)
	})

	return r.mappings, r.campaign, r.allMappings, r.totalCount, r.err
}
