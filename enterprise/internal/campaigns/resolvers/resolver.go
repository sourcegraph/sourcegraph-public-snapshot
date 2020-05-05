package resolvers

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"
	"time"

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

	svc := ee.NewService(r.store, r.httpFactory)
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

	svc := ee.NewService(r.store, r.httpFactory)
	campaign, detachedChangesets, err := svc.UpdateCampaign(ctx, updateArgs)
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

	// 🚨 SECURITY: Only site admins may update campaigns for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	campaignID, err := unmarshalCampaignID(args.Campaign)
	if err != nil {
		return nil, err
	}

	svc := ee.NewService(r.store, r.httpFactory)
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

	svc := ee.NewService(r.store, r.httpFactory)
	patchSet, err := svc.CreatePatchSetFromPatches(ctx, patches, user.ID, true)
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

	svc := ee.NewService(r.store, r.httpFactory)

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

	svc := ee.NewService(r.store, r.httpFactory)
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

	svc := ee.NewService(r.store, r.httpFactory)
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

func (r *Resolver) Actions(ctx context.Context, args *graphqlbackend.ListActionsArgs) (_ graphqlbackend.ActionConnectionResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.Actions", fmt.Sprintf("First: %d", args.First))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may read actions for now.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	return &actionConnectionResolver{store: r.store, first: args.First}, nil
}

func (r *Resolver) ActionJobs(ctx context.Context, args *graphqlbackend.ListActionJobsArgs) (_ graphqlbackend.ActionJobConnectionResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.ActionJobs", fmt.Sprintf("First: %d", args.First))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may read jobs for now.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	return &actionJobConnectionResolver{store: r.store, first: args.First, state: args.State}, nil
}

func (r *Resolver) Agents(ctx context.Context, args *graphqlbackend.ListAgentsArgs) (_ graphqlbackend.AgentConnectionResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.Agents", fmt.Sprintf("First: %d", args.First))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may list agents for now.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	return &agentConnectionResolver{store: r.store, first: args.First, state: args.State}, nil
}

func (r *Resolver) CreateAction(ctx context.Context, args *graphqlbackend.CreateActionArgs) (_ graphqlbackend.ActionResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.CreateAction", fmt.Sprintf("Definition: %s", args.Definition))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may create executions for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	action, err := r.store.CreateAction(ctx, ee.CreateActionOpts{
		Name:  args.Name,
		Steps: args.Definition,
	})
	if err != nil {
		return nil, err
	}

	return &actionResolver{store: r.store, action: action}, nil
}

func (r *Resolver) UpdateAction(ctx context.Context, args *graphqlbackend.UpdateActionArgs) (_ graphqlbackend.ActionResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.UpdateAction", fmt.Sprintf("Action: %s", args.Action))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may create executions for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	actionID, err := unmarshalActionID(args.Action)
	if err != nil {
		return nil, err
	}

	// Check for existence.
	_, err = r.store.GetAction(ctx, ee.GetActionOpts{ID: actionID})
	if err != nil {
		return nil, err
	}

	action, err := r.store.UpdateAction(ctx, ee.UpdateActionOpts{
		ActionID: actionID,
		Steps:    args.NewDefinition,
	})
	if err != nil {
		return nil, err
	}

	return &actionResolver{store: r.store, action: action}, nil
}

func (r *Resolver) CreateActionExecution(ctx context.Context, args *graphqlbackend.CreateActionExecutionArgs) (_ graphqlbackend.ActionExecutionResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.CreateActionExecution", fmt.Sprintf("Action: %s", args.Action))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may create executions for now.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	actionID, err := unmarshalActionID(args.Action)
	if err != nil {
		return nil, err
	}

	// Check if exists.
	action, err := r.store.GetAction(ctx, ee.GetActionOpts{ID: actionID})
	if err != nil {
		return nil, err
	}
	actionExecution, actionJobs, err := createActionExecutionForAction(ctx, r.store, action, campaigns.ActionExecutionInvocationReasonManual)
	if err != nil {
		return nil, err
	}

	return &actionExecutionResolver{store: r.store, actionExecution: actionExecution, actionJobs: &actionJobs}, nil
}

