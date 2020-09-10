package campaigns

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/schema"
)

type GitserverClient interface {
	CreateCommitFromPatch(ctx context.Context, req protocol.CreateCommitFromPatchRequest) (string, error)
}

// reconciler processes changesets and reconciles their current state — in
// Sourcegraph or on the code host — with that described in the current
// ChangesetSpec associated with the changeset.
type reconciler struct {
	gitserverClient GitserverClient
	sourcer         repos.Sourcer
	store           *Store

	// This is used to disable a time.Sleep in updateChangeset so that the
	// tests don't run slower.
	noSleepBeforeSync bool
}

// HandlerFunc returns a dbworker.HandlerFunc that can be passed to a
// workerutil.Worker to process queued changesets.
func (r *reconciler) HandlerFunc() dbworker.HandlerFunc {
	return func(ctx context.Context, tx dbworkerstore.Store, record workerutil.Record) error {
		return r.process(ctx, r.store.With(tx), record.(*campaigns.Changeset))
	}
}

// process is the main entry point of the reconciler and processes changesets
// that were marked as queued in the database.
//
// For each changeset, the reconciler computes an action to take to reconcile a
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
func (r *reconciler) process(ctx context.Context, tx *Store, ch *campaigns.Changeset) error {
	action, err := determineAction(ctx, tx, ch)
	if err != nil {
		return err
	}

	log15.Info("Reconciler processing changeset", "changeset", ch.ID, "action", action.actionType)

	switch action.actionType {
	case actionSync:
		return r.syncChangeset(ctx, tx, ch)

	case actionPublish:
		return r.publishChangeset(ctx, tx, ch, action.spec)

	case actionUpdate:
		return r.updateChangeset(ctx, tx, ch, action.spec, action.delta)

	case actionClose:
		return r.closeChangeset(ctx, tx, ch)

	case actionNone:
		return nil

	default:
		return fmt.Errorf("Reconciler action %q not implemented", action.actionType)
	}
}

func (r *reconciler) syncChangeset(ctx context.Context, tx *Store, ch *campaigns.Changeset) error {
	rstore := repos.NewDBStore(tx.Handle().DB(), sql.TxOptions{})

	if err := SyncChangesets(ctx, rstore, tx, r.sourcer, ch); err != nil {
		return errors.Wrapf(err, "syncing changeset with external ID %q failed", ch.ExternalID)
	}

	return nil
}

// ErrPublishSameBranch is returned by publish changeset if a changeset with the same external branch
// already exists in the database and is owned by another campaign.
var ErrPublishSameBranch = errors.New("cannot create changeset on the same branch in multiple campaigns")

// publishChangeset creates the given changeset on its code host.
func (r *reconciler) publishChangeset(ctx context.Context, tx *Store, ch *campaigns.Changeset, spec *campaigns.ChangesetSpec) (err error) {
	repo, extSvc, err := loadAssociations(ctx, tx, ch)
	if err != nil {
		return errors.Wrap(err, "failed to load associations")
	}

	existingSameBranch, err := tx.GetChangeset(ctx, GetChangesetOpts{
		ExternalServiceType: ch.ExternalServiceType,
		RepoID:              ch.RepoID,
		ExternalBranch:      git.AbbreviateRef(spec.Spec.HeadRef),
	})
	if err != nil && err != ErrNoResults {
		return err
	}

	if existingSameBranch != nil && existingSameBranch.ID != ch.ID {
		return ErrPublishSameBranch
	}

	// Set up a source with which we can create a changeset
	ccs, err := r.buildChangesetSource(repo, extSvc)
	if err != nil {
		return err
	}

	// Create a commit and push it
	opts, err := buildCommitOpts(repo, spec)
	if err != nil {
		return err
	}
	ref, err := r.pushCommit(ctx, opts)
	if err != nil {
		return err
	}

	// Now create the actual pull request on the code host
	cs := &repos.Changeset{
		Title:     spec.Spec.Title,
		Body:      spec.Spec.Body,
		BaseRef:   spec.Spec.BaseRef,
		HeadRef:   git.EnsureRefPrefix(ref),
		Repo:      repo,
		Changeset: ch,
	}

	// If we're running this method a second time, because we failed due to an
	// ephemeral error, there's a race condition here.
	// It's possible that `CreateChangeset` doesn't return the newest head ref
	// commit yet, because the API of the codehost doesn't return it yet.
	exists, err := ccs.CreateChangeset(ctx, cs)
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
			if err := ccs.UpdateChangeset(ctx, cs); err != nil {
				return errors.Wrap(err, "updating changeset")
			}
		}
	}

	events := ch.Events()
	SetDerivedState(ctx, ch, events)

	if err := tx.UpsertChangesetEvents(ctx, events...); err != nil {
		log15.Error("UpsertChangesetEvents", "err", err)
		return err
	}

	ch.CreatedByCampaign = true
	ch.PublicationState = campaigns.ChangesetPublicationStatePublished
	ch.FailureMessage = nil
	return tx.UpdateChangeset(ctx, ch)
}

