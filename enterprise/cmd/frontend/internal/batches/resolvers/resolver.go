package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/search"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/service"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/deviceid"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Resolver is the GraphQL resolver of all things related to batch changes.
type Resolver struct {
	store *store.Store
}

// New returns a new Resolver whose store uses the given database
func New(store *store.Store) graphqlbackend.BatchChangesResolver {
	return &Resolver{store: store}
}

// batchChangesCreateAccess returns true if the current user has batch changes enabled for
// them and can create batchChanges/changesetSpecs/batchSpecs.
func batchChangesCreateAccess(ctx context.Context, db database.DB) error {
	if err := enterprise.BatchChangesEnabledForUser(ctx, db); err != nil {
		return err
	}

	act := actor.FromContext(ctx)
	if !act.IsAuthenticated() {
		return backend.ErrNotAuthenticated
	}
	return nil
}

// checkLicense returns a user-facing error if the batchChanges feature is not purchased
// with the current license or any error occurred while validating the license.
func checkLicense() error {
	batchChangesErr := licensing.Check(licensing.FeatureBatchChanges)
	if batchChangesErr == nil {
		return nil
	}

	if licensing.IsFeatureNotActivated(batchChangesErr) {
		// Let's fallback and check whether (deprecated) campaigns are enabled:
		campaignsErr := licensing.Check(licensing.FeatureCampaigns)
		if campaignsErr == nil {
			return nil
		}
		return batchChangesErr
	}

	return errors.New("Unable to check license feature, please refer to logs for actual error message.")
}

// maxUnlicensedChangesets is the maximum number of changesets that can be
// attached to a batch change when Sourcegraph is unlicensed or the Batch
// Changes feature is disabled.
const maxUnlicensedChangesets = 5

type batchSpecCreatedArg struct {
	ChangesetSpecsCount int `json:"changeset_specs_count"`
}

type batchChangeEventArg struct {
	BatchChangeID int64 `json:"batch_change_id"`
}

func logBackendEvent(ctx context.Context, db database.DB, name string, args any, publicArgs any) error {
	actor := actor.FromContext(ctx)
	jsonArg, err := json.Marshal(args)
	if err != nil {
		return err
	}
	jsonPublicArg, err := json.Marshal(publicArgs)
	if err != nil {
		return err
	}

	return usagestats.LogBackendEvent(db, actor.UID, deviceid.FromContext(ctx), name, jsonArg, jsonPublicArg, featureflag.GetEvaluatedFlagSet(ctx), nil)
}

func (r *Resolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]graphqlbackend.NodeByIDFunc{
		batchChangeIDKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return r.batchChangeByID(ctx, id)
		},
		batchSpecIDKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return r.batchSpecByID(ctx, id)
		},
		changesetSpecIDKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return r.changesetSpecByID(ctx, id)
		},
		changesetIDKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return r.changesetByID(ctx, id)
		},
		batchChangesCredentialIDKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return r.batchChangesCredentialByID(ctx, id)
		},
		bulkOperationIDKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return r.bulkOperationByID(ctx, id)
		},
		batchSpecWorkspaceIDKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return r.batchSpecWorkspaceByID(ctx, id)
		},
	}
}

func (r *Resolver) changesetByID(ctx context.Context, id graphql.ID) (graphqlbackend.ChangesetResolver, error) {
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	changesetID, err := unmarshalChangesetID(id)
	if err != nil {
		return nil, err
	}

	if changesetID == 0 {
		return nil, nil
	}

	changeset, err := r.store.GetChangeset(ctx, store.GetChangesetOpts{ID: changesetID})
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	// 🚨 SECURITY: database.Repos.Get uses the authzFilter under the hood and
	// filters out repositories that the user doesn't have access to.
	repo, err := r.store.Repos().Get(ctx, changeset.RepoID)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, err
	}

	return NewChangesetResolver(r.store, changeset, repo), nil
}

func (r *Resolver) batchChangeByID(ctx context.Context, id graphql.ID) (graphqlbackend.BatchChangeResolver, error) {
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	batchChangeID, err := unmarshalBatchChangeID(id)
	if err != nil {
		return nil, err
	}

	if batchChangeID == 0 {
		return nil, nil
	}

	batchChange, err := r.store.GetBatchChange(ctx, store.GetBatchChangeOpts{ID: batchChangeID})
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &batchChangeResolver{store: r.store, batchChange: batchChange}, nil
}

func (r *Resolver) BatchChange(ctx context.Context, args *graphqlbackend.BatchChangeArgs) (graphqlbackend.BatchChangeResolver, error) {
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	opts := store.GetBatchChangeOpts{Name: args.Name}

	err := graphqlbackend.UnmarshalNamespaceID(graphql.ID(args.Namespace), &opts.NamespaceUserID, &opts.NamespaceOrgID)
	if err != nil {
		return nil, err
	}

	batchChange, err := r.store.GetBatchChange(ctx, opts)
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &batchChangeResolver{store: r.store, batchChange: batchChange}, nil
}

func (r *Resolver) ResolveWorkspacesForBatchSpec(ctx context.Context, args *graphqlbackend.ResolveWorkspacesForBatchSpecArgs) ([]graphqlbackend.ResolvedBatchSpecWorkspaceResolver, error) {
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	// Parse the batch spec.
	evaluatableSpec, err := batcheslib.ParseBatchSpec([]byte(args.BatchSpec))
	if err != nil {
		return nil, err
	}

	// Verify the user is authenticated.
	act := actor.FromContext(ctx)
	if !act.IsAuthenticated() {
		return nil, backend.ErrNotAuthenticated
	}

	// Run the resolution.
	resolver := service.NewWorkspaceResolver(r.store)
	workspaces, err := resolver.ResolveWorkspacesForBatchSpec(ctx, evaluatableSpec)
	if err != nil {
		return nil, err
	}

	// Transform the result into resolvers.
	resolvers := make([]graphqlbackend.ResolvedBatchSpecWorkspaceResolver, 0, len(workspaces))
	for _, w := range workspaces {
		resolvers = append(resolvers, &resolvedBatchSpecWorkspaceResolver{store: r.store, workspace: w})
	}

	return resolvers, nil
}

func (r *Resolver) batchSpecByID(ctx context.Context, id graphql.ID) (graphqlbackend.BatchSpecResolver, error) {
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	batchSpecRandID, err := unmarshalBatchSpecID(id)
	if err != nil {
		return nil, err
	}

	if batchSpecRandID == "" {
		return nil, nil
	}

	batchSpec, err := r.store.GetBatchSpec(ctx, store.GetBatchSpecOpts{RandID: batchSpecRandID})
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	// Everyone can see batch specs, if they have the ID.
	return &batchSpecResolver{store: r.store, batchSpec: batchSpec}, nil
}

func (r *Resolver) changesetSpecByID(ctx context.Context, id graphql.ID) (graphqlbackend.ChangesetSpecResolver, error) {
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	changesetSpecRandID, err := unmarshalChangesetSpecID(id)
	if err != nil {
		return nil, err
	}

	if changesetSpecRandID == "" {
		return nil, nil
	}

	opts := store.GetChangesetSpecOpts{RandID: changesetSpecRandID}
	changesetSpec, err := r.store.GetChangesetSpec(ctx, opts)
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return NewChangesetSpecResolver(ctx, r.store, changesetSpec)
}

type batchChangesCredentialResolver interface {
	graphqlbackend.BatchChangesCredentialResolver
	authenticator(ctx context.Context) (auth.Authenticator, error)
}

