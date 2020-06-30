package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestGetDumpByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	// Dump does not exist initially
	if _, exists, err := store.GetDumpByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting dump: %s", err)
	} else if exists {
		t.Fatal("unexpected record")
	}

	uploadedAt := time.Unix(1587396557, 0).UTC()
	startedAt := uploadedAt.Add(time.Minute)
	finishedAt := uploadedAt.Add(time.Minute * 2)
	expected := Dump{
		ID:             1,
		Commit:         makeCommit(1),
		Root:           "sub/",
		VisibleAtTip:   true,
		UploadedAt:     uploadedAt,
		State:          "completed",
		FailureMessage: nil,
		StartedAt:      &startedAt,
		FinishedAt:     &finishedAt,
		RepositoryID:   50,
		RepositoryName: "n-50",
		Indexer:        "lsif-go",
	}

	insertUploads(t, dbconn.Global, Upload{
		ID:             expected.ID,
		Commit:         expected.Commit,
		Root:           expected.Root,
		VisibleAtTip:   expected.VisibleAtTip,
		UploadedAt:     expected.UploadedAt,
		State:          expected.State,
		FailureMessage: expected.FailureMessage,
		StartedAt:      expected.StartedAt,
		FinishedAt:     expected.FinishedAt,
		ProcessAfter:   expected.ProcessAfter,
		NumResets:      expected.NumResets,
		RepositoryID:   expected.RepositoryID,
		RepositoryName: expected.RepositoryName,
		Indexer:        expected.Indexer,
	})

	if dump, exists, err := store.GetDumpByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting dump: %s", err)
	} else if !exists {
		t.Fatal("expected record to exist")
	} else if diff := cmp.Diff(expected, dump); diff != "" {
		t.Errorf("unexpected dump (-want +got):\n%s", diff)
	}
}

func TestFindClosestDumps(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	// This database has the following commit graph:
	//
	// [1] --+--- 2 --------+--5 -- 6 --+-- [7]
	//       |              |           |
	//       +-- [3] -- 4 --+           +--- 8

	if err := store.UpdateCommits(context.Background(), 50, map[string][]string{
		makeCommit(1): {},
		makeCommit(2): {makeCommit(1)},
		makeCommit(3): {makeCommit(1)},
		makeCommit(4): {makeCommit(3)},
		makeCommit(5): {makeCommit(2), makeCommit(4)},
		makeCommit(6): {makeCommit(5)},
		makeCommit(7): {makeCommit(6)},
		makeCommit(8): {makeCommit(6)},
	}); err != nil {
		t.Fatalf("unexpected error updating commits: %s", err)
	}

	insertUploads(t, dbconn.Global,
		Upload{ID: 1, Commit: makeCommit(1)},
		Upload{ID: 2, Commit: makeCommit(3)},
		Upload{ID: 3, Commit: makeCommit(7)},
	)

	testFindClosestDumps(t, store, []FindClosestDumpsTestCase{
		{commit: makeCommit(1), file: "file.ts", rootMustEnclosePath: true, anyOfIDs: []int{1}},
		{commit: makeCommit(2), file: "file.ts", rootMustEnclosePath: true, anyOfIDs: []int{1}},
		{commit: makeCommit(3), file: "file.ts", rootMustEnclosePath: true, anyOfIDs: []int{2}},
		{commit: makeCommit(4), file: "file.ts", rootMustEnclosePath: true, anyOfIDs: []int{2}},
		{commit: makeCommit(6), file: "file.ts", rootMustEnclosePath: true, anyOfIDs: []int{3}},
		{commit: makeCommit(7), file: "file.ts", rootMustEnclosePath: true, anyOfIDs: []int{3}},
		{commit: makeCommit(5), file: "file.ts", rootMustEnclosePath: true, anyOfIDs: []int{1, 2, 3}},
		{commit: makeCommit(8), file: "file.ts", rootMustEnclosePath: true, anyOfIDs: []int{1, 2}},
	})
}

