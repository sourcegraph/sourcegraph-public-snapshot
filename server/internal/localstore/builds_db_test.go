// +build pgsqltest

package localstore

import (
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/util/jsonutil"
	"sourcegraph.com/sqs/pbtypes"
)

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

// TestBuilds_Get tests that the behavior of Builds.Get indirectly via the
// assertBuildExists method.
func TestBuilds_Get(t *testing.T) {
	t.Parallel()

	ctx, _, done := testContext()
	defer done()

	s := &builds{}
	want := &sourcegraph.Build{ID: 5, Repo: "x/x", CommitID: strings.Repeat("a", 40), Host: "localhost"}
	s.mustCreateBuilds(ctx, t, []*sourcegraph.Build{want})
	assertBuildExists(ctx, s, want, t)
}

// TestBuilds_List verifies the correct functioning of the Builds.List method.
func TestBuilds_List(t *testing.T) {
	t.Parallel()

	ctx, _, done := testContext()
	defer done()

	s := &builds{}
	want := []*sourcegraph.Build{
		{ID: 1, Repo: "r", CommitID: "c1"},
		{ID: 2, Repo: "r", CommitID: "c2"},
		{ID: 3, Repo: "r", CommitID: "c3"},
	}
	s.mustCreateBuilds(ctx, t, want)
	builds, err := s.List(ctx, &sourcegraph.BuildListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(builds, want) {
		t.Errorf("got %v, want %v", builds, want)
	}
}

// TestBuilds_List_byRepoAndCommitID verifies the correct functioning of
// the Builds.List method when filtering by a repo and commit ID.
func TestBuilds_List_byRepoAndCommitID(t *testing.T) {
	ctx, _, done := testContext()
	defer done()

	s := &builds{}
	data := []*sourcegraph.Build{
		{ID: 1, Repo: "r1", CommitID: "c1"},
		{ID: 2, Repo: "r1", CommitID: "c2"},
		{ID: 3, Repo: "r2", CommitID: "c1"},
	}
	s.mustCreateBuilds(ctx, t, data)
	builds, err := s.List(ctx, &sourcegraph.BuildListOptions{Repo: "r1", CommitID: "c1"})
	if err != nil {
		t.Fatal(err)
	}

	if want := []*sourcegraph.Build{data[0]}; !reflect.DeepEqual(builds, want) {
		t.Errorf("got %v, want %v", builds, want)
	}
}

// TestBuilds_ListBuildTasks verifies the correct functioning of the
// Builds.ListBuildTasks method.
func TestBuilds_ListBuildTasks(t *testing.T) {
	t.Parallel()

	ctx, _, done := testContext()
	defer done()

	s := builds{}
	tasks := []*sourcegraph.BuildTask{
		{ID: 10, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 1}, Label: "a"}, // test order
		{ID: 1, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 1}, Label: "b"},
		{ID: 2, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 1}, Label: "a"},
		{ID: 2, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 2}, Label: "a"},
	}
	s.mustCreateTasks(ctx, t, tasks)
	ts, err := s.ListBuildTasks(ctx, tasks[0].Spec().Build, nil)
	if err != nil {
		t.Fatalf("errored out: %s", err)
	}
	want := []*sourcegraph.BuildTask{tasks[1], tasks[2], tasks[0]}
	if !reflect.DeepEqual(ts, want) {
		t.Errorf("expected %#v, got %#v", want, ts)
	}
}

// TestBuilds_Create tests the behavior of Builds.Create and that it correctly
// creates the passed in build.
func TestBuilds_Create(t *testing.T) {
	t.Parallel()

	ctx, _, done := testContext()
	defer done()

	s := &builds{}
	want := &sourcegraph.Build{ID: 33, Repo: "y/y", CommitID: strings.Repeat("a", 40), Host: "localhost"}
	b, err := s.Create(ctx, want)
	if err != nil {
		t.Fatalf("errored out: %s", err)
	}
	if !reflect.DeepEqual(b, want) {
		t.Errorf("expected (on create): %#v, got %#v", want, b)
	}
	assertBuildExists(ctx, s, want, t)
}

// TestBuilds_Create_New verifies that passing a Build with ID == 0 to
// Builds.Create will generate an ID for it.
func TestBuilds_Create_New(t *testing.T) {
	ctx, _, done := testContext()
	defer done()

	s := &builds{}
	// No ID specified.
	want := &sourcegraph.Build{Repo: "y/y", CommitID: strings.Repeat("a", 40), Host: "localhost"}
	b, err := s.Create(ctx, want)
	if err != nil {
		t.Fatalf("errored out: %s", err)
	}
	if b.ID == 0 {
		t.Errorf("expected (on create new) id to be other than 0, but got %d", b.ID)
	}
	want.ID = b.ID
	assertBuildExists(ctx, s, want, t)
}

