package campaigns

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/schema"
)

type GitserverClient interface {
	CreateCommitFromPatch(ctx context.Context, req protocol.CreateCommitFromPatchRequest) (string, error)
}

// Reconciler processes changesets and reconciles their current state — in
// Sourcegraph or on the code host — with that described in the current
// ChangesetSpec associated with the changeset.
type Reconciler struct {
	GitserverClient GitserverClient
	Sourcer         repos.Sourcer
	Store           *Store

	// This is used to disable a time.Sleep for operationSleep so that the
	// tests don't run slower.
	noSleepBeforeSync bool
}

// HandlerFunc returns a dbworker.HandlerFunc that can be passed to a
// workerutil.Worker to process queued changesets.
func (r *Reconciler) HandlerFunc() dbworker.HandlerFunc {
	return func(ctx context.Context, tx dbworkerstore.Store, record workerutil.Record) error {
		return r.process(ctx, r.Store.With(tx), record.(*campaigns.Changeset))
	}
}

// process is the main entry point of the reconciler and processes changesets
// that were marked as queued in the database.
//
// For each changeset, the reconciler computes an execution plan to run to reconcile a
// possible divergence between the changeset's current state and the desired
// state (for example expressed in a changeset spec).
//
// To do that, the reconciler looks at the changeset's current state
// (publication state, external state, sync state, ...), its (if set) current
// ChangesetSpec, and (if it exists) its previous ChangesetSpec.
//
// If an error is returned, the workerutil.Worker that called this function
// (through the HandlerFunc) will set the changeset's ReconcilerState to
// errored and set its FailureMessage to the error.
func (r *Reconciler) process(ctx context.Context, tx *Store, ch *campaigns.Changeset) error {
	// Reset the error message.
	ch.FailureMessage = nil

	prev, curr, err := loadChangesetSpecs(ctx, tx, ch)
	if err != nil {
		return nil
	}

	plan, err := DetermineReconcilerPlan(prev, curr, ch)
	if err != nil {
		return err
	}

	log15.Info("Reconciler processing changeset", "changeset", ch.ID, "operations", plan.Ops)

	e := &executor{
		sourcer:           r.Sourcer,
		gitserverClient:   r.GitserverClient,
		noSleepBeforeSync: r.noSleepBeforeSync,

		tx: tx,
		ch: ch,

		spec:  curr,
		delta: plan.Delta,
	}

	return e.ExecutePlan(ctx, plan)
}

// ErrPublishSameBranch is returned by publish changeset if a changeset with
// the same external branch already exists in the database and is owned by
// another campaign.
// It is a terminal error that won't be fixed by retrying to publish the
// changeset with the same spec.
type ErrPublishSameBranch struct{}

func (e ErrPublishSameBranch) Error() string {
	return "cannot create changeset on the same branch in multiple campaigns"
}

func (e ErrPublishSameBranch) NonRetryable() bool { return true }

type executor struct {
	gitserverClient   GitserverClient
	sourcer           repos.Sourcer
	noSleepBeforeSync bool

	tx  *Store
	ccs repos.ChangesetSource

	repo   *repos.Repo
	extSvc *types.ExternalService

	// au is nil if we want to use the global credentials stored in the external
	// service configuration.
	au auth.Authenticator

	ch    *campaigns.Changeset
	spec  *campaigns.ChangesetSpec
	delta *ChangesetSpecDelta
}

