package resolvers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
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
func NewResolver(db *sql.DB) graphqlbackend.A8NResolver {
	return &Resolver{store: ee.NewStore(db)}
}

func (r *Resolver) ChangesetByID(ctx context.Context, id graphql.ID) (graphqlbackend.ExternalChangesetResolver, error) {
	// 🚨 SECURITY: Only site admins may access changesets for now.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
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
	// 🚨 SECURITY: Only site admins may access campaigns for now.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
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

func (r *Resolver) CampaignPlanByID(ctx context.Context, id graphql.ID) (graphqlbackend.CampaignPlanResolver, error) {
	// 🚨 SECURITY: Only site admins may access campaign plans for now.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
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

	campaign := &a8n.Campaign{
		Name:        args.Input.Name,
		Description: args.Input.Description,
		AuthorID:    user.ID,
	}

	if args.Input.Plan != nil {
		planID, err := unmarshalCampaignPlanID(*args.Input.Plan)
		if err != nil {
			return nil, err
		}
		campaign.CampaignPlanID = planID
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
	err = svc.CreateCampaign(ctx, campaign)
	if err != nil {
		return nil, err
	}

	go func() {
		ctx := trace.ContextWithTrace(context.Background(), tr)
		err := svc.RunChangesetJobs(ctx, campaign)
		if err != nil {
			log15.Error("RunChangesetJobs", "err", err)
		}
	}()

	return &campaignResolver{store: r.store, Campaign: campaign}, nil
}

func (r *Resolver) UpdateCampaign(ctx context.Context, args *graphqlbackend.UpdateCampaignArgs) (graphqlbackend.CampaignResolver, error) {
	// 🚨 SECURITY: Only site admins may update campaigns for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	campaignID, err := unmarshalCampaignID(args.Input.ID)
	if err != nil {
		return nil, err
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

	if args.Input.Name != nil {
		campaign.Name = *args.Input.Name
	}

	if args.Input.Description != nil {
		campaign.Description = *args.Input.Description
	}

	if err := tx.UpdateCampaign(ctx, campaign); err != nil {
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

	svc := ee.NewService(r.store, gitserver.DefaultClient, r.httpFactory)
	go func() {
		ctx := trace.ContextWithTrace(context.Background(), tr)
		err := svc.RunChangesetJobs(ctx, campaign)
		if err != nil {
			log15.Error("RunChangesetJobs", "err", err)
		}
	}()

	return &campaignResolver{store: r.store, Campaign: campaign}, nil
}

func (r *Resolver) Campaigns(ctx context.Context, args *graphqlutil.ConnectionArgs) (graphqlbackend.CampaignsConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins may read campaigns for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	return &campaignsConnectionResolver{
		store: r.store,
		opts: ee.ListCampaignsOpts{
			Limit: int(args.GetFirst()),
		},
	}, nil
}

func (r *Resolver) CreateChangesets(ctx context.Context, args *graphqlbackend.CreateChangesetsArgs) (_ []graphqlbackend.ExternalChangesetResolver, err error) {
	// 🚨 SECURITY: Only site admins may create changesets for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	var repoIDs []uint32
	repoSet := map[uint32]*repos.Repo{}
	cs := make([]*a8n.Changeset, 0, len(args.Input))

	for _, c := range args.Input {
		repoID, err := unmarshalRepositoryID(c.Repository)
		if err != nil {
			return nil, err
		}

		id := uint32(repoID)
		if _, ok := repoSet[id]; !ok {
			repoSet[id] = nil
			repoIDs = append(repoIDs, id)
		}

		cs = append(cs, &a8n.Changeset{
			RepoID:     int32(id),
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
		if !a8n.IsRepoSupported(&r.ExternalRepo) {
			err = errors.Errorf(
				"External service type %s of repository %q is currently not supported in Automation features",
				r.ExternalRepo.ServiceType,
				r.Name,
			)
			return nil, err
		}

		repoSet[r.ID] = r
	}

	for id, r := range repoSet {
		if r == nil {
			return nil, errors.Errorf("repo %v not found", marshalRepositoryID(api.RepoID(id)))
		}
	}

	for _, c := range cs {
		c.ExternalServiceType = repoSet[uint32(c.RepoID)].ExternalRepo.ServiceType
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
			preloadedRepo: repoSet[uint32(cs[i].RepoID)],
		}
	}

	return csr, nil
}

func (r *Resolver) Changesets(ctx context.Context, args *graphqlutil.ConnectionArgs) (graphqlbackend.ExternalChangesetsConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins may read changesets for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	return &changesetsConnectionResolver{
		store: r.store,
		opts: ee.ListChangesetsOpts{
			Limit: int(args.GetFirst()),
		},
	}, nil
}

func (r *Resolver) PreviewCampaignPlan(ctx context.Context, args graphqlbackend.PreviewCampaignPlanArgs) (graphqlbackend.CampaignPlanResolver, error) {
	var err error
	tr, ctx := trace.New(ctx, "Resolver.PreviewCampaignPlan", args.Specification.Type)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// 🚨 SECURITY: Only site admins may update campaigns for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	specArgs := string(args.Specification.Arguments)
	typeName := strings.ToLower(args.Specification.Type)
	if typeName == "" {
		return nil, errors.New("cannot create CampaignPlan without Type")
	}
	campaignType, err := ee.NewCampaignType(typeName, specArgs, r.httpFactory)
	if err != nil {
		return nil, err
	}
	plan := &a8n.CampaignPlan{CampaignType: typeName, Arguments: specArgs}
	runner := ee.NewRunner(r.store, campaignType, graphqlbackend.SearchRepos, nil)

	tr.LogFields(log.Int64("plan_id", plan.ID), log.Bool("Wait", args.Wait))

	if args.Wait {
		err := runner.Run(ctx, plan)
		if err != nil {
			return nil, err
		}
		err = runner.Wait()
		if err != nil {
			return nil, err
		}
	} else {
		act := actor.FromContext(ctx)
		ctx := trace.ContextWithTrace(context.Background(), tr)
		ctx = actor.WithActor(ctx, act)
		err := runner.Run(ctx, plan)
		if err != nil {
			return nil, err
		}
	}

	return &campaignPlanResolver{store: r.store, campaignPlan: plan}, nil
}

func (r *Resolver) CancelCampaignPlan(ctx context.Context, args graphqlbackend.CancelCampaignPlanArgs) (res *graphqlbackend.EmptyResponse, err error) {
	// 🚨 SECURITY: Only site admins may update campaigns for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	id, err := unmarshalCampaignPlanID(args.Plan)
	if err != nil {
		return nil, err
	}

	tx, err := r.store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	defer tx.Done(&err)

	plan, err := tx.GetCampaignPlan(ctx, ee.GetCampaignPlanOpts{ID: id})
	if err != nil {
		return nil, err
	}

	if !plan.CanceledAt.IsZero() {
		return &graphqlbackend.EmptyResponse{}, nil
	}

	status, err := tx.GetCampaignPlanStatus(ctx, id)
	if err != nil {
		return nil, err
	}

	if status.Finished() {
		return &graphqlbackend.EmptyResponse{}, nil
	}

	plan.CanceledAt = time.Now().UTC()

	err = tx.UpdateCampaignPlan(ctx, plan)

	return &graphqlbackend.EmptyResponse{}, err
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
