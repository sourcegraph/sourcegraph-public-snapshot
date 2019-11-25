package resolvers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ee "github.com/sourcegraph/sourcegraph/enterprise/pkg/a8n"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type changesetsConnectionResolver struct {
	store *ee.Store
	opts  ee.ListChangesetsOpts

	// cache results because they are used by multiple fields
	once       sync.Once
	changesets []*a8n.Changeset
	reposByID  map[int32]*repos.Repo
	next       int64
	err        error
}

func (r *changesetsConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ExternalChangesetResolver, error) {
	changesets, reposByID, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := make([]graphqlbackend.ExternalChangesetResolver, 0, len(changesets))
	for _, c := range changesets {
		repo, ok := reposByID[c.RepoID]
		if !ok {
			return nil, fmt.Errorf("failed to load repo %d", c.RepoID)
		}

		resolvers = append(resolvers, &changesetResolver{
			store:         r.store,
			Changeset:     c,
			preloadedRepo: repo,
		})
	}

	return resolvers, nil
}

func (r *changesetsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	opts := ee.CountChangesetsOpts{CampaignID: r.opts.CampaignID}
	count, err := r.store.CountChangesets(ctx, opts)
	return int32(count), err
}

func (r *changesetsConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, _, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(next != 0), nil
}

func (r *changesetsConnectionResolver) compute(ctx context.Context) ([]*a8n.Changeset, map[int32]*repos.Repo, int64, error) {
	r.once.Do(func() {
		r.changesets, r.next, r.err = r.store.ListChangesets(ctx, r.opts)
		if r.err != nil {
			return
		}

		reposStore := repos.NewDBStore(r.store.DB(), sql.TxOptions{})
		repoIDs := make([]uint32, len(r.changesets))
		for i, c := range r.changesets {
			repoIDs[i] = uint32(c.RepoID)
		}

		rs, err := reposStore.ListRepos(ctx, repos.StoreListReposArgs{IDs: repoIDs})
		if err != nil {
			r.err = err
			return
		}

		r.reposByID = make(map[int32]*repos.Repo, len(rs))
		for _, repo := range rs {
			r.reposByID[int32(repo.ID)] = repo
		}
	})

	return r.changesets, r.reposByID, r.next, r.err
}

type changesetResolver struct {
	store *ee.Store
	*a8n.Changeset
	preloadedRepo *repos.Repo

	// cache repo because it's called more than once
	once sync.Once
	repo *graphqlbackend.RepositoryResolver
	err  error
}

const changesetIDKind = "Changeset"

func marshalChangesetID(id int64) graphql.ID {
	return relay.MarshalID(changesetIDKind, id)
}

func unmarshalChangesetID(id graphql.ID) (cid int64, err error) {
	err = relay.UnmarshalSpec(id, &cid)
	return
}

func (r *changesetResolver) computeRepo(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	r.once.Do(func() {
		if r.preloadedRepo != nil {
			r.repo = newRepositoryResolver(r.preloadedRepo)
		} else {
			r.repo, r.err = graphqlbackend.RepositoryByIDInt32(ctx, api.RepoID(r.RepoID))
			if r.err != nil {
				return
			}
		}
	})
	return r.repo, r.err
}

func (r *changesetResolver) ID() graphql.ID {
	return marshalChangesetID(r.Changeset.ID)
}

func (r *changesetResolver) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return r.computeRepo(ctx)
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

	repo, err := r.computeRepo(ctx)
	if err != nil {
		return nil, err
	}

	base, err := r.Changeset.BaseRefOid()
	if err != nil {
		return nil, err
	}
	if base == "" {
		// Fallback to the ref name if we can't get the OID
		base, err = r.Changeset.BaseRefName()
		if err != nil {
			return nil, err
		}
	}

	head, err := r.Changeset.HeadRefOid()
	if err != nil {
		return nil, err
	}
	if head == "" {
		// Fallback to the ref name if we can't get the OID
		head, err = r.Changeset.HeadRefName()
		if err != nil {
			return nil, err
		}
	}

	return graphqlbackend.NewRepositoryComparison(ctx, repo, &graphqlbackend.RepositoryComparisonInput{
		Base: &base,
		Head: &head,
	})
}

func (r *changesetResolver) Head(ctx context.Context) (*graphqlbackend.GitRefResolver, error) {
	name, err := r.Changeset.HeadRefName()
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, errors.New("changeset head ref name could not be determined")
	}

	oid, err := r.Changeset.HeadRefOid()
	if err != nil {
		return nil, err
	}

	return r.gitRef(ctx, name, oid)
}

func (r *changesetResolver) Base(ctx context.Context) (*graphqlbackend.GitRefResolver, error) {
	name, err := r.Changeset.BaseRefName()
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, errors.New("changeset base ref name could not be determined")
	}

	oid, err := r.Changeset.BaseRefOid()
	if err != nil {
		return nil, err
	}

	return r.gitRef(ctx, name, oid)
}

func (r *changesetResolver) gitRef(ctx context.Context, name, oid string) (*graphqlbackend.GitRefResolver, error) {
	repo, err := r.computeRepo(ctx)
	if err != nil {
		return nil, err
	}

	if oid == "" {
		commitID, err := r.commitID(ctx, repo, name)
		if err != nil {
			return nil, err
		}
		oid = string(commitID)
	}

	return graphqlbackend.NewGitRefResolver(repo, name, graphqlbackend.GitObjectID(oid)), nil
}

func (r *changesetResolver) commitID(ctx context.Context, repo *graphqlbackend.RepositoryResolver, refName string) (api.CommitID, error) {
	grepo, err := backend.CachedGitRepo(ctx, &types.Repo{
		ExternalRepo: *repo.ExternalRepo(),
		Name:         api.RepoName(repo.Name()),
	})
	if err != nil {
		return api.CommitID(""), err
	}
	// Call ResolveRevision to trigger fetches from remote (in case base/head commits don't
	// exist).
	return git.ResolveRevision(ctx, *grepo, nil, refName, nil)
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
