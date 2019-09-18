package a8n

import (
	"context"
	"database/sql"
	"path"
	"sync"

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
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
)

// Resolver is the GraphQL resolver of all things A8N.
type Resolver struct {
	store       *Store
	HTTPFactory *httpcli.Factory
}

func (r *Resolver) ChangesetByID(ctx context.Context, s *Store, id graphql.ID) (graphqlbackend.ChangesetResolver, error) {
	// 🚨 SECURITY: Only site admins may access changesets for now.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	changesetID, err := unmarshalChangesetID(id)
	if err != nil {
		return nil, err
	}

	changeset, err := s.GetChangeset(ctx, GetChangesetOpts{ID: changesetID})
	if err != nil {
		return nil, err
	}

	return &changesetResolver{store: s, Changeset: changeset}, nil
}

func (r *Resolver) CampaignByID(ctx context.Context, s *Store, id graphql.ID) (graphqlbackend.CampaignResolver, error) {
	// 🚨 SECURITY: Only site admins may access campaigns for now.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	campaignID, err := unmarshalCampaignID(id)
	if err != nil {
		return nil, err
	}

	campaign, err := s.GetCampaign(ctx, GetCampaignOpts{ID: campaignID})
	if err != nil {
		return nil, err
	}

	return &campaignResolver{store: s, Campaign: campaign}, nil
}

func (r *Resolver) AddChangesetsToCampaign(ctx context.Context, args *graphqlbackend.AddChangesetsToCampaignArgs) (_ *campaignResolver, err error) {
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

	campaign, err := tx.GetCampaign(ctx, GetCampaignOpts{ID: campaignID})
	if err != nil {
		return nil, err
	}

	changesets, _, err := tx.ListChangesets(ctx, ListChangesetsOpts{IDs: changesetIDs})
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

func (r *Resolver) CreateCampaign(ctx context.Context, args *graphqlbackend.CreateCampaignArgs) (*campaignResolver, error) {
	user, err := db.Users.GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "%v", backend.ErrNotAuthenticated)
	}

	// 🚨 SECURITY: Only site admins may create a campaign for now.
	if !user.SiteAdmin {
		return nil, backend.ErrMustBeSiteAdmin
	}

	campaign := &Campaign{
		Name:        args.Input.Name,
		Description: args.Input.Description,
		AuthorID:    user.ID,
	}

	node, err := NodeByID(ctx, r.store, args.Input.Namespace)
	if err != nil {
		return nil, err
	}

	switch ns := node.(type) {
	case *UserResolver:
		campaign.NamespaceUserID = ns.DatabaseID()
	case *OrgResolver:
		campaign.NamespaceOrgID = ns.OrgID()
	default:
		return nil, errors.Errorf("Invalid namespace of type %T", ns)
	}

	if err := r.store.CreateCampaign(ctx, campaign); err != nil {
		return nil, err
	}

	return &campaignResolver{store: r.store, Campaign: campaign}, nil
}

func (r *Resolver) Campaigns(ctx context.Context, args *graphqlutil.ConnectionArgs) (*campaignsConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins may read campaigns for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	return &campaignsConnectionResolver{
		store: r.store,
		opts: ListCampaignsOpts{
			Limit: int(args.GetFirst()),
		},
	}, nil
}

type campaignsConnectionResolver struct {
	store *Store
	opts  ListCampaignsOpts

	// cache results because they are used by multiple fields
	once      sync.Once
	campaigns []*Campaign
	next      int64
	err       error
}

func (r *campaignsConnectionResolver) Nodes(ctx context.Context) ([]*campaignResolver, error) {
	campaigns, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*campaignResolver, 0, len(campaigns))
	for _, c := range campaigns {
		resolvers = append(resolvers, &campaignResolver{store: r.store, Campaign: c})
	}
	return resolvers, nil
}