// ExecutePlan executes the given reconciler plan.
func (e *executor) ExecutePlan(ctx context.Context, plan *ReconcilerPlan) (err error) {
	if plan.Ops.IsNone() {
		return nil
	}

	reposStore := repos.NewDBStore(e.tx.Handle().DB(), sql.TxOptions{})

	e.repo, err = loadRepo(ctx, reposStore, e.ch.RepoID)
	if err != nil {
		return errors.Wrap(err, "failed to load repository")
	}

	e.extSvc, err = loadExternalService(ctx, reposStore, e.repo)
	if err != nil {
		return errors.Wrap(err, "failed to load external service")
	}

	// Figure out which authenticator we should use to modify the changeset.
	e.au, err = e.loadAuthenticator(ctx)
	if err != nil {
		return err
	}

	// Set up a source with which we can modify the changeset.
	e.ccs, err = e.buildChangesetSource(e.repo, e.extSvc)
	if err != nil {
		return err
	}

	upsertChangesetEvents := true
	for _, op := range plan.Ops.ExecutionOrder() {
		switch op {
		case campaigns.ReconcilerOperationSync:
			err = e.syncChangeset(ctx)

		case campaigns.ReconcilerOperationImport:
			err = e.importChangeset(ctx)

		case campaigns.ReconcilerOperationPush:
			err = e.pushChangesetPatch(ctx)

		case campaigns.ReconcilerOperationPublish:
			err = e.publishChangeset(ctx, false)

		case campaigns.ReconcilerOperationPublishDraft:
			err = e.publishChangeset(ctx, true)

		case campaigns.ReconcilerOperationReopen:
			err = e.reopenChangeset(ctx)

		case campaigns.ReconcilerOperationUpdate:
			err = e.updateChangeset(ctx)

		case campaigns.ReconcilerOperationUndraft:
			err = e.undraftChangeset(ctx)

		case campaigns.ReconcilerOperationClose:
			err = e.closeChangeset(ctx)

		case campaigns.ReconcilerOperationSleep:
			e.sleep()

		default:
			err = fmt.Errorf("executor operation %q not implemented", op)
		}

		if err != nil {
			return err
		}
	}

	if upsertChangesetEvents {
		events := e.ch.Events()
		SetDerivedState(ctx, e.ch, events)

		if err := e.tx.UpsertChangesetEvents(ctx, events...); err != nil {
			log15.Error("UpsertChangesetEvents", "err", err)
			return err
		}
	}

	return e.tx.UpdateChangeset(ctx, e.ch)
}

func (e *executor) buildChangesetSource(repo *repos.Repo, extSvc *types.ExternalService) (repos.ChangesetSource, error) {
	sources, err := e.sourcer(extSvc)
	if err != nil {
		return nil, err
	}
	if len(sources) != 1 {
		return nil, errors.New("invalid number of sources for external service")
	}
	src := sources[0]

	if e.au != nil {
		// If e.au == nil that means the user that applied that last
		// campaign/changeset spec is a site-admin and we can fall back to the
		// global credentials stored in extSvc.
		ucs, ok := src.(repos.UserSource)
		if !ok {
			return nil, errors.Errorf("using user credentials on code host of repo %q is not implemented", repo.Name)
		}

		if src, err = ucs.WithAuthenticator(e.au); err != nil {
			return nil, errors.Wrapf(err, "unable to use this specific user credential on code host of repo %q", repo.Name)
		}
	}

	ccs, ok := src.(repos.ChangesetSource)
	if !ok {
		return nil, errors.Errorf("creating changesets on code host of repo %q is not implemented", repo.Name)
	}

	return ccs, nil
}

// loadAuthenticator determines the correct Authenticator to use when
// reconciling the current changeset. It will return nil, nil if the code host's
// global configuration should be used (ie the applying user is an admin and
// doesn't have a credential configured for the code host, or the changeset
// isn't owned by a campaign).
func (e *executor) loadAuthenticator(ctx context.Context) (auth.Authenticator, error) {
	if e.ch.OwnedByCampaignID == 0 {
		// Unowned changesets are imported, and therefore don't need to use a user
		// credential, since reconciliation isn't a mutating process.
		return nil, nil
	}

	// If the changeset is owned by a campaign, we want to reconcile using
	// the user's credentials, which means we need to know which user last
	// applied the owning campaign. Let's go find out.
	campaign, err := loadCampaign(ctx, e.tx, e.ch.OwnedByCampaignID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load owning campaign")
	}

	cred, err := loadUserCredential(ctx, campaign.LastApplierID, e.repo)
	if err != nil {
		if errcode.IsNotFound(err) {
			// We need to check if the user is an admin: if they are, then
			// we can use the nil return from loadUserCredential() to fall
			// back to the global credentials used for the code host. If
			// not, then we need to error out.
			user, err := loadUser(ctx, campaign.LastApplierID)
			if err != nil {
				return nil, errors.Wrap(err, "failed to load user applying the campaign")
			}

			if user.SiteAdmin {
				return nil, nil
			}

			return nil, ErrMissingCredentials{repo: e.repo.Name}
		}
		return nil, errors.Wrap(err, "failed to load user credential")
	}

	return cred.Credential, nil
}

// ErrMissingCredentials is returned by loadAuthenticator if the user that
// applied the last campaign/changeset spec doesn't have UserCredentials for
// the given repository and is not a site-admin (so no fallback to the global
// credentials is possible).
type ErrMissingCredentials struct{ repo string }

