// +build pgsqltest

package pgsql

import (
	"strings"
	"sync"
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
	// This test ensures that DequeueNext will dequeue a build exactly
	// once and that concurrent processes will not dequeue the same
	// build. It may not always trigger the race condition, but if it
	// even does once, it is very important that we fix it.

	// TODO port this test to the testsuite

	t.Parallel()

	const (
		nbuilds  = 90
		nworkers = 30
	)

	var allBuilds []*sourcegraph.Build
	for i := 0; i < nbuilds; i++ {
		allBuilds = append(allBuilds, &sourcegraph.Build{
			Repo: "r", BuildConfig: sourcegraph.BuildConfig{Queue: true, Priority: int32(i)},
			CommitID: strings.Repeat("a", 40),
		})
	}

	var s builds
	ctx, done := testContext()
	defer done()

	for i, b := range allBuilds {
		allBuilds[i] = s.mustCreate(ctx, t, b)
	}
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
