package a8n

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// NewService returns a Service.
func NewService(store *Store, git GitserverClient, repoResolveRevision repoResolveRevision, cf *httpcli.Factory) *Service {
	return NewServiceWithClock(store, git, repoResolveRevision, cf, store.Clock())
}

// NewServiceWithClock returns a Service the given clock used
// to generate timestamps.
func NewServiceWithClock(store *Store, git GitserverClient, repoResolveRevision repoResolveRevision, cf *httpcli.Factory, clock func() time.Time) *Service {
	svc := &Service{
		store:               store,
		git:                 git,
		repoResolveRevision: repoResolveRevision,
		cf:                  cf,
		clock:               clock,
	}
	if svc.repoResolveRevision == nil {
		svc.repoResolveRevision = defaultRepoResolveRevision
	}

	return svc
}

type GitserverClient interface {
	CreateCommitFromPatch(ctx context.Context, req protocol.CreateCommitFromPatchRequest) (string, error)
}

type Service struct {
	store               *Store
	git                 GitserverClient
	repoResolveRevision repoResolveRevision
	cf                  *httpcli.Factory

	clock func() time.Time
}

// repoResolveRevision resolves a Git revspec in a repository and returns the resolved commit ID.
type repoResolveRevision func(context.Context, *repos.Repo, string) (api.CommitID, error)

// defaultRepoResolveRevision is an implementation of repoResolveRevision that talks to gitserver to
// resolve a Git revspec.
var defaultRepoResolveRevision = func(ctx context.Context, repo *repos.Repo, revspec string) (api.CommitID, error) {
	return backend.Repos.ResolveRev(ctx,
		&types.Repo{Name: api.RepoName(repo.Name), ExternalRepo: repo.ExternalRepo},
		revspec,
	)
}

// CreateCampaignPlanFromPatches creates a CampaignPlan and its associated CampaignJobs from patches
// computed by the caller. There is no diff execution or computation performed during creation of
// the CampaignJobs in this case (unlike when using Runner to create a CampaignPlan from a
// specification).
//
// If resolveRevision is nil, a default implementation is used.
func (s *Service) CreateCampaignPlanFromPatches(ctx context.Context, patches []a8n.CampaignPlanPatch, userID int32) (*a8n.CampaignPlan, error) {
	if userID == 0 {
		return nil, backend.ErrNotAuthenticated
	}
	// Look up all repositories
	reposStore := repos.NewDBStore(s.store.DB(), sql.TxOptions{})
	repoIDs := make([]uint32, len(patches))
	for i, patch := range patches {
		repoIDs[i] = uint32(patch.Repo)
	}
	allRepos, err := reposStore.ListRepos(ctx, repos.StoreListReposArgs{IDs: repoIDs})
	if err != nil {
		return nil, err
	}
	reposByID := make(map[uint32]*repos.Repo, len(patches))
	for _, repo := range allRepos {
		reposByID[repo.ID] = repo
	}

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Done(&err)

	plan := &a8n.CampaignPlan{
		CampaignType: campaignTypePatch,
		Arguments:    "", // intentionally empty to avoid needless duplication with CampaignJob diffs
		UserID:       userID,
	}

	err = tx.CreateCampaignPlan(ctx, plan)
	if err != nil {
		return nil, err
	}

	for _, patch := range patches {
		repo := reposByID[uint32(patch.Repo)]
		if repo == nil {
			return nil, fmt.Errorf("repository ID %d not found", patch.Repo)
		}
		if !a8n.IsRepoSupported(&repo.ExternalRepo) {
			continue
		}

		commit, err := s.repoResolveRevision(ctx, repo, patch.BaseRevision)
		if err != nil {
			return nil, errors.Wrapf(err, "repository %q", repo.Name)
		}

		job := &a8n.CampaignJob{
			CampaignPlanID: plan.ID,
			RepoID:         int32(patch.Repo),
			BaseRef:        patch.BaseRevision,
			Rev:            commit,
			Diff:           patch.Patch,
			StartedAt:      s.clock(),
			FinishedAt:     s.clock(),
		}
		if err := tx.CreateCampaignJob(ctx, job); err != nil {
			return nil, err
		}
	}

	return plan, nil
}

