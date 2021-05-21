package executor

import (
	"context"
	"reflect"
	"strconv"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/log"
	"github.com/sourcegraph/src-cli/internal/batches/workspace"
)

type taskExecutor interface {
	Start(context.Context, []*Task, taskStatusHandler)
	Wait(context.Context) ([]taskResult, error)
}

// Coordinates coordinates the execution of Tasks. It makes use of an executor,
// checks the ExecutionCache whether execution is necessary, builds
// batches.ChangesetSpecs out of the executionResults.
type Coordinator struct {
	opts NewCoordinatorOpts

	cache      ExecutionCache
	exec       taskExecutor
	logManager *log.Manager
}

type repoNameResolver func(ctx context.Context, name string) (*graphql.Repository, error)

type NewCoordinatorOpts struct {
	// Dependencies
	ResolveRepoName repoNameResolver
	Creator         workspace.Creator
	Client          api.Client

	// Everything that follows are either command-line flags or features.

	// TODO: We could probably have a wrapper around flags and features,
	// something like ExecutionArgs, that we can pass around
	CacheDir   string
	ClearCache bool
	SkipErrors bool

	// Used by createChangesetSpecs
	AutoAuthorDetails bool

	CleanArchives bool
	Parallelism   int
	Timeout       time.Duration
	KeepLogs      bool
	TempDir       string
}

func NewCoordinator(opts NewCoordinatorOpts) *Coordinator {
	cache := NewCache(opts.CacheDir)
	logManager := log.NewManager(opts.TempDir, opts.KeepLogs)

	exec := newExecutor(newExecutorOpts{
		Fetcher: batches.NewRepoFetcher(opts.Client, opts.CacheDir, opts.CleanArchives),
		Creator: opts.Creator,
		Logger:  logManager,

		AutoAuthorDetails: opts.AutoAuthorDetails,
		Parallelism:       opts.Parallelism,
		Timeout:           opts.Timeout,
		TempDir:           opts.TempDir,
	})

	return &Coordinator{
		opts: opts,

		cache:      cache,
		exec:       exec,
		logManager: logManager,
	}
}

// CheckCache checks whether the internal ExecutionCache contains
// ChangesetSpecs for the given Tasks. If cached ChangesetSpecs exist, those
// are returned, otherwise the Task, to be executed later.
func (c *Coordinator) CheckCache(ctx context.Context, tasks []*Task) (uncached []*Task, specs []*batches.ChangesetSpec, err error) {
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

func (c *Coordinator) checkCacheForTask(ctx context.Context, task *Task) (specs []*batches.ChangesetSpec, found bool, err error) {
	// Check if the task is cached.
	cacheKey := task.cacheKey()
	if c.opts.ClearCache {
		if err := c.cache.Clear(ctx, cacheKey); err != nil {
			return specs, false, errors.Wrapf(err, "clearing cache for %q", task.Repository.Name)
		}

		return specs, false, nil
	}

	var result executionResult
	result, found, err = c.cache.Get(ctx, cacheKey)
	if err != nil {
		return specs, false, errors.Wrapf(err, "checking cache for %q", task.Repository.Name)
	}

	if !found {
		return specs, false, nil
	}

	// If the cached result resulted in an empty diff, we don't need to
	// add it to the list of specs that are displayed to the user and
	// send to the server. Instead, we can just report that the task is
	// complete and move on.
	if result.Diff == "" {
		return specs, true, nil
	}

	specs, err = createChangesetSpecs(task, result, c.opts.AutoAuthorDetails)
	if err != nil {
		return specs, false, err
	}

	return specs, true, nil
}

func (c *Coordinator) cacheAndBuildSpec(ctx context.Context, taskResult taskResult, status taskStatusHandler) (specs []*batches.ChangesetSpec, err error) {
	defer func() {
		// Set these two fields in any case
		status.Update(taskResult.task, func(status *TaskStatus) {
			status.ChangesetSpecsDone = true
			status.ChangesetSpecs = specs
		})
	}()

	// Add to the cache, even if no diff was produced.
	cacheKey := taskResult.task.cacheKey()
	if err := c.cache.Set(ctx, cacheKey, taskResult.result); err != nil {
		return nil, errors.Wrapf(err, "caching result for %q", taskResult.task.Repository.Name)
	}

	// If the steps didn't result in any diff, we don't need to create a
	// changeset spec that's displayed to the user and send to the server.
	if taskResult.result.Diff == "" {
		return nil, nil
	}

	// Build the changeset specs.
	specs, err = createChangesetSpecs(taskResult.task, taskResult.result, c.opts.AutoAuthorDetails)
	if err != nil {
		return specs, err
	}

	return specs, nil
}

type executionProgressPrinter func([]*TaskStatus)

// Execute executes the given Tasks and the importChangeset statements in the
// given spec. It regularly calls the executionProgressPrinter with the
// current TaskStatuses.
func (c *Coordinator) Execute(ctx context.Context, tasks []*Task, spec *batches.BatchSpec, printer executionProgressPrinter) ([]*batches.ChangesetSpec, []string, error) {
	var (
		specs []*batches.ChangesetSpec
		errs  *multierror.Error
	)

	// Start the goroutine that updates the UI
	status := NewTaskStatusCollection(tasks)

	done := make(chan struct{})
	if printer != nil {
		go func() {
			status.CopyStatuses(printer)

			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					status.CopyStatuses(printer)

				case <-done:
					return
				}
			}
		}()
	}

	// Run executor
	c.exec.Start(ctx, tasks, status)
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
		taskSpecs, err := c.cacheAndBuildSpec(ctx, taskResult, status)
		if err != nil {
			return nil, nil, err
		}

		specs = append(specs, taskSpecs...)
	}

	// Now that we've built the specs too we can mark the progress as done
	if printer != nil {
		status.CopyStatuses(printer)
		done <- struct{}{}
	}

	// Add external changeset specs.
	for _, ic := range spec.ImportChangesets {
		repo, err := c.opts.ResolveRepoName(ctx, ic.Repository)
		if err != nil {
			wrapped := errors.Wrapf(err, "resolving repository name %q", ic.Repository)
			if c.opts.SkipErrors {
				errs = multierror.Append(errs, wrapped)
				continue
			} else {
				return nil, nil, wrapped
			}
		}

		for _, id := range ic.ExternalIDs {
			var sid string

			switch tid := id.(type) {
			case string:
				sid = tid
			case int, int8, int16, int32, int64:
				sid = strconv.FormatInt(reflect.ValueOf(id).Int(), 10)
			case uint, uint8, uint16, uint32, uint64:
				sid = strconv.FormatUint(reflect.ValueOf(id).Uint(), 10)
			case float32:
				sid = strconv.FormatFloat(float64(tid), 'f', -1, 32)
			case float64:
				sid = strconv.FormatFloat(tid, 'f', -1, 64)
			default:
				return nil, nil, errors.Errorf("cannot convert value of type %T into a valid external ID: expected string or int", id)
			}

			specs = append(specs, &batches.ChangesetSpec{
				BaseRepository:    repo.ID,
				ExternalChangeset: &batches.ExternalChangeset{ExternalID: sid},
			})
		}
	}

	return specs, c.logManager.LogFiles(), errs.ErrorOrNil()
}
