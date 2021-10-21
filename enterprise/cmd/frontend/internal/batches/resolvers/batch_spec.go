package resolvers

import (
	"context"
	"strconv"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/search"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/service"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/lib/batches"
)

const batchSpecIDKind = "BatchSpec"

func marshalBatchSpecRandID(id string) graphql.ID {
	return relay.MarshalID(batchSpecIDKind, id)
}

func unmarshalBatchSpecID(id graphql.ID) (batchSpecRandID string, err error) {
	err = relay.UnmarshalSpec(id, &batchSpecRandID)
	return
}

var _ graphqlbackend.BatchSpecResolver = &batchSpecResolver{}

type batchSpecResolver struct {
	store *store.Store

	batchSpec          *btypes.BatchSpec
	preloadedNamespace *graphqlbackend.NamespaceResolver

	// We cache the namespace on the resolver, since it's accessed more than once.
	namespaceOnce sync.Once
	namespace     *graphqlbackend.NamespaceResolver
	namespaceErr  error

	resolutionOnce sync.Once
	resolution     *btypes.BatchSpecResolutionJob
	resolutionErr  error

	workspacesOnce sync.Once
	workspaces     []*btypes.BatchSpecWorkspace
	workspacesErr  error

	validateSpecsOnce sync.Once
	validateSpecsErr  error

	statsOnce sync.Once
	stats     btypes.BatchSpecStats
	statsErr  error

	stateOnce sync.Once
	state     btypes.BatchSpecState
	stateErr  error
}

func (r *batchSpecResolver) ID() graphql.ID {
	// 🚨 SECURITY: This needs to be the RandID! We can't expose the
	// sequential, guessable ID.
	return marshalBatchSpecRandID(r.batchSpec.RandID)
}

func (r *batchSpecResolver) OriginalInput() (string, error) {
	return r.batchSpec.RawSpec, nil
}

func (r *batchSpecResolver) ParsedInput() (graphqlbackend.JSONValue, error) {
	return graphqlbackend.JSONValue{Value: r.batchSpec.Spec}, nil
}

func (r *batchSpecResolver) ChangesetSpecs(ctx context.Context, args *graphqlbackend.ChangesetSpecsConnectionArgs) (graphqlbackend.ChangesetSpecConnectionResolver, error) {
	opts := store.ListChangesetSpecsOpts{
		BatchSpecID: r.batchSpec.ID,
	}
	if err := validateFirstParamDefaults(args.First); err != nil {
		return nil, err
	}
	opts.Limit = int(args.First)
	if args.After != nil {
		id, err := strconv.Atoi(*args.After)
		if err != nil {
			return nil, err
		}
		opts.Cursor = int64(id)
	}

	return &changesetSpecConnectionResolver{
		store: r.store,
		opts:  opts,
	}, nil
}

func (r *batchSpecResolver) ApplyPreview(ctx context.Context, args *graphqlbackend.ChangesetApplyPreviewConnectionArgs) (graphqlbackend.ChangesetApplyPreviewConnectionResolver, error) {
	if args.CurrentState != nil {
		if !btypes.ChangesetState(*args.CurrentState).Valid() {
			return nil, errors.Errorf("invalid currentState %q", *args.CurrentState)
		}
	}
	if err := validateFirstParamDefaults(args.First); err != nil {
		return nil, err
	}
	opts := store.GetRewirerMappingsOpts{
		LimitOffset: &database.LimitOffset{
			Limit: int(args.First),
		},
		CurrentState: (*btypes.ChangesetState)(args.CurrentState),
	}
	if args.After != nil {
		id, err := strconv.Atoi(*args.After)
		if err != nil {
			return nil, err
		}
		opts.LimitOffset.Offset = id
	}
	if args.Search != nil {
		var err error
		opts.TextSearch, err = search.ParseTextSearch(*args.Search)
		if err != nil {
			return nil, errors.Wrap(err, "parsing search")
		}
	}
	if args.Action != nil {
		if !btypes.ReconcilerOperation(*args.Action).Valid() {
			return nil, errors.Errorf("invalid action %q", *args.Action)
		}
	}
	publicationStates, err := newPublicationStateMap(args.PublicationStates)
	if err != nil {
		return nil, err
	}

	return &changesetApplyPreviewConnectionResolver{
		store:             r.store,
		opts:              opts,
		action:            (*btypes.ReconcilerOperation)(args.Action),
		batchSpecID:       r.batchSpec.ID,
		publicationStates: publicationStates,
	}, nil
}