// updateChangeset updates the given changeset's attribute on the code host
// according to its ChangesetSpec and the delta previously computed.
// If the delta includes only changes to the commit, updateChangeset will only
// create and force push a new commit.
// If the delta requires updates to the changeset on the code host, it will
// update the changeset there.
func (r *reconciler) updateChangeset(ctx context.Context, tx *Store, ch *campaigns.Changeset, spec *campaigns.ChangesetSpec, delta *changesetSpecDelta) (err error) {
	repo, extSvc, err := loadAssociations(ctx, tx, ch)
	if err != nil {
		return errors.Wrap(err, "failed to load associations")
	}

	// Set up a source with which we can update the changeset on the code host.
	ccs, err := r.buildChangesetSource(repo, extSvc)
	if err != nil {
		return err
	}

	if delta.NeedCommitUpdate() {
		opts, err := buildCommitOpts(repo, spec)
		if err != nil {
			return err
		}

		if _, err = r.pushCommit(ctx, opts); err != nil {
			return err
		}
	}

	// If we only need to update the diff, we're done, because we already
	// pushed the commit. We don't need to update anything on the codehost.
	if !delta.NeedCodeHostUpdate() {
		ch.FailureMessage = nil
		// But we need to sync the changeset so that it has the new commit.
		//
		// The problem: the code host might not have updated the changeset to
		// have the new commit SHA as its head ref oid (and the check states,
		// ...).
		//
		// That's why we give them 3 seconds to update the changesets.
		//
		// Why 3 seconds? Well... 1 or 2 seem to be too short and 4 too long?
		if !r.noSleepBeforeSync {
			time.Sleep(3 * time.Second)
		}
		return r.syncChangeset(ctx, tx, ch)
	}

	// Otherwise, we need to update the pull request on the code host.
	cs := repos.Changeset{
		Title:     spec.Spec.Title,
		Body:      spec.Spec.Body,
		BaseRef:   spec.Spec.BaseRef,
		HeadRef:   git.EnsureRefPrefix(spec.Spec.HeadRef),
		Repo:      repo,
		Changeset: ch,
	}

	if err := ccs.UpdateChangeset(ctx, &cs); err != nil {
		return errors.Wrap(err, "updating changeset")
	}

	// We extract the events, compute derived state and upsert events because
	// the update of the pull request might have changed the changeset on the
	// code host.
	events := ch.Events()
	SetDerivedState(ctx, ch, events)
	if err := tx.UpsertChangesetEvents(ctx, events...); err != nil {
		log15.Error("UpsertChangesetEvents", "err", err)
		return err
	}

	ch.FailureMessage = nil
	return tx.UpdateChangeset(ctx, ch)
}

