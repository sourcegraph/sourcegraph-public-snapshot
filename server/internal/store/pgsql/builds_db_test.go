// +build pgsqltest

package pgsql

import (
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/store/testsuite"
	"src.sourcegraph.com/sourcegraph/util/jsonutil"
)

func TestBuilds_Get(t *testing.T) {
	t.Parallel()

	var s Builds
	ctx, done := testContext()
	defer done()

	wantBuild := s.mustCreate(ctx, t, &sourcegraph.Build{Repo: "r"})

	build, err := s.Get(ctx, sourcegraph.BuildSpec{Attempt: 1, Repo: sourcegraph.RepoSpec{URI: "r"}})
	if err != nil {
		t.Fatal(err)
	}

	if !jsonutil.JSONEqual(t, build, wantBuild) {
		t.Errorf("got build %+v, want %+v", build, wantBuild)
	}
}

func TestBuilds_List(t *testing.T) {
	t.Parallel()

	wantBuild := &sourcegraph.Build{Repo: "r"}
	wantBuilds := []*sourcegraph.Build{wantBuild}

	var s Builds
	ctx, done := testContext()
	defer done()

	wantBuild = s.mustCreate(ctx, t, wantBuild)

	builds, err := s.List(ctx, &sourcegraph.BuildListOptions{
		Sort:        "priority",
		Direction:   "desc",
		ListOptions: sourcegraph.ListOptions{Page: 1, PerPage: 10},
	})
	if err != nil {
		t.Fatal(err)
	}

	if n := len(builds); n != len(wantBuilds) {
		t.Errorf("got len(builds) == %d, want %d", n, len(wantBuilds))
	}
	if !jsonutil.JSONEqual(t, builds[0], wantBuild) {
		t.Errorf("got build %+v, want %+v", builds[0], wantBuild)
	}
}

func TestBuilds_GetFirstInCommitOrder_firstCommitIDMatch(t *testing.T) {
	var s Builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_GetFirstInCommitOrder_firstCommitIDMatch(ctx, t, &s, s.mustCreateBuilds)
}

func TestBuilds_GetFirstInCommitOrder_secondCommitIDMatch(t *testing.T) {
	var s Builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_GetFirstInCommitOrder_secondCommitIDMatch(ctx, t, &s, s.mustCreateBuilds)
}

func TestBuilds_GetFirstInCommitOrder_successfulOnly(t *testing.T) {
	var s Builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_GetFirstInCommitOrder_successfulOnly(ctx, t, &s, s.mustCreateBuilds)
}

func TestBuilds_GetFirstInCommitOrder_noneFound(t *testing.T) {
	var s Builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_GetFirstInCommitOrder_noneFound(ctx, t, &s, s.mustCreateBuilds)
}

func TestBuilds_GetFirstInCommitOrder_returnNewest(t *testing.T) {
	var s Builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_GetFirstInCommitOrder_returnNewest(ctx, t, &s, s.mustCreateBuilds)
}

func TestBuilds_ListBuildTasks(t *testing.T) {
	t.Parallel()

	var s Builds
	ctx, done := testContext()
	defer done()

	newTasks := []*sourcegraph.BuildTask{{Repo: "r", Op: "a"}}
	wantTasks := s.mustCreateTasks(ctx, t, newTasks)

	tasks, err := s.ListBuildTasks(ctx, sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "r"}}, &sourcegraph.BuildTaskListOptions{ListOptions: sourcegraph.ListOptions{Page: 1, PerPage: 10}})
	if err != nil {
		t.Fatal(err)
	}

	if n := len(tasks); n != len(wantTasks) {
		t.Errorf("got len(tasks) == %d, want %d", n, len(wantTasks))
	}
	if !reflect.DeepEqual(tasks, wantTasks) {
		t.Errorf("got build tasks %+v, want %+v", tasks, wantTasks)
	}
}

func TestBuilds_Create(t *testing.T) {
	t.Parallel()

	var s Builds
	ctx, done := testContext()
	defer done()

	b, err := s.Create(ctx, &sourcegraph.Build{Repo: "r", CommitID: "c"})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := s.Get(ctx, b.Spec()); err != nil {
		t.Fatal(err)
	}
}

func TestBuilds_Create_New(t *testing.T) {
	var s Builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_Create_New(ctx, t, &s)
}

func TestBuilds_Create_SequentialAttempt(t *testing.T) {
	var s Builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_Create_SequentialAttempt(ctx, t, &s)
}

func TestBuilds_Update(t *testing.T) {
	t.Parallel()

	var s Builds
	ctx, done := testContext()
	defer done()

	b := s.mustCreate(ctx, t, &sourcegraph.Build{Repo: "r", CommitID: "c"})

	if err := s.Update(ctx, b.Spec(), sourcegraph.BuildUpdate{Success: true}); err != nil {
		t.Fatal(err)
	}

	updated, err := s.Get(ctx, b.Spec())
	if err != nil {
		t.Fatal(err)
	}
	if !updated.Success {
		t.Errorf("got updated build Success == %v, want %v", updated.Success, true)
	}
}