func (r *campaignsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	opts := CountCampaignsOpts{ChangesetID: r.opts.ChangesetID}
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

func (r *campaignsConnectionResolver) compute(ctx context.Context) ([]*Campaign, int64, error) {
	r.once.Do(func() {
		r.campaigns, r.next, r.err = r.store.ListCampaigns(ctx, r.opts)
	})
	return r.campaigns, r.next, r.err
}

type campaignResolver struct {
	store *Store
	*Campaign
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

func (r *campaignResolver) Author(ctx context.Context) (*UserResolver, error) {
	return UserByIDInt32(ctx, r.AuthorID)
}

func (r *campaignResolver) URL(ctx context.Context) (string, error) {
	// TODO(tsenart): Query for namespace only once
	ns, err := r.Namespace(ctx)
	if err != nil {
		return "", err
	}
	return path.Join(ns.URL(), "campaigns", string(r.ID())), nil
}

func (r *campaignResolver) Namespace(ctx context.Context) (n namespaceResolver, err error) {
	if r.NamespaceUserID != 0 {
		n.Namespace, err = UserByIDInt32(ctx, r.NamespaceUserID)
	} else {
		n.Namespace, err = OrgByIDInt32(ctx, r.NamespaceOrgID)
	}

	return n, err
}

func (r *campaignResolver) CreatedAt() DateTime {
	return DateTime{Time: r.Campaign.CreatedAt}
}

func (r *campaignResolver) UpdatedAt() DateTime {
	return DateTime{Time: r.Campaign.UpdatedAt}
}

func (r *campaignResolver) Changesets(ctx context.Context, args struct {
	graphqlutil.ConnectionArgs
}) *changesetsConnectionResolver {
	return &changesetsConnectionResolver{
		store: r.store,
		opts: ListChangesetsOpts{
			CampaignID: r.Campaign.ID,
			Limit:      int(args.ConnectionArgs.GetFirst()),
		},
	}
}

type createChangesetInput struct {
	Repository graphql.ID
	ExternalID string
}

func (r *Resolver) CreateChangesets(ctx context.Context, args *CreateChangesetsArgs) (_ []*changesetResolver, err error) {
	// 🚨 SECURITY: Only site admins may create changesets for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	var repoIDs []uint32
	repoSet := map[uint32]*repos.Repo{}
	cs := make([]*Changeset, 0, len(args.Input))

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

		cs = append(cs, &Changeset{
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
	es, err := store.ListExternalServices(ctx, repos.StoreListExternalServicesArgs{RepoIDs: repoIDs})
	if err != nil {
		return nil, err
	}

	byRepo := make(map[uint32]int64, len(rs))
	for _, r := range rs {
		eids := r.ExternalServiceIDs()
		for _, id := range eids {
			if _, ok := byRepo[r.ID]; !ok {
				byRepo[r.ID] = id
				break
			}
		}
	}

	type batch struct {
		repos.ChangesetSource
		Changesets []*repos.Changeset
	}

	batches := make(map[int64]*batch, len(es))
	for _, e := range es {
		src, err := repos.NewSource(e, r.HTTPFactory)
		if err != nil {
			return nil, err
		}

		css, ok := src.(repos.ChangesetSource)
		if !ok {
			return nil, errors.Errorf("unsupported repo type %q", e.Kind)
		}

		batches[e.ID] = &batch{ChangesetSource: css}
	}

	for _, c := range cs {
		repoID := uint32(c.RepoID)
		b := batches[byRepo[repoID]]
		b.Changesets = append(b.Changesets, &repos.Changeset{
			Changeset: c,
			Repo:      repoSet[repoID],
		})
	}

	for _, b := range batches {
		if err = b.LoadChangesets(ctx, b.Changesets...); err != nil {
			return nil, err
		}
	}

	if err = r.store.UpdateChangesets(ctx, cs...); err != nil {
		return nil, err
	}

	csr := make([]*changesetResolver, len(cs))
	for i := range cs {
		csr[i] = &changesetResolver{
			store:     r.store,
			Changeset: cs[i],
			repo:      repoSet[uint32(cs[i].RepoID)],
		}
	}

	return csr, nil
}

func (r *Resolver) Changesets(ctx context.Context, args *graphqlutil.ConnectionArgs) (*changesetsConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins may read changesets for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	return &changesetsConnectionResolver{
		store: r.store,
		opts: ListChangesetsOpts{
			Limit: int(args.GetFirst()),
		},
	}, nil
}

type changesetsConnectionResolver struct {
	store *Store
	opts  ListChangesetsOpts

	// cache results because they are used by multiple fields
	once       sync.Once
	changesets []*Changeset
	next       int64
	err        error
}

func (r *changesetsConnectionResolver) Nodes(ctx context.Context) ([]*changesetResolver, error) {
	changesets, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*changesetResolver, 0, len(changesets))
	for _, c := range changesets {
		resolvers = append(resolvers, &changesetResolver{store: r.store, Changeset: c})
	}
	return resolvers, nil
}

func (r *changesetsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	opts := CountChangesetsOpts{CampaignID: r.opts.CampaignID}
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

func (r *changesetsConnectionResolver) compute(ctx context.Context) ([]*Changeset, int64, error) {
	r.once.Do(func() {
		r.changesets, r.next, r.err = r.store.ListChangesets(ctx, r.opts)
	})
	return r.changesets, r.next, r.err
}

type changesetResolver struct {
	store *Store
	*Changeset
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

func (r *changesetResolver) Repository(ctx context.Context) (*RepositoryResolver, error) {
	if r.repo != nil {
		return &RepositoryResolver{
			repo: &types.Repo{
				ID:           api.RepoID(r.repo.ID),
				ExternalRepo: r.repo.ExternalRepo,
				Name:         api.RepoName(r.repo.Name),
				RepoFields: &types.RepoFields{
					URI:         r.repo.URI,
					Description: r.repo.Description,
					Language:    r.repo.Language,
					Fork:        r.repo.Fork,
				},
			},
		}, nil
	}
	return repositoryByIDInt32(ctx, api.RepoID(r.Changeset.RepoID))
}

func (r *changesetResolver) Campaigns(ctx context.Context, args struct {
	graphqlutil.ConnectionArgs
}) *campaignsConnectionResolver {
	return &campaignsConnectionResolver{
		store: r.store,
		opts: ListCampaignsOpts{
			ChangesetID: r.Changeset.ID,
			Limit:       int(args.ConnectionArgs.GetFirst()),
		},
	}
}

func (r *changesetResolver) CreatedAt() DateTime {
	return DateTime{Time: r.Changeset.CreatedAt}
}

func (r *changesetResolver) UpdatedAt() DateTime {
	return DateTime{Time: r.Changeset.UpdatedAt}
}

func (r *changesetResolver) Title() (string, error) {
	return r.Changeset.Title()
}

func (r *changesetResolver) Body() (string, error) {
	return r.Changeset.Body()
}

func (r *changesetResolver) State() (string, error) {
	// TODO(mrnugget): Let's see if we can use a type instead of string,
	// without circular reference between a8n/graphqlbackend
	return string(r.Changeset.State())
}

func (r *changesetResolver) ExternalURL() (*externallink.Resolver, error) {
	url, err := r.Changeset.URL()
	if err != nil {
		return nil, err
	}
	return externallink.NewResolver(url, r.Changeset.ExternalServiceType), nil
}

func (r *changesetResolver) ReviewState() (string, error) {
	// TODO(mrnugget): Let's see if we can use a type instead of string,
	// without circular reference between a8n/graphqlbackend
	return string(r.Changeset.ReviewState())
}
