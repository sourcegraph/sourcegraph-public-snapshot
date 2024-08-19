package resolvers

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/batches/service"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/batches/syncer"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ graphqlbackend.ChangesetApplyPreviewConnectionResolver = &changesetApplyPreviewConnectionResolver{}

type changesetApplyPreviewConnectionResolver struct {
	store           *store.Store
	gitserverClient gitserver.Client
	logger          log.Logger

	opts              store.GetRewirerMappingsOpts
	action            *btypes.ReconcilerOperation
	batchSpecID       int64
	publicationStates publicationStateMap

	once     sync.Once
	mappings *rewirerMappingsFacade
	err      error
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

func (r *changesetApplyPreviewConnectionResolver) PageInfo(ctx context.Context) (*gqlutil.PageInfo, error) {
	if r.opts.LimitOffset == nil {
		return gqlutil.HasNextPage(false), nil
	}
	mappings, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if (r.opts.LimitOffset.Limit + r.opts.LimitOffset.Offset) >= len(mappings.All) {
		return gqlutil.HasNextPage(false), nil
	}
	return gqlutil.NextPageCursor(strconv.Itoa(r.opts.LimitOffset.Limit + r.opts.LimitOffset.Offset)), nil
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
			scheduledSyncs[d.ChangesetID] = syncer.NextSync(r.store.Clock(), d)
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
	archive      int32
	reattach     int32

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
func (r *changesetApplyPreviewConnectionStatsResolver) Archive() int32 {
	return r.archive
}
func (r *changesetApplyPreviewConnectionStatsResolver) Reattach() int32 {
	return r.reattach
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
		var ops []string
		if _, ok := res.ToHiddenChangesetApplyPreview(); ok {
			// Hidden ones never perform operations.
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
			if len(ops) > 0 && len(mapping.Changeset.BatchChanges) > 0 {
				stats.modified++
			} else if len(mapping.Changeset.BatchChanges) == 0 {
				stats.added++
			}
		}
		if _, ok := targets.ToVisibleApplyPreviewTargetsDetach(); ok {
			stats.removed++
		}
		for _, op := range ops {
			switch op {
			case string(btypes.ReconcilerOperationPush):
				stats.push++
			case string(btypes.ReconcilerOperationUpdate):
				stats.update++
			case string(btypes.ReconcilerOperationUndraft):
				stats.undraft++
			case string(btypes.ReconcilerOperationPublish):
				stats.publish++
			case string(btypes.ReconcilerOperationPublishDraft):
				stats.publishDraft++
			case string(btypes.ReconcilerOperationSync):
				stats.sync++
			case string(btypes.ReconcilerOperationImport):
				stats._import++
			case string(btypes.ReconcilerOperationClose):
				stats.close++
			case string(btypes.ReconcilerOperationReopen):
				stats.reopen++
			case string(btypes.ReconcilerOperationSleep):
				stats.sleep++
			case string(btypes.ReconcilerOperationDetach):
				stats.detach++
			case string(btypes.ReconcilerOperationArchive):
				stats.archive++
			case string(btypes.ReconcilerOperationReattach):
				stats.reattach++
			}
		}
	}

	return stats, nil
}

func (r *changesetApplyPreviewConnectionResolver) compute(ctx context.Context) (*rewirerMappingsFacade, error) {
	r.once.Do(func() {
		r.mappings = newRewirerMappingsFacade(r.store, r.gitserverClient, r.logger, r.batchSpecID, r.publicationStates)
		r.err = r.mappings.compute(ctx, r.opts)
	})

	return r.mappings, r.err
}

// rewirerMappingsFacade wraps btypes.RewirerMappings to provide memoised pagination
// and filtering functionality.
type rewirerMappingsFacade struct {
	All btypes.RewirerMappings

	// Inputs from outside the resolver that we need to build other resolvers.
	batchSpecID       int64
	publicationStates publicationStateMap
	store             *store.Store
	gitserverClient   gitserver.Client
	logger            log.Logger

	// This field is set when ReconcileBatchChange is called.
	batchChange *btypes.BatchChange

	// Cache of filtered pages.
	pagesMu sync.Mutex
	pages   map[rewirerMappingPageOpts]*rewirerMappingPage

	// Cache of rewirer mapping resolvers.
	resolversMu sync.Mutex
	resolvers   map[*btypes.RewirerMapping]graphqlbackend.ChangesetApplyPreviewResolver
}

// newRewirerMappingsFacade creates a new rewirer mappings object, which
// includes dry running the batch change reconciliation.
func newRewirerMappingsFacade(s *store.Store, gitserverClient gitserver.Client, logger log.Logger, batchSpecID int64, publicationStates publicationStateMap) *rewirerMappingsFacade {
	return &rewirerMappingsFacade{
		batchSpecID:       batchSpecID,
		publicationStates: publicationStates,
		store:             s,
		logger:            logger,
		gitserverClient:   gitserverClient,
		pages:             make(map[rewirerMappingPageOpts]*rewirerMappingPage),
		resolvers:         make(map[*btypes.RewirerMapping]graphqlbackend.ChangesetApplyPreviewResolver),
	}
}

func (rmf *rewirerMappingsFacade) compute(ctx context.Context, opts store.GetRewirerMappingsOpts) error {
	svc := service.New(rmf.store)
	batchSpec, err := rmf.store.GetBatchSpec(ctx, store.GetBatchSpecOpts{ID: rmf.batchSpecID})
	if err != nil {
		return err
	}
	// Dry-run reconcile the batch change with the new batch spec.
	if rmf.batchChange, _, err = svc.ReconcileBatchChange(ctx, batchSpec); err != nil {
		return err
	}

	opts = store.GetRewirerMappingsOpts{
		BatchSpecID:   rmf.batchSpecID,
		BatchChangeID: rmf.batchChange.ID,
		TextSearch:    opts.TextSearch,
		CurrentState:  opts.CurrentState,
	}
	rmf.All, err = rmf.store.GetRewirerMappings(ctx, opts)
	return err
}

type rewirerMappingPageOpts struct {
	*database.LimitOffset
	Op *btypes.ReconcilerOperation
}

type rewirerMappingPage struct {
	Mappings btypes.RewirerMappings

	// TotalCount represents the total count of filtered results, but not
	// necessarily the full set of results.
	TotalCount int
}

// Page applies the given filter, and paginates the results.
func (rmf *rewirerMappingsFacade) Page(ctx context.Context, opts rewirerMappingPageOpts) (*rewirerMappingPage, error) {
	rmf.pagesMu.Lock()
	defer rmf.pagesMu.Unlock()

	if page := rmf.pages[opts]; page != nil {
		return page, nil
	}

	var filtered btypes.RewirerMappings
	if opts.Op != nil {
		filtered = btypes.RewirerMappings{}
		for _, mapping := range rmf.All {
			res, ok := rmf.Resolver(mapping).ToVisibleChangesetApplyPreview()
			if !ok {
				continue
			}

			ops, err := res.Operations(ctx)
			if err != nil {
				return nil, err
			}

			for _, op := range ops {
				if op == string(*opts.Op) {
					filtered = append(filtered, mapping)
					break
				}
			}
		}
	} else {
		filtered = rmf.All
	}

	var page btypes.RewirerMappings
	if lo := opts.LimitOffset; lo != nil {
		if limit, offset := lo.Limit, lo.Offset; limit < 0 || offset < 0 || offset > len(filtered) {
			// The limit and/or offset are outside the possible bounds, so we
			// just need to make the slice not nil.
			page = btypes.RewirerMappings{}
		} else if limit == 0 {
			page = filtered[offset:]
		} else {
			if end := limit + offset; end > len(filtered) {
				page = filtered[offset:]
			} else {
				page = filtered[offset:end]
			}
		}
	} else {
		page = filtered
	}

	rmf.pages[opts] = &rewirerMappingPage{
		Mappings:   page,
		TotalCount: len(filtered),
	}
	return rmf.pages[opts], nil
}

func (rmf *rewirerMappingsFacade) Resolver(mapping *btypes.RewirerMapping) graphqlbackend.ChangesetApplyPreviewResolver {
	rmf.resolversMu.Lock()
	defer rmf.resolversMu.Unlock()

	if resolver := rmf.resolvers[mapping]; resolver != nil {
		return resolver
	}

	// We build the resolver without a preloadedNextSync, since not all callers
	// will have calculated that.
	rmf.resolvers[mapping] = &changesetApplyPreviewResolver{
		store:                rmf.store,
		gitserverClient:      rmf.gitserverClient,
		logger:               rmf.logger,
		mapping:              mapping,
		preloadedBatchChange: rmf.batchChange,
		batchSpecID:          rmf.batchSpecID,
		publicationStates:    rmf.publicationStates,
	}
	return rmf.resolvers[mapping]
}

func (rmf *rewirerMappingsFacade) ResolverWithNextSync(mapping *btypes.RewirerMapping, nextSync time.Time) graphqlbackend.ChangesetApplyPreviewResolver {
	// As the apply target resolvers don't cache the preloaded next sync value
	// when creating the changeset resolver, we can shallow copy and update the
	// field rather than having to build a whole new resolver.
	//
	// Since objects can only end up in the resolvers map via Resolver(), it's
	// safe to type-assert to *changesetApplyPreviewResolver here.
	resolver := *rmf.Resolver(mapping).(*changesetApplyPreviewResolver)
	resolver.preloadedNextSync = nextSync

	return &resolver
}

// publicationStateMap maps changeset specs (by random ID) to their desired UI
// publication state.
type publicationStateMap map[string]batches.PublishedValue

// newPublicationStateMap creates a new publicationStateMap from the given
// publication state input, validating that there are no duplicates or invalid
// changeset spec GraphQL IDs.
func newPublicationStateMap(in *[]graphqlbackend.ChangesetSpecPublicationStateInput) (publicationStateMap, error) {
	out := publicationStateMap{}
	if in != nil {
		var errs error
		for _, ps := range *in {
			id, err := unmarshalChangesetSpecID(ps.ChangesetSpec)
			if err != nil {
				errs = errors.Append(errs, errors.Wrapf(err, "malformed changeset spec ID %q", string(ps.ChangesetSpec)))
				continue
			}

			if _, ok := out[id]; ok {
				errs = errors.Append(errs, errors.Newf("duplicate changeset spec ID %q", string(ps.ChangesetSpec)))
				continue
			}
			out[id] = ps.PublicationState
		}
		if errs != nil {
			return nil, errs
		}
	}

	return out, nil
}
