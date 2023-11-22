package resolvers

import (
	"context"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/gitserver"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	bgql "github.com/sourcegraph/sourcegraph/internal/batches/graphql"
	"github.com/sourcegraph/sourcegraph/internal/batches/service"
	"github.com/sourcegraph/sourcegraph/internal/batches/state"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ graphqlbackend.BatchChangeResolver = &batchChangeResolver{}

type batchChangeResolver struct {
	store           *store.Store
	gitserverClient gitserver.Client
	logger          log.Logger

	batchChange *btypes.BatchChange

	// Cache the namespace on the resolver, since it's accessed more than once.
	namespaceOnce sync.Once
	namespace     graphqlbackend.NamespaceResolver
	namespaceErr  error

	batchSpecOnce sync.Once
	batchSpec     *btypes.BatchSpec
	batchSpecErr  error

	canAdministerOnce sync.Once
	canAdminister     bool
	canAdministerErr  error
}

const batchChangeIDKind = "BatchChange"

func unmarshalBatchChangeID(id graphql.ID) (batchChangeID int64, err error) {
	err = relay.UnmarshalSpec(id, &batchChangeID)
	return
}

func (r *batchChangeResolver) ID() graphql.ID {
	return bgql.MarshalBatchChangeID(r.batchChange.ID)
}

func (r *batchChangeResolver) Name() string {
	return r.batchChange.Name
}

func (r *batchChangeResolver) Description() *string {
	if r.batchChange.Description == "" {
		return nil
	}
	return &r.batchChange.Description
}

func (r *batchChangeResolver) State() string {
	var batchChangeState btypes.BatchChangeState
	if r.batchChange.Closed() {
		batchChangeState = btypes.BatchChangeStateClosed
	} else if r.batchChange.IsDraft() {
		batchChangeState = btypes.BatchChangeStateDraft
	} else {
		batchChangeState = btypes.BatchChangeStateOpen
	}

	return batchChangeState.ToGraphQL()
}

func (r *batchChangeResolver) Creator(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	user, err := graphqlbackend.UserByIDInt32(ctx, r.store.DatabaseDB(), r.batchChange.CreatorID)
	if errcode.IsNotFound(err) {
		return nil, nil
	}
	return user, err
}

func (r *batchChangeResolver) LastApplier(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	if r.batchChange.LastApplierID == 0 {
		return nil, nil
	}

	user, err := graphqlbackend.UserByIDInt32(ctx, r.store.DatabaseDB(), r.batchChange.LastApplierID)
	if errcode.IsNotFound(err) {
		return nil, nil
	}

	return user, err
}

func (r *batchChangeResolver) LastAppliedAt() *gqlutil.DateTime {
	if r.batchChange.LastAppliedAt.IsZero() {
		return nil
	}

	return &gqlutil.DateTime{Time: r.batchChange.LastAppliedAt}
}

func (r *batchChangeResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	r.canAdministerOnce.Do(func() {
		svc := service.New(r.store)
		r.canAdminister, r.canAdministerErr = svc.CheckViewerCanAdminister(ctx, r.batchChange.NamespaceUserID, r.batchChange.NamespaceOrgID)
	})
	return r.canAdminister, r.canAdministerErr
}

func (r *batchChangeResolver) URL(ctx context.Context) (string, error) {
	n, err := r.Namespace(ctx)
	if err != nil {
		return "", err
	}
	return batchChangeURL(n, r), nil
}

func (r *batchChangeResolver) Namespace(ctx context.Context) (graphqlbackend.NamespaceResolver, error) {
	return r.computeNamespace(ctx)
}

func (r *batchChangeResolver) computeNamespace(ctx context.Context) (graphqlbackend.NamespaceResolver, error) {
	r.namespaceOnce.Do(func() {
		if r.batchChange.NamespaceUserID != 0 {
			r.namespace.Namespace, r.namespaceErr = graphqlbackend.UserByIDInt32(
				ctx,
				r.store.DatabaseDB(),
				r.batchChange.NamespaceUserID,
			)
		} else {
			r.namespace.Namespace, r.namespaceErr = graphqlbackend.OrgByIDInt32(
				ctx,
				r.store.DatabaseDB(),
				r.batchChange.NamespaceOrgID,
			)
		}
		if errcode.IsNotFound(r.namespaceErr) {
			r.namespace.Namespace = nil
			r.namespaceErr = errors.New("namespace of batch change has been deleted")
		}
	})

	return r.namespace, r.namespaceErr
}

func (r *batchChangeResolver) computeBatchSpec(ctx context.Context) (*btypes.BatchSpec, error) {
	r.batchSpecOnce.Do(func() {
		r.batchSpec, r.batchSpecErr = r.store.GetBatchSpec(ctx, store.GetBatchSpecOpts{
			ID: r.batchChange.BatchSpecID,
		})
	})

	return r.batchSpec, r.batchSpecErr
}

func (r *batchChangeResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.batchChange.CreatedAt}
}

func (r *batchChangeResolver) UpdatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.batchChange.UpdatedAt}
}

func (r *batchChangeResolver) ClosedAt() *gqlutil.DateTime {
	if !r.batchChange.Closed() {
		return nil
	}
	return &gqlutil.DateTime{Time: r.batchChange.ClosedAt}
}

func (r *batchChangeResolver) ChangesetsStats(ctx context.Context) (graphqlbackend.ChangesetsStatsResolver, error) {
	stats, err := r.store.GetChangesetsStats(ctx, r.batchChange.ID)
	if err != nil {
		return nil, err
	}
	return &changesetsStatsResolver{stats: stats}, nil
}

