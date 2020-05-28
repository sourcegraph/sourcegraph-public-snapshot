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
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
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
	// Look up all repositories
	reposStore := repos.NewDBStore(s.store.DB(), sql.TxOptions{})
	repoIDs := make([]api.RepoID, len(patches))
	for i, patch := range patches {
		repoIDs[i] = api.RepoID(patch.RepoID)
	}
	allRepos, err := reposStore.ListRepos(ctx, repos.StoreListReposArgs{IDs: repoIDs})
	if err != nil {
		return nil, err
	}
	reposByID := make(map[api.RepoID]*repos.Repo, len(patches))
	for _, repo := range allRepos {
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
		repo := reposByID[patch.RepoID]
		if repo == nil {
			return nil, fmt.Errorf("repository ID %d not found", patch.RepoID)
		}
		if !campaigns.IsRepoSupported(&repo.ExternalRepo) {
			continue
		}

		patch.PatchSetID = patchSet.ID
		if err := tx.CreatePatch(ctx, patch); err != nil {
			return nil, err
		}
	}

	return patchSet, nil
}

// CreateCampaign creates the Campaign. When a PatchSetID is set on the
// Campaign and the Campaign is not created as a draft, it calls
// CreateChangesetJobs inside the same transaction in which it creates the
// Campaign.
func (s *Service) CreateCampaign(ctx context.Context, c *campaigns.Campaign, draft bool) error {
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
		_, err := tx.GetCampaign(ctx, GetCampaignOpts{PatchSetID: c.PatchSetID})
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

	if err = tx.CreateCampaign(ctx, c); err != nil {
		return err
	}

	if c.PatchSetID != 0 {
		if err := validateCampaignBranch(c.Branch); err != nil {
			return err
		}
	}

	if c.PatchSetID == 0 || draft {
		return nil
	}

	err = s.createChangesetJobsWithStore(ctx, tx, c)
	return err
}

// ErrNoPatches is returned by CreateCampaign or UpdateCampaign if a
// PatchSetID was specified but the PatchSet does not have any
// (finished) Patches.
var ErrNoPatches = errors.New("cannot create or update a Campaign without any changesets")

func (s *Service) createChangesetJobsWithStore(ctx context.Context, store *Store, c *campaigns.Campaign) error {
	if c.PatchSetID == 0 {
		return errors.New("cannot create changesets for campaign with no patch set")
	}

	jobs, _, err := store.ListPatches(ctx, ListPatchesOpts{
		PatchSetID:                c.PatchSetID,
		Limit:                     -1,
		OnlyWithDiff:              true,
		OnlyUnpublishedInCampaign: c.ID,
	})
	if err != nil {
		return err
	}

	if len(jobs) == 0 {
		return ErrNoPatches
	}

	for _, job := range jobs {
		changesetJob := &campaigns.ChangesetJob{
			CampaignID: c.ID,
			PatchID:    job.ID,
		}
		err = store.CreateChangesetJob(ctx, changesetJob)
		if err != nil {
			return err
		}
	}

	return nil
}

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

