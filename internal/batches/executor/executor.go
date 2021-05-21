package executor

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/neelance/parallel"
	"github.com/pkg/errors"
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

// taskResult is a combination of a Task and the result of its execution.
type taskResult struct {
	task   *Task
	result executionResult
}

type newExecutorOpts struct {
	// Dependencies
	Creator workspace.Creator
	Fetcher batches.RepoFetcher
	Logger  *log.Manager

	// Config
	AutoAuthorDetails bool
	Parallelism       int
	Timeout           time.Duration
	TempDir           string
}

type executor struct {
	opts newExecutorOpts

	par           *parallel.Run
	doneEnqueuing chan struct{}

	results   []taskResult
	resultsMu sync.Mutex
}

func newExecutor(opts newExecutorOpts) *executor {
	return &executor{
		opts: opts,

		doneEnqueuing: make(chan struct{}),
		par:           parallel.NewRun(opts.Parallelism),
	}
}

type taskStatusHandler interface {
	Update(task *Task, callback func(status *TaskStatus))
}

// Start starts the execution of the given Tasks in goroutines, calling the
// given taskStatusHandler to update the progress of the tasks.
func (x *executor) Start(ctx context.Context, tasks []*Task, status taskStatusHandler) {
	defer func() { close(x.doneEnqueuing) }()

	for _, task := range tasks {
		select {
		case <-ctx.Done():
			return
		default:
		}

		x.par.Acquire()

		go func(task *Task, status taskStatusHandler) {
			defer x.par.Release()

			select {
			case <-ctx.Done():
				return
			default:
				err := x.do(ctx, task, status)
				if err != nil {
					x.par.Error(err)
				}
			}
		}(task, status)
	}
}

// Wait blocks until all Tasks enqueued with Start have been executed.
func (x *executor) Wait(ctx context.Context) ([]taskResult, error) {
	<-x.doneEnqueuing

	result := make(chan error, 1)

	go func(ch chan error) {
		ch <- x.par.Wait()
	}(result)

	select {
	case <-ctx.Done():
		return x.results, ctx.Err()
	case err := <-result:
		close(result)
		if err != nil {
			return x.results, err
		}
	}

	return x.results, nil
}

func (x *executor) do(ctx context.Context, task *Task, status taskStatusHandler) (err error) {
	// Ensure that the status is updated when we're done.
	defer func() {
		status.Update(task, func(status *TaskStatus) {
			status.CurrentlyExecuting = ""
			status.Err = err
		})
	}()

	// We're away!
	status.Update(task, func(status *TaskStatus) {
		status.StartedAt = time.Now()
	})

	// Let's set up our logging.
	log, err := x.opts.Logger.AddTask(task.Repository.SlugForPath(task.Path))
	if err != nil {
		return errors.Wrap(err, "creating log file")
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
	task.Archive = x.opts.Fetcher.Checkout(task.Repository, task.ArchivePathToFetch())

	// Set up our timeout.
	runCtx, cancel := context.WithTimeout(ctx, x.opts.Timeout)
	defer cancel()

	// Actually execute the steps.
	opts := &executionOpts{
		archive:               task.Archive,
		batchChangeAttributes: task.BatchChangeAttributes,
		repo:                  task.Repository,
		path:                  task.Path,
		steps:                 task.Steps,
		wc:                    x.opts.Creator,
		logger:                log,
		tempDir:               x.opts.TempDir,
		reportProgress: func(currentlyExecuting string) {
			status.Update(task, func(status *TaskStatus) {
				status.CurrentlyExecuting = currentlyExecuting
			})
		},
	}

	result, err := runSteps(runCtx, opts)
	if err != nil {
		if reachedTimeout(runCtx, err) {
			err = &errTimeoutReached{timeout: x.opts.Timeout}
		}
		return err
	}

	x.addResult(task, result)

	return nil
}

func (x *executor) addResult(task *Task, result executionResult) {
	x.resultsMu.Lock()
	defer x.resultsMu.Unlock()

	x.results = append(x.results, taskResult{
		task:   task,
		result: result,
	})
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
