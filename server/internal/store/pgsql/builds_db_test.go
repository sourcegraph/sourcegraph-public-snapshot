// +build pgsqltest

package pgsql

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/context"

	"sourcegraph.com/sqs/pbtypes"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/store/testsuite"
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

// TestBuilds_Get tests that the behavior of Builds.Get indirectly via the
// assertBuildExists method.
func TestBuilds_Get(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	s := &builds{}
	want := &sourcegraph.Build{ID: 5, Repo: "x/x", CommitID: strings.Repeat("a", 40), Host: "localhost"}
	s.mustCreateBuilds(ctx, t, []*sourcegraph.Build{want})
	assertBuildExists(ctx, s, want, t)
}

// TestBuilds_List verifies the correct functioning of the Builds.List method.
func TestBuilds_List(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
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
	ctx, done := testContext()
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

// TestBuilds_GetFirstInCommitOrder_firstCommitIDMatch tests the behavior
// of Builds.GetFirstInCommitOrder when the first commit ID has
// multiple builds (it should return the newest).
func TestBuilds_GetFirstInCommitOrder_firstCommitIDMatch(t *testing.T) {
	ctx, done := testContext()
	defer done()

	s := &builds{}
	s.mustCreateBuilds(ctx, t, []*sourcegraph.Build{{ID: 1, Repo: "r", CommitID: "a"}})

	build, nth, err := s.GetFirstInCommitOrder(ctx, "r", []string{"a"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if build == nil {
		t.Fatal("build == nil")
	}
	if build.ID != 1 {
		t.Errorf("got ID %d, want %d", build.ID, 1)
	}
	if want := 0; nth != want {
		t.Errorf("got nth == %d, want %d", nth, want)
	}
}

// TestBuilds_GetFirstInCommitOrder_secondCommitIDMatch tests the behavior
// of Builds.GetSecondInCommitOrder when the *second* (but not second)
// commit ID has multiple builds (it should return the newest).
func TestBuilds_GetFirstInCommitOrder_secondCommitIDMatch(t *testing.T) {
	ctx, done := testContext()
	defer done()

	s := &builds{}
	s.mustCreateBuilds(ctx, t, []*sourcegraph.Build{{ID: 2, Repo: "r", CommitID: "b"}})

	build, nth, err := s.GetFirstInCommitOrder(ctx, "r", []string{"a", "b"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if build == nil {
		t.Fatal("build == nil")
	}
	if build.ID != 2 {
		t.Errorf("got ID %d, want %d", build.ID, 2)
	}
	if want := 1; nth != want {
		t.Errorf("got nth == %d, want %d", nth, want)
	}
}

// TestBuilds_GetFirstInCommitOrder_successfulOnly tests the behavior of
// Builds.GetFirstInCommitOrder when successfulOnly is true and there
// are no successful builds.
func TestBuilds_GetFirstInCommitOrder_successfulOnly(t *testing.T) {
	ctx, done := testContext()
	defer done()

	s := &builds{}
	s.mustCreateBuilds(ctx, t, []*sourcegraph.Build{{ID: 1, Repo: "r", CommitID: "a", Success: false}})

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

// TestBuilds_GetFirstInCommitOrder_noneFound tests the behavior of
// Builds.GetFirstInCommitOrder when there are no builds with any of
// the specified commitIDs.
func TestBuilds_GetFirstInCommitOrder_noneFound(t *testing.T) {
	ctx, done := testContext()
	defer done()

	s := &builds{}
	s.mustCreateBuilds(ctx, t, []*sourcegraph.Build{{ID: 1, Repo: "r", CommitID: "a"}})

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

// TestBuilds_GetFirstInCommitOrder_returnNewest tests the behavior of
// Builds.GetFirstInCommitOrder when there are multiple builds for a
// specified commit ID (it should pick the newest build).
func TestBuilds_GetFirstInCommitOrder_returnNewest(t *testing.T) {
	ctx, done := testContext()
	defer done()

	s := &builds{}
	t0 := pbtypes.NewTimestamp(time.Unix(0, 0)) // oldest
	t1 := pbtypes.NewTimestamp(time.Unix(1, 0))
	t2 := pbtypes.NewTimestamp(time.Unix(2, 0)) // newest
	s.mustCreateBuilds(ctx, t, []*sourcegraph.Build{
		{ID: 1, Repo: "r", CommitID: "a", StartedAt: &t0},
		{ID: 2, Repo: "r", CommitID: "a", StartedAt: &t2}, // newest
		{ID: 3, Repo: "r", CommitID: "a", StartedAt: &t1},
	})

	build, nth, err := s.GetFirstInCommitOrder(ctx, "r", []string{"a"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if build == nil {
		t.Fatal("build == nil")
	}
	if build.ID != 2 {
		t.Errorf("got ID %d, want %d", build.ID, 2)
	}
	if want := 0; nth != want {
		t.Errorf("got nth == %d, want %d", nth, want)
	}
}

func TestBuilds_ListBuildTasks(t *testing.T) {
	t.Parallel()

	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_ListBuildTasks(ctx, t, &s, s.mustCreateTasks)
}

// TestBuilds_Create tests the behavior of Builds.Create and that it correctly
// creates the passed in build.
func TestBuilds_Create(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
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
	ctx, done := testContext()
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
	ctx, done := testContext()
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

func TestBuilds_Update(t *testing.T) {
	t.Parallel()

	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_Update(ctx, t, &s, s.mustCreateBuilds)
}

func TestBuilds_Update_builderConfig(t *testing.T) {
	t.Parallel()

	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_Update_builderConfig(ctx, t, &s, s.mustCreateBuilds)
}

func TestBuilds_CreateTasks(t *testing.T) {
	t.Parallel()

	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_CreateTasks(ctx, t, &s, s.mustCreateTasks)
}

func TestBuilds_CreateTasks_SequentialID(t *testing.T) {
	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_CreateTasks_SequentialID(ctx, t, &s)
}

func TestBuilds_UpdateTask(t *testing.T) {
	t.Parallel()

	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_UpdateTask(ctx, t, &s, s.mustCreateTasks)
}

func TestBuilds_GetTask(t *testing.T) {
	t.Parallel()

	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_GetTask(ctx, t, &s, s.mustCreateTasks)
}

func TestBuilds_DequeueNext(t *testing.T) {
	t.Parallel()

	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_DequeueNext(ctx, t, &s, s.mustCreateBuilds)
}

func TestBuilds_DequeueNext_ordered(t *testing.T) {
	t.Parallel()

	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_DequeueNext_ordered(ctx, t, &s, s.mustCreateBuilds)
}

func TestBuilds_DequeueNext_noRaceCondition(t *testing.T) {
	t.Parallel()

	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_DequeueNext_noRaceCondition(ctx, t, &s, s.mustCreateBuilds)
}