func TestFindClosestDumpsAlternateCommitGraph(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	// This database has the following commit graph:
	//
	// 1 --+-- [2] ---- 3
	//     |
	//     +--- 4 --+-- 5 -- 6
	//              |
	//              +-- 7 -- 8

	if err := store.UpdateCommits(context.Background(), 50, map[string][]string{
		makeCommit(1): {},
		makeCommit(2): {makeCommit(1)},
		makeCommit(3): {makeCommit(2)},
		makeCommit(4): {makeCommit(1)},
		makeCommit(5): {makeCommit(4)},
		makeCommit(6): {makeCommit(5)},
		makeCommit(7): {makeCommit(4)},
		makeCommit(8): {makeCommit(7)},
	}); err != nil {
		t.Fatalf("unexpected error updating commits: %s", err)
	}

	insertUploads(t, dbconn.Global,
		Upload{ID: 1, Commit: makeCommit(2)},
	)

	testFindClosestDumps(t, store, []FindClosestDumpsTestCase{
		{commit: makeCommit(1), allOfIDs: []int{1}},
		{commit: makeCommit(2), allOfIDs: []int{1}},
		{commit: makeCommit(3), allOfIDs: []int{1}},
		{commit: makeCommit(4)},
		{commit: makeCommit(6)},
		{commit: makeCommit(7)},
		{commit: makeCommit(5)},
		{commit: makeCommit(8)},
	})
}

func TestFindClosestDumpsDistinctRoots(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	// This database has the following commit graph:
	//
	// 1 --+-- [2]

	if err := store.UpdateCommits(context.Background(), 50, map[string][]string{
		makeCommit(1): {},
		makeCommit(2): {makeCommit(1)},
	}); err != nil {
		t.Fatalf("unexpected error updating commits: %s", err)
	}

	insertUploads(t, dbconn.Global,
		Upload{ID: 1, Commit: makeCommit(2), Root: "root1/"},
		Upload{ID: 2, Commit: makeCommit(2), Root: "root2/"},
	)

	testFindClosestDumps(t, store, []FindClosestDumpsTestCase{
		{commit: makeCommit(1), file: "blah", rootMustEnclosePath: true},
		{commit: makeCommit(2), file: "root1/file.ts", rootMustEnclosePath: true, allOfIDs: []int{1}},
		{commit: makeCommit(1), file: "root2/file.ts", rootMustEnclosePath: true, allOfIDs: []int{2}},
		{commit: makeCommit(2), file: "root2/file.ts", rootMustEnclosePath: true, allOfIDs: []int{2}},
		{commit: makeCommit(1), file: "root3/file.ts", rootMustEnclosePath: true},
	})
}

func TestFindClosestDumpsOverlappingRoots(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	// This database has the following commit graph:
	//
	// 1 -- 2 --+-- 3 --+-- 5 -- 6
	//          |       |
	//          +-- 4 --+
	//
	// With the following LSIF dumps:
	//
	// | Commit | Root    | Indexer |
	// | ------ + ------- + ------- |
	// | 1      | root3/  | lsif-go |
	// | 1      | root4/  | lsif-py |
	// | 2      | root1/  | lsif-go |
	// | 2      | root2/  | lsif-go |
	// | 2      |         | lsif-py | (overwrites root4/ at commit 1)
	// | 3      | root1/  | lsif-go | (overwrites root1/ at commit 2)
	// | 4      |         | lsif-py | (overwrites (root) at commit 2)
	// | 5      | root2/  | lsif-go | (overwrites root2/ at commit 2)
	// | 6      | root1/  | lsif-go | (overwrites root1/ at commit 2)

	if err := store.UpdateCommits(context.Background(), 50, map[string][]string{
		makeCommit(1): {},
		makeCommit(2): {makeCommit(1)},
		makeCommit(3): {makeCommit(2)},
		makeCommit(4): {makeCommit(2)},
		makeCommit(5): {makeCommit(3), makeCommit(4)},
		makeCommit(6): {makeCommit(5)},
	}); err != nil {
		t.Fatalf("unexpected error updating commits: %s", err)
	}

	insertUploads(t, dbconn.Global,
		Upload{ID: 1, Commit: makeCommit(1), Root: "root3/"},
		Upload{ID: 2, Commit: makeCommit(1), Root: "root4/", Indexer: "lsif-py"},
		Upload{ID: 3, Commit: makeCommit(2), Root: "root1/"},
		Upload{ID: 4, Commit: makeCommit(2), Root: "root2/"},
		Upload{ID: 5, Commit: makeCommit(2), Root: "", Indexer: "lsif-py"},
		Upload{ID: 6, Commit: makeCommit(3), Root: "root1/"},
		Upload{ID: 7, Commit: makeCommit(4), Root: "", Indexer: "lsif-py"},
		Upload{ID: 8, Commit: makeCommit(5), Root: "root2/"},
		Upload{ID: 9, Commit: makeCommit(6), Root: "root1/"},
	)

	testFindClosestDumps(t, store, []FindClosestDumpsTestCase{
		{commit: makeCommit(4), file: "root1/file.ts", rootMustEnclosePath: true, allOfIDs: []int{7, 3}},
		{commit: makeCommit(5), file: "root2/file.ts", rootMustEnclosePath: true, allOfIDs: []int{8, 7}},
		{commit: makeCommit(3), file: "root3/file.ts", rootMustEnclosePath: true, allOfIDs: []int{5, 1}},
		{commit: makeCommit(1), file: "root4/file.ts", rootMustEnclosePath: true, allOfIDs: []int{2}},
		{commit: makeCommit(2), file: "root4/file.ts", rootMustEnclosePath: true, allOfIDs: []int{5}},
	})
}

