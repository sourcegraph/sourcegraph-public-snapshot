package a8n

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

// NewCampaignsService returns a CampaignsService.
func NewCampaignsService(store *Store, git GitserverClient) *CampaignsService {
	return NewCampaignsServiceWithClock(store, git, func() time.Time {
		return time.Now().UTC().Truncate(time.Microsecond)
	})
}

// NewCampaignsServiceWithClock returns a CampaignsService the given clock used
// to generate timestamps.
func NewCampaignsServiceWithClock(store *Store, git GitserverClient, clock func() time.Time) *CampaignsService {
	return &CampaignsService{
		store: store,
		git:   git,
		clock: clock,
	}
}

type GitserverClient interface {
	CreateCommitFromPatch(ctx context.Context, req protocol.CreateCommitFromPatchRequest) (string, error)
}

type CampaignsService struct {
	store *Store
	git   GitserverClient

	clock func() time.Time
}

func (cc *CampaignsService) CreateCampaign(ctx context.Context, c *a8n.Campaign) error {
	if err := cc.store.CreateCampaign(ctx, c); err != nil {
		return err
	}

	if c.CampaignPlanID == 0 {
		return nil
	}

	jobs, _, err := cc.store.ListCampaignJobs(ctx, ListCampaignJobsOpts{
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

	repoStore := repos.NewDBStore(cc.store.DB(), sql.TxOptions{})
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

		err = cc.store.CreateChangesetJob(ctx, rel.ChangesetJob)
		if err != nil {
			return err
		}
	}

	// TODO: This part can happen asynchronously
	for _, rel := range rels {
		err := cc.runChangesetJob(ctx, c, rel.ChangesetJob, rel.Repo, rel.CampaignJob)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cc *CampaignsService) runChangesetJob(
	ctx context.Context,
	c *a8n.Campaign,
	job *a8n.ChangesetJob,
	repo *repos.Repo,
	campaignJob *a8n.CampaignJob,
) (err error) {
	// TODO(a8n):
	//   - Ensure all of these calls are idempotent so they can be safely retried.
	defer func() {
		if err != nil {
			job.Error = err.Error()
		}
		job.FinishedAt = cc.clock()
		err = cc.store.UpdateChangesetJob(ctx, job)
		return
	}()

	job.StartedAt = cc.clock()

	rev, err := cc.git.CreateCommitFromPatch(ctx, protocol.CreateCommitFromPatchRequest{
		Repo:       api.RepoName(repo.Name),
		BaseCommit: campaignJob.Rev,
		// IMPORTANT: We add a trailing newline here, otherwise `git apply`
		// will fail with "corrupt patch at line <N>" where N is the last line.
		Patch:     campaignJob.Diff + "\n",
		TargetRef: "sourcegraph/campaign-" + strconv.FormatInt(c.ID, 10),
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

	fmt.Printf("pushed rev: %s\n", rev)
	// client := github.NewClient()
	// client.CreatePullRequest(rev)

	// TODO(a8n):
	//   - Create a Changeset once we have an external ID for the created Pull Request
	//   - Update `ChangesetID` on `ChangesetJob`
	//   - Add the Changeset to the Campaign, add Campaign to Changeset
	return nil
}
