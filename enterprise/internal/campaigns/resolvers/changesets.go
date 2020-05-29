package resolvers

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type changesetsConnectionResolver struct {
	store *ee.Store
	opts  ee.ListChangesetsOpts

	// cache results because they are used by multiple fields
	once           sync.Once
	changesets     []*campaigns.Changeset
	scheduledSyncs map[int64]time.Time
	reposByID      map[api.RepoID]*types.Repo
	next           int64
	err            error
}

func (r *changesetsConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ChangesetResolver, error) {
	changesets, reposByID, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := make([]graphqlbackend.ChangesetResolver, 0, len(changesets))
	for _, c := range changesets {
		repo, ok := reposByID[c.RepoID]
		if !ok {
			// If it's not in reposByID the repository was either deleted or
			// filtered out by the authz-filter.
			// In both cases: use hiddenChangesetResolver.
			resolvers = append(resolvers, &hiddenChangesetResolver{
				store:      r.store,
				Changeset:  c,
				nextSyncAt: r.scheduledSyncs[c.ID],
			})
			continue
		}

		resolvers = append(resolvers, &changesetResolver{
			store:         r.store,
			Changeset:     c,
			preloadedRepo: repo,
			nextSyncAt:    r.scheduledSyncs[c.ID],
		})
	}

	return resolvers, nil
}

func (r *changesetsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	opts := ee.CountChangesetsOpts{
		CampaignID:          r.opts.CampaignID,
		ExternalState:       r.opts.ExternalState,
		ExternalCheckState:  r.opts.ExternalCheckState,
		ExternalReviewState: r.opts.ExternalReviewState,
	}
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

func (r *changesetsConnectionResolver) compute(ctx context.Context) ([]*campaigns.Changeset, map[api.RepoID]*types.Repo, int64, error) {
	r.once.Do(func() {
		r.changesets, r.next, r.err = r.store.ListChangesets(ctx, r.opts)
		if r.err != nil {
			return
		}

		changesetIDs := make([]int64, len(r.changesets))
		for i, c := range r.changesets {
			changesetIDs[i] = c.ID
		}

		syncData, err := r.store.ListChangesetSyncData(ctx, ee.ListChangesetSyncDataOpts{ChangesetIDs: changesetIDs})
		if err != nil {
			r.err = err
			return
		}
		r.scheduledSyncs = make(map[int64]time.Time)
		for _, d := range syncData {
			r.scheduledSyncs[d.ChangesetID] = ee.NextSync(time.Now, d)
		}

		repoIDs := make([]api.RepoID, len(r.changesets))
		for i, c := range r.changesets {
			repoIDs[i] = c.RepoID
		}

		// 🚨 SECURITY: db.Repos.GetByIDs uses the authzFilter under the hood and
		// filters out repositories that the user doesn't have access to.
		rs, err := db.Repos.GetByIDs(ctx, repoIDs...)
		if err != nil {
			r.err = err
			return
		}

		r.reposByID = make(map[api.RepoID]*types.Repo, len(rs))
		for _, repo := range rs {
			r.reposByID[repo.ID] = repo
		}
	})

	return r.changesets, r.reposByID, r.next, r.err
}

type changesetResolver struct {
	store *ee.Store
	*campaigns.Changeset
	preloadedRepo *types.Repo

	// cache repo because it's called more than once
	repoOnce sync.Once
	repo     *graphqlbackend.RepositoryResolver
	repoErr  error

	// cache changeset events as they are used more than once
	eventsOnce sync.Once
	events     []*campaigns.ChangesetEvent
	eventsErr  error

	// When the next sync is scheduled
	nextSyncAt time.Time
}

const changesetIDKind = "ExternalChangeset"

func marshalChangesetID(id int64) graphql.ID {
	return relay.MarshalID(changesetIDKind, id)
}

func unmarshalChangesetID(id graphql.ID) (cid int64, err error) {
	err = relay.UnmarshalSpec(id, &cid)
	return
}

func (r *changesetResolver) ToExternalChangeset() (graphqlbackend.ExternalChangesetResolver, bool) {
	return r, true
}

func (r *changesetResolver) ToHiddenExternalChangeset() (graphqlbackend.HiddenExternalChangesetResolver, bool) {
	return nil, false
}

func (r *changesetResolver) computeRepo(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	r.repoOnce.Do(func() {
		if r.preloadedRepo != nil {
			r.repo = graphqlbackend.NewRepositoryResolver(r.preloadedRepo)
		} else {
			r.repo, r.repoErr = graphqlbackend.RepositoryByIDInt32(ctx, r.RepoID)
			if r.repoErr != nil {
				return
			}
		}
	})
	return r.repo, r.repoErr
}

func (r *changesetResolver) computeEvents(ctx context.Context) ([]*campaigns.ChangesetEvent, error) {
	r.eventsOnce.Do(func() {
		opts := ee.ListChangesetEventsOpts{
			ChangesetIDs: []int64{r.Changeset.ID},
			Limit:        -1,
		}
		es, _, err := r.store.ListChangesetEvents(ctx, opts)
		r.events = es
		r.eventsErr = err
	})
	return r.events, r.eventsErr
}

func (r *changesetResolver) ID() graphql.ID {
	return marshalChangesetID(r.Changeset.ID)
}

func (r *changesetResolver) ExternalID() string {
	return r.Changeset.ExternalID
}

func (r *changesetResolver) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return r.computeRepo(ctx)
}

