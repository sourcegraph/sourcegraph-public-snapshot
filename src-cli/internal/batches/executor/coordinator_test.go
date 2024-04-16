package executor

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/overridable"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution/cache"
	"github.com/sourcegraph/sourcegraph/lib/batches/git"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"

	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/mock"
	"github.com/sourcegraph/src-cli/internal/batches/util"
)

func TestCoordinator_Execute(t *testing.T) {
	publishedFalse := overridable.FromBoolOrString(false)
	srcCLITask := &Task{Repository: testRepo1, Steps: []batcheslib.Step{{Run: "echo Hello World"}}}
	sourcegraphTask := &Task{Repository: testRepo2, Steps: []batcheslib.Step{{Run: "echo Hello Sourcegraph"}}}

	buildSpecFor := func(repo *graphql.Repository, modify func(*batcheslib.ChangesetSpec)) *batcheslib.ChangesetSpec {
		spec := &batcheslib.ChangesetSpec{
			BaseRepository: repo.ID,

			BaseRef:        repo.BaseRef(),
			BaseRev:        repo.Rev(),
			HeadRepository: repo.ID,
			HeadRef:        util.EnsureRefPrefix(testChangesetTemplate.Branch),
			Title:          testChangesetTemplate.Title,
			Body:           testChangesetTemplate.Body,
			Commits: []batcheslib.GitCommitDescription{
				{
					Version:     2,
					Message:     testChangesetTemplate.Commit.Message,
					AuthorName:  testChangesetTemplate.Commit.Author.Name,
					AuthorEmail: testChangesetTemplate.Commit.Author.Email,
					Diff:        []byte(`dummydiff1`),
				},
			},
			Published: batcheslib.PublishedValue{Val: false},
		}

		modify(spec)
		return spec
	}

	tests := []struct {
		name string

		executor *dummyExecutor
		opts     NewCoordinatorOpts

		tasks     []*Task
		batchSpec *batcheslib.BatchSpec

		wantCacheEntries int
		wantSpecs        []*batcheslib.ChangesetSpec
		wantErrInclude   string
	}{
		{
			name: "success",

			tasks: []*Task{srcCLITask, sourcegraphTask},

			batchSpec: &batcheslib.BatchSpec{
				Name:              "my-batch-change",
				Description:       "the description",
				ChangesetTemplate: testChangesetTemplate,
			},

			executor: &dummyExecutor{
				results: []taskResult{
					{task: srcCLITask, stepResults: []execution.AfterStepResult{{Version: 2, Diff: []byte(`dummydiff1`)}}},
					{task: sourcegraphTask, stepResults: []execution.AfterStepResult{{Version: 2, Diff: []byte(`dummydiff2`)}}},
				},
			},
			opts: NewCoordinatorOpts{},

			wantCacheEntries: 2,
			wantSpecs: []*batcheslib.ChangesetSpec{
				buildSpecFor(testRepo1, func(spec *batcheslib.ChangesetSpec) {
					spec.Commits[0].Diff = []byte(`dummydiff1`)
				}),
				buildSpecFor(testRepo2, func(spec *batcheslib.ChangesetSpec) {
					spec.Commits[0].Diff = []byte(`dummydiff2`)
				}),
			},
		},
		{
			name:  "templated changesetTemplate",
			tasks: []*Task{srcCLITask},

			batchSpec: &batcheslib.BatchSpec{
				Name:        "my-batch-change",
				Description: "the description",
				ChangesetTemplate: &batcheslib.ChangesetTemplate{
					Title: `output1=${{ outputs.output1}}`,
					Body: `output1=${{ outputs.output1}}
		output2=${{ outputs.output2.subField }}

		modified_files=${{ steps.modified_files }}
		added_files=${{ steps.added_files }}
		deleted_files=${{ steps.deleted_files }}
		renamed_files=${{ steps.renamed_files }}

		repository_name=${{ repository.name }}

		batch_change_name=${{ batch_change.name }}
		batch_change_description=${{ batch_change.description }}
		`,
					Branch: "templated-branch-${{ outputs.output1 }}",
					Commit: batcheslib.ExpandedGitCommitDescription{
						Message: "output1=${{ outputs.output1}},output2=${{ outputs.output2.subField }}",
						Author: &batcheslib.GitCommitAuthor{
							Name:  "output1=${{ outputs.output1}}",
							Email: "output1=${{ outputs.output1}}",
						},
					},
					Published: &publishedFalse,
				},
			},

			executor: &dummyExecutor{
				results: []taskResult{
					{
						task: srcCLITask,
						stepResults: []execution.AfterStepResult{{
							Version: 2,
							Diff:    []byte(`dummydiff1`),
							Outputs: map[string]interface{}{
								"output1": "myOutputValue1",
								"output2": map[string]interface{}{
									"subField": "subFieldValue",
								},
							},
							ChangedFiles: git.Changes{
								Modified: []string{"modified.txt"},
								Added:    []string{"added.txt"},
								Deleted:  []string{"deleted.txt"},
								Renamed:  []string{"renamed.txt"},
							},
						}},
					},
				},
			},
			opts: NewCoordinatorOpts{},

			wantCacheEntries: 1,
			wantSpecs: []*batcheslib.ChangesetSpec{
				buildSpecFor(testRepo1, func(spec *batcheslib.ChangesetSpec) {
					spec.HeadRef = "refs/heads/templated-branch-myOutputValue1"
					spec.Title = "output1=myOutputValue1"
					spec.Body = `output1=myOutputValue1
		output2=subFieldValue

		modified_files=[modified.txt]
		added_files=[added.txt]
		deleted_files=[deleted.txt]
		renamed_files=[renamed.txt]

		repository_name=github.com/sourcegraph/src-cli

		batch_change_name=my-batch-change
		batch_change_description=the description`
					spec.Commits = []batcheslib.GitCommitDescription{
						{
							Version:     2,
							Message:     "output1=myOutputValue1,output2=subFieldValue",
							AuthorName:  "output1=myOutputValue1",
							AuthorEmail: "output1=myOutputValue1",
							Diff:        []byte(`dummydiff1`),
						},
					}
				}),
			},
		},
		{
			name: "transform group",

			tasks: []*Task{srcCLITask, sourcegraphTask},

			batchSpec: &batcheslib.BatchSpec{
				ChangesetTemplate: testChangesetTemplate,
				TransformChanges: &batcheslib.TransformChanges{
					Group: []batcheslib.Group{
						{Directory: "a/b/c", Branch: "in-directory-c"},
						{Directory: "a/b", Branch: "in-directory-b", Repository: testRepo2.Name},
					},
				},
			},

			executor: &dummyExecutor{
				results: []taskResult{
					{task: srcCLITask, stepResults: []execution.AfterStepResult{{Version: 2, Diff: nestedChangesDiff}}},
					{task: sourcegraphTask, stepResults: []execution.AfterStepResult{{Version: 2, Diff: nestedChangesDiff}}},
				},
			},
			opts: NewCoordinatorOpts{},

			// TODO: Fix comment.
			// We have 4 ChangesetSpecs, but we only want 2 cache entries,
			// since we cache per Task, not per resulting changeset spec.
			wantCacheEntries: 2,
			wantSpecs: []*batcheslib.ChangesetSpec{
				buildSpecFor(testRepo1, func(spec *batcheslib.ChangesetSpec) {
					spec.HeadRef = "refs/heads/" + testChangesetTemplate.Branch
					spec.Commits[0].Diff = []byte(nestedChangesDiffSubdirA + nestedChangesDiffSubdirB)
				}),
				buildSpecFor(testRepo2, func(spec *batcheslib.ChangesetSpec) {
					spec.HeadRef = "refs/heads/in-directory-b"
					spec.Commits[0].Diff = []byte(nestedChangesDiffSubdirB + nestedChangesDiffSubdirC)
				}),
				buildSpecFor(testRepo1, func(spec *batcheslib.ChangesetSpec) {
					spec.HeadRef = "refs/heads/in-directory-c"
					spec.Commits[0].Diff = []byte(nestedChangesDiffSubdirC)
				}),
				buildSpecFor(testRepo2, func(spec *batcheslib.ChangesetSpec) {
					spec.HeadRef = util.EnsureRefPrefix(testChangesetTemplate.Branch)
					spec.Commits[0].Diff = []byte(nestedChangesDiffSubdirA)
				}),
			},
		},
		{
			name: "cache for step mount",

			tasks: []*Task{srcCLITask, sourcegraphTask},

			batchSpec: &batcheslib.BatchSpec{
				Name:              "my-batch-change",
				Description:       "the description",
				ChangesetTemplate: testChangesetTemplate,
				Steps: []batcheslib.Step{
					{
						Run:   "echo foo",
						Mount: []batcheslib.Mount{{Path: "/foo/bar/sample.sh", Mountpoint: "/tmp/sample.sh"}},
					},
				},
			},

			executor: &dummyExecutor{
				results: []taskResult{
					{task: srcCLITask, stepResults: []execution.AfterStepResult{{Version: 2, Diff: []byte(`dummydiff1`), StepIndex: 0}}},
					{task: sourcegraphTask, stepResults: []execution.AfterStepResult{{Version: 2, Diff: []byte(`dummydiff2`), StepIndex: 0}}},
				},
			},
			opts: NewCoordinatorOpts{},

			wantCacheEntries: 2,
			wantSpecs: []*batcheslib.ChangesetSpec{
				buildSpecFor(testRepo1, func(spec *batcheslib.ChangesetSpec) {
					spec.Commits[0].Diff = []byte(`dummydiff1`)
				}),
				buildSpecFor(testRepo2, func(spec *batcheslib.ChangesetSpec) {
					spec.Commits[0].Diff = []byte(`dummydiff2`)
				}),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			// Set attributes on Task which would be set by the TaskBuilder
			for _, t := range tc.tasks {
				t.BatchChangeAttributes = &template.BatchChangeAttributes{
					Name:        tc.batchSpec.Name,
					Description: tc.batchSpec.Description,
				}
			}

			logManager := mock.LogNoOpManager{}
			c := newInMemoryExecutionCache()

			tc.opts.Cache = c
			tc.opts.Logger = logManager
			tc.opts.BinaryDiffs = true

			coord := Coordinator{
				exec: tc.executor,
				opts: tc.opts,
			}

			// execute contains the actual logic for executing the tasks and
			// the batch spec. We'll run this multiple times to cover both the
			// cache and non-cache code paths.
			execute := func(t *testing.T) {
				specs, _, err := coord.ExecuteAndBuildSpecs(ctx, tc.batchSpec, tc.tasks, newDummyTaskExecutionUI())
				if tc.wantErrInclude == "" {
					if err != nil {
						t.Fatalf("execution failed: %s", err)
					}
				} else {
					if err == nil {
						t.Fatalf("expected error to include %q, but got no error", tc.wantErrInclude)
					}
					if !strings.Contains(err.Error(), tc.wantErrInclude) {
						t.Errorf("wrong error. have=%q want included=%q", err, tc.wantErrInclude)
					}
				}

				if have, want := len(specs), len(tc.wantSpecs); have != want {
					t.Fatalf("wrong number of changeset specs. want=%d, have=%d", want, have)
				}

				opts := []cmp.Option{
					cmpopts.EquateEmpty(),
					cmpopts.SortSlices(func(a, b *batcheslib.ChangesetSpec) bool {
						if a.BaseRepository == b.BaseRepository {
							return a.HeadRef < b.HeadRef
						}
						return a.BaseRepository < b.BaseRepository
					}),
				}
				if !cmp.Equal(tc.wantSpecs, specs, opts...) {
					t.Errorf("wrong ChangesetSpecs (-want +got):\n%s", cmp.Diff(tc.wantSpecs, specs, opts...))
				}
			}

			verifyCache := func(t *testing.T) {
				// Verify that there is a cache entry for each repo.
				if have, want := c.size(), tc.wantCacheEntries; have != want {
					t.Errorf("unexpected number of cache entries: have=%d want=%d cache=%+v", have, want, c)
				}
			}

			// Sanity check, since we're going to be looking at the side effects
			// on the cache.
			if c.size() != 0 {
				t.Fatalf("unexpected hot cache: %+v", c)
			}

			// Run with a cold cache.
			t.Run("cold cache", func(t *testing.T) {
				execute(t)
				verifyCache(t)
			})

			// Run with a warm cache.
			t.Run("warm cache", func(t *testing.T) {
				execute(t)
				verifyCache(t)
			})
		})
	}
}

