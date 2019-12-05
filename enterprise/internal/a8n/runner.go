package a8n

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/graph-gophers/graphql-go/relay"
	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// maxRepositories defines the maximum number of repositories over which a
// Runner executes CampaignJobs.
// This upper limit is set while Automation features are still under
// development.
var maxRepositories = env.Get("A8N_MAX_REPOS", "200", "maximum number of repositories over which to run campaigns")

// maxWorkers defines the maximum number of repositories over which a Runner
// executes CampaignJobs in parallel.
var maxWorkers = env.Get("A8N_MAX_WORKERS", "8", "maximum number of repositories run campaigns over in parallel")

// ErrTooManyResults is returned by the Runner's Run method when the
// CampaignType's searchQuery produced more than maxRepositories number of
// repositories.
var ErrTooManyResults = errors.New("search yielded too many results")

// A Runner executes a CampaignPlan by creating and running CampaignJobs
// according to the CampaignPlan's Arguments and CampaignType.
type Runner struct {
	store         *Store
	search        repoSearch
	defaultBranch repoDefaultBranch
	clock         func() time.Time

	ct CampaignType

	started bool
	wg      sync.WaitGroup
}

// repoSearch takes in a raw search query and returns the list of repositories
// associated with the search results.
type repoSearch func(ctx context.Context, query string) ([]*graphqlbackend.RepositoryResolver, error)

// repoDefaultBranch takes in a RepositoryResolver and returns the ref name of
// the repositories default branch and its target commit ID.
type repoDefaultBranch func(ctx context.Context, repo *graphqlbackend.RepositoryResolver) (string, api.CommitID, error)

// ErrNoDefaultBranch is returned by a repoDefaultBranch when no default branch
// could be determined for a given repo.
var ErrNoDefaultBranch = errors.New("could not determine default branch")

// defaultRepoDefaultBranch is an implementation of repoDefaultBranch that uses
// methods defined on RepositoryResolver to talk to gitserver to determine a
// repository's default branch and its target commit ID.
var defaultRepoDefaultBranch = func(ctx context.Context, repo *graphqlbackend.RepositoryResolver) (string, api.CommitID, error) {
	var branch string
	var commitID api.CommitID

	defaultBranch, err := repo.DefaultBranch(ctx)
	if err != nil {
		return branch, commitID, err
	}
	if defaultBranch == nil {
		return branch, commitID, ErrNoDefaultBranch
	}
	branch = defaultBranch.Name()

	commit, err := defaultBranch.Target().Commit(ctx)
	if err != nil {
		return branch, commitID, err
	}

	commitID = api.CommitID(commit.OID())
	return branch, commitID, nil
}

// NewRunner returns a Runner for a given CampaignType.
func NewRunner(store *Store, ct CampaignType, search repoSearch, defaultBranch repoDefaultBranch) *Runner {
	return NewRunnerWithClock(store, ct, search, defaultBranch, func() time.Time {
		return time.Now().UTC().Truncate(time.Microsecond)
	})
}

// NewRunnerWithClock returns a Runner for a given CampaignType with the given clock used
// to generate timestamps
func NewRunnerWithClock(store *Store, ct CampaignType, search repoSearch, defaultBranch repoDefaultBranch, clock func() time.Time) *Runner {
	runner := &Runner{
		store:         store,
		search:        search,
		defaultBranch: defaultBranch,
		ct:            ct,
		clock:         clock,
	}
	if runner.defaultBranch == nil {
		runner.defaultBranch = defaultRepoDefaultBranch
	}

	return runner
}

// The time after which a CampaignJob's execution times out
const jobTimeout = 2 * time.Minute