func (r *batchChangeResolver) Changesets(
	ctx context.Context,
	args *graphqlbackend.ListChangesetsArgs,
) (graphqlbackend.ChangesetsConnectionResolver, error) {
	opts, safe, err := listChangesetOptsFromArgs(args, r.batchChange.ID)
	if err != nil {
		return nil, err
	}
	opts.BatchChangeID = r.batchChange.ID
	return &changesetsConnectionResolver{
		store:           r.store,
		gitserverClient: r.gitserverClient,
		logger:          r.logger,
		opts:            opts,
		optsSafe:        safe,
	}, nil
}

func (r *batchChangeResolver) ChangesetCountsOverTime(
	ctx context.Context,
	args *graphqlbackend.ChangesetCountsArgs,
) ([]graphqlbackend.ChangesetCountsResolver, error) {
	publishedState := btypes.ChangesetPublicationStatePublished
	opts := store.ListChangesetsOpts{
		BatchChangeID:   r.batchChange.ID,
		IncludeArchived: args.IncludeArchived,
		// Only load fully-synced changesets, so that the data we use for computing the changeset counts is complete.
		PublicationState: &publishedState,
	}
	cs, _, err := r.store.ListChangesets(ctx, opts)
	if err != nil {
		return nil, err
	}

	var es []*btypes.ChangesetEvent
	changesetIDs := cs.IDs()
	if len(changesetIDs) > 0 {
		eventsOpts := store.ListChangesetEventsOpts{ChangesetIDs: changesetIDs, Kinds: state.RequiredEventTypesForHistory}
		es, _, err = r.store.ListChangesetEvents(ctx, eventsOpts)
		if err != nil {
			return nil, err
		}
	}
	// Sort all events once by their timestamps, CalcCounts depends on it.
	events := state.ChangesetEvents(es)
	sort.Sort(events)

	// Determine timeframe.
	now := r.store.Clock()()
	weekAgo := now.Add(-7 * 24 * time.Hour)
	start := r.batchChange.CreatedAt.UTC()
	if len(events) > 0 {
		start = events[0].Timestamp().UTC()
	}
	// At least a week lookback, more if the batch change was created earlier.
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

	counts, err := state.CalcCounts(start, end, cs, es...)
	if err != nil {
		return nil, err
	}

	resolvers := make([]graphqlbackend.ChangesetCountsResolver, 0, len(counts))
	for _, c := range counts {
		resolvers = append(resolvers, &changesetCountsResolver{counts: c})
	}

	return resolvers, nil
}

func (r *batchChangeResolver) DiffStat(ctx context.Context) (*graphqlbackend.DiffStat, error) {
	diffStat, err := r.store.GetBatchChangeDiffStat(ctx, store.GetBatchChangeDiffStatOpts{BatchChangeID: r.batchChange.ID})
	if err != nil {
		return nil, err
	}
	return graphqlbackend.NewDiffStat(*diffStat), nil
}

func (r *batchChangeResolver) CurrentSpec(ctx context.Context) (graphqlbackend.BatchSpecResolver, error) {
	batchSpec, err := r.computeBatchSpec(ctx)
	if err != nil {
		// This spec should always exist, so fail hard on not found errors as well.
		return nil, err
	}

	return &batchSpecResolver{store: r.store, batchSpec: batchSpec, logger: r.logger}, nil
}

func (r *batchChangeResolver) BulkOperations(
	ctx context.Context,
	args *graphqlbackend.ListBatchChangeBulkOperationArgs,
) (graphqlbackend.BulkOperationConnectionResolver, error) {
	if err := validateFirstParamDefaults(args.First); err != nil {
		return nil, err
	}
	opts := store.ListBulkOperationsOpts{
		LimitOpts: store.LimitOpts{
			Limit: int(args.First),
		},
	}
	if args.After != nil {
		id, err := strconv.Atoi(*args.After)
		if err != nil {
			return nil, err
		}
		opts.Cursor = int64(id)
	}

	if args.CreatedAfter != nil {
		opts.CreatedAfter = args.CreatedAfter.Time
	}

	return &bulkOperationConnectionResolver{
		store:           r.store,
		gitserverClient: r.gitserverClient,
		batchChangeID:   r.batchChange.ID,
		opts:            opts,
		logger:          r.logger,
	}, nil
}

func (r *batchChangeResolver) BatchSpecs(
	ctx context.Context,
	args *graphqlbackend.ListBatchSpecArgs,
) (graphqlbackend.BatchSpecConnectionResolver, error) {
	if err := validateFirstParamDefaults(args.First); err != nil {
		return nil, err
	}
	opts := store.ListBatchSpecsOpts{
		BatchChangeID: r.batchChange.ID,
		LimitOpts: store.LimitOpts{
			Limit: int(args.First),
		},
		// We want the batch spec connection to always show the latest one first.
		NewestFirst: true,
	}

	if args.IncludeLocallyExecutedSpecs != nil {
		opts.IncludeLocallyExecutedSpecs = *args.IncludeLocallyExecutedSpecs
	}

	if args.ExcludeEmptySpecs != nil {
		opts.ExcludeEmptySpecs = *args.ExcludeEmptySpecs
	}

	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.store.DatabaseDB()); err != nil {
		opts.ExcludeCreatedFromRawNotOwnedByUser = actor.FromContext(ctx).UID
	}

	if args.After != nil {
		id, err := strconv.Atoi(*args.After)
		if err != nil {
			return nil, err
		}
		opts.Cursor = int64(id)
	}

	return &batchSpecConnectionResolver{store: r.store, logger: r.logger, opts: opts}, nil
}