// TestBuilds_Create_SequentialID verifies that passing a Build with
// ID == 0 to Builds.Create will generate an ID for it that
// is greater than all other builds' IDs.
func TestBuilds_Create_SequentialID(t *testing.T) {
	ctx, _, done := testContext()
	defer done()

	s := &builds{}
	_, err := s.Create(ctx, &sourcegraph.Build{ID: 1, Repo: "y/y", CommitID: strings.Repeat("a", 40), Host: "localhost"})
	if err != nil {
		t.Fatal(err)
	}

	want := &sourcegraph.Build{Repo: "y/y", CommitID: strings.Repeat("a", 40), Host: "localhost"}
	b, err := s.Create(ctx, want)
	if err != nil {
		t.Fatal(err)
	}
	if want := uint64(2); b.ID != want {
		t.Errorf("got id == %d, want %d", b.ID, want)
	}
}

// TestBuilds_Update tests the correct functioning of the Builds.Update method
// by inserting a build, Updating it and verifying that it exists in its new
// form.
func TestBuilds_Update(t *testing.T) {
	t.Parallel()

	ctx, _, done := testContext()
	defer done()

	s := &builds{}
	orig := &sourcegraph.Build{ID: 33, Repo: "y/y", CommitID: strings.Repeat("a", 40), Host: "localhost"}
	t0 := pbtypes.NewTimestamp(time.Unix(1, 0))
	update := sourcegraph.BuildUpdate{
		StartedAt: &t0,
		Host:      "sourcegraph.com",
		Priority:  5,
		Killed:    true,
	}
	s.mustCreateBuilds(ctx, t, []*sourcegraph.Build{orig})

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

// TestBuilds_Update_builderConfig tests that updating BuilderConfig only updates
// the BuilderConfig without affecting other fields.
func TestBuilds_Update_builderConfig(t *testing.T) {
	t.Parallel()

	ctx, _, done := testContext()
	defer done()

	s := &builds{}
	t0 := pbtypes.NewTimestamp(time.Unix(1, 0))
	orig := &sourcegraph.Build{ID: 33, Repo: "y/y", CommitID: strings.Repeat("a", 40), Host: "localhost", StartedAt: &t0}
	s.mustCreateBuilds(ctx, t, []*sourcegraph.Build{orig})

	update := sourcegraph.BuildUpdate{BuilderConfig: "test"}
	err := s.Update(ctx, orig.Spec(), update)
	if err != nil {
		t.Fatalf("errored out: %s", err)
	}
	want := *orig
	want.BuilderConfig = update.BuilderConfig
	assertBuildExists(ctx, s, &want, t)
}

// TestBuilds_CreateTasks verifies that inserting a series of tasks via
// Builds.CreateTasks correctly creates these tasks in the store. The existence
// is asserted using the assertTaskExists method.
func TestBuilds_CreateTasks(t *testing.T) {
	t.Parallel()

	ctx, _, done := testContext()
	defer done()

	s := &builds{}
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

// TestBuilds_CreateTasks_SequentialID verifies that when creating tasks
// with unset IDs, IDs are generated such that they are sequential in
// the build.
func TestBuilds_CreateTasks_SequentialID(t *testing.T) {
	ctx, _, done := testContext()
	defer done()

	s := &builds{}
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

// TestBuilds_UpdateTask verifies the correct functioning of the
// Builds.UpdateTask method.
func TestBuilds_UpdateTask(t *testing.T) {
	t.Parallel()

	ctx, _, done := testContext()
	defer done()

	s := &builds{}
	tasks := []*sourcegraph.BuildTask{
		{ID: 1, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 1}, Label: "a"},
		{ID: 2, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 1}, Label: "a"},
		{ID: 3, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 2}, Label: "b"},
		{ID: 4, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "x/z"}, ID: 1}, Label: "b"},
	}
	s.mustCreateTasks(ctx, t, tasks)
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

func TestBuilds_GetTask(t *testing.T) {
	t.Parallel()

	ctx, _, done := testContext()
	defer done()

	s := &builds{}
	tasks := []*sourcegraph.BuildTask{
		{ID: 1, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 1}, Label: "b"},
		{ID: 2, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 1}, Label: "a"},
		{ID: 2, Build: sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "a/b"}, ID: 2}, Label: "a"},
	}
	s.mustCreateTasks(ctx, t, tasks)
	for _, tsk := range tasks {
		assertTaskExists(ctx, s, tsk, t)
	}
}