// Run executes the CampaignPlan by searching for relevant repositories using
// the CampaignType specific searchQuery and then executing CampaignJobs for
// each repository.
// Before it starts executing CampaignJobs it persists the CampaignPlan and the
// new CampaignJobs in a transaction.
// What each CampaignJob then does in each repository depends on the
// CampaignType set on CampaignPlan.
// This is a non-blocking method that will possibly return before all
// CampaignJobs are finished.
func (r *Runner) Run(ctx context.Context, plan *a8n.CampaignPlan) (err error) {
	tr, ctx := trace.New(ctx, "Runner.Run", fmt.Sprintf("plan_id %d", plan.ID))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	tr.LogFields(log.Bool("started", r.started))
	if r.started {
		return errors.New("already started")
	}
	r.started = true

	rs, err := r.search(ctx, r.ct.searchQuery())
	if err != nil {
		return err
	}
	max, err := strconv.ParseInt(maxRepositories, 10, 64)
	if err != nil {
		return err
	}
	if len(rs) > int(max) {
		err = ErrTooManyResults
		return err
	}

	jobs, err := r.createPlanAndJobs(ctx, plan, rs)
	if err != nil {
		return err
	}

	numWorkers, err := strconv.ParseInt(maxWorkers, 10, 64)
	if err != nil {
		return err
	}
	tr.LogFields(log.Int64("num_workers", numWorkers), log.Int("num_jobs", len(jobs)))

	queue := make(chan *a8n.CampaignJob)
	worker := func(queue chan *a8n.CampaignJob) {
		for job := range queue {
			r.runJob(ctx, job)
		}
	}

	for i := 0; i < int(numWorkers); i++ {
		go worker(queue)
	}

	r.wg.Add(len(jobs))

	go func() {
		for _, job := range jobs {
			queue <- job
		}

		close(queue)
	}()

	return nil
}

func (r *Runner) runJob(pctx context.Context, job *a8n.CampaignJob) {
	var err error
	tr, ctx := trace.New(pctx, "Runner.runJob", fmt.Sprintf("job_id %d", job.ID))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	ctx, cancel := context.WithTimeout(ctx, jobTimeout)
	defer cancel()

	defer func() {
		defer r.wg.Done()

		job.FinishedAt = r.clock()

		if ctx.Err() == context.DeadlineExceeded {
			job.Error = "Generating diff took longer than expected. Aborted."
		}

		// We're passing a new context here because we want to persist the job
		// even if we ran into a timeout earlier.
		err := r.store.UpdateCampaignJob(context.Background(), job)
		if err != nil {
			log15.Error("UpdateCampaignJob failed", "err", err)
		}
	}()

	// Check whether CampaignPlan has been canceled.
	p, err := r.store.GetCampaignPlan(ctx, GetCampaignPlanOpts{ID: job.CampaignPlanID})
	if err != nil {
		job.Error = err.Error()
		return
	}
	if !p.CanceledAt.IsZero() {
		job.Error = "Campaign execution canceled."
		return
	}

	job.StartedAt = r.clock()

	// We load the repository here again so that we decouple the
	// creation and running of jobs from the start.
	reposStore := repos.NewDBStore(r.store.DB(), sql.TxOptions{})
	opts := repos.StoreListReposArgs{IDs: []uint32{uint32(job.RepoID)}}
	rs, err := reposStore.ListRepos(ctx, opts)
	if err != nil {
		job.Error = err.Error()
		return
	}
	if len(rs) != 1 {
		job.Error = fmt.Sprintf("repository %d not found", job.RepoID)
		return
	}

	diff, desc, err := r.ct.generateDiff(ctx, api.RepoName(rs[0].Name), api.CommitID(job.Rev))
	if err != nil {
		job.Error = err.Error()
	}

	job.Diff = diff
	job.Description = desc
}

func (r *Runner) createPlanAndJobs(
	ctx context.Context,
	plan *a8n.CampaignPlan,
	rs []*graphqlbackend.RepositoryResolver,
) ([]*a8n.CampaignJob, error) {
	var (
		err error
		tx  *Store
	)
	tx, err = r.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Done(&err)

	err = tx.CreateCampaignPlan(ctx, plan)
	if err != nil {
		return nil, err
	}

	jobs := make([]*a8n.CampaignJob, 0, len(rs))
	for _, repo := range rs {
		if !a8n.IsRepoSupported(repo.ExternalRepo()) {
			continue
		}

		var repoID int32
		if err = relay.UnmarshalSpec(repo.ID(), &repoID); err != nil {
			return jobs, err
		}

		var (
			branch string
			rev    api.CommitID
		)
		branch, rev, err = r.defaultBranch(ctx, repo)
		if err == ErrNoDefaultBranch {
			err = nil
			continue
		}
		if err != nil {
			return jobs, err
		}

		job := &a8n.CampaignJob{
			CampaignPlanID: plan.ID,
			RepoID:         repoID,
			BaseRef:        branch,
			Rev:            rev,
		}
		if err = tx.CreateCampaignJob(ctx, job); err != nil {
			return jobs, err
		}
		jobs = append(jobs, job)
	}

	return jobs, err
}

// Wait blocks until all CampaignJobs created and started by Start have
// finished.
func (r *Runner) Wait() error {
	if !r.started {
		return errors.New("not started")
	}

	r.wg.Wait()

	return nil
}
