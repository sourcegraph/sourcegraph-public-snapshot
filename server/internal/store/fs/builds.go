package fs

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/kr/fs"
	"golang.org/x/net/context"

	"gopkg.in/inconshreveable/log15.v2"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

// Builds is a local FS-backed implementation of the Builds store. It stores
// Builds and BuildTasks in the filesystem (by default $SGPATH/buildstore, which
// can be re-configured via CLI flags).
//
// Builds are stored and identified uniquely by the combination of
// Repo + Build ID. A build is stored as a JSON-encoded structure in a
// file having the path <repo>/<build-id>. Tasks are stored as a
// JSON-encoded structure having the path:
// <repo>/<build-id>/tasks/<task-id> and are uniquely identifiable by
// a combination of the aforementioned variables.
//
// Queues are stored as indexes in JSON files at the root of the build filesystem.
// They do not store actual builds or tasks, only arrays of specs that point to
// the location of the actual tasks or builds that are enqueued.
//
// The Build FS is persistent both in terms of queue and data.
//
// TODO(sqs): Be clear about what concurrency guarantees we make.
type Builds struct {
	mu       sync.RWMutex                     // guards FS
	imported map[sourcegraph.RepoRevSpec]bool // tracks which repos have already been imported
}

var _ store.Builds = (*Builds)(nil)

const buildQueueFilename = "queue-builds.json"

func NewBuildStore() *Builds {
	return &Builds{imported: make(map[sourcegraph.RepoRevSpec]bool)}
}

func (s *Builds) Get(ctx context.Context, buildSpec sourcegraph.BuildSpec) (*sourcegraph.Build, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.get(ctx, buildSpec)
}

func (s *Builds) get(ctx context.Context, buildSpec sourcegraph.BuildSpec) (*sourcegraph.Build, error) {
	return s.getFromPath(ctx, filepath.Join(dirForBuild(buildSpec), "build.json"))
}

func (s *Builds) getFromPath(ctx context.Context, path string) (*sourcegraph.Build, error) {
	f, err := buildStoreVFS(ctx).Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		var errors MultiError
		err = errors.verify(f.Close(), err)
	}()
	var b sourcegraph.Build
	if err = json.NewDecoder(f).Decode(&b); err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *Builds) List(ctx context.Context, opt *sourcegraph.BuildListOptions) ([]*sourcegraph.Build, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if opt == nil {
		opt = &sourcegraph.BuildListOptions{}
	}

	log15.Debug("Listing builds", "pkg", "server/internal/store/fs", "Repo", opt.Repo, "CommitID", opt.CommitID)

	// selectFn returns true if the build matches the filters.
	selectFn := func(b *sourcegraph.Build) bool {
		if b.ID == 0 {
			// Omit legacy builds from prior to the migration to
			// numeric build IDs. This can be removed in a future
			// version.
			return false
		}
		return (!opt.Queued || (b.Queue && b.StartedAt == nil)) &&
			(!opt.Active || (b.StartedAt != nil && b.EndedAt == nil)) &&
			(!opt.Ended || b.EndedAt != nil) &&
			(!opt.Succeeded || b.Success) &&
			(!opt.Failed || b.Failure) &&
			(!opt.Purged || b.Purged) &&
			(opt.Repo == "" || b.Repo == opt.Repo) &&
			(opt.CommitID == "" || b.CommitID == opt.CommitID)
	}

	var builds []*sourcegraph.Build
	if opt.Repo != "" && opt.CommitID != "" {
		// Fastpath: consult index instead of iterating and
		// filtering over all.
		var err error
		builds, err = s.listBuildsIndexed(ctx, opt.Repo, opt.CommitID, selectFn)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		builds, err = s.listBuildsWalkFS(ctx, opt, selectFn)
		if err != nil {
			return nil, err
		}
	}

	builds = store.SortAndPaginateBuilds(builds, opt)
	return builds, nil
}

func (s *Builds) listBuildsIndexed(ctx context.Context, repo, commitID string, selectFn func(*sourcegraph.Build) bool) ([]*sourcegraph.Build, error) {
	buildIdx, err := getRepoBuildIndex(ctx, repo)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return nil, err
	}

	var builds []*sourcegraph.Build
	for _, bs := range buildIdx[commitID] {
		b, err := s.get(ctx, bs)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, err
		}
		if selectFn(b) {
			builds = append(builds, b)
		}
	}
	return builds, nil
}

