package resolvers

import (
	"context"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	sgactor "github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	bgql "github.com/sourcegraph/sourcegraph/internal/batches/graphql"
	"github.com/sourcegraph/sourcegraph/internal/batches/state"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/batches/syncer"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/batches/types/scheduler/config"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type changesetResolver struct {
	store           *store.Store
	gitserverClient gitserver.Client
	logger          log.Logger

	changeset *btypes.Changeset

	// When repo is nil, this resolver resolves to a `HiddenExternalChangeset` in the API.
	repo         *types.Repo
	repoResolver *graphqlbackend.RepositoryResolver

	attemptedPreloadNextSyncAt bool
	// When the next sync is scheduled
	preloadedNextSyncAt time.Time
	nextSyncAtOnce      sync.Once
	nextSyncAt          time.Time
	nextSyncAtErr       error

	// cache the current ChangesetSpec as it's accessed by multiple methods
	specOnce sync.Once
	spec     *btypes.ChangesetSpec
	specErr  error
}

func NewChangesetResolverWithNextSync(store *store.Store, gitserverClient gitserver.Client, logger log.Logger, changeset *btypes.Changeset, repo *types.Repo, nextSyncAt time.Time) *changesetResolver {
	r := NewChangesetResolver(store, gitserverClient, logger, changeset, repo)
	r.attemptedPreloadNextSyncAt = true
	r.preloadedNextSyncAt = nextSyncAt
	return r
}

func NewChangesetResolver(store *store.Store, gitserverClient gitserver.Client, logger log.Logger, changeset *btypes.Changeset, repo *types.Repo) *changesetResolver {
	return &changesetResolver{
		store:           store,
		gitserverClient: gitserverClient,

		repo:         repo,
		repoResolver: graphqlbackend.NewRepositoryResolver(store.DatabaseDB(), gitserverClient, repo),
		changeset:    changeset,
	}
}

const changesetIDKind = "Changeset"

func unmarshalChangesetID(id graphql.ID) (cid int64, err error) {
	err = relay.UnmarshalSpec(id, &cid)
	return
}

func (r *changesetResolver) ToExternalChangeset() (graphqlbackend.ExternalChangesetResolver, bool) {
	if !r.repoAccessible() {
		return nil, false
	}

	return r, true
}

func (r *changesetResolver) ToHiddenExternalChangeset() (graphqlbackend.HiddenExternalChangesetResolver, bool) {
	if r.repoAccessible() {
		return nil, false
	}

	return r, true
}

func (r *changesetResolver) repoAccessible() bool {
	// If the repository is not nil, it's accessible
	return r.repo != nil
}

func (r *changesetResolver) computeSpec(ctx context.Context) (*btypes.ChangesetSpec, error) {
	r.specOnce.Do(func() {
		if r.changeset.CurrentSpecID == 0 {
			r.specErr = errors.New("Changeset has no ChangesetSpec")
			return
		}

		r.spec, r.specErr = r.store.GetChangesetSpecByID(ctx, r.changeset.CurrentSpecID)
	})
	return r.spec, r.specErr
}

func (r *changesetResolver) computeNextSyncAt(ctx context.Context) (time.Time, error) {
	r.nextSyncAtOnce.Do(func() {
		if r.attemptedPreloadNextSyncAt {
			r.nextSyncAt = r.preloadedNextSyncAt
			return
		}
		syncData, err := r.store.ListChangesetSyncData(ctx, store.ListChangesetSyncDataOpts{ChangesetIDs: []int64{r.changeset.ID}})
		if err != nil {
			r.nextSyncAtErr = err
			return
		}
		for _, d := range syncData {
			if d.ChangesetID == r.changeset.ID {
				r.nextSyncAt = syncer.NextSync(r.store.Clock(), d)
				return
			}
		}
	})
	return r.nextSyncAt, r.nextSyncAtErr
}

func (r *changesetResolver) ID() graphql.ID {
	return bgql.MarshalChangesetID(r.changeset.ID)
}

func (r *changesetResolver) ExternalID() *string {
	if r.changeset.ExternalID == "" {
		return nil
	}
	return &r.changeset.ExternalID
}

func (r *changesetResolver) Repository(ctx context.Context) *graphqlbackend.RepositoryResolver {
	return r.repoResolver
}

func (r *changesetResolver) BatchChanges(ctx context.Context, args *graphqlbackend.ListBatchChangesArgs) (graphqlbackend.BatchChangesConnectionResolver, error) {
	opts := store.ListBatchChangesOpts{
		ChangesetID: r.changeset.ID,
	}

	bcState, err := parseBatchChangeState(args.State)
	if err != nil {
		return nil, err
	}
	if bcState != "" {
		opts.States = []btypes.BatchChangeState{bcState}
	}

	// If multiple `states` are provided, prefer them over `bcState`.
	if args.States != nil {
		states, err := parseBatchChangeStates(args.States)
		if err != nil {
			return nil, err
		}
		opts.States = states
	}

	if err := validateFirstParamDefaults(args.First); err != nil {
		return nil, err
	}
	opts.Limit = int(args.First)
	if args.After != nil {
		cursor, err := strconv.ParseInt(*args.After, 10, 32)
		if err != nil {
			return nil, err
		}
		opts.Cursor = cursor
	}

	authErr := auth.CheckCurrentUserIsSiteAdmin(ctx, r.store.DatabaseDB())
	if authErr != nil && authErr != auth.ErrMustBeSiteAdmin {
		return nil, err
	}
	isSiteAdmin := authErr != auth.ErrMustBeSiteAdmin
	if !isSiteAdmin {
		if args.ViewerCanAdminister != nil && *args.ViewerCanAdminister {
			actor := sgactor.FromContext(ctx)
			opts.OnlyAdministeredByUserID = actor.UID
		}
	}

	return &batchChangesConnectionResolver{store: r.store, gitserverClient: r.gitserverClient, opts: opts, logger: r.logger}, nil
}

// This points to the Batch Change that can close or open this changeset on its codehost. If this is nil,
// then the changeset is imported.
func (r *changesetResolver) OwnedByBatchChange() *graphql.ID {
	if batchChangeID := r.changeset.OwnedByBatchChangeID; batchChangeID != 0 {
		bcID := bgql.MarshalBatchChangeID(batchChangeID)
		return &bcID
	}
	return nil
}

func (r *changesetResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.changeset.CreatedAt}
}

