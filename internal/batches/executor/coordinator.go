package executor

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution/cache"

	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/docker"
	"github.com/sourcegraph/src-cli/internal/batches/log"
	"github.com/sourcegraph/src-cli/internal/batches/repozip"
	"github.com/sourcegraph/src-cli/internal/batches/workspace"
)

type taskExecutor interface {
	Start(context.Context, []*Task, TaskExecutionUI)
	Wait(context.Context) ([]taskResult, error)
}

// Coordinator coordinates the execution of Tasks. It makes use of an executor,
// checks the ExecutionCache whether execution is necessary, builds
// batcheslib.ChangesetSpecs out of the executionResults.
type Coordinator struct {
	opts NewCoordinatorOpts

	cache      cache.Cache
	exec       taskExecutor
	logManager log.LogManager
}

type imageEnsurer func(ctx context.Context, name string) (docker.Image, error)

type NewCoordinatorOpts struct {
	// Dependencies
	EnsureImage         imageEnsurer
	Creator             workspace.Creator
	Cache               cache.Cache
	RepoArchiveRegistry repozip.ArchiveRegistry

	GlobalEnv []string

	// Everything that follows are either command-line flags or features.

	// Used by batcheslib.BuildChangesetSpecs
	Features batches.FeatureFlags

	Parallelism int
	Timeout     time.Duration
	TempDir     string
	IsRemote    bool
}

func NewCoordinator(opts NewCoordinatorOpts, logger log.LogManager) *Coordinator {
	exec := newExecutor(newExecutorOpts{
		RepoArchiveRegistry: opts.RepoArchiveRegistry,
		EnsureImage:         opts.EnsureImage,
		Creator:             opts.Creator,
		Logger:              logger,

		Parallelism: opts.Parallelism,
		Timeout:     opts.Timeout,
		TempDir:     opts.TempDir,
		IsRemote:    opts.IsRemote,
		GlobalEnv:   opts.GlobalEnv,
	})

	return &Coordinator{
		opts:       opts,
		cache:      opts.Cache,
		exec:       exec,
		logManager: logger,
	}
}

// CheckCache checks whether the internal ExecutionCache contains
// ChangesetSpecs for the given Tasks. If cached ChangesetSpecs exist, those
// are returned, otherwise the Task, to be executed later.
func (c *Coordinator) CheckCache(ctx context.Context, batchSpec *batcheslib.BatchSpec, tasks []*Task) (uncached []*Task, specs []*batcheslib.ChangesetSpec, err error) {
	for _, t := range tasks {
		cachedSpecs, found, err := c.checkCacheForTask(ctx, batchSpec, t)
		if err != nil {
			return nil, nil, err
		}

		if !found {
			uncached = append(uncached, t)
			continue
		}

		specs = append(specs, cachedSpecs...)
	}

	return uncached, specs, nil
}

// CheckStepResultsCache checks the cache for each Task, but only for cached
// step results. This is used by `src batch exec` when executing server-side.
func (c *Coordinator) CheckStepResultsCache(ctx context.Context, tasks []*Task, globalEnv []string) error {
	for _, t := range tasks {
		if err := c.loadCachedStepResults(ctx, t, globalEnv); err != nil {
			return err
		}
	}
	return nil
}

func (c *Coordinator) ClearCache(ctx context.Context, tasks []*Task) error {
	for _, task := range tasks {
		for i := len(task.Steps) - 1; i > -1; i-- {
			key := task.cacheKey(c.opts.GlobalEnv, c.opts.IsRemote, i)
			if err := c.cache.Clear(ctx, key); err != nil {
				return errors.Wrapf(err, "clearing cache for step %d in %q", i, task.Repository.Name)
			}
		}
	}
	return nil
}

func (c *Coordinator) checkCacheForTask(ctx context.Context, batchSpec *batcheslib.BatchSpec, task *Task) (specs []*batcheslib.ChangesetSpec, found bool, err error) {
	if err := c.loadCachedStepResults(ctx, task, c.opts.GlobalEnv); err != nil {
		return specs, false, err
	}

	// If we have cached results and don't need to execute any more steps,
	// we build changeset specs and return.
	// TODO: This doesn't consider skipped steps.
	if task.CachedStepResultFound && task.CachedStepResult.StepIndex == len(task.Steps)-1 {
		// If the cached result resulted in an empty diff, we don't need to
		// add it to the list of specs that are displayed to the user and
		// send to the server. Instead, we can just report that the task is
		// complete and move on.
		if task.CachedStepResult.Diff == "" {
			return specs, true, nil
		}

		specs, err = c.buildChangesetSpecs(task, batchSpec, task.CachedStepResult)
		return specs, true, err
	}

	return specs, false, nil
}

