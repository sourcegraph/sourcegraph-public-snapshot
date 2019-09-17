package graphqlbackend

import (
	"context"
	"database/sql"
	"sync"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/pkg/a8n"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

type createChangesetInput struct {
	Repository graphql.ID
	ExternalID string
}

func (r *schemaResolver) CreateChangesets(ctx context.Context, args *struct {
	Input []createChangesetInput
}) (_ []*changesetResolver, err error) {
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

	tx, err := r.A8NStore.Transact(ctx)
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

	store = repos.NewDBStore(r.A8NStore.DB(), sql.TxOptions{})
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

	if err = r.A8NStore.UpdateChangesets(ctx, cs...); err != nil {
		return nil, err
	}

	csr := make([]*changesetResolver, len(cs))
	for i := range cs {
		csr[i] = &changesetResolver{
			store:     r.A8NStore,
			Changeset: cs[i],
			repo:      repoSet[uint32(cs[i].RepoID)],
		}
	}

	return csr, nil
}

func (r *schemaResolver) Changesets(ctx context.Context, args *struct {
	graphqlutil.ConnectionArgs
}) (*changesetsConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins may read changesets for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	return &changesetsConnectionResolver{
		store: r.A8NStore,
		opts: a8n.ListChangesetsOpts{
			Limit: int(args.ConnectionArgs.GetFirst()),
		},
	}, nil
}

type changesetsConnectionResolver struct {
	store *a8n.Store
	opts  a8n.ListChangesetsOpts

	// cache results because they are used by multiple fields
	once       sync.Once
	changesets []*a8n.Changeset
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
	opts := a8n.CountChangesetsOpts{CampaignID: r.opts.CampaignID}
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

func changesetByID(ctx context.Context, s *a8n.Store, id graphql.ID) (*changesetResolver, error) {
	// 🚨 SECURITY: Only site admins may access changesets for now.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	changesetID, err := unmarshalChangesetID(id)
	if err != nil {
		return nil, err
	}

	changeset, err := s.GetChangeset(ctx, a8n.GetChangesetOpts{ID: changesetID})
	if err != nil {
		return nil, err
	}

	return &changesetResolver{store: s, Changeset: changeset}, nil
}

type changesetResolver struct {
	store *a8n.Store
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
		opts: a8n.ListCampaignsOpts{
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
