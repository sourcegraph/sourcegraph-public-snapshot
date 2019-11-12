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

func (s *Service) CreateCampaign(ctx context.Context, c *a8n.Campaign) error {
	if err := s.store.CreateCampaign(ctx, c); err != nil {
		return err
	}

	if c.CampaignPlanID == 0 {
		return nil
	}

	jobs, _, err := s.store.ListCampaignJobs(ctx, ListCampaignJobsOpts{
		CampaignPlanID: c.CampaignPlanID,
		Limit:          10000,
	})
	if err != nil {
		return err
	}

	// RepoRels contains the joined repo, campaign job and the changeset job.
	type RepoRels struct {
		*repos.Repo
		*a8n.CampaignJob
		*a8n.ChangesetJob
	}

	rels := make(map[int32]*RepoRels, len(jobs))
	repoIDs := make([]uint32, len(jobs))

	for i, job := range jobs {
		rels[job.RepoID] = &RepoRels{CampaignJob: job}
		repoIDs[i] = uint32(job.RepoID)
	}

	repoStore := repos.NewDBStore(s.store.DB(), sql.TxOptions{})
	rs, err := repoStore.ListRepos(ctx, repos.StoreListReposArgs{IDs: repoIDs})
	if err != nil {
		return err
	}

	for _, repo := range rs {
		rel := rels[int32(repo.ID)]
		rel.Repo = repo

		rel.ChangesetJob = &a8n.ChangesetJob{
			CampaignID:    c.ID,
			CampaignJobID: rel.CampaignJob.ID,
		}

		err = s.store.CreateChangesetJob(ctx, rel.ChangesetJob)
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
	store := s.store

	// TODO(a8n):
	//   - Ensure all of these calls are idempotent so they can be safely retried.
	defer func() {
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
		return
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
		Push: true,
	})

	if err != nil {
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

			c := cfg.(*schema.GitHubConnection)
			if c.Token != "" {
				externalService = e
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

	cs := repos.Changeset{
		Title:       c.Name,
		Body:        c.Description,
		BaseRefName: "master",
		HeadRefName: headRefName,
		Repo:        repo,
		Changeset: &a8n.Changeset{
			RepoID:      int32(repo.ID),
			CampaignIDs: []int64{job.CampaignID},
		},
	}

	cc, ok := src.(interface {
		CreateChangeset(context.Context, *repos.Changeset) error
	})

	if !ok {
		return errors.Errorf("creating changesets on code host of repo %q is not implemented", repo.Name)
	}

	if err = cc.CreateChangeset(ctx, &cs); err != nil {
		return err
	}

	store, err = store.Transact(ctx)
	if err != nil {
		return err
	}

	defer store.Done(&err)

	if err = store.CreateChangesets(ctx, cs.Changeset); err != nil {
		return err
	}

	job.ChangesetID = cs.Changeset.ID
	c.ChangesetIDs = append(c.ChangesetIDs, cs.Changeset.ID)

	return store.UpdateCampaign(ctx, c)
}