// PublishCampaign publishes the Campaign with the given ID
// by turning the Patches attached to the PatchSet of
// the Campaign into ChangesetJobs and enqueuing them
func (s *Service) PublishCampaign(ctx context.Context, id int64) (campaign *campaigns.Campaign, err error) {
	traceTitle := fmt.Sprintf("campaign: %d", id)
	tr, ctx := trace.New(ctx, "service.PublishCampaign", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Done(&err)

	campaign, err = tx.GetCampaign(ctx, GetCampaignOpts{ID: id})
	if err != nil {
		return nil, errors.Wrap(err, "getting campaign")
	}

	err = backend.CheckSiteAdminOrSameUser(ctx, campaign.AuthorID)
	if err != nil {
		return nil, err
	}

	return campaign, s.createChangesetJobsWithStore(ctx, tx, campaign)
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
func (s *Service) CloseOpenChangesets(ctx context.Context, cs []*campaigns.Changeset) (err error) {
	cs = selectChangesets(cs, func(c *campaigns.Changeset) bool {
		return c.ExternalState == campaigns.ChangesetStateOpen
	})

	if len(cs) == 0 {
		return nil
	}

	reposStore := repos.NewDBStore(s.store.DB(), sql.TxOptions{})
	bySource, err := GroupChangesetsBySource(ctx, reposStore, s.cf, nil, cs...)
	if err != nil {
		return err
	}

	errs := &multierror.Error{}
	for _, s := range bySource {
		for _, c := range s.Changesets {
			if err := s.CloseChangeset(ctx, c); err != nil {
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
	// SyncChangesetsWithSources does) our burndown chart will be outdated
	// until the next run of campaigns.Syncer.
	return SyncChangesetsWithSources(ctx, s.store, bySource)
}

// AddChangesetsToCampaign adds the given changeset IDs to the given campaign's
// ChangesetIDs and the campaign ID to the CampaignIDs of each changeset.
// It updates the campaign and the changesets in the database.
// If one of the changeset IDs is invalid an error is returned.
func (s *Service) AddChangesetsToCampaign(ctx context.Context, campaignID int64, changesetIDs []int64) (campaign *campaigns.Campaign, err error) {
	traceTitle := fmt.Sprintf("campaign: %d, changesets: %v", campaignID, changesetIDs)
	tr, ctx := trace.New(ctx, "service.EnqueueChangesetSync", traceTitle)
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

	if campaign.PatchSetID != 0 {
		return nil, errors.New("Changesets can only be added to campaigns that don't create their own changesets")
	}

	set := map[int64]struct{}{}
	for _, id := range changesetIDs {
		set[id] = struct{}{}
	}

	changesets, _, err := tx.ListChangesets(ctx, ListChangesetsOpts{IDs: changesetIDs})
	if err != nil {
		return nil, err
	}

	for _, c := range changesets {
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
	if _, err := s.store.GetChangeset(ctx, GetChangesetOpts{ID: id}); err != nil {
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

	err = s.store.ResetFailedChangesetJobs(ctx, campaign.ID)
	if err != nil {
		return nil, errors.Wrap(err, "resetting failed changeset jobs")
	}

	return campaign, nil
}

// CreateChangesetJobForPatch creates a ChangesetJob for the
// Patch with the given ID. The Patch has to belong to a
// PatchSet that was attached to a Campaign.
func (s *Service) CreateChangesetJobForPatch(ctx context.Context, patchID int64) (err error) {
	traceTitle := fmt.Sprintf("patch: %d", patchID)
	tr, ctx := trace.New(ctx, "service.CreateChangesetJobForPatch", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	job, err := s.store.GetPatch(ctx, GetPatchOpts{ID: patchID})
	if err != nil {
		return err
	}

	campaign, err := s.store.GetCampaign(ctx, GetCampaignOpts{PatchSetID: job.PatchSetID})
	if err != nil {
		return err
	}

	err = backend.CheckSiteAdminOrSameUser(ctx, campaign.AuthorID)
	if err != nil {
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
		// Already exists
		return nil
	}

	changesetJob := &campaigns.ChangesetJob{
		CampaignID: campaign.ID,
		PatchID:    job.ID,
	}
	err = tx.CreateChangesetJob(ctx, changesetJob)
	if err != nil {
		return err
	}
	return nil
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

// ErrManualCampaignUpdatePatchIllegal is returned by UpdateCampaign if a patch set
// is to be attached to a manual campaign.
var ErrManualCampaignUpdatePatchIllegal = errors.New("cannot update a manual campaign to have a patch set")

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
	if oldPatchSetID == 0 && args.PatchSet != nil {
		return nil, nil, ErrManualCampaignUpdatePatchIllegal
	}
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

	status, err := tx.GetCampaignStatus(ctx, GetCampaignStatusOpts{ID: campaign.ID})
	if err != nil {
		return nil, nil, err
	}

	if status.Processing() {
		return nil, nil, ErrUpdateProcessingCampaign
	}

	published, err := campaignPublished(ctx, tx, campaign.ID)
	if err != nil {
		return nil, nil, err
	}
	partiallyPublished := !published && status.Total != 0

	if campaign.PatchSetID != 0 && updateBranch {
		if published || partiallyPublished {
			return nil, nil, ErrPublishedCampaignBranchChange
		}
	}

	if !published && !partiallyPublished {
		// If the campaign hasn't been published yet and no Changesets have
		// been individually published (through the `PublishChangeset`
		// mutation), we can simply update the attributes on the Campaign
		// because no ChangesetJobs have been created yet that need updating.
		return campaign, nil, tx.UpdateCampaign(ctx, campaign)
	}

	// If we do have to update ChangesetJobs/Changesets, here's a fast path: if
	// we don't update the PatchSet, we don't need to rewire ChangesetJobs,
	// but only update name/description if they changed.
	if !updatePatchSetID && updateAttributes {
		err := tx.UpdateCampaign(ctx, campaign)
		if err != nil {
			return campaign, nil, err
		}
		return campaign, nil, tx.ResetChangesetJobs(ctx, campaign.ID)
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

// campaignPublished returns true if all ChangesetJobs have been created yet
// (they might still be processing).
func campaignPublished(ctx context.Context, store *Store, campaign int64) (bool, error) {
	changesetCreation, err := store.GetLatestChangesetJobCreatedAt(ctx, campaign)
	if err != nil {
		return false, errors.Wrap(err, "getting latest changesetjob creation time")
	}
	// GetLatestChangesetJobCreatedAt returns a zero time.Time if not all
	// ChangesetJobs have been created yet.
	return !changesetCreation.IsZero(), nil
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

	for _, j := range newPatches {
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

	for _, group := range byRepoID {
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

func selectChangesets(cs []*campaigns.Changeset, predicate func(*campaigns.Changeset) bool) []*campaigns.Changeset {
	i := 0
	for _, c := range cs {
		if predicate(c) {
			cs[i] = c
			i++
		}
	}

	return cs[:i]
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