func (r *Resolver) CreateActionExecutionsForSavedSearch(ctx context.Context, args *graphqlbackend.CreateActionExecutionsForSavedSearchArgs) (*graphqlbackend.EmptyResponse, error) {
	actions, err := r.store.ListActionsBySavedSearchQuery(ctx, ee.ListActionsBySavedSearchQueryOpts{SavedSearchQuery: args.SavedSearchQuery})
	if err != nil {
		return nil, err
	}
	for _, action := range actions {
		_, _, err := createActionExecutionForAction(ctx, r.store, action, campaigns.ActionExecutionInvocationReasonSavedSearch)
		if err != nil {
			return nil, err
		}
		log15.Info(fmt.Sprintf("Created new execution for action %d\n", action.ID))
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) PullActionJob(ctx context.Context, args *graphqlbackend.PullActionJobArgs) (_ graphqlbackend.ActionJobResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.PullActionJob", fmt.Sprintf("Agent: %q", args.Agent))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admin tokens can register as an agent for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	dbId, err := unmarshalAgentID(args.Agent)
	if err != nil {
		return nil, err
	}

	// Check if the agent actually exists.
	agent, err := r.store.GetAgent(ctx, ee.GetAgentOpts{ID: dbId})
	if err != nil {
		return nil, err
	}

	actionJob, err := r.store.PullActionJob(ctx, ee.PullActionJobOpts{AgentID: agent.ID})
	if err != nil {
		if err == ee.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &actionJobResolver{store: r.store, job: actionJob}, nil
}

func (r *Resolver) UpdateActionJob(ctx context.Context, args *graphqlbackend.UpdateActionJobArgs) (_ graphqlbackend.ActionJobResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.UpdateActionJob", fmt.Sprintf("Action job: %s", args.ActionJob))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may create executions for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	// todo: we need a user to associate the patch set with, but is the issuer of the runner token implicitly good enough?
	user, err := backend.CurrentUser(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "%v", backend.ErrNotAuthenticated)
	}
	if user == nil {
		return nil, backend.ErrNotAuthenticated
	}

	id, err := unmarshalActionJobID(args.ActionJob)
	if err != nil {
		return nil, err
	}

	if args.Patch != nil {
		// todo: Where are we logging this error? Append to log and mark as failed?
		// Ensure patch is a valid unified diff. This is the same check we do for manually uploaded patches in the CreatePatchSetFromPatches mutation.
		diffReader := diff.NewMultiFileDiffReader(strings.NewReader(*args.Patch))
		for {
			_, err := diffReader.ReadFile()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, errors.Wrap(err, "invalid patch submitted")
			}
		}
	}

	actionJob, err := r.store.GetActionJob(ctx, ee.GetActionJobOpts{ID: id})
	if err != nil {
		if err == ee.ErrNoResults {
			return nil, errors.New("ActionJob not found")
		}
		return nil, err
	}

	// check if is running, otherwise updating state is not allowed
	if actionJob.State != campaigns.ActionJobStateRunning {
		return nil, errors.New("Cannot update not running action job")
	}

	opts := ee.UpdateActionJobOpts{
		ID:    id,
		State: args.State,
		Patch: args.Patch,
	}
	// Set end time on status change from "running".
	if args.State != nil && *args.State != campaigns.ActionJobStateRunning {
		now := time.Now()
		opts.ExecutionEndAt = &now
	}

	tx, err := r.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Done(&err)

	actionJob, err = tx.UpdateActionJob(ctx, opts)
	if err != nil {
		return nil, err
	}

	// If this job is not yet completed, no further action needs to be taken.
	if actionJob.State == campaigns.ActionJobStatePending || actionJob.State == campaigns.ActionJobStateRunning {
		return &actionJobResolver{store: r.store, job: actionJob}, nil
	}
	// check if ALL are completed, timeouted, or failed now, then proceed with patch generation.
	actionJobs, _, err := tx.ListActionJobs(ctx, ee.ListActionJobsOpts{
		ExecutionID: actionJob.ExecutionID,
		Limit:       -1,
	})
	if err != nil {
		return nil, err
	}
	allCompleted := true
	patchCount := 0
	for _, j := range actionJobs {
		if j.Patch != nil {
			patchCount = patchCount + 1
		}
		// a job is completed when it timeouted, failed, or completed
		if j.State == campaigns.ActionJobStatePending || j.State == campaigns.ActionJobStateRunning {
			allCompleted = false
			break
		}
	}
	if !allCompleted {
		return &actionJobResolver{store: r.store, job: actionJob}, nil
	}
	patches := make([]*campaigns.Patch, patchCount)
	for _, job := range actionJobs {
		if job.Patch != nil {
			patches = append(patches, &campaigns.Patch{
				RepoID:  api.RepoID(job.RepoID),
				Rev:     api.CommitID(job.BaseRevision),
				BaseRef: job.BaseReference,
				Diff:    *job.Patch,
			})
		}
	}
	svc := ee.NewService(tx, gitserver.DefaultClient, r.httpFactory)
	// important: pass false for useTx, as our transaction will already be committed bu CreatePatchSetFromPatches
	// otherwise, and we cannot update the execution within the tx anymore
	patchSet, err := svc.CreatePatchSetFromPatches(ctx, patches, user.ID, false)
	if err != nil {
		return nil, err
	}
	// attach patch set to action execution
	actionExecution, err := tx.UpdateActionExecution(ctx, ee.UpdateActionExecutionOpts{ExecutionID: actionJob.ExecutionID, PatchSetID: patchSet.ID})
	if err != nil {
		return nil, err
	}

	// check if action is associated to a campaign, then we update that directly with the plan from above
	action, err := tx.GetAction(ctx, ee.GetActionOpts{ID: actionExecution.ActionID})
	if err != nil {
		if err == ee.ErrNoResults {
			return nil, errors.New("Action not found")
		}
		return nil, err
	}
	if action.CampaignID != 0 {
		// todo: the tx is completed when the background go func of updateCampaign executes
		if _, err = updateCampaign(ctx, tx, r.httpFactory, tr, ee.UpdateCampaignArgs{Campaign: action.CampaignID, PatchSet: &patchSet.ID}); err != nil {
			return nil, err
		}
	}
	return &actionJobResolver{store: r.store, job: actionJob}, nil
}

