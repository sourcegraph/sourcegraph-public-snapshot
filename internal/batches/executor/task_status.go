package executor

import (
	"sync"
	"time"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/src-cli/internal/batches"
)

// TaskStatusCollection is a collection of TaskStatuses that provides
// concurrency-safe access to the statuses by locking access and providing
// copies of the TaskStatuses.
type TaskStatusCollection struct {
	statuses   map[*Task]*TaskStatus
	statusesMu sync.RWMutex
}

func NewTaskStatusCollection(tasks []*Task) *TaskStatusCollection {
	tsc := &TaskStatusCollection{
		statuses: make(map[*Task]*TaskStatus),
	}

	for _, t := range tasks {
		tsc.statuses[t] = &TaskStatus{
			RepoName: t.Repository.Name,
			Path:     t.Path,
		}
	}

	return tsc
}

// Update updates the TaskStatus for the given Task by calling the provided
// update callback with the TaskStatus.
func (tsc *TaskStatusCollection) Update(task *Task, update func(status *TaskStatus)) {
	tsc.statusesMu.Lock()
	defer tsc.statusesMu.Unlock()

	status, ok := tsc.statuses[task]
	if ok {
		update(status)
	}
}

// CopyStatuses creates a copy of all TaskStatuses and calls the provided
// callback with list.
func (tsc *TaskStatusCollection) CopyStatuses(callback func([]*TaskStatus)) {
	tsc.statusesMu.RLock()
	defer tsc.statusesMu.RUnlock()

	var s []*TaskStatus
	for _, status := range tsc.statuses {
		s = append(s, status)
	}

	callback(s)
}

type TaskStatus struct {
	RepoName string
	Path     string

	LogFile            string
	StartedAt          time.Time
	FinishedAt         time.Time
	CurrentlyExecuting string

	// ChangesetSpecs are the specs produced by executing the Task in a
	// repository. One Task can produce multiple ChangesetSpecs (see
	// createChangesetSpec).
	// Only check this field once ChangesetSpecsDone is set.
	ChangesetSpecs []*batches.ChangesetSpec
	// ChangesetSpecsDone is set after the Coordinator attempted to build the
	// ChangesetSpecs of a task.
	ChangesetSpecsDone bool

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

func (ts *TaskStatus) FinishedExecution() bool {
	return !ts.StartedAt.IsZero() && !ts.FinishedAt.IsZero()
}

func (ts *TaskStatus) FinishedBuildingSpecs() bool {
	return ts.ChangesetSpecsDone
}

func (ts *TaskStatus) ExecutionTime() time.Duration {
	return ts.FinishedAt.Sub(ts.StartedAt).Truncate(time.Millisecond)
}

// FileDiffs returns the file diffs produced by the Task in the given
// repository.
// If no file diffs were produced, the task resulted in an error, or the task
// hasn't finished execution yet, the second return value is false.
func (ts *TaskStatus) FileDiffs() ([]*diff.FileDiff, bool, error) {
	if !ts.FinishedBuildingSpecs() || len(ts.ChangesetSpecs) == 0 || ts.Err != nil {
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