func (e ErrMissingCredentials) Error() string {
	return fmt.Sprintf("user does not have a valid credential for repository %q", e.repo)
}

func (e ErrMissingCredentials) NonRetryable() bool { return true }

// pushChangesetPatch creates the commits for the changeset on its codehost.
func (e *executor) pushChangesetPatch(ctx context.Context) (err error) {
	existingSameBranch, err := e.tx.GetChangeset(ctx, GetChangesetOpts{
		ExternalServiceType: e.ch.ExternalServiceType,
		RepoID:              e.ch.RepoID,
		ExternalBranch:      e.spec.Spec.HeadRef,
	})
	if err != nil && err != ErrNoResults {
		return err
	}

	if existingSameBranch != nil && existingSameBranch.ID != e.ch.ID {
		return ErrPublishSameBranch{}
	}

	// Create a commit and push it
	opts, err := buildCommitOpts(e.repo, e.extSvc, e.spec, e.au)
	if err != nil {
		return err
	}
	return e.pushCommit(ctx, opts)
}

// publishChangeset creates the given changeset on its code host.
func (e *executor) publishChangeset(ctx context.Context, asDraft bool) (err error) {
	cs := &repos.Changeset{
		Title:     e.spec.Spec.Title,
		Body:      e.spec.Spec.Body,
		BaseRef:   e.spec.Spec.BaseRef,
		HeadRef:   e.spec.Spec.HeadRef,
		Repo:      e.repo,
		Changeset: e.ch,
	}

	// Depending on the changeset, we may want to add to the body (for example,
	// to add a backlink to Sourcegraph).
	if err := decorateChangesetBody(ctx, e.tx, cs); err != nil {
		return errors.Wrapf(err, "decorating body for changeset %d", e.ch.ID)
	}

	var exists bool
	if asDraft {
		// If the changeset shall be published in draft mode, make sure the changeset source implements DraftChangesetSource.
		draftCcs, ok := e.ccs.(repos.DraftChangesetSource)
		if !ok {
			return errors.New("changeset operation is publish-draft, but changeset source doesn't implement DraftChangesetSource")
		}
		exists, err = draftCcs.CreateDraftChangeset(ctx, cs)
	} else {
		// If we're running this method a second time, because we failed due to an
		// ephemeral error, there's a race condition here.
		// It's possible that `CreateChangeset` doesn't return the newest head ref
		// commit yet, because the API of the codehost doesn't return it yet.
		exists, err = e.ccs.CreateChangeset(ctx, cs)
	}
	if err != nil {
		return errors.Wrap(err, "creating changeset")
	}
	// If the Changeset already exists and our source can update it, we try to update it
	if exists {
		outdated, err := cs.IsOutdated()
		if err != nil {
			return errors.Wrap(err, "could not determine whether changeset needs update")
		}

		if outdated {
			if err := e.ccs.UpdateChangeset(ctx, cs); err != nil {
				return errors.Wrap(err, "updating changeset")
			}
		}
	}
	// Set the changeset to published.
	e.ch.PublicationState = campaigns.ChangesetPublicationStatePublished
	return nil
}

func (e *executor) syncChangeset(ctx context.Context) error {
	if err := e.loadChangeset(ctx); err != nil {
		_, ok := err.(repos.ChangesetNotFoundError)
		if !ok {
			return err
		}

		// If we're syncing a changeset and it can't be found anymore, we mark
		// it as deleted.
		if !e.ch.IsDeleted() {
			e.ch.SetDeleted()
		}
	}

	return nil
}

func (e *executor) importChangeset(ctx context.Context) error {
	if err := e.loadChangeset(ctx); err != nil {
		return err
	}

	e.ch.Unsynced = false

	return nil
}

func (e *executor) loadChangeset(ctx context.Context) error {
	repoChangeset := &repos.Changeset{Repo: e.repo, Changeset: e.ch}
	return e.ccs.LoadChangeset(ctx, repoChangeset)
}

