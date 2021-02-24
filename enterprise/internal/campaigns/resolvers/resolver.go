package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	cauth "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/search"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/service"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
)

// Resolver is the GraphQL resolver of all things related to Campaigns.
type Resolver struct {
	store *store.Store
}

// New returns a new Resolver whose store uses the given database
func New(store *store.Store) graphqlbackend.CampaignsResolver {
	return &Resolver{store: store}
}

func campaignsEnabled(ctx context.Context) error {
	// On Sourcegraph.com nobody can read/create campaign entities
	if envvar.SourcegraphDotComMode() {
		return ErrCampaignsDotCom{}
	}

	if enabled := conf.CampaignsEnabled(); enabled {
		if conf.Get().CampaignsRestrictToAdmins && backend.CheckCurrentUserIsSiteAdmin(ctx) != nil {
			return ErrCampaignsDisabledForUser{}
		}
		return nil
	}

	return ErrCampaignsDisabled{}
}

// campaignsCreateAccess returns true if the current user can create
// campaigns/changesetSpecs/campaignSpecs.
func campaignsCreateAccess(ctx context.Context) error {
	// On Sourcegraph.com nobody can create campaigns/patchsets/changesets
	if envvar.SourcegraphDotComMode() {
		return ErrCampaignsDotCom{}
	}

	act := actor.FromContext(ctx)
	if !act.IsAuthenticated() {
		return backend.ErrNotAuthenticated
	}
	return nil
}

// checkLicense returns a user-facing error if the campaigns feature is not purchased
// with the current license or any error occurred while validating the license.
func checkLicense() error {
	if err := licensing.Check(licensing.FeatureCampaigns); err != nil {
		if licensing.IsFeatureNotActivated(err) {
			return err
		}
		return errors.New("Unable to check license feature, please refer to logs for actual error message.")
	}
	return nil
}

// maxUnlicensedChangesets is the maximum number of changesets that can be
// attached to a campaign when Sourcegraph is unlicensed or the campaign feature
// is disabled.
const maxUnlicensedChangesets = 5

