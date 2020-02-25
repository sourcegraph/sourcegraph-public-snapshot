package resolvers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"gopkg.in/inconshreveable/log15.v2"
)

// Resolver is the GraphQL resolver of all things A8N.
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
		return nil, err
	}

	return &campaignResolver{store: r.store, Campaign: campaign}, nil
}

func (r *Resolver) ChangesetPlanByID(ctx context.Context, id graphql.ID) (graphqlbackend.ChangesetPlanResolver, error) {
	// 🚨 SECURITY: Only site admins or users when read-access is enabled may access campaign jobs.
	if err := allowReadAccess(ctx); err != nil {
		return nil, err
	}

	campaignJobID, err := unmarshalCampaignJobID(id)
	if err != nil {
		return nil, err
	}

	job, err := r.store.GetCampaignJob(ctx, ee.GetCampaignJobOpts{ID: campaignJobID})
	if err != nil {
		return nil, err
	}

	return &campaignJobResolver{store: r.store, job: job}, nil
}

func (r *Resolver) CampaignPlanByID(ctx context.Context, id graphql.ID) (graphqlbackend.CampaignPlanResolver, error) {
	// 🚨 SECURITY: Only site admins or users when read-access is enabled may access campaign plans.
	if err := allowReadAccess(ctx); err != nil {
		return nil, err
	}

	planID, err := unmarshalCampaignPlanID(id)
	if err != nil {
		return nil, err
	}

	plan, err := r.store.GetCampaignPlan(ctx, ee.GetCampaignPlanOpts{ID: planID})
	if err != nil {
		return nil, err
	}

	return &campaignPlanResolver{store: r.store, campaignPlan: plan}, nil
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

	if campaign.CampaignPlanID != 0 {
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
		Name:        args.Input.Name,
		Description: args.Input.Description,
		AuthorID:    user.ID,
	}

	if args.Input.Branch != nil {
		campaign.Branch = *args.Input.Branch
	}

	if args.Input.Plan != nil {
		planID, err := unmarshalCampaignPlanID(*args.Input.Plan)
		if err != nil {
			return nil, err
		}
		campaign.CampaignPlanID = planID
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

	svc := ee.NewService(r.store, gitserver.DefaultClient, nil, r.httpFactory)
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

	if args.Input.Plan != nil {
		campaignPlanID, err := unmarshalCampaignPlanID(*args.Input.Plan)
		if err != nil {
			return nil, err
		}
		updateArgs.Plan = &campaignPlanID
	}

	svc := ee.NewService(r.store, gitserver.DefaultClient, nil, r.httpFactory)
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

	svc := ee.NewService(r.store, gitserver.DefaultClient, nil, r.httpFactory)
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
	var opts ee.ListCampaignsOpts
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
		return nil, err
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
	syncer := ee.ChangesetSyncer{
		ReposStore:  store,
		Store:       tx,
		HTTPFactory: r.httpFactory,
	}
	if err = syncer.SyncChangesets(ctx, cs...); err != nil {
		return nil, err
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

func (r *Resolver) Changesets(ctx context.Context, args *graphqlutil.ConnectionArgs) (graphqlbackend.ExternalChangesetsConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins or users when read-access is enabled may access changesets.
	if err := allowReadAccess(ctx); err != nil {
		return nil, err
	}
	return &changesetsConnectionResolver{
		store: r.store,
		opts: ee.ListChangesetsOpts{
			Limit: int(args.GetFirst()),
		},
	}, nil
}

func (r *Resolver) CreateCampaignPlanFromPatches(ctx context.Context, args graphqlbackend.CreateCampaignPlanFromPatchesArgs) (graphqlbackend.CampaignPlanResolver, error) {
	var err error
	tr, ctx := trace.New(ctx, "Resolver.CreateCampaignPlanFromPatches", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may create campaign plans for now
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

	patches := make([]campaigns.CampaignPlanPatch, len(args.Patches))
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

		patches[i] = campaigns.CampaignPlanPatch{
			Repo:         repo,
			BaseRevision: patch.BaseRevision,
			Patch:        patch.Patch,
		}
	}

	svc := ee.NewService(r.store, gitserver.DefaultClient, nil, r.httpFactory)
	plan, err := svc.CreateCampaignPlanFromPatches(ctx, patches, user.ID)
	if err != nil {
		return nil, err
	}

	return &campaignPlanResolver{store: r.store, campaignPlan: plan}, nil
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

	svc := ee.NewService(r.store, gitserver.DefaultClient, nil, r.httpFactory)

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

	svc := ee.NewService(r.store, gitserver.DefaultClient, nil, r.httpFactory)
	campaign, err := svc.PublishCampaign(ctx, campaignID)
	if err != nil {
		return nil, errors.Wrap(err, "publishing campaign")
	}

	return &campaignResolver{store: r.store, Campaign: campaign}, nil
}

func (r *Resolver) PublishChangeset(ctx context.Context, args *graphqlbackend.PublishChangesetArgs) (_ *graphqlbackend.EmptyResponse, err error) {
	tr, ctx := trace.New(ctx, "Resolver.PublishChangeset", fmt.Sprintf("ChangesetPlan: %q", args.ChangesetPlan))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may update campaigns for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	campaignJobID, err := unmarshalCampaignJobID(args.ChangesetPlan)
	if err != nil {
		return nil, err
	}

	svc := ee.NewService(r.store, gitserver.DefaultClient, nil, r.httpFactory)
	err = svc.CreateChangesetJobForCampaignJob(ctx, campaignJobID)
	if err != nil {
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

// actionEnvVarResolver

type actionEnvVarResolver struct {
	key   string
	value string
}

func (r actionEnvVarResolver) Key() string {
	return r.key
}

func (r actionEnvVarResolver) Value() string {
	return r.value
}

// actionConnectionResolver

type actionConnectionResolver struct {
	once  sync.Once
	store *ee.Store

	first *int32

	actions    []*campaigns.Action
	totalCount int64
	err        error
}

func (r *actionConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	_, totalCount, err := r.compute(ctx)
	if err != nil {
		return 0, err
	}
	// todo: dangerous
	return int32(totalCount), nil
}

func (r *actionConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ActionResolver, error) {
	nodes, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.ActionResolver, len(nodes))
	for i, node := range nodes {
		resolvers[i] = &actionResolver{store: r.store, action: *node}
	}
	return resolvers, nil
}

func (r *actionConnectionResolver) compute(ctx context.Context) ([]*campaigns.Action, int64, error) {
	r.once.Do(func() {
		limit := -1
		if r.first != nil {
			limit = int(*r.first)
		}
		r.actions, r.totalCount, r.err = r.store.ListActions(ctx, ee.ListActionsOpts{Limit: limit, Cursor: 0})
	})
	return r.actions, r.totalCount, r.err
}

// actionExecutionConnectionResolver

type actionExecutionConnectionResolver struct {
	store    *ee.Store
	actionID int64
	first    *int32

	once sync.Once

	actionExecutions []*campaigns.ActionExecution
	totalCount       int64
	err              error
}

func (r *actionExecutionConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	_, totalCount, err := r.compute(ctx)
	// todo: dangerous
	return int32(totalCount), err
}

func (r *actionExecutionConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ActionExecutionResolver, error) {
	nodes, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.ActionExecutionResolver, len(nodes))
	for i, node := range nodes {
		resolvers[i] = &actionExecutionResolver{store: r.store, actionExecution: *node}
	}
	return resolvers, nil
}

func (r *actionExecutionConnectionResolver) compute(ctx context.Context) ([]*campaigns.ActionExecution, int64, error) {
	r.once.Do(func() {
		limit := -1
		if r.first != nil {
			limit = int(*r.first)
		}
		r.actionExecutions, r.totalCount, r.err = r.store.ListActionExecutions(ctx, ee.ListActionExecutionsOpts{Limit: limit, Cursor: 0, ActionID: &r.actionID})
	})
	return r.actionExecutions, r.totalCount, r.err
}

// actionJobConnectionResolver

type actionJobConnectionResolver struct {
	store *ee.Store

	once       sync.Once
	jobs       *[]campaigns.ActionJob
	totalCount int32
	err        error
}

func (r *actionJobConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	_, totalCount, err := r.compute(ctx)
	return totalCount, err
}

func (r *actionJobConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ActionJobResolver, error) {
	jobs, _, err := r.compute(ctx)
	resolvers := make([]graphqlbackend.ActionJobResolver, len(jobs))
	for i, job := range jobs {
		resolvers[i] = &actionJobResolver{store: r.store, job: job}
	}
	return resolvers, err
}

func (r *actionJobConnectionResolver) compute(ctx context.Context) ([]campaigns.ActionJob, int32, error) {
	// this might have been passed down (CreateActionExecution already knows all jobs, so why fetch them again. TODO: paginate those as well)
	if r.jobs == nil {
		r.once.Do(func() {
			actionJobs := make([]campaigns.ActionJob, 1)
			actionJob, err := r.store.ActionJobByID(ctx, ee.ActionJobByIDOpts{ID: 123})
			if err != nil {
				r.jobs = nil
				r.totalCount = 0
				r.err = err
			} else if actionJob == nil {
				r.jobs = nil
				r.totalCount = 0
				r.err = nil
			} else {
				// todo: this is needs to be fetched from the parent action execution, not be static 123
				actionJobs[0] = *actionJob
				r.jobs = &actionJobs
				r.totalCount = 1
				r.err = err
			}
		})
	} else {
		// todo: unsafe
		r.totalCount = int32(len(*r.jobs))
	}
	return *r.jobs, r.totalCount, r.err
}

// runner resolver

type runnerResolver struct {
	// todo
}

func (r *runnerResolver) ID() graphql.ID {
	return "asd"
}

func (r *runnerResolver) Name() string {
	return "runner-sg-dev-123"
}

func (r *runnerResolver) Description() string {
	return "macOS 10.15.3, Docker 19.06.03, 8 CPU"
}

func (r *runnerResolver) State() campaigns.RunnerState {
	return campaigns.RunnerStateOnline
}

func (r *runnerResolver) RunningJobs() graphqlbackend.ActionJobConnectionResolver {
	return &actionJobConnectionResolver{}
}

// action definition resolver

type actionDefinitionResolver struct {
	steps  string
	envStr string
}

func (r *actionDefinitionResolver) Steps() graphqlbackend.JSONCString {
	return graphqlbackend.JSONCString(r.steps)
}

func (r *actionDefinitionResolver) ActionWorkspace() *graphqlbackend.GitTreeEntryResolver {
	return nil
}

func (r *actionDefinitionResolver) Env() ([]graphqlbackend.ActionEnvVarResolver, error) {
	if r.envStr == "" {
		return []graphqlbackend.ActionEnvVarResolver{}, nil
	}
	var parsed []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal([]byte(r.envStr), &parsed); err != nil {
		return nil, errors.Wrap(err, "invalid env stored")
	}
	envs := make([]graphqlbackend.ActionEnvVarResolver, len(parsed))
	for i, env := range parsed {
		envs[i] = &actionEnvVarResolver{
			key:   env.Key,
			value: env.Value,
		}
	}
	return envs, nil
}

// query and mutation resolvers

func (r *Resolver) Actions(ctx context.Context, args *graphqlbackend.ListActionsArgs) (_ graphqlbackend.ActionConnectionResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.Actions", fmt.Sprintf("First: %d", args.First))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may create executions for now
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

	// 🚨 SECURITY: Only site admins may create executions for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	// todo: pass down args
	return &actionJobConnectionResolver{store: r.store}, nil
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

	return nil, nil
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

	return nil, nil
}

func (r *Resolver) UploadWorkspace(ctx context.Context, args *graphqlbackend.UploadWorkspaceArgs) (_ *graphqlbackend.GitTreeEntryResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.UploadWorkspace", fmt.Sprintf("Size: %d", len(args.Content)))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may create executions for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	return nil, nil
}

func (r *Resolver) CreateActionExecution(ctx context.Context, args *graphqlbackend.CreateActionExecutionArgs) (_ graphqlbackend.ActionExecutionResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.CreateActionExecution", fmt.Sprintf("Action: %s", args.Action))
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

	action, err := r.store.ActionByID(ctx, ee.ActionByIDOpts{ID: actionID})
	if err != nil {
		return nil, err
	}
	if action.ID == 0 {
		return nil, errors.New("Action not found")
	}
	scopeQuery, err := scopeQueryForSteps(action.Steps)
	if err != nil {
		return nil, err
	}
	repos, err := findRepos(ctx, scopeQuery)
	if err != nil {
		return nil, err
	}
	if len(repos) == 0 {
		return nil, errors.New("Cannot create execution for action that yields 0 repositories")
	}
	// todo: the next steps need to happen in a transaction
	actionExecution, err := r.store.CreateActionExecution(ctx, ee.CreateActionExecutionOpts{
		InvokationReason: campaigns.ActionExecutionInvokationReasonManual,
		Steps:            action.Steps,
		EnvStr:           action.EnvStr,
		ActionID:         action.ID,
	})
	if err != nil {
		return nil, err
	}
	actionJobs := make([]campaigns.ActionJob, len(repos))
	for i, repo := range repos {
		repoID, err := graphqlbackend.UnmarshalRepositoryID(graphql.ID(repo.ID))
		if err != nil {
			return nil, err
		}
		actionJob, err := r.store.CreateActionJob(ctx, ee.CreateActionJobOpts{
			ExecutionID:  actionExecution.ID,
			RepositoryID: int64(repoID),
			BaseRevision: repo.Rev,
		})
		if err != nil {
			return nil, err
		}
		actionJobs[i] = *actionJob
	}

	return &actionExecutionResolver{store: r.store, actionExecution: *actionExecution, actionJobs: actionJobs}, nil
}

func (r *Resolver) PullActionJob(ctx context.Context, args *graphqlbackend.PullActionJobArgs) (_ graphqlbackend.ActionJobResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.PullActionJob", fmt.Sprintf("Runner: %q", args.Runner))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admin tokens can register as a runner for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	actionJob, err := r.store.PullActionJob(ctx)
	if err != nil {
		return nil, err
	}

	// todo better handling of this
	if actionJob.ID == 0 {
		return nil, nil
	}

	// set runner = args.Runner

	return &actionJobResolver{store: r.store, job: *actionJob}, nil
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

	id, err := unmarshalActionJobID(args.ActionJob)
	if err != nil {
		return nil, err
	}

	opts := ee.UpdateActionJobOpts{
		ID:    id,
		State: args.State,
		Patch: args.Patch,
	}
	// set finish time on completion
	if *args.State == campaigns.ActionJobStateCompleted {
		now := time.Now()
		opts.ExecutionEnd = &now
	}
	actionJob, err := r.store.UpdateActionJob(ctx, opts)
	if err != nil {
		return nil, err
	}

	// if args.state == "COMPLETED" {
	// 	// todo: check if was running before, otherwise updating state is not allowed
	// 	// todo: check if ALL are completed, timeouted, or failed now, then proceed with patch generation
	// 	patches := make([]CampaignPlanPatch)
	// 	actionJobs := make([]campaigns.ActionJob)
	// 	for _, aj := range actionJobs {
	// 		if aj.patch != nil {
	// 			append(patches, CampaignPlanPatch{
	// 				Repository:   aj.repo,
	// 				BaseRevision: aj.revision,
	// 				Patch:        aj.patch,
	// 			})
	// 		}
	// 	}
	// }
	return &actionJobResolver{store: r.store, job: *actionJob}, nil
}

func (r *Resolver) AppendLog(ctx context.Context, args *graphqlbackend.AppendLogArgs) (_ graphqlbackend.ActionJobResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.AppendLog", fmt.Sprintf("ActionJob: %q", args.ActionJob))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admin tokens can register as a runner for now, todo: this should only be allowed to runners. (we set RunnerSeenAt: time.Now())
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	id, err := unmarshalActionJobID(args.ActionJob)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	actionJob, err := r.store.UpdateActionJob(ctx, ee.UpdateActionJobOpts{
		ID:           id,
		Log:          &args.Content,
		RunnerSeenAt: &now,
	})
	if err != nil {
		return nil, err
	}

	// todo: test if this works
	if actionJob.ID == 0 {
		return nil, nil
	}

	return &actionJobResolver{store: r.store, job: *actionJob}, nil
}

func (r *Resolver) RetryActionJob(ctx context.Context, args *graphqlbackend.RetryActionJobArgs) (_ *graphqlbackend.EmptyResponse, err error) {
	tr, ctx := trace.New(ctx, "Resolver.RetryActionJob", fmt.Sprintf("Action job: %s", args.ActionJob))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may create executions for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, errors.Wrap(err, "checking if user is admin")
	}

	id, err := unmarshalActionJobID(args.ActionJob)
	if err != nil {
		return nil, err
	}

	if err := r.store.ClearActionJob(ctx, ee.ClearActionJobOpts{
		ID: id,
	}); err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, nil
}
