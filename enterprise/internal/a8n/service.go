package a8n

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewService returns a Service.
func NewService(store *Store, git GitserverClient, cf *httpcli.Factory) *Service {
	return NewServiceWithClock(store, git, cf, func() time.Time {
		return time.Now().UTC().Truncate(time.Microsecond)
	})
}

// NewServiceWithClock returns a Service the given clock used
// to generate timestamps.
func NewServiceWithClock(store *Store, git GitserverClient, cf *httpcli.Factory, clock func() time.Time) *Service {
	return &Service{
		store: store,
		git:   git,
		cf:    cf,
		clock: clock,
	}
}

type GitserverClient interface {
	CreateCommitFromPatch(ctx context.Context, req protocol.CreateCommitFromPatchRequest) (string, error)
}

type Service struct {
	store *Store
	git   GitserverClient
	cf    *httpcli.Factory

	clock func() time.Time
}

// CreateCampaign creates the Campaign. When a CampaignPlanID is set, it also
// creates one ChangesetJob for each CampaignJob belonging to the respective
// CampaignPlan, together with the Campaign in a transaction.
func (s *Service) CreateCampaign(ctx context.Context, c *a8n.Campaign) (err error) {
	tx, err := s.store.Transact(ctx)
	if err != nil {
		return err
	}

	defer tx.Done(&err)

	if err := tx.CreateCampaign(ctx, c); err != nil {
		return err
	}

	if c.CampaignPlanID == 0 {
		return nil
	}

	jobs, _, err := tx.ListCampaignJobs(ctx, ListCampaignJobsOpts{
		CampaignPlanID: c.CampaignPlanID,
		Limit:          10000,
		OnlyFinished:   true,
		OnlyWithDiff:   true,
	})
	if err != nil {
		return err
	}

	for _, job := range jobs {

		changesetJob := &a8n.ChangesetJob{
			CampaignID:    c.ID,
			CampaignJobID: job.ID,
		}
		err = tx.CreateChangesetJob(ctx, changesetJob)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) RunChangesetJobs(ctx context.Context, c *a8n.Campaign) error {
	jobs, _, err := s.store.ListChangesetJobs(ctx, ListChangesetJobsOpts{
		CampaignID: c.ID,
		Limit:      10000,
	})
	if err != nil {
		return err
	}

	errs := &multierror.Error{}
	for _, job := range jobs {
		err := s.runChangesetJob(ctx, c, job)
		if err != nil {
			err = errors.Wrapf(err, "ChangesetJob %d", job.ID)
			errs = multierror.Append(errs, err)
		}
	}

	return errs.ErrorOrNil()
}

func (s *Service) runChangesetJob(
	ctx context.Context,
	c *a8n.Campaign,
	job *a8n.ChangesetJob,
) (err error) {
	// TODO(a8n):
	//   - Ensure all of these calls are idempotent so they can be safely retried.
	defer func() {
		if err != nil {
			job.Error = err.Error()
		}
		job.FinishedAt = s.clock()

		if e := s.store.UpdateChangesetJob(ctx, job); e != nil {
			if err == nil {
				err = e
			} else {
				err = multierror.Append(err, e)
			}
		}
	}()

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

	headRefName := "sourcegraph/campaign-" + strconv.FormatInt(c.ID, 10)

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
			Date:        job.StartedAt,
		},
		// We use unified diffs, not git diffs, which means they're missing the
		// `a/` and `/b` filename prefixes. `-p0` tells `git apply` to not
		// expect and strip prefixes.
		GitApplyArgs: []string{"-p0"},
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

	cs := repos.Changeset{
		Title:       c.Name,
		Body:        c.Description,
		BaseRefName: baseRef,
		HeadRefName: headRefName,
		Repo:        repo,
		Changeset: &a8n.Changeset{
			RepoID:      int32(repo.ID),
			CampaignIDs: []int64{job.CampaignID},
		},
	}

	ccs, ok := src.(repos.ChangesetSource)
	if !ok {
		return errors.Errorf("creating changesets on code host of repo %q is not implemented", repo.Name)
	}

	if err = ccs.CreateChangeset(ctx, &cs); err != nil {
		return err
	}

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return err
	}

	defer tx.Done(&err)

	if err = tx.CreateChangesets(ctx, cs.Changeset); err != nil {
		return err
	}

	c.ChangesetIDs = append(c.ChangesetIDs, cs.Changeset.ID)
	if err = tx.UpdateCampaign(ctx, c); err != nil {
		return err
	}

	job.ChangesetID = cs.Changeset.ID
	return tx.UpdateChangesetJob(ctx, job)
}
