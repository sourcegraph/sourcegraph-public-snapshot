package campaigns

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/neelance/parallel"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-diff/diff"
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
	AddTask(repo *graphql.Repository, steps []Step, transform *TransformChanges, template *ChangesetTemplate)
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
	Outputs    map[string]interface{}

	Template         *ChangesetTemplate `json:"-"`
	TransformChanges *TransformChanges  `json:"-"`
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

	// ChangesetSpecs are the specs produced by executing the Task in a
	// repository. With the introduction of `transformChanges` to the campaign
	// spec, one Task can produce multiple ChangesetSpecs.
	ChangesetSpecs []*ChangesetSpec
	// Err is set if executing the Task lead to an error.
	Err error

	fileDiffs     []*diff.FileDiff
	fileDiffsErr  error
	fileDiffsOnce sync.Once
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

// FileDiffs returns the file diffs produced by the Task in the given
// repository.
// If no file diffs were produced, the task resulted in an error, or the task
// hasn't finished execution yet, the second return value is false.
func (ts *TaskStatus) FileDiffs() ([]*diff.FileDiff, bool, error) {
	if !ts.IsCompleted() || len(ts.ChangesetSpecs) == 0 || ts.Err != nil {
		return nil, false, nil
	}

	ts.fileDiffsOnce.Do(func() {
		var all []*diff.FileDiff

		for _, spec := range ts.ChangesetSpecs {
			fd, err := diff.ParseMultiFileDiff([]byte(spec.Commits[0].Diff))
			if err != nil {
				ts.fileDiffsErr = err
				return
			}

			all = append(all, fd...)
		}

		ts.fileDiffs = all
	})

	return ts.fileDiffs, len(ts.fileDiffs) != 0, ts.fileDiffsErr
}

type executor struct {
	ExecutorOpts

	cache    ExecutionCache
	client   api.Client
	features featureFlags
	logger   *LogManager
	creator  WorkspaceCreator
	fetcher  RepoFetcher

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
		fetcher:       opts.RepoFetcher,
		doneEnqueuing: make(chan struct{}),
		logger:        NewLogManager(opts.TempDir, opts.KeepLogs),
		tempDir:       opts.TempDir,
		par:           parallel.NewRun(opts.Parallelism),
		tasks:         []*Task{},
		statuses:      map[*Task]*TaskStatus{},
	}
}

