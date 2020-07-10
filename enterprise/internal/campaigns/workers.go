package campaigns

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/schema"
)

// maxWorkers defines the maximum number of changeset jobs to run in parallel.
var maxWorkers = env.Get("CAMPAIGNS_MAX_WORKERS", "8", "maximum number of repository jobs to run in parallel")

const defaultWorkerCount = 8

type GitserverClient interface {
	CreateCommitFromPatch(ctx context.Context, req protocol.CreateCommitFromPatchRequest) (string, error)
}

// RunWorkers should be executed in a background goroutine and is responsible
// for finding pending ChangesetJobs and executing them.
// ctx should be canceled to terminate the function.
func RunWorkers(ctx context.Context, s *Store, clock func() time.Time, gitClient GitserverClient, sourcer repos.Sourcer, backoffDuration time.Duration) {
	workerCount, err := strconv.Atoi(maxWorkers)
	if err != nil {
		log15.Error("Parsing max worker count failed. Falling back to default.", "default", defaultWorkerCount, "err", err)
		workerCount = defaultWorkerCount
	}

	externalURL := func() string {
		return conf.Cached(func() interface{} {
			return conf.Get().ExternalURL
		})().(string)
	}

	// process is executed inside a database transaction that's opened by
	// ProcessPendingChangesetJobs.
	process := func(ctx context.Context, s *Store, job campaigns.ChangesetJob) error {
		c, err := s.GetCampaign(ctx, GetCampaignOpts{
			ID: job.CampaignID,
		})
		if err != nil {
			return errors.Wrap(err, "getting campaign")
		}

		if runErr := ExecChangesetJob(ctx, c, &job, ExecChangesetJobOpts{
			Clock:       clock,
			ExternalURL: externalURL(),
			GitClient:   gitClient,
			Sourcer:     sourcer,
			Store:       s,
		}); runErr != nil {
			log15.Error("ExecChangesetJob", "jobID", job.ID, "err", runErr)
		}
		// We don't assign to err here so that we don't roll back the transaction
		// ExecChangesetJob will save the error in the job row
		return nil
	}
	worker := func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				didRun, err := s.ProcessPendingChangesetJobs(ctx, process)
				if err != nil {
					log15.Error("Running changeset job", "err", err)
				}
				// Back off on error or when no jobs available
				if err != nil || !didRun {
					time.Sleep(backoffDuration)
				}
			}
		}
	}
	for i := 0; i < workerCount; i++ {
		go worker()
	}
}

type ExecChangesetJobOpts struct {
	Clock       func() time.Time
	Store       *Store
	GitClient   GitserverClient
	Sourcer     repos.Sourcer
	ExternalURL string
}

