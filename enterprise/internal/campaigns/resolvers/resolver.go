package resolvers

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var ErrIDIsZero = errors.New("invalid node id")

// Resolver is the GraphQL resolver of all things related to Campaigns.
type Resolver struct {
	store       *ee.Store
	httpFactory *httpcli.Factory
}

// NewResolver returns a new Resolver whose store uses the given db
func NewResolver(db *sql.DB) graphqlbackend.CampaignsResolver {
	return &Resolver{store: ee.NewStore(db)}
}

func allowReadAccess(ctx context.Context) error {
	// 🚨 SECURITY: Only site admins or users when read-access is enabled may access changesets.
	if readAccess := conf.CampaignsReadAccessEnabled(); readAccess {
		return nil
	}

	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return err
	}

	return nil
}

func (r *Resolver) ChangesetByID(ctx context.Context, id graphql.ID) (graphqlbackend.ChangesetResolver, error) {
	// 🚨 SECURITY: Only site admins or users when read-access is enabled may access changesets.
	if err := allowReadAccess(ctx); err != nil {
		return nil, err
	}

	changesetID, err := unmarshalChangesetID(id)
	if err != nil {
		return nil, err
	}

	if changesetID == 0 {
		return nil, nil
	}

	changeset, err := r.store.GetChangeset(ctx, ee.GetChangesetOpts{ID: changesetID})
	if err != nil {
		if err == ee.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	// 🚨 SECURITY: db.Repos.Get uses the authzFilter under the hood and
	// filters out repositories that the user doesn't have access to.
	repo, err := db.Repos.Get(ctx, changeset.RepoID)
	if err != nil {
		if errcode.IsNotFound(err) {
			// TODO: nextSyncAt is not populated. See https://github.com/sourcegraph/sourcegraph/issues/11227
			return &hiddenChangesetResolver{
				store:       r.store,
				httpFactory: r.httpFactory,
				Changeset:   changeset,
			}, nil
		}
		return nil, err
	}

	return &changesetResolver{
		// TODO: nextSyncAt is not populated. See https://github.com/sourcegraph/sourcegraph/issues/11227
		store:         r.store,
		httpFactory:   r.httpFactory,
		Changeset:     changeset,
		preloadedRepo: repo,
	}, nil
}

func (r *Resolver) CampaignByID(ctx context.Context, id graphql.ID) (graphqlbackend.CampaignResolver, error) {
	// 🚨 SECURITY: Only site admins or users when read-access is enabled may access campaign.
	if err := allowReadAccess(ctx); err != nil {
		return nil, err
	}

	campaignID, err := campaigns.UnmarshalCampaignID(id)
	if err != nil {
		return nil, err
	}

	if campaignID == 0 {
		return nil, nil
	}

	campaign, err := r.store.GetCampaign(ctx, ee.GetCampaignOpts{ID: campaignID})
	if err != nil {
		if err == ee.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &campaignResolver{store: r.store, httpFactory: r.httpFactory, Campaign: campaign}, nil
}

func (r *Resolver) PatchByID(ctx context.Context, id graphql.ID) (graphqlbackend.PatchInterfaceResolver, error) {
	// 🚨 SECURITY: Only site admins or users when read-access is enabled may access patches.
	if err := allowReadAccess(ctx); err != nil {
		return nil, err
	}

	patchID, err := unmarshalPatchID(id)
	if err != nil {
		return nil, err
	}

	if patchID == 0 {
		return nil, nil
	}

	patch, err := r.store.GetPatch(ctx, ee.GetPatchOpts{ID: patchID})
	if err != nil {
		if err == ee.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	// 🚨 SECURITY: db.Repos.Get uses the authzFilter under the hood and
	// filters out repositories that the user doesn't have access to.
	repo, err := db.Repos.Get(ctx, patch.RepoID)
	if err != nil {
		if errcode.IsNotFound(err) {
			return &hiddenPatchResolver{patch: patch}, nil
		}
		return nil, err
	}

	return &patchResolver{store: r.store, patch: patch, preloadedRepo: repo}, nil
}

func (r *Resolver) PatchSetByID(ctx context.Context, id graphql.ID) (graphqlbackend.PatchSetResolver, error) {
	// 🚨 SECURITY: Only site admins or users when read-access is enabled may access patch sets.
	if err := allowReadAccess(ctx); err != nil {
		return nil, err
	}

	patchSetID, err := unmarshalPatchSetID(id)
	if err != nil {
		return nil, err
	}

	if patchSetID == 0 {
		return nil, nil
	}

	patchSet, err := r.store.GetPatchSet(ctx, ee.GetPatchSetOpts{ID: patchSetID})
	if err != nil {
		if err == ee.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &patchSetResolver{store: r.store, patchSet: patchSet}, nil
}

func (r *Resolver) AddChangesetsToCampaign(ctx context.Context, args *graphqlbackend.AddChangesetsToCampaignArgs) (_ graphqlbackend.CampaignResolver, err error) {
	campaignID, err := campaigns.UnmarshalCampaignID(args.Campaign)
	if err != nil {
		return nil, err
	}

	if campaignID == 0 {
		return nil, nil
	}

	changesetIDs := make([]int64, 0, len(args.Changesets))
	set := map[int64]struct{}{}
	for _, changesetID := range args.Changesets {
		id, err := unmarshalChangesetID(changesetID)
		if err != nil {
			return nil, err
		}
		if id == 0 {
			continue
		}

		if _, ok := set[id]; !ok {
			changesetIDs = append(changesetIDs, id)
			set[id] = struct{}{}
		}
	}

	svc := ee.NewService(r.store, r.httpFactory)
	// 🚨 SECURITY: AddChangesetsToCampaign checks whether current user is authorized.
	campaign, err := svc.AddChangesetsToCampaign(ctx, campaignID, changesetIDs)
	if err != nil {
		return nil, err
	}

	return &campaignResolver{store: r.store, httpFactory: r.httpFactory, Campaign: campaign}, nil
}

func (r *Resolver) CreateCampaign(ctx context.Context, args *graphqlbackend.CreateCampaignArgs) (graphqlbackend.CampaignResolver, error) {
	var err error
	tr, ctx := trace.New(ctx, "Resolver.CreateCampaign", args.Input.Name)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	user, err := db.Users.GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "%v", backend.ErrNotAuthenticated)
	}

	// 🚨 SECURITY: Only site admins may create a campaign for now.
	if !user.SiteAdmin {
		return nil, backend.ErrMustBeSiteAdmin
	}

	campaign := &campaigns.Campaign{
		Name:     args.Input.Name,
		AuthorID: user.ID,
	}

	if args.Input.Description != nil {
		campaign.Description = *args.Input.Description
	}
	if args.Input.Branch != nil {
		campaign.Branch = *args.Input.Branch
	}

	if args.Input.PatchSet != nil {
		patchSetID, err := unmarshalPatchSetID(*args.Input.PatchSet)
		if err != nil {
			return nil, err
		}
		campaign.PatchSetID = patchSetID
	}

	switch relay.UnmarshalKind(args.Input.Namespace) {
	case "User":
		err = relay.UnmarshalSpec(args.Input.Namespace, &campaign.NamespaceUserID)
	case "Org":
		err = relay.UnmarshalSpec(args.Input.Namespace, &campaign.NamespaceOrgID)
	default:
		err = errors.Errorf("Invalid namespace %q", args.Input.Namespace)
	}

	if err != nil {
		return nil, err
	}

	svc := ee.NewService(r.store, r.httpFactory)
	err = svc.CreateCampaign(ctx, campaign)
	if err != nil {
		return nil, err
	}

	return &campaignResolver{store: r.store, httpFactory: r.httpFactory, Campaign: campaign}, nil
}

func (r *Resolver) UpdateCampaign(ctx context.Context, args *graphqlbackend.UpdateCampaignArgs) (_ graphqlbackend.CampaignResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.UpdateCampaign", fmt.Sprintf("Campaign: %q", args.Input.ID))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	campaignID, err := campaigns.UnmarshalCampaignID(args.Input.ID)
	if err != nil {
		return nil, err
	}

	if campaignID == 0 {
		return nil, nil
	}

	updateArgs := ee.UpdateCampaignArgs{Campaign: campaignID}
	updateArgs.Name = args.Input.Name
	updateArgs.Description = args.Input.Description
	updateArgs.Branch = args.Input.Branch

	if args.Input.PatchSet != nil {
		patchSetID, err := unmarshalPatchSetID(*args.Input.PatchSet)
		if err != nil {
			return nil, err
		}
		updateArgs.PatchSet = &patchSetID
	}

	svc := ee.NewService(r.store, r.httpFactory)

	// 🚨 SECURITY: UpdateCampaign checks whether current user is authorized.
	campaign, detachedChangesets, err := svc.UpdateCampaign(ctx, updateArgs)
	if err != nil {
		return nil, err
	}

	if len(detachedChangesets) != 0 {
		go func() {
			ctx := trace.ContextWithTrace(context.Background(), tr)
			err := svc.CloseOpenChangesets(ctx, detachedChangesets)
			if err != nil {
				log15.Error("CloseOpenChangesets", "err", err)
			}
		}()
	}

	return &campaignResolver{store: r.store, httpFactory: r.httpFactory, Campaign: campaign}, nil
}

func (r *Resolver) DeleteCampaign(ctx context.Context, args *graphqlbackend.DeleteCampaignArgs) (_ *graphqlbackend.EmptyResponse, err error) {
	tr, ctx := trace.New(ctx, "Resolver.DeleteCampaign", fmt.Sprintf("Campaign: %q", args.Campaign))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	campaignID, err := campaigns.UnmarshalCampaignID(args.Campaign)
	if err != nil {
		return nil, err
	}

	if campaignID == 0 {
		return nil, ErrIDIsZero
	}

	svc := ee.NewService(r.store, r.httpFactory)
	// 🚨 SECURITY: DeleteCampaign checks whether current user is authorized.
	err = svc.DeleteCampaign(ctx, campaignID, args.CloseChangesets)
	return &graphqlbackend.EmptyResponse{}, err
}

func (r *Resolver) RetryCampaignChangesets(ctx context.Context, args *graphqlbackend.RetryCampaignChangesetsArgs) (graphqlbackend.CampaignResolver, error) {
	var err error
	tr, ctx := trace.New(ctx, "Resolver.RetryCampaignChangesets", fmt.Sprintf("Campaign: %q", args.Campaign))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	campaignID, err := campaigns.UnmarshalCampaignID(args.Campaign)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling campaign id")
	}

	if campaignID == 0 {
		return nil, ErrIDIsZero
	}

	svc := ee.NewService(r.store, r.httpFactory)
	// 🚨 SECURITY: RetryPublishCampaign checks whether current user is authorized.
	campaign, err := svc.RetryPublishCampaign(ctx, campaignID)
	if err != nil {
		return nil, errors.Wrap(err, "publishing campaign")
	}

	return &campaignResolver{store: r.store, httpFactory: r.httpFactory, Campaign: campaign}, nil
}

func (r *Resolver) Campaigns(ctx context.Context, args *graphqlbackend.ListCampaignArgs) (graphqlbackend.CampaignsConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins or users when read-access is enabled may access campaign.
	if err := allowReadAccess(ctx); err != nil {
		return nil, err
	}
	opts := ee.ListCampaignsOpts{
		HasPatchSet: args.HasPatchSet,
	}
	state, err := parseCampaignState(args.State)
	if err != nil {
		return nil, err
	}
	opts.State = state
	if args.First != nil {
		opts.Limit = int(*args.First)
	}
	authErr := backend.CheckCurrentUserIsSiteAdmin(ctx)
	if authErr != nil && authErr != backend.ErrMustBeSiteAdmin {
		return nil, err
	}
	isSiteAdmin := authErr != backend.ErrMustBeSiteAdmin
	if !isSiteAdmin {
		if args.ViewerCanAdminister != nil && *args.ViewerCanAdminister {
			actor := actor.FromContext(ctx)
			opts.OnlyForAuthor = actor.UID
		}
	}
	return &campaignsConnectionResolver{
		store:       r.store,
		httpFactory: r.httpFactory,
		opts:        opts,
	}, nil
}

func (r *Resolver) CreateChangesets(ctx context.Context, args *graphqlbackend.CreateChangesetsArgs) (_ []graphqlbackend.ExternalChangesetResolver, err error) {
	// 🚨 SECURITY: Only site admins may create changesets for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	var repoIDs []api.RepoID
	repoSet := map[api.RepoID]*types.Repo{}
	cs := make([]*campaigns.Changeset, 0, len(args.Input))

	for _, c := range args.Input {
		repoID, err := graphqlbackend.UnmarshalRepositoryID(c.Repository)
		if err != nil {
			return nil, err
		}

		if _, ok := repoSet[repoID]; !ok {
			repoSet[repoID] = nil
			repoIDs = append(repoIDs, repoID)
		}

		cs = append(cs, &campaigns.Changeset{
			RepoID:     repoID,
			ExternalID: c.ExternalID,
		})
	}

	tx, err := r.store.Transact(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "creating transaction")
	}
	defer tx.Done(&err)

	// 🚨 SECURITY: db.Repos.GetByIDs uses the authzFilter under the hood and
	// filters out repositories that the user doesn't have access to.
	rs, err := db.Repos.GetByIDs(ctx, repoIDs...)
	if err != nil {
		return nil, err
	}

	for _, r := range rs {
		if !campaigns.IsRepoSupported(&r.ExternalRepo) {
			err = errors.Errorf(
				"External service type %s of repository %q is currently not supported for use with campaigns",
				r.ExternalRepo.ServiceType,
				r.Name,
			)
			return nil, err
		}

		repoSet[r.ID] = r
	}

	for id, r := range repoSet {
		if r == nil {
			return nil, errors.Errorf("repo %v not found", graphqlbackend.MarshalRepositoryID(api.RepoID(id)))
		}
	}

	for _, c := range cs {
		c.ExternalServiceType = repoSet[c.RepoID].ExternalRepo.ServiceType
	}

	err = tx.CreateChangesets(ctx, cs...)
	if err != nil {
		if _, ok := err.(ee.AlreadyExistError); !ok {
			return nil, err
		}
	}

	repoStore := repos.NewDBStore(tx.DB(), sql.TxOptions{})

	// NOTE: We are performing a blocking sync here in order to ensure
	// that the remote changeset exists and also to remove the possibility
	// of an unsynced changeset entering our database
	if err = ee.SyncChangesets(ctx, repoStore, tx, r.httpFactory, cs...); err != nil {
		return nil, errors.Wrap(err, "syncing changesets")
	}

	csr := make([]graphqlbackend.ExternalChangesetResolver, len(cs))
	for i := range cs {
		csr[i] = &changesetResolver{
			store:         r.store,
			httpFactory:   r.httpFactory,
			Changeset:     cs[i],
			preloadedRepo: repoSet[cs[i].RepoID],
		}
	}

	return csr, nil
}

