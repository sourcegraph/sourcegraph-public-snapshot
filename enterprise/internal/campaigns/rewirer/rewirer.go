package rewirer

import (
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type ChangesetRewirer struct {
	// The mappings need to be hydrated for the ChangesetRewirer to consume them.
	mappings   store.RewirerMappings
	campaignID int64
}

func New(mappings store.RewirerMappings, campaignID int64) *ChangesetRewirer {
	return &ChangesetRewirer{
		mappings:   mappings,
		campaignID: campaignID,
	}
}

// Rewire uses RewirerMappings (mapping ChangesetSpecs to matching Changesets) generated by Store.GetRewirerMappings to update the Changesets
// for consumption by the background reconciler.
//
// It also updates the ChangesetIDs on the campaign.
func (r *ChangesetRewirer) Rewire() (changesets []*campaigns.Changeset, err error) {
	changesets = []*campaigns.Changeset{}

	for _, m := range r.mappings {
		// If a Changeset that's currently attached to the campaign wasn't matched to a ChangesetSpec, it needs to be closed/detached.
		if m.ChangesetSpecID == 0 {
			changeset := m.Changeset

			// If we don't have access to a repository, we don't detach nor close the changeset.
			if m.Repo == nil {
				continue
			}

			// If the changeset is currently not attached to this campaign, we don't want to modify it.
			if !changeset.AttachedTo(r.campaignID) {
				continue
			}

			r.closeChangeset(changeset)
			changesets = append(changesets, changeset)

			continue
		}

		spec := m.ChangesetSpec

		// If we don't have access to a repository, we return an error. Why not
		// simply skip the repository? If we skip it, the user can't reapply
		// the same campaign spec, since it's already applied and re-applying
		// would require a new spec.
		repo := m.Repo
		if repo == nil {
			return nil, &db.RepoNotFoundErr{ID: m.RepoID}
		}

		if err := checkRepoSupported(repo); err != nil {
			return nil, err
		}

		var changeset *campaigns.Changeset

		if m.ChangesetID != 0 {
			changeset = m.Changeset
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

func (r *ChangesetRewirer) createChangesetForSpec(repo *types.Repo, spec *campaigns.ChangesetSpec) *campaigns.Changeset {
	newChangeset := &campaigns.Changeset{
		RepoID:              spec.RepoID,
		ExternalServiceType: repo.ExternalRepo.ServiceType,

		CampaignIDs:       []int64{r.campaignID},
		OwnedByCampaignID: r.campaignID,
		CurrentSpecID:     spec.ID,

		PublicationState: campaigns.ChangesetPublicationStateUnpublished,
		ReconcilerState:  campaigns.ReconcilerStateQueued,
	}

	// Copy over diff stat from the spec.
	diffStat := spec.DiffStat()
	newChangeset.SetDiffStat(&diffStat)

	return newChangeset
}

func (r *ChangesetRewirer) updateChangesetToNewSpec(c *campaigns.Changeset, spec *campaigns.ChangesetSpec) {
	if c.ReconcilerState == campaigns.ReconcilerStateCompleted {
		c.PreviousSpecID = c.CurrentSpecID
	}
	c.CurrentSpecID = spec.ID

	// Ensure that the changeset is attached to the campaign
	c.CampaignIDs = append(c.CampaignIDs, r.campaignID)

	// Copy over diff stat from the new spec.
	diffStat := spec.DiffStat()
	c.SetDiffStat(&diffStat)

	// We need to enqueue it for the changeset reconciler, so the
	// reconciler wakes up, compares old and new spec and, if
	// necessary, updates the changesets accordingly.
	c.ResetQueued()
}

func (r *ChangesetRewirer) createTrackingChangeset(repo *types.Repo, externalID string) *campaigns.Changeset {
	newChangeset := &campaigns.Changeset{
		RepoID:              repo.ID,
		ExternalServiceType: repo.ExternalRepo.ServiceType,

		CampaignIDs: []int64{r.campaignID},
		ExternalID:  externalID,
		// Note: no CurrentSpecID, because we merely track this one

		PublicationState: campaigns.ChangesetPublicationStatePublished,

		// Enqueue it so the reconciler syncs it.
		ReconcilerState: campaigns.ReconcilerStateQueued,
		Unsynced:        true,
	}

	return newChangeset
}

func (r *ChangesetRewirer) attachTrackingChangeset(changeset *campaigns.Changeset) {
	// We already have a changeset with the given repoID and
	// externalID, so we can track it.
	changeset.CampaignIDs = append(changeset.CampaignIDs, r.campaignID)

	// If it's errored and not created by another campaign, we re-enqueue it.
	if changeset.OwnedByCampaignID == 0 && changeset.ReconcilerState == campaigns.ReconcilerStateErrored {
		changeset.ResetQueued()
	}
}

func (r *ChangesetRewirer) closeChangeset(changeset *campaigns.Changeset) {
	if changeset.CurrentSpecID != 0 && changeset.OwnedByCampaignID == r.campaignID {
		// If we have a current spec ID and the changeset was created by
		// _this_ campaign that means we should detach and close it.
		if changeset.Published() {
			// Store the current spec also as the previous spec.
			//
			// Why?
			//
			// When a changeset with (prev: A, curr: B) should be closed but
			// closing failed, it will still have (prev: A, curr: B) set.
			//
			// If someone then applies a new campaign spec and re-attaches that
			// changeset with changeset spec C, the changeset would end up with
			// (prev: A, curr: C), because we don't rotate specs on errors in
			// `updateChangesetToNewSpec`.
			//
			// That would mean, though, that the delta between A and C tells us
			// to repush and update the changeset on the code host, in addition
			// to 'reopen', which would actually be the only required action.
			//
			// So, when we mark a changeset as to-be-closed, we also rotate the
			// specs, so that it changeset is saved as (prev: B, curr: B) and
			// when somebody re-attaches it it's (prev: B, curr: C).
			// But we only rotate the spec, if applying the currentSpecID was
			// successful:
			if changeset.ReconcilerState == campaigns.ReconcilerStateCompleted {
				changeset.PreviousSpecID = changeset.CurrentSpecID
			}
			changeset.Closing = true
			changeset.ResetQueued()
		}
	}

	// Disassociate the changeset with the campaign.
	changeset.RemoveCampaignID(r.campaignID)
}

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