func (r *Resolver) batchChangesCredentialByID(ctx context.Context, id graphql.ID) (batchChangesCredentialResolver, error) {
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	dbID, isSiteCredential, err := unmarshalBatchChangesCredentialID(id)
	if err != nil {
		return nil, err
	}

	if dbID == 0 {
		return nil, nil
	}

	if isSiteCredential {
		return r.batchChangesSiteCredentialByID(ctx, dbID)
	}

	return r.batchChangesUserCredentialByID(ctx, dbID)
}

func (r *Resolver) batchChangesUserCredentialByID(ctx context.Context, id int64) (batchChangesCredentialResolver, error) {
	cred, err := r.store.UserCredentials().GetByID(ctx, id)
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	if err := backend.CheckSiteAdminOrSameUser(ctx, r.store.DatabaseDB(), cred.UserID); err != nil {
		return nil, err
	}

	return &batchChangesUserCredentialResolver{credential: cred}, nil
}

func (r *Resolver) batchChangesSiteCredentialByID(ctx context.Context, id int64) (batchChangesCredentialResolver, error) {
	// Todo: Is this required? Should everyone be able to see there are _some_ credentials?
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	cred, err := r.store.GetSiteCredential(ctx, store.GetSiteCredentialOpts{ID: id})
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &batchChangesSiteCredentialResolver{credential: cred}, nil
}

func (r *Resolver) bulkOperationByID(ctx context.Context, id graphql.ID) (graphqlbackend.BulkOperationResolver, error) {
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	dbID, err := unmarshalBulkOperationID(id)
	if err != nil {
		return nil, err
	}

	if dbID == "" {
		return nil, nil
	}

	return r.bulkOperationByIDString(ctx, dbID)
}

func (r *Resolver) bulkOperationByIDString(ctx context.Context, id string) (graphqlbackend.BulkOperationResolver, error) {
	bulkOperation, err := r.store.GetBulkOperation(ctx, store.GetBulkOperationOpts{ID: id})
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}
	return &bulkOperationResolver{store: r.store, bulkOperation: bulkOperation}, nil
}

func (r *Resolver) batchSpecWorkspaceByID(ctx context.Context, gqlID graphql.ID) (graphqlbackend.BatchSpecWorkspaceResolver, error) {
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	id, err := unmarshalBatchSpecWorkspaceID(gqlID)
	if err != nil {
		return nil, err
	}

	if id == 0 {
		return nil, nil
	}

	w, err := r.store.GetBatchSpecWorkspace(ctx, store.GetBatchSpecWorkspaceOpts{ID: id})
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	spec, err := r.store.GetBatchSpec(ctx, store.GetBatchSpecOpts{ID: w.BatchSpecID})
	if err != nil {
		return nil, err
	}

	ex, err := r.store.GetBatchSpecWorkspaceExecutionJob(ctx, store.GetBatchSpecWorkspaceExecutionJobOpts{BatchSpecWorkspaceID: w.ID})
	if err != nil && err != store.ErrNoResults {
		return nil, err
	}

	return newBatchSpecWorkspaceResolver(ctx, r.store, w, ex, spec.Spec)
}

func (r *Resolver) CreateBatchChange(ctx context.Context, args *graphqlbackend.CreateBatchChangeArgs) (graphqlbackend.BatchChangeResolver, error) {
	var err error
	tr, _ := trace.New(ctx, "Resolver.CreateBatchChange", fmt.Sprintf("BatchSpec %s", args.BatchSpec))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	opts := service.ApplyBatchChangeOpts{
		// This is what differentiates CreateBatchChange from ApplyBatchChange
		FailIfBatchChangeExists: true,
	}
	batchChange, err := r.applyOrCreateBatchChange(ctx, &graphqlbackend.ApplyBatchChangeArgs{
		BatchSpec:         args.BatchSpec,
		EnsureBatchChange: nil,
		PublicationStates: args.PublicationStates,
	}, opts)
	if err != nil {
		return nil, err
	}

	arg := &batchChangeEventArg{BatchChangeID: batchChange.ID}
	err = logBackendEvent(ctx, r.store.DatabaseDB(), "BatchChangeCreated", arg, arg)
	if err != nil {
		return nil, err
	}

	return &batchChangeResolver{store: r.store, batchChange: batchChange}, nil
}

func (r *Resolver) ApplyBatchChange(ctx context.Context, args *graphqlbackend.ApplyBatchChangeArgs) (graphqlbackend.BatchChangeResolver, error) {
	var err error
	tr, ctx := trace.New(ctx, "Resolver.ApplyBatchChange", fmt.Sprintf("BatchSpec %s", args.BatchSpec))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	batchChange, err := r.applyOrCreateBatchChange(ctx, args, service.ApplyBatchChangeOpts{})
	if err != nil {
		return nil, err
	}

	arg := &batchChangeEventArg{BatchChangeID: batchChange.ID}
	err = logBackendEvent(ctx, r.store.DatabaseDB(), "BatchChangeCreatedOrUpdated", arg, arg)
	if err != nil {
		return nil, err
	}

	return &batchChangeResolver{store: r.store, batchChange: batchChange}, nil
}

func addPublicationStatesToOptions(in *[]graphqlbackend.ChangesetSpecPublicationStateInput, opts *service.UiPublicationStates) error {
	var errs error

	if in != nil && *in != nil {
		for _, state := range *in {
			id, err := unmarshalChangesetSpecID(state.ChangesetSpec)
			if err != nil {
				return err
			}

			if err := opts.Add(id, state.PublicationState); err != nil {
				errs = errors.Append(errs, err)
			}
		}

	}

	return errs
}

func (r *Resolver) applyOrCreateBatchChange(ctx context.Context, args *graphqlbackend.ApplyBatchChangeArgs, opts service.ApplyBatchChangeOpts) (*btypes.BatchChange, error) {
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	var err error
	if opts.BatchSpecRandID, err = unmarshalBatchSpecID(args.BatchSpec); err != nil {
		return nil, err
	}

	if opts.BatchSpecRandID == "" {
		return nil, ErrIDIsZero{}
	}

	if licenseErr := checkLicense(); licenseErr != nil {
		if licensing.IsFeatureNotActivated(licenseErr) {
			batchSpec, err := r.store.GetBatchSpec(ctx, store.GetBatchSpecOpts{
				RandID: opts.BatchSpecRandID,
			})
			if err != nil {
				return nil, err
			}
			count, err := r.store.CountChangesetSpecs(ctx, store.CountChangesetSpecsOpts{BatchSpecID: batchSpec.ID})
			if err != nil {
				return nil, err
			}
			if count > maxUnlicensedChangesets {
				return nil, ErrBatchChangesUnlicensed{licenseErr}
			}
		} else {
			return nil, licenseErr
		}
	}

	if args.EnsureBatchChange != nil {
		opts.EnsureBatchChangeID, err = unmarshalBatchChangeID(*args.EnsureBatchChange)
		if err != nil {
			return nil, err
		}
	}

	if err := addPublicationStatesToOptions(args.PublicationStates, &opts.PublicationStates); err != nil {
		return nil, err
	}

	svc := service.New(r.store)
	// 🚨 SECURITY: ApplyBatchChange checks whether the user has permission to
	// apply the batch spec.
	batchChange, err := svc.ApplyBatchChange(ctx, opts)
	if err != nil {
		if err == service.ErrEnsureBatchChangeFailed {
			return nil, ErrEnsureBatchChangeFailed{}
		} else if err == service.ErrApplyClosedBatchChange {
			return nil, ErrApplyClosedBatchChange{}
		} else if err == service.ErrMatchingBatchChangeExists {
			return nil, ErrMatchingBatchChangeExists{}
		}
		return nil, err
	}

	return batchChange, nil
}

