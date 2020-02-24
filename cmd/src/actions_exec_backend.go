package main

import (
	"context"
	"sync"
	"time"

	"github.com/neelance/parallel"
)

type actionExecutorOptions struct {
	keepLogs bool
	timeout  time.Duration

	cache    actionExecutionCache
	onUpdate func(map[ActionRepo]ActionRepoStatus)
}

type actionExecutor struct {
	action Action
	opt    actionExecutorOptions

	reposMu sync.Mutex
	repos   map[ActionRepo]ActionRepoStatus

	par           *parallel.Run
	done          chan struct{}
	doneEnqueuing chan struct{}

	logger *actionLogger
}

func newActionExecutor(action Action, parallelism int, logger *actionLogger, opt actionExecutorOptions) *actionExecutor {
	if opt.cache == nil {
		opt.cache = actionExecutionNoOpCache{}
	}

	return &actionExecutor{
		action: action,
		opt:    opt,
		repos:  map[ActionRepo]ActionRepoStatus{},
		par:    parallel.NewRun(parallelism),
		logger: logger,

		done:          make(chan struct{}),
		doneEnqueuing: make(chan struct{}),
	}
}

func (x *actionExecutor) enqueueRepo(repo ActionRepo) {
	x.updateRepoStatus(repo, ActionRepoStatus{EnqueuedAt: time.Now()})
}

func (x *actionExecutor) updateRepoStatus(repo ActionRepo, status ActionRepoStatus) {
	x.reposMu.Lock()
	defer x.reposMu.Unlock()

	// Perform delta update.
	prev := x.repos[repo]
	if status.LogFile == "" {
		status.LogFile = prev.LogFile
	}
	if status.EnqueuedAt.IsZero() {
		status.EnqueuedAt = prev.EnqueuedAt
	}
	if status.StartedAt.IsZero() {
		status.StartedAt = prev.StartedAt
	}
	if status.FinishedAt.IsZero() {
		status.FinishedAt = prev.FinishedAt
	}
	if status.Patch == (CampaignPlanPatch{}) {
		status.Patch = prev.Patch
	}
	if status.Err == nil {
		status.Err = prev.Err
	}

	x.repos[repo] = status

	if x.opt.onUpdate != nil {
		x.opt.onUpdate(x.repos)
	}
}

func (x *actionExecutor) allPatches() []CampaignPlanPatch {
	patches := make([]CampaignPlanPatch, 0, len(x.repos))
	x.reposMu.Lock()
	defer x.reposMu.Unlock()
	for _, repoStatus := range x.repos {
		if patch := repoStatus.Patch; patch != (CampaignPlanPatch{}) {
			patches = append(patches, repoStatus.Patch)
		}
	}
	return patches
}

func (x *actionExecutor) start(ctx context.Context) {
	if x.opt.onUpdate != nil {
		go func() {
			for {
				select {
				case <-x.done:
					return
				default:
				}

				x.reposMu.Lock()
				x.opt.onUpdate(x.repos)
				x.reposMu.Unlock()
				time.Sleep(50 * time.Millisecond)
			}
		}()

	}

	x.reposMu.Lock()
	allRepos := make([]ActionRepo, 0, len(x.repos))
	for repo := range x.repos {
		allRepos = append(allRepos, repo)
	}
	x.reposMu.Unlock()

	for _, repo := range allRepos {
		x.par.Acquire()
		go func(repo ActionRepo) {
			defer x.par.Release()
			err := x.do(ctx, repo)
			if err != nil {
				x.par.Error(err)
			}
		}(repo)
	}

	close(x.doneEnqueuing)
}

func (x *actionExecutor) wait() error {
	<-x.doneEnqueuing
	err := x.par.Wait()
	close(x.done)
	return err
}
