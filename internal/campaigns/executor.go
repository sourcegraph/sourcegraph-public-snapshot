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
	AddTask(repo *graphql.Repository, steps []Step, template *ChangesetTemplate)
	LogFiles() []string
	Start(ctx context.Context)
	Wait() ([]*ChangesetSpec, error)

	// LockedTaskStatuses calls the given function with the current state of
	// the task statuses. Before calling the function, the statuses are locked
	// to provide a consistent view of all statuses, but that also means the
	// callback should be as fast as possible.
	LockedTaskStatuses(func([]*TaskStatus))
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
	RepoName string

	Cached bool

	LogFile    string
	EnqueuedAt time.Time
	StartedAt  time.Time
	FinishedAt time.Time

	// TODO: add current step and progress fields.
	CurrentlyExecuting string

	// Result fields.
	ChangesetSpec *ChangesetSpec
	Err           error
}

func (ts *TaskStatus) Clone() *TaskStatus {
	clone := *ts
	return &clone
}

func (ts *TaskStatus) IsRunning() bool {
	return !ts.StartedAt.IsZero() && ts.FinishedAt.IsZero()
}

func (ts *TaskStatus) IsCompleted() bool {
	return !ts.StartedAt.IsZero() && !ts.FinishedAt.IsZero()
}

func (ts *TaskStatus) ExecutionTime() time.Duration {
	return ts.FinishedAt.Sub(ts.StartedAt).Truncate(time.Millisecond)
}

type executor struct {
	ExecutorOpts

	cache    ExecutionCache
	client   api.Client
	features featureFlags
	logger   *LogManager
	creator  *WorkspaceCreator

	tasks      []*Task
	statuses   map[*Task]*TaskStatus
	statusesMu sync.RWMutex

	tempDir string

	par           *parallel.Run
	doneEnqueuing chan struct{}

	specs   []*ChangesetSpec
	specsMu sync.Mutex
}

func newExecutor(opts ExecutorOpts, client api.Client, features featureFlags) *executor {
	return &executor{
		ExecutorOpts:  opts,
		cache:         opts.Cache,
		creator:       opts.Creator,
		client:        client,
		features:      features,
		doneEnqueuing: make(chan struct{}),
		logger:        NewLogManager(opts.TempDir, opts.KeepLogs),
		tempDir:       opts.TempDir,
		par:           parallel.NewRun(opts.Parallelism),
		tasks:         []*Task{},
		statuses:      map[*Task]*TaskStatus{},
	}
}

func (x *executor) AddTask(repo *graphql.Repository, steps []Step, template *ChangesetTemplate) {
	task := &Task{repo, steps, template}
	x.tasks = append(x.tasks, task)

	x.statusesMu.Lock()
	x.statuses[task] = &TaskStatus{RepoName: repo.Name, EnqueuedAt: time.Now()}
	x.statusesMu.Unlock()
}

func (x *executor) LogFiles() []string {
	return x.logger.LogFiles()
}

func (x *executor) Start(ctx context.Context) {
	for _, task := range x.tasks {
		select {
		case <-ctx.Done():
			break
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
		}
		if result != nil {
			// Build a new changeset spec. We don't want to use `result` as is,
			// because the changesetTemplate may have changed. In that case
			// the diff would still be valid, so we take it from the cache,
			// but we still build a new ChangesetSpec from the task.
			var diff string

			if len(result.Commits) > 1 {
				panic("campaigns currently lack support for multiple commits per changeset")
			}
			if len(result.Commits) == 1 {
				diff = result.Commits[0].Diff
			}

			// If the cached result resulted in an empty diff, we don't need to
			// add it to the list of specs that are displayed to the user and
			// send to the server. Instead, we can just report that the task is
			// complete and move on.
			if len(diff) == 0 {
				x.updateTaskStatus(task, func(status *TaskStatus) {
					status.Cached = true
					status.FinishedAt = time.Now()

				})
				return
			}

			spec := createChangesetSpec(task, diff, x.features)

			x.updateTaskStatus(task, func(status *TaskStatus) {
				status.ChangesetSpec = spec
				status.Cached = true
				status.FinishedAt = time.Now()
			})

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
			err = TaskExecutionErr{
				Err:        err,
				Logfile:    log.Path(),
				Repository: task.Repository.Name,
			}
			log.MarkErrored()
		}
		log.Close()
	}()

	// Set up our timeout.
	runCtx, cancel := context.WithTimeout(ctx, x.Timeout)
	defer cancel()

	// Actually execute the steps.
	diff, err := runSteps(runCtx, x.creator, task.Repository, task.Steps, log, x.tempDir, func(currentlyExecuting string) {
		x.updateTaskStatus(task, func(status *TaskStatus) {
			status.CurrentlyExecuting = currentlyExecuting
		})

	})
	if err != nil {
		if reachedTimeout(runCtx, err) {
			err = &errTimeoutReached{timeout: x.Timeout}
		}
		return
	}

	// Build the changeset spec.
	spec := createChangesetSpec(task, string(diff), x.features)

	// Add to the cache. We don't use runCtx here because we want to write to
	// the cache even if we've now reached the timeout.
	if err = x.cache.Set(ctx, cacheKey, spec); err != nil {
		err = errors.Wrapf(err, "caching result for %q", task.Repository.Name)
	}

	// If the steps didn't result in any diff, we don't need to add it to the
	// list of specs that are displayed to the user and send to the server.
	if len(diff) == 0 {
		return
	}

	x.updateTaskStatus(task, func(status *TaskStatus) {
		status.ChangesetSpec = spec
	})

	// Add the spec to the executor's list of completed specs.
	x.specsMu.Lock()
	x.specs = append(x.specs, spec)
	x.specsMu.Unlock()
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

	return errors.Is(err, context.DeadlineExceeded)
}

func createChangesetSpec(task *Task, diff string, features featureFlags) *ChangesetSpec {
	repo := task.Repository.Name

	var authorName string
	var authorEmail string

	if task.Template.Commit.Author == nil {
		if features.includeAutoAuthorDetails {
			// user did not provide author info, so use defaults
			authorName = "Sourcegraph"
			authorEmail = "campaigns@sourcegraph.com"
		}
	} else {
		authorName = task.Template.Commit.Author.Name
		authorEmail = task.Template.Commit.Author.Email
	}

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
					Message:     task.Template.Commit.Message,
					AuthorName:  authorName,
					AuthorEmail: authorEmail,
					Diff:        string(diff),
				},
			},
			Published: task.Template.Published.Value(repo),
		},
	}
}