// CreateCampaign creates the Campaign. When a CampaignPlanID is set on the
// Campaign and the Campaign is not created as a draft, it calls
// CreateChangesetJobs inside the same transaction in which it creates the
// Campaign.
func (s *Service) CreateCampaign(ctx context.Context, c *a8n.Campaign, draft bool) error {
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

	c.CreatedAt = s.clock()
	c.UpdatedAt = c.CreatedAt

	if err := tx.CreateCampaign(ctx, c); err != nil {
		return err
	}

	if c.CampaignPlanID == 0 || draft {
		return nil
	}

	return s.createChangesetJobsWithStore(ctx, tx, c)
}

// ErrNoCampaignJobs is returned by CreateCampaign or UpdateCampaign if a
// CampaignPlanID was specified but the CampaignPlan does not have any
// (finished) CampaignJobs.
var ErrNoCampaignJobs = errors.New("cannot create or update a Campaign without any changesets")

func (s *Service) createChangesetJobsWithStore(ctx context.Context, store *Store, c *a8n.Campaign) error {
	if c.CampaignPlanID == 0 {
		return errors.New("cannot create changesets for campaign with no campaign plan")
	}

	jobs, _, err := store.ListCampaignJobs(ctx, ListCampaignJobsOpts{
		CampaignPlanID:            c.CampaignPlanID,
		Limit:                     -1,
		OnlyFinished:              true,
		OnlyWithDiff:              true,
		OnlyUnpublishedInCampaign: c.ID,
	})
	if err != nil {
		return err
	}

	if len(jobs) == 0 {
		return ErrNoCampaignJobs
	}

	for _, job := range jobs {
		changesetJob := &a8n.ChangesetJob{
			CampaignID:    c.ID,
			CampaignJobID: job.ID,
		}
		err = store.CreateChangesetJob(ctx, changesetJob)
		if err != nil {
			return err
		}
	}

	return nil
}

