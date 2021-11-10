package executor

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"

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

// Coordinates coordinates the execution of Tasks. It makes use of an executor,
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

	CleanArchives bool
	Parallelism   int
	Timeout       time.Duration
	KeepLogs      bool
	TempDir       string
}

func NewCoordinator(opts NewCoordinatorOpts) *Coordinator {
	logManager := log.NewManager(opts.TempDir, opts.KeepLogs)

	exec := newExecutor(newExecutorOpts{
		RepoArchiveRegistry: opts.RepoArchiveRegistry,
		EnsureImage:         opts.EnsureImage,
		Creator:             opts.Creator,
		Logger:              logManager,

		Parallelism: opts.Parallelism,
		Timeout:     opts.Timeout,
		TempDir:     opts.TempDir,
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
func (c *Coordinator) CheckCache(ctx context.Context, tasks []*Task) (uncached []*Task, specs []*batcheslib.ChangesetSpec, err error) {
	for _, t := range tasks {
		cachedSpecs, found, err := c.checkCacheForTask(ctx, t)
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
		cacheKey := task.cacheKey()
		if err := c.cache.Clear(ctx, cacheKey); err != nil {
			return errors.Wrapf(err, "clearing cache for %q", task.Repository.Name)
		}
		for i := len(task.Steps) - 1; i > -1; i-- {
			key := cache.StepsCacheKey{ExecutionKey: task.cacheKey(), StepIndex: i}

			if err := c.cache.Clear(ctx, key); err != nil {
				return errors.Wrapf(err, "clearing cache for step %d in %q", i, task.Repository.Name)
			}
		}
	}
	return nil
}

func (c *Coordinator) checkCacheForTask(ctx context.Context, task *Task) (specs []*batcheslib.ChangesetSpec, found bool, err error) {
	// Check if the task is cached.
	cacheKey := task.cacheKey()

	var result execution.Result
	result, found, err = c.cache.Get(ctx, cacheKey)
	if err != nil {
		return specs, false, errors.Wrapf(err, "checking cache for %q", task.Repository.Name)
	}

	if !found {
		// If we are here, that means we didn't find anything in the cache for the
		// complete task. So, what if we have cached results for the steps?
		if err := c.loadCachedStepResults(ctx, task); err != nil {
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

	specs, err = c.buildChangesetSpecs(task, result)
	if err != nil {
		return specs, false, err
	}

	return specs, true, nil
}

func (c Coordinator) buildChangesetSpecs(task *Task, result execution.Result) ([]*batcheslib.ChangesetSpec, error) {
	input := &batcheslib.ChangesetSpecInput{
		BaseRepositoryID: task.Repository.ID,
		HeadRepositoryID: task.Repository.ID,
		Repository: batcheslib.ChangesetSpecRepository{
			Name:        task.Repository.Name,
			FileMatches: task.Repository.SortedFileMatches(),
			BaseRef:     task.Repository.BaseRef(),
			BaseRev:     task.Repository.Rev(),
		},
		BatchChangeAttributes: task.BatchChangeAttributes,
		Template:              task.Template,
		TransformChanges:      task.TransformChanges,

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

func (c *Coordinator) loadCachedStepResults(ctx context.Context, task *Task) error {
	// We start at the back so that we can find the _last_ cached step,
	// then restart execution on the following step.
	for i := len(task.Steps) - 1; i > -1; i-- {
		key := cache.StepsCacheKey{ExecutionKey: task.cacheKey(), StepIndex: i}

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

func (c *Coordinator) cacheAndBuildSpec(ctx context.Context, taskResult taskResult, ui TaskExecutionUI) ([]*batcheslib.ChangesetSpec, error) {
	// Add to the cache, even if no diff was produced.
	cacheKey := taskResult.task.cacheKey()
	if err := c.cache.Set(ctx, cacheKey, taskResult.result); err != nil {
		return nil, errors.Wrapf(err, "caching result for %q", taskResult.task.Repository.Name)
	}

	// Save the per-step results
	for _, stepResult := range taskResult.stepResults {
		key := cache.StepsCacheKey{
			ExecutionKey: taskResult.task.cacheKey(),
			StepIndex:    stepResult.StepIndex,
		}
		if err := c.cache.SetStepResult(ctx, key, stepResult); err != nil {
			return nil, errors.Wrapf(err, "caching result for step %d in %q", stepResult.StepIndex, taskResult.task.Repository.Name)
		}
	}

	// If the steps didn't result in any diff, we don't need to create a
	// changeset spec that's displayed to the user and send to the server.
	if taskResult.result.Diff == "" {
		return nil, nil
	}

	// Build the changeset specs.
	specs, err := c.buildChangesetSpecs(taskResult.task, taskResult.result)
	if err != nil {
		return nil, err
	}

	ui.TaskChangesetSpecsBuilt(taskResult.task, specs)
	return specs, nil
}

// Execute executes the given Tasks and the importChangeset statements in the
// given spec. It regularly calls the executionProgressPrinter with the
// current TaskStatuses.
func (c *Coordinator) Execute(ctx context.Context, tasks []*Task, spec *batcheslib.BatchSpec, ui TaskExecutionUI) ([]*batcheslib.ChangesetSpec, []string, error) {
	var (
		specs []*batcheslib.ChangesetSpec
		errs  *multierror.Error
	)

	ui.Start(tasks)

	// Run executor
	c.exec.Start(ctx, tasks, ui)
	results, err := c.exec.Wait(ctx)
	if err != nil {
		if c.opts.SkipErrors {
			errs = multierror.Append(errs, err)
		} else {
			return nil, nil, err
		}
	}

	// Write results to cache, build ChangesetSpecs if possible and add to list.
	for _, taskResult := range results {
		taskSpecs, err := c.cacheAndBuildSpec(ctx, taskResult, ui)
		if err != nil {
			return nil, nil, err
		}

		specs = append(specs, taskSpecs...)
	}

	return specs, c.logManager.LogFiles(), errs.ErrorOrNil()
}
