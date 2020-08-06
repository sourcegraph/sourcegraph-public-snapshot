package resolvers

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type changesetResolver struct {
	store       *ee.Store
	httpFactory *httpcli.Factory

	changeset *campaigns.Changeset

	attemptedPreloadRepo bool
	preloadedRepo        *types.Repo

	// cache repo because it's called more than once
	repoOnce sync.Once
	repo     *graphqlbackend.RepositoryResolver
	repoErr  error
	// The context with which we try to load the repository if it's not
	// preloaded. We need an extra field for that, because the
	// ToExternalChangeset/ToHiddenExternalChangeset methods cannot take a
	// context.Context without graphql-go panic'ing.
	repoCtx context.Context

	// cache changeset events as they are used more than once
	eventsOnce sync.Once
	events     ee.ChangesetEvents
	eventsErr  error

	attemptedPreloadNextSyncAt bool
	// When the next sync is scheduled
	preloadedNextSyncAt *time.Time
	nextSyncAtOnce      sync.Once
	nextSyncAt          time.Time
	nextSyncAtErr       error

	// cache the current ChangesetSpec as it's accessed by multiple methods
	specOnce sync.Once
	spec     *campaigns.ChangesetSpec
	specErr  error
}

const changesetIDKind = "Changeset"

func marshalChangesetID(id int64) graphql.ID {
	return relay.MarshalID(changesetIDKind, id)
}

func unmarshalChangesetID(id graphql.ID) (cid int64, err error) {
	err = relay.UnmarshalSpec(id, &cid)
	return
}

func (r *changesetResolver) ToExternalChangeset() (graphqlbackend.ExternalChangesetResolver, bool) {
	accessible, err := r.repoAccessible()
	if err != nil {
		return r, true
	}

	if !accessible {
		return nil, false
	}

	return r, true
}

func (r *changesetResolver) ToHiddenExternalChangeset() (graphqlbackend.HiddenExternalChangesetResolver, bool) {
	accessible, err := r.repoAccessible()
	if err != nil {
		return r, true
	}

	if accessible {
		return nil, false
	}

	return r, true
}

func (r *changesetResolver) repoAccessible() (bool, error) {
	repo, err := r.computeRepo()
	if err != nil {
		// In case we couldn't load the repository because of an error, we
		// return the error
		return false, err
	}

	// If the repository is not nil, it's accessible
	return repo != nil, nil
}

func (r *changesetResolver) computeRepo() (*graphqlbackend.RepositoryResolver, error) {
	r.repoOnce.Do(func() {
		if r.attemptedPreloadRepo {
			if r.preloadedRepo != nil {
				r.repo = graphqlbackend.NewRepositoryResolver(r.preloadedRepo)
			}
		} else {
			if r.repoCtx == nil {
				r.repoErr = fmt.Errorf("no context available to query repository")
				return
			}

			// 🚨 SECURITY: graphqlbackend.RepositoryByIDInt32 uses the authzFilter under the hood and
			// filters out repositories that the user doesn't have access to.
			r.repo, r.repoErr = graphqlbackend.RepositoryByIDInt32(r.repoCtx, r.changeset.RepoID)
		}
	})
	return r.repo, r.repoErr
}

func (r *changesetResolver) computeSpec(ctx context.Context) (*campaigns.ChangesetSpec, error) {
	r.specOnce.Do(func() {
		if r.changeset.CurrentSpecID == 0 {
			r.specErr = errors.New("Changeset has no ChangesetSpec")
			return
		}

		r.spec, r.specErr = r.store.GetChangesetSpecByID(ctx, r.changeset.CurrentSpecID)
	})
	return r.spec, r.specErr
}

func (r *changesetResolver) computeEvents(ctx context.Context) ([]*campaigns.ChangesetEvent, error) {
	r.eventsOnce.Do(func() {
		opts := ee.ListChangesetEventsOpts{
			ChangesetIDs: []int64{r.changeset.ID},
			Limit:        -1,
		}
		es, _, err := r.store.ListChangesetEvents(ctx, opts)

		r.events = es
		sort.Sort(r.events)

		r.eventsErr = err
	})
	return r.events, r.eventsErr
}