// listChangesetOptsFromArgs turns the graphqlbackend.ListChangesetsArgs into
// ListChangesetsOpts.
// If the args do not include a filter that would reveal sensitive information
// about a changeset the user doesn't have access to, the second return value
// is false.
func listChangesetOptsFromArgs(args *graphqlbackend.ListChangesetsArgs) (opts ee.ListChangesetsOpts, optsSafe bool, err error) {
	if args == nil {
		return opts, true, nil
	}

	safe := true

	if args.First != nil {
		opts.Limit = int(*args.First)
	}

	if args.State != nil {
		state := campaigns.ChangesetState(*args.State)
		if !state.Valid() {
			return opts, false, errors.New("changeset state not valid")
		}
		opts.ExternalState = &state
		// hiddenChangesetResolver has a State property so filtering based on
		// that is safe.
	}
	if args.ReviewState != nil {
		state := campaigns.ChangesetReviewState(*args.ReviewState)
		if !state.Valid() {
			return opts, false, errors.New("changeset review state not valid")
		}
		opts.ExternalReviewState = &state
		// If the user filters by ReviewState we cannot include hidden
		// changesets, since that would leak information.
		safe = false
	}
	if args.CheckState != nil {
		state := campaigns.ChangesetCheckState(*args.CheckState)
		if !state.Valid() {
			return opts, false, errors.New("changeset check state not valid")
		}
		opts.ExternalCheckState = &state
		// If the user filters by CheckState we cannot include hidden
		// changesets, since that would leak information.
		safe = false
	}

	return opts, safe, nil
}

