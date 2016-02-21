// +build pgsqltest

package pgsql

import (
	"testing"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store/testsuite"
)

func TestBuilds_Get(t *testing.T) {
	t.Parallel()

	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_Get(ctx, t, &s, s.mustCreateBuilds)
}

func TestBuilds_List(t *testing.T) {
	t.Parallel()

	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_List(ctx, t, &s, s.mustCreateBuilds)
}

func TestBuilds_List_byRepoAndCommitID(t *testing.T) {
	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_List_byRepoAndCommitID(ctx, t, &s, s.mustCreateBuilds)
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

func TestBuilds_GetFirstInCommitOrder_returnNewest(t *testing.T) {
	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_GetFirstInCommitOrder_returnNewest(ctx, t, &s, s.mustCreateBuilds)
}

func TestBuilds_ListBuildTasks(t *testing.T) {
	t.Parallel()

	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_ListBuildTasks(ctx, t, &s, s.mustCreateTasks)
}

func TestBuilds_Create(t *testing.T) {
	t.Parallel()

	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_Create(ctx, t, &s)
}

func TestBuilds_Create_New(t *testing.T) {
	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_Create_New(ctx, t, &s)
}

func TestBuilds_Create_SequentialID(t *testing.T) {
	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_Create_SequentialID(ctx, t, &s)
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
