package campaigns

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/neelance/parallel"
	"github.com/pkg/errors"
)

type ActionRepoStatus struct {
	Cached bool

	LogFile    string
	EnqueuedAt time.Time
	StartedAt  time.Time
	FinishedAt time.Time

	Patch PatchInput
	Err   error
}

type ExecutorOpts struct {
	Endpoint    string
	AccessToken string

	KeepLogs bool
	Timeout  time.Duration

	ClearCache bool
	Cache      ExecutionCache
}

type Executor struct {
	action Action
	opt    ExecutorOpts

	reposMu sync.Mutex
	repos   map[ActionRepo]ActionRepoStatus

	par           *parallel.Run
	doneEnqueuing chan struct{}

	logger *ActionLogger
}

func NewExecutor(action Action, parallelism int, logger *ActionLogger, opt ExecutorOpts) *Executor {
	if opt.Cache == nil {
		opt.Cache = ExecutionNoOpCache{}
	}

	return &Executor{
		action: action,
		opt:    opt,
		repos:  map[ActionRepo]ActionRepoStatus{},
		par:    parallel.NewRun(parallelism),
		logger: logger,

		doneEnqueuing: make(chan struct{}),
	}
}

func (x *Executor) EnqueueRepo(repo ActionRepo) {
	x.updateRepoStatus(repo, ActionRepoStatus{EnqueuedAt: time.Now()})
}

func (x *Executor) updateRepoStatus(repo ActionRepo, status ActionRepoStatus) {
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

func (x *Executor) AllPatches() []PatchInput {
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

func (x *Executor) Start(ctx context.Context) {
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

func (x *Executor) Wait() error {
	<-x.doneEnqueuing
	return x.par.Wait()
}

func (x *Executor) do(ctx context.Context, repo ActionRepo) (err error) {
	// Check if cached.
	cacheKey := ExecutionCacheKey{Repo: repo, Runs: x.action.Steps}
	if x.opt.ClearCache {
		if err := x.opt.Cache.Clear(ctx, cacheKey); err != nil {
			return errors.Wrapf(err, "clearing cache for %s", repo.Name)
		}
	} else {
		if result, ok, err := x.opt.Cache.Get(ctx, cacheKey); err != nil {
			return errors.Wrapf(err, "checking cache for %s", repo.Name)
		} else if ok {
			status := ActionRepoStatus{Cached: true, Patch: result}
			x.updateRepoStatus(repo, status)
			x.logger.RepoCacheHit(repo, len(x.action.Steps), status.Patch != PatchInput{})
			return nil
		}
	}

	prefix := "action-" + strings.Replace(strings.Replace(repo.Name, "/", "-", -1), "github.com-", "", -1)

	logFileName, err := x.logger.AddRepo(repo)
	if err != nil {
		return errors.Wrapf(err, "failed to setup logging for repo %s", repo.Name)
	}

	x.updateRepoStatus(repo, ActionRepoStatus{
		LogFile:   logFileName,
		StartedAt: time.Now(),
	})

	runCtx, cancel := context.WithTimeout(ctx, x.opt.Timeout)
	defer cancel()

	patch, err := runAction(runCtx, x.opt.Endpoint, x.opt.AccessToken, prefix, repo.Name, repo.Rev, x.action.Steps, x.logger)
	status := ActionRepoStatus{
		FinishedAt: time.Now(),
	}
	if len(patch) > 0 {
		status.Patch = PatchInput{
			Repository:   repo.ID,
			BaseRevision: repo.Rev,
			BaseRef:      repo.BaseRef,
			Patch:        string(patch),
		}
	}
	if err != nil {
		if reachedTimeout(runCtx, err) {
			err = &errTimeoutReached{timeout: x.opt.Timeout}
		}
		status.Err = err
	}

	x.updateRepoStatus(repo, status)
	lerr := x.logger.RepoFinished(repo.Name, len(patch) > 0, err)
	if lerr != nil {
		return lerr
	}

	// Add to cache if successful.
	if err == nil {
		// We don't use runCtx here because we want to write to the cache even
		// if we've now reached the timeout
		if err := x.opt.Cache.Set(ctx, cacheKey, status.Patch); err != nil {
			return errors.Wrapf(err, "caching result for %s", repo.Name)
		}
	}

	return err
}

type errTimeoutReached struct{ timeout time.Duration }

func (e *errTimeoutReached) Error() string {
	return fmt.Sprintf("Timeout reached. Execution took longer than %s.", e.timeout)
}

func reachedTimeout(cmdCtx context.Context, err error) bool {
	if ee, ok := errors.Cause(err).(*exec.ExitError); ok {
		if ee.String() == "signal: killed" && cmdCtx.Err() == context.DeadlineExceeded {
			return true
		}
	}

	return errors.Is(err, context.DeadlineExceeded)
}