// updateChangeset updates the given changeset's attribute on the code host
// according to its ChangesetSpec and the delta previously computed.
func (e *executor) updateChangeset(ctx context.Context) (err error) {
	cs := repos.Changeset{
		Title:     e.spec.Spec.Title,
		Body:      e.spec.Spec.Body,
		BaseRef:   e.spec.Spec.BaseRef,
		HeadRef:   e.spec.Spec.HeadRef,
		Repo:      e.repo,
		Changeset: e.ch,
	}

	// Depending on the changeset, we may want to add to the body (for example,
	// to add a backlink to Sourcegraph).
	if err := decorateChangesetBody(ctx, e.tx, &cs); err != nil {
		return errors.Wrapf(err, "decorating body for changeset %d", e.ch.ID)
	}

	if err := e.ccs.UpdateChangeset(ctx, &cs); err != nil {
		return errors.Wrap(err, "updating changeset")
	}

	return nil
}

// reopenChangeset reopens the given changeset attribute on the code host.
func (e *executor) reopenChangeset(ctx context.Context) (err error) {
	cs := repos.Changeset{Repo: e.repo, Changeset: e.ch}
	if err := e.ccs.ReopenChangeset(ctx, &cs); err != nil {
		return errors.Wrap(err, "updating changeset")
	}
	return nil
}

// closeChangeset closes the given changeset on its code host if its ExternalState is OPEN or DRAFT.
func (e *executor) closeChangeset(ctx context.Context) (err error) {
	e.ch.Closing = false

	if e.ch.ExternalState != campaigns.ChangesetExternalStateDraft && e.ch.ExternalState != campaigns.ChangesetExternalStateOpen {
		return nil
	}

	cs := &repos.Changeset{Changeset: e.ch, Repo: e.repo}

	if err := e.ccs.CloseChangeset(ctx, cs); err != nil {
		return errors.Wrap(err, "closing changeset")
	}
	return nil
}

// undraftChangeset marks the given changeset on its code host as ready for review.
func (e *executor) undraftChangeset(ctx context.Context) (err error) {
	draftCcs, ok := e.ccs.(repos.DraftChangesetSource)
	if !ok {
		return errors.New("changeset operation is undraft, but changeset source doesn't implement DraftChangesetSource")
	}

	cs := &repos.Changeset{
		Title:     e.spec.Spec.Title,
		Body:      e.spec.Spec.Body,
		BaseRef:   e.spec.Spec.BaseRef,
		HeadRef:   e.spec.Spec.HeadRef,
		Repo:      e.repo,
		Changeset: e.ch,
	}

	if err := draftCcs.UndraftChangeset(ctx, cs); err != nil {
		return errors.Wrap(err, "undrafting changeset")
	}
	return nil
}

// sleep sleeps for 3 seconds.
func (e *executor) sleep() {
	if !e.noSleepBeforeSync {
		time.Sleep(3 * time.Second)
	}
}

func (e *executor) pushCommit(ctx context.Context, opts protocol.CreateCommitFromPatchRequest) error {
	_, err := e.gitserverClient.CreateCommitFromPatch(ctx, opts)
	if err != nil {
		if diffErr, ok := err.(*protocol.CreateCommitFromPatchError); ok {
			return errors.Errorf(
				"creating commit from patch for repository %q: %s\n"+
					"```\n"+
					"$ %s\n"+
					"%s\n"+
					"```",
				diffErr.RepositoryName, diffErr.InternalError, diffErr.Command, strings.TrimSpace(diffErr.CombinedOutput))
		}
		return err
	}

	return nil
}

func buildCommitOpts(repo *repos.Repo, extSvc *types.ExternalService, spec *campaigns.ChangesetSpec, a auth.Authenticator) (protocol.CreateCommitFromPatchRequest, error) {
	var opts protocol.CreateCommitFromPatchRequest

	desc := spec.Spec

	diff, err := desc.Diff()
	if err != nil {
		return opts, err
	}

	commitMessage, err := desc.CommitMessage()
	if err != nil {
		return opts, err
	}

	commitAuthorName, err := desc.AuthorName()
	if err != nil {
		return opts, err
	}

	commitAuthorEmail, err := desc.AuthorEmail()
	if err != nil {
		return opts, err
	}

	source, ok := repo.Sources[extSvc.URN()]
	if !ok {
		return opts, errors.New("repository was not cloned through given external service")
	}

	pushConf, err := buildPushConfig(repo.ExternalRepo.ServiceType, source.CloneURL, a)
	if err != nil {
		return opts, err
	}

	opts = protocol.CreateCommitFromPatchRequest{
		Repo:       api.RepoName(repo.Name),
		BaseCommit: api.CommitID(desc.BaseRev),
		// IMPORTANT: We add a trailing newline here, otherwise `git apply`
		// will fail with "corrupt patch at line <N>" where N is the last line.
		Patch:     diff + "\n",
		TargetRef: desc.HeadRef,

		// CAUTION: `UniqueRef` means that we'll push to the branch even if it
		// already exists.
		// So when we retry publishing a changeset, this will overwrite what we
		// pushed before.
		UniqueRef: false,

		CommitInfo: protocol.PatchCommitInfo{
			Message:     commitMessage,
			AuthorName:  commitAuthorName,
			AuthorEmail: commitAuthorEmail,
			Date:        spec.CreatedAt,
		},
		// We use unified diffs, not git diffs, which means they're missing the
		// `a/` and `b/` filename prefixes. `-p0` tells `git apply` to not
		// expect and strip prefixes.
		GitApplyArgs: []string{"-p0"},
		Push:         pushConf,
	}

	return opts, nil
}

