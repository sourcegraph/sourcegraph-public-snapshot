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
	"github.com/sourcegraph/sourcegraph/internal/database"
)

var _ graphqlbackend.ChangesetApplyPreviewConnectionResolver = &changesetApplyPreviewConnectionResolver{}

type changesetApplyPreviewConnectionResolver struct {
	store *store.Store

	opts           store.GetRewirerMappingsOpts
	action         *campaigns.ReconcilerOperation
	campaignSpecID int64

	once     sync.Once
	mappings *rewirerMappings
	err      error
}

// rewirerMappings wraps store.RewirerMappings to provide memoised pagination
// and filtering functionality.
type rewirerMappings struct {
	All store.RewirerMappings

	// Inputs from outside the resolver that we need to build other resolvers.
	campaignSpecID int64
	store          *store.Store

	// Calculated within the type with the dry run.
	campaign *campaigns.Campaign

	// Cache of filtered pages.
	pagesMu sync.Mutex
	pages   map[rewirerMappingPageOpts]*rewirerMappingPage

	// Cache of rewirer mapping resolvers.
	resolversMu sync.Mutex
	resolvers   map[*store.RewirerMapping]*changesetApplyPreviewResolver
}

// newRewirerMappings creates a new rewirer mappings object, which includes dry
// running the campaign reconciliation.
func newRewirerMappings(ctx context.Context, s *store.Store, opts store.GetRewirerMappingsOpts, campaignSpecID int64) (*rewirerMappings, error) {
	rm := &rewirerMappings{
		campaignSpecID: campaignSpecID,
		store:          s,
		pages:          make(map[rewirerMappingPageOpts]*rewirerMappingPage),
		resolvers:      make(map[*store.RewirerMapping]*changesetApplyPreviewResolver),
	}

	svc := service.New(rm.store)
	campaignSpec, err := rm.store.GetCampaignSpec(ctx, store.GetCampaignSpecOpts{ID: campaignSpecID})
	if err != nil {
		return nil, err
	}
	// Dry-run reconcile the campaign with the new campaign spec.
	if rm.campaign, _, err = svc.ReconcileCampaign(ctx, campaignSpec); err != nil {
		return nil, err
	}

	opts = store.GetRewirerMappingsOpts{
		CampaignSpecID: campaignSpecID,
		CampaignID:     rm.campaign.ID,
		TextSearch:     opts.TextSearch,
		CurrentState:   opts.CurrentState,
	}
	if rm.All, err = rm.store.GetRewirerMappings(ctx, opts); err != nil {
		return nil, err
	}
	if err := rm.All.Hydrate(ctx, rm.store); err != nil {
		return nil, err
	}

	return rm, nil
}

type rewirerMappingPageOpts struct {
	*database.LimitOffset
	Op *campaigns.ReconcilerOperation
}

type rewirerMappingPage struct {
	Mappings store.RewirerMappings

	// TotalCount represents the total count of filtered results, but not
	// necessarily the full set of results.
	TotalCount int
}

// Page applies the given filter, and paginates the results.
func (rw *rewirerMappings) Page(ctx context.Context, opts rewirerMappingPageOpts) (*rewirerMappingPage, error) {
	rw.pagesMu.Lock()
	defer rw.pagesMu.Unlock()

	if page := rw.pages[opts]; page != nil {
		return page, nil
	}

	var filtered store.RewirerMappings
	if opts.Op != nil {
		for _, mapping := range rw.All {
			res, ok := rw.Resolver(mapping).ToVisibleChangesetApplyPreview()
			if !ok {
				continue
			}

			ops, err := res.Operations(ctx)
			if err != nil {
				return nil, err
			}

			for _, op := range ops {
				if op == *opts.Op {
					filtered = append(filtered, mapping)
					break
				}
			}
		}
	} else {
		filtered = rw.All
	}

	page := store.RewirerMappings{}
	if limit, offset := opts.LimitOffset.Limit, opts.LimitOffset.Offset; limit <= 0 || offset < 0 || offset > len(filtered) {
		// The limit and/or offset are outside the possible bounds, so we have
		// nothing to do here: the empty mappings slice is correct.
	} else {
		if end := limit + offset; end > len(filtered) {
			page = filtered[offset:]
		} else {
			page = filtered[offset:end]
		}
	}

	rw.pages[opts] = &rewirerMappingPage{
		Mappings:   page,
		TotalCount: len(filtered),
	}
	return rw.pages[opts], nil
}