func (r *changesetResolver) UpdatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.changeset.UpdatedAt}
}

func (r *changesetResolver) NextSyncAt(ctx context.Context) (*gqlutil.DateTime, error) {
	// If code host syncs are disabled, the syncer is not actively syncing
	// changesets and the next sync time cannot be determined.
	if conf.Get().DisableAutoCodeHostSyncs {
		return nil, nil
	}

	nextSyncAt, err := r.computeNextSyncAt(ctx)
	if err != nil {
		return nil, err
	}
	if nextSyncAt.IsZero() {
		return nil, nil
	}
	return &gqlutil.DateTime{Time: nextSyncAt}, nil
}

func (r *changesetResolver) Title(ctx context.Context) (*string, error) {
	if r.changeset.IsImporting() {
		return nil, nil
	}

	if r.changeset.Published() {
		t, err := r.changeset.Title()
		if err != nil {
			return nil, err
		}
		return &t, nil
	}

	desc, err := r.getBranchSpecDescription(ctx)
	if err != nil {
		return nil, err
	}

	return &desc.Title, nil
}

func (r *changesetResolver) Author() (*graphqlbackend.PersonResolver, error) {
	if r.changeset.IsImporting() {
		return nil, nil
	}

	if !r.changeset.Published() {
		return nil, nil
	}

	name, err := r.changeset.AuthorName()
	if err != nil {
		return nil, err
	}
	email, err := r.changeset.AuthorEmail()
	if err != nil {
		return nil, err
	}

	// For many code hosts, we can't get the author information from the API.
	if name == "" && email == "" {
		return nil, nil
	}

	return graphqlbackend.NewPersonResolver(
		r.store.DatabaseDB(),
		name,
		email,
		// Try to find the corresponding Sourcegraph user.
		true,
	), nil
}

func (r *changesetResolver) Body(ctx context.Context) (*string, error) {
	if r.changeset.IsImporting() {
		return nil, nil
	}

	if r.changeset.Published() {
		b, err := r.changeset.Body()
		if err != nil {
			return nil, err
		}
		return &b, nil
	}

	desc, err := r.getBranchSpecDescription(ctx)
	if err != nil {
		return nil, err
	}

	return &desc.Body, nil
}