func buildPushConfig(extSvcType, cloneURL string, a auth.Authenticator) (*protocol.PushConfig, error) {
	u, err := url.Parse(cloneURL)
	if err != nil {
		return nil, errors.Wrap(err, "parsing repository clone URL")
	}

	switch av := a.(type) {
	case *auth.OAuthBearerToken:
		switch extSvcType {
		case extsvc.TypeGitHub:
			u.User = url.User(av.Token)

		case extsvc.TypeGitLab:
			u.User = url.UserPassword("git", av.Token)

		case extsvc.TypeBitbucketServer:
			return nil, errors.New("require username/token to push commits to BitbucketServer")
		}

	case *auth.BasicAuth:
		switch extSvcType {
		case extsvc.TypeGitHub, extsvc.TypeGitLab:
			return nil, errors.New("need token to push commits to " + extSvcType)

		case extsvc.TypeBitbucketServer:
			u.User = url.UserPassword(av.Username, av.Password)
		}

	case nil:
		// This is OK: we'll just send an empty token and gitserver will use
		// the credential stored in the clone URL of the repository.

	default:
		return nil, ErrNoPushCredentials{credentialsType: fmt.Sprintf("%T", a)}
	}

	return &protocol.PushConfig{RemoteURL: u.String()}, nil
}

// ErrNoPushCredentials is returned by buildCommitOpts if the credentials
// cannot be used by git to authenticate a `git push`.
type ErrNoPushCredentials struct{ credentialsType string }

func (e ErrNoPushCredentials) Error() string {
	return fmt.Sprintf("cannot use credentials of type %T to push commits", e.credentialsType)
}

func (e ErrNoPushCredentials) NonRetryable() bool { return true }

var operationPrecedence = map[campaigns.ReconcilerOperation]int{
	campaigns.ReconcilerOperationPush:         0,
	campaigns.ReconcilerOperationImport:       1,
	campaigns.ReconcilerOperationPublish:      1,
	campaigns.ReconcilerOperationPublishDraft: 1,
	campaigns.ReconcilerOperationClose:        1,
	campaigns.ReconcilerOperationReopen:       2,
	campaigns.ReconcilerOperationUndraft:      3,
	campaigns.ReconcilerOperationUpdate:       4,
	campaigns.ReconcilerOperationSleep:        5,
	campaigns.ReconcilerOperationSync:         6,
}

type ReconcilerOperations []campaigns.ReconcilerOperation

func (ops ReconcilerOperations) IsNone() bool {
	return len(ops) == 0
}

func (ops ReconcilerOperations) Equal(b ReconcilerOperations) bool {
	if len(ops) != len(b) {
		return false
	}
	bEntries := make(map[campaigns.ReconcilerOperation]struct{})
	for _, e := range b {
		bEntries[e] = struct{}{}
	}

	for _, op := range ops {
		if _, ok := bEntries[op]; !ok {
			return false
		}
	}

	return true
}

func (ops ReconcilerOperations) String() string {
	if ops.IsNone() {
		return "No operations required"
	}
	eo := ops.ExecutionOrder()
	ss := make([]string, len(eo))
	for i, val := range eo {
		ss[i] = strings.ToLower(string(val))
	}
	return strings.Join(ss, " => ")
}

func (ops ReconcilerOperations) ExecutionOrder() []campaigns.ReconcilerOperation {
	uniqueOps := []campaigns.ReconcilerOperation{}

	// Make sure ops are unique.
	seenOps := make(map[campaigns.ReconcilerOperation]struct{})
	for _, op := range ops {
		if _, ok := seenOps[op]; ok {
			continue
		}

		seenOps[op] = struct{}{}
		uniqueOps = append(uniqueOps, op)
	}

	sort.Slice(uniqueOps, func(i, j int) bool {
		return operationPrecedence[uniqueOps[i]] < operationPrecedence[uniqueOps[j]]
	})

	return uniqueOps
}

