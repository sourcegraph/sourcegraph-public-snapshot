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
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
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

// CreatePatchSetFromPatches creates a PatchSet and its associated Patches from patches
// computed by the caller. There is no diff execution or computation performed during creation of
// the Patches in this case (unlike when using Runner to create a PatchSet from a
// specification).
func (s *Service) CreatePatchSetFromPatches(ctx context.Context, patches []*campaigns.Patch, userID int32) (*campaigns.PatchSet, error) {
	if userID == 0 {
		return nil, backend.ErrNotAuthenticated
	}

	repoIDs := make([]api.RepoID, len(patches))
	for i, patch := range patches {
		repoIDs[i] = patch.RepoID
	}
	// 🚨 SECURITY: We use db.Repos.GetByIDs to check for which the user has access.
	repos, err := db.Repos.GetByIDs(ctx, repoIDs...)
	if err != nil {
		return nil, err
	}
	reposByID := make(map[api.RepoID]*types.Repo, len(patches))
	for _, repo := range repos {
		reposByID[repo.ID] = repo
	}

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Done(&err)

	patchSet := &campaigns.PatchSet{UserID: userID}
	err = tx.CreatePatchSet(ctx, patchSet)
	if err != nil {
		return nil, err
	}

	for _, patch := range patches {
		if _, ok := reposByID[patch.RepoID]; !ok {
			return nil, &db.RepoNotFoundErr{ID: patch.RepoID}
		}

		patch.PatchSetID = patchSet.ID
		if err := tx.CreatePatch(ctx, patch); err != nil {
			return nil, err
		}
	}

	return patchSet, nil
}

// CreateCampaign creates the Campaign. When a PatchSetID is set on the
// Campaign it validates that the PatchSet contains Patches.
func (s *Service) CreateCampaign(ctx context.Context, c *campaigns.Campaign) error {
	var err error
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
	defer tx.Done(&err)

	if c.PatchSetID != 0 {
		_, err = tx.GetCampaign(ctx, GetCampaignOpts{PatchSetID: c.PatchSetID})
		if err != nil && err != ErrNoResults {
			return err
		}
		if err != ErrNoResults {
			err = ErrPatchSetDuplicate
			return err
		}
	}

	c.CreatedAt = s.clock()
	c.UpdatedAt = c.CreatedAt

	err = tx.CreateCampaign(ctx, c)
	if err != nil {
		return err
	}

	if c.PatchSetID == 0 {
		return nil
	}
	err = validateCampaignBranch(c.Branch)
	if err != nil {
		return err
	}
	// Validate we don't have an empty patchset.
	var patchCount int64
	patchCount, err = tx.CountPatches(ctx, CountPatchesOpts{PatchSetID: c.PatchSetID, OnlyWithDiff: true, OnlyUnpublishedInCampaign: c.ID})
	if err != nil {
		return err
	}
	if patchCount == 0 {
		err = ErrNoPatches
		return err
	}

	return nil
}

// ErrNoPatches is returned by CreateCampaign or UpdateCampaign if a
// PatchSetID was specified but the PatchSet does not have any
// (finished) Patches.
var ErrNoPatches = errors.New("cannot create or update a Campaign without any changesets")

// ErrCloseProcessingCampaign is returned by CloseCampaign if the Campaign has
// been published at the time of closing but its ChangesetJobs have not
// finished execution.
var ErrCloseProcessingCampaign = errors.New("cannot close a Campaign while changesets are being created on codehosts")

