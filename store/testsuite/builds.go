package testsuite

import (
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/jsonutil"
)

// InsertBuildsFunc is called at the beginning of Builds_* test funcs
// to insert or mock the builds data source.
type InsertBuildsFunc func(ctx context.Context, t *testing.T, mockBuilds []*sourcegraph.Build)

// InsertTasksFunc is called at the beginning of test funcs that need
// to add tasks during their setup.
type InsertTasksFunc func(ctx context.Context, t *testing.T, mockTasks []*sourcegraph.BuildTask)

// ValidateQueueEntryFunc is called by tests that wish to validate the existence
// of a build in a queue. This is left to the implementation layer because the
// check method may vary.
type ValidateQueueEntryFunc func(ctx context.Context, want sourcegraph.BuildSpec, t *testing.T) bool

// assertBuildExists verifies that a build exists in the store by using its Get method.
func assertBuildExists(ctx context.Context, s store.Builds, want *sourcegraph.Build, t *testing.T) {
	b, err := s.Get(ctx, want.Spec())
	if err != nil {
		t.Fatalf("errored out: %s", err)
	}
	if !reflect.DeepEqual(b, want) {
		t.Errorf("expected %#v, got %#v", want, b)
	}
}

// assertTaskExists verifies that a build exists in the store by using its GetTask method.
func assertTaskExists(ctx context.Context, s store.Builds, want *sourcegraph.BuildTask, t *testing.T) {
	b, err := s.GetTask(ctx, want.Spec())
	if err != nil {
		t.Fatalf("errored out: %s", err)
	}
	if !reflect.DeepEqual(b, want) {
		t.Errorf("expected %#v, got %#v", want, b)
	}
}

// Builds_Update_builderConfig tests that updating BuilderConfig only updates
// the BuilderConfig without affecting other fields.
func Builds_Update_builderConfig(ctx context.Context, t *testing.T, s store.Builds, insert InsertBuildsFunc) {
	t0 := pbtypes.NewTimestamp(time.Unix(1, 0))
	orig := &sourcegraph.Build{ID: 33, Repo: "y/y", CommitID: strings.Repeat("a", 40), Host: "localhost", StartedAt: &t0}
	insert(ctx, t, []*sourcegraph.Build{orig})

	update := sourcegraph.BuildUpdate{BuilderConfig: "test"}
	err := s.Update(ctx, orig.Spec(), update)
	if err != nil {
		t.Fatalf("errored out: %s", err)
	}
	want := *orig
	want.BuilderConfig = update.BuilderConfig
	assertBuildExists(ctx, s, &want, t)
}

func Builds_DequeueNext(ctx context.Context, t *testing.T, s store.Builds, insert InsertBuildsFunc) {
	want := &sourcegraph.Build{ID: 5, Repo: "x/x", CommitID: strings.Repeat("a", 40), Host: "localhost", BuildConfig: sourcegraph.BuildConfig{Queue: true}}
	insert(ctx, t, []*sourcegraph.Build{want})
	build, err := s.DequeueNext(ctx)
	if err != nil {
		t.Fatalf("errored out: %s", err)
	}
	if build.StartedAt == nil {
		t.Errorf("got dequeued build StartedAt null, want it to be set to appx. now")
	}
	build.StartedAt = nil // don't compare since StartedAt is set from the current time
	if !reflect.DeepEqual(build, want) {
		t.Errorf("expected %#v, got %#v", want, build)
	}
}

func Builds_DequeueNext_ordered(ctx context.Context, t *testing.T, s store.Builds, insert InsertBuildsFunc) {
	t1 := pbtypes.NewTimestamp(time.Unix(100000, 0))
	t2 := pbtypes.NewTimestamp(time.Unix(200000, 0))

	b1 := &sourcegraph.Build{ID: 1, CommitID: strings.Repeat("A", 40), Repo: "r", CreatedAt: t1, BuildConfig: sourcegraph.BuildConfig{Queue: true, Priority: 10}}
	b2 := &sourcegraph.Build{ID: 2, CommitID: strings.Repeat("A", 40), Repo: "r", CreatedAt: t1, BuildConfig: sourcegraph.BuildConfig{Queue: true}}
	b3 := &sourcegraph.Build{ID: 3, CommitID: strings.Repeat("A", 40), Repo: "r", CreatedAt: t2, BuildConfig: sourcegraph.BuildConfig{Queue: true}}
	bNo1 := &sourcegraph.Build{ID: 4, CommitID: strings.Repeat("A", 40), Repo: "r", BuildConfig: sourcegraph.BuildConfig{Queue: false}}
	bNo2 := &sourcegraph.Build{ID: 5, CommitID: strings.Repeat("A", 40), Repo: "r", StartedAt: &t2, BuildConfig: sourcegraph.BuildConfig{Queue: true}}

	insert(ctx, t, []*sourcegraph.Build{b1, b2, b3, bNo1, bNo2})

	wantBuilds := []*sourcegraph.Build{
		b1, b2, b3, nil, // in order
	}

	for i, wantBuild := range wantBuilds {
		build, err := s.DequeueNext(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if build != nil {
			if build.StartedAt == nil {
				t.Errorf("got dequeued build #%d StartedAt null, want it to be set to appx. now", i+1)
			}
			build.StartedAt = nil // don't compare since StartedAt is set from the current time
		}
		if !jsonutil.JSONEqual(t, build, wantBuild) {
			t.Errorf("dequeued build #%d\n\nGOT\n%+v\n\nWANT\n%+v", i+1, build, wantBuild)
		}
	}
}

// Builds_DequeueNext_noRaceCondition ensures that DequeueNext will dequeue a
// build exactly once and that concurrent processes will not dequeue the same
// build. It may not always trigger the race condition, but if it even does
// once, it is very important that we fix it.
func Builds_DequeueNext_noRaceCondition(ctx context.Context, t *testing.T, s store.Builds, insert InsertBuildsFunc) {
	const (
		nbuilds  = 90
		nworkers = 30
	)

	var allBuilds []*sourcegraph.Build
	for i := 0; i < nbuilds; i++ {
		allBuilds = append(allBuilds, &sourcegraph.Build{
			ID:          uint64(i + 1),
			Repo:        "r",
			BuildConfig: sourcegraph.BuildConfig{Queue: true, Priority: int32(i)},
			CommitID:    strings.Repeat("a", 40),
		})
	}

	insert(ctx, t, allBuilds)
	t.Logf("enqueued %d builds", nbuilds)

	dq := map[uint64]bool{} // build attempt -> whether it has already been dequeued
	var dqMu sync.Mutex

	var wg sync.WaitGroup
	for i := 0; i < nworkers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for {
				b, err := s.DequeueNext(ctx)
				if err != nil {
					t.Fatal(err)
				}
				if b == nil {
					return
				}

				dqMu.Lock()
				if dq[b.ID] {
					dqMu.Unlock()
					t.Errorf("build %d was already dequeued (race condition)", b.ID)
					return
				}
				dq[b.ID] = true
				dqMu.Unlock()
				t.Logf("worker %d got build %d (priority %d)", i, b.ID, b.Priority)
			}
		}(i)
	}
	wg.Wait()

	for _, b := range allBuilds {
		if !dq[b.ID] {
			t.Errorf("build %d was never dequeued", b.ID)
		}
	}
}

