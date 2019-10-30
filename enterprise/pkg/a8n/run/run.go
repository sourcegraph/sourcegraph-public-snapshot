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

type Runner struct {
	store  *ee.Store
	plan   *a8n.CampaignPlan
	search repoSearch

	ct campaignType

	started bool
	wg      sync.WaitGroup
}

type campaignType interface {
	searchQuery() string
	validateArgs() error
	runJob(*a8n.CampaignJob)
}

type repoSearch func(ctx context.Context, query string) ([]*graphqlbackend.RepositoryResolver, error)

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

func (r *Runner) Start(ctx context.Context) error {
	r.started = true

	query := r.ct.searchQuery()
	repos, err := r.search(ctx, query)
	if err != nil {
		return err
	}

	jobs := make([]*a8n.CampaignJob, 0, len(repos))
	for _, repo := range repos {
		job := &a8n.CampaignJob{
			CampaignPlanID: r.plan.ID,
			StartedAt:      time.Now().UTC(),
		}

		err := relay.UnmarshalSpec(repo.ID(), &job.RepoID)
		if err != nil {
			return err
		}

		defaultBranch, err := repo.DefaultBranch(ctx)
		if err != nil {
			return err
		}
		if defaultBranch != nil {
			commit, err := defaultBranch.Target().Commit(ctx)
			if err != nil {
				return err
			}
			job.Rev = api.CommitID(commit.OID())
		}

		err = r.store.CreateCampaignJob(ctx, job)
		if err != nil {
			return err
		}
		jobs = append(jobs, job)
	}

	for _, job := range jobs {
		log15.Info("Launching job", "job", job.ID, "repo", job.RepoID)

		r.wg.Add(1)
		go func(job *a8n.CampaignJob) {
			r.ct.runJob(job)

			err := r.store.UpdateCampaignJob(ctx, job)
			if err != nil {
				log15.Error("UpdateCampaignJob failed", "err", err)
			}
			log15.Info("Job done", "job", job.ID)

			r.wg.Done()
		}(job)
	}

	return nil
}

func (r *Runner) Wait() error {
	if !r.started {
		return errors.New("not started")
	}

	r.wg.Wait()

	return nil
}