// ErrUnsupportedCodehost is returned by EnqueueChangesetJobForPatch if the target repo of a patch is an unsupported repo.
var ErrUnsupportedCodehost = errors.New("cannot publish patch for unsupported codehost")

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
		defer tx.Done(&err)

		campaign, err = tx.GetCampaign(ctx, GetCampaignOpts{ID: id})
		if err != nil {
			return errors.Wrap(err, "getting campaign")
		}

		if err := backend.CheckSiteAdminOrSameUser(ctx, campaign.AuthorID); err != nil {
			return err
		}

		processing, err := campaignIsProcessing(ctx, tx, id)
		if err != nil {
			return err
		}
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
func (s *Service) DeleteCampaign(ctx context.Context, id int64, closeChangesets bool) (err error) {
	traceTitle := fmt.Sprintf("campaign: %d, closeChangesets: %t", id, closeChangesets)
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

	transaction := func() (cs []*campaigns.Changeset, err error) {
		tx, err := s.store.Transact(ctx)
		if err != nil {
			return nil, err
		}
		defer tx.Done(&err)

		processing, err := campaignIsProcessing(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		if processing {
			return nil, ErrDeleteProcessingCampaign
		}

		// If we don't have to close the changesets, we can simply delete the
		// Campaign and return. The triggers in the database will remove the
		// campaign's ID from the changesets' CampaignIDs.
		if !closeChangesets {
			return nil, tx.DeleteCampaign(ctx, id)
		}

		// First load the Changesets with the given campaignID, before deleting
		// the campaign would remove the association.
		cs, _, err = tx.ListChangesets(ctx, ListChangesetsOpts{
			CampaignID: id,
			Limit:      -1,
		})
		if err != nil {
			return nil, err
		}

		// Remove the association manually, since we'll update the Changesets in
		// the database, after closing them and we can't update them with an
		// invalid CampaignID.
		for _, c := range cs {
			c.RemoveCampaignID(id)
		}

		return cs, tx.DeleteCampaign(ctx, id)
	}

	cs, err := transaction()
	if err != nil {
		return err
	}

	go func() {
		ctx := trace.ContextWithTrace(context.Background(), tr)
		err := s.CloseOpenChangesets(ctx, cs)
		if err != nil {
			log15.Error("CloseCampaignChangesets", "err", err)
		}
	}()

	return nil
}

// CloseOpenChangesets closes the given Changesets on their respective codehosts and syncs them.
func (s *Service) CloseOpenChangesets(ctx context.Context, cs campaigns.Changesets) (err error) {
	cs = cs.Filter(func(c *campaigns.Changeset) bool {
		return c.ExternalState == campaigns.ChangesetStateOpen
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

// AddChangesetsToCampaign adds the given changeset IDs to the given campaign's
// ChangesetIDs and the campaign ID to the CampaignIDs of each changeset.
// It updates the campaign and the changesets in the database.
// If one of the changeset IDs is invalid an error is returned.
func (s *Service) AddChangesetsToCampaign(ctx context.Context, campaignID int64, changesetIDs []int64) (campaign *campaigns.Campaign, err error) {
	traceTitle := fmt.Sprintf("campaign: %d, changesets: %v", campaignID, changesetIDs)
	tr, ctx := trace.New(ctx, "service.AddChangesetsToCampaign", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Done(&err)

	campaign, err = tx.GetCampaign(ctx, GetCampaignOpts{ID: campaignID})
	if err != nil {
		return nil, err
	}

	if err := backend.CheckSiteAdminOrSameUser(ctx, campaign.AuthorID); err != nil {
		return nil, err
	}

	set := map[int64]struct{}{}
	for _, id := range changesetIDs {
		set[id] = struct{}{}
	}

	changesets, _, err := tx.ListChangesets(ctx, ListChangesetsOpts{IDs: changesetIDs, Limit: -1})
	if err != nil {
		return nil, err
	}

	accessibleRepos, err := accessibleRepos(ctx, changesets.RepoIDs())
	if err != nil {
		return nil, err
	}

	for _, c := range changesets {
		if _, ok := accessibleRepos[c.RepoID]; !ok {
			return nil, &db.RepoNotFoundErr{ID: c.RepoID}
		}

		delete(set, c.ID)
		c.CampaignIDs = append(c.CampaignIDs, campaign.ID)
		c.AddedToCampaign = true
	}

	if len(set) > 0 {
		return nil, errors.Errorf("changesets %v not found", set)
	}

	if err = tx.UpdateChangesets(ctx, changesets...); err != nil {
		return nil, err
	}

	campaign.ChangesetIDs = append(campaign.ChangesetIDs, changesetIDs...)
	if err = tx.UpdateCampaign(ctx, campaign); err != nil {
		return nil, err
	}

	return campaign, nil
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

// RetryPublishCampaign resets the failed (!) ChangesetJobs for the given
// campaign, which causes them to be re-run in the background.
func (s *Service) RetryPublishCampaign(ctx context.Context, id int64) (campaign *campaigns.Campaign, err error) {
	traceTitle := fmt.Sprintf("campaign: %d", id)
	tr, ctx := trace.New(ctx, "service.RetryPublishCampaign", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	campaign, err = s.store.GetCampaign(ctx, GetCampaignOpts{ID: id})
	if err != nil {
		return nil, errors.Wrap(err, "getting campaign")
	}

	if err := backend.CheckSiteAdminOrSameUser(ctx, campaign.AuthorID); err != nil {
		return nil, err
	}

	patches, _, err := s.store.ListPatches(ctx, ListPatchesOpts{
		PatchSetID: campaign.PatchSetID,
		Limit:      -1,
		NoDiff:     true,
	})
	if err != nil {
		return nil, err
	}

	repoIDs, err := s.store.GetRepoIDsForFailedChangesetJobs(ctx, campaign.ID)
	if err != nil {
		return nil, err
	}

	accessibleRepos, err := accessibleRepos(ctx, repoIDs)
	if err != nil {
		return nil, err
	}

	var resetPatchIDs []int64
	for _, p := range patches {
		if _, ok := accessibleRepos[p.RepoID]; !ok {
			continue
		}

		resetPatchIDs = append(resetPatchIDs, p.ID)
	}

	err = s.store.ResetChangesetJobs(ctx, ResetChangesetJobsOpts{
		CampaignID: id,
		OnlyFailed: true,
		PatchIDs:   resetPatchIDs,
	})
	if err != nil {
		return nil, errors.Wrap(err, "resetting failed changeset jobs")
	}

	return campaign, nil
}

// EnqueueChangesetJobs enqueues a ChangesetJob for each Patch associated with
// the PatchSet in the given Campaign, creating it if necessary. The Patch has
// to belong to a PatchSet
func (s *Service) EnqueueChangesetJobs(ctx context.Context, campaignID int64) (_ *campaigns.Campaign, err error) {
	traceTitle := fmt.Sprintf("campaign: %d", campaignID)
	tr, ctx := trace.New(ctx, "service.EnqueueChangesetJobs", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	campaign, err := s.store.GetCampaign(ctx, GetCampaignOpts{ID: campaignID})
	if err != nil {
		return nil, err
	}

	err = backend.CheckSiteAdminOrSameUser(ctx, campaign.AuthorID)
	if err != nil {
		return nil, err
	}

	if campaign.PatchSetID == 0 {
		return nil, ErrNoPatches
	}

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Done(&err)

	patches, _, err := tx.ListPatches(ctx, ListPatchesOpts{
		PatchSetID:              campaign.PatchSetID,
		Limit:                   -1,
		OnlyWithDiff:            true,
		OnlyWithoutChangesetJob: campaign.ID,
		NoDiff:                  true,
	})
	if err != nil {
		return nil, err
	}

	repoIDs := make([]api.RepoID, 0, len(patches))
	for _, p := range patches {
		repoIDs = append(repoIDs, p.RepoID)
	}

	accessibleRepos, err := accessibleRepos(ctx, repoIDs)
	if err != nil {
		return nil, err
	}

	existingJobs, _, err := tx.ListChangesetJobs(ctx, ListChangesetJobsOpts{
		Limit:      -1,
		CampaignID: campaign.ID,
	})
	if err != nil {
		return nil, err
	}

	jobsByPatchID := make(map[int64]*campaigns.ChangesetJob, len(existingJobs))
	for _, j := range existingJobs {
		jobsByPatchID[j.PatchID] = j
	}

	for _, p := range patches {
		if _, ok := jobsByPatchID[p.ID]; ok {
			continue
		}

		r, ok := accessibleRepos[p.RepoID]
		if !ok {
			continue
		}

		// Check if the repo is on a supported codehost.
		if !campaigns.IsRepoSupported(&r.ExternalRepo) {
			continue
		}

		j := &campaigns.ChangesetJob{CampaignID: campaign.ID, PatchID: p.ID}
		if err := tx.CreateChangesetJob(ctx, j); err != nil {
			return nil, err
		}
	}

	return campaign, nil
}

// EnqueueChangesetJobForPatch queues a ChangesetJob for the Patch with the
// given ID, creating it if necessary. The Patch has to belong to a PatchSet
// that was attached to a Campaign.
func (s *Service) EnqueueChangesetJobForPatch(ctx context.Context, patchID int64) (err error) {
	traceTitle := fmt.Sprintf("patch: %d", patchID)
	tr, ctx := trace.New(ctx, "service.EnqueueChangesetJobForPatch", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	job, err := s.store.GetPatch(ctx, GetPatchOpts{ID: patchID})
	if err != nil {
		return err
	}
	repo, err := db.Repos.Get(ctx, job.RepoID)
	if err != nil {
		return err
	}
	if !campaigns.IsRepoSupported(&repo.ExternalRepo) {
		return ErrUnsupportedCodehost
	}

	campaign, err := s.store.GetCampaign(ctx, GetCampaignOpts{PatchSetID: job.PatchSetID})
	if err != nil {
		return err
	}

	err = backend.CheckSiteAdminOrSameUser(ctx, campaign.AuthorID)
	if err != nil {
		return err
	}

	// 🚨 SECURITY: We use db.Repos.Get to check whether the user has access to
	// the repository or not.
	if _, err = db.Repos.Get(ctx, job.RepoID); err != nil {
		return err
	}

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer tx.Done(&err)

	existing, err := tx.GetChangesetJob(ctx, GetChangesetJobOpts{
		CampaignID: campaign.ID,
		PatchID:    job.ID,
	})
	if err != nil && err != ErrNoResults {
		return err
	}
	if existing != nil {
		// An extant changeset job that failed should be reset so
		// ProcessPendingChangesetJobs can try to publish it again.
		if existing.UnsuccessfullyCompleted() {
			existing.Reset()
			return tx.UpdateChangesetJob(ctx, existing)
		}

		return nil
	}

	changesetJob := &campaigns.ChangesetJob{
		CampaignID: campaign.ID,
		PatchID:    job.ID,
	}
	return tx.CreateChangesetJob(ctx, changesetJob)
}

// GetCampaignStatus returns the BackgroundProcessStatus for the given campaign.
func (s *Service) GetCampaignStatus(ctx context.Context, c *campaigns.Campaign) (status *campaigns.BackgroundProcessStatus, err error) {
	traceTitle := fmt.Sprintf("campaign: %d", c.ID)
	tr, ctx := trace.New(ctx, "service.GetCampaignStatus", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	canAdmin, err := hasCampaignAdminPermissions(ctx, c)
	if err != nil {
		return nil, err
	}

	if !canAdmin {
		// If the user doesn't have admin permissions for this campaign, we
		// don't need to filter out specific errors, but can simply exclude
		// _all_ errors.
		return s.store.GetCampaignStatus(ctx, GetCampaignStatusOpts{
			ID:            c.ID,
			ExcludeErrors: true,
		})
	}

	// We need to filter out error messages the user is not allowed to see,
	// because they don't have permissions to access the repository associated
	// with a given patch/changesetJob.

	// First we load the repo IDs of the failed changesetJobs
	repoIDs, err := s.store.GetRepoIDsForFailedChangesetJobs(ctx, c.ID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: We use db.Repos.GetByIDs to filter out repositories the
	// user doesn't have access to.
	accessibleRepos, err := db.Repos.GetByIDs(ctx, repoIDs...)
	if err != nil {
		return nil, err
	}

	accessibleRepoIDs := make(map[api.RepoID]struct{}, len(accessibleRepos))
	for _, r := range accessibleRepos {
		accessibleRepoIDs[r.ID] = struct{}{}
	}

	// We now check which repositories in `repoIDs` are not in `accessibleRepoIDs`.
	// We have to filter the error messages associated with those out.
	excludedRepos := make([]api.RepoID, 0, len(accessibleRepoIDs))
	for _, id := range repoIDs {
		if _, ok := accessibleRepoIDs[id]; !ok {
			excludedRepos = append(excludedRepos, id)
		}
	}

	return s.store.GetCampaignStatus(ctx, GetCampaignStatusOpts{
		ID:                   c.ID,
		ExcludeErrorsInRepos: excludedRepos,
	})
}

// ErrUpdateProcessingCampaign is returned by UpdateCampaign if the Campaign
// has been published at the time of update but its ChangesetJobs have not
// finished execution.
var ErrUpdateProcessingCampaign = errors.New("cannot update a Campaign while changesets are being created on codehosts")

type UpdateCampaignArgs struct {
	Campaign    int64
	Name        *string
	Description *string
	Branch      *string
	PatchSet    *int64
}

// ErrCampaignNameBlank is returned by CreateCampaign or UpdateCampaign if the
// specified Campaign name is blank.
var ErrCampaignNameBlank = errors.New("Campaign title cannot be blank")

// ErrCampaignBranchBlank is returned by CreateCampaign or UpdateCampaign if the specified Campaign's
// branch is blank. This is only enforced when creating published campaigns with a patch set.
var ErrCampaignBranchBlank = errors.New("Campaign branch cannot be blank")

// ErrCampaignBranchInvalid is returned by CreateCampaign or UpdateCampaign if the specified Campaign's
// branch is invalid. This is only enforced when creating published campaigns with a patch set.
var ErrCampaignBranchInvalid = errors.New("Campaign branch is invalid")

// ErrPublishedCampaignBranchChange is returned by UpdateCampaign if there is an
// attempt to change the branch of a published campaign with a patch set (or a campaign with individually published changesets).
var ErrPublishedCampaignBranchChange = errors.New("Published campaign branch cannot be changed")

// ErrPatchSetDuplicate is return by CreateCampaign or UpdateCampaign if the
// specified patch set is already attached to another campaign.
var ErrPatchSetDuplicate = errors.New("Campaign cannot use the same patch set as another campaign")

// ErrUpdateClosedCampaign is returned by UpdateCampaign if the Campaign
// has been closed.
var ErrUpdateClosedCampaign = errors.New("cannot update a closed Campaign")

// UpdateCampaign updates the Campaign with the given arguments.
func (s *Service) UpdateCampaign(ctx context.Context, args UpdateCampaignArgs) (campaign *campaigns.Campaign, detachedChangesets []*campaigns.Changeset, err error) {
	traceTitle := fmt.Sprintf("campaign: %d", args.Campaign)
	tr, ctx := trace.New(ctx, "service.UpdateCampaign", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return nil, nil, err
	}

	defer tx.Done(&err)

	campaign, err = tx.GetCampaign(ctx, GetCampaignOpts{ID: args.Campaign})
	if err != nil {
		return nil, nil, errors.Wrap(err, "getting campaign")
	}

	err = backend.CheckSiteAdminOrSameUser(ctx, campaign.AuthorID)
	if err != nil {
		return nil, nil, err
	}

	if !campaign.ClosedAt.IsZero() {
		return nil, nil, ErrUpdateClosedCampaign
	}

	var updateAttributes, updatePatchSetID, updateBranch bool

	if args.Name != nil && campaign.Name != *args.Name {
		if *args.Name == "" {
			return nil, nil, ErrCampaignNameBlank
		}

		campaign.Name = *args.Name
		updateAttributes = true
	}

	if args.Description != nil && campaign.Description != *args.Description {
		campaign.Description = *args.Description
		updateAttributes = true
	}

	oldPatchSetID := campaign.PatchSetID
	if args.PatchSet != nil && oldPatchSetID != *args.PatchSet {
		// Check there is no other campaign attached to the args.PatchSet.
		_, err = tx.GetCampaign(ctx, GetCampaignOpts{PatchSetID: *args.PatchSet})
		if err != nil && err != ErrNoResults {
			return nil, nil, err
		}
		if err != ErrNoResults {
			return nil, nil, ErrPatchSetDuplicate
		}

		campaign.PatchSetID = *args.PatchSet
		updatePatchSetID = true
	}

	if args.Branch != nil && campaign.Branch != *args.Branch {
		if err := validateCampaignBranch(*args.Branch); err != nil {
			return nil, nil, err
		}

		campaign.Branch = *args.Branch
		updateBranch = true
	}

	if !updateAttributes && !updatePatchSetID && !updateBranch {
		return campaign, nil, nil
	}

	// If we're adding a patchset, we're done, since we don't have to compute a
	// diff and don't have to update changesets.
	if oldPatchSetID == 0 {
		return campaign, nil, tx.UpdateCampaign(ctx, campaign)
	}

	status, err := tx.GetCampaignStatus(ctx, GetCampaignStatusOpts{ID: campaign.ID})
	if err != nil {
		return nil, nil, err
	}

	// We check whether we have any ChangesetJobs currently being processed.
	// If yes, we don't allow the update.
	if status.Processing() {
		return nil, nil, ErrUpdateProcessingCampaign
	}

	// If they're not processing, we can assume that they've been published or
	// failed.
	// How many patches do we have that are not published or failed to publish?
	unpublished, err := tx.CountPatches(ctx, CountPatchesOpts{
		PatchSetID:              oldPatchSetID,
		OnlyWithoutChangesetJob: campaign.ID,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "getting unpublished patches count")
	}
	allPublished := unpublished == 0
	partiallyPublished := !allPublished && status.Total != 0

	if oldPatchSetID != 0 && updateBranch {
		if allPublished || partiallyPublished {
			return nil, nil, ErrPublishedCampaignBranchChange
		}
	}

	if !allPublished && !partiallyPublished {
		// If no ChangesetJobs have been created yet, we can simply update the
		// attributes on the Campaign because no ChangesetJobs have been
		// created yet that need updating.
		return campaign, nil, tx.UpdateCampaign(ctx, campaign)
	}

	// If we do have to update ChangesetJobs/Changesets, here's a fast path: if
	// we don't update the PatchSet, we don't need to rewire ChangesetJobs,
	// but only update name/description if they changed.
	if (!updatePatchSetID || oldPatchSetID == 0) && updateAttributes {
		err := tx.UpdateCampaign(ctx, campaign)
		if err != nil {
			return campaign, nil, err
		}

		return campaign, nil, resetAccessibleChangesetJobs(ctx, tx, campaign)
	}

	diff, err := computeCampaignUpdateDiff(ctx, tx, campaign, oldPatchSetID, updateAttributes)
	if err != nil {
		return nil, nil, err
	}

	for _, c := range diff.Update {
		err := tx.UpdateChangesetJob(ctx, c)
		if err != nil {
			return nil, nil, errors.Wrap(err, "updating changeset job")
		}
	}

	// When we're doing a partial update and only update the Changesets that
	// have already been published, we don't want to create new ChangesetJobs,
	// since they would be processed and publish the other Changesets.
	if !partiallyPublished {
		for _, c := range diff.Create {
			err := tx.CreateChangesetJob(ctx, c)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	var changesetsToCloseAndDetach []int64
	for _, j := range diff.Delete {
		err := tx.DeleteChangesetJob(ctx, j.ID)
		if err != nil {
			return nil, nil, err
		}
		changesetsToCloseAndDetach = append(changesetsToCloseAndDetach, j.ChangesetID)
	}

	if len(changesetsToCloseAndDetach) == 0 {
		return campaign, nil, tx.UpdateCampaign(ctx, campaign)
	}

	changesets, _, err := tx.ListChangesets(ctx, ListChangesetsOpts{
		IDs: changesetsToCloseAndDetach,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "listing changesets to close and detach")
	}

	for _, c := range changesets {
		c.RemoveCampaignID(campaign.ID)
		campaign.RemoveChangesetID(c.ID)
	}

	if err = tx.UpdateChangesets(ctx, changesets...); err != nil {
		return nil, nil, errors.Wrap(err, "updating changesets")
	}

	return campaign, changesets, tx.UpdateCampaign(ctx, campaign)
}

func validateCampaignBranch(branch string) error {
	if branch == "" {
		return ErrCampaignBranchBlank
	}
	if !git.ValidateBranchName(branch) {
		return ErrCampaignBranchInvalid
	}
	return nil
}

type campaignUpdateDiff struct {
	Delete []*campaigns.ChangesetJob
	Update []*campaigns.ChangesetJob
	Create []*campaigns.ChangesetJob
}

// repoGroup is a group of entities involved in a Campaign that are associated
// with the same repository.
type repoGroup struct {
	changesetJob *campaigns.ChangesetJob
	patch        *campaigns.Patch
	newPatch     *campaigns.Patch
	changeset    *campaigns.Changeset
}

func resetAccessibleChangesetJobs(ctx context.Context, tx *Store, campaign *campaigns.Campaign) error {
	patches, _, err := tx.ListPatches(ctx, ListPatchesOpts{
		PatchSetID:   campaign.PatchSetID,
		Limit:        -1,
		OnlyWithDiff: true,
		NoDiff:       true,
	})
	if err != nil {
		return errors.Wrap(err, "listing patches")
	}

	repoIDs := make([]api.RepoID, 0, len(patches))
	for _, p := range patches {
		repoIDs = append(repoIDs, p.RepoID)
	}

	accessibleRepos, err := accessibleRepos(ctx, repoIDs)
	if err != nil {
		return err
	}

	var resetPatchIDs []int64
	for _, p := range patches {
		if _, ok := accessibleRepos[p.RepoID]; !ok {
			continue
		}

		resetPatchIDs = append(resetPatchIDs, p.ID)
	}

	return tx.ResetChangesetJobs(ctx, ResetChangesetJobsOpts{
		CampaignID: campaign.ID,
		PatchIDs:   resetPatchIDs,
	})
}

func computeCampaignUpdateDiff(
	ctx context.Context,
	tx *Store,
	campaign *campaigns.Campaign,
	oldPatchSetID int64,
	updateAttributes bool,
) (*campaignUpdateDiff, error) {
	diff := &campaignUpdateDiff{}

	changesetJobs, _, err := tx.ListChangesetJobs(ctx, ListChangesetJobsOpts{
		CampaignID: campaign.ID,
		Limit:      -1,
	})
	if err != nil {
		return nil, errors.Wrap(err, "listing changesets jobs")
	}

	changesets, _, err := tx.ListChangesets(ctx, ListChangesetsOpts{
		CampaignID: campaign.ID,
		Limit:      -1,
	})
	if err != nil {
		return nil, errors.Wrap(err, "listing changesets")
	}

	// We need OnlyWithDiff because we don't create ChangesetJobs for others.
	patches, _, err := tx.ListPatches(ctx, ListPatchesOpts{
		PatchSetID:   oldPatchSetID,
		Limit:        -1,
		OnlyWithDiff: true,
	})
	if err != nil {
		return nil, errors.Wrap(err, "listing patches")
	}

	newPatches, _, err := tx.ListPatches(ctx, ListPatchesOpts{
		PatchSetID:   campaign.PatchSetID,
		Limit:        -1,
		OnlyWithDiff: true,
	})
	if err != nil {
		return nil, errors.Wrap(err, "listing new patches")
	}

	if len(newPatches) == 0 {
		return nil, ErrNoPatches
	}

	// We need to determine which current ChangesetJobs we want to keep and
	// which ones we want to delete.
	// We can find out which ones we want to keep by looking at the RepoID of
	// their Patches.

	byRepoID, err := mergeByRepoID(changesetJobs, patches, changesets)
	if err != nil {
		return nil, err
	}

	// repoIDs is a unique list of repositories involved in this update
	// operation. We use it to query which repositories the user has access to.
	repoIDs := make([]api.RepoID, 0, len(byRepoID))
	for repoID := range byRepoID {
		repoIDs = append(repoIDs, repoID)
	}
	for _, p := range newPatches {
		if _, ok := byRepoID[p.RepoID]; !ok {
			repoIDs = append(repoIDs, p.RepoID)
		}
	}
	// 🚨 SECURITY: Check which repositories the user has access to. If the
	// user doesn't have access, don't create/delete/update anything.
	accessibleRepos, err := accessibleRepos(ctx, repoIDs)
	if err != nil {
		return nil, err
	}

	for _, j := range newPatches {
		// If the user is missing permissions for any of the new patches, we
		// return an error instead of skipping patches, so we don't end up with
		// an unfixable state (i.e. unpublished patch + changeset for same
		// repo).
		repo, ok := accessibleRepos[j.RepoID]
		if !ok {
			return nil, &db.RepoNotFoundErr{ID: j.RepoID}
		}
		if !campaigns.IsRepoSupported(&repo.ExternalRepo) {
			continue
		}

		if group, ok := byRepoID[j.RepoID]; ok {
			group.newPatch = j
		} else {
			// If we have new Patches that don't match an existing
			// ChangesetJob we need to create new ChangesetJobs.
			diff.Create = append(diff.Create, &campaigns.ChangesetJob{
				CampaignID: campaign.ID,
				PatchID:    j.ID,
			})
		}
	}

	for repoID, group := range byRepoID {
		// If the user is lacking permissions for this repository we don't
		// delete/update the changeset.
		if _, ok := accessibleRepos[repoID]; !ok {
			continue
		}

		// Either we _don't_ have a matching _new_ Patch, then we delete
		// the ChangesetJob and detach & close Changeset.
		if group.newPatch == nil {
			// But if we have already created a changeset for this repo and its
			// merged or closed, we keep it.
			if group.changeset != nil {
				s := group.changeset.ExternalState
				if s == campaigns.ChangesetStateMerged || s == campaigns.ChangesetStateClosed {
					continue
				}
			}
			diff.Delete = append(diff.Delete, group.changesetJob)
			continue
		}

		// Or we have a matching _new_ Patch, then we keep the
		// ChangesetJob around, but need to rewire it.
		group.changesetJob.PatchID = group.newPatch.ID

		//  And, if the {Diff,Rev,BaseRef,Description} are different, we  need to
		// update the Changeset on the codehost...
		if updateAttributes || patchesDiffer(group.newPatch, group.patch) {
			// .. but if we already have a Changeset and that is merged, we
			// don't want to update it...
			if group.changeset != nil {
				s := group.changeset.ExternalState
				if s == campaigns.ChangesetStateMerged || s == campaigns.ChangesetStateClosed {
					// Note: in the future we want to create a new ChangesetJob here.
					continue
				}
			}

			// if we do want to update it, we _reset_ the ChangesetJob, so it
			// gets run again when RunChangesetJobs is called after
			// UpdateCampaign.
			group.changesetJob.Reset()
		}

		diff.Update = append(diff.Update, group.changesetJob)
	}

	return diff, nil
}

// patchesDiffer returns true if the Patches differ in a way that
// requires updating the Changeset on the codehost.
func patchesDiffer(a, b *campaigns.Patch) bool {
	return a.Diff != b.Diff ||
		a.Rev != b.Rev ||
		a.BaseRef != b.BaseRef
}

func isOutdated(c *repos.Changeset) (bool, error) {
	currentTitle, err := c.Changeset.Title()
	if err != nil {
		return false, err
	}

	if currentTitle != c.Title {
		return true, nil
	}

	currentBody, err := c.Changeset.Body()
	if err != nil {
		return false, err
	}

	if currentBody != c.Body {
		return true, nil
	}

	currentBaseRef, err := c.Changeset.BaseRef()
	if err != nil {
		return false, err
	}

	if git.EnsureRefPrefix(currentBaseRef) != git.EnsureRefPrefix(c.BaseRef) {
		return true, nil
	}

	return false, nil
}

func mergeByRepoID(
	chs []*campaigns.ChangesetJob,
	cas []*campaigns.Patch,
	cs []*campaigns.Changeset,
) (map[api.RepoID]*repoGroup, error) {
	jobs := make(map[api.RepoID]*repoGroup, len(chs))

	byID := make(map[int64]*campaigns.Patch, len(cas))
	for _, j := range cas {
		byID[j.ID] = j
	}

	for _, j := range chs {
		caj, ok := byID[j.PatchID]
		if !ok {
			return nil, fmt.Errorf("Patch with ID %d cannot be found for ChangesetJob %d", j.PatchID, j.ID)
		}
		jobs[caj.RepoID] = &repoGroup{changesetJob: j, patch: caj}
	}

	for _, c := range cs {
		if j, ok := jobs[c.RepoID]; ok {
			j.changeset = c
		}
	}

	return jobs, nil
}

func campaignIsProcessing(ctx context.Context, store *Store, campaign int64) (bool, error) {
	status, err := store.GetCampaignStatus(ctx, GetCampaignStatusOpts{ID: campaign})
	if err != nil {
		return false, err
	}
	return status.Processing(), nil
}

// hasCampaignAdminPermissions returns true when the actor in the given context
// is either a site-admin or the author of the given campaign.
func hasCampaignAdminPermissions(ctx context.Context, c *campaigns.Campaign) (bool, error) {
	// 🚨 SECURITY: Only site admins or the authors of a campaign have campaign admin rights.
	if err := backend.CheckSiteAdminOrSameUser(ctx, c.AuthorID); err != nil {
		if _, ok := err.(*backend.InsufficientAuthorizationError); ok {
			return false, nil
		}

		return false, err
	}
	return true, nil
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