func (r *Resolver) AppendLog(ctx context.Context, args *graphqlbackend.AppendLogArgs) (_ graphqlbackend.ActionJobResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.AppendLog", fmt.Sprintf("ActionJob: %q", args.ActionJob))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admin tokens can register as a runner for now, todo: this should only be allowed to runners. (we set AgentSeenAt: time.Now())
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	id, err := unmarshalActionJobID(args.ActionJob)
	if err != nil {
		return nil, err
	}

	// todo: when is the threshold for appending missing logs hit and appending any further logs is forbidden?

	actionJob, err := r.store.UpdateActionJob(ctx, ee.UpdateActionJobOpts{
		ID:  id,
		Log: &args.Content,
	})
	// todo: update agent last_seen_at
	if err != nil {
		// Return null if the job doesn't exist. TODO: Rly?
		if err == ee.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &actionJobResolver{store: r.store, job: actionJob}, nil
}

func (r *Resolver) RetryActionJob(ctx context.Context, args *graphqlbackend.RetryActionJobArgs) (_ graphqlbackend.ActionJobResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.RetryActionJob", fmt.Sprintf("Action job: %s", args.ActionJob))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may retry jobs for now.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	id, err := unmarshalActionJobID(args.ActionJob)
	if err != nil {
		return nil, err
	}

	// Check for existence.
	_, err = r.store.GetActionJob(ctx, ee.GetActionJobOpts{ID: id})
	if err != nil {
		return nil, err
	}

	job, err := r.store.ClearActionJob(ctx, ee.ClearActionJobOpts{
		ID: id,
	})
	if err != nil {
		return nil, err
	}

	return &actionJobResolver{store: r.store, job: job}, nil
}

func (r *Resolver) CancelActionExecution(ctx context.Context, args *graphqlbackend.CancelActionExecutionArgs) (_ *graphqlbackend.EmptyResponse, err error) {
	tr, ctx := trace.New(ctx, "Resolver.CancelActionExecution", fmt.Sprintf("Action execution: %s", args.ActionExecution))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may cancel executions for now.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	id, err := unmarshalActionExecutionID(args.ActionExecution)
	if err != nil {
		return nil, err
	}
	// Probe if execution exists, so we don't swallow errors.
	_, err = r.store.GetActionExecution(ctx, ee.GetActionExecutionOpts{ID: id})
	if err != nil {
		return nil, err
	}

	// Set all remaining unfinished tasks to canceled.
	err = r.store.CancelActionExecution(ctx, ee.CancelActionExecutionOpts{ExecutionID: id})
	if err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) RegisterAgent(ctx context.Context, args *graphqlbackend.RegisterAgentArgs) (_ graphqlbackend.AgentResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.RegisterAgent", fmt.Sprintf("Agent ID: %s", args.ID))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may register agents for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	agent, err := r.store.CreateAgent(ctx, ee.CreateAgentOpts{Name: args.ID, Specs: args.Specs})
	if err != nil {
		return nil, err
	}
	return &agentResolver{agent: agent, store: r.store}, nil
}

func (r *Resolver) UnregisterAgent(ctx context.Context, args *graphqlbackend.UnregisterAgentArgs) (_ *graphqlbackend.EmptyResponse, err error) {
	tr, ctx := trace.New(ctx, "Resolver.UnregisterAgent", fmt.Sprintf("Agent ID: %s", args.Agent))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may unregister agents for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	// todo
	return &graphqlbackend.EmptyResponse{}, nil
}
