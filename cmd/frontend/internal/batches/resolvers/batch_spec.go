package resolvers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/gitserver"

	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	sgactor "github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/batches/search"
	"github.com/sourcegraph/sourcegraph/internal/batches/service"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	store           *store.Store
	gitserverClient gitserver.Client
	logger          log.Logger

	batchSpec          *btypes.BatchSpec
	preloadedNamespace *graphqlbackend.NamespaceResolver

	// We cache the namespace on the resolver, since it's accessed more than once.
	namespaceOnce sync.Once
	namespace     *graphqlbackend.NamespaceResolver
	namespaceErr  error

	resolutionOnce sync.Once
	resolution     *btypes.BatchSpecResolutionJob
	resolutionErr  error

	validateSpecsOnce sync.Once
	validateSpecsErr  error

	statsOnce sync.Once
	stats     btypes.BatchSpecStats
	statsErr  error

	stateOnce sync.Once
	state     btypes.BatchSpecState
	stateErr  error

	canAdministerOnce sync.Once
	canAdminister     bool
	canAdministerErr  error
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
		gitserverClient:   r.gitserverClient,
		logger:            r.logger,
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
	user, err := graphqlbackend.UserByIDInt32(ctx, r.store.DatabaseDB(), r.batchSpec.UserID)
	if errcode.IsNotFound(err) {
		return nil, nil
	}
	return user, err
}

func (r *batchSpecResolver) Namespace(ctx context.Context) (*graphqlbackend.NamespaceResolver, error) {
	return r.computeNamespace(ctx)
}

func (r *batchSpecResolver) ApplyURL(ctx context.Context) (*string, error) {
	if r.batchSpec.CreatedFromRaw && !r.finishedExecutionWithoutValidationErrors(ctx) {
		return nil, nil
	}

	n, err := r.computeNamespace(ctx)
	if err != nil {
		return nil, err
	}
	url := batchChangesApplyURL(n, r)
	return &url, nil
}

func (r *batchSpecResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.batchSpec.CreatedAt}
}

func (r *batchSpecResolver) ExpiresAt() *gqlutil.DateTime {
	return &gqlutil.DateTime{Time: r.batchSpec.ExpiresAt()}
}

func (r *batchSpecResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	return r.computeCanAdminister(ctx)
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
	added, deleted, err := r.store.GetBatchSpecDiffStat(ctx, r.batchSpec.ID)
	if err != nil {
		return nil, err
	}

	return graphqlbackend.NewDiffStat(diff.Stat{
		Added:   int32(added),
		Deleted: int32(deleted),
	}), nil
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
		store:           r.store,
		gitserverClient: r.gitserverClient,
		batchChange:     batchChange,
		logger:          r.logger,
	}, nil
}

func (r *batchSpecResolver) SupersedingBatchSpec(ctx context.Context) (graphqlbackend.BatchSpecResolver, error) {
	namespace, err := r.computeNamespace(ctx)
	if err != nil {
		return nil, err
	}

	actor := sgactor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, errors.New("user is not authenticated")
	}

	svc := service.New(r.store)
	newest, err := svc.GetNewestBatchSpec(ctx, r.store, r.batchSpec, actor.UID)
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
		logger:             r.logger,
		batchSpec:          newest,
		preloadedNamespace: namespace,
	}

	return resolver, nil
}

