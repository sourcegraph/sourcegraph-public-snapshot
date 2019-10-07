package resolvers

import (
	"context"
	"database/sql"
	"path"
	"sync"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ee "github.com/sourcegraph/sourcegraph/enterprise/pkg/a8n"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
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

func (r *Resolver) ChangesetByID(ctx context.Context, id graphql.ID) (graphqlbackend.ChangesetResolver, error) {
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

	switch relay.UnmarshalKind(args.Input.Namespace) {
	case "User":
		relay.UnmarshalSpec(args.Input.Namespace, &campaign.NamespaceUserID)
	case "Org":
		relay.UnmarshalSpec(args.Input.Namespace, &campaign.NamespaceOrgID)
	default:
		return nil, errors.Errorf("Invalid namespace %q", args.Input.Namespace)
	}

	if err := r.store.CreateCampaign(ctx, campaign); err != nil {
		return nil, err
	}

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

func (r *Resolver) DeleteCampaign(ctx context.Context, args *graphqlbackend.DeleteCampaignArgs) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins may update campaigns for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	campaignID, err := unmarshalCampaignID(args.Campaign)
	if err != nil {
		return nil, err
	}

	err = r.store.DeleteCampaign(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, nil
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

type campaignsConnectionResolver struct {
	store *ee.Store
	opts  ee.ListCampaignsOpts

	// cache results because they are used by multiple fields
	once      sync.Once
	campaigns []*a8n.Campaign
	next      int64
	err       error
}

func (r *campaignsConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.CampaignResolver, error) {
	campaigns, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.CampaignResolver, 0, len(campaigns))
	for _, c := range campaigns {
		resolvers = append(resolvers, &campaignResolver{store: r.store, Campaign: c})
	}
	return resolvers, nil
}

func (r *campaignsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	opts := ee.CountCampaignsOpts{ChangesetID: r.opts.ChangesetID}
	count, err := r.store.CountCampaigns(ctx, opts)
	return int32(count), err
}

func (r *campaignsConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(next != 0), nil
}

func (r *campaignsConnectionResolver) compute(ctx context.Context) ([]*a8n.Campaign, int64, error) {
	r.once.Do(func() {
		r.campaigns, r.next, r.err = r.store.ListCampaigns(ctx, r.opts)
	})
	return r.campaigns, r.next, r.err
}

type campaignResolver struct {
	store *ee.Store
	*a8n.Campaign
}

const campaignIDKind = "Campaign"

func marshalCampaignID(id int64) graphql.ID {
	return relay.MarshalID(campaignIDKind, id)
}

func unmarshalCampaignID(id graphql.ID) (campaignID int64, err error) {
	err = relay.UnmarshalSpec(id, &campaignID)
	return
}

func (r *campaignResolver) ID() graphql.ID {
	return marshalCampaignID(r.Campaign.ID)
}

func (r *campaignResolver) Name() string {
	return r.Campaign.Name
}

func (r *campaignResolver) Description() string {
	return r.Campaign.Description
}

func (r *campaignResolver) Author(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	return graphqlbackend.UserByIDInt32(ctx, r.AuthorID)
}

func (r *campaignResolver) URL(ctx context.Context) (string, error) {
	// TODO(tsenart): Query for namespace only once
	ns, err := r.Namespace(ctx)
	if err != nil {
		return "", err
	}
	return path.Join(ns.URL(), "campaigns", string(r.ID())), nil
}

func (r *campaignResolver) Namespace(ctx context.Context) (n graphqlbackend.NamespaceResolver, err error) {
	if r.NamespaceUserID != 0 {
		n.Namespace, err = graphqlbackend.UserByIDInt32(ctx, r.NamespaceUserID)
	} else {
		n.Namespace, err = graphqlbackend.OrgByIDInt32(ctx, r.NamespaceOrgID)
	}

	return n, err
}

func (r *campaignResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.Campaign.CreatedAt}
}

func (r *campaignResolver) UpdatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.Campaign.UpdatedAt}
}

func (r *campaignResolver) Changesets(ctx context.Context, args struct {
	graphqlutil.ConnectionArgs
}) graphqlbackend.ChangesetsConnectionResolver {
	return &changesetsConnectionResolver{
		store: r.store,
		opts: ee.ListChangesetsOpts{
			CampaignID: r.Campaign.ID,
			Limit:      int(args.ConnectionArgs.GetFirst()),
		},
	}
}