func (r *Resolver) CreateBatchSpec(ctx context.Context, args *graphqlbackend.CreateBatchSpecArgs) (graphqlbackend.BatchSpecResolver, error) {
	var err error
	tr, ctx := trace.New(ctx, "CreateBatchSpec", fmt.Sprintf("Resolver.CreateBatchSpec %s, Spec %q", args.Namespace, args.BatchSpec))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if err := batchChangesCreateAccess(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	if err := checkLicense(); err != nil {
		if licensing.IsFeatureNotActivated(err) {
			if len(args.ChangesetSpecs) > maxUnlicensedChangesets {
				return nil, ErrBatchChangesUnlicensed{err}
			}
		} else {
			return nil, err
		}
	}

	opts := service.CreateBatchSpecOpts{RawSpec: args.BatchSpec}

	err = graphqlbackend.UnmarshalNamespaceID(args.Namespace, &opts.NamespaceUserID, &opts.NamespaceOrgID)
	if err != nil {
		return nil, err
	}

	for _, graphqlID := range args.ChangesetSpecs {
		randID, err := unmarshalChangesetSpecID(graphqlID)
		if err != nil {
			return nil, err
		}
		opts.ChangesetSpecRandIDs = append(opts.ChangesetSpecRandIDs, randID)
	}

	svc := service.New(r.store)
	batchSpec, err := svc.CreateBatchSpec(ctx, opts)
	if err != nil {
		return nil, err
	}

	eventArg := &batchSpecCreatedArg{ChangesetSpecsCount: len(opts.ChangesetSpecRandIDs)}
	if err := logBackendEvent(ctx, r.store.DatabaseDB(), "BatchSpecCreated", eventArg, eventArg); err != nil {
		return nil, err
	}

	specResolver := &batchSpecResolver{
		store:     r.store,
		batchSpec: batchSpec,
	}

	return specResolver, nil
}

func (r *Resolver) CreateChangesetSpec(ctx context.Context, args *graphqlbackend.CreateChangesetSpecArgs) (graphqlbackend.ChangesetSpecResolver, error) {
	var err error
	tr, ctx := trace.New(ctx, "Resolver.CreateChangesetSpec", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if err := batchChangesCreateAccess(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	act := actor.FromContext(ctx)
	// Actor MUST be logged in at this stage, because batchChangesCreateAccess checks that already.
	// To be extra safe, we'll just do the cheap check again here so if anyone ever modifies
	// batchChangesCreateAccess, we still enforce it here.
	if !act.IsAuthenticated() {
		return nil, backend.ErrNotAuthenticated
	}

	svc := service.New(r.store)
	spec, err := svc.CreateChangesetSpec(ctx, args.ChangesetSpec, act.UID)
	if err != nil {
		return nil, err
	}

	return NewChangesetSpecResolver(ctx, r.store, spec)
}

func (r *Resolver) MoveBatchChange(ctx context.Context, args *graphqlbackend.MoveBatchChangeArgs) (graphqlbackend.BatchChangeResolver, error) {
	var err error
	tr, ctx := trace.New(ctx, "Resolver.MoveBatchChange", fmt.Sprintf("BatchChange %s", args.BatchChange))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	batchChangeID, err := unmarshalBatchChangeID(args.BatchChange)
	if err != nil {
		return nil, err
	}

	if batchChangeID == 0 {
		return nil, ErrIDIsZero{}
	}

	opts := service.MoveBatchChangeOpts{
		BatchChangeID: batchChangeID,
	}

	if args.NewName != nil {
		opts.NewName = *args.NewName
	}

	if args.NewNamespace != nil {
		err := graphqlbackend.UnmarshalNamespaceID(*args.NewNamespace, &opts.NewNamespaceUserID, &opts.NewNamespaceOrgID)
		if err != nil {
			return nil, err
		}
	}

	svc := service.New(r.store)
	// 🚨 SECURITY: MoveBatchChange checks whether the current user is authorized.
	batchChange, err := svc.MoveBatchChange(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &batchChangeResolver{store: r.store, batchChange: batchChange}, nil
}

func (r *Resolver) DeleteBatchChange(ctx context.Context, args *graphqlbackend.DeleteBatchChangeArgs) (_ *graphqlbackend.EmptyResponse, err error) {
	tr, ctx := trace.New(ctx, "Resolver.DeleteBatchChange", fmt.Sprintf("BatchChange: %q", args.BatchChange))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	batchChangeID, err := unmarshalBatchChangeID(args.BatchChange)
	if err != nil {
		return nil, err
	}

	if batchChangeID == 0 {
		return nil, ErrIDIsZero{}
	}

	svc := service.New(r.store)
	// 🚨 SECURITY: DeleteBatchChange checks whether current user is authorized.
	err = svc.DeleteBatchChange(ctx, batchChangeID)
	if err != nil {
		return nil, err
	}

	arg := &batchChangeEventArg{BatchChangeID: batchChangeID}
	if err := logBackendEvent(ctx, r.store.DatabaseDB(), "BatchChangeDeleted", arg, arg); err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, err
}

func (r *Resolver) BatchChanges(ctx context.Context, args *graphqlbackend.ListBatchChangesArgs) (graphqlbackend.BatchChangesConnectionResolver, error) {
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	opts := store.ListBatchChangesOpts{}

	state, err := parseBatchChangeState(args.State)
	if err != nil {
		return nil, err
	}
	if state != "" {
		opts.States = []btypes.BatchChangeState{state}
	}

	// If multiple `states` are provided, prefer them over `state`.
	if args.States != nil {
		states, err := parseBatchChangeStates(args.States)
		if err != nil {
			return nil, err
		}
		opts.States = states
	}

	if err := validateFirstParamDefaults(args.First); err != nil {
		return nil, err
	}
	opts.Limit = int(args.First)
	if args.After != nil {
		cursor, err := strconv.ParseInt(*args.After, 10, 32)
		if err != nil {
			return nil, err
		}
		opts.Cursor = cursor
	}

	authErr := backend.CheckCurrentUserIsSiteAdmin(ctx, r.store.DatabaseDB())
	if authErr != nil && authErr != backend.ErrMustBeSiteAdmin {
		return nil, authErr
	}
	isSiteAdmin := authErr != backend.ErrMustBeSiteAdmin
	if !isSiteAdmin {
		actor := actor.FromContext(ctx)
		if args.ViewerCanAdminister != nil && *args.ViewerCanAdminister {
			opts.CreatorID = actor.UID
		}

		// 🚨 SECURITY: If the user is not an admin, we don't want to include
		// unapplied (draft) BatchChanges except those that the user owns.
		opts.ExcludeDraftsNotOwnedByUserID = actor.UID
	}

	if args.Namespace != nil {
		err := graphqlbackend.UnmarshalNamespaceID(*args.Namespace, &opts.NamespaceUserID, &opts.NamespaceOrgID)
		if err != nil {
			return nil, err
		}
	}

	if args.Repo != nil {
		repoID, err := graphqlbackend.UnmarshalRepositoryID(*args.Repo)
		if err != nil {
			return nil, err
		}
		opts.RepoID = repoID
	}

	return &batchChangesConnectionResolver{
		store: r.store,
		opts:  opts,
	}, nil
}

func (r *Resolver) RepoChangesetsStats(ctx context.Context, repo *graphql.ID) (graphqlbackend.RepoChangesetsStatsResolver, error) {
	repoID, err := graphqlbackend.UnmarshalRepositoryID(*repo)
	if err != nil {
		return nil, err
	}

	stats, err := r.store.GetRepoChangesetsStats(ctx, repoID)
	if err != nil {
		return nil, err
	}
	return &repoChangesetsStatsResolver{stats: *stats}, nil
}

func (r *Resolver) RepoDiffStat(ctx context.Context, repo *graphql.ID) (*graphqlbackend.DiffStat, error) {
	repoID, err := graphqlbackend.UnmarshalRepositoryID(*repo)
	if err != nil {
		return nil, err
	}

	diffStat, err := r.store.GetRepoDiffStat(ctx, repoID)
	if err != nil {
		return nil, err
	}
	return graphqlbackend.NewDiffStat(*diffStat), nil
}

func (r *Resolver) BatchChangesCodeHosts(ctx context.Context, args *graphqlbackend.ListBatchChangesCodeHostsArgs) (graphqlbackend.BatchChangesCodeHostConnectionResolver, error) {
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	if args.UserID != nil {
		// 🚨 SECURITY: Only viewable for self or by site admins.
		if err := backend.CheckSiteAdminOrSameUser(ctx, r.store.DatabaseDB(), *args.UserID); err != nil {
			return nil, err
		}
	}

	if err := validateFirstParamDefaults(args.First); err != nil {
		return nil, err
	}

	limitOffset := database.LimitOffset{
		Limit: int(args.First),
	}
	if args.After != nil {
		cursor, err := strconv.ParseInt(*args.After, 10, 32)
		if err != nil {
			return nil, err
		}
		limitOffset.Offset = int(cursor)
	}

	return &batchChangesCodeHostConnectionResolver{userID: args.UserID, limitOffset: limitOffset, store: r.store}, nil
}

// listChangesetOptsFromArgs turns the graphqlbackend.ListChangesetsArgs into
// ListChangesetsOpts.
// If the args do not include a filter that would reveal sensitive information
// about a changeset the user doesn't have access to, the second return value
// is false.
func listChangesetOptsFromArgs(args *graphqlbackend.ListChangesetsArgs, batchChangeID int64) (opts store.ListChangesetsOpts, optsSafe bool, err error) {
	if args == nil {
		return opts, true, nil
	}

	safe := true

	// TODO: This _could_ become problematic if a user has a batch change with > 10000 changesets, once
	// we use cursor based pagination in the frontend for ChangesetConnections this problem will disappear.
	// Currently we cannot enable it, though, because we want to re-fetch the whole list periodically to
	// check for a change in the changeset states.
	if err := validateFirstParamDefaults(args.First); err != nil {
		return opts, false, err
	}
	opts.Limit = int(args.First)

	if args.After != nil {
		cursor, err := strconv.ParseInt(*args.After, 10, 32)
		if err != nil {
			return opts, false, errors.Wrap(err, "parsing after cursor")
		}
		opts.Cursor = cursor
	}

	if args.OnlyClosable != nil && *args.OnlyClosable {
		if args.State != nil {
			return opts, false, errors.New("invalid combination of state and onlyClosable")
		}

		opts.States = []btypes.ChangesetState{btypes.ChangesetStateOpen, btypes.ChangesetStateDraft}
	}

	if args.State != nil {
		state := btypes.ChangesetState(*args.State)
		if !state.Valid() {
			return opts, false, errors.New("changeset state not valid")
		}

		opts.States = []btypes.ChangesetState{state}
	}

	if args.ReviewState != nil {
		state := btypes.ChangesetReviewState(*args.ReviewState)
		if !state.Valid() {
			return opts, false, errors.New("changeset review state not valid")
		}
		opts.ExternalReviewState = &state
		// If the user filters by ReviewState we cannot include hidden
		// changesets, since that would leak information.
		safe = false
	}
	if args.CheckState != nil {
		state := btypes.ChangesetCheckState(*args.CheckState)
		if !state.Valid() {
			return opts, false, errors.New("changeset check state not valid")
		}
		opts.ExternalCheckState = &state
		// If the user filters by CheckState we cannot include hidden
		// changesets, since that would leak information.
		safe = false
	}
	if args.OnlyPublishedByThisBatchChange != nil {
		published := btypes.ChangesetPublicationStatePublished

		opts.OwnedByBatchChangeID = batchChangeID
		opts.PublicationState = &published
	}
	if args.Search != nil {
		var err error
		opts.TextSearch, err = search.ParseTextSearch(*args.Search)
		if err != nil {
			return opts, false, errors.Wrap(err, "parsing search")
		}
		// Since we search for the repository name in text searches, the
		// presence or absence of results may leak information about hidden
		// repositories.
		safe = false
	}
	if args.OnlyArchived {
		opts.OnlyArchived = args.OnlyArchived
	}
	if args.Repo != nil {
		repoID, err := graphqlbackend.UnmarshalRepositoryID(*args.Repo)
		if err != nil {
			return opts, false, errors.Wrap(err, "unmarshalling repo id")
		}
		opts.RepoIDs = []api.RepoID{repoID}
	}

	return opts, safe, nil
}

func (r *Resolver) CloseBatchChange(ctx context.Context, args *graphqlbackend.CloseBatchChangeArgs) (_ graphqlbackend.BatchChangeResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.CloseBatchChange", fmt.Sprintf("BatchChange: %q", args.BatchChange))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	batchChangeID, err := unmarshalBatchChangeID(args.BatchChange)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling batch change id")
	}

	if batchChangeID == 0 {
		return nil, ErrIDIsZero{}
	}

	svc := service.New(r.store)
	// 🚨 SECURITY: CloseBatchChange checks whether current user is authorized.
	batchChange, err := svc.CloseBatchChange(ctx, batchChangeID, args.CloseChangesets)
	if err != nil {
		return nil, errors.Wrap(err, "closing batch change")
	}

	arg := &batchChangeEventArg{BatchChangeID: batchChangeID}
	if err := logBackendEvent(ctx, r.store.DatabaseDB(), "BatchChangeClosed", arg, arg); err != nil {
		return nil, err
	}

	return &batchChangeResolver{store: r.store, batchChange: batchChange}, nil
}

func (r *Resolver) SyncChangeset(ctx context.Context, args *graphqlbackend.SyncChangesetArgs) (_ *graphqlbackend.EmptyResponse, err error) {
	tr, ctx := trace.New(ctx, "Resolver.SyncChangeset", fmt.Sprintf("Changeset: %q", args.Changeset))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	changesetID, err := unmarshalChangesetID(args.Changeset)
	if err != nil {
		return nil, err
	}

	if changesetID == 0 {
		return nil, ErrIDIsZero{}
	}

	// 🚨 SECURITY: EnqueueChangesetSync checks whether current user is authorized.
	svc := service.New(r.store)
	if err = svc.EnqueueChangesetSync(ctx, changesetID); err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) ReenqueueChangeset(ctx context.Context, args *graphqlbackend.ReenqueueChangesetArgs) (_ graphqlbackend.ChangesetResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.ReenqueueChangeset", fmt.Sprintf("Changeset: %q", args.Changeset))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	changesetID, err := unmarshalChangesetID(args.Changeset)
	if err != nil {
		return nil, err
	}

	if changesetID == 0 {
		return nil, ErrIDIsZero{}
	}

	// 🚨 SECURITY: ReenqueueChangeset checks whether the current user is authorized and can administer the changeset.
	svc := service.New(r.store)
	changeset, repo, err := svc.ReenqueueChangeset(ctx, changesetID)
	if err != nil {
		return nil, err
	}

	return NewChangesetResolver(r.store, changeset, repo), nil
}

func (r *Resolver) CreateBatchChangesCredential(ctx context.Context, args *graphqlbackend.CreateBatchChangesCredentialArgs) (_ graphqlbackend.BatchChangesCredentialResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.CreateBatchChangesCredential", fmt.Sprintf("%q (%q)", args.ExternalServiceKind, args.ExternalServiceURL))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	var userID int32
	if args.User != nil {
		userID, err = graphqlbackend.UnmarshalUserID(*args.User)
		if err != nil {
			return nil, err
		}

		if userID == 0 {
			return nil, ErrIDIsZero{}
		}
	}

	// Need to validate externalServiceKind, otherwise this'll panic.
	kind, valid := extsvc.ParseServiceKind(args.ExternalServiceKind)
	if !valid {
		return nil, errors.New("invalid external service kind")
	}

	if args.Credential == "" {
		return nil, errors.New("empty credential not allowed")
	}

	if userID != 0 {
		return r.createBatchChangesUserCredential(ctx, args.ExternalServiceURL, extsvc.KindToType(kind), userID, args.Credential, args.Username)
	}

	return r.createBatchChangesSiteCredential(ctx, args.ExternalServiceURL, extsvc.KindToType(kind), args.Credential, args.Username)
}

func (r *Resolver) createBatchChangesUserCredential(ctx context.Context, externalServiceURL, externalServiceType string, userID int32, credential string, username *string) (graphqlbackend.BatchChangesCredentialResolver, error) {
	// 🚨 SECURITY: Check that the requesting user can create the credential.
	if err := backend.CheckSiteAdminOrSameUser(ctx, r.store.DatabaseDB(), userID); err != nil {
		return nil, err
	}

	// Throw error documented in schema.graphql.
	userCredentialScope := database.UserCredentialScope{
		Domain:              database.UserCredentialDomainBatches,
		ExternalServiceID:   externalServiceURL,
		ExternalServiceType: externalServiceType,
		UserID:              userID,
	}
	existing, err := r.store.UserCredentials().GetByScope(ctx, userCredentialScope)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDuplicateCredential{}
	}

	a, err := r.generateAuthenticatorForCredential(ctx, externalServiceType, externalServiceURL, credential, username)
	if err != nil {
		return nil, err
	}
	cred, err := r.store.UserCredentials().Create(ctx, userCredentialScope, a)
	if err != nil {
		return nil, err
	}

	return &batchChangesUserCredentialResolver{credential: cred}, nil
}

func (r *Resolver) createBatchChangesSiteCredential(ctx context.Context, externalServiceURL, externalServiceType string, credential string, username *string) (graphqlbackend.BatchChangesCredentialResolver, error) {
	// 🚨 SECURITY: Check that a site credential can only be created
	// by a site-admin.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	// Throw error documented in schema.graphql.
	existing, err := r.store.GetSiteCredential(ctx, store.GetSiteCredentialOpts{
		ExternalServiceType: externalServiceType,
		ExternalServiceID:   externalServiceURL,
	})
	if err != nil && err != store.ErrNoResults {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDuplicateCredential{}
	}

	a, err := r.generateAuthenticatorForCredential(ctx, externalServiceType, externalServiceURL, credential, username)
	if err != nil {
		return nil, err
	}
	cred := &btypes.SiteCredential{
		ExternalServiceID:   externalServiceURL,
		ExternalServiceType: externalServiceType,
	}
	if err := r.store.CreateSiteCredential(ctx, cred, a); err != nil {
		return nil, err
	}

	return &batchChangesSiteCredentialResolver{credential: cred}, nil
}

func (r *Resolver) generateAuthenticatorForCredential(ctx context.Context, externalServiceType, externalServiceURL, credential string, username *string) (auth.Authenticator, error) {
	svc := service.New(r.store)

	var a auth.Authenticator
	keypair, err := encryption.GenerateRSAKey()
	if err != nil {
		return nil, err
	}
	if externalServiceType == extsvc.TypeBitbucketServer {
		// We need to fetch the username for the token, as just an OAuth token isn't enough for some reason..
		username, err := svc.FetchUsernameForBitbucketServerToken(ctx, externalServiceURL, externalServiceType, credential)
		if err != nil {
			if bitbucketserver.IsUnauthorized(err) {
				return nil, &ErrVerifyCredentialFailed{SourceErr: err}
			}
			return nil, err
		}
		a = &auth.BasicAuthWithSSH{
			BasicAuth:  auth.BasicAuth{Username: username, Password: credential},
			PrivateKey: keypair.PrivateKey,
			PublicKey:  keypair.PublicKey,
			Passphrase: keypair.Passphrase,
		}
	} else if externalServiceType == extsvc.TypeBitbucketCloud {
		a = &auth.BasicAuthWithSSH{
			BasicAuth:  auth.BasicAuth{Username: *username, Password: credential},
			PrivateKey: keypair.PrivateKey,
			PublicKey:  keypair.PublicKey,
			Passphrase: keypair.Passphrase,
		}
	} else {
		a = &auth.OAuthBearerTokenWithSSH{
			OAuthBearerToken: auth.OAuthBearerToken{Token: credential},
			PrivateKey:       keypair.PrivateKey,
			PublicKey:        keypair.PublicKey,
			Passphrase:       keypair.Passphrase,
		}
	}

	// Validate the newly created authenticator.
	if err := svc.ValidateAuthenticator(ctx, externalServiceURL, externalServiceType, a); err != nil {
		return nil, &ErrVerifyCredentialFailed{SourceErr: err}
	}
	return a, nil
}

func (r *Resolver) DeleteBatchChangesCredential(ctx context.Context, args *graphqlbackend.DeleteBatchChangesCredentialArgs) (_ *graphqlbackend.EmptyResponse, err error) {
	tr, ctx := trace.New(ctx, "Resolver.DeleteBatchChangesCredential", fmt.Sprintf("Credential: %q", args.BatchChangesCredential))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	dbID, isSiteCredential, err := unmarshalBatchChangesCredentialID(args.BatchChangesCredential)
	if err != nil {
		return nil, err
	}

	if dbID == 0 {
		return nil, ErrIDIsZero{}
	}

	if isSiteCredential {
		return r.deleteBatchChangesSiteCredential(ctx, dbID)
	}

	return r.deleteBatchChangesUserCredential(ctx, dbID)
}

func (r *Resolver) deleteBatchChangesUserCredential(ctx context.Context, credentialDBID int64) (*graphqlbackend.EmptyResponse, error) {
	// Get existing credential.
	cred, err := r.store.UserCredentials().GetByID(ctx, credentialDBID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check that the requesting user may delete the credential.
	if err := backend.CheckSiteAdminOrSameUser(ctx, r.store.DatabaseDB(), cred.UserID); err != nil {
		return nil, err
	}

	// This also fails if the credential was not found.
	if err := r.store.UserCredentials().Delete(ctx, credentialDBID); err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) deleteBatchChangesSiteCredential(ctx context.Context, credentialDBID int64) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Check that the requesting user may delete the credential.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	// This also fails if the credential was not found.
	if err := r.store.DeleteSiteCredential(ctx, credentialDBID); err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) DetachChangesets(ctx context.Context, args *graphqlbackend.DetachChangesetsArgs) (_ graphqlbackend.BulkOperationResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.DetachChangesets", fmt.Sprintf("BatchChange: %q, len(Changesets): %d", args.BatchChange, len(args.Changesets)))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	batchChangeID, changesetIDs, err := unmarshalBulkOperationBaseArgs(args.BulkOperationBaseArgs)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: CreateChangesetJobs checks whether current user is authorized.
	svc := service.New(r.store)
	bulkGroupID, err := svc.CreateChangesetJobs(
		ctx,
		batchChangeID,
		changesetIDs,
		btypes.ChangesetJobTypeDetach,
		&btypes.ChangesetJobDetachPayload{},
		store.ListChangesetsOpts{
			// Only allow to run this on archived changesets.
			OnlyArchived: true,
		},
	)
	if err != nil {
		return nil, err
	}

	return r.bulkOperationByIDString(ctx, bulkGroupID)
}

func (r *Resolver) CreateChangesetComments(ctx context.Context, args *graphqlbackend.CreateChangesetCommentsArgs) (_ graphqlbackend.BulkOperationResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.CreateChangesetComments", fmt.Sprintf("BatchChange: %q, len(Changesets): %d", args.BatchChange, len(args.Changesets)))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	if args.Body == "" {
		return nil, errors.New("empty comment body is not allowed")
	}

	batchChangeID, changesetIDs, err := unmarshalBulkOperationBaseArgs(args.BulkOperationBaseArgs)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: CreateChangesetJobs checks whether current user is authorized.
	svc := service.New(r.store)
	published := btypes.ChangesetPublicationStatePublished
	bulkGroupID, err := svc.CreateChangesetJobs(
		ctx,
		batchChangeID,
		changesetIDs,
		btypes.ChangesetJobTypeComment,
		&btypes.ChangesetJobCommentPayload{
			Message: args.Body,
		},
		store.ListChangesetsOpts{
			// Also include archived changesets, we allow commenting on them as well.
			IncludeArchived: true,
			// We can only comment on published changesets.
			PublicationState: &published,
		},
	)
	if err != nil {
		return nil, err
	}

	return r.bulkOperationByIDString(ctx, bulkGroupID)
}

