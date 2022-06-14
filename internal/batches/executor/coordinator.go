package executor

import (
	"context"
	"os"
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

	// Everything that follows are either command-line flags or features.

	// TODO: We could probably have a wrapper around flags and features,
	// something like ExecutionArgs, that we can pass around
	CacheDir   string
	SkipErrors bool

	// Used by batcheslib.BuildChangesetSpecs
	Features batches.FeatureFlags

	CleanArchives   bool
	Parallelism     int
	Timeout         time.Duration
	KeepLogs        bool
	TempDir         string
	AllowPathMounts bool
}

func NewCoordinator(opts NewCoordinatorOpts) *Coordinator {
	logManager := log.NewManager(opts.TempDir, opts.KeepLogs)

	globalEnv := os.Environ()

	exec := newExecutor(newExecutorOpts{
		RepoArchiveRegistry: opts.RepoArchiveRegistry,
		EnsureImage:         opts.EnsureImage,
		Creator:             opts.Creator,
		Logger:              logManager,

		Parallelism:     opts.Parallelism,
		Timeout:         opts.Timeout,
		TempDir:         opts.TempDir,
		AllowPathMounts: opts.AllowPathMounts,
		WriteStepCacheResult: func(ctx context.Context, stepResult execution.AfterStepResult, task *Task) error {
			// Temporarily skip writing to the cache if a mount is present
			for _, step := range task.Steps {
				if len(step.Mount) > 0 {
					return nil
				}
			}
			cacheKey := task.cacheKey(globalEnv)
			return writeToCache(ctx, opts.Cache, stepResult, task, cacheKey)
		},
	})

	return &Coordinator{
		opts:       opts,
		cache:      opts.Cache,
		exec:       exec,
		logManager: logManager,
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
func (c *Coordinator) CheckStepResultsCache(ctx context.Context, tasks []*Task) error {
	globalEnv := os.Environ()
	for _, t := range tasks {
		if err := c.loadCachedStepResults(ctx, t, globalEnv); err != nil {
			return err
		}
	}
	return nil
}

func (c *Coordinator) ClearCache(ctx context.Context, tasks []*Task) error {
	globalEnv := os.Environ()

	for _, task := range tasks {
		cacheKey := task.cacheKey(globalEnv)
		if err := c.cache.Clear(ctx, cacheKey); err != nil {
			return errors.Wrapf(err, "clearing cache for %q", task.Repository.Name)
		}
		for i := len(task.Steps) - 1; i > -1; i-- {
			key := cacheKeyForStep(cacheKey, i)
			if err := c.cache.Clear(ctx, key); err != nil {
				return errors.Wrapf(err, "clearing cache for step %d in %q", i, task.Repository.Name)
			}
		}
	}
	return nil
}

func (c *Coordinator) checkCacheForTask(ctx context.Context, batchSpec *batcheslib.BatchSpec, task *Task) (specs []*batcheslib.ChangesetSpec, found bool, err error) {
	globalEnv := os.Environ()

	// Check if the task is cached.
	cacheKey := task.cacheKey(globalEnv)

	var result execution.Result
	result, found, err = c.cache.Get(ctx, cacheKey)
	if err != nil {
		return specs, false, errors.Wrapf(err, "checking cache for %q", task.Repository.Name)
	}

	if !found {
		// If we are here, that means we didn't find anything in the cache for the
		// complete task. So, what if we have cached results for the steps?
		if err := c.loadCachedStepResults(ctx, task, globalEnv); err != nil {
			return specs, false, err
		}

		return specs, false, nil
	}

	// If the cached result resulted in an empty diff, we don't need to
	// add it to the list of specs that are displayed to the user and
	// send to the server. Instead, we can just report that the task is
	// complete and move on.
	if result.Diff == "" {
		return specs, true, nil
	}

	specs, err = c.buildChangesetSpecs(task, batchSpec, result)
	if err != nil {
		return specs, false, err
	}

	return specs, true, nil
}

func (c Coordinator) buildChangesetSpecs(task *Task, batchSpec *batcheslib.BatchSpec, result execution.Result) ([]*batcheslib.ChangesetSpec, error) {
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

		Result: execution.Result{
			Diff:         result.Diff,
			ChangedFiles: result.ChangedFiles,
			Outputs:      result.Outputs,
			Path:         result.Path,
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
	taskKey := task.cacheKey(globalEnv)
	for i := len(task.Steps) - 1; i > -1; i-- {
		key := cacheKeyForStep(taskKey, i)

		result, found, err := c.cache.GetStepResult(ctx, key)
		if err != nil {
			return errors.Wrapf(err, "checking for cached diff for step %d", i)
		}

		// Found a cached result, we're done
		if found {
			task.CachedResultFound = true
			task.CachedResult = result
			return nil
		}
	}

	return nil
}

func writeToCache(ctx context.Context, cache cache.Cache, stepResult execution.AfterStepResult, task *Task, cacheKey *cache.ExecutionKeyWithGlobalEnv) error {
	key := cacheKeyForStep(cacheKey, stepResult.StepIndex)
	if err := cache.SetStepResult(ctx, key, stepResult); err != nil {
		return errors.Wrapf(err, "caching result for step %d in %q", stepResult.StepIndex, task.Repository.Name)
	}

	return nil
}

func (c *Coordinator) writeExecutionCacheResult(ctx context.Context, taskResult taskResult, ui TaskExecutionUI) error {
	// Add to the cache, even if no diff was produced.
	globalEnv := os.Environ()
	cacheKey := taskResult.task.cacheKey(globalEnv)
	if err := c.cache.Set(ctx, cacheKey, taskResult.result); err != nil {
		return errors.Wrapf(err, "caching result for %q", taskResult.task.Repository.Name)
	}

	return nil
}

func (c *Coordinator) writeCacheAndBuildSpecs(ctx context.Context, batchSpec *batcheslib.BatchSpec, taskResult taskResult, ui TaskExecutionUI) ([]*batcheslib.ChangesetSpec, error) {
	// Temporarily prevent writing to the cache when running a spec with a mount. Caching does not at the moment "know"
	// when a file that is being mounted has changed. This causes the execution not to re-run if a mounted file changes.
	hasMount := false
	for _, step := range batchSpec.Steps {
		if len(step.Mount) > 0 {
			hasMount = true
			break
		}
	}
	if !hasMount {
		c.writeExecutionCacheResult(ctx, taskResult, ui)
	}

	// If the steps didn't result in any diff, we don't need to create a
	// changeset spec that's displayed to the user and send to the server.
	if taskResult.result.Diff == "" {
		return nil, nil
	}

	// Build the changeset specs.
	specs, err := c.buildChangesetSpecs(taskResult.task, batchSpec, taskResult.result)
	if err != nil {
		return nil, err
	}

	ui.TaskChangesetSpecsBuilt(taskResult.task, specs)
	return specs, nil
}

// Execute executes the given tasks. It calls the ui on updates.
func (c *Coordinator) Execute(ctx context.Context, tasks []*Task, ui TaskExecutionUI) error {
	results, err := c.doExecute(ctx, tasks, ui)

	// Write results to cache.
	for _, taskResult := range results {
		if cacheErr := c.writeExecutionCacheResult(ctx, taskResult, ui); cacheErr != nil {
			return cacheErr
		}
	}

	return err
}

// ExecuteAndBuildSpecs executes the given tasks and builds changeset specs for the results.
// It calls the ui on updates.
func (c *Coordinator) ExecuteAndBuildSpecs(ctx context.Context, batchSpec *batcheslib.BatchSpec, tasks []*Task, ui TaskExecutionUI) ([]*batcheslib.ChangesetSpec, []string, error) {
	results, errs := c.doExecute(ctx, tasks, ui)

	var specs []*batcheslib.ChangesetSpec

	// Write results to cache, build ChangesetSpecs if possible and add to list.
	for _, taskResult := range results {
		taskSpecs, err := c.writeCacheAndBuildSpecs(ctx, batchSpec, taskResult, ui)
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
	return c.exec.Wait(ctx)
}