func TestFindClosestDumpsMaxTraversalLimit(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	// This repository has the following commit graph (ancestors to the left):
	//
	// MAX_TRAVERSAL_LIMIT -- ... -- 2 -- 1 -- 0
	//
	// At commit `50`, the traversal limit will be reached before visiting commit `0`
	// because commits are visited in this order:
	//
	// | depth | commit |
	// | ----- | ------ |
	// | 1     | 50     | (with direction 'A')
	// | 2     | 50     | (with direction 'D')
	// | 3     | 51     |
	// | 4     | 49     |
	// | 5     | 52     |
	// | 6     | 48     |
	// | ...   |        |
	// | 99    | 99     |
	// | 100   | 1      | (limit reached)

	commits := map[string][]string{}
	for i := 0; i < MaxTraversalLimit; i++ {
		commits[makeCommit(i)] = []string{makeCommit(i + 1)}
	}

	if err := store.UpdateCommits(context.Background(), 50, commits); err != nil {
		t.Fatalf("unexpected error updating commits: %s", err)
	}

	insertUploads(t, dbconn.Global, Upload{ID: 1, Commit: makeCommit(0)})

	testFindClosestDumps(t, store, []FindClosestDumpsTestCase{
		{commit: makeCommit(0), file: "file.ts", rootMustEnclosePath: true, allOfIDs: []int{1}},
		{commit: makeCommit(1), file: "file.ts", rootMustEnclosePath: true, allOfIDs: []int{1}},
		{commit: makeCommit(MaxTraversalLimit/2 - 1), file: "file.ts", rootMustEnclosePath: true, allOfIDs: []int{1}},
		{commit: makeCommit(MaxTraversalLimit / 2), file: "file.ts", rootMustEnclosePath: true},
	})
}

