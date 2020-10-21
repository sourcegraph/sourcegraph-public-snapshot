package campaigns

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type repoHeadRef struct {
	repo    api.RepoID
	headRef string
}

type repoExternalID struct {
	repo       api.RepoID
	externalID string
}

type rewirerPlan []rewireOperation

func (r rewirerPlan) String() string {
	ss := make([]string, len(r))
	for i, v := range r {
		ss[i] = string(v.op)
	}
	return strings.Join(ss, ", ")
}

type rewireOperation struct {
	c    *campaigns.Changeset
	spec *campaigns.ChangesetSpec
	repo *types.Repo
	op   rewireOperationKind
}

type rewireOperationKind string

const (
	rewireOperationNone          rewireOperationKind = "none"
	rewireOperationCreate        rewireOperationKind = "create"
	rewireOperationUpdate        rewireOperationKind = "update"
	rewireOperationReenqueue     rewireOperationKind = "reenqueue"
	rewireOperationClose         rewireOperationKind = "close"
	rewireOperationUnlink        rewireOperationKind = "unlink"
	rewireOperationDelete        rewireOperationKind = "delete"
	rewireOperationTrack         rewireOperationKind = "track"
	rewireOperationAttachTracked rewireOperationKind = "attach-tracked"
)

type changesetRewirer struct {
	campaign *campaigns.Campaign
	tx       *Store
	rstore   repos.Store

	// These fields are populated by loadAssociations
	changesets          campaigns.Changesets
	newChangesetSpecs   campaigns.ChangesetSpecs
	accessibleReposByID map[api.RepoID]*types.Repo

	// These are populated by indexAssociations
	changesetsByRepoHeadRef    map[repoHeadRef]*campaigns.Changeset
	changesetsByRepoExternalID map[repoExternalID]*campaigns.Changeset
	currentSpecsByChangeset    map[int64]*campaigns.ChangesetSpec
}