func (r *batchSpecResolver) ViewerBatchChangesCodeHosts(ctx context.Context, args *graphqlbackend.ListViewerBatchChangesCodeHostsArgs) (graphqlbackend.BatchChangesCodeHostConnectionResolver, error) {
	actor := sgactor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, auth.ErrNotAuthenticated
	}

	repoIDs, err := r.store.ListBatchSpecRepoIDs(ctx, r.batchSpec.ID)
	if err != nil {
		return nil, err
	}

	// If there are no code hosts, then we don't have to compute anything
	// further.
	if len(repoIDs) == 0 {
		return &emptyBatchChangesCodeHostConnectionResolver{}, nil
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
		logger:                r.logger,
		opts: store.ListCodeHostsOpts{
			RepoIDs:             repoIDs,
			OnlyWithoutWebhooks: args.OnlyWithoutWebhooks,
		},
		limitOffset: database.LimitOffset{
			Limit:  int(args.First),
			Offset: offset,
		},
		db: r.store.DatabaseDB(),
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

func (r *batchSpecResolver) NoCache() *bool {
	if r.batchSpec.CreatedFromRaw {
		return &r.batchSpec.NoCache
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

func (r *batchSpecResolver) StartedAt(ctx context.Context) (*gqlutil.DateTime, error) {
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

	return &gqlutil.DateTime{Time: stats.StartedAt}, nil
}

func (r *batchSpecResolver) FinishedAt(ctx context.Context) (*gqlutil.DateTime, error) {
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

	return &gqlutil.DateTime{Time: stats.FinishedAt}, nil
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

	f := false
	failedJobs, err := r.store.ListBatchSpecWorkspaceExecutionJobs(ctx, store.ListBatchSpecWorkspaceExecutionJobsOpts{
		OnlyWithFailureMessage: true,
		BatchSpecID:            r.batchSpec.ID,
		// Omit canceled, they don't contain useful error messages.
		Cancel:      &f,
		ExcludeRank: true,
	})
	if err != nil {
		return nil, err
	}
	if len(failedJobs) == 0 {
		return nil, nil
	}

	var message strings.Builder
	message.WriteString("Failures:\n\n")
	for i, job := range failedJobs {
		message.WriteString("* " + *job.FailureMessage + "\n")

		if i == 4 {
			break
		}
	}
	if len(failedJobs) > 5 {
		message.WriteString(fmt.Sprintf("\nand %d more", len(failedJobs)-5))
	}

	str := message.String()
	return &str, nil
}

func (r *batchSpecResolver) ImportingChangesets(ctx context.Context, args *graphqlbackend.ListImportingChangesetsArgs) (graphqlbackend.ChangesetSpecConnectionResolver, error) {
	opts := store.ListChangesetSpecsOpts{
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

	return &batchSpecWorkspaceResolutionResolver{store: r.store, logger: r.logger, resolution: resolution}, nil
}

func (r *batchSpecResolver) ViewerCanRetry(ctx context.Context) (bool, error) {
	if !r.batchSpec.CreatedFromRaw {
		return false, nil
	}

	ok, err := r.computeCanAdminister(ctx)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	state, err := r.computeState(ctx)
	if err != nil {
		return false, err
	}

	// If the spec finished successfully, there's nothing to retry.
	if state == btypes.BatchSpecStateCompleted {
		return false, nil
	}

	return state.Finished(), nil
}

func (r *batchSpecResolver) Source() string {
	if r.batchSpec.CreatedFromRaw {
		return btypes.BatchSpecSourceRemote.ToGraphQL()
	}
	return btypes.BatchSpecSourceLocal.ToGraphQL()
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
			n.Namespace, err = graphqlbackend.UserByIDInt32(ctx, r.store.DatabaseDB(), r.batchSpec.NamespaceUserID)
		} else {
			n.Namespace, err = graphqlbackend.OrgByIDInt32(ctx, r.store.DatabaseDB(), r.batchSpec.NamespaceOrgID)
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

func (r *batchSpecResolver) finishedExecutionWithoutValidationErrors(ctx context.Context) bool {
	state, err := r.computeState(ctx)
	if err != nil {
		return false
	}

	if !state.FinishedAndNotCanceled() {
		return false
	}

	validationErr := r.validateChangesetSpecs(ctx)
	return validationErr == nil
}

func (r *batchSpecResolver) validateChangesetSpecs(ctx context.Context) error {
	r.validateSpecsOnce.Do(func() {
		svc := service.New(r.store)
		r.validateSpecsErr = svc.ValidateChangesetSpecs(ctx, r.batchSpec.ID)
	})
	return r.validateSpecsErr
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

func (r *batchSpecResolver) computeCanAdminister(ctx context.Context) (bool, error) {
	r.canAdministerOnce.Do(func() {
		svc := service.New(r.store)
		r.canAdminister, r.canAdministerErr = svc.CheckViewerCanAdminister(ctx, r.batchSpec.NamespaceUserID, r.batchSpec.NamespaceOrgID)
	})
	return r.canAdminister, r.canAdministerErr
}

func (r *batchSpecResolver) Files(ctx context.Context, args *graphqlbackend.ListBatchSpecWorkspaceFilesArgs) (_ graphqlbackend.BatchSpecWorkspaceFileConnectionResolver, err error) {
	if err := validateFirstParamDefaults(args.First); err != nil {
		return nil, err
	}
	opts := store.ListBatchSpecWorkspaceFileOpts{
		LimitOpts: store.LimitOpts{
			Limit: int(args.First),
		},
		BatchSpecRandID: r.batchSpec.RandID,
	}

	if args.After != nil {
		id, err := strconv.Atoi(*args.After)
		if err != nil {
			return nil, err
		}
		opts.Cursor = int64(id)
	}

	return &batchSpecWorkspaceFileConnectionResolver{store: r.store, opts: opts}, nil
}
