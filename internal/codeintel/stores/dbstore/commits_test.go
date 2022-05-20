package dbstore

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/commitgraph"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

func TestHasRepository(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

	testCases := []struct {
		repositoryID int
		exists       bool
	}{
		{50, true},
		{51, false},
		{52, false},
	}

	insertUploads(t, db, Upload{ID: 1, RepositoryID: 50})
	insertUploads(t, db, Upload{ID: 2, RepositoryID: 51, State: "deleted"})

	for _, testCase := range testCases {
		name := fmt.Sprintf("repositoryID=%d", testCase.repositoryID)

		t.Run(name, func(t *testing.T) {
			exists, err := store.HasRepository(context.Background(), testCase.repositoryID)
			if err != nil {
				t.Fatalf("unexpected error checking if repository exists: %s", err)
			}
			if exists != testCase.exists {
				t.Errorf("unexpected exists. want=%v have=%v", testCase.exists, exists)
			}
		})
	}
}

func TestHasCommit(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

	testCases := []struct {
		repositoryID int
		commit       string
		exists       bool
	}{
		{50, makeCommit(1), true},
		{50, makeCommit(2), false},
		{51, makeCommit(1), false},
	}

	insertNearestUploads(t, db, 50, map[string][]commitgraph.UploadMeta{makeCommit(1): {{UploadID: 42, Distance: 1}}})
	insertNearestUploads(t, db, 51, map[string][]commitgraph.UploadMeta{makeCommit(2): {{UploadID: 43, Distance: 2}}})

	for _, testCase := range testCases {
		name := fmt.Sprintf("repositoryID=%d commit=%s", testCase.repositoryID, testCase.commit)

		t.Run(name, func(t *testing.T) {
			exists, err := store.HasCommit(context.Background(), testCase.repositoryID, testCase.commit)
			if err != nil {
				t.Fatalf("unexpected error checking if commit exists: %s", err)
			}
			if exists != testCase.exists {
				t.Errorf("unexpected exists. want=%v have=%v", testCase.exists, exists)
			}
		})
	}
}

func TestMarkRepositoryAsDirty(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

	for _, id := range []int{50, 51, 52} {
		insertRepo(t, db, id, "")
	}

	for _, repositoryID := range []int{50, 51, 52, 51, 52} {
		if err := store.MarkRepositoryAsDirty(context.Background(), repositoryID); err != nil {
			t.Errorf("unexpected error marking repository as dirty: %s", err)
		}
	}

	repositoryIDs, err := store.DirtyRepositories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error listing dirty repositories: %s", err)
	}

	var keys []int
	for repositoryID := range repositoryIDs {
		keys = append(keys, repositoryID)
	}
	sort.Ints(keys)

	if diff := cmp.Diff([]int{50, 51, 52}, keys); diff != "" {
		t.Errorf("unexpected repository ids (-want +got):\n%s", diff)
	}
}