func (rw *rewirerMappings) Resolver(mapping *store.RewirerMapping) *changesetApplyPreviewResolver {
	rw.resolversMu.Lock()
	defer rw.resolversMu.Unlock()

	if resolver := rw.resolvers[mapping]; resolver != nil {
		return resolver
	}

	// We build the resolver without a preloadedNextSync, since not all callers
	// will have calculated that.
	rw.resolvers[mapping] = &changesetApplyPreviewResolver{
		store:             rw.store,
		mapping:           mapping,
		preloadedCampaign: rw.campaign,
		campaignSpecID:    rw.campaignSpecID,
	}
	return rw.resolvers[mapping]
}

func (rw *rewirerMappings) ResolverWithNextSync(mapping *store.RewirerMapping, nextSync time.Time) graphqlbackend.ChangesetApplyPreviewResolver {
	// As the apply target resolvers don't cache the preloaded next sync value
	// when creating the changeset resolver, we can shallow copy and update the
	// field rather than having to build a whole new resolver.
	resolver := *rw.Resolver(mapping)
	resolver.preloadedNextSync = nextSync

	return &resolver
}

func (r *changesetApplyPreviewConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	mappings, err := r.compute(ctx)
	if err != nil {
		return 0, err
	}

	page, err := mappings.Page(ctx, rewirerMappingPageOpts{
		LimitOffset: r.opts.LimitOffset,
		Op:          r.action,
	})
	if err != nil {
		return 0, err
	}

	return int32(page.TotalCount), nil
}

func (r *changesetApplyPreviewConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if r.opts.LimitOffset == nil {
		return graphqlutil.HasNextPage(false), nil
	}
	mappings, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if (r.opts.LimitOffset.Limit + r.opts.LimitOffset.Offset) >= len(mappings.All) {
		return graphqlutil.HasNextPage(false), nil
	}
	return graphqlutil.NextPageCursor(strconv.Itoa(r.opts.LimitOffset.Limit + r.opts.LimitOffset.Offset)), nil
}

func (r *changesetApplyPreviewConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ChangesetApplyPreviewResolver, error) {
	mappings, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	page, err := mappings.Page(ctx, rewirerMappingPageOpts{
		LimitOffset: r.opts.LimitOffset,
		Op:          r.action,
	})
	if err != nil {
		return nil, err
	}

	scheduledSyncs := make(map[int64]time.Time)
	changesetIDs := page.Mappings.ChangesetIDs()
	if len(changesetIDs) > 0 {
		syncData, err := r.store.ListChangesetSyncData(ctx, store.ListChangesetSyncDataOpts{ChangesetIDs: changesetIDs})
		if err != nil {
			return nil, err
		}
		for _, d := range syncData {
			scheduledSyncs[d.ChangesetID] = syncer.NextSync(time.Now, d)
		}
	}

	resolvers := make([]graphqlbackend.ChangesetApplyPreviewResolver, 0, len(page.Mappings))
	for _, mapping := range page.Mappings {
		resolvers = append(resolvers, mappings.ResolverWithNextSync(mapping, scheduledSyncs[mapping.ChangesetID]))
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
	mappings, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	stats := &changesetApplyPreviewConnectionStatsResolver{}
	for _, mapping := range mappings.All {
		res := mappings.Resolver(mapping)
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

func (r *changesetApplyPreviewConnectionResolver) compute(ctx context.Context) (*rewirerMappings, error) {
	r.once.Do(func() {
		r.mappings, r.err = newRewirerMappings(ctx, r.store, r.opts, r.campaignSpecID)
	})

	return r.mappings, r.err
}