// ReconcilerPlan represents the possible operations the reconciler needs to do
// to reconcile the current and the desired state of a changeset.
type ReconcilerPlan struct {
	// The operations that need to be done to reconcile the changeset.
	Ops ReconcilerOperations

	// The Delta between a possible previous ChangesetSpec and the current
	// ChangesetSpec.
	Delta *ChangesetSpecDelta
}

func (p *ReconcilerPlan) AddOp(op campaigns.ReconcilerOperation) { p.Ops = append(p.Ops, op) }
func (p *ReconcilerPlan) SetOp(op campaigns.ReconcilerOperation) { p.Ops = ReconcilerOperations{op} }

// DetermineReconcilerPlan looks at the given changeset to determine what action the
// reconciler should take.
// It loads the current ChangesetSpec and if it exists also the previous one.
// If the current ChangesetSpec is not applied to a campaign, it returns an
// error.
func DetermineReconcilerPlan(previousSpec, currentSpec *campaigns.ChangesetSpec, ch *campaigns.Changeset) (*ReconcilerPlan, error) {
	pl := &ReconcilerPlan{}

	// If it doesn't have a spec, it's an imported changeset and we can't do
	// anything.
	if currentSpec == nil {
		if ch.Unsynced {
			pl.SetOp(campaigns.ReconcilerOperationImport)
		}
		return pl, nil
	}

	// If it's marked as closing, we don't need to look at the specs.
	if ch.Closing {
		pl.SetOp(campaigns.ReconcilerOperationClose)
		return pl, nil
	}

	delta, err := compareChangesetSpecs(previousSpec, currentSpec)
	if err != nil {
		return pl, nil
	}
	pl.Delta = delta

	switch ch.PublicationState {
	case campaigns.ChangesetPublicationStateUnpublished:
		if currentSpec.Spec.Published.True() {
			pl.SetOp(campaigns.ReconcilerOperationPublish)
			pl.AddOp(campaigns.ReconcilerOperationPush)
		} else if currentSpec.Spec.Published.Draft() && ch.SupportsDraft() {
			// If configured to be opened as draft, and the changeset supports
			// draft mode, publish as draft. Otherwise, take no action.
			pl.SetOp(campaigns.ReconcilerOperationPublishDraft)
			pl.AddOp(campaigns.ReconcilerOperationPush)
		}

	case campaigns.ChangesetPublicationStatePublished:
		// Don't take any actions for merged changesets.
		if ch.ExternalState == campaigns.ChangesetExternalStateMerged {
			return pl, nil
		}
		if reopenAfterDetach(ch) {
			pl.SetOp(campaigns.ReconcilerOperationReopen)
		}

		// Only do undraft, when the codehost supports draft changesets.
		if delta.Undraft && campaigns.ExternalServiceSupports(ch.ExternalServiceType, campaigns.CodehostCapabilityDraftChangesets) {
			pl.AddOp(campaigns.ReconcilerOperationUndraft)
		}

		if delta.AttributesChanged() {
			if delta.NeedCommitUpdate() {
				pl.AddOp(campaigns.ReconcilerOperationPush)
			}

			// If we only need to update the diff and we didn't change the state of the changeset,
			// we're done, because we already pushed the commit. We don't need to
			// update anything on the codehost.
			if !delta.NeedCodeHostUpdate() {
				// But we need to sync the changeset so that it has the new commit.
				//
				// The problem: the code host might not have updated the changeset to
				// have the new commit SHA as its head ref oid (and the check states,
				// ...).
				//
				// That's why we give them 3 seconds to update the changesets.
				//
				// Why 3 seconds? Well... 1 or 2 seem to be too short and 4 too long?
				pl.AddOp(campaigns.ReconcilerOperationSleep)
				pl.AddOp(campaigns.ReconcilerOperationSync)
			} else {
				// Otherwise, we need to update the pull request on the code host or, if we
				// need to reopen it, update it to make sure it has the newest state.
				pl.AddOp(campaigns.ReconcilerOperationUpdate)
			}
		}

	default:
		return pl, fmt.Errorf("unknown changeset publication state: %s", ch.PublicationState)
	}

	return pl, nil
}