func (r *Resolver) ReenqueueChangesets(ctx context.Context, args *graphqlbackend.ReenqueueChangesetsArgs) (_ graphqlbackend.BulkOperationResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.ReenqueueChangesets", fmt.Sprintf("BatchChange: %q, len(Changesets): %d", args.BatchChange, len(args.Changesets)))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	batchChangeID, changesetIDs, err := unmarshalBulkOperationBaseArgs(args.BulkOperationBaseArgs)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: CreateChangesetJobs checks whether current user is authorized.
	svc := service.New(r.store)
	bulkGroupID, err := svc.CreateChangesetJobs(
		ctx,
		batchChangeID,
		changesetIDs,
		btypes.ChangesetJobTypeReenqueue,
		&btypes.ChangesetJobReenqueuePayload{},
		store.ListChangesetsOpts{
			// Only allow to retry failed changesets.
			ReconcilerStates: []btypes.ReconcilerState{btypes.ReconcilerStateFailed},
		},
	)
	if err != nil {
		return nil, err
	}

	return r.bulkOperationByIDString(ctx, bulkGroupID)
}

func (r *Resolver) MergeChangesets(ctx context.Context, args *graphqlbackend.MergeChangesetsArgs) (_ graphqlbackend.BulkOperationResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.MergeChangesets", fmt.Sprintf("BatchChange: %q, len(Changesets): %d", args.BatchChange, len(args.Changesets)))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	batchChangeID, changesetIDs, err := unmarshalBulkOperationBaseArgs(args.BulkOperationBaseArgs)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: CreateChangesetJobs checks whether current user is authorized.
	svc := service.New(r.store)
	published := btypes.ChangesetPublicationStatePublished
	openState := btypes.ChangesetExternalStateOpen
	bulkGroupID, err := svc.CreateChangesetJobs(
		ctx,
		batchChangeID,
		changesetIDs,
		btypes.ChangesetJobTypeMerge,
		&btypes.ChangesetJobMergePayload{Squash: args.Squash},
		store.ListChangesetsOpts{
			PublicationState: &published,
			ReconcilerStates: []btypes.ReconcilerState{btypes.ReconcilerStateCompleted},
			ExternalStates:   []btypes.ChangesetExternalState{openState},
		},
	)
	if err != nil {
		return nil, err
	}

	return r.bulkOperationByIDString(ctx, bulkGroupID)
}