func (r *changesetResolver) computeNextSyncAt(ctx context.Context) (time.Time, error) {
	r.nextSyncAtOnce.Do(func() {
		if r.attemptedPreloadNextSyncAt {
			if r.preloadedNextSyncAt != nil {
				r.nextSyncAt = *r.preloadedNextSyncAt
			}
			return
		}
		syncData, err := r.store.ListChangesetSyncData(ctx, ee.ListChangesetSyncDataOpts{ChangesetIDs: []int64{r.changeset.ID}})
		if err != nil {
			r.nextSyncAtErr = err
			return
		}
		for _, d := range syncData {
			if d.ChangesetID == r.changeset.ID {
				r.nextSyncAt = ee.NextSync(time.Now, d)
				return
			}
		}
	})
	return r.nextSyncAt, r.nextSyncAtErr
}

func (r *changesetResolver) ID() graphql.ID {
	return marshalChangesetID(r.changeset.ID)
}

func (r *changesetResolver) ExternalID() *string {
	if r.changeset.PublicationState.Unpublished() {
		return nil
	}
	return &r.changeset.ExternalID
}

func (r *changesetResolver) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return r.computeRepo()
}

func (r *changesetResolver) Campaigns(ctx context.Context, args *graphqlbackend.ListCampaignArgs) (graphqlbackend.CampaignsConnectionResolver, error) {
	opts := ee.ListCampaignsOpts{
		ChangesetID: r.changeset.ID,
	}

	state, err := parseCampaignState(args.State)
	if err != nil {
		return nil, err
	}
	opts.State = state
	if args.First != nil {
		opts.Limit = int(*args.First)
	}

	authErr := backend.CheckCurrentUserIsSiteAdmin(ctx)
	if authErr != nil && authErr != backend.ErrMustBeSiteAdmin {
		return nil, err
	}
	isSiteAdmin := authErr != backend.ErrMustBeSiteAdmin
	if !isSiteAdmin {
		if args.ViewerCanAdminister != nil && *args.ViewerCanAdminister {
			actor := actor.FromContext(ctx)
			opts.OnlyForAuthor = actor.UID
		}
	}

	return &campaignsConnectionResolver{store: r.store, httpFactory: r.httpFactory, opts: opts}, nil
}

func (r *changesetResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.changeset.CreatedAt}
}

func (r *changesetResolver) UpdatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.changeset.UpdatedAt}
}

func (r *changesetResolver) NextSyncAt(ctx context.Context) (*graphqlbackend.DateTime, error) {
	nextSyncAt, err := r.computeNextSyncAt(ctx)
	if err != nil {
		return nil, err
	}
	if nextSyncAt.IsZero() {
		return nil, nil
	}
	return &graphqlbackend.DateTime{Time: nextSyncAt}, nil
}

func (r *changesetResolver) Title(ctx context.Context) (string, error) {
	if r.changeset.PublicationState.Unpublished() {
		desc, err := r.getBranchSpecDescription(ctx)
		if err != nil {
			return "", err
		}

		return desc.Title, nil
	}

	return r.changeset.Title()
}

func (r *changesetResolver) Body(ctx context.Context) (string, error) {
	if r.changeset.PublicationState.Unpublished() {
		desc, err := r.getBranchSpecDescription(ctx)
		if err != nil {
			return "", err
		}

		return desc.Body, nil
	}

	return r.changeset.Body()
}

func (r *changesetResolver) getBranchSpecDescription(ctx context.Context) (*campaigns.ChangesetSpecDescription, error) {
	spec, err := r.computeSpec(ctx)
	if err != nil {
		return nil, err
	}

	if spec.Spec.IsImportingExisting() {
		return nil, errors.New("ChangesetSpec imports a changeset")
	}

	return spec.Spec, nil
}

func (r *changesetResolver) PublicationState() campaigns.ChangesetPublicationState {
	return r.changeset.PublicationState
}

func (r *changesetResolver) ReconcilerState() campaigns.ReconcilerState {
	return r.changeset.ReconcilerState
}

func (r *changesetResolver) ExternalState() *campaigns.ChangesetExternalState {
	if r.changeset.PublicationState.Unpublished() {
		return nil
	}
	return &r.changeset.ExternalState
}

func (r *changesetResolver) ExternalURL() (*externallink.Resolver, error) {
	if r.changeset.PublicationState.Unpublished() {
		return nil, nil
	}
	url, err := r.changeset.URL()
	if err != nil {
		return nil, err
	}
	return externallink.NewResolver(url, r.changeset.ExternalServiceType), nil
}