func (r *batchSpecResolver) Description() graphqlbackend.BatchChangeDescriptionResolver {
	return &batchChangeDescriptionResolver{
		name:        r.batchSpec.Spec.Name,
		description: r.batchSpec.Spec.Description,
	}
}

func (r *batchSpecResolver) Creator(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	user, err := graphqlbackend.UserByIDInt32(ctx, r.store.DB(), r.batchSpec.UserID)
	if errcode.IsNotFound(err) {
		return nil, nil
	}
	return user, err
}

func (r *batchSpecResolver) Namespace(ctx context.Context) (*graphqlbackend.NamespaceResolver, error) {
	return r.computeNamespace(ctx)
}

func (r *batchSpecResolver) ApplyURL(ctx context.Context) (*string, error) {
	state, err := r.computeState(ctx)
	if err != nil {
		return nil, err
	}

	if r.batchSpec.CreatedFromRaw && state != btypes.BatchSpecStateCompleted {
		return nil, nil
	}

	n, err := r.computeNamespace(ctx)
	if err != nil {
		return nil, err
	}
	url := batchChangesApplyURL(n, r)
	return &url, nil
}

func (r *batchSpecResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.batchSpec.CreatedAt}
}

func (r *batchSpecResolver) ExpiresAt() *graphqlbackend.DateTime {
	return &graphqlbackend.DateTime{Time: r.batchSpec.ExpiresAt()}
}

func (r *batchSpecResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	return checkSiteAdminOrSameUser(ctx, r.store.DB(), r.batchSpec.UserID)
}

type batchChangeDescriptionResolver struct {
	name, description string
}

func (r *batchChangeDescriptionResolver) Name() string {
	return r.name
}

func (r *batchChangeDescriptionResolver) Description() string {
	return r.description
}

func (r *batchSpecResolver) DiffStat(ctx context.Context) (*graphqlbackend.DiffStat, error) {
	specsConnection := &changesetSpecConnectionResolver{
		store: r.store,
		opts:  store.ListChangesetSpecsOpts{BatchSpecID: r.batchSpec.ID},
	}

	specs, err := specsConnection.Nodes(ctx)
	if err != nil {
		return nil, err
	}

	totalStat := &graphqlbackend.DiffStat{}
	for _, spec := range specs {
		// If we can't convert it, that means it's hidden from the user and we
		// can simply skip it.
		if _, ok := spec.ToVisibleChangesetSpec(); !ok {
			continue
		}

		resolver, ok := spec.(*changesetSpecResolver)
		if !ok {
			// This should never happen.
			continue
		}

		stat := resolver.changesetSpec.DiffStat()
		totalStat.AddStat(stat)
	}

	return totalStat, nil
}

func (r *batchSpecResolver) AppliesToBatchChange(ctx context.Context) (graphqlbackend.BatchChangeResolver, error) {
	svc := service.New(r.store)
	batchChange, err := svc.GetBatchChangeMatchingBatchSpec(ctx, r.batchSpec)
	if err != nil {
		return nil, err
	}
	if batchChange == nil {
		return nil, nil
	}

	return &batchChangeResolver{
		store:       r.store,
		batchChange: batchChange,
	}, nil
}

func (r *batchSpecResolver) SupersedingBatchSpec(ctx context.Context) (graphqlbackend.BatchSpecResolver, error) {
	namespace, err := r.computeNamespace(ctx)
	if err != nil {
		return nil, err
	}

	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return nil, errors.New("user is not authenticated")
	}

	svc := service.New(r.store)
	newest, err := svc.GetNewestBatchSpec(ctx, r.store, r.batchSpec, a.UID)
	if err != nil {
		return nil, err
	}

	// If this is the newest spec, then we can just return nil.
	if newest == nil || newest.ID == r.batchSpec.ID {
		return nil, nil
	}

	// If this spec and the new spec have different creators, we shouldn't
	// return this as a superseding spec.
	if newest.UserID != r.batchSpec.UserID {
		return nil, nil
	}

	// Create our new resolver, reusing as many fields as we can from this one.
	resolver := &batchSpecResolver{
		store:              r.store,
		batchSpec:          newest,
		preloadedNamespace: namespace,
	}

	return resolver, nil
}