func (s *Builds) listBuildsWalkFS(ctx context.Context, opt *sourcegraph.BuildListOptions, selectFn func(*sourcegraph.Build) bool) ([]*sourcegraph.Build, error) {
	root := "."
	if opt.Repo != "" {
		root = filepath.Join(root, opt.Repo)
	}

	var builds []*sourcegraph.Build

	w := fs.WalkFS(root, buildStoreVFS(ctx))
	for w.Step() {
		if w.Err() != nil {
			break
		}
		if w.Stat().IsDir() {
			// do not descend into tasks
			// TODO(gbbr): This will potentially cause a repo called "tasks" to be ignored
			if w.Stat().Name() == "tasks" {
				w.SkipDir()
			}
			continue
		}

		if strings.HasPrefix(filepath.Base(w.Path()), ".") {
			continue
		}
		if name := w.Stat().Name(); name == buildsCommitIDIndexFilename || name == "builds-index.json" || name == buildQueueFilename {
			continue
		}

		b, err := s.getFromPath(ctx, w.Path())
		if err != nil {
			log.Printf("error reading build file %s: %s", w.Path(), err)
			continue
		}
		if selectFn(b) {
			builds = append(builds, b)
		}
	}
	if err := w.Err(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return builds, nil
}

func (s *Builds) GetFirstInCommitOrder(ctx context.Context, repo string, commitIDs []string, successfulOnly bool) (build *sourcegraph.Build, nth int, err error) {
	log15.Debug("Finding first built commit in order", "pkg", "server/internal/store/fs", "repo", repo, "commit order", commitIDs, "successfulOnly", successfulOnly)

	// TODO(gbbr): Get only commits in parameter, not all for the repo.
	buildIdx, err := getRepoBuildIndex(ctx, repo)
	if err != nil {
		return nil, 0, err
	}
	for i, commitID := range commitIDs {
		var builds []*sourcegraph.Build
		for _, bs := range buildIdx[commitID] {
			b, err := s.get(ctx, bs)
			if os.IsNotExist(err) {
				continue
			} else if err != nil {
				return nil, 0, err
			}
			builds = append(builds, b)
		}

		// Need to get the most recently started of all builds
		// that exist for this commit ID.
		store.SortAndPaginateBuilds(builds, &sourcegraph.BuildListOptions{
			Sort: "started_at", Direction: "desc",
			ListOptions: sourcegraph.ListOptions{PerPage: int32(len(builds))}, // ensure no page truncation occurs
		})
		for _, b := range builds {
			if !successfulOnly || b.Success {
				log15.Debug("Found first built commit in order", "pkg", "server/internal/store/fs", "repo", repo, "commitID", commitID, "nth", i, "successfulOnly", successfulOnly)
				return b, i, nil
			}
		}
	}
	log15.Debug("Found no built commits in commit order", "pkg", "server/internal/store/fs", "repo", repo, "commitIDs", commitIDs, "successfulOnly", successfulOnly)
	return nil, -1, nil
}

func (s *Builds) Create(ctx context.Context, newBuild *sourcegraph.Build) (*sourcegraph.Build, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b := *newBuild // copy
	if err := s.createAndUpdateIndex(ctx, &b); err != nil {
		return nil, err
	}

	return &b, nil
}

// This operation as a whole is non-atomic, meaning that in very rare scenarios
// where a build file is successfully created but the queue might not be updated
// for whatever reason (errors or server crashes), it is possible for the build
// to show up in the list as "Queued" even though it is not part of the actual
// queue.
func (s *Builds) create(ctx context.Context, b *sourcegraph.Build) error {
	f, err := createBuildFile(ctx, b)
	if err != nil {
		return err
	}
	defer func() {
		var errors MultiError
		if err != nil {
			errors.verify(buildStoreVFS(ctx).Remove(dirForBuild(b.Spec())))
		}
		err = errors.verify(f.Close(), err)
	}()
	if err = json.NewEncoder(f).Encode(b); err != nil {
		return err
	}
	if b.StartedAt == nil && b.BuildConfig.Queue {
		if err = createBuildQueueEntry(ctx, *b); err != nil {
			return err
		}
	}
	return nil
}

func (s *Builds) createAndUpdateIndex(ctx context.Context, b *sourcegraph.Build) error {
	if err := s.create(ctx, b); err != nil {
		return err
	}
	if err := updateRepoBuildCommitIDIndex(ctx, b.Spec(), b.CommitID); err != nil {
		return err
	}
	return nil
}

func (s *Builds) Update(ctx context.Context, spec sourcegraph.BuildSpec, info sourcegraph.BuildUpdate) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, err := s.get(ctx, spec)
	if err != nil {
		return err
	}

	if info.StartedAt != nil {
		b.StartedAt = info.StartedAt
	}
	if info.EndedAt != nil {
		b.EndedAt = info.EndedAt
	}
	if info.HeartbeatAt != nil {
		b.HeartbeatAt = info.HeartbeatAt
	}
	b.Host = info.Host
	b.Purged = info.Purged
	b.Success = info.Success
	b.Failure = info.Failure
	b.Priority = info.Priority
	b.Killed = info.Killed

	return s.create(ctx, b)
}

