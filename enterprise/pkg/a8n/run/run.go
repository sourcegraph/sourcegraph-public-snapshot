package run

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	ee "github.com/sourcegraph/sourcegraph/enterprise/pkg/a8n"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/api"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// A Runner executes a CampaignPlan by creating and running CampaignJobs
// according to the CampaignPlan's Arguments and CampaignType.
type Runner struct {
	store  *ee.Store
	plan   *a8n.CampaignPlan
	search repoSearch

	ct campaignType

	started bool
	wg      sync.WaitGroup
}

// A campaignType provides a search query, argument validation and execution of
// job for a specific CampaignType specified on CampaignPlan.
type campaignType interface {
	searchQuery() string
	validateArgs() error
	runJob(*a8n.CampaignJob)
}

// repoSearch takes in a raw search query and returns the list of repositories
// associated with the search results.
type repoSearch func(ctx context.Context, query string) ([]*graphqlbackend.RepositoryResolver, error)

// New returns a Runner for the given CampaignPlan. It validates the
// CampaignPlan's Arguments according to its CampaignType.
func New(store *ee.Store, plan *a8n.CampaignPlan, search repoSearch) (*Runner, error) {
	var ct campaignType

	switch strings.ToLower(plan.CampaignType) {
	case "comby":
		ct = &comby{plan: plan}
	default:
		return nil, fmt.Errorf("unknown campaign type: %s", plan.CampaignType)
	}

	err := ct.validateArgs()
	if err != nil {
		return nil, err
	}

	return &Runner{store: store, plan: plan, search: search, ct: ct}, nil
}

// Start executes the CampaignPlan set on the Runner by first searching for
// relevant repositories and then creating and running CampaignJobs for each
// repository.
// What each CampaignJob does in each repository depends on the CampaignType
// set on CampaignPlan.
// This is a non-blocking method that will possibly return before all
// CampaignJobs are finished.
func (r *Runner) Start(ctx context.Context) error {
	if r.started {
		return errors.New("already started")
	}
	r.started = true

	repos, err := r.search(ctx, r.ct.searchQuery())
	if err != nil {
		return err
	}

	jobs := make([]*a8n.CampaignJob, 0, len(repos))
	for _, repo := range repos {
		// TODO(a8n): Do we want to persist a failed job instead
		// of returning an error here?
		var repoID int32
		if err := relay.UnmarshalSpec(repo.ID(), &repoID); err != nil {
			return err
		}

		defaultBranch, err := repo.DefaultBranch(ctx)
		if err != nil {
			return err
		}
		if defaultBranch == nil {
			return fmt.Errorf("no default branch for %q", repo.Name())
		}

		commit, err := defaultBranch.Target().Commit(ctx)
		if err != nil {
			return err
		}
		rev := api.CommitID(commit.OID())

		job := &a8n.CampaignJob{
			CampaignPlanID: r.plan.ID,
			StartedAt:      time.Now().UTC(),
			RepoID:         repoID,
			Rev:            rev,
		}
		if err := r.store.CreateCampaignJob(ctx, job); err != nil {
			return err
		}
		jobs = append(jobs, job)
	}

	for _, job := range jobs {
		log15.Info("Launching job", "job", job.ID, "repo", job.RepoID)

		r.wg.Add(1)
		go func(job *a8n.CampaignJob) {
			r.ct.runJob(job)

			log15.Info("Job done", "job", job.ID)

			err := r.store.UpdateCampaignJob(ctx, job)
			if err != nil {
				log15.Error("UpdateCampaignJob failed", "err", err)
			}

			r.wg.Done()
		}(job)
	}

	return nil
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