func (r *Resolver) CreatePatchSetFromPatches(ctx context.Context, args graphqlbackend.CreatePatchSetFromPatchesArgs) (graphqlbackend.PatchSetResolver, error) {
	var err error
	tr, ctx := trace.New(ctx, "Resolver.CreatePatchSetFromPatches", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may create patch sets for now.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	user, err := backend.CurrentUser(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "%v", backend.ErrNotAuthenticated)
	}
	if user == nil {
		return nil, backend.ErrNotAuthenticated
	}

	patches := make([]*campaigns.Patch, len(args.Patches))
	for i, patch := range args.Patches {
		repo, err := graphqlbackend.UnmarshalRepositoryID(patch.Repository)
		if err != nil {
			return nil, err
		}

		p := &campaigns.Patch{
			RepoID:  repo,
			Rev:     patch.BaseRevision,
			BaseRef: patch.BaseRef,
			Diff:    patch.Patch,
		}
		// Ensure patch is a valid unified diff by computing diff stats.
		err = p.ComputeDiffStat()
		if err != nil {
			return nil, errors.Wrapf(err, "patch for repository ID %q (base revision %q)", patch.Repository, patch.BaseRevision)
		}

		patches[i] = p
	}

	svc := ee.NewService(r.store, r.httpFactory)
	patchSet, err := svc.CreatePatchSetFromPatches(ctx, patches, user.ID)
	if err != nil {
		return nil, err
	}

	return &patchSetResolver{store: r.store, patchSet: patchSet}, nil
}