func (r *Resolver) CloseChangesets(ctx context.Context, args *graphqlbackend.CloseChangesetsArgs) (_ graphqlbackend.BulkOperationResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.CloseChangesets", fmt.Sprintf("BatchChange: %q, len(Changesets): %d", args.BatchChange, len(args.Changesets)))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	batchChangeID, changesetIDs, err := unmarshalBulkOperationBaseArgs(args.BulkOperationBaseArgs)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: CreateChangesetJobs checks whether current user is authorized.
	svc := service.New(r.store)
	published := btypes.ChangesetPublicationStatePublished
	bulkGroupID, err := svc.CreateChangesetJobs(
		ctx,
		batchChangeID,
		changesetIDs,
		btypes.ChangesetJobTypeClose,
		&btypes.ChangesetJobClosePayload{},
		store.ListChangesetsOpts{
			PublicationState: &published,
			ReconcilerStates: []btypes.ReconcilerState{btypes.ReconcilerStateCompleted},
			ExternalStates:   []btypes.ChangesetExternalState{btypes.ChangesetExternalStateOpen, btypes.ChangesetExternalStateDraft},
		},
	)
	if err != nil {
		return nil, err
	}

	return r.bulkOperationByIDString(ctx, bulkGroupID)
}

func (r *Resolver) PublishChangesets(ctx context.Context, args *graphqlbackend.PublishChangesetsArgs) (_ graphqlbackend.BulkOperationResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.PublishChangesets", fmt.Sprintf("BatchChange: %q, len(Changesets): %d", args.BatchChange, len(args.Changesets)))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	batchChangeID, changesetIDs, err := unmarshalBulkOperationBaseArgs(args.BulkOperationBaseArgs)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: CreateChangesetJobs checks whether current user is authorized.
	svc := service.New(r.store)
	bulkGroupID, err := svc.CreateChangesetJobs(
		ctx,
		batchChangeID,
		changesetIDs,
		btypes.ChangesetJobTypePublish,
		&btypes.ChangesetJobPublishPayload{Draft: args.Draft},
		store.ListChangesetsOpts{},
	)
	if err != nil {
		return nil, err
	}

	return r.bulkOperationByIDString(ctx, bulkGroupID)
}