func (r *campaignResolver) ChangesetCountsOverTime(
	ctx context.Context,
	args *graphqlbackend.ChangesetCountsArgs,
) ([]graphqlbackend.ChangesetCountsResolver, error) {
	resolvers := []graphqlbackend.ChangesetCountsResolver{}
	resolvers = append(resolvers, &changesetCountsResolver{
		date:                 time.Now(),
		total:                99,
		merged:               99,
		closed:               99,
		open:                 99,
		openApproved:         99,
		openChangesRequested: 99,
		openPending:          99,
	})
	return resolvers, nil
}

func (r *Resolver) CreateChangesets(ctx context.Context, args *graphqlbackend.CreateChangesetsArgs) (_ []graphqlbackend.ChangesetResolver, err error) {
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

	if err = tx.CreateChangesets(ctx, cs...); err != nil {
		return nil, err
	}

	tx.Done()

	// Only fetch metadata if none of these changesets existed before.
	// We do this outside of a transaction.

	store = repos.NewDBStore(r.store.DB(), sql.TxOptions{})
	syncer := ee.ChangesetSyncer{
		ReposStore:  store,
		Store:       r.store,
		HTTPFactory: r.httpFactory,
	}
	if err = syncer.SyncChangesets(ctx, cs...); err != nil {
		return nil, err
	}

	csr := make([]graphqlbackend.ChangesetResolver, len(cs))
	for i := range cs {
		csr[i] = &changesetResolver{
			store:     r.store,
			Changeset: cs[i],
			repo:      repoSet[uint32(cs[i].RepoID)],
		}
	}

	return csr, nil
}

func (r *Resolver) Changesets(ctx context.Context, args *graphqlutil.ConnectionArgs) (graphqlbackend.ChangesetsConnectionResolver, error) {
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

type changesetsConnectionResolver struct {
	store *ee.Store
	opts  ee.ListChangesetsOpts

	// cache results because they are used by multiple fields
	once       sync.Once
	changesets []*a8n.Changeset
	next       int64
	err        error
}

func (r *changesetsConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ChangesetResolver, error) {
	changesets, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.ChangesetResolver, 0, len(changesets))
	for _, c := range changesets {
		resolvers = append(resolvers, &changesetResolver{store: r.store, Changeset: c})
	}
	return resolvers, nil
}

func (r *changesetsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	opts := ee.CountChangesetsOpts{CampaignID: r.opts.CampaignID}
	count, err := r.store.CountChangesets(ctx, opts)
	return int32(count), err
}

func (r *changesetsConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(next != 0), nil
}

func (r *changesetsConnectionResolver) compute(ctx context.Context) ([]*a8n.Changeset, int64, error) {
	r.once.Do(func() {
		r.changesets, r.next, r.err = r.store.ListChangesets(ctx, r.opts)
	})
	return r.changesets, r.next, r.err
}

type changesetResolver struct {
	store *ee.Store
	*a8n.Changeset
	repo *repos.Repo
}

const changesetIDKind = "Changeset"

func marshalChangesetID(id int64) graphql.ID {
	return relay.MarshalID(changesetIDKind, id)
}

func unmarshalChangesetID(id graphql.ID) (cid int64, err error) {
	err = relay.UnmarshalSpec(id, &cid)
	return
}

func (r *changesetResolver) ID() graphql.ID {
	return marshalChangesetID(r.Changeset.ID)
}

func (r *changesetResolver) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	if r.repo != nil {
		return graphqlbackend.NewRepositoryResolver(&types.Repo{
			ID:           api.RepoID(r.repo.ID),
			ExternalRepo: r.repo.ExternalRepo,
			Name:         api.RepoName(r.repo.Name),
			RepoFields: &types.RepoFields{
				URI:         r.repo.URI,
				Description: r.repo.Description,
				Language:    r.repo.Language,
				Fork:        r.repo.Fork,
			},
		}), nil
	}
	return graphqlbackend.RepositoryByIDInt32(ctx, api.RepoID(r.Changeset.RepoID))
}

func (r *changesetResolver) Campaigns(ctx context.Context, args *struct {
	graphqlutil.ConnectionArgs
}) (graphqlbackend.CampaignsConnectionResolver, error) {
	return &campaignsConnectionResolver{
		store: r.store,
		opts: ee.ListCampaignsOpts{
			ChangesetID: r.Changeset.ID,
			Limit:       int(args.ConnectionArgs.GetFirst()),
		},
	}, nil
}

func (r *changesetResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.Changeset.CreatedAt}
}