// closeChangeset closes the given changeset on its code host if its ExternalState is OPEN.
func (r *reconciler) closeChangeset(ctx context.Context, tx *Store, ch *campaigns.Changeset) (err error) {
	ch.Closing = false
	ch.FailureMessage = nil

	if ch.ExternalState != campaigns.ChangesetExternalStateOpen {
		return tx.UpdateChangeset(ctx, ch)
	}

	repo, extSvc, err := loadAssociations(ctx, tx, ch)
	if err != nil {
		return errors.Wrap(err, "failed to load associations")
	}

	// Set up a source with which we can close the changeset
	ccs, err := r.buildChangesetSource(repo, extSvc)
	if err != nil {
		return err
	}

	cs := &repos.Changeset{Changeset: ch}

	if err := ccs.CloseChangeset(ctx, cs); err != nil {
		return errors.Wrap(err, "creating changeset")
	}

	// syncChangeset updates the changeset in the same transaction
	return r.syncChangeset(ctx, tx, ch)
}

func (r *reconciler) pushCommit(ctx context.Context, opts protocol.CreateCommitFromPatchRequest) (string, error) {
	ref, err := r.gitserverClient.CreateCommitFromPatch(ctx, opts)
	if err != nil {
		if diffErr, ok := err.(*protocol.CreateCommitFromPatchError); ok {
			return "", errors.Errorf(
				"creating commit from patch for repository %q: %s\n"+
					"```\n"+
					"$ %s\n"+
					"%s\n"+
					"```",
				diffErr.RepositoryName, diffErr.InternalError, diffErr.Command, strings.TrimSpace(diffErr.CombinedOutput))
		}
		return "", err
	}

	return ref, nil
}

func (r *reconciler) buildChangesetSource(repo *repos.Repo, extSvc *repos.ExternalService) (repos.ChangesetSource, error) {
	sources, err := r.sourcer(extSvc)
	if err != nil {
		return nil, err
	}
	if len(sources) != 1 {
		return nil, errors.New("invalid number of sources for external service")
	}
	src := sources[0]
	ccs, ok := src.(repos.ChangesetSource)
	if !ok {
		return nil, errors.Errorf("creating changesets on code host of repo %q is not implemented", repo.Name)
	}

	return ccs, nil
}

func buildCommitOpts(repo *repos.Repo, spec *campaigns.ChangesetSpec) (protocol.CreateCommitFromPatchRequest, error) {
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
		// `a/` and `/b` filename prefixes. `-p0` tells `git apply` to not
		// expect and strip prefixes.
		GitApplyArgs: []string{"-p0"},
		Push:         true,
	}

	return opts, nil
}

// actionType is an enum to distinguish between different reconcilerActions.
type actionType string

const (
	actionNone    actionType = "none"
	actionUpdate  actionType = "update"
	actionPublish actionType = "publish"
	actionSync    actionType = "sync"
	actionClose   actionType = "close"
)

// reconcilerAction represents the possible actions the reconciler can take for
// a given changeset.
type reconcilerAction struct {
	// The type of actionType.
	actionType actionType

	// The current spec of the changeset.
	spec *campaigns.ChangesetSpec

	// The delta between a possible previous ChangesetSpec and the current
	// ChangesetSpec.
	delta *changesetSpecDelta
}

// determineAction looks at the given changeset to determine what action the
// reconciler should take.
// It loads the current ChangesetSpec and if it exists also the previous one.
// If the current ChangesetSpec is not applied to a campaign, it returns an
// error.
func determineAction(ctx context.Context, tx *Store, ch *campaigns.Changeset) (reconcilerAction, error) {
	action := reconcilerAction{actionType: actionNone}

	// If it doesn't have a spec, it's an imported changeset and we can't do
	// anything.
	if ch.CurrentSpecID == 0 {
		if ch.Unsynced {
			action.actionType = actionSync
		}
		return action, nil
	}

	// If it's marked as closing, we don't need to look at the specs.
	if ch.Closing {
		action.actionType = actionClose
		return action, nil
	}

	curr, err := tx.GetChangesetSpecByID(ctx, ch.CurrentSpecID)
	if err != nil {
		return action, err
	}
	action.spec = curr

	if err := checkSpecAppliedToCampaign(ctx, tx, curr); err != nil {
		return action, err
	}

	var prev *campaigns.ChangesetSpec
	if ch.PreviousSpecID != 0 {
		prev, err = tx.GetChangesetSpecByID(ctx, ch.PreviousSpecID)
		if err != nil {
			return action, err
		}
	}

	switch ch.PublicationState {
	case campaigns.ChangesetPublicationStateUnpublished:
		if curr.Spec.Published {
			action.actionType = actionPublish
		}
	case campaigns.ChangesetPublicationStatePublished:
		delta, err := CompareChangesetSpecs(prev, curr)
		if err != nil {
			return action, nil
		}
		if delta.AttributesChanged() {
			action.actionType = actionUpdate
			action.delta = delta
		}
	default:
		return action, fmt.Errorf("unknown changeset publication state: %s", ch.PublicationState)
	}

	return action, nil
}