// RunChangesetJob will run the given ChangesetJob for the given campaign. It
// is idempotent and if the job has already been run it will not be rerun.
func RunChangesetJob(
	ctx context.Context,
	clock func() time.Time,
	store *Store,
	gitClient GitserverClient,
	cf *httpcli.Factory,
	c *a8n.Campaign,
	job *a8n.ChangesetJob,
) (err error) {
	// Store should already have an open transaction but ensure here anyway
	store, err = store.Transact(ctx)
	if err != nil {
		return errors.Wrap(err, "creating transaction")
	}

	tr, ctx := trace.New(ctx, "service.RunChangesetJob", fmt.Sprintf("job_id: %d", job.ID))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	tr.LogFields(log.Bool("completed", job.SuccessfullyCompleted()), log.Int64("job_id", job.ID), log.Int64("campaign_id", c.ID))

	if job.SuccessfullyCompleted() {
		log15.Info("ChangesetJob already completed", "id", job.ID)
		return nil
	}

	// We'll always run a final update but in the happy path it will run as
	// part of a transaction in which case we don't want to run it again in
	// the defer below
	var changesetJobUpdated bool
	runFinalUpdate := func(ctx context.Context, store *Store) {
		if changesetJobUpdated {
			// Don't run again
			return
		}
		if err != nil {
			job.Error = err.Error()
		}
		job.FinishedAt = clock()

		if e := store.UpdateChangesetJob(ctx, job); e != nil {
			if err == nil {
				err = e
			} else {
				err = multierror.Append(err, e)
			}
		}
		changesetJobUpdated = true
	}
	defer runFinalUpdate(ctx, store)

	job.StartedAt = clock()

	campaignJob, err := store.GetCampaignJob(ctx, GetCampaignJobOpts{ID: job.CampaignJobID})
	if err != nil {
		return err
	}

	reposStore := repos.NewDBStore(store.DB(), sql.TxOptions{})
	rs, err := reposStore.ListRepos(ctx, repos.StoreListReposArgs{IDs: []uint32{uint32(campaignJob.RepoID)}})
	if err != nil {
		return err
	}
	if len(rs) != 1 {
		return errors.Errorf("repo not found: %d", campaignJob.RepoID)
	}
	repo := rs[0]

	// TODO: The "campaign" is just here so that updates don't create new
	// branches and new changesets.
	// We should probably persist the `headRefName` on `ChangesetJob` and keep
	// it stable across retries and only set it the first time.
	headRefName := fmt.Sprintf("sourcegraph/%s-%d", "campaign", c.CreatedAt.Unix())

	_, err = gitClient.CreateCommitFromPatch(ctx, protocol.CreateCommitFromPatchRequest{
		Repo:       api.RepoName(repo.Name),
		BaseCommit: campaignJob.Rev,
		// IMPORTANT: We add a trailing newline here, otherwise `git apply`
		// will fail with "corrupt patch at line <N>" where N is the last line.
		Patch:     campaignJob.Diff + "\n",
		TargetRef: headRefName,
		CommitInfo: protocol.PatchCommitInfo{
			Message:     c.Name,
			AuthorName:  "Sourcegraph Bot",
			AuthorEmail: "automation@sourcegraph.com",
			Date:        job.CreatedAt,
		},
		// We use unified diffs, not git diffs, which means they're missing the
		// `a/` and `/b` filename prefixes. `-p0` tells `git apply` to not
		// expect and strip prefixes.
		// Since we also produce diffs manually, we might not have context lines,
		// so we need to disable that check with `--unidiff-zero`.
		GitApplyArgs: []string{"-p0", "--unidiff-zero"},
		Push:         true,
	})

	if err != nil {
		if diffErr, ok := err.(*protocol.CreateCommitFromPatchError); ok {
			return errors.Errorf("creating commit from patch for repo %q: %v (command: %q)", diffErr.RepositoryName, diffErr.Err, diffErr.Command)
		}
		return err
	}

	var externalService *repos.ExternalService
	{
		args := repos.StoreListExternalServicesArgs{IDs: repo.ExternalServiceIDs()}

		es, err := reposStore.ListExternalServices(ctx, args)
		if err != nil {
			return err
		}

		for _, e := range es {
			cfg, err := e.Configuration()
			if err != nil {
				return err
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
			}
			if externalService != nil {
				break
			}
		}
	}

	if externalService == nil {
		return errors.Errorf("no external services found for repo %q", repo.Name)
	}

	src, err := repos.NewSource(externalService, cf)
	if err != nil {
		return err
	}

	baseRef := "refs/heads/master"
	if campaignJob.BaseRef != "" {
		baseRef = campaignJob.BaseRef
	}

	body := c.Description
	if campaignJob.Description != "" {
		body += "\n\n---\n\n" + campaignJob.Description
	}

	cs := repos.Changeset{
		Title:   c.Name,
		Body:    body,
		BaseRef: baseRef,
		HeadRef: git.EnsureRefPrefix(headRefName),
		Repo:    repo,
		Changeset: &a8n.Changeset{
			RepoID:      int32(repo.ID),
			CampaignIDs: []int64{job.CampaignID},
		},
	}

	ccs, ok := src.(repos.ChangesetSource)
	if !ok {
		return errors.Errorf("creating changesets on code host of repo %q is not implemented", repo.Name)
	}

	// TODO: If we're updating the changeset, there's a race condition here.
	// It's possible that `CreateChangeset` doesn't return the newest head ref
	// commit yet, because the API of the codehost doesn't return it yet.
	exists, err := ccs.CreateChangeset(ctx, &cs)
	if err != nil {
		return errors.Wrap(err, "creating changeset")
	}
	// If the Changeset already exists and our source can update it, we try to update it
	if exists {
		outdated, err := isOutdated(&cs)
		if err != nil {
			return errors.Wrap(err, "could not determine whether changeset needs update")
		}

		if outdated {
			err := ccs.UpdateChangeset(ctx, &cs)
			if err != nil {
				return errors.Wrap(err, "updating changeset")
			}
		}
	}

	// We keep a clone because CreateChangesets might overwrite the changeset
	// with outdated metadata.
	clone := cs.Changeset.Clone()
	if err = store.CreateChangesets(ctx, clone); err != nil {
		if _, ok := err.(AlreadyExistError); !ok {
			return err
		}

		// Changeset already exists and the call to CreateChangesets overwrote
		// the Metadata field with the metadata in the database that's possibly
		// outdated.
		// We restore the newest metadata returned by the
		// `ccs.CreateChangesets` call above and then update the Changeset in
		// the database.
		clone.Metadata = cs.Changeset.Metadata
		if err = store.UpdateChangesets(ctx, clone); err != nil {
			return err
		}
	}

	c.ChangesetIDs = append(c.ChangesetIDs, clone.ID)
	if err = store.UpdateCampaign(ctx, c); err != nil {
		return err
	}

	job.ChangesetID = clone.ID
	runFinalUpdate(ctx, store)
	return
}