func TestCoordinator_Execute_StepCaching(t *testing.T) {
	// Setup dependencies
	cache := newInMemoryExecutionCache()
	logManager := mock.LogNoOpManager{}

	task := &Task{
		Steps: []batcheslib.Step{
			{Run: `echo "one"`},
			{Run: `echo "two"`},
			{Run: `echo "three"`},
		},
		Repository:            testRepo1,
		BatchChangeAttributes: &template.BatchChangeAttributes{},
	}

	executor := &dummyExecutor{}
	executor.results = []taskResult{{
		task: task,
		stepResults: []execution.AfterStepResult{
			{Version: 2, StepIndex: 0, Diff: []byte(`step-0-diff`)},
			{Version: 2, StepIndex: 1, Diff: []byte(`step-1-diff`)},
			{Version: 2, StepIndex: 2, Diff: []byte(`step-2-diff`)},
		},
	}}

	// Build Coordinator
	coord := &Coordinator{
		opts: NewCoordinatorOpts{
			Cache:  cache,
			Logger: logManager,
		},
		exec: executor,
	}

	batchSpec := &batcheslib.BatchSpec{ChangesetTemplate: testChangesetTemplate}

	// First execution. Make sure that the Task executes all steps.
	execAndEnsure(t, coord, executor, batchSpec, task, assertNoCachedResult(t))
	// We now expect the cache to have N entries for N steps.
	assertCacheSize(t, cache, 3)

	// Reset task
	task.CachedStepResultFound = false

	// Change the 2nd step's definition:
	task.Steps[1].Run = `echo "two modified"`
	// Re-execution should start with the diff produced by steps[0] as the
	// start state from which steps[1] is then re-executed.
	execAndEnsure(t, coord, executor, batchSpec, task, func(ctx context.Context, t []*Task, teu TaskExecutionUI) {})
	// Cache now contains old entries, plus another "complete task" entry and
	// two entries for newly executed steps.
	assertCacheSize(t, cache, 5)

	// Reset task
	task.CachedStepResultFound = false

	// Change the 3rd step's definition:
	task.Steps[2].Run = `echo "three modified"`
	// Re-execution should use the diff from steps[1] as start state
	execAndEnsure(t, coord, executor, batchSpec, task, func(ctx context.Context, t []*Task, teu TaskExecutionUI) {})
	// Cache now contains old entries, plus another "complete task" entry and
	// a single new step entry
	assertCacheSize(t, cache, 6)

	// Reset task
	task.CachedStepResultFound = false

	// Now we execute the spec with -clear-cache:
	if err := coord.ClearCache(context.Background(), []*Task{task}); err != nil {
		t.Fatal(err)
	}
	assertCacheSize(t, cache, 3)

	// We don't want any cached results set on the task:
	execAndEnsure(t, coord, executor, batchSpec, task, assertNoCachedResult(t))
	// Cache should have the same number of entries: the cached step results should
	// have been cleared (the complete-task-result is cleared in another
	// code path) and the same amount of cached entries has been added.
	assertCacheSize(t, cache, 6)
}