func checkSpecAppliedToCampaign(ctx context.Context, tx *Store, spec *campaigns.ChangesetSpec) error {
	campaignSpec, err := tx.GetCampaignSpec(ctx, GetCampaignSpecOpts{ID: spec.CampaignSpecID})
	if err != nil {
		return errors.Wrap(err, "failed to load campaign spec")
	}

	campaign, err := tx.GetCampaign(ctx, GetCampaignOpts{CampaignSpecID: campaignSpec.ID})
	if err != nil && err != ErrNoResults {
		return errors.Wrap(err, "failed to load campaign")
	}

	if campaign == nil || err == ErrNoResults {
		return errors.New("campaign spec is not applied to a campaign")
	}

	return nil
}

func loadAssociations(ctx context.Context, tx *Store, ch *campaigns.Changeset) (*repos.Repo, *repos.ExternalService, error) {
	reposStore := repos.NewDBStore(tx.Handle().DB(), sql.TxOptions{})

	repo, err := loadRepo(ctx, reposStore, ch.RepoID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to load repository")
	}

	extSvc, err := loadExternalService(ctx, reposStore, repo)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to load external service")
	}

	return repo, extSvc, nil
}

func loadRepo(ctx context.Context, tx repos.Store, id api.RepoID) (*repos.Repo, error) {
	rs, err := tx.ListRepos(ctx, repos.StoreListReposArgs{IDs: []api.RepoID{id}})
	if err != nil {
		return nil, err
	}
	if len(rs) != 1 {
		return nil, errors.Errorf("repo not found: %d", id)
	}
	return rs[0], nil
}

func loadExternalService(ctx context.Context, reposStore repos.Store, repo *repos.Repo) (*repos.ExternalService, error) {
	var externalService *repos.ExternalService
	{
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
	}

	if externalService == nil {
		return nil, errors.Errorf("no external services found for repo %q", repo.Name)
	}

	return externalService, nil
}

func CompareChangesetSpecs(previous, current *campaigns.ChangesetSpec) (*changesetSpecDelta, error) {
	delta := &changesetSpecDelta{}

	if previous == nil {
		return delta, nil
	}

	if previous.Spec.Title != current.Spec.Title {
		delta.titleChanged = true
	}
	if previous.Spec.Body != current.Spec.Body {
		delta.bodyChanged = true
	}
	if previous.Spec.BaseRef != current.Spec.BaseRef {
		delta.baseRefChanged = true
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
		delta.diffChanged = true
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
		delta.commitMessageChanged = true
	}

	return delta, nil
}

type changesetSpecDelta struct {
	titleChanged         bool
	bodyChanged          bool
	baseRefChanged       bool
	diffChanged          bool
	commitMessageChanged bool
}

func (d *changesetSpecDelta) String() string { return fmt.Sprintf("%#v", d) }

func (d *changesetSpecDelta) NeedCommitUpdate() bool {
	return d.diffChanged || d.commitMessageChanged
}

func (d *changesetSpecDelta) NeedCodeHostUpdate() bool {
	return d.titleChanged || d.bodyChanged || d.baseRefChanged
}

func (d *changesetSpecDelta) AttributesChanged() bool {
	return d.NeedCommitUpdate() || d.NeedCodeHostUpdate()
}
