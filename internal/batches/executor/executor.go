package executor

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
	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
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

type Task struct {
	Repository *graphql.Repository

	// Path is the folder relative to the repository's root in which the steps
	// should be executed.
	Path string
	// OnlyFetchWorkspace determines whether the repository archive contains
	// the complete repository or just the files in Path (and additional files,
	// see RepoFetcher).
	// If Path is "" then this setting has no effect.
	OnlyFetchWorkspace bool

	Steps   []batches.Step
	Outputs map[string]interface{}

	// TODO(mrnugget): this should just be a single BatchSpec field instead, if
	// we can make it work with caching
	BatchChangeAttributes *BatchChangeAttributes     `json:"-"`
	Template              *batches.ChangesetTemplate `json:"-"`
	TransformChanges      *batches.TransformChanges  `json:"-"`

	Archive batches.RepoZip `json:"-"`
}

func (t *Task) ArchivePathToFetch() string {
	if t.OnlyFetchWorkspace {
		return t.Path
	}
	return ""
}

func (t *Task) cacheKey() ExecutionCacheKey {
	return ExecutionCacheKey{t}
}

type TaskStatus struct {
	RepoName string
	Path     string

	Cached bool

	LogFile    string
	EnqueuedAt time.Time
	StartedAt  time.Time
	FinishedAt time.Time

	// TODO: add current step and progress fields.
	CurrentlyExecuting string

	// ChangesetSpecs are the specs produced by executing the Task in a
	// repository. With the introduction of `transformChanges` to the batch
	// spec, one Task can produce multiple ChangesetSpecs.
	ChangesetSpecs []*batches.ChangesetSpec
	// Err is set if executing the Task lead to an error.
	Err error

	fileDiffs     []*diff.FileDiff
	fileDiffsErr  error
	fileDiffsOnce sync.Once
}

func (ts *TaskStatus) DisplayName() string {
	if ts.Path != "" {
		return ts.RepoName + ":" + ts.Path
	}
	return ts.RepoName
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

type Opts struct {
	CacheDir   string
	ClearCache bool

	CleanArchives bool

	Creator     workspace.Creator
	Parallelism int
	Timeout     time.Duration

	KeepLogs bool
	TempDir  string
}

type executor struct {
	cache      ExecutionCache
	clearCache bool

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

// TODO(mrnugget): Why are client and features not part of Opts?
func New(opts Opts, client api.Client, features batches.FeatureFlags) *executor {
	return &executor{
		cache:      NewCache(opts.CacheDir),
		clearCache: opts.ClearCache,

		logger:  log.NewManager(opts.TempDir, opts.KeepLogs),
		creator: opts.Creator,

		fetcher: batches.NewRepoFetcher(client, opts.CacheDir, opts.CleanArchives),

		client:   client,
		features: features,

		tempDir: opts.TempDir,
		timeout: opts.Timeout,

		doneEnqueuing: make(chan struct{}),
		par:           parallel.NewRun(opts.Parallelism),
		tasks:         []*Task{},
		statuses:      map[*Task]*TaskStatus{},
	}
}

func (x *executor) AddTask(task *Task) {
	task.Archive = x.fetcher.Checkout(task.Repository, task.ArchivePathToFetch())
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
			return nil, err
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

	// Check if the task is cached.
	cacheKey := task.cacheKey()
	if x.clearCache {
		if err = x.cache.Clear(ctx, cacheKey); err != nil {
			err = errors.Wrapf(err, "clearing cache for %q", task.Repository.Name)
			return
		}
	} else {
		var (
			result executionResult
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
			if result.Diff == "" {
				x.updateTaskStatus(task, func(status *TaskStatus) {
					status.Cached = true
					status.FinishedAt = time.Now()

				})
				return
			}

			var specs []*batches.ChangesetSpec
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
			if err := x.addCompletedSpecs(task.Repository, specs); err != nil {
				return err
			}

			return
		}
	}

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
	if result.Diff == "" {
		return
	}

	x.updateTaskStatus(task, func(status *TaskStatus) {
		status.ChangesetSpecs = specs
	})

	if err := x.addCompletedSpecs(task.Repository, specs); err != nil {
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

func (x *executor) addCompletedSpecs(repository *graphql.Repository, specs []*batches.ChangesetSpec) error {
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

func createChangesetSpecs(task *Task, result executionResult, features batches.FeatureFlags) ([]*batches.ChangesetSpec, error) {
	repo := task.Repository.Name

	tmplCtx := &ChangesetTemplateContext{
		BatchChangeAttributes: *task.BatchChangeAttributes,
		Steps: StepsContext{
			Changes: result.ChangedFiles,
			Path:    result.Path,
		},
		Outputs:    result.Outputs,
		Repository: *task.Repository,
	}

	var authorName string
	var authorEmail string

	if task.Template.Commit.Author == nil {
		if features.IncludeAutoAuthorDetails {
			// user did not provide author info, so use defaults
			authorName = "Sourcegraph"
			authorEmail = "batch-changes@sourcegraph.com"
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

	newSpec := func(branch, diff string) *batches.ChangesetSpec {
		return &batches.ChangesetSpec{
			BaseRepository: task.Repository.ID,
			CreatedChangeset: &batches.CreatedChangeset{
				BaseRef:        task.Repository.BaseRef(),
				BaseRev:        task.Repository.Rev(),
				HeadRepository: task.Repository.ID,
				HeadRef:        "refs/heads/" + branch,
				Title:          title,
				Body:           body,
				Commits: []batches.GitCommitDescription{
					{
						Message:     message,
						AuthorName:  authorName,
						AuthorEmail: authorEmail,
						Diff:        diff,
					},
				},
				Published: task.Template.Published.ValueWithSuffix(repo, branch),
			},
		}
	}

	var specs []*batches.ChangesetSpec

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

func groupsForRepository(repo string, transform *batches.TransformChanges) []batches.Group {
	var groups []batches.Group

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

func validateGroups(repo, defaultBranch string, groups []batches.Group) error {
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

func groupFileDiffs(completeDiff, defaultBranch string, groups []batches.Group) (map[string]string, error) {
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
