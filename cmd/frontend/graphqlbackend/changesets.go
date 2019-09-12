package graphqlbackend

import (
	"context"
	"sync"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/pkg/a8n"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func (r *schemaResolver) CreateChangeset(ctx context.Context, args *struct {
	Repository graphql.ID
	ExternalID string
}) (_ *changesetResolver, err error) {
	// 🚨 SECURITY: Only site admins may create changesets for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	repoID, err := unmarshalRepositoryID(args.Repository)
	if err != nil {
		return nil, err
	}

	changeset := &a8n.Changeset{
		RepoID:     int32(repoID),
		ExternalID: args.ExternalID,
	}

	if err = r.A8NStore.CreateChangeset(ctx, changeset); err != nil {
		return nil, err
	}

	// TODO(tsenart): Sync change-set metadata.

	return &changesetResolver{store: r.A8NStore, Changeset: changeset}, nil
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

type changesetResolver struct {
	store *a8n.Store
	*a8n.Changeset
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
