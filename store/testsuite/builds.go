package testsuite

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/store"
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

// Builds_GetFirstInCommitOrder_firstCommitIDMatch tests the behavior
// of Builds.GetFirstInCommitOrder when the first commit ID has
// multiple builds (it should return the newest).
func Builds_GetFirstInCommitOrder_firstCommitIDMatch(ctx context.Context, t *testing.T, s store.Builds, insert InsertBuildsFunc) {
	insert(ctx, t, []*sourcegraph.Build{{Attempt: 1, Repo: "r", CommitID: "a"}})

	build, nth, err := s.GetFirstInCommitOrder(ctx, "r", []string{"a"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if build == nil {
		t.Fatal("build == nil")
	}
	if build.Attempt != 1 {
		t.Errorf("got Attempt %d, want %d", build.Attempt, 1)
	}
	if want := 0; nth != want {
		t.Errorf("got nth == %d, want %d", nth, want)
	}
}

// Builds_GetFirstInCommitOrder_secondCommitIDMatch tests the behavior
// of Builds.GetSecondInCommitOrder when the *second* (but not second)
// commit ID has multiple builds (it should return the newest).
func Builds_GetFirstInCommitOrder_secondCommitIDMatch(ctx context.Context, t *testing.T, s store.Builds, insert InsertBuildsFunc) {
	insert(ctx, t, []*sourcegraph.Build{{Attempt: 2, Repo: "r", CommitID: "b"}})

	build, nth, err := s.GetFirstInCommitOrder(ctx, "r", []string{"a", "b"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if build == nil {
		t.Fatal("build == nil")
	}
	if build.Attempt != 2 {
		t.Errorf("got Attempt %d, want %d", build.Attempt, 2)
	}
	if want := 1; nth != want {
		t.Errorf("got nth == %d, want %d", nth, want)
	}
}

// Builds_GetFirstInCommitOrder_successfulOnly tests the behavior of
// Builds.GetFirstInCommitOrder when successfulOnly is true and there
// are no successful builds.
func Builds_GetFirstInCommitOrder_successfulOnly(ctx context.Context, t *testing.T, s store.Builds, insert InsertBuildsFunc) {
	insert(ctx, t, []*sourcegraph.Build{{Attempt: 1, Repo: "r", CommitID: "a", Success: false}})

	build, nth, err := s.GetFirstInCommitOrder(ctx, "r", []string{"a"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if build != nil {
		t.Error("build != nil")
	}
	if want := -1; nth != want {
		t.Errorf("got nth == %d, want %d", nth, want)
	}
}

// Builds_GetFirstInCommitOrder_noneFound tests the behavior of
// Builds.GetFirstInCommitOrder when there are no builds with any of
// the specified commitIDs.
func Builds_GetFirstInCommitOrder_noneFound(ctx context.Context, t *testing.T, s store.Builds, insert InsertBuildsFunc) {
	insert(ctx, t, []*sourcegraph.Build{{Attempt: 1, Repo: "r", CommitID: "a"}})

	build, nth, err := s.GetFirstInCommitOrder(ctx, "r", []string{"b"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if build != nil {
		t.Error("build != nil")
	}
	if want := -1; nth != want {
		t.Errorf("got nth == %d, want %d", nth, want)
	}
}

// Builds_GetFirstInCommitOrder_returnNewest tests the behavior of
// Builds.GetFirstInCommitOrder when there are multiple builds for a
// specified commit ID (it should pick the newest build).
func Builds_GetFirstInCommitOrder_returnNewest(ctx context.Context, t *testing.T, s store.Builds, insert InsertBuildsFunc) {
	t0 := pbtypes.NewTimestamp(time.Unix(0, 0)) // oldest
	t1 := pbtypes.NewTimestamp(time.Unix(1, 0))
	t2 := pbtypes.NewTimestamp(time.Unix(2, 0)) // newest
	insert(ctx, t, []*sourcegraph.Build{
		{Attempt: 1, Repo: "r", CommitID: "a", StartedAt: &t0},
		{Attempt: 2, Repo: "r", CommitID: "a", StartedAt: &t2}, // newest
		{Attempt: 3, Repo: "r", CommitID: "a", StartedAt: &t1},
	})

	build, nth, err := s.GetFirstInCommitOrder(ctx, "r", []string{"a"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if build == nil {
		t.Fatal("build == nil")
	}
	if build.Attempt != 2 {
		t.Errorf("got Attempt %d, want %d", build.Attempt, 2)
	}
	if want := 0; nth != want {
		t.Errorf("got nth == %d, want %d", nth, want)
	}
}

// Builds_Get tests that the behavior of Builds.Get indirectly via the assertBuildExists method.
func Builds_Get(ctx context.Context, t *testing.T, s store.Builds, insert InsertBuildsFunc) {
	want := &sourcegraph.Build{Attempt: 5, Repo: "x/x", CommitID: strings.Repeat("a", 40), Host: "localhost"}
	insert(ctx, t, []*sourcegraph.Build{want})
	assertBuildExists(ctx, s, want, t)
}

// Builds_Create tests the behavior of Builds.Create and that it correctly creates the passed
// in build.
func Builds_Create(ctx context.Context, t *testing.T, s store.Builds) {
	want := &sourcegraph.Build{Attempt: 33, Repo: "y/y", CommitID: strings.Repeat("a", 40), Host: "localhost"}
	b, err := s.Create(ctx, want)
	if err != nil {
		t.Fatalf("errored out: %s", err)
	}
	if !reflect.DeepEqual(b, want) {
		t.Errorf("expected (on create): %#v, got %#v", want, b)
	}
	assertBuildExists(ctx, s, want, t)
}

// Builds_Create_Queue verifies that passing a Build with StartedAt=nil to the Builds.Create method
// will make it available in the queue.
func Builds_Create_Queue(ctx context.Context, t *testing.T, s store.Builds, queueEntryExists ValidateQueueEntryFunc) {
	want := &sourcegraph.Build{Attempt: 33, Repo: "y/y", CommitID: strings.Repeat("a", 40), Host: "localhost", BuildConfig: sourcegraph.BuildConfig{Queue: true}}
	_, err := s.Create(ctx, want)
	if err != nil {
		t.Fatalf("errored out: %s", err)
	}
	if !queueEntryExists(ctx, want.Spec(), t) {
		t.Errorf("%#v not in queue", want.Spec())
	}
}

// Builds_Create_New verifies that passing a Build with Attempt == 0 to Builds.Create will
// generate an Attempt for it.
func Builds_Create_New(ctx context.Context, t *testing.T, s store.Builds) {
	// no attempt
	want := &sourcegraph.Build{Repo: "y/y", CommitID: strings.Repeat("a", 40), Host: "localhost"}
	b, err := s.Create(ctx, want)
	if err != nil {
		t.Fatalf("errored out: %s", err)
	}
	if b.Attempt == 0 {
		t.Errorf("expected (on create new) attempt to be other than 0, but got %d", b.Attempt)
	}
	want.Attempt = b.Attempt
	assertBuildExists(ctx, s, want, t)
}

// Builds_Create_SequentialAttempt verifies that passing a Build with
// Attempt == 0 to Builds.Create will generate an Attempt for it that
// is greater than all other builds' Attempts.
func Builds_Create_SequentialAttempt(ctx context.Context, t *testing.T, s store.Builds) {
	_, err := s.Create(ctx, &sourcegraph.Build{Attempt: 1, Repo: "y/y", CommitID: strings.Repeat("a", 40), Host: "localhost"})
	if err != nil {
		t.Fatal(err)
	}

	want := &sourcegraph.Build{Repo: "y/y", CommitID: strings.Repeat("a", 40), Host: "localhost"}
	b, err := s.Create(ctx, want)
	if err != nil {
		t.Fatal(err)
	}
	if want := uint32(2); b.Attempt != want {
		t.Errorf("got attempt == %d, want %d", b.Attempt, want)
	}
}

// Builds_Update tests the correct functioning of the Builds.Update method by inserting a build,
// Updating it and verifying that it exists in its new form.
func Builds_Update(ctx context.Context, t *testing.T, s store.Builds, insert InsertBuildsFunc) {
	orig := &sourcegraph.Build{Attempt: 33, Repo: "y/y", CommitID: strings.Repeat("a", 40), Host: "localhost"}
	t0 := pbtypes.NewTimestamp(time.Unix(1, 0))
	update := sourcegraph.BuildUpdate{
		StartedAt: &t0,
		Host:      "sourcegraph.com",
		Priority:  5,
		Killed:    true,
	}
	insert(ctx, t, []*sourcegraph.Build{orig})

	err := s.Update(ctx, orig.Spec(), update)
	if err != nil {
		t.Fatalf("errored out: %s", err)
	}
	want := *orig
	want.StartedAt = update.StartedAt
	want.Host = update.Host
	want.Priority = update.Priority
	want.Killed = update.Killed
	assertBuildExists(ctx, s, &want, t)
}

func Builds_DequeueNext(ctx context.Context, t *testing.T, s store.Builds, insert InsertBuildsFunc) {
	want := &sourcegraph.Build{Attempt: 5, Repo: "x/x", CommitID: strings.Repeat("a", 40), Host: "localhost", BuildConfig: sourcegraph.BuildConfig{Queue: true}}
	insert(ctx, t, []*sourcegraph.Build{want})
	build, err := s.DequeueNext(ctx)
	if err != nil {
		t.Fatalf("errored out: %s", err)
	}
	if !reflect.DeepEqual(build, want) {
		t.Errorf("expected %#v, got %#v", want, build)
	}
}

// Builds_CreateTasks verifies that inserting a series of tasks via Builds.CreateTasks correctly
// creates these tasks in the store. The existence is asserted using the assertTaskExists method.
func Builds_CreateTasks(ctx context.Context, t *testing.T, s store.Builds, _ InsertTasksFunc) {
	tasks := []*sourcegraph.BuildTask{
		{TaskID: 1, Repo: "a/b", CommitID: strings.Repeat("a", 40), Attempt: 1, Op: "import"},
		{TaskID: 2, Repo: "a/b", CommitID: strings.Repeat("b", 40), Attempt: 1, Op: "import"},
		{TaskID: 3, Repo: "a/b", CommitID: strings.Repeat("b", 40), Attempt: 2, Op: "graph"},
		{TaskID: 4, Repo: "x/z", CommitID: strings.Repeat("v", 40), Attempt: 1, Op: "graph"},
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

// Builds_UpdateTask verifies the correct functioning of the Builds.UpdateTask method.
func Builds_UpdateTask(ctx context.Context, t *testing.T, s store.Builds, insert InsertTasksFunc) {
	tasks := []*sourcegraph.BuildTask{
		{TaskID: 1, Repo: "a/b", CommitID: strings.Repeat("a", 40), Attempt: 1, Op: "import"},
		{TaskID: 2, Repo: "a/b", CommitID: strings.Repeat("b", 40), Attempt: 1, Op: "import"},
		{TaskID: 3, Repo: "a/b", CommitID: strings.Repeat("b", 40), Attempt: 2, Op: "graph"},
		{TaskID: 4, Repo: "x/z", CommitID: strings.Repeat("v", 40), Attempt: 1, Op: "graph"},
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
		{TaskID: 10, Repo: "a/b", CommitID: strings.Repeat("a", 40), Attempt: 1, Op: "graph"}, // test order
		{TaskID: 1, Repo: "a/b", CommitID: strings.Repeat("a", 40), Attempt: 1, Op: "import"},
		{TaskID: 2, Repo: "a/b", CommitID: strings.Repeat("a", 40), Attempt: 1, Op: "graph"},
		{TaskID: 2, Repo: "a/b", CommitID: strings.Repeat("a", 40), Attempt: 2, Op: "graph"},
	}
	insert(ctx, t, tasks)
	ts, err := s.ListBuildTasks(ctx, tasks[0].Spec().BuildSpec, nil)
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
		{TaskID: 1, Repo: "a/b", CommitID: strings.Repeat("a", 40), Attempt: 1, Op: "import"},
		{TaskID: 2, Repo: "a/b", CommitID: strings.Repeat("a", 40), Attempt: 1, Op: "graph"},
		{TaskID: 2, Repo: "a/b", CommitID: strings.Repeat("a", 40), Attempt: 2, Op: "graph"},
	}
	insert(ctx, t, tasks)
	for _, tsk := range tasks {
		assertTaskExists(ctx, s, tsk, t)
	}
}