// ExecChangesetJob will execute the given ChangesetJob for the given campaign.
// It must be executed inside a transaction.
// It is idempotent and if the job has already been executed it will not be
// executed.
// ProcessPendingChangesetJobs opens a transaction before ultimately calling
// ExecChangesetJob. If ExecChangesetJob is called outside of that context, a
// transaction needs to be opened.
func ExecChangesetJob(
	ctx context.Context,
	c *campaigns.Campaign,
	job *campaigns.ChangesetJob,
	opts ExecChangesetJobOpts,
) (err error) {
	tr, ctx := trace.New(ctx, "service.ExecChangesetJob", fmt.Sprintf("job_id: %d", job.ID))
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
		job.FinishedAt = opts.Clock()

		if e := store.UpdateChangesetJob(ctx, job); e != nil {
			if err == nil {
				err = e
			} else {
				err = multierror.Append(err, e)
			}
		}
		changesetJobUpdated = true
	}
	defer runFinalUpdate(ctx, opts.Store)

	job.StartedAt = opts.Clock()

	patch, err := opts.Store.GetPatch(ctx, GetPatchOpts{ID: job.PatchID})
	if err != nil {
		return err
	}

	reposStore := repos.NewDBStore(opts.Store.DB(), sql.TxOptions{})
	rs, err := reposStore.ListRepos(ctx, repos.StoreListReposArgs{IDs: []api.RepoID{patch.RepoID}})
	if err != nil {
		return err
	}
	if len(rs) != 1 {
		return errors.Errorf("repo not found: %d", patch.RepoID)
	}
	repo := rs[0]

	branch := c.Branch
	ensureUniqueRef := true
	if job.Branch != "" {
		// If job.Branch is set that means this method has already been
		// executed for the given job. In that case, we want to use job.Branch
		// as the ref, since we created it, and not fallback to another ref.
		branch = job.Branch
		ensureUniqueRef = false
	}

	ref, err := opts.GitClient.CreateCommitFromPatch(ctx, protocol.CreateCommitFromPatchRequest{
		Repo:       api.RepoName(repo.Name),
		BaseCommit: patch.Rev,
		// IMPORTANT: We add a trailing newline here, otherwise `git apply`
		// will fail with "corrupt patch at line <N>" where N is the last line.
		Patch:     patch.Diff + "\n",
		TargetRef: branch,
		UniqueRef: ensureUniqueRef,
		CommitInfo: protocol.PatchCommitInfo{
			Message:     c.Name,
			AuthorName:  "Sourcegraph Bot",
			AuthorEmail: "campaigns@sourcegraph.com",
			Date:        job.CreatedAt,
		},
		// We use unified diffs, not git diffs, which means they're missing the
		// `a/` and `/b` filename prefixes. `-p0` tells `git apply` to not
		// expect and strip prefixes.
		GitApplyArgs: []string{"-p0"},
		Push:         true,
	})
	if err != nil {
		if diffErr, ok := err.(*protocol.CreateCommitFromPatchError); ok {
			return errors.Errorf(
				"creating commit from patch for repository %q: %s\n"+
					"```\n"+
					"$ %s\n"+
					"%s\n"+
					"```",
				diffErr.RepositoryName, diffErr.InternalError, diffErr.Command, strings.TrimSpace(diffErr.CombinedOutput))
		}
		return err
	}
	if job.Branch != "" && job.Branch != ref {
		return fmt.Errorf("ref %q doesn't match ChangesetJob's branch %q", ref, job.Branch)
	}
	job.Branch = ref

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

	sources, err := opts.Sourcer(externalService)
	if err != nil {
		return err
	}
	if len(sources) != 1 {
		return errors.New("invalid number of sources for external service")
	}
	src := sources[0]

	baseRef := "refs/heads/master"
	if patch.BaseRef != "" {
		baseRef = patch.BaseRef
	}

	cs := repos.Changeset{
		Title:   c.Name,
		Body:    c.GenChangesetBody(opts.ExternalURL),
		BaseRef: baseRef,
		HeadRef: git.EnsureRefPrefix(ref),
		Repo:    repo,
		Changeset: &campaigns.Changeset{
			RepoID:            repo.ID,
			CampaignIDs:       []int64{job.CampaignID},
			CreatedByCampaign: true,
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
	events := clone.Events()
	SetDerivedState(ctx, clone, events)
	if err = opts.Store.CreateChangesets(ctx, clone); err != nil {
		if _, ok := err.(AlreadyExistError); !ok {
			return err
		}

		// Changeset already exists and the call to CreateChangesets overwrote
		// the Metadata field with the metadata in the database that's possibly
		// outdated.
		// We restore the newest metadata returned by the
		// `ccs.CreateChangesets` call above and then update the Changeset in
		// the database.
		if err := clone.SetMetadata(cs.Changeset.Metadata); err != nil {
			return errors.Wrap(err, "setting changeset metadata")
		}
		events = clone.Events()
		SetDerivedState(ctx, clone, events)

		clone.CampaignIDs = append(clone.CampaignIDs, job.CampaignID)
		clone.CreatedByCampaign = true

		if err = opts.Store.UpdateChangesets(ctx, clone); err != nil {
			return err
		}
	}
	// the events don't have the changesetID yet, because it's not known at the point of cloning
	for _, e := range events {
		e.ChangesetID = clone.ID
	}
	if err := opts.Store.UpsertChangesetEvents(ctx, events...); err != nil {
		log15.Error("UpsertChangesetEvents", "err", err)
		return err
	}

	c.ChangesetIDs = append(c.ChangesetIDs, clone.ID)
	if err = opts.Store.UpdateCampaign(ctx, c); err != nil {
		return err
	}

	job.ChangesetID = clone.ID
	runFinalUpdate(ctx, opts.Store)
	return err
}