func (r *changesetRewirer) Plan(ctx context.Context) (pl rewirerPlan, err error) {
	// First we need to load the associations
	err = r.loadAssociations(ctx)
	if err != nil {
		return pl, err
	}

	// Now we put them into buckets so we can match easily
	err = r.indexAssociations(ctx)
	if err != nil {
		return pl, err
	}

	attachedChangesets := map[int64]bool{}
	for _, spec := range r.newChangesetSpecs {
		// If we don't have access to a repository, we return an error. Why not
		// simply skip the repository? If we skip it, the user can't reapply
		// the same campaign spec, since it's already applied and re-applying
		// would require a new spec.
		repo, ok := r.accessibleReposByID[spec.RepoID]
		if !ok {
			return pl, &db.RepoNotFoundErr{ID: spec.RepoID}
		}

		if err := checkRepoSupported(repo); err != nil {
			return pl, err
		}

		// If we need to track a changeset, we need to find it.
		if spec.Spec.IsImportingExisting() {
			k := repoExternalID{repo: spec.RepoID, externalID: spec.Spec.ExternalID}

			c, ok := r.changesetsByRepoExternalID[k]
			if ok {
				// If it's already attached to the campaign and errored, we re-enqueue it.
				if c.ReconcilerState == campaigns.ReconcilerStateErrored {
					pl = append(pl, rewireOperation{c: c, op: rewireOperationReenqueue, spec: spec, repo: repo})
				} else {
					pl = append(pl, rewireOperation{c: c, op: rewireOperationNone, spec: spec, repo: repo})
				}
				attachedChangesets[c.ID] = true
			} else {
				// If we don't have a changeset attached to the campaign, we need to find or create one with the externalID in that repository.
				existing, err := r.tx.GetChangeset(ctx, GetChangesetOpts{
					RepoID:              repo.ID,
					ExternalID:          k.externalID,
					ExternalServiceType: repo.ExternalRepo.ServiceType,
				})
				if err != nil && err != ErrNoResults {
					return nil, err
				}

				if err != ErrNoResults {
					pl = append(pl, rewireOperation{c: existing, op: rewireOperationAttachTracked, spec: spec, repo: repo})
					// If it's already attached to the campaign, we need to keep it
					// there. And if it's new, we want to attach it:
					attachedChangesets[existing.ID] = true
				} else {
					pl = append(pl, rewireOperation{op: rewireOperationTrack, spec: spec, repo: repo})
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
		c, ok := r.changesetsByRepoHeadRef[k]
		if !ok {
			// No, we don't have a changeset on that branch in this repo.
			// We're going to create one so the changeset reconciler picks it up,
			// creates a commit and pushes it to the branch.
			// Except, of course, if spec.Spec.Published is false, then it doesn't do anything.
			pl = append(pl, rewireOperation{op: rewireOperationCreate, spec: spec, repo: repo})
		} else {
			// But if we already have a changeset in the given repository with
			// the given branch, we need to update it to have the new spec
			// and possibly re-attach it to the campaign:
			pl = append(pl, rewireOperation{c: c, op: rewireOperationUpdate, spec: spec, repo: repo})
			attachedChangesets[c.ID] = true
		}
	}

	// But it's possible that changesets are now detached, like Changeset 3 in
	// the example above.
	// This we need to detach and close.
	for _, c := range r.changesets {
		if _, ok := attachedChangesets[c.ID]; ok {
			continue
		}

		// If we don't have access to a repository, we don't detach nor close the changeset.
		_, ok := r.accessibleReposByID[c.RepoID]
		if !ok {
			continue
		}

		if c.CurrentSpecID != 0 && c.OwnedByCampaignID == r.campaign.ID {
			// If we have a current spec ID and the changeset was created by
			// _this_ campaign that means we should detach and close it.

			// But only if it was created on the code host:
			if c.Published() {
				pl = append(pl, rewireOperation{c: c, op: rewireOperationClose})
			} else {
				// otherwise we simply delete it.
				pl = append(pl, rewireOperation{c: c, op: rewireOperationDelete})
			}
		} else {
			pl = append(pl, rewireOperation{c: c, op: rewireOperationUnlink})
		}
	}
	return pl, nil
}

// Rewire loads the current changesets of the given campaign, the changeset
// specs attached to the new campaign spec and rewires them so that the
// changesets are associated with the correct changeset specs and with the
// campaign.
//
// It also updates the ChangesetIDs on the campaign.
func (r *changesetRewirer) Rewire(ctx context.Context) (err error) {
	pl, err := r.Plan(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Calculated plan is %s\n", pl)
	// Now we have two lists, the current changesets and the new changeset specs:

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

	attachedChangesets := map[int64]bool{}
	for _, s := range pl {
		switch s.op {
		case rewireOperationNone:
			attachedChangesets[s.c.ID] = true
		case rewireOperationUnlink:
			s.c.RemoveCampaignID(r.campaign.ID)
			if err := r.tx.UpdateChangeset(ctx, s.c); err != nil {
				return err
			}
		case rewireOperationClose:
			s.c.Closing = true
			s.c.ReconcilerState = campaigns.ReconcilerStateQueued
			s.c.RemoveCampaignID(r.campaign.ID)
			if err := r.tx.UpdateChangeset(ctx, s.c); err != nil {
				return err
			}
		case rewireOperationDelete:
			if err := r.tx.DeleteChangeset(ctx, s.c.ID); err != nil {
				return err
			}
		case rewireOperationCreate:
			c, err := r.createChangesetForSpec(ctx, s.repo, s.spec)
			if err != nil {
				return err
			}
			attachedChangesets[c.ID] = true
		case rewireOperationUpdate:
			if err := r.updateChangesetToNewSpec(ctx, s.c, s.spec); err != nil {
				return err
			}
			attachedChangesets[s.c.ID] = true
		case rewireOperationReenqueue:
			if err := r.updateAndReenqueue(ctx, s.c); err != nil {
				return err
			}
			attachedChangesets[s.c.ID] = true
		case rewireOperationTrack:
			newChangeset := &campaigns.Changeset{
				RepoID:              s.repo.ID,
				ExternalServiceType: s.repo.ExternalRepo.ServiceType,

				CampaignIDs:     []int64{r.campaign.ID},
				ExternalID:      s.spec.Spec.ExternalID,
				AddedToCampaign: true,
				// Note: no CurrentSpecID, because we merely track this one

				PublicationState: campaigns.ChangesetPublicationStatePublished,

				// Enqueue it so the reconciler syncs it.
				ReconcilerState: campaigns.ReconcilerStateQueued,
				Unsynced:        true,
			}

			if err := r.tx.CreateChangeset(ctx, newChangeset); err != nil {
				return err
			}

			attachedChangesets[newChangeset.ID] = true
		case rewireOperationAttachTracked:
			// We already have a changeset with the given repoID and
			// externalID, so we can track it.
			s.c.AddedToCampaign = true
			s.c.CampaignIDs = append(s.c.CampaignIDs, r.campaign.ID)

			// If it errored, we re-enqueue it.
			if s.c.ReconcilerState == campaigns.ReconcilerStateErrored {
				if err := r.updateAndReenqueue(ctx, s.c); err != nil {
					return err
				}
			} else {
				return r.tx.UpdateChangeset(ctx, s.c)
			}

			attachedChangesets[s.c.ID] = true
		default:
			panic("blah not implemented")
		}
	}

	// We went through all the new changeset specs and either created or
	// updated a changeset.
	// Their IDs are all the IDs of changesets that should be in the campaign:
	r.campaign.ChangesetIDs = []int64{}
	for changesetID := range attachedChangesets {
		r.campaign.ChangesetIDs = append(r.campaign.ChangesetIDs, changesetID)
	}

	return r.tx.UpdateCampaign(ctx, r.campaign)
}

func (r *changesetRewirer) createChangesetForSpec(ctx context.Context, repo *types.Repo, spec *campaigns.ChangesetSpec) (*campaigns.Changeset, error) {
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

	return newChangeset, r.tx.CreateChangeset(ctx, newChangeset)
}

func (r *changesetRewirer) updateChangesetToNewSpec(ctx context.Context, c *campaigns.Changeset, spec *campaigns.ChangesetSpec) error {
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
	return r.updateAndReenqueue(ctx, c)
}

// loadAssociations populates the changesets, newChangesetSpecs and
// accessibleReposByID on changesetRewirer.
func (r *changesetRewirer) loadAssociations(ctx context.Context) (err error) {
	// Load all of the new ChangesetSpecs
	r.newChangesetSpecs, _, err = r.tx.ListChangesetSpecs(ctx, ListChangesetSpecsOpts{
		CampaignSpecID: r.campaign.CampaignSpecID,
	})
	if err != nil {
		return err
	}

	// Load all Changesets attached to this campaign, or owned by this campaign but detached.
	r.changesets, err = r.tx.ListChangesetsAttachedOrOwnedByCampaign(ctx, r.campaign.ID)
	if err != nil {
		return err
	}

	repoIDs := append(r.newChangesetSpecs.RepoIDs(), r.changesets.RepoIDs()...)
	// 🚨 SECURITY: db.Repos.GetRepoIDsSet uses the authzFilter under the hood and
	// filters out repositories that the user doesn't have access to.
	r.accessibleReposByID, err = db.Repos.GetReposSetByIDs(ctx, repoIDs...)
	return err
}

func (r *changesetRewirer) indexAssociations(ctx context.Context) (err error) {
	r.changesetsByRepoHeadRef = map[repoHeadRef]*campaigns.Changeset{}
	r.changesetsByRepoExternalID = map[repoExternalID]*campaigns.Changeset{}
	r.currentSpecsByChangeset = map[int64]*campaigns.ChangesetSpec{}

	for _, c := range r.changesets {
		// This is an n+1
		s, err := r.tx.GetChangesetSpecByID(ctx, c.CurrentSpecID)
		if err != nil {
			return err
		}
		r.currentSpecsByChangeset[c.ID] = s

		if c.ExternalID != "" {
			k := repoExternalID{repo: c.RepoID, externalID: c.ExternalID}
			r.changesetsByRepoExternalID[k] = c

			// If it has an externalID but no CurrentSpecID, it is a tracked
			// changeset, and we're done and don't need to match it by HeadRef
			if c.CurrentSpecID == 0 {
				continue
			}
		}

		k := repoHeadRef{repo: c.RepoID}
		if c.ExternalBranch != "" {
			k.headRef = git.EnsureRefPrefix(c.ExternalBranch)
			r.changesetsByRepoHeadRef[k] = c
			continue
		}

		// If we don't have an ExternalBranch, the changeset hasn't been
		// published yet (or hasn't been synced yet).
		if c.CurrentSpecID != 0 {
			// If we're here, the changeset doesn't have an external branch
			//
			// So we use the spec to get the branch where we _would_ push
			// the commit.

			k.headRef = git.EnsureRefPrefix(s.Spec.HeadRef)
			r.changesetsByRepoHeadRef[k] = c
		}
	}

	return nil
}

func (r *changesetRewirer) updateAndReenqueue(ctx context.Context, ch *campaigns.Changeset) error {
	ch.ResetQueued()
	return r.tx.UpdateChangeset(ctx, ch)
}
