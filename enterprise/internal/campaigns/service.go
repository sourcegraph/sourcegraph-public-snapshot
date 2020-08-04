package campaigns

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

// NewService returns a Service.
func NewService(store *Store, cf *httpcli.Factory) *Service {
	return NewServiceWithClock(store, cf, store.Clock())
}

// NewServiceWithClock returns a Service the given clock used
// to generate timestamps.
func NewServiceWithClock(store *Store, cf *httpcli.Factory, clock func() time.Time) *Service {
	svc := &Service{store: store, cf: cf, clock: clock}

	return svc
}

type Service struct {
	store *Store
	cf    *httpcli.Factory

	sourcer repos.Sourcer

	clock func() time.Time
}

// CreateCampaign creates the Campaign.
func (s *Service) CreateCampaign(ctx context.Context, c *campaigns.Campaign) (err error) {
	tr, ctx := trace.New(ctx, "Service.CreateCampaign", fmt.Sprintf("Name: %q", c.Name))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if c.Name == "" {
		return ErrCampaignNameBlank
	}

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	c.CreatedAt = s.clock()
	c.UpdatedAt = c.CreatedAt

	err = tx.CreateCampaign(ctx, c)
	if err != nil {
		return err
	}

	if c.Branch != "" {
		err = validateCampaignBranch(c.Branch)
		if err != nil {
			return err
		}
	}

	return nil
}

type CreateCampaignSpecOpts struct {
	RawSpec string

	NamespaceUserID int32
	NamespaceOrgID  int32

	ChangesetSpecRandIDs []string
}

// CreateCampaignSpec creates the CampaignSpec.
func (s *Service) CreateCampaignSpec(ctx context.Context, opts CreateCampaignSpecOpts) (spec *campaigns.CampaignSpec, err error) {
	actor := actor.FromContext(ctx)
	tr, ctx := trace.New(ctx, "Service.CreateCampaignSpec", fmt.Sprintf("Actor %s", actor))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	spec, err = campaigns.NewCampaignSpecFromRaw(opts.RawSpec)
	if err != nil {
		return nil, err
	}

	// Check whether the current user has access to either one of the namespaces.
	err = checkNamespaceAccess(ctx, opts.NamespaceUserID, opts.NamespaceOrgID)
	if err != nil {
		return nil, err
	}
	spec.NamespaceOrgID = opts.NamespaceOrgID
	spec.NamespaceUserID = opts.NamespaceUserID
	spec.UserID = actor.UID

	listOpts := ListChangesetSpecsOpts{Limit: -1, RandIDs: opts.ChangesetSpecRandIDs}
	cs, _, err := s.store.ListChangesetSpecs(ctx, listOpts)
	if err != nil {
		return nil, err
	}

	repoIDs := make([]api.RepoID, 0, len(cs))
	for _, c := range cs {
		repoIDs = append(repoIDs, c.RepoID)
	}

	accessibleReposByID, err := accessibleRepos(ctx, repoIDs)
	if err != nil {
		return nil, err
	}

	byRandID := make(map[string]*campaigns.ChangesetSpec, len(cs))
	for _, changesetSpec := range cs {
		// 🚨 SECURITY: We return an error if the user doesn't have access to one
		// of the repositories associated with a ChangesetSpec.
		if _, ok := accessibleReposByID[changesetSpec.RepoID]; !ok {
			return nil, &db.RepoNotFoundErr{ID: changesetSpec.RepoID}
		}
		byRandID[changesetSpec.RandID] = changesetSpec
	}

	// Check if a changesetSpec was not found
	for _, randID := range opts.ChangesetSpecRandIDs {
		if _, ok := byRandID[randID]; !ok {
			return nil, &changesetSpecNotFoundErr{RandID: randID}
		}
	}

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.CreateCampaignSpec(ctx, spec); err != nil {
		return nil, err
	}

	for _, changesetSpec := range cs {
		changesetSpec.CampaignSpecID = spec.ID

		if err := tx.UpdateChangesetSpec(ctx, changesetSpec); err != nil {
			return nil, err
		}
	}

	return spec, nil
}