func reopenAfterDetach(ch *campaigns.Changeset) bool {
	closed := ch.ExternalState == campaigns.ChangesetExternalStateClosed
	if !closed {
		return false
	}

	// Sanity check: if it's not owned by a campaign, it's simply being tracked.
	if ch.OwnedByCampaignID == 0 {
		return false
	}
	// Sanity check 2: if it's marked as to-be-closed, then we don't reopen it.
	if ch.Closing {
		return false
	}

	// At this point the changeset is closed and not marked as to-be-closed.

	// TODO: What if somebody closed the changeset on purpose on the codehost?
	return ch.AttachedTo(ch.OwnedByCampaignID)
}

func loadRepo(ctx context.Context, tx RepoStore, id api.RepoID) (*repos.Repo, error) {
	rs, err := tx.ListRepos(ctx, repos.StoreListReposArgs{IDs: []api.RepoID{id}})
	if err != nil {
		return nil, err
	}
	if len(rs) != 1 {
		return nil, errors.Errorf("repo not found: %d", id)
	}
	return rs[0], nil
}

func loadExternalService(ctx context.Context, reposStore RepoStore, repo *repos.Repo) (*types.ExternalService, error) {
	var externalService *types.ExternalService
	args := repos.StoreListExternalServicesArgs{IDs: repo.ExternalServiceIDs()}

	es, err := reposStore.ListExternalServices(ctx, args)
	if err != nil {
		return nil, err
	}

	for _, e := range es {
		cfg, err := e.Configuration()
		if err != nil {
			return nil, err
		}

		switch cfg := cfg.(type) {
		case *schema.GitHubConnection:
			if cfg.Token != "" {
				externalService = e
			}
		case *schema.BitbucketServerConnection:
			if cfg.Token != "" {
				externalService = e
			}
		case *schema.GitLabConnection:
			if cfg.Token != "" {
				externalService = e
			}
		}
		if externalService != nil {
			break
		}
	}

	if externalService == nil {
		return nil, errors.Errorf("no external services found for repo %q", repo.Name)
	}

	return externalService, nil
}

func loadCampaign(ctx context.Context, tx *Store, id int64) (*campaigns.Campaign, error) {
	if id == 0 {
		return nil, errors.New("changeset has no owning campaign")
	}

	campaign, err := tx.GetCampaign(ctx, GetCampaignOpts{ID: id})
	if err != nil && err != ErrNoResults {
		return nil, errors.Wrapf(err, "retrieving owning campaign: %d", id)
	} else if campaign == nil {
		return nil, errors.Errorf("campaign not found: %d", id)
	}

	return campaign, nil
}

func loadChangesetSpecs(ctx context.Context, tx *Store, ch *campaigns.Changeset) (prev, curr *campaigns.ChangesetSpec, err error) {
	if ch.CurrentSpecID != 0 {
		curr, err = tx.GetChangesetSpecByID(ctx, ch.CurrentSpecID)
		if err != nil {
			return
		}
	}
	if ch.PreviousSpecID != 0 {
		prev, err = tx.GetChangesetSpecByID(ctx, ch.PreviousSpecID)
		if err != nil {
			return
		}
	}
	return
}

func loadUser(ctx context.Context, id int32) (*types.User, error) {
	return db.Users.GetByID(ctx, id)
}

func loadUserCredential(ctx context.Context, userID int32, repo *repos.Repo) (*db.UserCredential, error) {
	return db.UserCredentials.GetByScope(ctx, db.UserCredentialScope{
		Domain:              db.UserCredentialDomainCampaigns,
		UserID:              userID,
		ExternalServiceType: repo.ExternalRepo.ServiceType,
		ExternalServiceID:   repo.ExternalRepo.ServiceID,
	})
}

func decorateChangesetBody(ctx context.Context, tx *Store, cs *repos.Changeset) error {
	campaign, err := loadCampaign(ctx, tx, cs.OwnedByCampaignID)
	if err != nil {
		return errors.Wrap(err, "failed to load campaign")
	}

	// We need to get the namespace, since external campaign URLs are
	// namespaced.
	ns, err := db.Namespaces.GetByID(ctx, campaign.NamespaceOrgID, campaign.NamespaceUserID)
	if err != nil {
		return errors.Wrap(err, "retrieving namespace")
	}

	url, err := campaignURL(ctx, ns, campaign)
	if err != nil {
		return errors.Wrap(err, "building URL")
	}

	cs.Body = fmt.Sprintf(
		"%s\n\n[_Created by Sourcegraph campaign `%s/%s`._](%s)",
		cs.Body, ns.Name, campaign.Name, url,
	)

	return nil
}