func (s *Builds) DequeueNext(ctx context.Context) (*sourcegraph.Build, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var queue []sourcegraph.BuildSpec
	if err := getQueue(ctx, buildQueueFilename, &queue); err != nil {
		return nil, err
	}
	if len(queue) == 0 {
		return nil, nil
	}
	var first sourcegraph.BuildSpec
	first, queue = queue[0], queue[1:]
	if err := replaceQueue(ctx, buildQueueFilename, queue); err != nil {
		return nil, err
	}

	return s.get(ctx, first)
}

func (s *Builds) CreateTasks(ctx context.Context, tasks []*sourcegraph.BuildTask) ([]*sourcegraph.BuildTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, task := range tasks {
		if err := s.updateTask(ctx, task); err != nil {
			return nil, err
		}
	}
	return tasks, nil
}

// updateTask updates or creates a task.
//
// This operation as a whole is non-atomic, meaning that in very rare scenarios
// where a build file is successfully created but the queue might not be updated
// for whatever reason (errors or server crashes), it is possible for the build
// to show up in the list as "Queued" even though it is not part of the actual
// queue.
func (s *Builds) updateTask(ctx context.Context, task *sourcegraph.BuildTask) error {
	f, err := createTaskFile(ctx, task)
	if err != nil {
		return err
	}
	defer func() {
		var errors MultiError
		if err != nil {
			errors.verify(buildStoreVFS(ctx).Remove(filenameForTask(task.Spec())))
		}
		err = errors.verify(f.Close(), err)
	}()
	if err = json.NewEncoder(f).Encode(task); err != nil {
		return err
	}
	return nil
}

func (s *Builds) UpdateTask(ctx context.Context, taskSpec sourcegraph.TaskSpec, info sourcegraph.TaskUpdate) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, err := s.getTask(ctx, taskSpec)
	if err != nil {
		return err
	}

	if info.StartedAt != nil {
		t.StartedAt = info.StartedAt
	}
	if info.EndedAt != nil {
		t.EndedAt = info.EndedAt
	}
	if info.Success {
		t.Success = true
	}
	if info.Failure {
		t.Failure = true
	}

	return s.updateTask(ctx, t)
}

type byTaskID []*sourcegraph.BuildTask

func (s byTaskID) Len() int           { return len(s) }
func (s byTaskID) Less(i, j int) bool { return s[i].ID < s[j].ID }
func (s byTaskID) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s *Builds) ListBuildTasks(ctx context.Context, buildSpec sourcegraph.BuildSpec, opt *sourcegraph.BuildTaskListOptions) ([]*sourcegraph.BuildTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*sourcegraph.BuildTask, 0)
	root := filepath.Join(dirForBuild(buildSpec), "tasks")
	// if the directory does not exist, there are no tasks
	w := fs.WalkFS(root, buildStoreVFS(ctx))
	for w.Step() {
		if w.Err() != nil {
			break
		}
		if w.Stat().IsDir() {
			continue
		}
		f, err := buildStoreVFS(ctx).Open(w.Path())
		if err != nil {
			return nil, err
		}
		t := new(sourcegraph.BuildTask)
		err = json.NewDecoder(f).Decode(t)
		f.Close()
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	if err := w.Err(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	sort.Sort(byTaskID(tasks))

	return tasks, nil
}

func (s *Builds) GetTask(ctx context.Context, taskSpec sourcegraph.TaskSpec) (*sourcegraph.BuildTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getTask(ctx, taskSpec)
}

func (s *Builds) getTask(ctx context.Context, taskSpec sourcegraph.TaskSpec) (*sourcegraph.BuildTask, error) {
	f, err := buildStoreVFS(ctx).Open(filenameForTask(taskSpec))
	if err != nil {
		return nil, err
	}
	defer func() {
		var errors MultiError
		err = errors.verify(f.Close(), err)
	}()
	var t sourcegraph.BuildTask
	if err = json.NewDecoder(f).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

// MultiError collects a slice of errors and implements the error interface.
type MultiError struct{ Errs []error }

func (e MultiError) Error() string {
	switch len(e.Errs) {
	case 0:
		return "<nil>"
	case 1:
		return e.Errs[0].Error()
	default:
		errors := "multiple errors: "
		for k, err := range e.Errs {
			errors += err.Error()
			if k < len(e.Errs)-1 {
				errors += ", "
			}
		}
		return errors
	}
}

// err returns different values based on the state of the collection.
// If empty, it returns nil. If it contains one error, it returns it,
// otherwise it returns itself.
func (e MultiError) err() error {
	switch len(e.Errs) {
	case 0:
		return nil
	case 1:
		return e.Errs[0]
	default:
		return e
	}
}

// verify validates the passed error, and if it is non-nil it adds it to the
// collection, returning the resulting error
func (e *MultiError) verify(errs ...error) error {
	if e.Errs == nil {
		e.Errs = make([]error, 0, len(errs))
	}
	for _, err := range errs {
		if err == nil {
			continue
		}
		e.Errs = append(e.Errs, err)
	}
	return e.err()
}