// CreateChangesetSpec validates the given raw spec input and creates the ChangesetSpec.
func (s *Service) CreateChangesetSpec(ctx context.Context, rawSpec string, userID int32) (spec *campaigns.ChangesetSpec, err error) {
	tr, ctx := trace.New(ctx, "Service.CreateChangesetSpec", fmt.Sprintf("User %d", userID))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	spec, err = campaigns.NewChangesetSpecFromRaw(rawSpec)
	if err != nil {
		return nil, err
	}
	spec.UserID = userID
	spec.RepoID, err = graphqlbackend.UnmarshalRepositoryID(spec.Spec.BaseRepository)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: We use db.Repos.Get to check whether the user has access to
	// the repository or not.
	if _, err = db.Repos.Get(ctx, spec.RepoID); err != nil {
		return nil, err
	}

	return spec, s.store.CreateChangesetSpec(ctx, spec)
}

// changesetSpecNotFoundErr is returned by CreateCampaignSpec if a
// ChangesetSpec with the given RandID doesn't exist.
// It fulfills the interface required by errcode.IsNotFound.
type changesetSpecNotFoundErr struct {
	RandID string
}

func (e *changesetSpecNotFoundErr) Error() string {
	if e.RandID != "" {
		return fmt.Sprintf("changesetSpec not found: id=%s", e.RandID)
	}
	return "changesetSpec not found"
}

func (e *changesetSpecNotFoundErr) NotFound() bool { return true }

type ApplyCampaignOpts struct {
	CampaignSpecRandID string
	EnsureCampaignID   int64
}

func (o ApplyCampaignOpts) String() string {
	return fmt.Sprintf(
		"CampaignSpec %s, EnsureCampaignID %d",
		o.CampaignSpecRandID,
		o.EnsureCampaignID,
	)
}

// mockApplyCampaignCloseChangesets is used to test ApplyCampaign closing
// detached changesets.
// This is a temporary mock that should be removed once we move closing of
// changesets into the background.
var mockApplyCampaignCloseChangesets func(campaigns.Changesets)

