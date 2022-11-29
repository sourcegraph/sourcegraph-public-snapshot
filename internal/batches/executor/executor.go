package executor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/neelance/parallel"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/batches/docker"
	"github.com/sourcegraph/src-cli/internal/batches/log"
	"github.com/sourcegraph/src-cli/internal/batches/repozip"
	"github.com/sourcegraph/src-cli/internal/batches/util"
	"github.com/sourcegraph/src-cli/internal/batches/workspace"

	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
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
	task        *Task
	stepResults []execution.AfterStepResult
	err         error
}

type imageEnsurer func(ctx context.Context, name string) (docker.Image, error)

type NewExecutorOpts struct {
	// Dependencies
	Creator             workspace.Creator
	RepoArchiveRegistry repozip.ArchiveRegistry
	EnsureImage         imageEnsurer
	Logger              log.LogManager

	// Config
	Parallelism      int
	Timeout          time.Duration
	WorkingDirectory string
	TempDir          string
	IsRemote         bool
	GlobalEnv        []string
	ForceRoot        bool

	BinaryDiffs bool
}

type executor struct {
	opts NewExecutorOpts

	par           *parallel.Run
	doneEnqueuing chan struct{}

	results   []taskResult
	resultsMu sync.Mutex
}

func NewExecutor(opts NewExecutorOpts) *executor {
	return &executor{
		opts: opts,

		doneEnqueuing: make(chan struct{}),
		par:           parallel.NewRun(opts.Parallelism),
	}
}

// Start starts the execution of the given Tasks in goroutines, calling the
// given taskStatusHandler to update the progress of the tasks.
func (x *executor) Start(ctx context.Context, tasks []*Task, ui TaskExecutionUI) {
	defer func() { close(x.doneEnqueuing) }()

	for _, task := range tasks {
		select {
		case <-ctx.Done():
			return
		default:
		}

		x.par.Acquire()

		go func(task *Task, ui TaskExecutionUI) {
			defer x.par.Release()

			select {
			case <-ctx.Done():
				return
			default:
				err := x.do(ctx, task, ui)
				if err != nil {
					x.par.Error(err)
				}
			}
		}(task, ui)
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

func (x *executor) do(ctx context.Context, task *Task, ui TaskExecutionUI) (err error) {
	// Ensure that the status is updated when we're done.
	defer func() {
		ui.TaskFinished(task, err)
	}()

	// We're away!
	ui.TaskStarted(task)

	// Let's set up our logging.
	l, err := x.opts.Logger.AddTask(util.SlugForPathInRepo(task.Repository.Name, task.Repository.Rev(), task.Path))
	if err != nil {
		return errors.Wrap(err, "creating log file")
	}
	defer l.Close()

	// Now checkout the archive.
	repoArchive := x.opts.RepoArchiveRegistry.Checkout(
		repozip.RepoRevision{
			RepoName: task.Repository.Name,
			Commit:   task.Repository.Rev(),
		},
		task.ArchivePathToFetch(),
	)

	// Actually execute the steps.
	opts := &RunStepsOpts{
		Task:             task,
		Logger:           l,
		WC:               x.opts.Creator,
		EnsureImage:      x.opts.EnsureImage,
		TempDir:          x.opts.TempDir,
		GlobalEnv:        x.opts.GlobalEnv,
		Timeout:          x.opts.Timeout,
		RepoArchive:      repoArchive,
		WorkingDirectory: x.opts.WorkingDirectory,
		ForceRoot:        x.opts.ForceRoot,
		BinaryDiffs:      x.opts.BinaryDiffs,

		UI: ui.StepsExecutionUI(task),
	}
	stepResults, err := RunSteps(ctx, opts)
	if err != nil {
		// Create a more visual error for the UI.
		err = TaskExecutionErr{
			Err:        err,
			Logfile:    l.Path(),
			Repository: task.Repository.Name,
		}
		l.MarkErrored()
	}
	x.addResult(task, stepResults, err)

	return err
}

func (x *executor) addResult(task *Task, stepResults []execution.AfterStepResult, err error) {
	x.resultsMu.Lock()
	defer x.resultsMu.Unlock()

	x.results = append(x.results, taskResult{
		task:        task,
		stepResults: stepResults,
		err:         err,
	})
}