// Builds_CreateTasks verifies that inserting a series of tasks via Builds.CreateTasks correctly
// creates these tasks in the store. The existence is asserted using the assertTaskExists method.
func Builds_CreateTasks(ctx context.Context, t *testing.T, s store.Builds, _ InsertTasksFunc) {
	tasks := []*sourcegraph.BuildTask{
		{ID: 1, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 1}, Label: "a"},
		{ID: 2, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 1}, Label: "a"},
		{ID: 3, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 2}, Label: "b"},
		{ID: 4, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "x/z"}, ID: 1}, Label: "b"},
	}
	tsk, err := s.CreateTasks(ctx, tasks)
	if err != nil {
		t.Fatalf("errored out: %s", err)
	}
	if !reflect.DeepEqual(tsk, tasks) {
		t.Errorf("created tasks do not match params. Expected %#v, got %#v", tasks, t)
	}
	for _, tsk := range tasks {
		assertTaskExists(ctx, s, tsk, t)
	}
}

// Builds_CreateTasks_SequentialID verifies that when creating tasks
// with unset IDs, IDs are generated such that they are sequential in
// the build.
func Builds_CreateTasks_SequentialID(ctx context.Context, t *testing.T, s store.Builds) {
	build := sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "x/z"}, ID: 1}

	for i := 1; i < 4; i++ {
		tasks, err := s.CreateTasks(ctx, []*sourcegraph.BuildTask{{Build: build}})
		if err != nil {
			t.Fatal(err)
		}
		if want := uint64(i); tasks[0].ID != want {
			t.Errorf("got id == %d, want %d", tasks[0].ID, want)
		}
	}
}

// Builds_UpdateTask verifies the correct functioning of the Builds.UpdateTask method.
func Builds_UpdateTask(ctx context.Context, t *testing.T, s store.Builds, insert InsertTasksFunc) {
	tasks := []*sourcegraph.BuildTask{
		{ID: 1, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 1}, Label: "a"},
		{ID: 2, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 1}, Label: "a"},
		{ID: 3, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 2}, Label: "b"},
		{ID: 4, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "x/z"}, ID: 1}, Label: "b"},
	}
	insert(ctx, t, tasks)
	t0 := pbtypes.NewTimestamp(time.Unix(1, 0))
	err := s.UpdateTask(ctx, tasks[2].Spec(), sourcegraph.TaskUpdate{
		EndedAt: &t0,
		Failure: true,
	})
	if err != nil {
		t.Fatalf("errored out: %s", err)
	}
	want := *(tasks[2])
	want.EndedAt = &t0
	want.Failure = true
	assertTaskExists(ctx, s, &want, t)
}

// Builds_ListBuildTasks verifies the correct functioning of the Builds.ListBuildTasks method.
func Builds_ListBuildTasks(ctx context.Context, t *testing.T, s store.Builds, insert InsertTasksFunc) {
	tasks := []*sourcegraph.BuildTask{
		{ID: 10, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 1}, Label: "a"}, // test order
		{ID: 1, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 1}, Label: "b"},
		{ID: 2, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 1}, Label: "a"},
		{ID: 2, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 2}, Label: "a"},
	}
	insert(ctx, t, tasks)
	ts, err := s.ListBuildTasks(ctx, tasks[0].Spec().Build, nil)
	if err != nil {
		t.Fatalf("errored out: %s", err)
	}
	want := []*sourcegraph.BuildTask{tasks[1], tasks[2], tasks[0]}
	if !reflect.DeepEqual(ts, want) {
		t.Errorf("expected %#v, got %#v", want, ts)
	}
}

func Builds_GetTask(ctx context.Context, t *testing.T, s store.Builds, insert InsertTasksFunc) {
	tasks := []*sourcegraph.BuildTask{
		{ID: 1, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 1}, Label: "b"},
		{ID: 2, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 1}, Label: "a"},
		{ID: 2, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 2}, Label: "a"},
	}
	insert(ctx, t, tasks)
	for _, tsk := range tasks {
		assertTaskExists(ctx, s, tsk, t)
	}
}