func TestMaxStaleAge(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

	for _, id := range []int{50, 51, 52} {
		insertRepo(t, db, id, "")
	}

	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO lsif_dirty_repositories (
			repository_id,
			update_token,
			dirty_token,
			set_dirty_at
		)
		VALUES
			(50, 10, 10, NOW() - '45 minutes'::interval), -- not dirty
			(51, 20, 25, NOW() - '30 minutes'::interval), -- dirty
			(52, 30, 35, NOW() - '20 minutes'::interval), -- dirty
			(53, 40, 45, NOW() - '30 minutes'::interval); -- no associated repo
	`); err != nil {
		t.Fatalf("unexpected error marking repostiory as dirty: %s", err)
	}

	age, err := store.MaxStaleAge(context.Background())
	if err != nil {
		t.Fatalf("unexpected error listing dirty repositories: %s", err)
	}
	if age.Round(time.Second) != 30*time.Minute {
		t.Fatalf("unexpected max age. want=%s have=%s", 30*time.Minute, age)
	}
}

func TestSkipsDeletedRepositories(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

	insertRepo(t, db, 50, "should not be dirty")
	deleteRepo(t, db, 50, time.Now())

	insertRepo(t, db, 51, "should be dirty")

	// NOTE: We did not insert 52, so it should not show up as dirty, even though we mark it below.

	for _, repositoryID := range []int{50, 51, 52} {
		if err := store.MarkRepositoryAsDirty(context.Background(), repositoryID); err != nil {
			t.Fatalf("unexpected error marking repository as dirty: %s", err)
		}
	}

	repositoryIDs, err := store.DirtyRepositories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error listing dirty repositories: %s", err)
	}

	var keys []int
	for repositoryID := range repositoryIDs {
		keys = append(keys, repositoryID)
	}
	sort.Ints(keys)

	if diff := cmp.Diff([]int{51}, keys); diff != "" {
		t.Errorf("unexpected repository ids (-want +got):\n%s", diff)
	}
}

func TestCommitGraphMetadata(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

	if err := store.MarkRepositoryAsDirty(context.Background(), 50); err != nil {
		t.Errorf("unexpected error marking repository as dirty: %s", err)
	}

	updatedAt := time.Unix(1587396557, 0).UTC()
	query := sqlf.Sprintf("INSERT INTO lsif_dirty_repositories VALUES (%s, %s, %s, %s)", 51, 10, 10, updatedAt)
	if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error inserting commit graph metadata: %s", err)
	}

	testCases := []struct {
		RepositoryID int
		Stale        bool
		UpdatedAt    *time.Time
	}{
		{50, true, nil},
		{51, false, &updatedAt},
		{52, false, nil},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("repositoryID=%d", testCase.RepositoryID), func(t *testing.T) {
			stale, updatedAt, err := store.CommitGraphMetadata(context.Background(), testCase.RepositoryID)
			if err != nil {
				t.Fatalf("unexpected error getting commit graph metadata: %s", err)
			}

			if stale != testCase.Stale {
				t.Errorf("unexpected value for stale. want=%v have=%v", testCase.Stale, stale)
			}

			if diff := cmp.Diff(testCase.UpdatedAt, updatedAt); diff != "" {
				t.Errorf("unexpected value for uploadedAt (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCalculateVisibleUploads(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

	// This database has the following commit graph:
	//
	// [1] --+--- 2 --------+--5 -- 6 --+-- [7]
	//       |              |           |
	//       +-- [3] -- 4 --+           +--- 8

	uploads := []Upload{
		{ID: 1, Commit: makeCommit(1)},
		{ID: 2, Commit: makeCommit(3)},
		{ID: 3, Commit: makeCommit(7)},
	}
	insertUploads(t, db, uploads...)

	graph := gitdomain.ParseCommitGraph([]string{
		strings.Join([]string{makeCommit(8), makeCommit(6)}, " "),
		strings.Join([]string{makeCommit(7), makeCommit(6)}, " "),
		strings.Join([]string{makeCommit(6), makeCommit(5)}, " "),
		strings.Join([]string{makeCommit(5), makeCommit(2), makeCommit(4)}, " "),
		strings.Join([]string{makeCommit(4), makeCommit(3)}, " "),
		strings.Join([]string{makeCommit(3), makeCommit(1)}, " "),
		strings.Join([]string{makeCommit(2), makeCommit(1)}, " "),
		strings.Join([]string{makeCommit(1)}, " "),
	})

	refDescriptions := map[string][]gitdomain.RefDescription{
		makeCommit(8): {{IsDefaultBranch: true}},
	}

	if err := store.CalculateVisibleUploads(context.Background(), 50, graph, refDescriptions, time.Hour, time.Hour, 0); err != nil {
		t.Fatalf("unexpected error while calculating visible uploads: %s", err)
	}

	expectedVisibleUploads := map[string][]int{
		makeCommit(1): {1},
		makeCommit(2): {1},
		makeCommit(3): {2},
		makeCommit(4): {2},
		makeCommit(5): {1},
		makeCommit(6): {1},
		makeCommit(7): {3},
		makeCommit(8): {1},
	}
	if diff := cmp.Diff(expectedVisibleUploads, getVisibleUploads(t, db, 50, keysOf(expectedVisibleUploads))); diff != "" {
		t.Errorf("unexpected visible uploads (-want +got):\n%s", diff)
	}

	// Ensure data can be queried in reverse direction as well
	assertCommitsVisibleFromUploads(t, store, uploads, expectedVisibleUploads)

	if diff := cmp.Diff([]int{1}, getUploadsVisibleAtTip(t, db, 50)); diff != "" {
		t.Errorf("unexpected uploads visible at tip (-want +got):\n%s", diff)
	}
}

func TestCalculateVisibleUploadsAlternateCommitGraph(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

	// This database has the following commit graph:
	//
	// 1 --+-- [2] ---- 3
	//     |
	//     +--- 4 --+-- 5 -- 6
	//              |
	//              +-- 7 -- 8

	uploads := []Upload{
		{ID: 1, Commit: makeCommit(2)},
	}
	insertUploads(t, db, uploads...)

	graph := gitdomain.ParseCommitGraph([]string{
		strings.Join([]string{makeCommit(8), makeCommit(7)}, " "),
		strings.Join([]string{makeCommit(7), makeCommit(4)}, " "),
		strings.Join([]string{makeCommit(6), makeCommit(5)}, " "),
		strings.Join([]string{makeCommit(5), makeCommit(4)}, " "),
		strings.Join([]string{makeCommit(4), makeCommit(1)}, " "),
		strings.Join([]string{makeCommit(3), makeCommit(2)}, " "),
		strings.Join([]string{makeCommit(2), makeCommit(1)}, " "),
		strings.Join([]string{makeCommit(1)}, " "),
	})

	refDescriptions := map[string][]gitdomain.RefDescription{
		makeCommit(3): {{IsDefaultBranch: true}},
	}

	if err := store.CalculateVisibleUploads(context.Background(), 50, graph, refDescriptions, time.Hour, time.Hour, 0); err != nil {
		t.Fatalf("unexpected error while calculating visible uploads: %s", err)
	}

	expectedVisibleUploads := map[string][]int{
		makeCommit(2): {1},
		makeCommit(3): {1},
	}
	if diff := cmp.Diff(expectedVisibleUploads, getVisibleUploads(t, db, 50, keysOf(expectedVisibleUploads))); diff != "" {
		t.Errorf("unexpected visible uploads (-want +got):\n%s", diff)
	}

	// Ensure data can be queried in reverse direction as well
	assertCommitsVisibleFromUploads(t, store, uploads, expectedVisibleUploads)

	if diff := cmp.Diff([]int{1}, getUploadsVisibleAtTip(t, db, 50)); diff != "" {
		t.Errorf("unexpected uploads visible at tip (-want +got):\n%s", diff)
	}
}

func TestCalculateVisibleUploadsDistinctRoots(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

	// This database has the following commit graph:
	//
	// 1 -- [2]

	uploads := []Upload{
		{ID: 1, Commit: makeCommit(2), Root: "root1/"},
		{ID: 2, Commit: makeCommit(2), Root: "root2/"},
	}
	insertUploads(t, db, uploads...)

	graph := gitdomain.ParseCommitGraph([]string{
		strings.Join([]string{makeCommit(2), makeCommit(1)}, " "),
		strings.Join([]string{makeCommit(1)}, " "),
	})

	refDescriptions := map[string][]gitdomain.RefDescription{
		makeCommit(2): {{IsDefaultBranch: true}},
	}

	if err := store.CalculateVisibleUploads(context.Background(), 50, graph, refDescriptions, time.Hour, time.Hour, 0); err != nil {
		t.Fatalf("unexpected error while calculating visible uploads: %s", err)
	}

	expectedVisibleUploads := map[string][]int{
		makeCommit(2): {1, 2},
	}
	if diff := cmp.Diff(expectedVisibleUploads, getVisibleUploads(t, db, 50, keysOf(expectedVisibleUploads))); diff != "" {
		t.Errorf("unexpected visible uploads (-want +got):\n%s", diff)
	}

	// Ensure data can be queried in reverse direction as well
	assertCommitsVisibleFromUploads(t, store, uploads, expectedVisibleUploads)

	if diff := cmp.Diff([]int{1, 2}, getUploadsVisibleAtTip(t, db, 50)); diff != "" {
		t.Errorf("unexpected uploads visible at tip (-want +got):\n%s", diff)
	}
}

func TestCalculateVisibleUploadsOverlappingRoots(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

	// This database has the following commit graph:
	//
	// 1 -- 2 --+-- 3 --+-- 5 -- 6
	//          |       |
	//          +-- 4 --+
	//
	// With the following LSIF dumps:
	//
	// | UploadID | Commit | Root    | Indexer |
	// | -------- + ------ + ------- + ------- |
	// | 1        | 1      | root3/  | lsif-go |
	// | 2        | 1      | root4/  | lsif-py |
	// | 3        | 2      | root1/  | lsif-go |
	// | 4        | 2      | root2/  | lsif-go |
	// | 5        | 2      |         | lsif-py | (overwrites root4/ at commit 1)
	// | 6        | 3      | root1/  | lsif-go | (overwrites root1/ at commit 2)
	// | 7        | 4      |         | lsif-py | (overwrites (root) at commit 2)
	// | 8        | 5      | root2/  | lsif-go | (overwrites root2/ at commit 2)
	// | 9        | 6      | root1/  | lsif-go | (overwrites root1/ at commit 2)

	uploads := []Upload{
		{ID: 1, Commit: makeCommit(1), Indexer: "lsif-go", Root: "root3/"},
		{ID: 2, Commit: makeCommit(1), Indexer: "lsif-py", Root: "root4/"},
		{ID: 3, Commit: makeCommit(2), Indexer: "lsif-go", Root: "root1/"},
		{ID: 4, Commit: makeCommit(2), Indexer: "lsif-go", Root: "root2/"},
		{ID: 5, Commit: makeCommit(2), Indexer: "lsif-py", Root: ""},
		{ID: 6, Commit: makeCommit(3), Indexer: "lsif-go", Root: "root1/"},
		{ID: 7, Commit: makeCommit(4), Indexer: "lsif-py", Root: ""},
		{ID: 8, Commit: makeCommit(5), Indexer: "lsif-go", Root: "root2/"},
		{ID: 9, Commit: makeCommit(6), Indexer: "lsif-go", Root: "root1/"},
	}
	insertUploads(t, db, uploads...)

	graph := gitdomain.ParseCommitGraph([]string{
		strings.Join([]string{makeCommit(6), makeCommit(5)}, " "),
		strings.Join([]string{makeCommit(5), makeCommit(3), makeCommit(4)}, " "),
		strings.Join([]string{makeCommit(4), makeCommit(2)}, " "),
		strings.Join([]string{makeCommit(3), makeCommit(2)}, " "),
		strings.Join([]string{makeCommit(2), makeCommit(1)}, " "),
		strings.Join([]string{makeCommit(1)}, " "),
	})

	refDescriptions := map[string][]gitdomain.RefDescription{
		makeCommit(6): {{IsDefaultBranch: true}},
	}

	if err := store.CalculateVisibleUploads(context.Background(), 50, graph, refDescriptions, time.Hour, time.Hour, 0); err != nil {
		t.Fatalf("unexpected error while calculating visible uploads: %s", err)
	}

	expectedVisibleUploads := map[string][]int{
		makeCommit(1): {1, 2},
		makeCommit(2): {1, 2, 3, 4, 5},
		makeCommit(3): {1, 2, 4, 5, 6},
		makeCommit(4): {1, 2, 3, 4, 7},
		makeCommit(5): {1, 2, 6, 7, 8},
		makeCommit(6): {1, 2, 7, 8, 9},
	}
	if diff := cmp.Diff(expectedVisibleUploads, getVisibleUploads(t, db, 50, keysOf(expectedVisibleUploads))); diff != "" {
		t.Errorf("unexpected visible uploads (-want +got):\n%s", diff)
	}

	// Ensure data can be queried in reverse direction as well
	assertCommitsVisibleFromUploads(t, store, uploads, expectedVisibleUploads)

	if diff := cmp.Diff([]int{1, 2, 7, 8, 9}, getUploadsVisibleAtTip(t, db, 50)); diff != "" {
		t.Errorf("unexpected uploads visible at tip (-want +got):\n%s", diff)
	}
}

func TestCalculateVisibleUploadsIndexerName(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

	// This database has the following commit graph:
	//
	// [1] -- [2] -- [3] -- [4] -- 5

	uploads := []Upload{
		{ID: 1, Commit: makeCommit(1), Root: "root1/", Indexer: "idx1"},
		{ID: 2, Commit: makeCommit(2), Root: "root2/", Indexer: "idx1"},
		{ID: 3, Commit: makeCommit(3), Root: "root3/", Indexer: "idx1"},
		{ID: 4, Commit: makeCommit(4), Root: "root4/", Indexer: "idx1"},
		{ID: 5, Commit: makeCommit(1), Root: "root1/", Indexer: "idx2"},
		{ID: 6, Commit: makeCommit(2), Root: "root2/", Indexer: "idx2"},
		{ID: 7, Commit: makeCommit(3), Root: "root3/", Indexer: "idx2"},
		{ID: 8, Commit: makeCommit(4), Root: "root4/", Indexer: "idx2"},
	}
	insertUploads(t, db, uploads...)

	graph := gitdomain.ParseCommitGraph([]string{
		strings.Join([]string{makeCommit(5), makeCommit(4)}, " "),
		strings.Join([]string{makeCommit(4), makeCommit(3)}, " "),
		strings.Join([]string{makeCommit(3), makeCommit(2)}, " "),
		strings.Join([]string{makeCommit(2), makeCommit(1)}, " "),
		strings.Join([]string{makeCommit(1)}, " "),
	})

	refDescriptions := map[string][]gitdomain.RefDescription{
		makeCommit(5): {{IsDefaultBranch: true}},
	}

	if err := store.CalculateVisibleUploads(context.Background(), 50, graph, refDescriptions, time.Hour, time.Hour, 0); err != nil {
		t.Fatalf("unexpected error while calculating visible uploads: %s", err)
	}

	expectedVisibleUploads := map[string][]int{
		makeCommit(1): {1, 5},
		makeCommit(2): {1, 2, 5, 6},
		makeCommit(3): {1, 2, 3, 5, 6, 7},
		makeCommit(4): {1, 2, 3, 4, 5, 6, 7, 8},
		makeCommit(5): {1, 2, 3, 4, 5, 6, 7, 8},
	}
	if diff := cmp.Diff(expectedVisibleUploads, getVisibleUploads(t, db, 50, keysOf(expectedVisibleUploads))); diff != "" {
		t.Errorf("unexpected visible uploads (-want +got):\n%s", diff)
	}

	// Ensure data can be queried in reverse direction as well
	assertCommitsVisibleFromUploads(t, store, uploads, expectedVisibleUploads)

	if diff := cmp.Diff([]int{1, 2, 3, 4, 5, 6, 7, 8}, getUploadsVisibleAtTip(t, db, 50)); diff != "" {
		t.Errorf("unexpected uploads visible at tip (-want +got):\n%s", diff)
	}
}

func TestCalculateVisibleUploadsResetsDirtyFlag(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

	uploads := []Upload{
		{ID: 1, Commit: makeCommit(1)},
		{ID: 2, Commit: makeCommit(2)},
		{ID: 3, Commit: makeCommit(3)},
	}
	insertUploads(t, db, uploads...)

	graph := gitdomain.ParseCommitGraph([]string{
		strings.Join([]string{makeCommit(3), makeCommit(2)}, " "),
		strings.Join([]string{makeCommit(2), makeCommit(1)}, " "),
		strings.Join([]string{makeCommit(1)}, " "),
	})

	refDescriptions := map[string][]gitdomain.RefDescription{
		makeCommit(3): {{IsDefaultBranch: true}},
	}

	for i := 0; i < 3; i++ {
		// Set dirty token to 3
		if err := store.MarkRepositoryAsDirty(context.Background(), 50); err != nil {
			t.Fatalf("unexpected error marking repository as dirty: %s", err)
		}
	}

	now := time.Unix(1587396557, 0).UTC()

	// Non-latest dirty token - should not clear flag
	if err := store.calculateVisibleUploadsWithTime(context.Background(), 50, graph, refDescriptions, time.Hour, time.Hour, 2, now); err != nil {
		t.Fatalf("unexpected error while calculating visible uploads: %s", err)
	}
	repositoryIDs, err := store.DirtyRepositories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error listing dirty repositories: %s", err)
	}
	if len(repositoryIDs) == 0 {
		t.Errorf("did not expect repository to be unmarked")
	}

	// Latest dirty token - should clear flag
	if err := store.calculateVisibleUploadsWithTime(context.Background(), 50, graph, refDescriptions, time.Hour, time.Hour, 3, now); err != nil {
		t.Fatalf("unexpected error while calculating visible uploads: %s", err)
	}
	repositoryIDs, err = store.DirtyRepositories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error listing dirty repositories: %s", err)
	}
	if len(repositoryIDs) != 0 {
		t.Errorf("expected repository to be unmarked")
	}

	stale, updatedAt, err := store.CommitGraphMetadata(context.Background(), 50)
	if err != nil {
		t.Fatalf("unexpected error getting commit graph metadata: %s", err)
	}
	if stale {
		t.Errorf("unexpected value for stale. want=%v have=%v", false, stale)
	}
	if diff := cmp.Diff(&now, updatedAt); diff != "" {
		t.Errorf("unexpected value for uploadedAt (-want +got):\n%s", diff)
	}
}

func TestCalculateVisibleUploadsResetsDirtyFlagTransactionTimestamp(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

	uploads := []Upload{
		{ID: 1, Commit: makeCommit(1)},
		{ID: 2, Commit: makeCommit(2)},
		{ID: 3, Commit: makeCommit(3)},
	}
	insertUploads(t, db, uploads...)

	graph := gitdomain.ParseCommitGraph([]string{
		strings.Join([]string{makeCommit(3), makeCommit(2)}, " "),
		strings.Join([]string{makeCommit(2), makeCommit(1)}, " "),
		strings.Join([]string{makeCommit(1)}, " "),
	})

	refDescriptions := map[string][]gitdomain.RefDescription{
		makeCommit(3): {{IsDefaultBranch: true}},
	}

	for i := 0; i < 3; i++ {
		// Set dirty token to 3
		if err := store.MarkRepositoryAsDirty(context.Background(), 50); err != nil {
			t.Fatalf("unexpected error marking repository as dirty: %s", err)
		}
	}

	// This test is mainly a syntax check against `transaction_timestamp()`
	if err := store.CalculateVisibleUploads(context.Background(), 50, graph, refDescriptions, time.Hour, time.Hour, 3); err != nil {
		t.Fatalf("unexpected error while calculating visible uploads: %s", err)
	}
}

func TestCalculateVisibleUploadsNonDefaultBranches(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

	// This database has the following commit graph:
	//
	//                +-- [08] ----- {09} --+
	//                |                     |
	// [01] -- {02} --+-- [03] --+-- {04} --+-- {05} -- [06] -- {07}
	//                           |
	//                           +--- 10 ------ [11] -- {12}
	//
	// 02: tag v1
	// 04: tag v2
	// 05: tag v3
	// 07: tip of main branch
	// 09: tip of branch feat1
	// 12: tip of branch feat2

	uploads := []Upload{
		{ID: 1, Commit: makeCommit(1)},
		{ID: 2, Commit: makeCommit(3)},
		{ID: 3, Commit: makeCommit(6)},
		{ID: 4, Commit: makeCommit(8)},
		{ID: 5, Commit: makeCommit(11)},
	}
	insertUploads(t, db, uploads...)

	graph := gitdomain.ParseCommitGraph([]string{
		strings.Join([]string{makeCommit(12), makeCommit(11)}, " "),
		strings.Join([]string{makeCommit(11), makeCommit(10)}, " "),
		strings.Join([]string{makeCommit(10), makeCommit(3)}, " "),
		strings.Join([]string{makeCommit(7), makeCommit(6)}, " "),
		strings.Join([]string{makeCommit(6), makeCommit(5)}, " "),
		strings.Join([]string{makeCommit(5), makeCommit(4), makeCommit(9)}, " "),
		strings.Join([]string{makeCommit(9), makeCommit(8)}, " "),
		strings.Join([]string{makeCommit(8), makeCommit(2)}, " "),
		strings.Join([]string{makeCommit(4), makeCommit(3)}, " "),
		strings.Join([]string{makeCommit(3), makeCommit(2)}, " "),
		strings.Join([]string{makeCommit(2), makeCommit(1)}, " "),
		strings.Join([]string{makeCommit(1)}, " "),
	})

	t1 := time.Now().Add(-time.Minute * 90) // > 1 hr
	t2 := time.Now().Add(-time.Minute * 30) // < 1 hr

	refDescriptions := map[string][]gitdomain.RefDescription{
		// stale
		makeCommit(2): {{Name: "v1", Type: gitdomain.RefTypeTag, CreatedDate: &t1}},
		makeCommit(9): {{Name: "feat1", Type: gitdomain.RefTypeBranch, CreatedDate: &t1}},

		// fresh
		makeCommit(4):  {{Name: "v2", Type: gitdomain.RefTypeTag, CreatedDate: &t2}},
		makeCommit(5):  {{Name: "v3", Type: gitdomain.RefTypeTag, CreatedDate: &t2}},
		makeCommit(7):  {{Name: "main", Type: gitdomain.RefTypeBranch, IsDefaultBranch: true, CreatedDate: &t2}},
		makeCommit(12): {{Name: "feat2", Type: gitdomain.RefTypeBranch, CreatedDate: &t2}},
	}

	if err := store.CalculateVisibleUploads(context.Background(), 50, graph, refDescriptions, time.Hour, time.Hour, 0); err != nil {
		t.Fatalf("unexpected error while calculating visible uploads: %s", err)
	}

	expectedVisibleUploads := map[string][]int{
		makeCommit(1):  {1},
		makeCommit(2):  {1},
		makeCommit(3):  {2},
		makeCommit(4):  {2},
		makeCommit(5):  {2},
		makeCommit(6):  {3},
		makeCommit(7):  {3},
		makeCommit(8):  {4},
		makeCommit(9):  {4},
		makeCommit(10): {2},
		makeCommit(11): {5},
		makeCommit(12): {5},
	}
	if diff := cmp.Diff(expectedVisibleUploads, getVisibleUploads(t, db, 50, keysOf(expectedVisibleUploads))); diff != "" {
		t.Errorf("unexpected visible uploads (-want +got):\n%s", diff)
	}

	// Ensure data can be queried in reverse direction as well
	assertCommitsVisibleFromUploads(t, store, uploads, expectedVisibleUploads)

	if diff := cmp.Diff([]int{3}, getUploadsVisibleAtTip(t, db, 50)); diff != "" {
		t.Errorf("unexpected uploads visible at tip (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff([]int{2, 3, 5}, getProtectedUploads(t, db, 50)); diff != "" {
		t.Errorf("unexpected protected uploads (-want +got):\n%s", diff)
	}
}

func TestCalculateVisibleUploadsNonDefaultBranchesWithCustomRetentionConfiguration(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

	// This database has the following commit graph:
	//
	//                +-- [08] ----- {09} --+
	//                |                     |
	// [01] -- {02} --+-- [03] --+-- {04} --+-- {05} -- [06] -- {07}
	//                           |
	//                           +--- 10 ------ [11] -- {12}
	//
	// 02: tag v1
	// 04: tag v2
	// 05: tag v3
	// 07: tip of main branch
	// 09: tip of branch feat1
	// 12: tip of branch feat2

	uploads := []Upload{
		{ID: 1, Commit: makeCommit(1)},
		{ID: 2, Commit: makeCommit(3)},
		{ID: 3, Commit: makeCommit(6)},
		{ID: 4, Commit: makeCommit(8)},
		{ID: 5, Commit: makeCommit(11)},
	}
	insertUploads(t, db, uploads...)

	retentionConfigurationQuery := `
		INSERT INTO lsif_retention_configuration (
			id,
			repository_id,
			max_age_for_non_stale_branches_seconds,
			max_age_for_non_stale_tags_seconds
		) VALUES (
			1,
			50,
			3600,
			3600
		)
	`
	if _, err := db.Exec(retentionConfigurationQuery); err != nil {
		t.Fatalf("unexpected error inserting retention configuration: %s", err)
	}

	graph := gitdomain.ParseCommitGraph([]string{
		strings.Join([]string{makeCommit(12), makeCommit(11)}, " "),
		strings.Join([]string{makeCommit(11), makeCommit(10)}, " "),
		strings.Join([]string{makeCommit(10), makeCommit(3)}, " "),
		strings.Join([]string{makeCommit(7), makeCommit(6)}, " "),
		strings.Join([]string{makeCommit(6), makeCommit(5)}, " "),
		strings.Join([]string{makeCommit(5), makeCommit(4), makeCommit(9)}, " "),
		strings.Join([]string{makeCommit(9), makeCommit(8)}, " "),
		strings.Join([]string{makeCommit(8), makeCommit(2)}, " "),
		strings.Join([]string{makeCommit(4), makeCommit(3)}, " "),
		strings.Join([]string{makeCommit(3), makeCommit(2)}, " "),
		strings.Join([]string{makeCommit(2), makeCommit(1)}, " "),
		strings.Join([]string{makeCommit(1)}, " "),
	})

	t1 := time.Now().Add(-time.Minute * 90) // > 1 hr
	t2 := time.Now().Add(-time.Minute * 30) // < 1 hr

	refDescriptions := map[string][]gitdomain.RefDescription{
		// stale
		makeCommit(2): {{Name: "v1", Type: gitdomain.RefTypeTag, CreatedDate: &t1}},
		makeCommit(9): {{Name: "feat1", Type: gitdomain.RefTypeBranch, CreatedDate: &t1}},

		// fresh
		makeCommit(4):  {{Name: "v2", Type: gitdomain.RefTypeTag, CreatedDate: &t2}},
		makeCommit(5):  {{Name: "v3", Type: gitdomain.RefTypeTag, CreatedDate: &t2}},
		makeCommit(7):  {{Name: "main", Type: gitdomain.RefTypeBranch, IsDefaultBranch: true, CreatedDate: &t2}},
		makeCommit(12): {{Name: "feat2", Type: gitdomain.RefTypeBranch, CreatedDate: &t2}},
	}

	if err := store.CalculateVisibleUploads(context.Background(), 50, graph, refDescriptions, time.Second, time.Second, 0); err != nil {
		t.Fatalf("unexpected error while calculating visible uploads: %s", err)
	}

	expectedVisibleUploads := map[string][]int{
		makeCommit(1):  {1},
		makeCommit(2):  {1},
		makeCommit(3):  {2},
		makeCommit(4):  {2},
		makeCommit(5):  {2},
		makeCommit(6):  {3},
		makeCommit(7):  {3},
		makeCommit(8):  {4},
		makeCommit(9):  {4},
		makeCommit(10): {2},
		makeCommit(11): {5},
		makeCommit(12): {5},
	}
	if diff := cmp.Diff(expectedVisibleUploads, getVisibleUploads(t, db, 50, keysOf(expectedVisibleUploads))); diff != "" {
		t.Errorf("unexpected visible uploads (-want +got):\n%s", diff)
	}

	// Ensure data can be queried in reverse direction as well
	assertCommitsVisibleFromUploads(t, store, uploads, expectedVisibleUploads)

	if diff := cmp.Diff([]int{3}, getUploadsVisibleAtTip(t, db, 50)); diff != "" {
		t.Errorf("unexpected uploads visible at tip (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff([]int{2, 3, 5}, getProtectedUploads(t, db, 50)); diff != "" {
		t.Errorf("unexpected protected uploads (-want +got):\n%s", diff)
	}
}

func assertCommitsVisibleFromUploads(t *testing.T, store *Store, uploads []Upload, expectedVisibleUploads map[string][]int) {
	expectedVisibleCommits := map[int][]string{}
	for commit, uploadIDs := range expectedVisibleUploads {
		for _, uploadID := range uploadIDs {
			expectedVisibleCommits[uploadID] = append(expectedVisibleCommits[uploadID], commit)
		}
	}
	for _, commits := range expectedVisibleCommits {
		sort.Strings(commits)
	}

	// Test pagination by requesting only a couple of
	// results at a time in this assertion helper.
	testPageSize := 2

	for _, upload := range uploads {
		var token *string
		var allCommits []string

		for {
			commits, nextToken, err := store.CommitsVisibleToUpload(context.Background(), upload.ID, testPageSize, token)
			if err != nil {
				t.Fatalf("unexpected error getting commits visible to upload %d: %s", upload.ID, err)
			}
			if nextToken == nil {
				break
			}

			allCommits = append(allCommits, commits...)
			token = nextToken
		}

		if diff := cmp.Diff(expectedVisibleCommits[upload.ID], allCommits); diff != "" {
			t.Errorf("unexpected commits visible to upload %d (-want +got):\n%s", upload.ID, diff)
		}
	}
}

func keysOf(m map[string][]int) (keys []string) {
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

//
// Benchmarks
//

func BenchmarkCalculateVisibleUploads(b *testing.B) {
	db := dbtest.NewDB(b)
	store := testStore(db)

	graph, err := readBenchmarkCommitGraph()
	if err != nil {
		b.Fatalf("unexpected error reading benchmark commit graph: %s", err)
	}

	refDescriptions := map[string][]gitdomain.RefDescription{
		makeCommit(3): {{IsDefaultBranch: true}},
	}

	uploads, err := readBenchmarkCommitGraphView()
	if err != nil {
		b.Fatalf("unexpected error reading benchmark uploads: %s", err)
	}
	insertUploads(b, db, uploads...)

	b.ResetTimer()
	b.ReportAllocs()

	if err := store.CalculateVisibleUploads(context.Background(), 50, graph, refDescriptions, time.Hour, time.Hour, 0); err != nil {
		b.Fatalf("unexpected error while calculating visible uploads: %s", err)
	}
}

func readBenchmarkCommitGraph() (*gitdomain.CommitGraph, error) {
	contents, err := readBenchmarkFile("../../commitgraph/testdata/commits.txt.gz")
	if err != nil {
		return nil, err
	}

	return gitdomain.ParseCommitGraph(strings.Split(string(contents), "\n")), nil
}

func readBenchmarkCommitGraphView() ([]Upload, error) {
	contents, err := readBenchmarkFile("../../commitgraph/testdata/uploads.txt.gz")
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(bytes.NewReader(contents))

	var uploads []Upload
	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		id, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, err
		}

		uploads = append(uploads, Upload{
			ID:           id,
			RepositoryID: 50,
			Commit:       record[1],
			Root:         record[2],
		})
	}

	return uploads, nil
}

func readBenchmarkFile(path string) ([]byte, error) {
	uploadsFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer uploadsFile.Close()

	r, err := gzip.NewReader(uploadsFile)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	contents, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return contents, nil
}