func (r *changesetResolver) UpdatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.Changeset.UpdatedAt}
}

func (r *changesetResolver) Title() (string, error) {
	return r.Changeset.Title()
}

func (r *changesetResolver) Body() (string, error) {
	return r.Changeset.Body()
}

func (r *changesetResolver) State() (a8n.ChangesetState, error) {
	return r.Changeset.State()
}

func (r *changesetResolver) ExternalURL() (*externallink.Resolver, error) {
	url, err := r.Changeset.URL()
	if err != nil {
		return nil, err
	}
	return externallink.NewResolver(url, r.Changeset.ExternalServiceType), nil
}

func (r *changesetResolver) ReviewState() (a8n.ChangesetReviewState, error) {
	return r.Changeset.ReviewState()
}

func (r *changesetResolver) Events(ctx context.Context, args *struct {
	graphqlutil.ConnectionArgs
}) (graphqlbackend.ChangesetEventsConnectionResolver, error) {
	return &changesetEventsConnectionResolver{
		store:     r.store,
		changeset: r.Changeset,
		opts: ee.ListChangesetEventsOpts{
			ChangesetID: r.Changeset.ID,
			Limit:       int(args.ConnectionArgs.GetFirst()),
		},
	}, nil
}

type changesetEventsConnectionResolver struct {
	store     *ee.Store
	changeset *a8n.Changeset
	opts      ee.ListChangesetEventsOpts

	// cache results because they are used by multiple fields
	once            sync.Once
	changesetEvents []*a8n.ChangesetEvent
	next            int64
	err             error
}

func (r *changesetEventsConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ChangesetEventResolver, error) {
	changesetEvents, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.ChangesetEventResolver, 0, len(changesetEvents))
	for _, c := range changesetEvents {
		resolvers = append(resolvers, &changesetEventResolver{store: r.store, changeset: r.changeset, ChangesetEvent: c})
	}
	return resolvers, nil
}

func (r *changesetEventsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	opts := ee.CountChangesetEventsOpts{ChangesetID: r.opts.ChangesetID}
	count, err := r.store.CountChangesetEvents(ctx, opts)
	return int32(count), err
}

func (r *changesetEventsConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(next != 0), nil
}

func (r *changesetEventsConnectionResolver) compute(ctx context.Context) ([]*a8n.ChangesetEvent, int64, error) {
	r.once.Do(func() {
		r.changesetEvents, r.next, r.err = r.store.ListChangesetEvents(ctx, r.opts)
	})
	return r.changesetEvents, r.next, r.err
}

type changesetEventResolver struct {
	store     *ee.Store
	changeset *a8n.Changeset
	*a8n.ChangesetEvent
}

const changesetEventIDKind = "ChangesetEvent"

func marshalchangesetEventID(id int64) graphql.ID {
	return relay.MarshalID(changesetEventIDKind, id)
}

func (r *changesetEventResolver) ID() graphql.ID {
	return marshalchangesetEventID(r.ChangesetEvent.ID)
}

func (r *changesetEventResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.ChangesetEvent.CreatedAt}
}

func (r *changesetEventResolver) Changeset(ctx context.Context) (graphqlbackend.ChangesetResolver, error) {
	return &changesetResolver{store: r.store, Changeset: r.changeset}, nil
}

func marshalRepositoryID(repo api.RepoID) graphql.ID { return relay.MarshalID("Repository", repo) }
func unmarshalRepositoryID(id graphql.ID) (repo api.RepoID, err error) {
	err = relay.UnmarshalSpec(id, &repo)
	return
}

type changesetCountsResolver struct {
	date                 time.Time
	total                int32
	merged               int32
	closed               int32
	open                 int32
	openApproved         int32
	openChangesRequested int32
	openPending          int32
}

func (r *changesetCountsResolver) Date() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{time.Now()}
}

func (r *changesetCountsResolver) Total() int32 {
	return r.total
}

func (r *changesetCountsResolver) Merged() int32 {
	return r.merged
}

func (r *changesetCountsResolver) Closed() int32 {
	return r.closed
}

func (r *changesetCountsResolver) Open() int32 {
	return r.open
}

func (r *changesetCountsResolver) OpenApproved() int32 {
	return r.openApproved
}

func (r *changesetCountsResolver) OpenChangesRequested() int32 {
	return r.openChangesRequested
}

func (r *changesetCountsResolver) OpenPending() int32 {
	return r.openPending
}
