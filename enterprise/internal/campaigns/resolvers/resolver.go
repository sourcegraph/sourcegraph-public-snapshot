package resolvers

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

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
	if readAccess := conf.CampaignsReadAccessEnabled(); readAccess {
		return nil
	}

	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return err
	}

	return nil
}

func (r *Resolver) ChangesetByID(ctx context.Context, id graphql.ID) (graphqlbackend.ExternalChangesetResolver, error) {
	// 🚨 SECURITY: Only site admins or users when read-access is enabled may access changesets.
	if err := allowReadAccess(ctx); err != nil {
		return nil, err
	}

	changesetID, err := unmarshalChangesetID(id)
	if err != nil {
		return nil, err
	}

	changeset, err := r.store.GetChangeset(ctx, ee.GetChangesetOpts{ID: changesetID})
	if err != nil {
		if err == ee.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &changesetResolver{store: r.store, Changeset: changeset}, nil
}

func (r *Resolver) CampaignByID(ctx context.Context, id graphql.ID) (graphqlbackend.CampaignResolver, error) {
	// 🚨 SECURITY: Only site admins or users when read-access is enabled may access campaign.
	if err := allowReadAccess(ctx); err != nil {
		return nil, err
	}

	campaignID, err := unmarshalCampaignID(id)
	if err != nil {
		return nil, err
	}

	campaign, err := r.store.GetCampaign(ctx, ee.GetCampaignOpts{ID: campaignID})
	if err != nil {
		if err == ee.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &campaignResolver{store: r.store, Campaign: campaign}, nil
}

func (r *Resolver) PatchByID(ctx context.Context, id graphql.ID) (graphqlbackend.PatchResolver, error) {
	// 🚨 SECURITY: Only site admins or users when read-access is enabled may access patches.
	if err := allowReadAccess(ctx); err != nil {
		return nil, err
	}

	patchID, err := unmarshalPatchID(id)
	if err != nil {
		return nil, err
	}

	patch, err := r.store.GetPatch(ctx, ee.GetPatchOpts{ID: patchID})
	if err != nil {
		if err == ee.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &patchResolver{store: r.store, patch: patch}, nil
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
	// 🚨 SECURITY: Only site admins may modify changesets and campaigns for now.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	campaignID, err := unmarshalCampaignID(args.Campaign)
	if err != nil {
		return nil, err
	}

	changesetIDs := make([]int64, 0, len(args.Changesets))
	set := map[int64]struct{}{}
	for _, changesetID := range args.Changesets {
		id, err := unmarshalChangesetID(changesetID)
		if err != nil {
			return nil, err
		}

		if _, ok := set[id]; !ok {
			changesetIDs = append(changesetIDs, id)
			set[id] = struct{}{}
		}
	}

	tx, err := r.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Done(&err)

	campaign, err := tx.GetCampaign(ctx, ee.GetCampaignOpts{ID: campaignID})
	if err != nil {
		return nil, err
	}

	if campaign.PatchSetID != 0 {
		return nil, errors.New("Changesets can only be added to campaigns that don't create their own changesets")
	}

	changesets, _, err := tx.ListChangesets(ctx, ee.ListChangesetsOpts{IDs: changesetIDs})
	if err != nil {
		return nil, err
	}

	for _, c := range changesets {
		delete(set, c.ID)
		c.CampaignIDs = append(c.CampaignIDs, campaign.ID)
	}

	if len(set) > 0 {
		return nil, errors.Errorf("changesets %v not found", set)
	}

	if err = tx.UpdateChangesets(ctx, changesets...); err != nil {
		return nil, err
	}

	campaign.ChangesetIDs = append(campaign.ChangesetIDs, changesetIDs...)
	if err = tx.UpdateCampaign(ctx, campaign); err != nil {
		return nil, err
	}

	return &campaignResolver{store: r.store, Campaign: campaign}, nil
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

	var draft bool
	if args.Input.Draft != nil {
		draft = *args.Input.Draft
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

	svc := ee.NewService(r.store, gitserver.DefaultClient, r.httpFactory)
	err = svc.CreateCampaign(ctx, campaign, draft)
	if err != nil {
		return nil, err
	}

	return &campaignResolver{store: r.store, Campaign: campaign}, nil
}

func (r *Resolver) UpdateCampaign(ctx context.Context, args *graphqlbackend.UpdateCampaignArgs) (_ graphqlbackend.CampaignResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.UpdateCampaign", fmt.Sprintf("Campaign: %q", args.Input.ID))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may update campaigns for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	campaignID, err := unmarshalCampaignID(args.Input.ID)
	if err != nil {
		return nil, err
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

	svc := ee.NewService(r.store, gitserver.DefaultClient, r.httpFactory)
	campaign, detachedChangesets, err := svc.UpdateCampaign(ctx, updateArgs)
	if err != nil {
		return nil, err
	}

	if detachedChangesets != nil {
		go func() {
			ctx := trace.ContextWithTrace(context.Background(), tr)
			err := svc.CloseOpenChangesets(ctx, detachedChangesets)
			if err != nil {
				log15.Error("CloseOpenChangesets", "err", err)
			}
		}()
	}

	return &campaignResolver{store: r.store, Campaign: campaign}, nil
}

func (r *Resolver) DeleteCampaign(ctx context.Context, args *graphqlbackend.DeleteCampaignArgs) (_ *graphqlbackend.EmptyResponse, err error) {
	tr, ctx := trace.New(ctx, "Resolver.DeleteCampaign", fmt.Sprintf("Campaign: %q", args.Campaign))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may update campaigns for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	campaignID, err := unmarshalCampaignID(args.Campaign)
	if err != nil {
		return nil, err
	}

	svc := ee.NewService(r.store, gitserver.DefaultClient, r.httpFactory)
	err = svc.DeleteCampaign(ctx, campaignID, args.CloseChangesets)
	return &graphqlbackend.EmptyResponse{}, err
}

func (r *Resolver) RetryCampaign(ctx context.Context, args *graphqlbackend.RetryCampaignArgs) (graphqlbackend.CampaignResolver, error) {
	var err error
	tr, ctx := trace.New(ctx, "Resolver.RetryCampaign", fmt.Sprintf("Campaign: %q", args.Campaign))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may update campaigns for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	campaignID, err := unmarshalCampaignID(args.Campaign)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling campaign id")
	}

	campaign, err := r.store.GetCampaign(ctx, ee.GetCampaignOpts{ID: campaignID})
	if err != nil {
		return nil, errors.Wrap(err, "getting campaign")
	}

	err = r.store.ResetFailedChangesetJobs(ctx, campaign.ID)
	if err != nil {
		return nil, errors.Wrap(err, "resetting failed changeset jobs")
	}

	return &campaignResolver{store: r.store, Campaign: campaign}, nil
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
	return &campaignsConnectionResolver{
		store: r.store,
		opts:  opts,
	}, nil
}

func (r *Resolver) CreateChangesets(ctx context.Context, args *graphqlbackend.CreateChangesetsArgs) (_ []graphqlbackend.ExternalChangesetResolver, err error) {
	// 🚨 SECURITY: Only site admins may create changesets for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	var repoIDs []api.RepoID
	repoSet := map[api.RepoID]*repos.Repo{}
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

	store := repos.NewDBStore(tx.DB(), sql.TxOptions{})

	rs, err := store.ListRepos(ctx, repos.StoreListReposArgs{IDs: repoIDs})
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

	store = repos.NewDBStore(tx.DB(), sql.TxOptions{})

	// NOTE: We are performing a blocking sync here in order to ensure
	// that the remote changeset exists and also to remove the possibility
	// of an unsynced changeset entering our database
	if err = ee.SyncChangesets(ctx, store, tx, r.httpFactory, cs...); err != nil {
		return nil, errors.Wrap(err, "syncing changesets")
	}

	csr := make([]graphqlbackend.ExternalChangesetResolver, len(cs))
	for i := range cs {
		csr[i] = &changesetResolver{
			store:         r.store,
			Changeset:     cs[i],
			preloadedRepo: repoSet[cs[i].RepoID],
		}
	}

	return csr, nil
}

func (r *Resolver) Changesets(ctx context.Context, args *graphqlbackend.ListChangesetsArgs) (graphqlbackend.ExternalChangesetsConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins or users when read-access is enabled may access changesets.
	if err := allowReadAccess(ctx); err != nil {
		return nil, err
	}
	opts, err := listChangesetOptsFromArgs(args)
	if err != nil {
		return nil, err
	}
	return &changesetsConnectionResolver{
		store: r.store,
		opts:  opts,
	}, nil
}

func listChangesetOptsFromArgs(args *graphqlbackend.ListChangesetsArgs) (ee.ListChangesetsOpts, error) {
	var opts ee.ListChangesetsOpts
	if args == nil {
		return opts, nil
	}
	if args.First != nil {
		opts.Limit = int(*args.First)
	}
	if args.State != nil {
		state := campaigns.ChangesetState(*args.State)
		if !state.Valid() {
			return opts, errors.New("changeset state not valid")
		}
		opts.ExternalState = &state
	}
	if args.ReviewState != nil {
		state := campaigns.ChangesetReviewState(*args.ReviewState)
		if !state.Valid() {
			return opts, errors.New("changeset review state not valid")
		}
		opts.ExternalReviewState = &state
	}
	if args.CheckState != nil {
		state := campaigns.ChangesetCheckState(*args.CheckState)
		if !state.Valid() {
			return opts, errors.New("changeset check state not valid")
		}
		opts.ExternalCheckState = &state
	}
	return opts, nil
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

		// Ensure patch is a valid unified diff.
		diffReader := diff.NewMultiFileDiffReader(strings.NewReader(patch.Patch))
		for {
			_, err := diffReader.ReadFile()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, errors.Wrapf(err, "patch for repository ID %q (base revision %q)", patch.Repository, patch.BaseRevision)
			}
		}

		patches[i] = &campaigns.Patch{
			RepoID:  repo,
			Rev:     patch.BaseRevision,
			BaseRef: patch.BaseRef,
			Diff:    patch.Patch,
		}
	}

	svc := ee.NewService(r.store, gitserver.DefaultClient, r.httpFactory)
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

	// 🚨 SECURITY: Only site admins may update campaigns for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	campaignID, err := unmarshalCampaignID(args.Campaign)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling campaign id")
	}

	svc := ee.NewService(r.store, gitserver.DefaultClient, r.httpFactory)

	campaign, err := svc.CloseCampaign(ctx, campaignID, args.CloseChangesets)
	if err != nil {
		return nil, errors.Wrap(err, "closing campaign")
	}

	return &campaignResolver{store: r.store, Campaign: campaign}, nil
}

func (r *Resolver) PublishCampaign(ctx context.Context, args *graphqlbackend.PublishCampaignArgs) (_ graphqlbackend.CampaignResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.PublishCampaign", fmt.Sprintf("Campaign: %q", args.Campaign))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may update campaigns for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	campaignID, err := unmarshalCampaignID(args.Campaign)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling campaign id")
	}

	svc := ee.NewService(r.store, gitserver.DefaultClient, r.httpFactory)
	campaign, err := svc.PublishCampaign(ctx, campaignID)
	if err != nil {
		return nil, errors.Wrap(err, "publishing campaign")
	}

	return &campaignResolver{store: r.store, Campaign: campaign}, nil
}

func (r *Resolver) PublishChangeset(ctx context.Context, args *graphqlbackend.PublishChangesetArgs) (_ *graphqlbackend.EmptyResponse, err error) {
	tr, ctx := trace.New(ctx, "Resolver.PublishChangeset", fmt.Sprintf("Patch: %q", args.Patch))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may update campaigns for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	patchID, err := unmarshalPatchID(args.Patch)
	if err != nil {
		return nil, err
	}

	svc := ee.NewService(r.store, gitserver.DefaultClient, r.httpFactory)
	err = svc.CreateChangesetJobForPatch(ctx, patchID)
	if err != nil {
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

	// 🚨 SECURITY: Only site admins may update campaigns for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	changesetID, err := unmarshalChangesetID(args.Changeset)
	if err != nil {
		return nil, err
	}

	// Check for existence of changeset so we don't swallow that error.
	if _, err = r.store.GetChangeset(ctx, ee.GetChangesetOpts{ID: changesetID}); err != nil {
		return nil, err
	}

	if err := repoupdater.DefaultClient.EnqueueChangesetSync(ctx, []int64{changesetID}); err != nil {
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