func TestFindClosestDumpsIndexerName(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	// This database has the following commit graph:
	//
	// [1] --+-- [2] --+-- [3] --+-- [4] --+-- 5

	if err := store.UpdateCommits(context.Background(), 50, map[string][]string{
		makeCommit(1): {},
		makeCommit(2): {makeCommit(1)},
		makeCommit(3): {makeCommit(2)},
		makeCommit(4): {makeCommit(3)},
		makeCommit(5): {makeCommit(4)},
	}); err != nil {
		t.Fatalf("unexpected error updating commits: %s", err)
	}

	insertUploads(t, dbconn.Global,
		Upload{ID: 1, Commit: makeCommit(1), Root: "root1/", Indexer: "idx1"},
		Upload{ID: 2, Commit: makeCommit(2), Root: "root2/", Indexer: "idx1"},
		Upload{ID: 3, Commit: makeCommit(3), Root: "root3/", Indexer: "idx1"},
		Upload{ID: 4, Commit: makeCommit(4), Root: "root4/", Indexer: "idx1"},
		Upload{ID: 5, Commit: makeCommit(1), Root: "root1/", Indexer: "idx2"},
		Upload{ID: 6, Commit: makeCommit(2), Root: "root2/", Indexer: "idx2"},
		Upload{ID: 7, Commit: makeCommit(3), Root: "root3/", Indexer: "idx2"},
		Upload{ID: 8, Commit: makeCommit(4), Root: "root4/", Indexer: "idx2"},
	)

	testFindClosestDumps(t, store, []FindClosestDumpsTestCase{
		{commit: makeCommit(5), file: "root1/file.ts", indexer: "idx1", allOfIDs: []int{1}},
		{commit: makeCommit(5), file: "root2/file.ts", indexer: "idx1", allOfIDs: []int{2}},
		{commit: makeCommit(5), file: "root3/file.ts", indexer: "idx1", allOfIDs: []int{3}},
		{commit: makeCommit(5), file: "root4/file.ts", indexer: "idx1", allOfIDs: []int{4}},
		{commit: makeCommit(5), file: "root1/file.ts", indexer: "idx2", allOfIDs: []int{5}},
		{commit: makeCommit(5), file: "root2/file.ts", indexer: "idx2", allOfIDs: []int{6}},
		{commit: makeCommit(5), file: "root3/file.ts", indexer: "idx2", allOfIDs: []int{7}},
		{commit: makeCommit(5), file: "root4/file.ts", indexer: "idx2", allOfIDs: []int{8}},
	})
}

func TestFindClosestDumpsIntersectingPath(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	if err := store.UpdateCommits(context.Background(), 50, map[string][]string{
		makeCommit(1): {},
	}); err != nil {
		t.Fatalf("unexpected error updating commits: %s", err)
	}

	insertUploads(t, dbconn.Global,
		Upload{ID: 1, Commit: makeCommit(1), Root: "web/src/", Indexer: "lsif-eslint"},
	)

	testFindClosestDumps(t, store, []FindClosestDumpsTestCase{
		{commit: makeCommit(1), file: "", rootMustEnclosePath: false, allOfIDs: []int{1}},
		{commit: makeCommit(1), file: "web/", rootMustEnclosePath: false, allOfIDs: []int{1}},
		{commit: makeCommit(1), file: "web/src/file.ts", rootMustEnclosePath: false, allOfIDs: []int{1}},
	})
}

func TestDeleteOldestDump(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	// Cannot prune empty dump set
	if _, prunable, err := store.DeleteOldestDump(context.Background()); err != nil {
		t.Fatalf("unexpected error pruning dumps: %s", err)
	} else if prunable {
		t.Fatal("unexpectedly prunable")
	}

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(time.Minute)
	t3 := t1.Add(time.Minute * 2)
	t4 := t1.Add(time.Minute * 3)

	insertUploads(t, dbconn.Global,
		Upload{ID: 1, UploadedAt: t1},
		Upload{ID: 2, UploadedAt: t2, VisibleAtTip: true},
		Upload{ID: 3, UploadedAt: t3},
		Upload{ID: 4, UploadedAt: t4},
	)

	// Prune oldest
	if id, prunable, err := store.DeleteOldestDump(context.Background()); err != nil {
		t.Fatalf("unexpected error pruning dumps: %s", err)
	} else if !prunable {
		t.Fatal("unexpectedly non-prunable")
	} else if id != 1 {
		t.Errorf("unexpected pruned identifier. want=%d have=%d", 1, id)
	}

	// Prune next oldest (skips visible at tip)
	if id, prunable, err := store.DeleteOldestDump(context.Background()); err != nil {
		t.Fatalf("unexpected error pruning dumps: %s", err)
	} else if !prunable {
		t.Fatal("unexpectedly non-prunable")
	} else if id != 3 {
		t.Errorf("unexpected pruned identifier. want=%d have=%d", 3, id)
	}
}

type FindClosestDumpsTestCase struct {
	commit              string
	file                string
	rootMustEnclosePath bool
	indexer             string
	anyOfIDs            []int
	allOfIDs            []int
}