func TestBuilds_CreateTasks(t *testing.T) {
	t.Parallel()

	var s Builds
	ctx, done := testContext()
	defer done()

	buildSpec := sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "r"}}
	tasks := []*sourcegraph.BuildTask{
		{Repo: "r", Op: "foo", UnitType: "t", Unit: "u"},
		{Repo: "r", Op: "bar", UnitType: "t", Unit: "u"},
	}
	if _, err := s.CreateTasks(ctx, tasks); err != nil {
		t.Fatal(err)
	}

	tasks2, err := s.ListBuildTasks(ctx, buildSpec, nil)
	if err != nil {
		t.Fatal(err)
	}
	if want := 2; len(tasks2) != want {
		t.Errorf("got len(tasks2) == %d, want %d", len(tasks2), want)
	}
}

func TestBuilds_UpdateTask(t *testing.T) {
	t.Parallel()

	var s Builds
	ctx, done := testContext()
	defer done()

	task := &sourcegraph.BuildTask{Attempt: 1, CommitID: strings.Repeat("a", 40), Repo: "r", Op: "foo", UnitType: "t", Unit: "u"}

	tasks, err := s.CreateTasks(ctx, []*sourcegraph.BuildTask{task})
	if err != nil {
		t.Fatal(err)
	}
	task = tasks[0]

	ts := pbtypes.NewTimestamp(time.Unix(123, 0))
	if err := s.UpdateTask(ctx, task.Spec(), sourcegraph.TaskUpdate{Success: true, StartedAt: &ts}); err != nil {
		t.Fatal(err)
	}

	updated, err := s.GetTask(ctx, task.Spec())
	if err != nil {
		t.Fatal(err)
	}
	if !updated.Success {
		t.Errorf("got updated task Success == %v, want %v", updated.Success, true)
	}
}

func TestBuilds_DequeueNext(t *testing.T) {
	t.Parallel()

	var s Builds
	ctx, done := testContext()
	defer done()

	t1 := time.Unix(100000, 0)
	t2 := time.Unix(200000, 0)

	b1 := &sourcegraph.Build{Attempt: 1, CommitID: strings.Repeat("A", 40), Repo: "r", CreatedAt: pbtypes.NewTimestamp(t1), BuildConfig: sourcegraph.BuildConfig{Queue: true, Priority: 10}}
	b2 := &sourcegraph.Build{Attempt: 2, CommitID: strings.Repeat("A", 40), Repo: "r", CreatedAt: pbtypes.NewTimestamp(t1), BuildConfig: sourcegraph.BuildConfig{Queue: true}}
	b3 := &sourcegraph.Build{Attempt: 3, CommitID: strings.Repeat("A", 40), Repo: "r", CreatedAt: pbtypes.NewTimestamp(t2), BuildConfig: sourcegraph.BuildConfig{Queue: true}}
	bNo1 := &sourcegraph.Build{Attempt: 4, CommitID: strings.Repeat("A", 40), Repo: "r", BuildConfig: sourcegraph.BuildConfig{Queue: false}}
	bNo2 := &sourcegraph.Build{Attempt: 5, CommitID: strings.Repeat("A", 40), Repo: "r", StartedAt: ts(&t1), BuildConfig: sourcegraph.BuildConfig{Queue: true}}

	b1 = s.mustCreate(ctx, t, b1)
	b2 = s.mustCreate(ctx, t, b2)
	b3 = s.mustCreate(ctx, t, b3)
	bNo1 = s.mustCreate(ctx, t, bNo1)
	bNo2 = s.mustCreate(ctx, t, bNo2)

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

func TestBuilds_DequeueNext_noRaceCondition(t *testing.T) {
	// This test ensures that DequeueNext will dequeue a build exactly
	// once and that concurrent processes will not dequeue the same
	// build. It may not always trigger the race condition, but if it
	// even does once, it is very important that we fix it.

	t.Parallel()

	const (
		nbuilds  = 90
		nworkers = 30
	)

	var builds []*sourcegraph.Build
	for i := 0; i < nbuilds; i++ {
		builds = append(builds, &sourcegraph.Build{
			Repo: "r", BuildConfig: sourcegraph.BuildConfig{Queue: true, Priority: int32(i)},
			CommitID: strings.Repeat("a", 40),
		})
	}

	var s Builds
	ctx, done := testContext()
	defer done()

	for i, b := range builds {
		builds[i] = s.mustCreate(ctx, t, b)
	}
	t.Logf("enqueued %d builds", nbuilds)

	dq := map[uint32]bool{} // build attempt -> whether it has already been dequeued
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
				if dq[b.Attempt] {
					dqMu.Unlock()
					t.Errorf("build %d was already dequeued (race condition)", b.Attempt)
					return
				}
				dq[b.Attempt] = true
				dqMu.Unlock()
				t.Logf("worker %d got build %d (priority %d)", i, b.Attempt, b.Priority)
			}
		}(i)
	}
	wg.Wait()

	for _, b := range builds {
		if !dq[b.Attempt] {
			t.Errorf("build %d was never dequeued", b.Attempt)
		}
	}
}
