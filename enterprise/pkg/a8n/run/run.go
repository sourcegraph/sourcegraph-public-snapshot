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
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

type Runner interface {
	Start(context.Context) error
	Wait() error

	validate() error
}

type repoSearch func(ctx context.Context, query string) ([]*graphqlbackend.RepositoryResolver, error)

func New(store *ee.Store, plan *a8n.CampaignPlan, search repoSearch) (Runner, error) {
	var runner Runner

	switch strings.ToLower(plan.CampaignType) {
	case "comby":
		runner = &comby{store: store, plan: plan, search: search}
	default:
		return nil, fmt.Errorf("unknown campaign type: %s", plan.CampaignType)
	}

	return runner, runner.validate()
}

var combyServiceURL string

func init() {
	combyServiceURL = env.Get("COMBY_URL", "http://replacer:3185", "replacer server URL")
}

type combyArgs struct {
	ScopeQuery      string `json:"scopeQuery"`
	MatchTemplate   string `json:"matchTemplate"`
	RewriteTemplate string `json:"rewriteTemplate"`
}

type comby struct {
	store  *ee.Store
	plan   *a8n.CampaignPlan
	search repoSearch

	args combyArgs

	started bool
	wg      sync.WaitGroup
}

func (c *comby) validate() error {
	var args combyArgs

	if err := jsonc.Unmarshal(c.plan.Arguments, &args); err != nil {
		return err
	}

	if args.ScopeQuery == "" {
		return errors.New("missing argument in specification: scopeQuery")
	}

	if args.MatchTemplate == "" {
		return errors.New("missing argument in specification: matchTemplate")
	}

	if args.RewriteTemplate == "" {
		return errors.New("missing argument in specification: rewriteTemplate")
	}

	return nil
}

func (c *comby) Start(ctx context.Context) error {
	c.started = true

	repos, err := c.search(ctx, c.args.ScopeQuery)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, repo := range repos {
		job := &a8n.CampaignJob{
			CampaignPlanID: c.plan.ID,
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

		err = c.store.CreateCampaignJob(ctx, job)
		if err != nil {
			return err
		}

		wg.Add(1)
		go func(plan *a8n.CampaignPlan, job *a8n.CampaignJob) {
			// TODO(a8n): Do real work.
			// Send request to service with Repo, Ref, Arguments.
			// Receive diff.
			job.Diff = bogusDiff
			job.FinishedAt = time.Now()

			err := c.store.UpdateCampaignJob(ctx, job)
			if err != nil {
				log15.Error("RunCampaign.UpdateCampaignJob failed", "err", err)
			}

			wg.Done()
		}(c.plan, job)
	}

	return nil
}

func (c *comby) Wait() error {
	if !c.started {
		return errors.New("not started")
	}

	c.wg.Wait()

	return nil
}

const bogusDiff = `diff --git a/README.md b/README.md
index 323fae0..34a3ec2 100644
--- a/README.md
+++ b/README.md
@@ -1 +1 @@
-foobar
+barfoo
`