func (r *Resolver) ChangesetByID(ctx context.Context, id graphql.ID) (graphqlbackend.ChangesetResolver, error) {
	if err := campaignsEnabled(ctx); err != nil {
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

func (r *Resolver) CampaignByID(ctx context.Context, id graphql.ID) (graphqlbackend.CampaignResolver, error) {
	if err := campaignsEnabled(ctx); err != nil {
		return nil, err
	}

	campaignID, err := unmarshalCampaignID(id)
	if err != nil {
		return nil, err
	}

	if campaignID == 0 {
		return nil, nil
	}

	campaign, err := r.store.GetCampaign(ctx, store.GetCampaignOpts{ID: campaignID})
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &campaignResolver{store: r.store, Campaign: campaign}, nil
}

func (r *Resolver) Campaign(ctx context.Context, args *graphqlbackend.CampaignArgs) (graphqlbackend.CampaignResolver, error) {
	if err := campaignsEnabled(ctx); err != nil {
		return nil, err
	}

	opts := store.GetCampaignOpts{Name: args.Name}

	err := graphqlbackend.UnmarshalNamespaceID(graphql.ID(args.Namespace), &opts.NamespaceUserID, &opts.NamespaceOrgID)
	if err != nil {
		return nil, err
	}

	campaign, err := r.store.GetCampaign(ctx, opts)
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &campaignResolver{store: r.store, Campaign: campaign}, nil
}

func (r *Resolver) CampaignSpecByID(ctx context.Context, id graphql.ID) (graphqlbackend.CampaignSpecResolver, error) {
	if err := campaignsEnabled(ctx); err != nil {
		return nil, err
	}

	campaignSpecRandID, err := unmarshalCampaignSpecID(id)
	if err != nil {
		return nil, err
	}

	if campaignSpecRandID == "" {
		return nil, nil
	}

	opts := store.GetCampaignSpecOpts{RandID: campaignSpecRandID}
	campaignSpec, err := r.store.GetCampaignSpec(ctx, opts)
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &campaignSpecResolver{store: r.store, campaignSpec: campaignSpec}, nil
}

func (r *Resolver) ChangesetSpecByID(ctx context.Context, id graphql.ID) (graphqlbackend.ChangesetSpecResolver, error) {
	if err := campaignsEnabled(ctx); err != nil {
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

func (r *Resolver) CampaignsCredentialByID(ctx context.Context, id graphql.ID) (graphqlbackend.CampaignsCredentialResolver, error) {
	if err := campaignsEnabled(ctx); err != nil {
		return nil, err
	}

	dbID, err := unmarshalCampaignsCredentialID(id)
	if err != nil {
		return nil, err
	}

	if dbID == 0 {
		return nil, nil
	}

	cred, err := r.store.UserCredentials().GetByID(ctx, dbID)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	if err := backend.CheckSiteAdminOrSameUser(ctx, cred.UserID); err != nil {
		return nil, err
	}

	return &campaignsCredentialResolver{credential: cred}, nil
}

func (r *Resolver) CreateCampaign(ctx context.Context, args *graphqlbackend.CreateCampaignArgs) (graphqlbackend.CampaignResolver, error) {
	var err error
	tr, _ := trace.New(ctx, "Resolver.CreateCampaign", fmt.Sprintf("CampaignSpec %s", args.CampaignSpec))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if err := campaignsEnabled(ctx); err != nil {
		return nil, err
	}

	opts := service.ApplyCampaignOpts{
		// This is what differentiates CreateCampaign from ApplyCampaign
		FailIfCampaignExists: true,
	}

	opts.CampaignSpecRandID, err = unmarshalCampaignSpecID(args.CampaignSpec)
	if err != nil {
		return nil, err
	}

	if opts.CampaignSpecRandID == "" {
		return nil, ErrIDIsZero{}
	}

	svc := service.New(r.store)
	campaign, err := svc.ApplyCampaign(ctx, opts)
	if err != nil {
		if err == service.ErrEnsureCampaignFailed {
			return nil, ErrEnsureCampaignFailed{}
		} else if err == service.ErrApplyClosedCampaign {
			return nil, ErrApplyClosedCampaign{}
		} else if err == service.ErrMatchingCampaignExists {
			return nil, ErrMatchingCampaignExists{}
		}
		return nil, err
	}

	return &campaignResolver{store: r.store, Campaign: campaign}, nil
}

func (r *Resolver) ApplyCampaign(ctx context.Context, args *graphqlbackend.ApplyCampaignArgs) (graphqlbackend.CampaignResolver, error) {
	var err error
	tr, ctx := trace.New(ctx, "Resolver.ApplyCampaign", fmt.Sprintf("CampaignSpec %s", args.CampaignSpec))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if err := campaignsEnabled(ctx); err != nil {
		return nil, err
	}

	opts := service.ApplyCampaignOpts{}

	opts.CampaignSpecRandID, err = unmarshalCampaignSpecID(args.CampaignSpec)
	if err != nil {
		return nil, err
	}

	if opts.CampaignSpecRandID == "" {
		return nil, ErrIDIsZero{}
	}

	if args.EnsureCampaign != nil {
		opts.EnsureCampaignID, err = unmarshalCampaignID(*args.EnsureCampaign)
		if err != nil {
			return nil, err
		}
	}

	svc := service.New(r.store)
	// 🚨 SECURITY: ApplyCampaign checks whether the user has permission to
	// apply the campaign spec
	campaign, err := svc.ApplyCampaign(ctx, opts)
	if err != nil {
		if err == service.ErrEnsureCampaignFailed {
			return nil, ErrEnsureCampaignFailed{}
		} else if err == service.ErrApplyClosedCampaign {
			return nil, ErrApplyClosedCampaign{}
		} else if err == service.ErrMatchingCampaignExists {
			return nil, ErrMatchingCampaignExists{}
		}
		return nil, err
	}

	return &campaignResolver{store: r.store, Campaign: campaign}, nil
}

func (r *Resolver) CreateCampaignSpec(ctx context.Context, args *graphqlbackend.CreateCampaignSpecArgs) (graphqlbackend.CampaignSpecResolver, error) {
	var err error
	tr, ctx := trace.New(ctx, "Resolver.CreateCampaignSpec", fmt.Sprintf("Namespace %s, Spec %q", args.Namespace, args.CampaignSpec))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if err := campaignsEnabled(ctx); err != nil {
		return nil, err
	}

	if err := campaignsCreateAccess(ctx); err != nil {
		return nil, err
	}

	if err := checkLicense(); err != nil {
		if licensing.IsFeatureNotActivated(err) {
			if len(args.ChangesetSpecs) > maxUnlicensedChangesets {
				return nil, ErrCampaignsUnlicensed{err}
			}
		} else {
			return nil, err
		}
	}

	opts := service.CreateCampaignSpecOpts{RawSpec: args.CampaignSpec}

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
	campaignSpec, err := svc.CreateCampaignSpec(ctx, opts)
	if err != nil {
		return nil, err
	}

	if err := logCampaignSpecCreated(ctx, r.store.DB(), &opts); err != nil {
		return nil, err
	}

	specResolver := &campaignSpecResolver{
		store:        r.store,
		campaignSpec: campaignSpec,
	}

	return specResolver, nil
}

func logCampaignSpecCreated(ctx context.Context, db dbutil.DB, opts *service.CreateCampaignSpecOpts) error {
	// Log an analytics event when a CampaignSpec has been created.
	// See internal/usagestats/campaigns.go.
	actor := actor.FromContext(ctx)

	type eventArg struct {
		ChangesetSpecsCount int `json:"changeset_specs_count"`
	}
	arg := eventArg{ChangesetSpecsCount: len(opts.ChangesetSpecRandIDs)}

	jsonArg, err := json.Marshal(arg)
	if err != nil {
		return err
	}

	return usagestats.LogBackendEvent(db, actor.UID, "CampaignSpecCreated", json.RawMessage(jsonArg))
}

func (r *Resolver) CreateChangesetSpec(ctx context.Context, args *graphqlbackend.CreateChangesetSpecArgs) (graphqlbackend.ChangesetSpecResolver, error) {
	var err error
	tr, ctx := trace.New(ctx, "Resolver.CreateChangesetSpec", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if err := campaignsEnabled(ctx); err != nil {
		return nil, err
	}

	if err := campaignsCreateAccess(ctx); err != nil {
		return nil, err
	}

	act := actor.FromContext(ctx)
	// Actor MUST be logged in at this stage, because campaignsCreateAccess checks that already.
	// To be extra safe, we'll just do the cheap check again here so if anyone ever modifies
	// campaignsCreateAccess, we still enforce it here.
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

func (r *Resolver) MoveCampaign(ctx context.Context, args *graphqlbackend.MoveCampaignArgs) (graphqlbackend.CampaignResolver, error) {
	var err error
	tr, ctx := trace.New(ctx, "Resolver.MoveCampaign", fmt.Sprintf("Campaign %s", args.Campaign))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if err := campaignsEnabled(ctx); err != nil {
		return nil, err
	}

	campaignID, err := unmarshalCampaignID(args.Campaign)
	if err != nil {
		return nil, err
	}

	if campaignID == 0 {
		return nil, ErrIDIsZero{}
	}

	opts := service.MoveCampaignOpts{
		CampaignID: campaignID,
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
	// 🚨 SECURITY: MoveCampaign checks whether the current user is authorized.
	campaign, err := svc.MoveCampaign(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &campaignResolver{store: r.store, Campaign: campaign}, nil
}

func (r *Resolver) DeleteCampaign(ctx context.Context, args *graphqlbackend.DeleteCampaignArgs) (_ *graphqlbackend.EmptyResponse, err error) {
	tr, ctx := trace.New(ctx, "Resolver.DeleteCampaign", fmt.Sprintf("Campaign: %q", args.Campaign))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	if err := campaignsEnabled(ctx); err != nil {
		return nil, err
	}

	campaignID, err := unmarshalCampaignID(args.Campaign)
	if err != nil {
		return nil, err
	}

	if campaignID == 0 {
		return nil, ErrIDIsZero{}
	}

	svc := service.New(r.store)
	// 🚨 SECURITY: DeleteCampaign checks whether current user is authorized.
	err = svc.DeleteCampaign(ctx, campaignID)
	return &graphqlbackend.EmptyResponse{}, err
}

func (r *Resolver) Campaigns(ctx context.Context, args *graphqlbackend.ListCampaignsArgs) (graphqlbackend.CampaignsConnectionResolver, error) {
	if err := campaignsEnabled(ctx); err != nil {
		return nil, err
	}

	opts := store.ListCampaignsOpts{}

	state, err := parseCampaignState(args.State)
	if err != nil {
		return nil, err
	}
	opts.State = state
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

	authErr := backend.CheckCurrentUserIsSiteAdmin(ctx)
	if authErr != nil && authErr != backend.ErrMustBeSiteAdmin {
		return nil, authErr
	}
	isSiteAdmin := authErr != backend.ErrMustBeSiteAdmin
	if !isSiteAdmin {
		if args.ViewerCanAdminister != nil && *args.ViewerCanAdminister {
			actor := actor.FromContext(ctx)
			opts.InitialApplierID = actor.UID
		}
	}

	if args.Namespace != nil {
		err := graphqlbackend.UnmarshalNamespaceID(*args.Namespace, &opts.NamespaceUserID, &opts.NamespaceOrgID)
		if err != nil {
			return nil, err
		}
	}

	return &campaignsConnectionResolver{
		store: r.store,
		opts:  opts,
	}, nil
}

func (r *Resolver) CampaignsCodeHosts(ctx context.Context, args *graphqlbackend.ListCampaignsCodeHostsArgs) (graphqlbackend.CampaignsCodeHostConnectionResolver, error) {
	if err := campaignsEnabled(ctx); err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only viewable for self or by site admins.
	if err := backend.CheckSiteAdminOrSameUser(ctx, args.UserID); err != nil {
		return nil, err
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

	return &campaignsCodeHostConnectionResolver{userID: args.UserID, limitOffset: limitOffset, store: r.store}, nil
}

// listChangesetOptsFromArgs turns the graphqlbackend.ListChangesetsArgs into
// ListChangesetsOpts.
// If the args do not include a filter that would reveal sensitive information
// about a changeset the user doesn't have access to, the second return value
// is false.
func listChangesetOptsFromArgs(args *graphqlbackend.ListChangesetsArgs, campaignID int64) (opts store.ListChangesetsOpts, optsSafe bool, err error) {
	if args == nil {
		return opts, true, nil
	}

	safe := true

	// TODO: This _could_ become problematic if a user has a campaign with > 10000 changesets, once
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

	if args.State != nil {
		state := *args.State
		if !state.Valid() {
			return opts, false, errors.New("changeset state not valid")
		}

		switch state {
		case campaigns.ChangesetStateOpen:
			externalState := campaigns.ChangesetExternalStateOpen
			publicationState := campaigns.ChangesetPublicationStatePublished
			opts.ExternalState = &externalState
			opts.ReconcilerStates = []campaigns.ReconcilerState{campaigns.ReconcilerStateCompleted}
			opts.PublicationState = &publicationState
		case campaigns.ChangesetStateDraft:
			externalState := campaigns.ChangesetExternalStateDraft
			publicationState := campaigns.ChangesetPublicationStatePublished
			opts.ExternalState = &externalState
			opts.ReconcilerStates = []campaigns.ReconcilerState{campaigns.ReconcilerStateCompleted}
			opts.PublicationState = &publicationState
		case campaigns.ChangesetStateClosed:
			externalState := campaigns.ChangesetExternalStateClosed
			publicationState := campaigns.ChangesetPublicationStatePublished
			opts.ExternalState = &externalState
			opts.ReconcilerStates = []campaigns.ReconcilerState{campaigns.ReconcilerStateCompleted}
			opts.PublicationState = &publicationState
		case campaigns.ChangesetStateMerged:
			externalState := campaigns.ChangesetExternalStateMerged
			publicationState := campaigns.ChangesetPublicationStatePublished
			opts.ExternalState = &externalState
			opts.ReconcilerStates = []campaigns.ReconcilerState{campaigns.ReconcilerStateCompleted}
			opts.PublicationState = &publicationState
		case campaigns.ChangesetStateDeleted:
			externalState := campaigns.ChangesetExternalStateDeleted
			publicationState := campaigns.ChangesetPublicationStatePublished
			opts.ExternalState = &externalState
			opts.ReconcilerStates = []campaigns.ReconcilerState{campaigns.ReconcilerStateCompleted}
			opts.PublicationState = &publicationState
		case campaigns.ChangesetStateUnpublished:
			publicationState := campaigns.ChangesetPublicationStateUnpublished
			opts.ReconcilerStates = []campaigns.ReconcilerState{campaigns.ReconcilerStateCompleted}
			opts.PublicationState = &publicationState
		case campaigns.ChangesetStateProcessing:
			opts.ReconcilerStates = []campaigns.ReconcilerState{campaigns.ReconcilerStateQueued, campaigns.ReconcilerStateProcessing}
		case campaigns.ChangesetStateRetrying:
			opts.ReconcilerStates = []campaigns.ReconcilerState{campaigns.ReconcilerStateErrored}
		case campaigns.ChangesetStateFailed:
			opts.ReconcilerStates = []campaigns.ReconcilerState{campaigns.ReconcilerStateFailed}
		default:
			return opts, false, errors.Errorf("changeset state %q not supported in filtering", state)
		}
	}

	if args.ReviewState != nil {
		state := *args.ReviewState
		if !state.Valid() {
			return opts, false, errors.New("changeset review state not valid")
		}
		opts.ExternalReviewState = &state
		// If the user filters by ReviewState we cannot include hidden
		// changesets, since that would leak information.
		safe = false
	}
	if args.CheckState != nil {
		state := *args.CheckState
		if !state.Valid() {
			return opts, false, errors.New("changeset check state not valid")
		}
		opts.ExternalCheckState = &state
		// If the user filters by CheckState we cannot include hidden
		// changesets, since that would leak information.
		safe = false
	}
	if args.OnlyPublishedByThisCampaign != nil {
		published := campaigns.ChangesetPublicationStatePublished

		opts.OwnedByCampaignID = campaignID
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

	return opts, safe, nil
}

func (r *Resolver) CloseCampaign(ctx context.Context, args *graphqlbackend.CloseCampaignArgs) (_ graphqlbackend.CampaignResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.CloseCampaign", fmt.Sprintf("Campaign: %q", args.Campaign))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if err := campaignsEnabled(ctx); err != nil {
		return nil, err
	}

	campaignID, err := unmarshalCampaignID(args.Campaign)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling campaign id")
	}

	if campaignID == 0 {
		return nil, ErrIDIsZero{}
	}

	svc := service.New(r.store)
	// 🚨 SECURITY: CloseCampaign checks whether current user is authorized.
	campaign, err := svc.CloseCampaign(ctx, campaignID, args.CloseChangesets)
	if err != nil {
		return nil, errors.Wrap(err, "closing campaign")
	}

	return &campaignResolver{store: r.store, Campaign: campaign}, nil
}

func (r *Resolver) SyncChangeset(ctx context.Context, args *graphqlbackend.SyncChangesetArgs) (_ *graphqlbackend.EmptyResponse, err error) {
	tr, ctx := trace.New(ctx, "Resolver.SyncChangeset", fmt.Sprintf("Changeset: %q", args.Changeset))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	if err := campaignsEnabled(ctx); err != nil {
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
	if err := campaignsEnabled(ctx); err != nil {
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

func (r *Resolver) CreateCampaignsCredential(ctx context.Context, args *graphqlbackend.CreateCampaignsCredentialArgs) (_ graphqlbackend.CampaignsCredentialResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.CreateCampaignsCredential", fmt.Sprintf("%q (%q)", args.ExternalServiceKind, args.ExternalServiceURL))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	if err := campaignsEnabled(ctx); err != nil {
		return nil, err
	}

	userID, err := graphqlbackend.UnmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	if userID == 0 {
		return nil, ErrIDIsZero{}
	}

	// 🚨 SECURITY: Check that the requesting user can create the credential.
	if err := backend.CheckSiteAdminOrSameUser(ctx, userID); err != nil {
		return nil, err
	}

	// Need to validate externalServiceKind, otherwise this'll panic.
	kind, valid := extsvc.ParseServiceKind(args.ExternalServiceKind)
	if !valid {
		return nil, errors.New("invalid external service kind")
	}

	// TODO: Do we want to validate the URL, or even if such an external service exists? Or better, would the DB have a constraint?

	if args.Credential == "" {
		return nil, errors.New("empty credential not allowed")
	}

	scope := database.UserCredentialScope{
		Domain:              database.UserCredentialDomainCampaigns,
		ExternalServiceID:   args.ExternalServiceURL,
		ExternalServiceType: extsvc.KindToType(kind),
		UserID:              userID,
	}

	// Throw error documented in schema.graphql.
	existing, err := r.store.UserCredentials().GetByScope(ctx, scope)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDuplicateCredential{}
	}

	keypair, err := cauth.GenerateRSAKey()
	if err != nil {
		return nil, err
	}

	var a auth.Authenticator
	if kind == extsvc.KindBitbucketServer {
		svc := service.New(r.store)
		username, err := svc.FetchUsernameForBitbucketServerToken(ctx, args.ExternalServiceURL, extsvc.KindToType(kind), args.Credential)
		if err != nil {
			return nil, err
		}
		a = &auth.BasicAuthWithSSH{
			BasicAuth:  auth.BasicAuth{Username: username, Password: args.Credential},
			PrivateKey: keypair.PrivateKey,
			PublicKey:  keypair.PublicKey,
			Passphrase: keypair.Passphrase,
		}
	} else {
		a = &auth.OAuthBearerTokenWithSSH{
			OAuthBearerToken: auth.OAuthBearerToken{Token: args.Credential},
			PrivateKey:       keypair.PrivateKey,
			PublicKey:        keypair.PublicKey,
			Passphrase:       keypair.Passphrase,
		}
	}

	cred, err := r.store.UserCredentials().Create(ctx, scope, a)
	if err != nil {
		return nil, err
	}

	return &campaignsCredentialResolver{credential: cred}, nil
}

func (r *Resolver) DeleteCampaignsCredential(ctx context.Context, args *graphqlbackend.DeleteCampaignsCredentialArgs) (_ *graphqlbackend.EmptyResponse, err error) {
	tr, ctx := trace.New(ctx, "Resolver.DeleteCampaignsCredential", fmt.Sprintf("Credential: %q", args.CampaignsCredential))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	if err := campaignsEnabled(ctx); err != nil {
		return nil, err
	}

	dbID, err := unmarshalCampaignsCredentialID(args.CampaignsCredential)
	if err != nil {
		return nil, err
	}

	if dbID == 0 {
		return nil, ErrIDIsZero{}
	}

	// Get existing credential.
	cred, err := r.store.UserCredentials().GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check that the requesting user may delete the credential.
	if err := backend.CheckSiteAdminOrSameUser(ctx, cred.UserID); err != nil {
		return nil, err
	}

	// This also fails if the credential was not found.
	if err := r.store.UserCredentials().Delete(ctx, dbID); err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func parseCampaignState(s *string) (campaigns.CampaignState, error) {
	if s == nil {
		return campaigns.CampaignStateAny, nil
	}
	switch *s {
	case "OPEN":
		return campaigns.CampaignStateOpen, nil
	case "CLOSED":
		return campaigns.CampaignStateClosed, nil
	default:
		return campaigns.CampaignStateAny, fmt.Errorf("unknown state %q", *s)
	}
}

func checkSiteAdminOrSameUser(ctx context.Context, userID int32) (bool, error) {
	// 🚨 SECURITY: Only site admins or the authors of a campaign have campaign
	// admin rights.
	if err := backend.CheckSiteAdminOrSameUser(ctx, userID); err != nil {
		if _, ok := err.(*backend.InsufficientAuthorizationError); ok {
			return false, nil
		}

		return false, err
	}
	return true, nil
}

type ErrInvalidFirstParameter struct {
	Min, Max, First int
}

func (e ErrInvalidFirstParameter) Error() string {
	return fmt.Sprintf("first param %d is out of range (min=%d, max=%d)", e.First, e.Min, e.Max)
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