func (r *Resolver) CloseCampaign(ctx context.Context, args *graphqlbackend.CloseCampaignArgs) (_ graphqlbackend.CampaignResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.CloseCampaign", fmt.Sprintf("Campaign: %q", args.Campaign))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	campaignID, err := campaigns.UnmarshalCampaignID(args.Campaign)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling campaign id")
	}

	if campaignID == 0 {
		return nil, ErrIDIsZero
	}

	svc := ee.NewService(r.store, r.httpFactory)
	// 🚨 SECURITY: CloseCampaign checks whether current user is authorized.
	campaign, err := svc.CloseCampaign(ctx, campaignID, args.CloseChangesets)
	if err != nil {
		return nil, errors.Wrap(err, "closing campaign")
	}

	return &campaignResolver{store: r.store, httpFactory: r.httpFactory, Campaign: campaign}, nil
}

func (r *Resolver) PublishCampaignChangesets(ctx context.Context, args *graphqlbackend.PublishCampaignChangesetsArgs) (_ graphqlbackend.CampaignResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.PublishCampaignChangesets", fmt.Sprintf("Campaign: %q", args.Campaign))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	campaignID, err := campaigns.UnmarshalCampaignID(args.Campaign)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling campaign id")
	}

	if campaignID == 0 {
		return nil, ErrIDIsZero
	}

	svc := ee.NewService(r.store, r.httpFactory)
	// 🚨 SECURITY: EnqueueChangesetJobs checks whether current user is authorized.
	campaign, err := svc.EnqueueChangesetJobs(ctx, campaignID)
	if err != nil {
		return nil, errors.Wrap(err, "publishing campaign changesets")
	}

	return &campaignResolver{store: r.store, httpFactory: r.httpFactory, Campaign: campaign}, nil
}
func (r *Resolver) PublishChangeset(ctx context.Context, args *graphqlbackend.PublishChangesetArgs) (_ *graphqlbackend.EmptyResponse, err error) {
	tr, ctx := trace.New(ctx, "Resolver.PublishChangeset", fmt.Sprintf("Patch: %q", args.Patch))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	patchID, err := unmarshalPatchID(args.Patch)
	if err != nil {
		return nil, err
	}

	if patchID == 0 {
		return nil, ErrIDIsZero
	}

	// 🚨 SECURITY: EnqueueChangesetJobForPatch checks whether current user is authorized.
	svc := ee.NewService(r.store, r.httpFactory)
	if err = svc.EnqueueChangesetJobForPatch(ctx, patchID); err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) SyncChangeset(ctx context.Context, args *graphqlbackend.SyncChangesetArgs) (_ *graphqlbackend.EmptyResponse, err error) {
	tr, ctx := trace.New(ctx, "Resolver.SyncChangeset", fmt.Sprintf("Changeset: %q", args.Changeset))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	changesetID, err := unmarshalChangesetID(args.Changeset)
	if err != nil {
		return nil, err
	}

	if changesetID == 0 {
		return nil, ErrIDIsZero
	}

	// 🚨 SECURITY: EnqueueChangesetSync checks whether current user is authorized.
	svc := ee.NewService(r.store, r.httpFactory)
	if err = svc.EnqueueChangesetSync(ctx, changesetID); err != nil {
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

func currentUserCanAdministerCampaign(ctx context.Context, c *campaigns.Campaign) (bool, error) {
	// 🚨 SECURITY: Only site admins or the authors of a campaign have campaign admin rights.
	if err := backend.CheckSiteAdminOrSameUser(ctx, c.AuthorID); err != nil {
		if _, ok := err.(*backend.InsufficientAuthorizationError); ok {
			return false, nil
		}

		return false, err
	}
	return true, nil
}