func TestBuilds_DequeueNext(t *testing.T) {
	t.Parallel()

	ctx, _, done := testContext()
	defer done()

	ctx = withTestIDKey(t, ctx)

	s := &builds{}
	want := &sourcegraph.Build{ID: 5, Repo: "x/x", CommitID: strings.Repeat("a", 40), Host: "localhost", BuildConfig: sourcegraph.BuildConfig{Queue: true}}
	s.mustCreateBuilds(ctx, t, []*sourcegraph.Build{want})
	build, err := s.DequeueNext(ctx)
	if err != nil {
		t.Fatalf("errored out: %s", err)
	}
	build.AccessToken = "" // do not compare access token
	if !reflect.DeepEqual(build, newBuildJobForTest(t, ctx, want)) {
		t.Errorf("expected %#v, got %#v", newBuildJobForTest(t, ctx, want), build)
	}
}

func TestBuilds_DequeueNext_ordered(t *testing.T) {
	t.Parallel()

	ctx, _, done := testContext()
	defer done()

	ctx = withTestIDKey(t, ctx)

	s := &builds{}
	t1 := pbtypes.NewTimestamp(time.Unix(100000, 0))
	t2 := pbtypes.NewTimestamp(time.Unix(200000, 0))

	b1 := &sourcegraph.Build{ID: 1, CommitID: strings.Repeat("A", 40), Repo: "r", CreatedAt: t1, BuildConfig: sourcegraph.BuildConfig{Queue: true, Priority: 10}}
	b2 := &sourcegraph.Build{ID: 2, CommitID: strings.Repeat("A", 40), Repo: "r", CreatedAt: t1, BuildConfig: sourcegraph.BuildConfig{Queue: true}}
	b3 := &sourcegraph.Build{ID: 3, CommitID: strings.Repeat("A", 40), Repo: "r", CreatedAt: t2, BuildConfig: sourcegraph.BuildConfig{Queue: true}}
	bNo1 := &sourcegraph.Build{ID: 4, CommitID: strings.Repeat("A", 40), Repo: "r", BuildConfig: sourcegraph.BuildConfig{Queue: false}}
	bNo2 := &sourcegraph.Build{ID: 5, CommitID: strings.Repeat("A", 40), Repo: "r", StartedAt: &t2, BuildConfig: sourcegraph.BuildConfig{Queue: true}}

	s.mustCreateBuilds(ctx, t, []*sourcegraph.Build{b1, b2, b3, bNo1, bNo2})

	wantBuilds := []*sourcegraph.BuildJob{
		newBuildJobForTest(t, ctx, b1), newBuildJobForTest(t, ctx, b2), newBuildJobForTest(t, ctx, b3), nil, // in order
	}

	for i, wantBuild := range wantBuilds {
		build, err := s.DequeueNext(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if build != nil {
			build.AccessToken = "" // do not compare access token
		}
		if !jsonutil.JSONEqual(t, build, wantBuild) {
			t.Errorf("dequeued build #%d\n\nGOT\n%+v\n\nWANT\n%+v", i+1, build, wantBuild)
		}
	}
}

func newBuildJobForTest(t *testing.T, ctx context.Context, b *sourcegraph.Build) *sourcegraph.BuildJob {
	j, err := newBuildJob(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	j.AccessToken = "" // do not compare access token
	return j
}

// TestBuilds_DequeueNext_noRaceCondition ensures that DequeueNext will dequeue
// a build exactly once and that concurrent processes will not dequeue the same
// build. It may not always trigger the race condition, but if it even does
// once, it is very important that we fix it.
func TestBuilds_DequeueNext_noRaceCondition(t *testing.T) {
	t.Parallel()

	ctx, _, done := testContext()
	defer done()

	ctx = withTestIDKey(t, ctx)

	s := &builds{}
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

	s.mustCreateBuilds(ctx, t, allBuilds)
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
				if dq[b.Spec.ID] {
					dqMu.Unlock()
					t.Errorf("build %d was already dequeued (race condition)", b.Spec.ID)
					return
				}
				dq[b.Spec.ID] = true
				dqMu.Unlock()
				t.Logf("worker %d got build %d", i, b.Spec.ID)
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

func withTestIDKey(t *testing.T, ctx context.Context) context.Context {
	idkey.SetTestEnvironment(512)
	k, err := idkey.Generate()
	if err != nil {
		t.Fatal(err)
	}
	return idkey.NewContext(ctx, k)
}