// execAndEnsure executes the given Task with the given cache and dummyExecutor
// in a new Coordinator, setting cb as the startCallback on the executor.
func execAndEnsure(t *testing.T, coord *Coordinator, exec *dummyExecutor, batchSpec *batcheslib.BatchSpec, task *Task, cb startCallback) {
	t.Helper()

	// Setup the callback
	exec.startCb = cb

	// Check cache
	uncached, cachedSpecs, err := coord.CheckCache(context.Background(), batchSpec, []*Task{task})
	if err != nil {
		t.Fatal(err)
	}

	// Execute
	freshSpecs, _, err := coord.ExecuteAndBuildSpecs(context.Background(), batchSpec, uncached, newDummyTaskExecutionUI())
	if err != nil {
		t.Fatalf("execution of task failed: %s", err)
	}

	specs := append(cachedSpecs, freshSpecs...)

	// Sanity check, because we're not interested in the specs
	if have, want := len(specs), 1; have != want {
		t.Fatalf("Wrong number of specs. want=%d, have=%d", want, have)
	}

	// Ensure callback was called
	if !exec.startCbCalled {
		t.Fatalf("expected the startCallback to be called, but was not")
	}
	exec.startCbCalled = false
}

// assertCacheSize asserts the cache's size.
func assertCacheSize(t *testing.T, cache *inMemoryExecutionCache, want int) {
	t.Helper()

	if have := cache.size(); have != want {
		t.Fatalf("wrong cache size. have=%d, want=%d", have, want)
	}
}

