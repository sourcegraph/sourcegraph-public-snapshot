package campaigns

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// ErrApplyClosedCampaign is returned by ApplyCampaign when the campaign
// matched by the campaign spec is already closed.
var ErrApplyClosedCampaign = errors.New("existing campaign matched by campaign spec is closed")

// ErrMatchingCampaignExists is returned by ApplyCampaign if a campaign matching the
// campaign spec already exists and FailIfExists was set.
var ErrMatchingCampaignExists = errors.New("a campaign matching the given campaign spec already exists")

type ApplyCampaignOpts struct {
	CampaignSpecRandID string
	EnsureCampaignID   int64

	// When FailIfCampaignExists is true, ApplyCampaign will fail if a Campaign
	// matching the given CampaignSpec already exists.
	FailIfCampaignExists bool
}

func (o ApplyCampaignOpts) String() string {
	return fmt.Sprintf(
		"CampaignSpec %s, EnsureCampaignID %d",
		o.CampaignSpecRandID,
		o.EnsureCampaignID,
	)
}

// ApplyCampaign creates the CampaignSpec.
func (s *Service) ApplyCampaign(ctx context.Context, opts ApplyCampaignOpts) (campaign *campaigns.Campaign, err error) {
	tr, ctx := trace.New(ctx, "Service.ApplyCampaign", opts.String())
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	campaignSpec, err := s.store.GetCampaignSpec(ctx, GetCampaignSpecOpts{
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

	campaign, previousSpecID, err := s.ReconcileCampaign(ctx, campaignSpec)
	if err != nil {
		return nil, err
	}

	if campaign.ID != 0 && opts.FailIfCampaignExists {
		return nil, ErrMatchingCampaignExists
	}

	if opts.EnsureCampaignID != 0 && campaign.ID != opts.EnsureCampaignID {
		return nil, ErrEnsureCampaignFailed
	}

	if campaign.Closed() {
		return nil, ErrApplyClosedCampaign
	}

	if previousSpecID == campaignSpec.ID {
		return campaign, nil
	}

	// Before we write to the database in a transaction, we cancel all
	// currently enqueued/errored-and-retryable changesets the campaign might
	// have.
	// We do this so we don't continue to possibly create changesets on the
	// codehost while we're applying a new campaign spec.
	// This is blocking, because the changeset rows currently being processed by the
	// reconciler are locked.
	if err := s.store.CancelQueuedCampaignChangesets(ctx, campaign.ID); err != nil {
		return campaign, nil
	}

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	if campaign.ID == 0 {
		err := tx.CreateCampaign(ctx, campaign)
		if err != nil {
			return nil, err
		}
	}

	rstore := repos.NewDBStore(tx.DB(), sql.TxOptions{})

	// Now we need to wire up the ChangesetSpecs of the new CampaignSpec
	// correctly with the Changesets so that the reconciler can create/update
	// them.

	// Load the mapping between ChangesetSpecs and existing Changesets in the target campaign.
	mappings, err := tx.GetRewirerMappings(ctx, GetRewirerMappingsOpts{
		CampaignSpecID: campaign.CampaignSpecID,
		CampaignID:     campaign.ID,
	})
	if err != nil {
		return nil, err
	}

	// And execute the mapping.
	rewirer := &changesetRewirer{
		tx:       tx,
		rstore:   rstore,
		campaign: campaign,
		mappings: mappings,
	}
	changesets, err := rewirer.Rewire(ctx)
	if err != nil {
		return nil, err
	}

	// Reset the attached changesets.
	campaign.ChangesetIDs = []int64{}
	for _, changeset := range changesets {
		if err := tx.UpsertChangeset(ctx, changeset); err != nil {
			return nil, err
		}
		campaign.ChangesetIDs = append(campaign.ChangesetIDs, changeset.ID)
	}

	if err := tx.UpdateCampaign(ctx, campaign); err != nil {
		return nil, err
	}

	return campaign, nil
}

func (s *Service) ReconcileCampaign(ctx context.Context, campaignSpec *campaigns.CampaignSpec) (campaign *campaigns.Campaign, previousSpecID int64, err error) {
	campaign, err = s.GetCampaignMatchingCampaignSpec(ctx, campaignSpec)
	if err != nil {
		return nil, 0, err
	}
	if campaign == nil {
		campaign = &campaigns.Campaign{}
	} else {
		previousSpecID = campaign.CampaignSpecID
	}
	// Populate the campaign with the values from the campaign spec.
	campaign.CampaignSpecID = campaignSpec.ID
	campaign.NamespaceOrgID = campaignSpec.NamespaceOrgID
	campaign.NamespaceUserID = campaignSpec.NamespaceUserID
	campaign.Name = campaignSpec.Spec.Name
	actor := actor.FromContext(ctx)
	if campaign.InitialApplierID == 0 {
		campaign.InitialApplierID = actor.UID
	}
	campaign.LastApplierID = actor.UID
	campaign.LastAppliedAt = s.clock()
	campaign.Description = campaignSpec.Spec.Description
	return campaign, previousSpecID, nil
}

type changesetRewirer struct {
	mappings RewirerMappings
	campaign *campaigns.Campaign
	tx       *Store
	rstore   repos.Store
}

// Rewire uses RewirerMappings (mapping ChangesetSpecs to matching Changesets) generated by Store.GetRewirerMappings to update the Changesets
// for consumption by the background reconciler.
//
// It also updates the ChangesetIDs on the campaign.
func (r *changesetRewirer) Rewire(ctx context.Context) (changesets []*campaigns.Changeset, err error) {
	// First we need to load the associations.
	associations, err := r.loadAssociations(ctx)
	if err != nil {
		return nil, err
	}

	changesets = []*campaigns.Changeset{}

	for _, m := range r.mappings {
		// If a Changeset that's currently attached to the campaign wasn't matched to a ChangesetSpec, it needs to be closed/detached.
		if m.ChangesetSpecID == 0 {
			changeset, ok := associations.changesetsByID[m.ChangesetID]
			if !ok {
				// This should never happen.
				return nil, errors.New("changeset not found")
			}

			// If we don't have access to a repository, we don't detach nor close the changeset.
			_, ok = associations.accessibleReposByID[m.RepoID]
			if !ok {
				continue
			}

			if err := r.closeChangeset(ctx, changeset); err != nil {
				return nil, err
			}

			continue
		}

		spec, ok := associations.changesetSpecsByID[m.ChangesetSpecID]
		if !ok {
			// This should never happen.
			return nil, errors.New("spec not found")
		}

		// If we don't have access to a repository, we return an error. Why not
		// simply skip the repository? If we skip it, the user can't reapply
		// the same campaign spec, since it's already applied and re-applying
		// would require a new spec.
		repo, ok := associations.accessibleReposByID[m.RepoID]
		if !ok {
			return nil, &db.RepoNotFoundErr{ID: m.RepoID}
		}

		if err := checkRepoSupported(repo); err != nil {
			return nil, err
		}

		var changeset *campaigns.Changeset

		if m.ChangesetID != 0 {
			changeset, ok = associations.changesetsByID[m.ChangesetID]
			if !ok {
				// This should never happen.
				return nil, errors.New("changeset not found")
			}
			if spec.Spec.IsImportingExisting() {
				r.attachTrackingChangeset(changeset)
			} else if spec.Spec.IsBranch() {
				r.updateChangesetToNewSpec(changeset, spec)
			}
		} else {
			if spec.Spec.IsImportingExisting() {
				changeset = r.createTrackingChangeset(repo, spec.Spec.ExternalID)
			} else if spec.Spec.IsBranch() {
				changeset = r.createChangesetForSpec(repo, spec)
			}
		}
		changesets = append(changesets, changeset)
	}

	return changesets, nil
}

func (r *changesetRewirer) createChangesetForSpec(repo *types.Repo, spec *campaigns.ChangesetSpec) *campaigns.Changeset {
	newChangeset := &campaigns.Changeset{
		RepoID:              spec.RepoID,
		ExternalServiceType: repo.ExternalRepo.ServiceType,

		CampaignIDs:       []int64{r.campaign.ID},
		OwnedByCampaignID: r.campaign.ID,
		CurrentSpecID:     spec.ID,

		PublicationState: campaigns.ChangesetPublicationStateUnpublished,
		ReconcilerState:  campaigns.ReconcilerStateQueued,
	}

	// Copy over diff stat from the spec.
	diffStat := spec.DiffStat()
	newChangeset.SetDiffStat(&diffStat)

	return newChangeset
}

func (r *changesetRewirer) updateChangesetToNewSpec(c *campaigns.Changeset, spec *campaigns.ChangesetSpec) {
	c.PreviousSpecID = c.CurrentSpecID
	c.CurrentSpecID = spec.ID

	// Ensure that the changeset is attached to the campaign
	c.CampaignIDs = append(c.CampaignIDs, r.campaign.ID)

	// Copy over diff stat from the new spec.
	diffStat := spec.DiffStat()
	c.SetDiffStat(&diffStat)

	// We need to enqueue it for the changeset reconciler, so the
	// reconciler wakes up, compares old and new spec and, if
	// necessary, updates the changesets accordingly.
	c.ResetQueued()
}

func (r *changesetRewirer) createTrackingChangeset(repo *types.Repo, externalID string) *campaigns.Changeset {
	newChangeset := &campaigns.Changeset{
		RepoID:              repo.ID,
		ExternalServiceType: repo.ExternalRepo.ServiceType,

		CampaignIDs:     []int64{r.campaign.ID},
		ExternalID:      externalID,
		AddedToCampaign: true,
		// Note: no CurrentSpecID, because we merely track this one

		PublicationState: campaigns.ChangesetPublicationStatePublished,

		// Enqueue it so the reconciler syncs it.
		ReconcilerState: campaigns.ReconcilerStateQueued,
		Unsynced:        true,
	}

	return newChangeset
}

func (r *changesetRewirer) attachTrackingChangeset(changeset *campaigns.Changeset) {
	// We already have a changeset with the given repoID and
	// externalID, so we can track it.
	changeset.AddedToCampaign = true
	changeset.CampaignIDs = append(changeset.CampaignIDs, r.campaign.ID)

	// If it's errored and not created by another campaign, we re-enqueue it.
	if changeset.OwnedByCampaignID == 0 && changeset.ReconcilerState == campaigns.ReconcilerStateErrored {
		changeset.ResetQueued()
	}
}

func (r *changesetRewirer) closeChangeset(ctx context.Context, changeset *campaigns.Changeset) error {
	if changeset.CurrentSpecID != 0 && changeset.OwnedByCampaignID == r.campaign.ID {
		// If we have a current spec ID and the changeset was created by
		// _this_ campaign that means we should detach and close it.

		// But only if it was created on the code host:
		if changeset.Published() {
			changeset.Closing = true
			changeset.ResetQueued()
		} else {
			// otherwise we simply delete it.
			return r.tx.DeleteChangeset(ctx, changeset.ID)
		}
	}

	// Disassociate the changeset with the campaign.
	changeset.RemoveCampaignID(r.campaign.ID)
	return r.tx.UpdateChangeset(ctx, changeset)
}

type rewirerAssociations struct {
	accessibleReposByID map[api.RepoID]*types.Repo
	changesetsByID      map[int64]*campaigns.Changeset
	changesetSpecsByID  map[int64]*campaigns.ChangesetSpec
}

// loadAssociations retrieves all entities required to rewire the changesets in a campaign.
func (r *changesetRewirer) loadAssociations(ctx context.Context) (associations *rewirerAssociations, err error) {
	// Fetch the changeset specs involved in this rewiring. This should always be the same as omitting the `IDs` section,
	// we just make sure people know why that is the case here.
	changesetSpecs, _, err := r.tx.ListChangesetSpecs(ctx, ListChangesetSpecsOpts{
		CampaignSpecID: r.campaign.CampaignSpecID,
		IDs:            r.mappings.ChangesetSpecIDs(),
	})
	if err != nil {
		return nil, err
	}

	// Then fetch the changesets involved in this rewiring.
	changesets, _, err := r.tx.ListChangesets(ctx, ListChangesetsOpts{IDs: r.mappings.ChangesetIDs()})
	if err != nil {
		return nil, err
	}

	associations = &rewirerAssociations{}
	// Fetch all repos involved. We use them later to enforce repo permissions.
	//
	// 🚨 SECURITY: db.Repos.GetRepoIDsSet uses the authzFilter under the hood and
	// filters out repositories that the user doesn't have access to.
	associations.accessibleReposByID, err = db.Repos.GetReposSetByIDs(ctx, r.mappings.RepoIDs()...)
	if err != nil {
		return nil, err
	}

	associations.changesetsByID = map[int64]*campaigns.Changeset{}
	associations.changesetSpecsByID = map[int64]*campaigns.ChangesetSpec{}

	for _, c := range changesets {
		associations.changesetsByID[c.ID] = c
	}
	for _, c := range changesetSpecs {
		associations.changesetSpecsByID[c.ID] = c
	}

	return associations, nil
}
