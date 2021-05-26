package executor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/git"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/mock"
)

func TestCoordinator_Execute(t *testing.T) {
	srcCLITask := &Task{Repository: testRepo1}
	sourcegraphTask := &Task{Repository: testRepo2}

	buildSpecFor := func(repo *graphql.Repository, modify func(*batches.ChangesetSpec)) *batches.ChangesetSpec {
		spec := &batches.ChangesetSpec{
			BaseRepository: repo.ID,
			CreatedChangeset: &batches.CreatedChangeset{
				BaseRef:        repo.BaseRef(),
				BaseRev:        repo.Rev(),
				HeadRepository: repo.ID,
				HeadRef:        "refs/heads/" + testChangesetTemplate.Branch,
				Title:          testChangesetTemplate.Title,
				Body:           testChangesetTemplate.Body,
				Commits: []batches.GitCommitDescription{
					{
						Message:     testChangesetTemplate.Commit.Message,
						AuthorName:  testChangesetTemplate.Commit.Author.Name,
						AuthorEmail: testChangesetTemplate.Commit.Author.Email,
						Diff:        `dummydiff1`,
					},
				},
				Published: false,
			},
		}

		modify(spec)
		return spec
	}

	tests := []struct {
		name string

		executor *dummyExecutor
		opts     NewCoordinatorOpts

		tasks     []*Task
		batchSpec *batches.BatchSpec

		wantCacheEntries int
		wantSpecs        []*batches.ChangesetSpec
		wantErrInclude   string
	}{
		{
			name: "success",

			tasks: []*Task{srcCLITask, sourcegraphTask},

			batchSpec: &batches.BatchSpec{
				Name:              "my-batch-change",
				Description:       "the description",
				ChangesetTemplate: testChangesetTemplate,
			},

			executor: &dummyExecutor{
				results: []taskResult{
					{task: srcCLITask, result: executionResult{Diff: `dummydiff1`}},
					{task: sourcegraphTask, result: executionResult{Diff: `dummydiff2`}},
				},
			},

			wantCacheEntries: 2,
			wantSpecs: []*batches.ChangesetSpec{
				buildSpecFor(testRepo1, func(spec *batches.ChangesetSpec) {
					spec.CreatedChangeset.Commits[0].Diff = `dummydiff1`
				}),
				buildSpecFor(testRepo2, func(spec *batches.ChangesetSpec) {
					spec.CreatedChangeset.Commits[0].Diff = `dummydiff2`
				}),
			},
		},
		{
			name:  "templated changesetTemplate",
			tasks: []*Task{srcCLITask},

			batchSpec: &batches.BatchSpec{
				Name:        "my-batch-change",
				Description: "the description",
				ChangesetTemplate: &batches.ChangesetTemplate{
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
					Commit: batches.ExpandedGitCommitDescription{
						Message: "output1=${{ outputs.output1}},output2=${{ outputs.output2.subField }}",
						Author: &batches.GitCommitAuthor{
							Name:  "output1=${{ outputs.output1}}",
							Email: "output1=${{ outputs.output1}}",
						},
					},
				},
			},

			executor: &dummyExecutor{
				results: []taskResult{
					{
						task: srcCLITask,
						result: executionResult{
							Diff: `dummydiff1`,
							Outputs: map[string]interface{}{
								"output1": "myOutputValue1",
								"output2": map[string]interface{}{
									"subField": "subFieldValue",
								},
							},
							ChangedFiles: &git.Changes{
								Modified: []string{"modified.txt"},
								Added:    []string{"added.txt"},
								Deleted:  []string{"deleted.txt"},
								Renamed:  []string{"renamed.txt"},
							},
						},
					},
				},
			},

			wantCacheEntries: 1,
			wantSpecs: []*batches.ChangesetSpec{
				buildSpecFor(testRepo1, func(spec *batches.ChangesetSpec) {
					spec.CreatedChangeset.HeadRef = "refs/heads/templated-branch-myOutputValue1"
					spec.CreatedChangeset.Title = "output1=myOutputValue1"
					spec.CreatedChangeset.Body = `output1=myOutputValue1
		output2=subFieldValue

		modified_files=[modified.txt]
		added_files=[added.txt]
		deleted_files=[deleted.txt]
		renamed_files=[renamed.txt]

		repository_name=github.com/sourcegraph/src-cli

		batch_change_name=my-batch-change
		batch_change_description=the description`
					spec.CreatedChangeset.Commits = []batches.GitCommitDescription{
						{
							Message:     "output1=myOutputValue1,output2=subFieldValue",
							AuthorName:  "output1=myOutputValue1",
							AuthorEmail: "output1=myOutputValue1",
							Diff:        `dummydiff1`,
						},
					}
				}),
			},
		},
		{
			name: "transform group",

			tasks: []*Task{srcCLITask, sourcegraphTask},

			batchSpec: &batches.BatchSpec{
				ChangesetTemplate: testChangesetTemplate,
				TransformChanges: &batches.TransformChanges{
					Group: []batches.Group{
						{Directory: "a/b/c", Branch: "in-directory-c"},
						{Directory: "a/b", Branch: "in-directory-b", Repository: testRepo2.Name},
					},
				},
			},

			executor: &dummyExecutor{
				results: []taskResult{
					{task: srcCLITask, result: executionResult{Diff: nestedChangesDiff}},
					{task: sourcegraphTask, result: executionResult{Diff: nestedChangesDiff}},
				},
			},

			// We have 4 ChangesetSpecs, but we only want 2 cache entries,
			// since we cache per Task, not per resulting changeset spec.
			wantCacheEntries: 2,
			wantSpecs: []*batches.ChangesetSpec{
				buildSpecFor(testRepo1, func(spec *batches.ChangesetSpec) {
					spec.CreatedChangeset.HeadRef = "refs/heads/" + testChangesetTemplate.Branch
					spec.CreatedChangeset.Commits[0].Diff = nestedChangesDiffSubdirA + nestedChangesDiffSubdirB
				}),
				buildSpecFor(testRepo2, func(spec *batches.ChangesetSpec) {
					spec.CreatedChangeset.HeadRef = "refs/heads/in-directory-b"
					spec.CreatedChangeset.Commits[0].Diff = nestedChangesDiffSubdirB + nestedChangesDiffSubdirC
				}),
				buildSpecFor(testRepo1, func(spec *batches.ChangesetSpec) {
					spec.CreatedChangeset.HeadRef = "refs/heads/in-directory-c"
					spec.CreatedChangeset.Commits[0].Diff = nestedChangesDiffSubdirC
				}),
				buildSpecFor(testRepo2, func(spec *batches.ChangesetSpec) {
					spec.CreatedChangeset.HeadRef = "refs/heads/" + testChangesetTemplate.Branch
					spec.CreatedChangeset.Commits[0].Diff = nestedChangesDiffSubdirA
				}),
			},
		},
		{
			name: "skip errors",
			opts: NewCoordinatorOpts{SkipErrors: true},

			tasks: []*Task{srcCLITask, sourcegraphTask},

			batchSpec: &batches.BatchSpec{
				Name:              "my-batch-change",
				Description:       "the description",
				ChangesetTemplate: testChangesetTemplate,
			},

			// Execution succeeded in srcCLIRepo, but fails in sourcegraphRepo
			executor: &dummyExecutor{
				results: []taskResult{
					{task: srcCLITask, result: executionResult{Diff: `dummydiff1`}},
				},
				waitErr: stepFailedErr{
					Err:         fmt.Errorf("something went wrong"),
					Run:         "broken command",
					Container:   "alpine:3",
					TmpFilename: "/tmp/foobar",
					Stderr:      "unknown command: broken",
				},
			},

			wantErrInclude: "run: broken command",
			// We want 1 cache entry and 1 spec
			wantCacheEntries: 1,
			wantSpecs: []*batches.ChangesetSpec{
				buildSpecFor(testRepo1, func(spec *batches.ChangesetSpec) {
					spec.CreatedChangeset.Commits[0].Diff = `dummydiff1`
				}),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			// Set attributes on Task which would be set by the TaskBuilder
			for _, t := range tc.tasks {
				t.TransformChanges = tc.batchSpec.TransformChanges
				t.Template = tc.batchSpec.ChangesetTemplate
				t.BatchChangeAttributes = &BatchChangeAttributes{
					Name:        tc.batchSpec.Name,
					Description: tc.batchSpec.Description,
				}
			}

			logManager := mock.LogNoOpManager{}

			cache := newInMemoryExecutionCache()
			noopPrinter := func([]*TaskStatus) {}
			coord := Coordinator{
				cache:      cache,
				exec:       tc.executor,
				logManager: logManager,
				opts:       tc.opts,
			}

			// execute contains the actual logic for executing the tasks and
			// the batch spec. We'll run this multiple times to cover both the
			// cache and non-cache code paths.
			execute := func(t *testing.T) {
				specs, _, err := coord.Execute(ctx, tc.tasks, tc.batchSpec, noopPrinter)
				if tc.wantErrInclude == "" {
					if err != nil {
						t.Fatalf("execution failed: %s", err)
					}
				} else {
					if err == nil {
						t.Fatalf("expected error to include %q, but got no error", tc.wantErrInclude)
					} else {
						if !strings.Contains(err.Error(), tc.wantErrInclude) {
							t.Errorf("wrong error. have=%q want included=%q", err, tc.wantErrInclude)
						}
					}
				}

				if have, want := len(specs), len(tc.wantSpecs); have != want {
					t.Fatalf("wrong number of changeset specs. want=%d, have=%d", want, have)
				}

				opts := []cmp.Option{
					cmpopts.EquateEmpty(),
					cmpopts.SortSlices(func(a, b *batches.ChangesetSpec) bool {
						if a.BaseRepository == b.BaseRepository && a.CreatedChangeset != nil && b.CreatedChangeset != nil {
							return a.CreatedChangeset.HeadRef < b.CreatedChangeset.HeadRef
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
				if have, want := cache.size(), tc.wantCacheEntries; have != want {
					t.Errorf("unexpected number of cache entries: have=%d want=%d cache=%+v", have, want, cache)
				}
			}

			// Sanity check, since we're going to be looking at the side effects
			// on the cache.
			if cache.size() != 0 {
				t.Fatalf("unexpectedly hot cache: %+v", cache)
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
	cache := newInMemoryExecutionCache()

	task := &Task{
		Steps: []batches.Step{
			{Run: `echo "one"`},
			{Run: `echo "two"`},
			{Run: `echo "three"`},
		},
		Repository:            testRepo1,
		BatchChangeAttributes: &BatchChangeAttributes{},
	}

	executor := &dummyExecutor{}
	executor.results = []taskResult{{
		task: task,
		result: executionResult{
			Diff:         "dummydiff",
			ChangedFiles: &git.Changes{},
			Outputs:      map[string]interface{}{},
			Path:         "",
		},
		stepResults: []stepExecutionResult{
			{StepIndex: 0, Diff: []byte(`step-0-diff`)},
			{StepIndex: 1, Diff: []byte(`step-1-diff`)},
			{StepIndex: 2, Diff: []byte(`step-2-diff`)},
		},
	}}

	// First execution. Make sure that the Task executes all steps.
	execAndEnsure(t, cache, executor, task, assertNoCachedResult(t))
	// We now expect the cache to have 1+N entries: 1 for the complete task, N
	// for the steps.
	wantCacheSize := len(task.Steps) + 1
	assertCacheSize(t, cache, wantCacheSize)

	// Change the 2nd step's definition:
	task.Steps[1].Run = `echo "two modified"`
	// Re-execution should start with the diff produced by steps[0] as the
	// start state from which steps[1] is then re-executed.
	execAndEnsure(t, cache, executor, task, assertCachedResultForStep(t, 0))
	// Cache now contains old entries, plus another "complete task" entry and
	// two entries for newly executed steps.
	wantCacheSize += 1 + 2
	assertCacheSize(t, cache, wantCacheSize)

	// Change the 3rd step's definition:
	task.Steps[2].Run = `echo "three modified"`
	// Re-execution should use the diff from steps[1] as start state
	execAndEnsure(t, cache, executor, task, assertCachedResultForStep(t, 1))
	// Cache now contains old entries, plus another "complete task" entry and
	// a single new step entry
	wantCacheSize += 1 + 1
	assertCacheSize(t, cache, wantCacheSize)
}

// execAndEnsure executes the given Task with the given cache and dummyExecutor
// in a new Coordinator, setting cb as the startCallback on the executor.
func execAndEnsure(t *testing.T, cache ExecutionCache, exec *dummyExecutor, task *Task, cb startCallback) {
	t.Helper()

	// Setup dependencies
	batchSpec := &batches.BatchSpec{ChangesetTemplate: testChangesetTemplate}
	logManager := mock.LogNoOpManager{}
	noopPrinter := func([]*TaskStatus) {}

	// Build Coordinator
	coord := &Coordinator{cache: cache, exec: exec, logManager: logManager}

	// Set the ChangesetTemplate on Task
	task.Template = batchSpec.ChangesetTemplate

	// Setup the callback
	exec.startCb = cb

	// Execute
	specs, _, err := coord.Execute(context.Background(), []*Task{task}, batchSpec, noopPrinter)
	if err != nil {
		t.Fatalf("execution of task failed: %s", err)
	}

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

// assertCachedResultForStep returns a function that can be used as a
// startCallback on dummyExecutor to assert that the first Task has a cached
// result for the given step.
func assertCachedResultForStep(t *testing.T, step int) func(context.Context, []*Task, taskStatusHandler) {
	return func(c context.Context, tasks []*Task, tsh taskStatusHandler) {
		t.Helper()

		task := tasks[0]
		if !task.CachedResultFound {
			t.Fatalf("CachedResultFound not set")
		}

		if have, want := task.CachedResult.StepIndex, step; have != want {
			t.Fatalf("CachedResult.Step wrong. have=%d, want=%d", have, want)
		}
	}
}

// expectCachedResultForStep returns a function that can be used as a
// startCallback on dummyExecutor to assert that the first Task has no cached results.
func assertNoCachedResult(t *testing.T) func(context.Context, []*Task, taskStatusHandler) {
	return func(c context.Context, tasks []*Task, tsh taskStatusHandler) {
		t.Helper()

		task := tasks[0]
		if task.CachedResultFound {
			t.Fatalf("CachedResultFound but not expected")
		}
	}
}

type startCallback func(context.Context, []*Task, taskStatusHandler)

type dummyExecutor struct {
	startCb       startCallback
	startCbCalled bool

	results []taskResult
	waitErr error
}

func (d *dummyExecutor) Start(ctx context.Context, ts []*Task, status taskStatusHandler) {
	if d.startCb != nil {
		d.startCb(ctx, ts, status)
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

func (c *inMemoryExecutionCache) getCacheItem(key CacheKeyer) (interface{}, bool, error) {
	k, err := key.Key()
	if err != nil {
		return executionResult{}, false, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	res, ok := c.cache[k]
	return res, ok, nil
}

func (c *inMemoryExecutionCache) Get(ctx context.Context, key CacheKeyer) (executionResult, bool, error) {
	res, ok, err := c.getCacheItem(key)
	if err != nil || !ok {
		return executionResult{}, ok, err
	}

	execResult, ok := res.(executionResult)
	return execResult, ok, nil
}

func (c *inMemoryExecutionCache) Set(ctx context.Context, key CacheKeyer, result executionResult) error {
	k, err := key.Key()
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[k] = result
	return nil
}

func (c *inMemoryExecutionCache) GetStepResult(ctx context.Context, key CacheKeyer) (stepExecutionResult, bool, error) {
	res, ok, err := c.getCacheItem(key)
	if err != nil || !ok {
		return stepExecutionResult{}, ok, err
	}

	execResult, ok := res.(stepExecutionResult)
	return execResult, ok, nil
}

func (c *inMemoryExecutionCache) SetStepResult(ctx context.Context, key CacheKeyer, result stepExecutionResult) error {
	k, err := key.Key()
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[k] = result
	return nil
}

func (c *inMemoryExecutionCache) Clear(ctx context.Context, key CacheKeyer) error {
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

const nestedChangesDiff = nestedChangesDiffSubdirA + nestedChangesDiffSubdirB + nestedChangesDiffSubdirC
