package campaigns

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/neelance/parallel"
	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/campaigns/graphql"
)

type Executor interface {
	AddTask(repo *graphql.Repository, steps []Step, template *ChangesetTemplate) *TaskStatus
	LogFiles() []string
	Start(ctx context.Context)
	Wait() ([]*ChangesetSpec, error)
}

type Task struct {
	Repository *graphql.Repository
	Steps      []Step
	Template   *ChangesetTemplate `json:"-"`
}

func (t *Task) cacheKey() ExecutionCacheKey {
	return ExecutionCacheKey{t}
}

type TaskStatus struct {
	Cached bool

	LogFile    string
	EnqueuedAt time.Time
	StartedAt  time.Time
	FinishedAt time.Time

	// TODO: add current step and progress fields.

	// Result fields.
	ChangesetSpec *ChangesetSpec
	Err           error
}

type ExecutorUpdateCallback func(*Task, TaskStatus)

type executor struct {
	ExecutorOpts

	cache  ExecutionCache
	client api.Client
	logger *LogManager
	tasks  sync.Map

	par           *parallel.Run
	doneEnqueuing chan struct{}

	update ExecutorUpdateCallback

	specs   []*ChangesetSpec
	specsMu sync.Mutex
}

func newExecutor(opts ExecutorOpts, client api.Client, update ExecutorUpdateCallback) *executor {
	return &executor{
		ExecutorOpts:  opts,
		cache:         opts.Cache,
		client:        client,
		doneEnqueuing: make(chan struct{}),
		logger:        NewLogManager(opts.KeepLogs),
		par:           parallel.NewRun(opts.Parallelism),
		update:        update,
	}
}

func (x *executor) AddTask(repo *graphql.Repository, steps []Step, template *ChangesetTemplate) *TaskStatus {
	task := &Task{repo, steps, template}
	ts := &TaskStatus{EnqueuedAt: time.Now()}
	x.tasks.Store(task, ts)
	return ts
}

func (x *executor) LogFiles() []string {
	return x.logger.LogFiles()
}

func (x *executor) Start(ctx context.Context) {
	x.tasks.Range(func(k, v interface{}) bool {
		x.par.Acquire()

		go func(task *Task) {
			defer x.par.Release()
			err := x.do(ctx, task)
			if err != nil {
				x.par.Error(err)
			}
		}(k.(*Task))

		return true
	})

	close(x.doneEnqueuing)
}

func (x *executor) Wait() ([]*ChangesetSpec, error) {
	<-x.doneEnqueuing
	if err := x.par.Wait(); err != nil {
		return nil, err
	}
	return x.specs, nil
}

func (x *executor) do(ctx context.Context, task *Task) (err error) {
	// Set up the task status so we can update it as we progress.
	ts, _ := x.tasks.LoadOrStore(task, &TaskStatus{})
	status, ok := ts.(*TaskStatus)
	if !ok {
		return errors.Errorf("unexpected non-TaskStatus value of type %T: %+v", ts, ts)
	}

	// Ensure that the status is updated when we're done.
	defer func() {
		status.FinishedAt = time.Now()
		status.Err = err
		x.updateTaskStatus(task, status)
	}()

	// We're away!
	status.StartedAt = time.Now()
	x.updateTaskStatus(task, status)

	// Check if the task is cached.
	cacheKey := task.cacheKey()
	if x.ClearCache {
		if err = x.cache.Clear(ctx, cacheKey); err != nil {
			err = errors.Wrapf(err, "clearing cache for %q", task.Repository.Name)
			return
		}
	} else {
		var result *ChangesetSpec
		if result, err = x.cache.Get(ctx, cacheKey); err != nil {
			err = errors.Wrapf(err, "checking cache for %q", task.Repository.Name)
			return
		} else if result != nil && len(result.Commits) == 1 {
			// Build a new changeset spec. We don't want to use `result` as is,
			// because the changesetTemplate may have changed. In that case
			// the diff would still be valid, so we take it from the cache,
			// but we still build a new ChangesetSpec from the task.
			diff := result.Commits[0].Diff
			spec := createChangesetSpec(task, diff)

			status.Cached = true
			status.ChangesetSpec = spec
			status.FinishedAt = time.Now()
			x.updateTaskStatus(task, status)

			// Add the spec to the executor's list of completed specs.
			x.specsMu.Lock()
			x.specs = append(x.specs, spec)
			x.specsMu.Unlock()

			return
		}
	}

	// It isn't, so let's get ready to run the task. First, let's set up our
	// logging.
	log, err := x.logger.AddTask(task)
	if err != nil {
		err = errors.Wrap(err, "creating log file")
		return
	}
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "Error executing task. The log file can be found at:\n%s\n", log.Path())
			log.MarkErrored()
		}
		log.Close()
	}()

	// Set up our timeout.
	runCtx, cancel := context.WithTimeout(ctx, x.Timeout)
	defer cancel()

	// Actually execute the steps.
	diff, err := runSteps(runCtx, x.client, task.Repository, task.Steps, log)
	if err != nil {
		if reachedTimeout(runCtx, err) {
			err = &errTimeoutReached{timeout: x.Timeout}
		}
		return
	}

	// Build the changeset spec.
	spec := createChangesetSpec(task, string(diff))

	status.ChangesetSpec = spec
	x.updateTaskStatus(task, status)

	// Add the spec to the executor's list of completed specs.
	x.specsMu.Lock()
	x.specs = append(x.specs, spec)
	x.specsMu.Unlock()

	// Add to the cache. We don't use runCtx here because we want to write to
	// the cache even if we've now reached the timeout.
	if err = x.cache.Set(ctx, cacheKey, spec); err != nil {
		err = errors.Wrapf(err, "caching result for %q", task.Repository.Name)
	}

	return
}

func (x *executor) updateTaskStatus(task *Task, status *TaskStatus) {
	x.tasks.Store(task, status)
	if x.update != nil {
		x.update(task, *status)
	}
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

func createChangesetSpec(task *Task, diff string) *ChangesetSpec {
	return &ChangesetSpec{
		BaseRepository: task.Repository.ID,
		CreatedChangeset: &CreatedChangeset{
			BaseRef:        task.Repository.BaseRef(),
			BaseRev:        task.Repository.Rev(),
			HeadRepository: task.Repository.ID,
			HeadRef:        "refs/heads/" + task.Template.Branch,
			Title:          task.Template.Title,
			Body:           task.Template.Body,
			Commits: []GitCommitDescription{
				{
					Message: task.Template.Commit.Message,
					Diff:    string(diff),
				},
			},
			Published: task.Template.Published,
		},
	}
}