func (r *batchSpecResolver) ViewerBatchChangesCodeHosts(ctx context.Context, args *graphqlbackend.ListViewerBatchChangesCodeHostsArgs) (graphqlbackend.BatchChangesCodeHostConnectionResolver, error) {
	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, backend.ErrNotAuthenticated
	}

	specs, _, err := r.store.ListChangesetSpecs(ctx, store.ListChangesetSpecsOpts{BatchSpecID: r.batchSpec.ID})
	if err != nil {
		return nil, err
	}

	offset := 0
	if args.After != nil {
		offset, err = strconv.Atoi(*args.After)
		if err != nil {
			return nil, err
		}
	}

	return &batchChangesCodeHostConnectionResolver{
		userID:                &actor.UID,
		onlyWithoutCredential: args.OnlyWithoutCredential,
		store:                 r.store,
		opts: store.ListCodeHostsOpts{
			RepoIDs: specs.RepoIDs(),
		},
		limitOffset: database.LimitOffset{
			Limit:  int(args.First),
			Offset: offset,
		},
	}, nil
}

func (r *batchSpecResolver) AllowUnsupported() *bool {
	if r.batchSpec.CreatedFromRaw {
		return &r.batchSpec.AllowUnsupported
	}
	return nil
}

func (r *batchSpecResolver) AllowIgnored() *bool {
	if r.batchSpec.CreatedFromRaw {
		return &r.batchSpec.AllowIgnored
	}
	return nil
}

func (r *batchSpecResolver) AutoApplyEnabled() bool {
	// TODO(ssbc): not implemented
	return false
}

func (r *batchSpecResolver) State(ctx context.Context) (string, error) {
	state, err := r.computeState(ctx)
	if err != nil {
		return "", err
	}
	return state.ToGraphQL(), nil
}

func (r *batchSpecResolver) StartedAt(ctx context.Context) (*graphqlbackend.DateTime, error) {
	if !r.batchSpec.CreatedFromRaw {
		return nil, nil
	}

	state, err := r.computeState(ctx)
	if err != nil {
		return nil, err
	}

	if !state.Started() {
		return nil, nil
	}

	stats, err := r.computeStats(ctx)
	if err != nil {
		return nil, err
	}
	if stats.StartedAt.IsZero() {
		return nil, nil
	}

	return graphqlbackend.DateTimeOrNil(&stats.StartedAt), nil
}

func (r *batchSpecResolver) FinishedAt(ctx context.Context) (*graphqlbackend.DateTime, error) {
	if !r.batchSpec.CreatedFromRaw {
		return nil, nil
	}

	state, err := r.computeState(ctx)
	if err != nil {
		return nil, err
	}

	if !state.Finished() {
		return nil, nil
	}

	stats, err := r.computeStats(ctx)
	if err != nil {
		return nil, err
	}
	if stats.FinishedAt.IsZero() {
		return nil, nil
	}

	return graphqlbackend.DateTimeOrNil(&stats.FinishedAt), nil
}

func (r *batchSpecResolver) FailureMessage(ctx context.Context) (*string, error) {
	resolution, err := r.computeResolutionJob(ctx)
	if err != nil {
		return nil, err
	}
	if resolution != nil && resolution.FailureMessage != nil {
		return resolution.FailureMessage, nil
	}

	validationErr := r.validateChangesetSpecs(ctx)
	if validationErr != nil {
		message := validationErr.Error()
		return &message, nil
	}

	// TODO: look at execution jobs.
	return nil, nil
}

func (r *batchSpecResolver) ImportingChangesets(ctx context.Context, args *graphqlbackend.ListImportingChangesetsArgs) (graphqlbackend.ChangesetSpecConnectionResolver, error) {
	workspaces, err := r.computeBatchSpecWorkspaces(ctx)
	if err != nil {
		return nil, err
	}

	uniqueCSIDs := make(map[int64]struct{})
	for _, w := range workspaces {
		for _, id := range w.ChangesetSpecIDs {
			if _, ok := uniqueCSIDs[id]; !ok {
				uniqueCSIDs[id] = struct{}{}
			}
		}
	}
	specIDs := make([]int64, 0, len(uniqueCSIDs))
	for id := range uniqueCSIDs {
		specIDs = append(specIDs, id)
	}

	opts := store.ListChangesetSpecsOpts{
		IDs:         specIDs,
		BatchSpecID: r.batchSpec.ID,
		Type:        batches.ChangesetSpecDescriptionTypeExisting,
	}
	if err := validateFirstParamDefaults(args.First); err != nil {
		return nil, err
	}
	opts.Limit = int(args.First)
	if args.After != nil {
		id, err := strconv.Atoi(*args.After)
		if err != nil {
			return nil, err
		}
		opts.Cursor = int64(id)
	}

	return &changesetSpecConnectionResolver{store: r.store, opts: opts}, nil
}