func (r *changesetResolver) ReviewState(ctx context.Context) *campaigns.ChangesetReviewState {
	if r.changeset.PublicationState.Unpublished() {
		return nil
	}
	return &r.changeset.ExternalReviewState
}

func (r *changesetResolver) CheckState() *campaigns.ChangesetCheckState {
	if r.changeset.PublicationState.Unpublished() {
		return nil
	}

	state := r.changeset.ExternalCheckState
	if state == campaigns.ChangesetCheckStateUnknown {
		return nil
	}

	return &state
}

func (r *changesetResolver) Error() *string { return r.changeset.FailureMessage }

func (r *changesetResolver) Labels(ctx context.Context) ([]graphqlbackend.ChangesetLabelResolver, error) {
	// Not every code host supports labels on changesets so don't make a DB call unless we need to.
	if ok := r.changeset.SupportsLabels(); !ok {
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
	labels := ee.ComputeLabels(r.changeset, es)
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
		changeset: r.changeset,
		opts: ee.ListChangesetEventsOpts{
			ChangesetIDs: []int64{r.changeset.ID},
			Limit:        int(args.ConnectionArgs.GetFirst()),
		},
	}, nil
}

func (r *changesetResolver) Diff(ctx context.Context) (graphqlbackend.RepositoryComparisonInterface, error) {
	if r.changeset.PublicationState.Unpublished() {
		repo, err := r.computeRepo()
		if err != nil {
			return nil, err
		}

		desc, err := r.getBranchSpecDescription(ctx)
		if err != nil {
			return nil, err
		}

		diff, err := desc.Diff()
		if err != nil {
			return nil, errors.New("ChangesetSpec has no diff")
		}

		return graphqlbackend.NewPreviewRepositoryComparisonResolver(
			ctx,
			repo,
			desc.BaseRev,
			diff,
		)
	}

	// Only return diffs for open changesets, otherwise we can't guarantee that
	// we have the refs on gitserver
	if r.changeset.ExternalState != campaigns.ChangesetExternalStateOpen {
		return nil, nil
	}

	repo, err := r.computeRepo()
	if err != nil {
		return nil, err
	}

	base, err := r.changeset.BaseRefOid()
	if err != nil {
		return nil, err
	}
	if base == "" {
		// Fallback to the ref if we can't get the OID
		base, err = r.changeset.BaseRef()
		if err != nil {
			return nil, err
		}
	}

	head, err := r.changeset.HeadRefOid()
	if err != nil {
		return nil, err
	}
	if head == "" {
		// Fallback to the ref if we can't get the OID
		head, err = r.changeset.HeadRef()
		if err != nil {
			return nil, err
		}
	}

	return graphqlbackend.NewRepositoryComparison(ctx, repo, &graphqlbackend.RepositoryComparisonInput{
		Base:         &base,
		Head:         &head,
		FetchMissing: true,
	})
}

func (r *changesetResolver) DiffStat(ctx context.Context) (*graphqlbackend.DiffStat, error) {
	if stat := r.changeset.DiffStat(); stat != nil {
		return graphqlbackend.NewDiffStat(*stat), nil
	}
	return nil, nil
}

func (r *changesetResolver) Head(ctx context.Context) (*graphqlbackend.GitRefResolver, error) {
	if r.changeset.PublicationState.Unpublished() {
		return nil, nil
	}

	name, err := r.changeset.HeadRef()
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, errors.New("changeset head ref could not be determined")
	}

	var oid string
	if r.changeset.ExternalState == campaigns.ChangesetExternalStateMerged {
		// The PR was merged, find the merge commit
		events, err := r.computeEvents(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "fetching changeset events")
		}
		oid = ee.ChangesetEvents(events).FindMergeCommitID()
	}
	if oid == "" {
		// Fall back to the head ref
		oid, err = r.changeset.HeadRefOid()
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
	if r.changeset.PublicationState.Unpublished() {
		return nil, nil
	}

	name, err := r.changeset.BaseRef()
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, errors.New("changeset base ref could not be determined")
	}

	oid, err := r.changeset.BaseRefOid()
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
	repo, err := r.computeRepo()
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
	return git.ResolveRevision(ctx, *grepo, nil, refName, git.ResolveRevisionOptions{})
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
	if r.label.Description == "" {
		return nil
	}
	return &r.label.Description
}