func (r *changesetResolver) getBranchSpecDescription(ctx context.Context) (*btypes.ChangesetSpec, error) {
	spec, err := r.computeSpec(ctx)
	if err != nil {
		return nil, err
	}

	if spec.Type == btypes.ChangesetSpecTypeExisting {
		return nil, errors.New("ChangesetSpec imports a changeset")
	}

	return spec, nil
}

func (r *changesetResolver) State() string {
	return string(r.changeset.State)
}

func (r *changesetResolver) ExternalURL() (*externallink.Resolver, error) {
	if !r.changeset.Published() {
		return nil, nil
	}
	if r.changeset.ExternalState == btypes.ChangesetExternalStateDeleted {
		return nil, nil
	}
	url, err := r.changeset.URL()
	if err != nil {
		return nil, err
	}
	if url == "" {
		return nil, nil
	}
	return externallink.NewResolver(url, r.changeset.ExternalServiceType), nil
}

func (r *changesetResolver) ForkNamespace() *string {
	if namespace := r.changeset.ExternalForkNamespace; namespace != "" {
		return &namespace
	}
	return nil
}

func (r *changesetResolver) ForkName() *string {
	if name := r.changeset.ExternalForkName; name != "" {
		return &name
	}
	return nil
}

func (r *changesetResolver) CommitVerification(ctx context.Context) (graphqlbackend.CommitVerificationResolver, error) {
	switch r.changeset.ExternalServiceType {
	case extsvc.TypeGitHub:
		if r.changeset.CommitVerification != nil {
			return &commitVerificationResolver{
				commitVerification: r.changeset.CommitVerification,
			}, nil
		}
	}
	return nil, nil
}

func (r *changesetResolver) ReviewState(ctx context.Context) *string {
	if !r.changeset.Published() {
		return nil
	}
	reviewState := string(r.changeset.ExternalReviewState)
	return &reviewState
}

func (r *changesetResolver) CheckState() *string {
	if !r.changeset.Published() {
		return nil
	}

	checkState := string(r.changeset.ExternalCheckState)
	if checkState == string(btypes.ChangesetCheckStateUnknown) {
		return nil
	}

	return &checkState
}

// Error: `FailureMessage` is set by the reconciler worker if it fails when processing
// a changeset job. However, for most reconciler operations, we automatically retry the
// operation a number of times. When the reconciler worker picks up a failed changeset job
// to restart processing, it clears out the `FailureMessage`, resulting in the error
// disappearing in the UI where we use this resolver field. To retain this context even as
// we retry to process the changeset, we copy over the error to `PreviousFailureMessage`
// when re-enqueueing a changeset and clearing its original `FailureMessage`. Only when a
// changeset is successfully processed will the `PreviousFailureMessage` be cleared.
//
// When resolving this field, we still prefer the latest `FailureMessage` we have, but if
// there's not a `FailureMessage` and there is a `Previous` one, we return that.
func (r *changesetResolver) Error() *string {
	if r.changeset.FailureMessage != nil {
		return r.changeset.FailureMessage
	}
	return r.changeset.PreviousFailureMessage
}

func (r *changesetResolver) SyncerError() *string { return r.changeset.SyncErrorMessage }