// ApplyCampaign creates the CampaignSpec.
func (s *Service) ApplyCampaign(ctx context.Context, opts ApplyCampaignOpts) (campaign *campaigns.Campaign, err error) {
	tr, ctx := trace.New(ctx, "Service.ApplyCampaign", opts.String())
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// Setup a defer func that gets executed _after_ the `tx.Done(err)` below.
	toClose := campaigns.Changesets{}
	defer func() {
		if mockApplyCampaignCloseChangesets != nil {
			mockApplyCampaignCloseChangesets(toClose)
			return
		}

		// So if err is not nil, the transaction has been rolled back.
		if err != nil {
			return
		}
		// If not, we launch a goroutine that closes the changesets added to
		// toClose in the background.
		go func() {
			ctx := trace.ContextWithTrace(context.Background(), tr)

			// Close only the changesets that are open
			err := s.CloseOpenChangesets(ctx, toClose)
			if err != nil {
				log15.Error("CloseCampaignChangesets", "err", err)
			}
		}()
	}()

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	rstore := repos.NewDBStore(tx.DB(), sql.TxOptions{})

	campaignSpec, err := tx.GetCampaignSpec(ctx, GetCampaignSpecOpts{
		RandID: opts.CampaignSpecRandID,
	})
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only site-admins or the creator of campaignSpec can apply
	// campaignSpec.
	if err := backend.CheckSiteAdminOrSameUser(ctx, campaignSpec.UserID); err != nil {
		return nil, err
	}

	getOpts := GetCampaignOpts{
		Name:            campaignSpec.Spec.Name,
		NamespaceUserID: campaignSpec.NamespaceUserID,
		NamespaceOrgID:  campaignSpec.NamespaceOrgID,
	}

	campaign, err = tx.GetCampaign(ctx, getOpts)
	if err != nil {
		if err != ErrNoResults {
			return nil, err
		}
		err = nil
	}
	if campaign == nil {
		campaign = &campaigns.Campaign{}
	}

	if opts.EnsureCampaignID != 0 && campaign.ID != opts.EnsureCampaignID {
		return nil, ErrEnsureCampaignFailed
	}

	if campaign.CampaignSpecID == campaignSpec.ID {
		return campaign, nil
	}

	campaign.CampaignSpecID = campaignSpec.ID
	campaign.AuthorID = campaignSpec.UserID
	campaign.NamespaceOrgID = campaignSpec.NamespaceOrgID
	campaign.NamespaceUserID = campaignSpec.NamespaceUserID
	campaign.Name = campaignSpec.Spec.Name

	campaign.Description = campaignSpec.Spec.Description
	// TODO(mrnugget): This doesn't need to be populated, since the branch is
	// now ChangesetSpec.Spec.HeadRef.
	campaign.Branch = campaignSpec.Spec.ChangesetTemplate.Branch

	if campaign.ID == 0 {
		err := tx.CreateCampaign(ctx, campaign)
		if err != nil {
			return nil, err
		}
	}

	// Now we need to wire up the ChangesetSpecs of the new CampaignSpec
	// correctly with the Changesets so that the reconciler can create/update
	// them.

	// Load all of the new ChangesetSpecs
	newChangesetSpecs, _, err := tx.ListChangesetSpecs(ctx, ListChangesetSpecsOpts{
		Limit:          -1,
		CampaignSpecID: campaign.CampaignSpecID,
	})
	if err != nil {
		return nil, err
	}

	// Load all Changesets attached to this Campaign.
	changesets, _, err := tx.ListChangesets(ctx, ListChangesetsOpts{CampaignID: campaign.ID})
	if err != nil {
		return nil, err
	}

	// We load all the repositories involved, checking for repository permissions
	// under the hood.
	repoIDs := make([]api.RepoID, 0, len(newChangesetSpecs)+len(changesets))
	for _, spec := range newChangesetSpecs {
		repoIDs = append(repoIDs, spec.RepoID)
	}
	for _, changeset := range changesets {
		repoIDs = append(repoIDs, changeset.RepoID)
	}
	accessibleReposByID, err := accessibleRepos(ctx, repoIDs)
	if err != nil {
		return nil, err
	}

	// Now we have two lists:

	// ┌───────────────────────────────────────┐   ┌───────────────────────────────┐
	// │Changeset 1 | Repo A | #111 | run-gofmt│   │  Spec 1 | Repo A | run-gofmt  │
	// └───────────────────────────────────────┘   └───────────────────────────────┘
	// ┌───────────────────────────────────────┐   ┌───────────────────────────────┐
	// │Changeset 2 | Repo B |      | run-gofmt│   │  Spec 2 | Repo B | run-gofmt  │
	// └───────────────────────────────────────┘   └───────────────────────────────┘
	// ┌───────────────────────────────────────┐   ┌───────────────────────────────────┐
	// │Changeset 3 | Repo C | #222 | run-gofmt│   │  Spec 3 | Repo C | run-goimports  │
	// └───────────────────────────────────────┘   └───────────────────────────────────┘
	// ┌───────────────────────────────────────┐   ┌───────────────────────────────┐
	// │Changeset 4 | Repo C | #333 | older-pr │   │    Spec 4 | Repo C | #333     │
	// └───────────────────────────────────────┘   └───────────────────────────────┘

	// We need to:
	// 1. Find out whether our new specs should _update_ an existing
	//    changeset, or whether we need to create a new one.
	// 2. Since we can have multiple changesets per repository, we need to match
	//    based on repo and external ID.
	// 3. But if a changeset wasn't published yet, it doesn't have an external ID.
	//    In that case, we need to check whether the branch on which we _might_
	//    push the commit (because the changeset might not be published
	//    yet) is the same.

	// What we want:
	//
	// ┌───────────────────────────────────────┐    ┌───────────────────────────────┐
	// │Changeset 1 | Repo A | #111 | run-gofmt│───▶│  Spec 1 | Repo A | run-gofmt  │
	// └───────────────────────────────────────┘    └───────────────────────────────┘
	// ┌───────────────────────────────────────┐    ┌───────────────────────────────┐
	// │Changeset 2 | Repo B |      | run-gofmt│───▶│  Spec 2 | Repo B | run-gofmt  │
	// └───────────────────────────────────────┘    └───────────────────────────────┘
	// ┌───────────────────────────────────────┐
	// │Changeset 3 | Repo C | #222 | run-gofmt│
	// └───────────────────────────────────────┘
	// ┌───────────────────────────────────────┐    ┌───────────────────────────────┐
	// │Changeset 4 | Repo C | #333 | older-pr │───▶│    Spec 4 | Repo C | #333     │
	// └───────────────────────────────────────┘    └───────────────────────────────┘
	// ┌───────────────────────────────────────┐    ┌───────────────────────────────────┐
	// │Changeset 5 | Repo C | | run-goimports │───▶│  Spec 3 | Repo C | run-goimports  │
	// └───────────────────────────────────────┘    └───────────────────────────────────┘
	//
	// Spec 1 should be attached to Changeset 1 and (possibly) update its title/body/diff.
	// Spec 2 should be attached to Changeset 2 and publish it on the code host.
	// Spec 3 should get a new Changeset, since its branch doesn't match Changeset 3's branch.
	// Spec 4 should be attached to Changeset 4, since it tracks PR #333 in Repo C.
	// Changeset 3 doesn't have a matching spec and should be detached from the campaign (and closed).

	type repoHeadRef struct {
		repo    api.RepoID
		headRef string
	}
	changesetsByRepoHeadRef := map[repoHeadRef]*campaigns.Changeset{}

	type repoExternalID struct {
		repo       api.RepoID
		externalID string
	}
	changesetsByRepoExternalID := map[repoExternalID]*campaigns.Changeset{}

	currentSpecsByChangeset := map[int64]*campaigns.ChangesetSpec{}

	for _, c := range changesets {
		// This is an n+1
		s, err := tx.GetChangesetSpecByID(ctx, c.CurrentSpecID)
		if err != nil {
			return nil, err
		}
		currentSpecsByChangeset[c.ID] = s

		if c.ExternalID != "" {
			k := repoExternalID{repo: c.RepoID, externalID: c.ExternalID}
			changesetsByRepoExternalID[k] = c

			// If it has an externalID but no CurrentSpecID, it is a tracked
			// changeset, and we're done and don't need to match it by HeadRef
			if c.CurrentSpecID == 0 {
				continue
			}
		}

		k := repoHeadRef{repo: c.RepoID}
		if c.ExternalBranch != "" {
			k.headRef = git.EnsureRefPrefix(c.ExternalBranch)
			changesetsByRepoHeadRef[k] = c
			continue
		}

		// If we don't have an ExternalBranch, the changeset hasn't been
		// published yet (or hasn't been synced yet).
		if c.CurrentSpecID != 0 {
			// If we're here, the changeset doesn't have an external branch
			//
			// So we load the spec to get the branch where we _would_ push
			// the commit.

			k.headRef = git.EnsureRefPrefix(s.Spec.HeadRef)
			changesetsByRepoHeadRef[k] = c
		}
	}

	attachedChangesets := map[int64]bool{}
	for _, spec := range newChangesetSpecs {
		// If we don't have access to a repository, we return an error. Why not
		// simply skip the repository? If we skip it, the user can't reapply
		// the same campaign spec, since it's already applied and re-applying
		// would require a new spec.
		repo, ok := accessibleReposByID[spec.RepoID]
		if !ok {
			return nil, &db.RepoNotFoundErr{ID: spec.RepoID}
		}

		if err := checkRepoSupported(repo); err != nil {
			return nil, err
		}

		// If we need to track a changeset, we need to find it.
		if spec.Spec.IsImportingExisting() {
			k := repoExternalID{repo: spec.RepoID, externalID: spec.Spec.ExternalID}

			c, ok := changesetsByRepoExternalID[k]
			if ok {
				// If we have the changeset, it's already attached to the campaign
				// but we need to keep track of all changesets in campaign
				attachedChangesets[c.ID] = true
			} else {
				// We don't have a changeset with the given repoID and external ID
				existing, err := tx.GetChangeset(ctx, GetChangesetOpts{
					RepoID:              repo.ID,
					ExternalID:          spec.Spec.ExternalID,
					ExternalServiceType: repo.ExternalRepo.ServiceType,
				})
				if err != nil && err != ErrNoResults {
					return nil, err
				}
				if existing != nil {
					// We already have a changeset with the given repoID and
					// externalID, so we can track it.
					existing.AddedToCampaign = true
					existing.CampaignIDs = append(existing.CampaignIDs, campaign.ID)
					if err = tx.UpdateChangeset(ctx, existing); err != nil {
						return nil, err
					}
					attachedChangesets[existing.ID] = true
				} else {
					newChangeset := &campaigns.Changeset{
						RepoID:              spec.RepoID,
						ExternalServiceType: repo.ExternalRepo.ServiceType,

						CampaignIDs:     []int64{campaign.ID},
						ExternalID:      k.externalID,
						AddedToCampaign: true,
						// Note: no CurrentSpecID, because we merely track this one

						PublicationState: campaigns.ChangesetPublicationStatePublished,
						ReconcilerState:  campaigns.ReconcilerStateCompleted,
					}

					if err = tx.CreateChangeset(ctx, newChangeset); err != nil {
						return nil, err
					}

					// TODO: Now we're syncing in the request path to ensure
					// that the remote changeset exists and also to remove the possibility
					// of an unsynced changeset entering our database
					// IMPORTANT: We need to move that to the reconciler/syncer/background.
					if err = SyncChangesets(ctx, rstore, tx, s.cf, newChangeset); err != nil {
						return nil, errors.Wrapf(err, "syncing changeset failed. repo=%q, externalID=%q", repo.Name, k.externalID)
					}

					attachedChangesets[newChangeset.ID] = true
				}

			}
			// We handled both cases for "track existing changeset" spec:
			// 1. Add existing changeset to campaign
			// 2. Create new changeset and sync it
			continue
		}

		// What we're now looking at is a spec that says:
		//   1. Create a PR on this branch in this repo with this title/body/diff
		// or, if the a PR on this branch with this repo already exists:
		//   2. Update the PR on this branch in this repo to have this new title/body/diff
		//
		// So, let's check:
		// Do we already have a changeset on this branch in this repo?
		k := repoHeadRef{repo: spec.RepoID, headRef: git.EnsureRefPrefix(spec.Spec.HeadRef)}
		c, ok := changesetsByRepoHeadRef[k]
		if !ok {
			// No, we don't have a changeset on that branch in this repo.
			// We're going to create one so the changeset reconciler picks it up,
			// creates a commit and pushes it to the branch.
			// Except, of course, if spec.Spec.Published is false, then it doesn't do anything.
			newChangeset := &campaigns.Changeset{
				RepoID:              spec.RepoID,
				ExternalServiceType: repo.ExternalRepo.ServiceType,

				CampaignIDs:       []int64{campaign.ID},
				OwnedByCampaignID: campaign.ID,
				CurrentSpecID:     spec.ID,

				PublicationState: campaigns.ChangesetPublicationStateUnpublished,
				ReconcilerState:  campaigns.ReconcilerStateQueued,
			}

			if err = tx.CreateChangeset(ctx, newChangeset); err != nil {
				return nil, err
			}
			attachedChangesets[newChangeset.ID] = true
		} else {
			// But if we already have a changeset in the given repository with
			// the given branch:
			//
			// We know we want to keep it in the campaign
			attachedChangesets[c.ID] = true

			// And we need to update it to have the new spec
			c.PreviousSpecID = c.CurrentSpecID
			c.CurrentSpecID = spec.ID

			// And we need to enqueue it for the changeset reconciler, so the
			// reconciler wakes up, compares old and new spec and, if
			// necessary, updates the changesets accordingly.
			c.ReconcilerState = campaigns.ReconcilerStateQueued

			if err = tx.UpdateChangeset(ctx, c); err != nil {
				return nil, err
			}
		}
	}

	// We went through all the new changeset specs and either created or
	// updated a changeset.
	// Their IDs are all the IDs of changesets that should be in the campaign:
	campaign.ChangesetIDs = []int64{}
	for changesetID := range attachedChangesets {
		campaign.ChangesetIDs = append(campaign.ChangesetIDs, changesetID)
	}

	// But it's possible that changesets are now detached, like Changeset 3 in
	// the example above.
	// This we need to detach and close.
	for _, c := range changesets {
		if _, ok := attachedChangesets[c.ID]; ok {
			continue
		}

		// If we don't have access to a repository, we don't detach nor close the changeset.
		_, ok := accessibleReposByID[c.RepoID]
		if !ok {
			continue
		}

		if c.CurrentSpecID != 0 && c.OwnedByCampaignID == campaign.ID {
			// If we have a current spec ID and the changeset was created by
			// _this_ campaign that means we should detach and close it.

			// But only if it was created on the code host:
			if c.PublicationState.Published() {
				toClose = append(toClose, c)
			} else {
				// otherwise we simply delete it.
				if err = tx.DeleteChangeset(ctx, c.ID); err != nil {
					return nil, err
				}
				continue
			}
		}

		c.RemoveCampaignID(campaign.ID)
		if err = tx.UpdateChangeset(ctx, c); err != nil {
			return nil, err
		}
	}

	return campaign, tx.UpdateCampaign(ctx, campaign)
}