func testFindClosestDumps(t *testing.T, store Store, testCases []FindClosestDumpsTestCase) {
	for _, testCase := range testCases {
		name := fmt.Sprintf(
			"commit=%s file=%s rootMustEnclosePath=%v indexer=%s",
			testCase.commit,
			testCase.file,
			testCase.rootMustEnclosePath,
			testCase.indexer,
		)

		t.Run(name, func(t *testing.T) {
			dumps, err := store.FindClosestDumps(context.Background(), 50, testCase.commit, testCase.file, testCase.rootMustEnclosePath, testCase.indexer)
			if err != nil {
				t.Fatalf("unexpected error finding closest dumps: %s", err)
			}

			if len(testCase.anyOfIDs) > 0 {
				testAnyOf(t, dumps, testCase.anyOfIDs)
				return
			}

			if len(testCase.allOfIDs) > 0 {
				testAllOf(t, dumps, testCase.allOfIDs)
				return
			}

			if len(dumps) != 0 {
				t.Errorf("unexpected nearest dump length. want=%d have=%d", 0, len(dumps))
				return
			}
		})
	}
}

func testAnyOf(t *testing.T, dumps []Dump, expectedIDs []int) {
	if len(dumps) != 1 {
		t.Errorf("unexpected nearest dump length. want=%d have=%d", 1, len(dumps))
		return
	}

	if !testPresence(dumps[0].ID, expectedIDs) {
		t.Errorf("unexpected nearest dump ids. want one of %v have=%v", expectedIDs, dumps[0].ID)
	}
}

func testAllOf(t *testing.T, dumps []Dump, expectedIDs []int) {
	if len(dumps) != len(expectedIDs) {
		t.Errorf("unexpected nearest dump length. want=%d have=%d", 1, len(dumps))
	}

	var dumpIDs []int
	for _, dump := range dumps {
		dumpIDs = append(dumpIDs, dump.ID)
	}

	for _, expectedID := range expectedIDs {
		if !testPresence(expectedID, dumpIDs) {
			t.Errorf("unexpected nearest dump ids. want all of %v have=%v", expectedIDs, dumpIDs)
			return
		}
	}

}

func testPresence(needle int, haystack []int) bool {
	for _, candidate := range haystack {
		if needle == candidate {
			return true
		}
	}

	return false
}

func TestUpdateDumpsVisibleFromTipOverlappingRootsSameIndexer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	// This database has the following commit graph:
	//
	// [1] -- [2] -- [3] -- [4] -- 5 -- [6] -- [7]

	insertUploads(t, dbconn.Global,
		Upload{ID: 1, Commit: makeCommit(1), Root: "r1/"},
		Upload{ID: 2, Commit: makeCommit(2), Root: "r2/"},
		Upload{ID: 3, Commit: makeCommit(3)},
		Upload{ID: 4, Commit: makeCommit(4), Root: "r3/"},
		Upload{ID: 5, Commit: makeCommit(6), Root: "r4/"},
		Upload{ID: 6, Commit: makeCommit(7), Root: "r5/"},
	)

	if err := store.UpdateCommits(context.Background(), 50, map[string][]string{
		makeCommit(1): {},
		makeCommit(2): {makeCommit(1)},
		makeCommit(3): {makeCommit(2)},
		makeCommit(4): {makeCommit(3)},
		makeCommit(5): {makeCommit(4)},
		makeCommit(6): {makeCommit(5)},
		makeCommit(7): {makeCommit(6)},
	}); err != nil {
		t.Fatalf("unexpected error updating commits: %s", err)
	}

	err := store.UpdateDumpsVisibleFromTip(context.Background(), 50, makeCommit(6))
	if err != nil {
		t.Fatalf("unexpected error updating dumps visible from tip: %s", err)
	}

	visibilities := getDumpVisibilities(t, dbconn.Global)
	expected := map[int]bool{1: false, 2: false, 3: false, 4: true, 5: true, 6: false}

	if diff := cmp.Diff(expected, visibilities); diff != "" {
		t.Errorf("unexpected visibility (-want +got):\n%s", diff)
	}
}