func (r *changesetResolver) ScheduleEstimateAt(ctx context.Context) (*gqlutil.DateTime, error) {
	// We need to find out how deep in the queue this changeset is.
	place, err := r.store.GetChangesetPlaceInSchedulerQueue(ctx, r.changeset.ID)
	if err == store.ErrNoResults {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	// Now we can ask the scheduler to estimate where this item would fall in
	// the schedule.
	return gqlutil.DateTimeOrNil(config.ActiveWindow().Estimate(r.store.Clock()(), place)), nil
}

func (r *changesetResolver) CurrentSpec(ctx context.Context) (graphqlbackend.VisibleChangesetSpecResolver, error) {
	if r.changeset.CurrentSpecID == 0 {
		return nil, nil
	}

	spec, err := r.computeSpec(ctx)
	if err != nil {
		return nil, err
	}

	return NewChangesetSpecResolverWithRepo(r.store, r.repo, spec), nil
}

func (r *changesetResolver) Labels(ctx context.Context) ([]graphqlbackend.ChangesetLabelResolver, error) {
	if !r.changeset.Published() {
		return []graphqlbackend.ChangesetLabelResolver{}, nil
	}

	// Not every code host supports labels on changesets so don't make a DB call unless we need to.
	if ok := r.changeset.SupportsLabels(); !ok {
		return []graphqlbackend.ChangesetLabelResolver{}, nil
	}

	opts := store.ListChangesetEventsOpts{
		ChangesetIDs: []int64{r.changeset.ID},
		Kinds:        state.ComputeLabelsRequiredEventTypes,
	}
	es, _, err := r.store.ListChangesetEvents(ctx, opts)
	if err != nil {
		return nil, err
	}
	// ComputeLabels expects the events to be pre-sorted.
	sort.Sort(state.ChangesetEvents(es))

	// We use changeset labels as the source of truth as they can be renamed
	// or removed but we'll also take into account any changeset events that
	// have happened since the last sync in order to reflect changes that
	// have come in via webhooks
	labels := state.ComputeLabels(r.changeset, es)
	resolvers := make([]graphqlbackend.ChangesetLabelResolver, 0, len(labels))
	for _, l := range labels {
		resolvers = append(resolvers, &changesetLabelResolver{label: l})
	}
	return resolvers, nil
}

func (r *changesetResolver) Events(ctx context.Context, args *graphqlbackend.ChangesetEventsConnectionArgs) (graphqlbackend.ChangesetEventsConnectionResolver, error) {
	if err := validateFirstParamDefaults(args.First); err != nil {
		return nil, err
	}
	var cursor int64
	if args.After != nil {
		var err error
		cursor, err = strconv.ParseInt(*args.After, 10, 32)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse after cursor")
		}
	}
	// TODO: We already need to fetch all events for ReviewState and Labels
	// perhaps we can use the cached data here
	return &changesetEventsConnectionResolver{
		store:             r.store,
		changesetResolver: r,
		first:             int(args.First),
		cursor:            cursor,
	}, nil
}

func (r *changesetResolver) Diff(ctx context.Context) (graphqlbackend.RepositoryComparisonInterface, error) {
	if r.changeset.IsImporting() {
		return nil, nil
	}

	db := r.store.DatabaseDB()
	// If the Changeset is from a code host that doesn't push to branches (like Gerrit), we can just use the branch spec description.
	if r.changeset.Unpublished() || r.changeset.SyncState.BaseRefOid == r.changeset.SyncState.HeadRefOid {
		desc, err := r.getBranchSpecDescription(ctx)
		if err != nil {
			return nil, err
		}

		return graphqlbackend.NewPreviewRepositoryComparisonResolver(
			ctx,
			db,
			r.gitserverClient,
			r.repoResolver,
			desc.BaseRev,
			desc.Diff,
		)
	}

	if !r.changeset.HasDiff() {
		return nil, nil
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

	return graphqlbackend.NewRepositoryComparison(ctx, db, r.gitserverClient, r.repoResolver, &graphqlbackend.RepositoryComparisonInput{
		Base: &base,
		Head: &head,
	})
}

func (r *changesetResolver) DiffStat(ctx context.Context) (*graphqlbackend.DiffStat, error) {
	if stat := r.changeset.DiffStat(); stat != nil {
		return graphqlbackend.NewDiffStat(*stat), nil
	}
	return nil, nil
}

type changesetLabelResolver struct {
	label btypes.ChangesetLabel
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

var _ graphqlbackend.CommitVerificationResolver = &commitVerificationResolver{}

type commitVerificationResolver struct {
	commitVerification *github.Verification
}

func (c *commitVerificationResolver) ToGitHubCommitVerification() (graphqlbackend.GitHubCommitVerificationResolver, bool) {
	if c.commitVerification != nil {
		return &gitHubCommitVerificationResolver{commitVerification: c.commitVerification}, true
	}

	return nil, false
}

var _ graphqlbackend.GitHubCommitVerificationResolver = &gitHubCommitVerificationResolver{}

type gitHubCommitVerificationResolver struct {
	commitVerification *github.Verification
}

func (r *gitHubCommitVerificationResolver) Verified() bool {
	return r.commitVerification.Verified
}

func (r *gitHubCommitVerificationResolver) Reason() string {
	return r.commitVerification.Reason
}

func (r *gitHubCommitVerificationResolver) Signature() string {
	return r.commitVerification.Signature
}

func (r *gitHubCommitVerificationResolver) Payload() string {
	return r.commitVerification.Payload
}