type MoveCampaignOpts struct {
	CampaignID int64

	NewName string

	NewNamespaceUserID int32
	NewNamespaceOrgID  int32
}

func (o MoveCampaignOpts) String() string {
	return fmt.Sprintf(
		"CampaignID %d, NewName %q, NewNamespaceUserID %d, NewNamespaceOrgID %d",
		o.CampaignID,
		o.NewName,
		o.NewNamespaceUserID,
		o.NewNamespaceOrgID,
	)
}

// MoveCampaign moves the campaign from one namespace to another and/or renames
// the campaign.
func (s *Service) MoveCampaign(ctx context.Context, opts MoveCampaignOpts) (campaign *campaigns.Campaign, err error) {
	tr, ctx := trace.New(ctx, "Service.MoveCampaign", opts.String())
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	campaign, err = tx.GetCampaign(ctx, GetCampaignOpts{ID: opts.CampaignID})
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only the Author of the campaign can move it.
	if err := backend.CheckSiteAdminOrSameUser(ctx, campaign.AuthorID); err != nil {
		return nil, err
	}
	// Check if current user has access to target namespace if set.
	if opts.NewNamespaceOrgID != 0 || opts.NewNamespaceUserID != 0 {
		err = checkNamespaceAccess(ctx, opts.NewNamespaceUserID, opts.NewNamespaceOrgID)
		if err != nil {
			return nil, err
		}
	}

	if opts.NewNamespaceOrgID != 0 {
		campaign.NamespaceOrgID = opts.NewNamespaceOrgID
		campaign.NamespaceUserID = 0
	} else if opts.NewNamespaceUserID != 0 {
		campaign.NamespaceUserID = opts.NewNamespaceUserID
		campaign.NamespaceOrgID = 0
	}

	if opts.NewName != "" {
		campaign.Name = opts.NewName
	}

	return campaign, tx.UpdateCampaign(ctx, campaign)
}