// CloseCampaign closes the Campaign with the given ID if it has not been closed yet.
func (s *Service) CloseCampaign(ctx context.Context, id int64, closeChangesets bool) (campaign *a8n.Campaign, err error) {
	traceTitle := fmt.Sprintf("campaign: %d, closeChangesets: %t", id, closeChangesets)
	tr, ctx := trace.New(ctx, "service.CloseCampaign", traceTitle)
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

	if !campaign.ClosedAt.IsZero() {
		return campaign, nil
	}

	campaign.ClosedAt = time.Now().UTC()

	if err = tx.UpdateCampaign(ctx, campaign); err != nil {
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
// by turning the CampaignJobs attached to the CampaignPlan of
// the Campaign into ChangesetJobs and enqueuing them
func (s *Service) PublishCampaign(ctx context.Context, id int64) (campaign *a8n.Campaign, err error) {
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
	return campaign, s.createChangesetJobsWithStore(ctx, tx, campaign)
}

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

	// If we don't have to close the changesets, we can simply delete the
	// Campaign and return. The triggers in the database will remove the
	// campaign's ID from the changesets' CampaignIDs.
	if !closeChangesets {
		return s.store.DeleteCampaign(ctx, id)
	}

	// First load the Changesets with the given campaignID, before deleting
	// the campaign would remove the association.
	cs, _, err := s.store.ListChangesets(ctx, ListChangesetsOpts{
		CampaignID: id,
		Limit:      -1,
	})
	if err != nil {
		return err
	}

	// Remove the association manually, since we'll update the Changesets in
	// the database, after closing them and we can't update them with an
	// invalid CampaignID.
	for _, c := range cs {
		c.RemoveCampaignID(id)
	}

	err = s.store.DeleteCampaign(ctx, id)
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
func (s *Service) CloseOpenChangesets(ctx context.Context, cs []*a8n.Changeset) (err error) {
	cs = selectChangesets(cs, func(c *a8n.Changeset) bool {
		s, err := c.State()
		if err != nil {
			log15.Warn("could not determine changeset state", "err", err)
			return false
		}
		return s == a8n.ChangesetStateOpen
	})

	if len(cs) == 0 {
		return nil
	}

	reposStore := repos.NewDBStore(s.store.DB(), sql.TxOptions{})
	syncer := ChangesetSyncer{
		ReposStore:  reposStore,
		Store:       s.store,
		HTTPFactory: s.cf,
	}

	bySource, err := syncer.GroupChangesetsBySource(ctx, cs...)
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
	// until the next run of a8n.Syncer.
	return syncer.SyncChangesetsWithSources(ctx, bySource)
}

// CreateChangesetJobForCampaignJob creates a ChangesetJob for the
// CampaignJob with the given ID. The CampaignJob has to belong to a
// CampaignPlan that was attached to a Campaign.
func (s *Service) CreateChangesetJobForCampaignJob(ctx context.Context, campaignJobID int64) (err error) {
	traceTitle := fmt.Sprintf("campaignJob: %d", campaignJobID)
	tr, ctx := trace.New(ctx, "service.CreateChangesetJobForCampaignJob", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	job, err := s.store.GetCampaignJob(ctx, GetCampaignJobOpts{ID: campaignJobID})
	if err != nil {
		return err
	}

	campaign, err := s.store.GetCampaign(ctx, GetCampaignOpts{CampaignPlanID: job.CampaignPlanID})
	if err != nil {
		return err
	}

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer tx.Done(&err)

	existing, err := tx.GetChangesetJob(ctx, GetChangesetJobOpts{
		CampaignID:    campaign.ID,
		CampaignJobID: job.ID,
	})
	if err != nil && err != ErrNoResults {
		return err
	}
	if existing != nil {
		// Already exists
		return nil
	}
	changesetJob := &a8n.ChangesetJob{
		CampaignID:    campaign.ID,
		CampaignJobID: job.ID,
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
	Plan        *int64
}

// ErrCampaignNameBlank is returned by CreateCampaign or UpdateCampaign if the
// specified Campaign name is blank.
var ErrCampaignNameBlank = errors.New("Campaign title cannot be blank")

// UpdateCampaign updates the Campaign with the given arguments.
func (s *Service) UpdateCampaign(ctx context.Context, args UpdateCampaignArgs) (campaign *a8n.Campaign, detachedChangesets []*a8n.Changeset, err error) {
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

	var updateAttributes, updatePlanID bool

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

	oldPlanID := campaign.CampaignPlanID
	if args.Plan != nil && oldPlanID != *args.Plan {
		campaign.CampaignPlanID = *args.Plan
		updatePlanID = true
	}

	if !updateAttributes && !updatePlanID {
		return campaign, nil, nil
	}

	status, err := tx.GetCampaignStatus(ctx, campaign.ID)
	if err != nil {
		return nil, nil, err
	}
	if !status.Finished() && !status.Canceled {
		return nil, nil, ErrUpdateProcessingCampaign
	}

	published, err := campaignPublished(ctx, tx, campaign.ID)
	if err != nil {
		return nil, nil, err
	}
	partiallyPublished := !published && status.Total != 0

	if !published && !partiallyPublished {
		// If the campaign hasn't been published yet and no Changesets have
		// been individually published (through the `PublishChangeset`
		// mutation), we can simply update the attributes on the Campaign
		// because no ChangesetJobs have been created yet that need updating.
		return campaign, nil, tx.UpdateCampaign(ctx, campaign)
	}

	// If we do have to update ChangesetJobs/Changesets, here's a fast path: if
	// we don't update the CampaignPlan, we don't need to rewire ChangesetJobs,
	// but only update name/description if they changed.
	if !updatePlanID && updateAttributes {
		return campaign, nil, tx.ResetChangesetJobs(ctx, campaign.ID)
	}

	diff, err := computeCampaignUpdateDiff(ctx, tx, campaign, oldPlanID, updateAttributes)
	if err != nil {
		return nil, nil, err
	}

	for _, c := range diff.Update {
		err := tx.UpdateChangesetJob(ctx, c)
		if err != nil {
			return nil, nil, err
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
		return nil, nil, err
	}

	return campaign, changesets, tx.UpdateCampaign(ctx, campaign)
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
	Delete []*a8n.ChangesetJob
	Update []*a8n.ChangesetJob
	Create []*a8n.ChangesetJob
}

// repoJobs is a triplet of jobs that are associated with the same repository.
type repoJobs struct {
	changesetJob   *a8n.ChangesetJob
	campaignJob    *a8n.CampaignJob
	newCampaignJob *a8n.CampaignJob
}

func computeCampaignUpdateDiff(
	ctx context.Context,
	tx *Store,
	campaign *a8n.Campaign,
	oldPlanID int64,
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

	// We need OnlyFinished and OnlyWithDiff because we don't create
	// ChangesetJobs for others.
	campaignJobs, _, err := tx.ListCampaignJobs(ctx, ListCampaignJobsOpts{
		CampaignPlanID: oldPlanID,
		Limit:          -1,
		OnlyFinished:   true,
		OnlyWithDiff:   true,
	})
	if err != nil {
		return nil, errors.Wrap(err, "listing campaign jobs")
	}

	newCampaignJobs, _, err := tx.ListCampaignJobs(ctx, ListCampaignJobsOpts{
		CampaignPlanID: campaign.CampaignPlanID,
		Limit:          -1,
		OnlyFinished:   true,
		OnlyWithDiff:   true,
	})
	if err != nil {
		return nil, errors.Wrap(err, "listing new campaign jobs")
	}

	if len(newCampaignJobs) == 0 {
		return nil, ErrNoCampaignJobs
	}

	// We need to determine which current ChangesetJobs we want to keep and
	// which ones we want to delete.
	// We can find out which ones we want to keep by looking at the RepoID of
	// their CampaignJobs.

	jobsByRepoID, err := mergeByRepoID(changesetJobs, campaignJobs)
	if err != nil {
		return nil, err
	}

	for _, j := range newCampaignJobs {
		if jobs, ok := jobsByRepoID[j.RepoID]; ok {
			jobs.newCampaignJob = j
		} else {
			// If we have new CampaignJobs that don't match an existing
			// ChangesetJob we need to create new ChangesetJobs.
			diff.Create = append(diff.Create, &a8n.ChangesetJob{
				CampaignID:    campaign.ID,
				CampaignJobID: j.ID,
			})
		}
	}

	for _, jobs := range jobsByRepoID {
		// Either we _don't_ have a matching _new_ CampaignJob, then we delete
		// the ChangesetJob and detach & close Changeset.
		if jobs.newCampaignJob == nil {
			diff.Delete = append(diff.Delete, jobs.changesetJob)
			continue
		}

		// Or we have a matching _new_ CampaignJob, then we keep the
		// ChangesetJob around, but need to rewire it.
		jobs.changesetJob.CampaignJobID = jobs.newCampaignJob.ID

		//  And, if the {Diff,Rev,BaseRef,Description} are different, we  need to
		// update the Changeset on the codehost...
		if updateAttributes || campaignJobsDiffer(jobs.newCampaignJob, jobs.campaignJob) {
			// ... to do that, we _reset_ the ChangesetJob, so it gets run again
			// when RunChangesetJobs is called after UpdateCampaign.
			jobs.changesetJob.Error = ""
			jobs.changesetJob.StartedAt = time.Time{}
			jobs.changesetJob.FinishedAt = time.Time{}
		}

		diff.Update = append(diff.Update, jobs.changesetJob)
	}

	return diff, nil
}

// campaignJobsDiffer returns true if the CampaignJobs differ in a way that
// requires updating the Changeset on the codehost.
func campaignJobsDiffer(a, b *a8n.CampaignJob) bool {
	return a.Diff != b.Diff ||
		a.Rev != b.Rev ||
		a.BaseRef != b.BaseRef ||
		a.Description != b.Description
}

func selectChangesets(cs []*a8n.Changeset, predicate func(*a8n.Changeset) bool) []*a8n.Changeset {
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

func mergeByRepoID(chs []*a8n.ChangesetJob, cas []*a8n.CampaignJob) (map[int32]*repoJobs, error) {
	jobs := make(map[int32]*repoJobs, len(chs))

	byID := make(map[int64]*a8n.CampaignJob, len(cas))
	for _, j := range cas {
		byID[j.ID] = j
	}

	for _, j := range chs {
		caj, ok := byID[j.CampaignJobID]
		if !ok {
			return nil, fmt.Errorf("CampaignJob with ID %d cannot be found for ChangesetJob %d", j.CampaignJobID, j.ID)
		}
		jobs[caj.RepoID] = &repoJobs{changesetJob: j, campaignJob: caj}
	}

	return jobs, nil
}