func TestUpdateDumpsVisibleFromTipOverlappingRoots(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	// This database has the following commit graph:
	//
	// [1] -- 2 -- [3] -- [4] -- [5] -- [6] -- [7]

	insertUploads(t, dbconn.Global,
		Upload{ID: 1, Commit: makeCommit(1), Root: "r1/"},
		Upload{ID: 2, Commit: makeCommit(3), Root: "r2/"},
		Upload{ID: 3, Commit: makeCommit(4), Root: "r1/"},
		Upload{ID: 4, Commit: makeCommit(6), Root: "r3/"},
		Upload{ID: 5, Commit: makeCommit(7), Root: "r4/"},
		Upload{ID: 6, Commit: makeCommit(1), Root: "r1/", Indexer: "lsif-tsc"},
		Upload{ID: 7, Commit: makeCommit(3), Root: "r2/", Indexer: "lsif-tsc"},
		Upload{ID: 8, Commit: makeCommit(4), Indexer: "lsif-tsc"},
		Upload{ID: 9, Commit: makeCommit(5), Root: "r3/", Indexer: "lsif-tsc"},
	)

	if err := store.UpdateCommits(context.Background(), 50, map[string][]string{
		makeCommit(1): {},
		makeCommit(2): {makeCommit(1)},
		makeCommit(3): {makeCommit(2)},
		makeCommit(4): {makeCommit(3)},
		makeCommit(5): {makeCommit(4)},
		makeCommit(6): {makeCommit(5)},
		makeCommit(7): {makeCommit(6)},
	}); err != nil {
		t.Fatalf("unexpected error updating commits: %s", err)
	}

	err := store.UpdateDumpsVisibleFromTip(context.Background(), 50, makeCommit(6))
	if err != nil {
		t.Fatalf("unexpected error updating dumps visible from tip: %s", err)
	}

	visibilities := getDumpVisibilities(t, dbconn.Global)
	expected := map[int]bool{1: false, 2: true, 3: true, 4: true, 5: false, 6: false, 7: false, 8: false, 9: true}
	if diff := cmp.Diff(expected, visibilities); diff != "" {
		t.Errorf("unexpected visibility. want=%v have=%v", expected, visibilities)
	}
}

func TestUpdateDumpsVisibleFromTipBranchingPaths(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	// This database has the following commit graph:
	//
	// 1 --+-- [2] --- 3 ---+
	//     |                |
	//     +--- 4 --- [5] --+ -- [8] --+-- [9]
	//     |                           |
	//     +-- [6] --- 7 --------------+

	insertUploads(t, dbconn.Global,
		Upload{ID: 1, Commit: makeCommit(2), Root: "r2/"},
		Upload{ID: 2, Commit: makeCommit(5), Root: "r2/a/"},
		Upload{ID: 3, Commit: makeCommit(5), Root: "r2/b/"},
		Upload{ID: 4, Commit: makeCommit(6), Root: "r1/a/"},
		Upload{ID: 5, Commit: makeCommit(6), Root: "r1/b/"},
		Upload{ID: 6, Commit: makeCommit(8), Root: "r1/"},
		Upload{ID: 7, Commit: makeCommit(9), Root: "r3/"},
	)

	if err := store.UpdateCommits(context.Background(), 50, map[string][]string{
		makeCommit(1): {},
		makeCommit(2): {makeCommit(1)},
		makeCommit(3): {makeCommit(2)},
		makeCommit(4): {makeCommit(1)},
		makeCommit(5): {makeCommit(4)},
		makeCommit(8): {makeCommit(5), makeCommit(3)},
		makeCommit(9): {makeCommit(7), makeCommit(8)},
		makeCommit(6): {makeCommit(1)},
		makeCommit(7): {makeCommit(6)},
	}); err != nil {
		t.Fatalf("unexpected error updating commits: %s", err)
	}

	err := store.UpdateDumpsVisibleFromTip(context.Background(), 50, makeCommit(9))
	if err != nil {
		t.Fatalf("unexpected error updating dumps visible from tip: %s", err)
	}

	visibilities := getDumpVisibilities(t, dbconn.Global)
	expected := map[int]bool{1: false, 2: true, 3: true, 4: false, 5: false, 6: true, 7: true}
	if diff := cmp.Diff(expected, visibilities); diff != "" {
		t.Errorf("unexpected visibility (-want +got):\n%s", diff)
	}
}