func (r *changesetResolver) Campaigns(ctx context.Context, args *graphqlbackend.ListCampaignArgs) (graphqlbackend.CampaignsConnectionResolver, error) {
	return newChangesetCampaignsConnectionsResolver(r.store, r.Changeset.ID, args)
}

func (r *changesetResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.Changeset.CreatedAt}
}

func (r *changesetResolver) UpdatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.Changeset.UpdatedAt}
}

func (r *changesetResolver) NextSyncAt() *graphqlbackend.DateTime {
	if r.nextSyncAt.IsZero() {
		return nil
	}
	return &graphqlbackend.DateTime{Time: r.nextSyncAt}
}

func (r *changesetResolver) Title() (string, error) {
	return r.Changeset.Title()
}

func (r *changesetResolver) Body() (string, error) {
	return r.Changeset.Body()
}

func (r *changesetResolver) State() campaigns.ChangesetState {
	return r.ExternalState
}

func (r *changesetResolver) ExternalURL() (*externallink.Resolver, error) {
	url, err := r.Changeset.URL()
	if err != nil {
		return nil, err
	}
	return externallink.NewResolver(url, r.Changeset.ExternalServiceType), nil
}

func (r *changesetResolver) ReviewState(ctx context.Context) campaigns.ChangesetReviewState {
	return r.ExternalReviewState
}

func (r *changesetResolver) CheckState(ctx context.Context) (*campaigns.ChangesetCheckState, error) {
	state := r.ExternalCheckState
	if state == campaigns.ChangesetCheckStateUnknown {
		return nil, nil
	}
	return &state, nil
}

func (r *changesetResolver) Labels(ctx context.Context) ([]graphqlbackend.ChangesetLabelResolver, error) {
	// Only GitHub supports labels on pull requests so don't make a DB call unless we need to
	if _, ok := r.Changeset.Metadata.(*github.PullRequest); !ok {
		return []graphqlbackend.ChangesetLabelResolver{}, nil
	}
	es, err := r.computeEvents(ctx)
	if err != nil {
		return nil, err
	}
	// We use changeset labels as the source of truth as they can be renamed
	// or removed but we'll also take into account any changeset events that
	// have happened since the last sync in order to reflect changes that
	// have come in via webhooks
	events := ee.ChangesetEvents(es)
	labels := events.UpdateLabelsSince(r.Changeset)
	sort.Slice(labels, func(i, j int) bool {
		return labels[i].Name < labels[j].Name
	})
	resolvers := make([]graphqlbackend.ChangesetLabelResolver, 0, len(labels))
	for _, l := range labels {
		resolvers = append(resolvers, &changesetLabelResolver{label: l})
	}
	return resolvers, nil
}

