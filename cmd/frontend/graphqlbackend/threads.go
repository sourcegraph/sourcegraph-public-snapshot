package graphqlbackend

import (
	"context"
	"sync"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/a8n"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func (r *schemaResolver) AddThreadFromURLToCampaign(ctx context.Context, args *struct {
	URL      string
	Campaign graphql.ID
}) (_ *threadResolver, err error) {
	var user *types.User
	user, err = db.Users.GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "%v", backend.ErrNotAuthenticated)
	}

	// 🚨 SECURITY: Only site admins may create a thread for now.
	if !user.SiteAdmin {
		return nil, backend.ErrMustBeSiteAdmin
	}

	var campaignID int64
	campaignID, err = unmarshalCampaignID(args.Campaign)
	if err != nil {
		return nil, err
	}

	var tx *a8n.Store
	tx, err = r.A8NStore.Transact(ctx)
	if err != nil {
		return nil, err
	}

	defer tx.Done(&err)

	var campaign *a8n.Campaign
	campaign, err = tx.GetCampaign(ctx, a8n.GetCampaignOpts{ID: campaignID})
	if err != nil {
		return nil, err
	}

	var repo *types.Repo
	repo, err = db.Repos.GetByName(ctx, api.RepoName(issueURLToRepoURL(args.URL)))
	if err != nil {
		return nil, err
	}

	thread := &a8n.Thread{
		RepoID:      int32(repo.ID),
		CampaignIDs: []int64{campaign.ID},
	}

	if err = tx.CreateThread(ctx, thread); err != nil {
		return nil, err
	}

	return &threadResolver{store: r.A8NStore, Thread: thread}, nil
}

func (r *schemaResolver) Threads(ctx context.Context, args *struct {
	graphqlutil.ConnectionArgs
}) (*threadsConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins may read external services (they have secrets).
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	return &threadsConnectionResolver{
		store: r.A8NStore,
		opts: a8n.ListThreadsOpts{
			Limit: int(args.ConnectionArgs.GetFirst()),
		},
	}, nil
}

type threadsConnectionResolver struct {
	store *a8n.Store
	opts  a8n.ListThreadsOpts

	// cache results because they are used by multiple fields
	once    sync.Once
	threads []*a8n.Thread
	next    int64
	err     error
}

func (r *threadsConnectionResolver) Nodes(ctx context.Context) ([]*threadResolver, error) {
	threads, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*threadResolver, 0, len(threads))
	for _, c := range threads {
		resolvers = append(resolvers, &threadResolver{Thread: c})
	}
	return resolvers, nil
}

func (r *threadsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	opts := a8n.CountThreadsOpts{CampaignID: r.opts.CampaignID}
	count, err := r.store.CountThreads(ctx, opts)
	return int32(count), err
}

func (r *threadsConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(next != 0), nil
}

func (r *threadsConnectionResolver) compute(ctx context.Context) ([]*a8n.Thread, int64, error) {
	r.once.Do(func() {
		r.threads, r.next, r.err = r.store.ListThreads(ctx, r.opts)
	})
	return r.threads, r.next, r.err
}

type threadResolver struct {
	store *a8n.Store
	*a8n.Thread
}

const threadIDKind = "Thread"

func marshalThreadID(id int64) graphql.ID {
	return relay.MarshalID(threadIDKind, id)
}

func (r *threadResolver) ID() graphql.ID {
	return marshalThreadID(r.Thread.ID)
}

func (r *threadResolver) Repository(ctx context.Context) (*RepositoryResolver, error) {
	return repositoryByIDInt32(ctx, api.RepoID(r.Thread.RepoID))
}

func (r *threadResolver) Campaigns(ctx context.Context, args struct {
	graphqlutil.ConnectionArgs
}) *campaignsConnectionResolver {
	return &campaignsConnectionResolver{
		store: r.store,
		opts: a8n.ListCampaignsOpts{
			ThreadID: r.Thread.ID,
			Limit:    int(args.ConnectionArgs.GetFirst()),
		},
	}
}

func (r *threadResolver) CreatedAt() DateTime {
	return DateTime{Time: r.Thread.CreatedAt}
}

func (r *threadResolver) UpdatedAt() DateTime {
	return DateTime{Time: r.Thread.UpdatedAt}
}

func issueURLToRepoURL(url string) string {
	// TODO: here be dragons
	// 1. Parse URL
	// 2. Determine code host
	// 3. According to which code host it is, go from issue URL to repoURL (i.e. cut off "issues/1")
	return "github.com/sourcegraph/sourcegraph"
}