func TestUpdateDumpsVisibleFromTipMaxTraversalLimit(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	// This repository has the following commit graph (ancestors to the left):
	//
	// (MAX_TRAVERSAL_LIMIT + 1) -- ... -- 2 -- 1 -- 0

	commits := map[string][]string{}
	for i := 0; i < MaxTraversalLimit+1; i++ {
		commits[makeCommit(i)] = []string{makeCommit(i + 1)}
	}

	if err := store.UpdateCommits(context.Background(), 50, commits); err != nil {
		t.Fatalf("unexpected error updating commits: %s", err)
	}

	insertUploads(t, dbconn.Global, Upload{ID: 1, Commit: makeCommit(MaxTraversalLimit)})

	if err := store.UpdateDumpsVisibleFromTip(context.Background(), 50, makeCommit(MaxTraversalLimit)); err != nil {
		t.Fatalf("unexpected error updating dumps visible from tip: %s", err)
	} else {
		visibilities := getDumpVisibilities(t, dbconn.Global)
		expected := map[int]bool{1: true}
		if diff := cmp.Diff(expected, visibilities); diff != "" {
			t.Errorf("unexpected visibility (-want +got):\n%s", diff)
		}
	}

	if err := store.UpdateDumpsVisibleFromTip(context.Background(), 50, makeCommit(1)); err != nil {
		t.Fatalf("unexpected error updating dumps visible from tip: %s", err)
	} else {
		visibilities := getDumpVisibilities(t, dbconn.Global)
		expected := map[int]bool{1: true}
		if diff := cmp.Diff(expected, visibilities); diff != "" {
			t.Errorf("unexpected visibility (-want +got):\n%s", diff)
		}
	}

	if err := store.UpdateDumpsVisibleFromTip(context.Background(), 50, makeCommit(0)); err != nil {
		t.Fatalf("unexpected error updating dumps visible from tip: %s", err)
	} else {
		visibilities := getDumpVisibilities(t, dbconn.Global)
		expected := map[int]bool{1: false}
		if diff := cmp.Diff(expected, visibilities); diff != "" {
			t.Errorf("unexpected visibility (-want +got):\n%s", diff)
		}
	}
}

func TestDeleteOverlappingDumps(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	insertUploads(t, dbconn.Global, Upload{
		ID:      1,
		Commit:  makeCommit(1),
		Root:    "cmd/",
		Indexer: "lsif-go",
	})

	err := store.DeleteOverlappingDumps(context.Background(), 50, makeCommit(1), "cmd/", "lsif-go")
	if err != nil {
		t.Fatalf("unexpected error deleting dump: %s", err)
	}

	// Original dump no longer exists
	if _, exists, err := store.GetDumpByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting dump: %s", err)
	} else if exists {
		t.Fatal("unexpected record")
	}
}

func TestDeleteOverlappingDumpsNoMatches(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	insertUploads(t, dbconn.Global, Upload{
		ID:      1,
		Commit:  makeCommit(1),
		Root:    "cmd/",
		Indexer: "lsif-go",
	})

	testCases := []struct {
		commit  string
		root    string
		indexer string
	}{
		{makeCommit(2), "cmd/", "lsif-go"},
		{makeCommit(1), "cmds/", "lsif-go"},
		{makeCommit(1), "cmd/", "lsif-tsc"},
	}

	for _, testCase := range testCases {
		err := store.DeleteOverlappingDumps(context.Background(), 50, testCase.commit, testCase.root, testCase.indexer)
		if err != nil {
			t.Fatalf("unexpected error deleting dump: %s", err)
		}
	}

	// Original dump still exists
	if _, exists, err := store.GetDumpByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting dump: %s", err)
	} else if !exists {
		t.Fatal("expected dump record to still exist")
	}
}

func TestDeleteOverlappingDumpsIgnoresIncompleteUploads(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	insertUploads(t, dbconn.Global, Upload{
		ID:      1,
		Commit:  makeCommit(1),
		Root:    "cmd/",
		Indexer: "lsif-go",
		State:   "queued",
	})

	err := store.DeleteOverlappingDumps(context.Background(), 50, makeCommit(1), "cmd/", "lsif-go")
	if err != nil {
		t.Fatalf("unexpected error deleting dump: %s", err)
	}

	// Original upload still exists
	if _, exists, err := store.GetUploadByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting dump: %s", err)
	} else if !exists {
		t.Fatal("expected dump record to still exist")
	}
}
