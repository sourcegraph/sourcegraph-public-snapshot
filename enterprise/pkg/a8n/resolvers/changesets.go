package resolvers

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ee "github.com/sourcegraph/sourcegraph/enterprise/pkg/a8n"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

type changesetsConnectionResolver struct {
	store *ee.Store
	opts  ee.ListChangesetsOpts

	// cache results because they are used by multiple fields
	once       sync.Once
	changesets []*a8n.Changeset
	next       int64
	err        error
}

func (r *changesetsConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ExternalChangesetResolver, error) {
	changesets, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.ExternalChangesetResolver, 0, len(changesets))
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
	return r.repoResolver(ctx)
}

func (r *changesetResolver) repoResolver(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	if r.repo != nil {
		return newRepositoryResolver(r.repo), nil
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

func (r *changesetResolver) ReviewState(ctx context.Context) (a8n.ChangesetReviewState, error) {
	// ChangesetEvents are currently only implemented for GitHub. For other
	// codehosts we compute the ReviewState from the Metadata field of a
	// Changeset.
	if _, ok := r.Changeset.Metadata.(*github.PullRequest); !ok {
		return r.Changeset.ReviewState()
	}

	opts := ee.ListChangesetEventsOpts{
		ChangesetIDs: []int64{r.Changeset.ID},
		Limit:        -1,
	}
	es, _, err := r.store.ListChangesetEvents(ctx, opts)
	if err != nil {
		return a8n.ChangesetReviewStatePending, err
	}

	events := make(a8n.ChangesetEvents, len(es))
	for i, e := range es {
		events[i] = e
	}

	sort.Sort(events)

	return events.ReviewState()
}

func (r *changesetResolver) Events(ctx context.Context, args *struct {
	graphqlutil.ConnectionArgs
}) (graphqlbackend.ChangesetEventsConnectionResolver, error) {
	return &changesetEventsConnectionResolver{
		store:     r.store,
		changeset: r.Changeset,
		opts: ee.ListChangesetEventsOpts{
			ChangesetIDs: []int64{r.Changeset.ID},
			Limit:        int(args.ConnectionArgs.GetFirst()),
		},
	}, nil
}

func (r *changesetResolver) Diff(ctx context.Context) (*graphqlbackend.RepositoryComparisonResolver, error) {
	s, err := r.Changeset.State()
	if err != nil {
		return nil, err
	}

	// Only return diffs for open changesets, otherwise we can't guarantee that
	// we have the refs on gitserver
	if s != a8n.ChangesetStateOpen {
		return nil, nil
	}

	repo, err := r.repoResolver(ctx)
	if err != nil {
		return nil, err
	}

	base, err := r.Changeset.BaseRefOid()
	if err != nil {
		return nil, err
	}
	if base == "" {
		return nil, errors.New("changeset base ref name could not be determined")
	}

	head, err := r.Changeset.HeadRefOid()
	if err != nil {
		return nil, err
	}
	if head == "" {
		return nil, errors.New("changeset head ref name could not be determined")
	}

	return graphqlbackend.NewRepositoryComparison(ctx, repo, &graphqlbackend.RepositoryComparisonInput{
		Base: &base,
		Head: &head,
	})
}

func newRepositoryResolver(r *repos.Repo) *graphqlbackend.RepositoryResolver {
	return graphqlbackend.NewRepositoryResolver(&types.Repo{
		ID:           api.RepoID(r.ID),
		ExternalRepo: r.ExternalRepo,
		Name:         api.RepoName(r.Name),
		RepoFields: &types.RepoFields{
			URI:         r.URI,
			Description: r.Description,
			Language:    r.Language,
			Fork:        r.Fork,
		},
	})
}