// internalClient is here for mocking reasons.
var internalClient interface {
	ExternalURL(context.Context) (string, error)
} = api.InternalClient

func campaignURL(ctx context.Context, ns *db.Namespace, c *campaigns.Campaign) (string, error) {
	// To build the absolute URL, we need to know where Sourcegraph is!
	extStr, err := internalClient.ExternalURL(ctx)
	if err != nil {
		return "", errors.Wrap(err, "getting external Sourcegraph URL")
	}

	extURL, err := url.Parse(extStr)
	if err != nil {
		return "", errors.Wrap(err, "parsing external Sourcegraph URL")
	}

	// This needs to be kept consistent with resolvers.campaignURL().
	// (Refactoring the resolver to use the same function is difficult due to
	// the different querying and caching behaviour in GraphQL resolvers, so we
	// simply replicate the logic here.)
	u := extURL.ResolveReference(&url.URL{Path: namespaceURL(ns) + "/campaigns/" + c.Name})

	return u.String(), nil
}

func namespaceURL(ns *db.Namespace) string {
	prefix := "/users/"
	if ns.Organization != 0 {
		prefix = "/organizations/"
	}

	return prefix + ns.Name
}

func compareChangesetSpecs(previous, current *campaigns.ChangesetSpec) (*ChangesetSpecDelta, error) {
	delta := &ChangesetSpecDelta{}

	if previous == nil {
		return delta, nil
	}

	if previous.Spec.Title != current.Spec.Title {
		delta.TitleChanged = true
	}
	if previous.Spec.Body != current.Spec.Body {
		delta.BodyChanged = true
	}
	if previous.Spec.BaseRef != current.Spec.BaseRef {
		delta.BaseRefChanged = true
	}

	// If was set to "draft" and now "true", need to undraft the changeset.
	// We currently ignore going from "true" to "draft".
	if previous.Spec.Published.Draft() && current.Spec.Published.True() {
		delta.Undraft = true
	}

	// Diff
	currentDiff, err := current.Spec.Diff()
	if err != nil {
		return nil, nil
	}
	previousDiff, err := previous.Spec.Diff()
	if err != nil {
		return nil, err
	}
	if previousDiff != currentDiff {
		delta.DiffChanged = true
	}

	// CommitMessage
	currentCommitMessage, err := current.Spec.CommitMessage()
	if err != nil {
		return nil, nil
	}
	previousCommitMessage, err := previous.Spec.CommitMessage()
	if err != nil {
		return nil, err
	}
	if previousCommitMessage != currentCommitMessage {
		delta.CommitMessageChanged = true
	}

	// AuthorName
	currentAuthorName, err := current.Spec.AuthorName()
	if err != nil {
		return nil, nil
	}
	previousAuthorName, err := previous.Spec.AuthorName()
	if err != nil {
		return nil, err
	}
	if previousAuthorName != currentAuthorName {
		delta.AuthorNameChanged = true
	}

	// AuthorEmail
	currentAuthorEmail, err := current.Spec.AuthorEmail()
	if err != nil {
		return nil, nil
	}
	previousAuthorEmail, err := previous.Spec.AuthorEmail()
	if err != nil {
		return nil, err
	}
	if previousAuthorEmail != currentAuthorEmail {
		delta.AuthorEmailChanged = true
	}

	return delta, nil
}

type ChangesetSpecDelta struct {
	TitleChanged         bool
	BodyChanged          bool
	Undraft              bool
	BaseRefChanged       bool
	DiffChanged          bool
	CommitMessageChanged bool
	AuthorNameChanged    bool
	AuthorEmailChanged   bool
}

func (d *ChangesetSpecDelta) String() string { return fmt.Sprintf("%#v", d) }

func (d *ChangesetSpecDelta) NeedCommitUpdate() bool {
	return d.DiffChanged || d.CommitMessageChanged || d.AuthorNameChanged || d.AuthorEmailChanged
}

func (d *ChangesetSpecDelta) NeedCodeHostUpdate() bool {
	return d.TitleChanged || d.BodyChanged || d.BaseRefChanged
}

func (d *ChangesetSpecDelta) AttributesChanged() bool {
	return d.NeedCommitUpdate() || d.NeedCodeHostUpdate()
}