func (x *executor) AddTask(repo *graphql.Repository, steps []Step, transform *TransformChanges, template *ChangesetTemplate) {
	task := &Task{
		Repository:       repo,
		Steps:            steps,
		Template:         template,
		TransformChanges: transform,
	}

	x.tasks = append(x.tasks, task)

	x.statusesMu.Lock()
	x.statuses[task] = &TaskStatus{RepoName: repo.Name, EnqueuedAt: time.Now()}
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
		var (
			result ExecutionResult
			found  bool
		)

		result, found, err = x.cache.Get(ctx, cacheKey)
		if err != nil {
			err = errors.Wrapf(err, "checking cache for %q", task.Repository.Name)
			return
		}
		if found {
			// If the cached result resulted in an empty diff, we don't need to
			// add it to the list of specs that are displayed to the user and
			// send to the server. Instead, we can just report that the task is
			// complete and move on.
			if len(result.Diff) == 0 {
				x.updateTaskStatus(task, func(status *TaskStatus) {
					status.Cached = true
					status.FinishedAt = time.Now()

				})
				return
			}

			var specs []*ChangesetSpec
			specs, err = createChangesetSpecs(task, result, x.features)
			if err != nil {
				return err
			}

			x.updateTaskStatus(task, func(status *TaskStatus) {
				status.ChangesetSpecs = specs
				status.Cached = true
				status.FinishedAt = time.Now()
			})

			// Add the spec to the executor's list of completed specs.
			x.specsMu.Lock()
			x.specs = append(x.specs, specs...)
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
	result, err := runSteps(runCtx, x.fetcher, x.creator, task.Repository, task.Steps, log, x.tempDir, func(currentlyExecuting string) {
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

	// Build the changeset specs.
	specs, err := createChangesetSpecs(task, result, x.features)
	if err != nil {
		return err
	}

	// Add to the cache. We don't use runCtx here because we want to write to
	// the cache even if we've now reached the timeout.
	if err = x.cache.Set(ctx, cacheKey, result); err != nil {
		err = errors.Wrapf(err, "caching result for %q", task.Repository.Name)
	}

	// If the steps didn't result in any diff, we don't need to add it to the
	// list of specs that are displayed to the user and send to the server.
	if len(result.Diff) == 0 {
		return
	}

	x.updateTaskStatus(task, func(status *TaskStatus) {
		status.ChangesetSpecs = specs
	})

	// Add the spec to the executor's list of completed specs.
	x.specsMu.Lock()
	x.specs = append(x.specs, specs...)
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

func createChangesetSpecs(task *Task, result ExecutionResult, features featureFlags) ([]*ChangesetSpec, error) {
	repo := task.Repository.Name

	tmplCtx := &ChangesetTemplateContext{
		Steps:      result.ChangedFiles,
		Outputs:    result.Outputs,
		Repository: *task.Repository,
	}

	var authorName string
	var authorEmail string

	if task.Template.Commit.Author == nil {
		if features.includeAutoAuthorDetails {
			// user did not provide author info, so use defaults
			authorName = "Sourcegraph"
			authorEmail = "campaigns@sourcegraph.com"
		}
	} else {
		var err error
		authorName, err = renderChangesetTemplateField("authorName", task.Template.Commit.Author.Name, tmplCtx)
		if err != nil {
			return nil, err
		}
		authorEmail, err = renderChangesetTemplateField("authorEmail", task.Template.Commit.Author.Email, tmplCtx)
		if err != nil {
			return nil, err
		}
	}

	title, err := renderChangesetTemplateField("title", task.Template.Title, tmplCtx)
	if err != nil {
		return nil, err
	}

	body, err := renderChangesetTemplateField("body", task.Template.Body, tmplCtx)
	if err != nil {
		return nil, err
	}

	message, err := renderChangesetTemplateField("message", task.Template.Commit.Message, tmplCtx)
	if err != nil {
		return nil, err
	}

	// TODO: As a next step, we should extend the ChangesetTemplateContext to also include
	// TransformChanges.Group and then change validateGroups and groupFileDiffs to, for each group,
	// render the branch name *before* grouping the diffs.
	defaultBranch, err := renderChangesetTemplateField("branch", task.Template.Branch, tmplCtx)
	if err != nil {
		return nil, err
	}

	newSpec := func(branch, diff string) *ChangesetSpec {
		return &ChangesetSpec{
			BaseRepository: task.Repository.ID,
			CreatedChangeset: &CreatedChangeset{
				BaseRef:        task.Repository.BaseRef(),
				BaseRev:        task.Repository.Rev(),
				HeadRepository: task.Repository.ID,
				HeadRef:        "refs/heads/" + branch,
				Title:          title,
				Body:           body,
				Commits: []GitCommitDescription{
					{
						Message:     message,
						AuthorName:  authorName,
						AuthorEmail: authorEmail,
						Diff:        diff,
					},
				},
				Published: task.Template.Published.Value(repo),
			},
		}
	}

	var specs []*ChangesetSpec

	groups := groupsForRepository(task.Repository.Name, task.TransformChanges)
	if len(groups) != 0 {
		err := validateGroups(task.Repository.Name, task.Template.Branch, groups)
		if err != nil {
			return specs, err
		}

		// TODO: Regarding 'defaultBranch', see comment above
		diffsByBranch, err := groupFileDiffs(result.Diff, defaultBranch, groups)
		if err != nil {
			return specs, errors.Wrap(err, "grouping diffs failed")
		}

		for branch, diff := range diffsByBranch {
			specs = append(specs, newSpec(branch, diff))
		}
	} else {
		specs = append(specs, newSpec(defaultBranch, result.Diff))
	}

	return specs, nil
}

func groupsForRepository(repo string, transform *TransformChanges) []Group {
	var groups []Group

	if transform == nil {
		return groups
	}

	for _, g := range transform.Group {
		if g.Repository != "" {
			if g.Repository == repo {
				groups = append(groups, g)
			}
		} else {
			groups = append(groups, g)
		}
	}

	return groups
}

func validateGroups(repo, defaultBranch string, groups []Group) error {
	uniqueBranches := make(map[string]struct{}, len(groups))

	for _, g := range groups {
		if _, ok := uniqueBranches[g.Branch]; ok {
			return fmt.Errorf("transformChanges would lead to multiple changesets in repository %s to have the same branch %q", repo, g.Branch)
		} else {
			uniqueBranches[g.Branch] = struct{}{}
		}

		if g.Branch == defaultBranch {
			return fmt.Errorf("transformChanges group branch for repository %s is the same as branch %q in changesetTemplate", repo, defaultBranch)
		}
	}

	return nil
}

func groupFileDiffs(completeDiff, defaultBranch string, groups []Group) (map[string]string, error) {
	fileDiffs, err := diff.ParseMultiFileDiff([]byte(completeDiff))
	if err != nil {
		return nil, err
	}

	// Housekeeping: we setup these two datastructures so we can
	// - access the group.Branch by the directory for which they should be used
	// - check against the given directories, in order.
	branchesByDirectory := make(map[string]string, len(groups))
	dirs := make([]string, len(branchesByDirectory))
	for _, g := range groups {
		branchesByDirectory[g.Directory] = g.Branch
		dirs = append(dirs, g.Directory)
	}

	byBranch := make(map[string][]*diff.FileDiff, len(groups))
	byBranch[defaultBranch] = []*diff.FileDiff{}

	// For each file diff...
	for _, f := range fileDiffs {
		name := f.NewName
		if name == "/dev/null" {
			name = f.OrigName
		}

		// .. we check whether it matches one of the given directories in the
		// group transformations, with the last match winning:
		var matchingDir string
		for _, d := range dirs {
			if strings.Contains(name, d) {
				matchingDir = d
			}
		}

		// If the diff didn't match a rule, it goes into the default branch and
		// the default changeset.
		if matchingDir == "" {
			byBranch[defaultBranch] = append(byBranch[defaultBranch], f)
			continue
		}

		// If it *did* match a directory, we look up which branch we should use:
		branch, ok := branchesByDirectory[matchingDir]
		if !ok {
			panic("this should not happen: " + matchingDir)
		}

		byBranch[branch] = append(byBranch[branch], f)
	}

	finalDiffsByBranch := make(map[string]string, len(byBranch))
	for branch, diffs := range byBranch {
		printed, err := diff.PrintMultiFileDiff(diffs)
		if err != nil {
			return nil, errors.Wrap(err, "printing multi file diff failed")
		}
		finalDiffsByBranch[branch] = string(printed)
	}
	return finalDiffsByBranch, nil
}
