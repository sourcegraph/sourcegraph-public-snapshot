package executor

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/neelance/parallel"
	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/log"
	"github.com/sourcegraph/src-cli/internal/batches/workspace"
)

type TaskExecutionErr struct {
	Err        error
	Logfile    string
	Repository string
}

func (e TaskExecutionErr) Cause() error {
	return e.Err
}

func (e TaskExecutionErr) Error() string {
	return fmt.Sprintf(
		"execution in %s failed: %s (see %s for details)",
		e.Repository,
		e.Err,
		e.Logfile,
	)
}

func (e TaskExecutionErr) StatusText() string {
	if stepErr, ok := e.Err.(stepFailedErr); ok {
		return stepErr.SingleLineError()
	}
	return e.Err.Error()
}

type Executor interface {
	AddTask(*Task)
	LogFiles() []string
	Start(ctx context.Context)
	Wait(ctx context.Context) ([]*batches.ChangesetSpec, error)

	// LockedTaskStatuses calls the given function with the current state of
	// the task statuses. Before calling the function, the statuses are locked
	// to provide a consistent view of all statuses, but that also means the
	// callback should be as fast as possible.
	LockedTaskStatuses(func([]*TaskStatus))
}

type NewExecutorOpts struct {
	Cache    ExecutionCache
	Client   api.Client
	Features batches.FeatureFlags
	Creator  workspace.Creator

	CleanArchives bool

	CacheDir string

	Parallelism int
	Timeout     time.Duration

	KeepLogs bool
	TempDir  string
}

type executor struct {
	cache ExecutionCache

	features batches.FeatureFlags

	client  api.Client
	logger  *log.Manager
	creator workspace.Creator
	fetcher batches.RepoFetcher

	tasks      []*Task
	statuses   map[*Task]*TaskStatus
	statusesMu sync.RWMutex

	tempDir string
	timeout time.Duration

	par           *parallel.Run
	doneEnqueuing chan struct{}

	specs   []*batches.ChangesetSpec
	specsMu sync.Mutex
}

func New(opts NewExecutorOpts) *executor {
	return &executor{
		cache:    opts.Cache,
		client:   opts.Client,
		features: opts.Features,
		creator:  opts.Creator,

		logger: log.NewManager(opts.TempDir, opts.KeepLogs),

		fetcher: batches.NewRepoFetcher(opts.Client, opts.CacheDir, opts.CleanArchives),

		tempDir: opts.TempDir,
		timeout: opts.Timeout,

		doneEnqueuing: make(chan struct{}),
		par:           parallel.NewRun(opts.Parallelism),
		tasks:         []*Task{},
		statuses:      map[*Task]*TaskStatus{},
	}
}

func (x *executor) AddTask(task *Task) {
	x.tasks = append(x.tasks, task)

	x.statusesMu.Lock()
	x.statuses[task] = &TaskStatus{RepoName: task.Repository.Name, Path: task.Path, EnqueuedAt: time.Now()}
	x.statusesMu.Unlock()
}

func (x *executor) LogFiles() []string {
	return x.logger.LogFiles()
}

func (x *executor) Start(ctx context.Context) {
	defer func() { close(x.doneEnqueuing) }()

	for _, task := range x.tasks {
		select {
		case <-ctx.Done():
			return
		default:
		}

		x.par.Acquire()

		go func(task *Task) {
			defer x.par.Release()

			select {
			case <-ctx.Done():
				return
			default:
				err := x.do(ctx, task)
				if err != nil {
					x.par.Error(err)
				}
			}
		}(task)
	}
}

func (x *executor) Wait(ctx context.Context) ([]*batches.ChangesetSpec, error) {
	<-x.doneEnqueuing

	result := make(chan error, 1)

	go func(ch chan error) {
		ch <- x.par.Wait()
	}(result)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-result:
		close(result)
		if err != nil {
			return x.specs, err
		}
	}

	return x.specs, nil
}

func (x *executor) do(ctx context.Context, task *Task) (err error) {
	// Ensure that the status is updated when we're done.
	defer func() {
		x.updateTaskStatus(task, func(status *TaskStatus) {
			status.FinishedAt = time.Now()
			status.CurrentlyExecuting = ""
			status.Err = err
		})
	}()

	// We're away!
	x.updateTaskStatus(task, func(status *TaskStatus) {
		status.StartedAt = time.Now()
	})

	// It isn't, so let's get ready to run the task. First, let's set up our
	// logging.
	log, err := x.logger.AddTask(task.Repository.SlugForPath(task.Path))
	if err != nil {
		err = errors.Wrap(err, "creating log file")
		return
	}
	defer func() {
		if err != nil {
			err = TaskExecutionErr{
				Err:        err,
				Logfile:    log.Path(),
				Repository: task.Repository.Name,
			}
			log.MarkErrored()
		}
		log.Close()
	}()

	// Now checkout the archive
	task.Archive = x.fetcher.Checkout(task.Repository, task.ArchivePathToFetch())

	// Set up our timeout.
	runCtx, cancel := context.WithTimeout(ctx, x.timeout)
	defer cancel()

	// Actually execute the steps.
	opts := &executionOpts{
		archive:               task.Archive,
		wc:                    x.creator,
		batchChangeAttributes: task.BatchChangeAttributes,
		repo:                  task.Repository,
		path:                  task.Path,
		steps:                 task.Steps,
		logger:                log,
		tempDir:               x.tempDir,
		reportProgress: func(currentlyExecuting string) {
			x.updateTaskStatus(task, func(status *TaskStatus) {
				status.CurrentlyExecuting = currentlyExecuting
			})
		},
	}
	result, err := runSteps(runCtx, opts)
	if err != nil {
		if reachedTimeout(runCtx, err) {
			err = &errTimeoutReached{timeout: x.timeout}
		}
		return
	}

	// Check if the task is cached.
	cacheKey := task.cacheKey()

	// Add to the cache. We don't use runCtx here because we want to write to
	// the cache even if we've now reached the timeout.
	if err = x.cache.Set(ctx, cacheKey, result); err != nil {
		err = errors.Wrapf(err, "caching result for %q", task.Repository.Name)
	}

	// If the steps didn't result in any diff, we don't need to add it to the
	// list of specs that are displayed to the user and send to the server.
	if result.Diff == "" {
		return
	}

	// Build the changeset specs.
	specs, err := createChangesetSpecs(task, result, x.features)
	if err != nil {
		return err
	}

	x.updateTaskStatus(task, func(status *TaskStatus) {
		status.ChangesetSpecs = specs
	})

	if err := x.addCompletedSpecs(specs); err != nil {
		return err
	}

	return
}

func (x *executor) updateTaskStatus(task *Task, update func(status *TaskStatus)) {
	x.statusesMu.Lock()
	defer x.statusesMu.Unlock()

	status, ok := x.statuses[task]
	if ok {
		update(status)
	}
}

func (x *executor) addCompletedSpecs(specs []*batches.ChangesetSpec) error {
	x.specsMu.Lock()
	defer x.specsMu.Unlock()

	x.specs = append(x.specs, specs...)
	return nil
}

func (x *executor) LockedTaskStatuses(callback func([]*TaskStatus)) {
	x.statusesMu.RLock()
	defer x.statusesMu.RUnlock()

	var s []*TaskStatus
	for _, status := range x.statuses {
		s = append(s, status)
	}

	callback(s)
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

	return errors.Is(errors.Cause(err), context.DeadlineExceeded)
}
