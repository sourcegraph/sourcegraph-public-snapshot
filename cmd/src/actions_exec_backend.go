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

	clearCache bool
	cache      actionExecutionCache
}

type actionExecutor struct {
	action Action
	opt    actionExecutorOptions

	reposMu sync.Mutex
	repos   map[ActionRepo]ActionRepoStatus

	par           *parallel.Run
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
	if status.Patch == (PatchInput{}) {
		status.Patch = prev.Patch
	}
	if status.Err == nil {
		status.Err = prev.Err
	}

	x.repos[repo] = status
}

func (x *actionExecutor) allPatches() []PatchInput {
	patches := make([]PatchInput, 0, len(x.repos))
	x.reposMu.Lock()
	defer x.reposMu.Unlock()
	for _, status := range x.repos {
		if patch := status.Patch; patch != (PatchInput{}) && status.Err == nil {
			patches = append(patches, status.Patch)
		}
	}
	return patches
}

func (x *actionExecutor) start(ctx context.Context) {
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
	return x.par.Wait()
}