func (r *changesetResolver) Events(ctx context.Context, args *struct {
	graphqlutil.ConnectionArgs
}) (graphqlbackend.ChangesetEventsConnectionResolver, error) {
	// TODO: We already need to fetch all events for ReviewState and Labels
	// perhaps we can use the cached data here
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
	// Only return diffs for open changesets, otherwise we can't guarantee that
	// we have the refs on gitserver
	if r.ExternalState != campaigns.ChangesetStateOpen {
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
		// Fallback to the ref if we can't get the OID
		base, err = r.Changeset.BaseRef()
		if err != nil {
			return nil, err
		}
	}

	head, err := r.Changeset.HeadRefOid()
	if err != nil {
		return nil, err
	}
	if head == "" {
		// Fallback to the ref if we can't get the OID
		head, err = r.Changeset.HeadRef()
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
	name, err := r.Changeset.HeadRef()
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, errors.New("changeset head ref could not be determined")
	}

	var oid string
	if r.ExternalState == campaigns.ChangesetStateMerged {
		// The PR was merged, find the merge commit
		events, err := r.computeEvents(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "fetching changeset events")
		}
		oid = ee.ChangesetEvents(events).FindMergeCommitID()
	}
	if oid == "" {
		// Fall back to the head ref
		oid, err = r.Changeset.HeadRefOid()
		if err != nil {
			return nil, err
		}
	}

	resolver, err := r.gitRef(ctx, name, oid)
	if err != nil {
		if gitserver.IsRevisionNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return resolver, nil
}

func (r *changesetResolver) Base(ctx context.Context) (*graphqlbackend.GitRefResolver, error) {
	name, err := r.Changeset.BaseRef()
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, errors.New("changeset base ref could not be determined")
	}

	oid, err := r.Changeset.BaseRefOid()
	if err != nil {
		return nil, err
	}

	resolver, err := r.gitRef(ctx, name, oid)
	if err != nil {
		if gitserver.IsRevisionNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return resolver, nil
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
		return "", err
	}
	// Call ResolveRevision to trigger fetches from remote (in case base/head commits don't
	// exist).
	return git.ResolveRevision(ctx, *grepo, nil, refName, nil)
}

type changesetLabelResolver struct {
	label campaigns.ChangesetLabel
}

func (r *changesetLabelResolver) Text() string {
	return r.label.Name
}

func (r *changesetLabelResolver) Color() string {
	return r.label.Color
}

func (r *changesetLabelResolver) Description() *string {
	return &r.label.Description
}

type hiddenChangesetResolver struct {
	store *ee.Store
	*campaigns.Changeset

	// When the next sync is scheduled
	nextSyncAt time.Time
}

const hiddenChangesetIDKind = "HiddenExternalChangeset"

func marshalHiddenChangesetID(id int64) graphql.ID {
	return relay.MarshalID(hiddenChangesetIDKind, id)
}

func unmarshalHiddenChangesetID(id graphql.ID) (cid int64, err error) {
	err = relay.UnmarshalSpec(id, &cid)
	return
}

func (r *hiddenChangesetResolver) ToExternalChangeset() (graphqlbackend.ExternalChangesetResolver, bool) {
	return nil, false
}

func (r *hiddenChangesetResolver) ToHiddenExternalChangeset() (graphqlbackend.HiddenExternalChangesetResolver, bool) {
	return r, true
}

func (r *hiddenChangesetResolver) ID() graphql.ID { return marshalChangesetID(r.Changeset.ID) }

func (r *hiddenChangesetResolver) Campaigns(ctx context.Context, args *graphqlbackend.ListCampaignArgs) (graphqlbackend.CampaignsConnectionResolver, error) {
	return newChangesetCampaignsConnectionsResolver(r.store, r.Changeset.ID, args)
}

func (r *hiddenChangesetResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.Changeset.CreatedAt}
}

func (r *hiddenChangesetResolver) UpdatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.Changeset.UpdatedAt}
}

func (r *hiddenChangesetResolver) NextSyncAt() *graphqlbackend.DateTime {
	if r.nextSyncAt.IsZero() {
		return nil
	}
	return &graphqlbackend.DateTime{Time: r.nextSyncAt}
}

func (r *hiddenChangesetResolver) State() campaigns.ChangesetState {
	return r.Changeset.ExternalState
}

func newChangesetCampaignsConnectionsResolver(
	s *ee.Store,
	changeset int64,
	args *graphqlbackend.ListCampaignArgs,
) (graphqlbackend.CampaignsConnectionResolver, error) {
	opts := ee.ListCampaignsOpts{
		ChangesetID: changeset,
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

	return &campaignsConnectionResolver{store: s, opts: opts}, nil
}