func (c Coordinator) buildChangesetSpecs(task *Task, batchSpec *batcheslib.BatchSpec, result execution.AfterStepResult) ([]*batcheslib.ChangesetSpec, error) {
	input := &batcheslib.ChangesetSpecInput{
		Repository: batcheslib.Repository{
			ID:          task.Repository.ID,
			Name:        task.Repository.Name,
			FileMatches: task.Repository.SortedFileMatches(),
			BaseRef:     task.Repository.BaseRef(),
			BaseRev:     task.Repository.Rev(),
		},
		BatchChangeAttributes: task.BatchChangeAttributes,
		Template:              batchSpec.ChangesetTemplate,
		TransformChanges:      batchSpec.TransformChanges,

		Result: execution.AfterStepResult{
			Diff:         result.Diff,
			ChangedFiles: result.ChangedFiles,
			Outputs:      result.Outputs,
		},
	}

	return batcheslib.BuildChangesetSpecs(input, batcheslib.ChangesetSpecFeatureFlags{
		IncludeAutoAuthorDetails: c.opts.Features.IncludeAutoAuthorDetails,
		AllowOptionalPublished:   c.opts.Features.AllowOptionalPublished,
	})
}

func (c *Coordinator) loadCachedStepResults(ctx context.Context, task *Task, globalEnv []string) error {
	// We start at the back so that we can find the _last_ cached step,
	// then restart execution on the following step.
	for i := len(task.Steps) - 1; i > -1; i-- {
		key := task.cacheKey(globalEnv, c.opts.IsRemote, i)

		result, found, err := c.cache.Get(ctx, key)
		if err != nil {
			return errors.Wrapf(err, "checking for cached diff for step %d", i)
		}

		// Found a cached result, we're done.
		if found {
			task.CachedStepResultFound = true
			task.CachedStepResult = result
			return nil
		}
	}

	return nil
}

func (c *Coordinator) buildSpecs(ctx context.Context, batchSpec *batcheslib.BatchSpec, taskResult taskResult, ui TaskExecutionUI) ([]*batcheslib.ChangesetSpec, error) {
	if len(taskResult.stepResults) == 0 {
		return nil, nil
	}

	lastStepResult := taskResult.stepResults[len(taskResult.stepResults)-1]

	// If the steps didn't result in any diff, we don't need to create a
	// changeset spec that's displayed to the user and send to the server.
	if lastStepResult.Diff == "" {
		return nil, nil
	}

	// Build the changeset specs.
	specs, err := c.buildChangesetSpecs(taskResult.task, batchSpec, lastStepResult)
	if err != nil {
		return nil, err
	}

	ui.TaskChangesetSpecsBuilt(taskResult.task, specs)
	return specs, nil
}

// Execute executes the given tasks. It calls the ui on updates.
func (c *Coordinator) Execute(ctx context.Context, tasks []*Task, ui TaskExecutionUI) error {
	_, err := c.doExecute(ctx, tasks, ui)

	return err
}

// ExecuteAndBuildSpecs executes the given tasks and builds changeset specs for the results.
// It calls the ui on updates.
func (c *Coordinator) ExecuteAndBuildSpecs(ctx context.Context, batchSpec *batcheslib.BatchSpec, tasks []*Task, ui TaskExecutionUI) ([]*batcheslib.ChangesetSpec, []string, error) {
	results, errs := c.doExecute(ctx, tasks, ui)

	var specs []*batcheslib.ChangesetSpec

	// Write results to cache, build ChangesetSpecs if possible and add to list.
	for _, taskResult := range results {
		// Don't build changeset specs for failed workspaces.
		if taskResult.err != nil {
			continue
		}

		taskSpecs, err := c.buildSpecs(ctx, batchSpec, taskResult, ui)
		if err != nil {
			return nil, nil, err
		}

		specs = append(specs, taskSpecs...)
	}

	return specs, c.logManager.LogFiles(), errs
}

func (c *Coordinator) doExecute(ctx context.Context, tasks []*Task, ui TaskExecutionUI) (results []taskResult, err error) {
	ui.Start(tasks)

	// Run executor
	c.exec.Start(ctx, tasks, ui)
	results, err = c.exec.Wait(ctx)

	// Write all step cache results for all results.
	for _, res := range results {
		for _, stepRes := range res.stepResults {
			cacheKey := res.task.cacheKey(c.opts.GlobalEnv, c.opts.IsRemote, stepRes.StepIndex)
			if err := c.cache.Set(ctx, cacheKey, stepRes); err != nil {
				return nil, errors.Wrapf(err, "caching result for step %d", stepRes.StepIndex)
			}
		}
	}

	return results, err
}