func (r *batchSpecResolver) WorkspaceResolution(ctx context.Context) (graphqlbackend.BatchSpecWorkspaceResolutionResolver, error) {
	if !r.batchSpec.CreatedFromRaw {
		return nil, nil
	}
	resolution, err := r.computeResolutionJob(ctx)
	if err != nil {
		return nil, err
	}
	if resolution == nil {
		return nil, nil
	}

	return &batchSpecWorkspaceResolutionResolver{store: r.store, resolution: resolution}, nil
}

func (r *batchSpecResolver) computeNamespace(ctx context.Context) (*graphqlbackend.NamespaceResolver, error) {
	r.namespaceOnce.Do(func() {
		if r.preloadedNamespace != nil {
			r.namespace = r.preloadedNamespace
			return
		}
		var (
			err error
			n   = &graphqlbackend.NamespaceResolver{}
		)

		if r.batchSpec.NamespaceUserID != 0 {
			n.Namespace, err = graphqlbackend.UserByIDInt32(ctx, r.store.DB(), r.batchSpec.NamespaceUserID)
		} else {
			n.Namespace, err = graphqlbackend.OrgByIDInt32(ctx, r.store.DB(), r.batchSpec.NamespaceOrgID)
		}

		if errcode.IsNotFound(err) {
			r.namespace = nil
			r.namespaceErr = errors.New("namespace of batch spec has been deleted")
			return
		}

		r.namespace = n
		r.namespaceErr = err
	})
	return r.namespace, r.namespaceErr
}

func (r *batchSpecResolver) computeResolutionJob(ctx context.Context) (*btypes.BatchSpecResolutionJob, error) {
	r.resolutionOnce.Do(func() {
		var err error
		r.resolution, err = r.store.GetBatchSpecResolutionJob(ctx, store.GetBatchSpecResolutionJobOpts{BatchSpecID: r.batchSpec.ID})
		if err != nil {
			if err == store.ErrNoResults {
				return
			}
			r.resolutionErr = err
		}
	})
	return r.resolution, r.resolutionErr
}

func (r *batchSpecResolver) validateChangesetSpecs(ctx context.Context) error {
	r.validateSpecsOnce.Do(func() {
		svc := service.New(r.store)
		r.validateSpecsErr = svc.ValidateChangesetSpecs(ctx, r.batchSpec.ID)
	})
	return r.validateSpecsErr
}

func (r *batchSpecResolver) computeBatchSpecWorkspaces(ctx context.Context) ([]*btypes.BatchSpecWorkspace, error) {
	r.workspacesOnce.Do(func() {
		r.workspaces, _, r.workspacesErr = r.store.ListBatchSpecWorkspaces(ctx, store.ListBatchSpecWorkspacesOpts{BatchSpecID: r.batchSpec.ID})
	})
	return r.workspaces, r.workspacesErr
}

func (r *batchSpecResolver) computeStats(ctx context.Context) (btypes.BatchSpecStats, error) {
	r.statsOnce.Do(func() {
		svc := service.New(r.store)
		r.stats, r.statsErr = svc.LoadBatchSpecStats(ctx, r.batchSpec)
	})
	return r.stats, r.statsErr
}

func (r *batchSpecResolver) computeState(ctx context.Context) (btypes.BatchSpecState, error) {
	r.stateOnce.Do(func() {
		r.state, r.stateErr = func() (btypes.BatchSpecState, error) {
			stats, err := r.computeStats(ctx)
			if err != nil {
				return "", err
			}

			state := btypes.ComputeBatchSpecState(r.batchSpec, stats)

			// If the BatchSpec finished execution successfully, we validate
			// the changeset specs.
			if state == btypes.BatchSpecStateCompleted {
				validationErr := r.validateChangesetSpecs(ctx)
				if validationErr != nil {
					return btypes.BatchSpecStateFailed, nil
				}
			}

			return state, nil
		}()
	})
	return r.state, r.stateErr
}
