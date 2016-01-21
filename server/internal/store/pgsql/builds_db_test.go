// +build pgsqltest

package pgsql

import (
	"testing"

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

func TestBuilds_GetFirstInCommitOrder_firstCommitIDMatch(t *testing.T) {
	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_GetFirstInCommitOrder_firstCommitIDMatch(ctx, t, &s, s.mustCreateBuilds)
}

func TestBuilds_GetFirstInCommitOrder_secondCommitIDMatch(t *testing.T) {
	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_GetFirstInCommitOrder_secondCommitIDMatch(ctx, t, &s, s.mustCreateBuilds)
}

func TestBuilds_GetFirstInCommitOrder_successfulOnly(t *testing.T) {
	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_GetFirstInCommitOrder_successfulOnly(ctx, t, &s, s.mustCreateBuilds)
}

func TestBuilds_GetFirstInCommitOrder_noneFound(t *testing.T) {
	var s builds
	ctx, done := testContext()
	defer done()

	testsuite.Builds_GetFirstInCommitOrder_noneFound(ctx, t, &s, s.mustCreateBuilds)
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