func (r *Resolver) BatchSpecs(ctx context.Context, args *graphqlbackend.ListBatchSpecArgs) (_ graphqlbackend.BatchSpecConnectionResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.BatchSpecs", fmt.Sprintf("First: %d, After: %v", args.First, args.After))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	if err := validateFirstParamDefaults(args.First); err != nil {
		return nil, err
	}

	opts := store.ListBatchSpecsOpts{
		LimitOpts: store.LimitOpts{
			Limit: int(args.First),
		},
		NewestFirst: true,
	}

	if args.IncludeLocallyExecutedSpecs != nil {
		opts.IncludeLocallyExecutedSpecs = *args.IncludeLocallyExecutedSpecs
	}

	// 🚨 SECURITY: If the user is not an admin, we don't want to include
	// BatchSpecs that were created with CreateBatchSpecFromRaw and not owned
	// by the user
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.store.DatabaseDB()); err != nil {
		opts.ExcludeCreatedFromRawNotOwnedByUser = actor.FromContext(ctx).UID
	}

	if args.After != nil {
		id, err := strconv.Atoi(*args.After)
		if err != nil {
			return nil, err
		}
		opts.Cursor = int64(id)
	}

	return &batchSpecConnectionResolver{store: r.store, opts: opts}, nil
}

func (r *Resolver) CreateEmptyBatchChange(ctx context.Context, args *graphqlbackend.CreateEmptyBatchChangeArgs) (_ graphqlbackend.BatchChangeResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.CreateEmptyBatchChange", fmt.Sprintf("Namespace: %+v", args.Namespace))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	svc := service.New(r.store)

	var uid, oid int32
	if err := graphqlbackend.UnmarshalNamespaceID(args.Namespace, &uid, &oid); err != nil {
		return nil, err
	}

	batchChange, err := svc.CreateEmptyBatchChange(ctx, service.CreateEmptyBatchChangeOpts{
		NamespaceUserID: uid,
		NamespaceOrgID:  oid,
		Name:            args.Name,
	})

	if err != nil {
		return nil, err
	}

	return &batchChangeResolver{store: r.store, batchChange: batchChange}, nil
}