// ErrEnsureCampaignFailed is returned by ApplyCampaign when a ensureCampaignID
// is provided but a campaign with the name specified the campaignSpec exists
// in the given namespace but has a different ID.
var ErrEnsureCampaignFailed = errors.New("a campaign in the given namespace and with the given name exists but does not match the given ID")

// ErrCloseProcessingCampaign is returned by CloseCampaign if the Campaign has
// been published at the time of closing but its ChangesetJobs have not
// finished execution.
var ErrCloseProcessingCampaign = errors.New("cannot close a Campaign while changesets are being created on codehosts")

// CloseCampaign closes the Campaign with the given ID if it has not been closed yet.
func (s *Service) CloseCampaign(ctx context.Context, id int64, closeChangesets bool) (campaign *campaigns.Campaign, err error) {
	traceTitle := fmt.Sprintf("campaign: %d, closeChangesets: %t", id, closeChangesets)
	tr, ctx := trace.New(ctx, "service.CloseCampaign", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	transaction := func() (err error) {
		tx, err := s.store.Transact(ctx)
		if err != nil {
			return err
		}
		defer func() { err = tx.Done(err) }()

		campaign, err = tx.GetCampaign(ctx, GetCampaignOpts{ID: id})
		if err != nil {
			return errors.Wrap(err, "getting campaign")
		}

		if err := backend.CheckSiteAdminOrSameUser(ctx, campaign.AuthorID); err != nil {
			return err
		}

		// TODO: Implement logic to find changesets in PUBLISHING state.
		processing := false
		if processing {
			err = ErrCloseProcessingCampaign
			return err
		}

		if !campaign.ClosedAt.IsZero() {
			return nil
		}

		campaign.ClosedAt = time.Now().UTC()

		return tx.UpdateCampaign(ctx, campaign)
	}

	err = transaction()
	if err != nil {
		return nil, err
	}

	if closeChangesets {
		go func() {
			ctx := trace.ContextWithTrace(context.Background(), tr)

			cs, _, err := s.store.ListChangesets(ctx, ListChangesetsOpts{
				CampaignID: campaign.ID,
				Limit:      -1,
			})
			if err != nil {
				log15.Error("ListChangesets", "err", err)
				return
			}

			// Close only the changesets that are open
			err = s.CloseOpenChangesets(ctx, cs)
			if err != nil {
				log15.Error("CloseCampaignChangesets", "err", err)
			}
		}()
	}

	return campaign, nil
}

// ErrDeleteProcessingCampaign is returned by DeleteCampaign if the Campaign
// has been published at the time of deletion but its ChangesetJobs have not
// finished execution.
var ErrDeleteProcessingCampaign = errors.New("cannot delete a Campaign while changesets are being created on codehosts")

// DeleteCampaign deletes the Campaign with the given ID if it hasn't been
// deleted yet. If closeChangesets is true, the changesets associated with the
// Campaign will be closed on the codehosts.
func (s *Service) DeleteCampaign(ctx context.Context, id int64) (err error) {
	traceTitle := fmt.Sprintf("campaign: %d", id)
	tr, ctx := trace.New(ctx, "service.DeleteCampaign", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	campaign, err := s.store.GetCampaign(ctx, GetCampaignOpts{ID: id})
	if err != nil {
		return err
	}

	if err := backend.CheckSiteAdminOrSameUser(ctx, campaign.AuthorID); err != nil {
		return err
	}

	transaction := func() (err error) {
		tx, err := s.store.Transact(ctx)
		if err != nil {
			return err
		}
		defer func() { err = tx.Done(err) }()

		// TODO: Implement logic to find changesets in PUBLISHING state.
		processing := false
		if processing {
			return ErrDeleteProcessingCampaign
		}

		return tx.DeleteCampaign(ctx, id)
	}

	return transaction()
}

// CloseOpenChangesets closes the given Changesets on their respective codehosts and syncs them.
func (s *Service) CloseOpenChangesets(ctx context.Context, cs campaigns.Changesets) (err error) {
	cs = cs.Filter(func(c *campaigns.Changeset) bool {
		return c.ExternalState == campaigns.ChangesetExternalStateOpen
	})

	if len(cs) == 0 {
		return nil
	}

	accessibleReposByID, err := accessibleRepos(ctx, cs.RepoIDs())
	if err != nil {
		return err
	}

	reposStore := repos.NewDBStore(s.store.DB(), sql.TxOptions{})
	bySource, err := groupChangesetsBySource(ctx, reposStore, s.cf, s.sourcer, cs...)
	if err != nil {
		return err
	}

	errs := &multierror.Error{}
	for _, group := range bySource {
		for _, c := range group.Changesets {
			if _, ok := accessibleReposByID[c.RepoID]; !ok {
				continue
			}

			if err := group.CloseChangeset(ctx, c); err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}

	if len(errs.Errors) != 0 {
		return errs
	}

	// Here we need to sync the just-closed changesets (even though
	// CloseChangesets updates the given Changesets too), because closing a
	// Changeset often produces a ChangesetEvent on the codehost and if we were
	// to close the Changesets and not update the events (which is what
	// syncChangesetsWithSources does) our burndown chart will be outdated
	// until the next run of campaigns.Syncer.
	return syncChangesetsWithSources(ctx, s.store, bySource)
}

// EnqueueChangesetSync loads the given changeset from the database, checks
// whether the actor in the context has permission to enqueue a sync and then
// enqueues a sync by calling the repoupdater client.
func (s *Service) EnqueueChangesetSync(ctx context.Context, id int64) (err error) {
	traceTitle := fmt.Sprintf("changeset: %d", id)
	tr, ctx := trace.New(ctx, "service.EnqueueChangesetSync", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// Check for existence of changeset so we don't swallow that error.
	changeset, err := s.store.GetChangeset(ctx, GetChangesetOpts{ID: id})
	if err != nil {
		return err
	}

	// 🚨 SECURITY: We use db.Repos.Get to check whether the user has access to
	// the repository or not.
	if _, err = db.Repos.Get(ctx, changeset.RepoID); err != nil {
		return err
	}

	campaigns, _, err := s.store.ListCampaigns(ctx, ListCampaignsOpts{ChangesetID: id})
	if err != nil {
		return err
	}

	// Check whether the user has admin rights for one of the campaigns.
	var (
		authErr        error
		hasAdminRights bool
	)

	for _, c := range campaigns {
		err := backend.CheckSiteAdminOrSameUser(ctx, c.AuthorID)
		if err != nil {
			authErr = err
		} else {
			hasAdminRights = true
			break
		}
	}

	if !hasAdminRights {
		return authErr
	}

	if err := repoupdater.DefaultClient.EnqueueChangesetSync(ctx, []int64{id}); err != nil {
		return err
	}

	return nil
}

// ErrCampaignNameBlank is returned by CreateCampaign or UpdateCampaign if the
// specified Campaign name is blank.
var ErrCampaignNameBlank = errors.New("Campaign title cannot be blank")

// ErrCampaignBranchBlank is returned by CreateCampaign or UpdateCampaign if the specified Campaign's
// branch is blank.
var ErrCampaignBranchBlank = errors.New("Campaign branch cannot be blank")

// ErrCampaignBranchInvalid is returned by CreateCampaign or UpdateCampaign if the specified Campaign's
// branch is invalid.
var ErrCampaignBranchInvalid = errors.New("Campaign branch is invalid")

func validateCampaignBranch(branch string) error {
	if branch == "" {
		return ErrCampaignBranchBlank
	}
	if !git.ValidateBranchName(branch) {
		return ErrCampaignBranchInvalid
	}
	return nil
}

// accessibleRepos collects the RepoIDs of the changesets and returns a set of
// the api.RepoID for which the subset of repositories for which the actor in
// ctx has read permissions.
func accessibleRepos(ctx context.Context, ids []api.RepoID) (map[api.RepoID]*types.Repo, error) {
	// 🚨 SECURITY: We use db.Repos.GetByIDs to filter out repositories the
	// user doesn't have access to.
	accessibleRepos, err := db.Repos.GetByIDs(ctx, ids...)
	if err != nil {
		return nil, err
	}

	accessibleRepoIDs := make(map[api.RepoID]*types.Repo, len(accessibleRepos))
	for _, r := range accessibleRepos {
		accessibleRepoIDs[r.ID] = r
	}

	return accessibleRepoIDs, nil
}

// checkNamespaceAccess checks whether the current user in the ctx has access
// to either the user ID or the org ID as a namespace.
// If the userID is non-zero that will be checked. Otherwise the org ID will be
// checked.
// If the current user is an admin, true will be returned.
// Otherwise it checks whether the current user _is_ the namespace user or has
// access to the namespace org.
// If both values are zero, an error is returned.
func checkNamespaceAccess(ctx context.Context, namespaceUserID, namespaceOrgID int32) error {
	if namespaceOrgID != 0 {
		return backend.CheckOrgAccess(ctx, namespaceOrgID)
	} else if namespaceUserID != 0 {
		return backend.CheckSiteAdminOrSameUser(ctx, namespaceUserID)
	} else {
		return ErrNoNamespace
	}
}

// ErrNoNamespace is returned by checkNamespaceAccess if no valid namespace ID is given.
var ErrNoNamespace = errors.New("no namespace given")

// checkRepoSupported checks whether the given repository is supported by campaigns
// and if not it returns an error.
func checkRepoSupported(repo *types.Repo) error {
	if campaigns.IsRepoSupported(&repo.ExternalRepo) {
		return nil
	}

	return errors.Errorf(
		"External service type %s of repository %q is currently not supported for use with campaigns",
		repo.ExternalRepo.ServiceType,
		repo.Name,
	)
}
