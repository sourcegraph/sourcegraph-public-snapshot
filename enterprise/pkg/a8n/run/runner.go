package run

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ee "github.com/sourcegraph/sourcegraph/enterprise/pkg/a8n"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/env"
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
	store    *ee.Store
	search   repoSearch
	commitID repoCommitID
	clock    func() time.Time

	ct CampaignType

	started bool
	wg      sync.WaitGroup
}

// repoSearch takes in a raw search query and returns the list of repositories
// associated with the search results.
type repoSearch func(ctx context.Context, query string) ([]*graphqlbackend.RepositoryResolver, error)

// repoCommitID takes in a RepositoryResolver and returns the target commit ID
// of the repository's default branch.
type repoCommitID func(ctx context.Context, repo *graphqlbackend.RepositoryResolver) (api.CommitID, error)

// ErrNoDefaultBranch is returned by a repoCommitID when no default branch
// could be determined for a given repo.
var ErrNoDefaultBranch = errors.New("could not determine default branch")

// defaultRepoCommitID is an implementation of repoCommit that uses methods
// defined on RepositoryResolver to talk to gitserver to determine a
// repository's default branch target commit ID.
var defaultRepoCommitID = func(ctx context.Context, repo *graphqlbackend.RepositoryResolver) (api.CommitID, error) {
	var commitID api.CommitID

	defaultBranch, err := repo.DefaultBranch(ctx)
	if err != nil {
		return commitID, err
	}
	if defaultBranch == nil {
		return commitID, ErrNoDefaultBranch
	}

	commit, err := defaultBranch.Target().Commit(ctx)
	if err != nil {
		return commitID, err
	}

	commitID = api.CommitID(commit.OID())
	return commitID, nil
}

// New returns a Runner for a given CampaignType.
func New(store *ee.Store, ct CampaignType, search repoSearch, commitID repoCommitID) *Runner {
	return NewWithClock(store, ct, search, commitID, func() time.Time {
		return time.Now().UTC().Truncate(time.Microsecond)
	})
}

// NewWithClock returns a Runner for a given CampaignType with the given clock used
// to generate timestamps
func NewWithClock(store *ee.Store, ct CampaignType, search repoSearch, commitID repoCommitID, clock func() time.Time) *Runner {
	runner := &Runner{
		store:    store,
		search:   search,
		commitID: commitID,
		ct:       ct,
		clock:    clock,
	}
	if runner.commitID == nil {
		runner.commitID = defaultRepoCommitID
	}

	return runner
}

// Run executes the CampaignPlan by searching for relevant repositories using
// the CampaignType specific searchQuery and then executing CampaignJobs for
// each repository.
// Before it starts executing CampaignJobs it persists the CampaignPlan and the
// new CampaignJobs in a transaction.
// What each CampaignJob then does in each repository depends on the
// CampaignType set on CampaignPlan.
// This is a non-blocking method that will possibly return before all
// CampaignJobs are finished.
func (r *Runner) Run(ctx context.Context, plan *a8n.CampaignPlan) error {
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
		return ErrTooManyResults
	}

	jobs, err := r.createPlanAndJobs(ctx, plan, rs)
	if err != nil {
		return err
	}

	numWorkers, err := strconv.ParseInt(maxWorkers, 10, 64)
	if err != nil {
		return err
	}

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

func (r *Runner) runJob(ctx context.Context, job *a8n.CampaignJob) {
	defer func() {
		defer r.wg.Done()

		job.FinishedAt = r.clock()

		// We're passing a new context here because we want to persist the job
		// even if we ran into a timeout earlier.
		err := r.store.UpdateCampaignJob(context.Background(), job)
		if err != nil {
			log15.Error("UpdateCampaignJob failed", "err", err)
		}
	}()

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

	diff, err := r.ct.generateDiff(ctx, api.RepoName(rs[0].Name), api.CommitID(job.Rev))
	if err != nil {
		job.Error = err.Error()
	}

	job.Diff = diff
}

func (r *Runner) createPlanAndJobs(
	ctx context.Context,
	plan *a8n.CampaignPlan,
	rs []*graphqlbackend.RepositoryResolver,
) ([]*a8n.CampaignJob, error) {
	var (
		err error
		tx  *ee.Store
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

		var rev api.CommitID
		rev, err = r.commitID(ctx, repo)
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