func (r *Resolver) UpsertEmptyBatchChange(ctx context.Context, args *graphqlbackend.UpsertEmptyBatchChangeArgs) (_ graphqlbackend.BatchChangeResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.UpsertEmptyBatchChange", fmt.Sprintf("Namespace: %s", args.Namespace))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	svc := service.New(r.store)

	var uid, oid int32
	if err := graphqlbackend.UnmarshalNamespaceID(args.Namespace, &uid, &oid); err != nil {
		return nil, err
	}

	batchChange, err := svc.UpsertEmptyBatchChange(ctx, service.UpsertEmptyBatchChangeOpts{
		NamespaceUserID: uid,
		NamespaceOrgID:  oid,
		Name:            args.Name,
	})

	if err != nil {
		return nil, err
	}

	return &batchChangeResolver{store: r.store, batchChange: batchChange}, nil
}

func (r *Resolver) CreateBatchSpecFromRaw(ctx context.Context, args *graphqlbackend.CreateBatchSpecFromRawArgs) (_ graphqlbackend.BatchSpecResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.CreateBatchSpecFromRaw", fmt.Sprintf("Namespace: %+v", args.Namespace))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if err := batchChangesCreateAccess(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	svc := service.New(r.store)

	var uid, oid int32
	if err := graphqlbackend.UnmarshalNamespaceID(args.Namespace, &uid, &oid); err != nil {
		return nil, err
	}

	bid, err := unmarshalBatchChangeID(args.BatchChange)
	if err != nil {
		return nil, err
	}

	batchSpec, err := svc.CreateBatchSpecFromRaw(ctx, service.CreateBatchSpecFromRawOpts{
		NamespaceUserID:  uid,
		NamespaceOrgID:   oid,
		RawSpec:          args.BatchSpec,
		AllowIgnored:     args.AllowIgnored,
		AllowUnsupported: args.AllowUnsupported,
		NoCache:          args.NoCache,
		BatchChange:      bid,
	})
	if err != nil {
		return nil, err
	}

	return &batchSpecResolver{store: r.store, batchSpec: batchSpec}, nil
}

func (r *Resolver) ExecuteBatchSpec(ctx context.Context, args *graphqlbackend.ExecuteBatchSpecArgs) (_ graphqlbackend.BatchSpecResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.ExecuteBatchSpec", fmt.Sprintf("BatchSpec: %+v", args.BatchSpec))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	batchSpecRandID, err := unmarshalBatchSpecID(args.BatchSpec)
	if err != nil {
		return nil, err
	}

	if batchSpecRandID == "" {
		return nil, ErrIDIsZero{}
	}

	// 🚨 SECURITY: ExecuteBatchSpec checks whether current user is authorized
	// and has access to namespace.
	// Right now we also only allow creating batch specs in a user-namespace,
	// so the check makes sure the current user is the creator of the batch
	// spec or an admin.
	svc := service.New(r.store)
	batchSpec, err := svc.ExecuteBatchSpec(ctx, service.ExecuteBatchSpecOpts{
		BatchSpecRandID: batchSpecRandID,
		// TODO: args not yet implemented: NoCache, AutoApply
	})
	if err != nil {
		return nil, err
	}

	return &batchSpecResolver{store: r.store, batchSpec: batchSpec}, nil
}

func (r *Resolver) CancelBatchSpecExecution(ctx context.Context, args *graphqlbackend.CancelBatchSpecExecutionArgs) (_ graphqlbackend.BatchSpecResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.CancelBatchSpecExecution", fmt.Sprintf("BatchSpec: %+v", args.BatchSpec))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	batchSpecRandID, err := unmarshalBatchSpecID(args.BatchSpec)
	if err != nil {
		return nil, err
	}

	if batchSpecRandID == "" {
		return nil, ErrIDIsZero{}
	}

	svc := service.New(r.store)
	batchSpec, err := svc.CancelBatchSpec(ctx, service.CancelBatchSpecOpts{
		BatchSpecRandID: batchSpecRandID,
	})
	if err != nil {
		return nil, err
	}

	return &batchSpecResolver{store: r.store, batchSpec: batchSpec}, nil
}

func (r *Resolver) RetryBatchSpecWorkspaceExecution(ctx context.Context, args *graphqlbackend.RetryBatchSpecWorkspaceExecutionArgs) (_ *graphqlbackend.EmptyResponse, err error) {
	tr, ctx := trace.New(ctx, "Resolver.RetryBatchSpecWorkspaceExecution", fmt.Sprintf("Workspaces: %+v", args.BatchSpecWorkspaces))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	var workspaceIDs []int64
	for _, raw := range args.BatchSpecWorkspaces {
		id, err := unmarshalBatchSpecWorkspaceID(raw)
		if err != nil {
			return nil, err
		}

		if id == 0 {
			return nil, ErrIDIsZero{}
		}

		workspaceIDs = append(workspaceIDs, id)
	}

	// 🚨 SECURITY: RetryBatchSpecWorkspaces checks whether current user is authorized
	// and has access to namespace.
	// Right now we also only allow creating batch specs in a user-namespace,
	// so the check makes sure the current user is the creator of the batch
	// spec or an admin.
	svc := service.New(r.store)
	err = svc.RetryBatchSpecWorkspaces(ctx, workspaceIDs)
	if err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) ReplaceBatchSpecInput(ctx context.Context, args *graphqlbackend.ReplaceBatchSpecInputArgs) (_ graphqlbackend.BatchSpecResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.ReplaceBatchSpecInput", fmt.Sprintf("BatchSpec: %+v", args.BatchSpec))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	batchSpecRandID, err := unmarshalBatchSpecID(args.PreviousSpec)
	if err != nil {
		return nil, err
	}

	if batchSpecRandID == "" {
		return nil, ErrIDIsZero{}
	}

	// 🚨 SECURITY: ReplaceBatchSpecInput checks whether current user is authorized
	// and has access to namespace.
	// Right now we also only allow creating batch specs in a user-namespace,
	// so the check makes sure the current user is the creator of the batch
	// spec or an admin.
	svc := service.New(r.store)
	batchSpec, err := svc.ReplaceBatchSpecInput(ctx, service.ReplaceBatchSpecInputOpts{
		BatchSpecRandID:  batchSpecRandID,
		RawSpec:          args.BatchSpec,
		AllowIgnored:     args.AllowIgnored,
		AllowUnsupported: args.AllowUnsupported,
		NoCache:          args.NoCache,
	})
	if err != nil {
		return nil, err
	}

	return &batchSpecResolver{store: r.store, batchSpec: batchSpec}, nil
}

func (r *Resolver) UpsertBatchSpecInput(ctx context.Context, args *graphqlbackend.UpsertBatchSpecInputArgs) (_ graphqlbackend.BatchSpecResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.UpsertBatchSpecInput", fmt.Sprintf("BatchSpec: %+v", args.BatchSpec))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if err := batchChangesCreateAccess(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	svc := service.New(r.store)

	var uid, oid int32
	if err := graphqlbackend.UnmarshalNamespaceID(args.Namespace, &uid, &oid); err != nil {
		return nil, err
	}

	// 🚨 SECURITY: UpsertBatchSpecInput checks whether current user is
	// authorised and has access to the namespace.
	//
	// Right now we also only allow creating batch specs in a user namespace, so
	// the check makes sure the current user is the creator of the batch spec or
	// an admin.
	batchSpec, err := svc.UpsertBatchSpecInput(ctx, service.UpsertBatchSpecInputOpts{
		NamespaceUserID:  uid,
		NamespaceOrgID:   oid,
		RawSpec:          args.BatchSpec,
		AllowIgnored:     args.AllowIgnored,
		AllowUnsupported: args.AllowUnsupported,
		NoCache:          args.NoCache,
	})
	if err != nil {
		return nil, err
	}

	return &batchSpecResolver{store: r.store, batchSpec: batchSpec}, nil
}