// expectCachedResultForStep returns a function that can be used as a
// startCallback on dummyExecutor to assert that the first Task has no cached results.
func assertNoCachedResult(t *testing.T) func(context.Context, []*Task, TaskExecutionUI) {
	return func(c context.Context, tasks []*Task, ui TaskExecutionUI) {
		t.Helper()

		task := tasks[0]
		if task.CachedStepResultFound {
			t.Fatalf("CachedResultFound but not expected")
		}
	}
}

type startCallback func(context.Context, []*Task, TaskExecutionUI)

var _ TaskExecutionUI = &dummyTaskExecutionUI{}

func newDummyTaskExecutionUI() *dummyTaskExecutionUI {
	return &dummyTaskExecutionUI{
		started:         map[*Task]struct{}{},
		finished:        map[*Task]struct{}{},
		finishedWithErr: map[*Task]struct{}{},
		specs:           map[*Task][]*batcheslib.ChangesetSpec{},
	}
}

type dummyTaskExecutionUI struct {
	mu sync.Mutex

	started         map[*Task]struct{}
	finished        map[*Task]struct{}
	finishedWithErr map[*Task]struct{}
	specs           map[*Task][]*batcheslib.ChangesetSpec
}

func (d *dummyTaskExecutionUI) Start([]*Task)    {}
func (d *dummyTaskExecutionUI) Success()         {}
func (d *dummyTaskExecutionUI) Failed(err error) {}
func (d *dummyTaskExecutionUI) TaskStarted(t *Task) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.started[t] = struct{}{}
}
func (d *dummyTaskExecutionUI) TaskFinished(t *Task, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.started, t)
	if err != nil {
		d.finishedWithErr[t] = struct{}{}
	} else {
		d.finished[t] = struct{}{}
	}
}
func (d *dummyTaskExecutionUI) TaskChangesetSpecsBuilt(t *Task, specs []*batcheslib.ChangesetSpec) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.specs[t] = specs
}

