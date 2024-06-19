package executor

import (
	"context"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution/cache"

	"github.com/sourcegraph/src-cli/internal/batches/log"
)

type taskExecutor interface {
	Start(context.Context, []*Task, TaskExecutionUI)
	Wait(context.Context) ([]taskResult, error)
}

// Coordinator coordinates the execution of Tasks. It makes use of an executor,
// checks the ExecutionCache whether execution is necessary, and builds
// batcheslib.ChangesetSpecs out of the executionResults.
type Coordinator struct {
	opts NewCoordinatorOpts

	exec taskExecutor
}

type NewCoordinatorOpts struct {
	ExecOpts NewExecutorOpts

	Cache       cache.Cache
	Logger      log.LogManager
	GlobalEnv   []string
	BinaryDiffs bool

	IsRemote bool
}

func NewCoordinator(opts NewCoordinatorOpts) *Coordinator {
	return &Coordinator{
		opts: opts,
		exec: NewExecutor(opts.ExecOpts),
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

func (c *Coordinator) ClearCache(ctx context.Context, tasks []*Task) error {
	for _, task := range tasks {
		for i := len(task.Steps) - 1; i > -1; i-- {
			key := task.CacheKey(c.opts.GlobalEnv, c.opts.ExecOpts.WorkingDirectory, i)
			if err := c.opts.Cache.Clear(ctx, key); err != nil {
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
		if len(task.CachedStepResult.Diff) == 0 {
			return specs, true, nil
		}

		specs, err = c.buildChangesetSpecs(task, batchSpec, task.CachedStepResult)
		return specs, true, err
	}

	return specs, false, nil
}

func (c *Coordinator) buildChangesetSpecs(task *Task, batchSpec *batcheslib.BatchSpec, result execution.AfterStepResult) ([]*batcheslib.ChangesetSpec, error) {
	version := 1
	if c.opts.BinaryDiffs {
		version = 2
	}
	input := &batcheslib.ChangesetSpecInput{
		Repository: batcheslib.Repository{
			ID:          task.Repository.ID,
			Name:        task.Repository.Name,
			FileMatches: task.Repository.SortedFileMatches(),
			BaseRef:     task.Repository.BaseRef(),
			BaseRev:     task.Repository.Rev(),
		},
		Path:                  task.Path,
		BatchChangeAttributes: task.BatchChangeAttributes,
		Template:              batchSpec.ChangesetTemplate,
		TransformChanges:      batchSpec.TransformChanges,

		Result: execution.AfterStepResult{
			Version:      version,
			Diff:         result.Diff,
			ChangedFiles: result.ChangedFiles,
			Outputs:      result.Outputs,
		},
	}

	return batcheslib.BuildChangesetSpecs(input, c.opts.BinaryDiffs, nil)
}

func (c *Coordinator) loadCachedStepResults(ctx context.Context, task *Task, globalEnv []string) error {
	// We start at the back so that we can find the _last_ cached step,
	// then restart execution on the following step.
	for i := len(task.Steps) - 1; i > -1; i-- {
		key := task.CacheKey(globalEnv, c.opts.ExecOpts.WorkingDirectory, i)

		result, found, err := c.opts.Cache.Get(ctx, key)
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
	if len(lastStepResult.Diff) == 0 {
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

// ExecuteAndBuildSpecs executes the given tasks and builds changeset specs for the results.
// It calls the ui on updates.
func (c *Coordinator) ExecuteAndBuildSpecs(ctx context.Context, batchSpec *batcheslib.BatchSpec, tasks []*Task, ui TaskExecutionUI) ([]*batcheslib.ChangesetSpec, []string, error) {
	ui.Start(tasks)

	// Run executor.
	c.exec.Start(ctx, tasks, ui)
	results, errs := c.exec.Wait(ctx)

	// Write all step cache results to the cache.
	for _, res := range results {
		for _, stepRes := range res.stepResults {
			cacheKey := res.task.CacheKey(c.opts.GlobalEnv, c.opts.ExecOpts.WorkingDirectory, stepRes.StepIndex)
			if err := c.opts.Cache.Set(ctx, cacheKey, stepRes); err != nil {
				return nil, nil, errors.Wrapf(err, "caching result for step %d", stepRes.StepIndex)
			}
		}
	}

	var specs []*batcheslib.ChangesetSpec

	// Build ChangesetSpecs if possible and add to list.
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

	return specs, c.opts.Logger.LogFiles(), errs
}