func (r *Resolver) CancelBatchSpecWorkspaceExecution(ctx context.Context, args *graphqlbackend.CancelBatchSpecWorkspaceExecutionArgs) (*graphqlbackend.EmptyResponse, error) {
	// TODO(ssbc): currently admin only.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}
	// TODO(ssbc): not implemented
	return nil, errors.New("not implemented yet")
}

func (r *Resolver) RetryBatchSpecExecution(ctx context.Context, args *graphqlbackend.RetryBatchSpecExecutionArgs) (_ graphqlbackend.BatchSpecResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.RetryBatchSpecExecution", fmt.Sprintf("BatchSpec: %+v", args.BatchSpec))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	batchSpecRandID, err := unmarshalBatchSpecID(args.BatchSpec)
	if err != nil {
		return nil, err
	}

	if batchSpecRandID == "" {
		return nil, ErrIDIsZero{}
	}

	// 🚨 SECURITY: RetryBatchSpecExecution checks whether current user is authorized
	// and has access to namespace.
	svc := service.New(r.store)
	if err = svc.RetryBatchSpecExecution(ctx, service.RetryBatchSpecExecutionOpts{
		BatchSpecRandID:  batchSpecRandID,
		IncludeCompleted: args.IncludeCompleted,
	}); err != nil {
		return nil, err
	}

	return r.batchSpecByID(ctx, args.BatchSpec)
}

func (r *Resolver) EnqueueBatchSpecWorkspaceExecution(ctx context.Context, args *graphqlbackend.EnqueueBatchSpecWorkspaceExecutionArgs) (*graphqlbackend.EmptyResponse, error) {
	// TODO(ssbc): currently admin only.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}
	// TODO(ssbc): not implemented
	return nil, errors.New("not implemented yet")
}

func (r *Resolver) ToggleBatchSpecAutoApply(ctx context.Context, args *graphqlbackend.ToggleBatchSpecAutoApplyArgs) (graphqlbackend.BatchSpecResolver, error) {
	// TODO(ssbc): currently admin only.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}
	// TODO(ssbc): not implemented
	return nil, errors.New("not implemented yet")
}

func (r *Resolver) DeleteBatchSpec(ctx context.Context, args *graphqlbackend.DeleteBatchSpecArgs) (*graphqlbackend.EmptyResponse, error) {
	// TODO(ssbc): currently admin only.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}
	// TODO(ssbc): not implemented
	return nil, errors.New("not implemented yet")
}

func (r *Resolver) AvailableBulkOperations(ctx context.Context, args *graphqlbackend.AvailableBulkOperationsArgs) (availableBulkOperations []string, err error) {
	tr, ctx := trace.New(ctx, "Resolver.AvailableBulkOperations", fmt.Sprintf("BatchChange: %q, len(Changesets): %d", args.BatchChange, len(args.Changesets)))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if err := enterprise.BatchChangesEnabledForUser(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	if len(args.Changesets) == 0 {
		return nil, errors.New("no changesets provided")
	}

	unmarshalledBatchChangeID, err := unmarshalBatchChangeID(args.BatchChange)
	if err != nil {
		return nil, err
	}

	changesetIDs := make([]int64, 0, len(args.Changesets))
	for _, changesetID := range args.Changesets {
		unmarshalledChangesetID, err := unmarshalChangesetID(changesetID)
		if err != nil {
			return nil, err
		}

		changesetIDs = append(changesetIDs, unmarshalledChangesetID)
	}

	svc := service.New(r.store)
	availableBulkOperations, err = svc.GetAvailableBulkOperations(ctx, service.GetAvailableBulkOperationsOpts{
		BatchChange: unmarshalledBatchChangeID,
		Changesets:  changesetIDs,
	})

	if err != nil {
		return nil, err
	}

	return availableBulkOperations, nil
}

func (r *Resolver) CheckBatchChangesCredential(ctx context.Context, args *graphqlbackend.CheckBatchChangesCredentialArgs) (_ *graphqlbackend.EmptyResponse, err error) {
	tr, ctx := trace.New(ctx, "Resolver.CheckBatchChangesCredential", fmt.Sprintf("Credential: %q", args.BatchChangesCredential))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	cred, err := r.batchChangesCredentialByID(ctx, args.BatchChangesCredential)
	if err != nil {
		return nil, err
	}
	if cred == nil {
		return nil, ErrIDIsZero{}
	}

	a, err := cred.authenticator(ctx)
	if err != nil {
		return nil, err
	}

	svc := service.New(r.store)
	if err := svc.ValidateAuthenticator(ctx, cred.ExternalServiceURL(), extsvc.KindToType(cred.ExternalServiceKind()), a); err != nil {
		return nil, &ErrVerifyCredentialFailed{SourceErr: err}
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) MaxUnlicensedChangesets(ctx context.Context) int32 {
	return maxUnlicensedChangesets
}

func parseBatchChangeStates(ss *[]string) ([]btypes.BatchChangeState, error) {
	states := []btypes.BatchChangeState{}
	if ss == nil || len(*ss) == 0 {
		return states, nil
	}
	for _, s := range *ss {
		state, err := parseBatchChangeState(&s)
		if err != nil {
			return nil, err
		}
		if state != "" {
			states = append(states, state)
		}
	}
	return states, nil
}

func parseBatchChangeState(s *string) (btypes.BatchChangeState, error) {
	if s == nil {
		return "", nil
	}
	switch *s {
	case "OPEN":
		return btypes.BatchChangeStateOpen, nil
	case "CLOSED":
		return btypes.BatchChangeStateClosed, nil
	case "DRAFT":
		return btypes.BatchChangeStateDraft, nil
	default:
		return "", errors.Errorf("unknown state %q", *s)
	}
}

func checkSiteAdminOrSameUser(ctx context.Context, db database.DB, userID int32) (bool, error) {
	// 🚨 SECURITY: Only site admins or the authors of a batch change have batch change
	// admin rights.
	if err := backend.CheckSiteAdminOrSameUser(ctx, db, userID); err != nil {
		if errors.HasType(err, &backend.InsufficientAuthorizationError{}) {
			return false, nil
		}

		return false, err
	}
	return true, nil
}

func validateFirstParam(first int32, max int) error {
	if first < 0 || first > int32(max) {
		return ErrInvalidFirstParameter{Min: 0, Max: max, First: int(first)}
	}
	return nil
}

const defaultMaxFirstParam = 10000

func validateFirstParamDefaults(first int32) error {
	return validateFirstParam(first, defaultMaxFirstParam)
}

func unmarshalBulkOperationBaseArgs(args graphqlbackend.BulkOperationBaseArgs) (batchChangeID int64, changesetIDs []int64, err error) {
	batchChangeID, err = unmarshalBatchChangeID(args.BatchChange)
	if err != nil {
		return 0, nil, err
	}

	if batchChangeID == 0 {
		return 0, nil, ErrIDIsZero{}
	}

	for _, raw := range args.Changesets {
		id, err := unmarshalChangesetID(raw)
		if err != nil {
			return 0, nil, err
		}

		if id == 0 {
			return 0, nil, ErrIDIsZero{}
		}

		changesetIDs = append(changesetIDs, id)
	}

	if len(changesetIDs) == 0 {
		return 0, nil, errors.New("specify at least one changeset")
	}

	return batchChangeID, changesetIDs, nil
}