func (d *dummyTaskExecutionUI) StepsExecutionUI(t *Task) StepsExecutionUI {
	return NoopStepsExecUI{}
}

var _ taskExecutor = &dummyExecutor{}

type dummyExecutor struct {
	startCb       startCallback
	startCbCalled bool

	results []taskResult
	waitErr error
}

func (d *dummyExecutor) Start(ctx context.Context, ts []*Task, ui TaskExecutionUI) {
	if d.startCb != nil {
		d.startCb(ctx, ts, ui)
		d.startCbCalled = true
	}
	// "noop noop noop", the crowd screams
}

func (d *dummyExecutor) Wait(context.Context) ([]taskResult, error) {
	return d.results, d.waitErr
}

// inMemoryExecutionCache provides an in-memory cache for testing purposes.
type inMemoryExecutionCache struct {
	cache map[string]interface{}
	mu    sync.RWMutex
}

func newInMemoryExecutionCache() *inMemoryExecutionCache {
	return &inMemoryExecutionCache{
		cache: make(map[string]interface{}),
	}
}

func (c *inMemoryExecutionCache) size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache)
}

func (c *inMemoryExecutionCache) getCacheItem(key cache.Keyer) (interface{}, bool, error) {
	k, err := key.Key()
	if err != nil {
		return execution.AfterStepResult{}, false, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	res, ok := c.cache[k]
	return res, ok, nil
}

func (c *inMemoryExecutionCache) Get(ctx context.Context, key cache.Keyer) (execution.AfterStepResult, bool, error) {
	res, ok, err := c.getCacheItem(key)
	if err != nil || !ok {
		return execution.AfterStepResult{}, ok, err
	}

	execResult, ok := res.(execution.AfterStepResult)
	return execResult, ok, nil
}

func (c *inMemoryExecutionCache) Set(ctx context.Context, key cache.Keyer, result execution.AfterStepResult) error {
	k, err := key.Key()
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[k] = result
	return nil
}

func (c *inMemoryExecutionCache) Clear(ctx context.Context, key cache.Keyer) error {
	k, err := key.Key()
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, k)
	return nil
}

const nestedChangesDiffSubdirA = `diff --git a/a/a.go b/a/a.go
index 2a93cde..a83f668 100644
--- a/a/a.go
+++ b/a/a.go
@@ -1,1 +1,3 @@
 package a
+
+var a = 1
`

const nestedChangesDiffSubdirB = `diff --git a/a/b/b.go b/a/b/b.go
index e0836a8..c977beb 100644
--- a/a/b/b.go
+++ b/a/b/b.go
@@ -1,1 +1,3 @@
 package b
+
+var b = 2
`

const nestedChangesDiffSubdirC = `diff --git a/a/b/c/b.go b/a/b/c/b.go
index 7f96c22..43df362 100644
--- a/a/b/c/b.go
+++ b/a/b/c/b.go
@@ -1,1 +1,3 @@
 package c
+
+var c = 3
`

var nestedChangesDiff = []byte(nestedChangesDiffSubdirA + nestedChangesDiffSubdirB + nestedChangesDiffSubdirC)
