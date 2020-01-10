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
func (s *Service) CreateCampaignPlanFromPatches(ctx context.Context, patches []a8n.CampaignPlanPatch) (*a8n.CampaignPlan, error) {
	// Look up all repositories.
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
// When draft is true it also does not set the PublishedAt field on the
// Campaign.
func (s *Service) CreateCampaign(ctx context.Context, c *a8n.Campaign, draft bool) error {
	var err error
	tr, ctx := trace.New(ctx, "Service.CreateCampaign", fmt.Sprintf("Name: %q", c.Name))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	tx, err := s.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer tx.Done(&err)

	c.CreatedAt = s.clock()
	c.UpdatedAt = c.CreatedAt
	if !draft {
		c.PublishedAt = c.CreatedAt
	}

	if err := tx.CreateCampaign(ctx, c); err != nil {
		return err
	}

	if c.CampaignPlanID == 0 || draft {
		return nil
	}

	return s.createChangesetJobsWithStore(ctx, tx, c)
}

// ErrNoCampaignJobs is returned by CreateCampaign if a CampaignPlanID was
// specified but the CampaignPlan does not have any (finished) CampaignJobs.
var ErrNoCampaignJobs = errors.New("cannot create a Campaign without any changesets")

func (s *Service) createChangesetJobsWithStore(ctx context.Context, store *Store, c *a8n.Campaign) error {
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

// RunChangesetJobs will run all the changeset jobs for the supplied campaign.
// It is idempotent and jobs that have already completed will not be rerun.
func (s *Service) RunChangesetJobs(ctx context.Context, c *a8n.Campaign) error {
	var err error
	tr, ctx := trace.New(ctx, "Service.RunChangesetJobs", fmt.Sprintf("Campaign: %q", c.Name))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	jobs, _, err := s.store.ListChangesetJobs(ctx, ListChangesetJobsOpts{
		CampaignID: c.ID,
		Limit:      -1,
	})
	if err != nil {
		return err
	}

	errs := &multierror.Error{}
	for _, job := range jobs {
		err := s.RunChangesetJob(ctx, c, job)
		if err != nil {
			err = errors.Wrapf(err, "ChangesetJob %d", job.ID)
			errs = multierror.Append(errs, err)
		}
	}

	return errs.ErrorOrNil()
}

// RunChangesetJob will run the given ChangesetJob for the given campaign. It
// is idempotent and if the job has already been run it will not be rerun.
func (s *Service) RunChangesetJob(
	ctx context.Context,
	c *a8n.Campaign,
	job *a8n.ChangesetJob,
) (err error) {
	tr, ctx := trace.New(ctx, "service.RunChangeSetJob", fmt.Sprintf("job_id: %d", job.ID))
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
		job.FinishedAt = s.clock()

		if e := store.UpdateChangesetJob(ctx, job); e != nil {
			if err == nil {
				err = e
			} else {
				err = multierror.Append(err, e)
			}
		}
		changesetJobUpdated = true
	}
	defer runFinalUpdate(ctx, s.store)

	// We start a transaction here so that we can grab a lock
	tx, err := s.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer tx.Done(&err)

	lockKey := fmt.Sprintf("RunChangesetJob: %d", job.ID)
	acquired, err := tx.TryAcquireAdvisoryLock(ctx, lockKey)
	if err != nil {
		return errors.Wrap(err, "acquiring lock")
	}
	if !acquired {
		return errors.New("could not acquire lock")
	}

	job.StartedAt = s.clock()

	campaignJob, err := s.store.GetCampaignJob(ctx, GetCampaignJobOpts{ID: job.CampaignJobID})
	if err != nil {
		return err
	}

	reposStore := repos.NewDBStore(s.store.DB(), sql.TxOptions{})
	rs, err := reposStore.ListRepos(ctx, repos.StoreListReposArgs{IDs: []uint32{uint32(campaignJob.RepoID)}})
	if err != nil {
		return err
	}
	if len(rs) != 1 {
		return errors.Errorf("repo not found: %d", campaignJob.RepoID)
	}
	repo := rs[0]

	headRefName := fmt.Sprintf("sourcegraph/%s-%d", git.HumanReadableBranchName(c.Name), c.CreatedAt.Unix())

	_, err = s.git.CreateCommitFromPatch(ctx, protocol.CreateCommitFromPatchRequest{
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

	src, err := repos.NewSource(externalService, s.cf)
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

	err = ccs.CreateChangeset(ctx, &cs)
	if err != nil {
		return errors.Wrap(err, "creating changeset")
	}

	if err = tx.CreateChangesets(ctx, cs.Changeset); err != nil {
		if _, ok := err.(AlreadyExistError); !ok {
			return err
		}
		// Changeset already exists so continue
	}

	c.ChangesetIDs = append(c.ChangesetIDs, cs.Changeset.ID)
	if err = tx.UpdateCampaign(ctx, c); err != nil {
		return err
	}

	job.ChangesetID = cs.Changeset.ID
	runFinalUpdate(ctx, tx)
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

// PublishCampaign publishes the Campaign with the given ID if it has not been
// published yet by turning the CampaignJobs attached to the CampaignPlan of
// the Campaign into ChangesetJobs and running them.
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

	if !campaign.PublishedAt.IsZero() {
		return campaign, nil
	}

	campaign.PublishedAt = s.clock()

	if err = tx.UpdateCampaign(ctx, campaign); err != nil {
		return campaign, err
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

// CreateChangesetJob creates a ChangesetJob for the CampaignJob with the given
// ID. The CampaignJob has to belong to a CampaignPlan that was attached to a
// Campaign.
// It returns the newly created ChangesetJob and its Campaign, which can then
// be passed to RunChangesetJob.
func (s *Service) CreateChangesetJobForCampaignJob(ctx context.Context, id int64) (_ *a8n.ChangesetJob, _ *a8n.Campaign, err error) {
	traceTitle := fmt.Sprintf("campaignJob: %d", id)
	tr, ctx := trace.New(ctx, "service.CreateChangesetJobForCampaignJob", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	job, err := s.store.GetCampaignJob(ctx, GetCampaignJobOpts{ID: id})
	if err != nil {
		return nil, nil, err
	}

	campaign, err := s.store.GetCampaign(ctx, GetCampaignOpts{CampaignPlanID: job.CampaignPlanID})
	if err != nil {
		return nil, nil, err
	}

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Done(&err)

	existing, err := tx.GetChangesetJob(ctx, GetChangesetJobOpts{
		CampaignID:    campaign.ID,
		CampaignJobID: job.ID,
	})
	if existing != nil && err == nil {
		return existing, campaign, nil
	}
	if err != nil && err != ErrNoResults {
		return nil, nil, err
	}

	changesetJob := &a8n.ChangesetJob{CampaignID: campaign.ID, CampaignJobID: job.ID}
	err = tx.CreateChangesetJob(ctx, changesetJob)
	if err != nil {
		return nil, nil, err
	}

	return changesetJob, campaign, nil
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
