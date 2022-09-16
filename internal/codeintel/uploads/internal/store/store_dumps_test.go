package store

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/commitgraph"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGetDumpsByIDs(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	// Dumps do not exist initially
	if dumps, err := store.GetDumpsByIDs(context.Background(), []int{1, 2}); err != nil {
		t.Fatalf("unexpected error getting dump: %s", err)
	} else if len(dumps) > 0 {
		t.Fatal("unexpected record")
	}

	uploadedAt := time.Unix(1587396557, 0).UTC()
	startedAt := uploadedAt.Add(time.Minute)
	finishedAt := uploadedAt.Add(time.Minute * 2)
	expectedAssociatedIndexID := 42
	expected1 := shared.Dump{
		ID:                1,
		Commit:            makeCommit(1),
		Root:              "sub/",
		VisibleAtTip:      true,
		UploadedAt:        uploadedAt,
		State:             "completed",
		FailureMessage:    nil,
		StartedAt:         &startedAt,
		FinishedAt:        &finishedAt,
		RepositoryID:      50,
		RepositoryName:    "n-50",
		Indexer:           "lsif-go",
		IndexerVersion:    "latest",
		AssociatedIndexID: &expectedAssociatedIndexID,
	}
	expected2 := shared.Dump{
		ID:                2,
		Commit:            makeCommit(2),
		Root:              "other/",
		VisibleAtTip:      false,
		UploadedAt:        uploadedAt,
		State:             "completed",
		FailureMessage:    nil,
		StartedAt:         &startedAt,
		FinishedAt:        &finishedAt,
		RepositoryID:      50,
		RepositoryName:    "n-50",
		Indexer:           "scip-typescript",
		IndexerVersion:    "1.2.3",
		AssociatedIndexID: nil,
	}

	insertUploads(t, db, dumpToUpload(expected1), dumpToUpload(expected2))
	insertVisibleAtTip(t, db, 50, 1)

	if dumps, err := store.GetDumpsByIDs(context.Background(), []int{1}); err != nil {
		t.Fatalf("unexpected error getting dump: %s", err)
	} else if len(dumps) != 1 {
		t.Fatal("expected one record")
	} else if diff := cmp.Diff(expected1, dumps[0]); diff != "" {
		t.Errorf("unexpected dump (-want +got):\n%s", diff)
	}

	if dumps, err := store.GetDumpsByIDs(context.Background(), []int{1, 2}); err != nil {
		t.Fatalf("unexpected error getting dump: %s", err)
	} else if len(dumps) != 2 {
		t.Fatal("expected two records")
	} else if diff := cmp.Diff(expected1, dumps[0]); diff != "" {
		t.Errorf("unexpected dump (-want +got):\n%s", diff)
	} else if diff := cmp.Diff(expected2, dumps[1]); diff != "" {
		t.Errorf("unexpected dump (-want +got):\n%s", diff)
	}
}

func TestFindClosestDumps(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	// This database has the following commit graph:
	//
	// [1] --+--- 2 --------+--5 -- 6 --+-- [7]
	//       |              |           |
	//       +-- [3] -- 4 --+           +--- 8

	uploads := []shared.Upload{
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

	visibleUploads, links := commitgraph.NewGraph(graph, toCommitGraphView(uploads)).Gather()

	expectedVisibleUploads := map[string][]commitgraph.UploadMeta{
		makeCommit(1): {{UploadID: 1, Distance: 0}},
		makeCommit(2): {{UploadID: 1, Distance: 1}},
		makeCommit(3): {{UploadID: 2, Distance: 0}},
		makeCommit(4): {{UploadID: 2, Distance: 1}},
		makeCommit(5): {{UploadID: 1, Distance: 2}},
		makeCommit(6): {{UploadID: 1, Distance: 3}},
		makeCommit(7): {{UploadID: 3, Distance: 0}},
		makeCommit(8): {{UploadID: 1, Distance: 4}},
	}
	if diff := cmp.Diff(expectedVisibleUploads, normalizeVisibleUploads(visibleUploads)); diff != "" {
		t.Errorf("unexpected visible uploads (-want +got):\n%s", diff)
	}
	expectedLinks := map[string]commitgraph.LinkRelationship{}
	if diff := cmp.Diff(expectedLinks, links); diff != "" {
		t.Errorf("unexpected visible links (-want +got):\n%s", diff)
	}

	// Prep
	insertNearestUploads(t, db, 50, visibleUploads)
	insertLinks(t, db, 50, links)

	// Test
	testFindClosestDumps(t, store, []FindClosestDumpsTestCase{
		{commit: makeCommit(1), file: "file.ts", rootMustEnclosePath: true, graph: graph, anyOfIDs: []int{1}},
		{commit: makeCommit(2), file: "file.ts", rootMustEnclosePath: true, graph: graph, anyOfIDs: []int{1}},
		{commit: makeCommit(3), file: "file.ts", rootMustEnclosePath: true, graph: graph, anyOfIDs: []int{2}},
		{commit: makeCommit(4), file: "file.ts", rootMustEnclosePath: true, graph: graph, anyOfIDs: []int{2}},
		{commit: makeCommit(6), file: "file.ts", rootMustEnclosePath: true, graph: graph, anyOfIDs: []int{1}},
		{commit: makeCommit(7), file: "file.ts", rootMustEnclosePath: true, graph: graph, anyOfIDs: []int{3}},
		{commit: makeCommit(5), file: "file.ts", rootMustEnclosePath: true, graph: graph, anyOfIDs: []int{1, 2, 3}},
		{commit: makeCommit(8), file: "file.ts", rootMustEnclosePath: true, graph: graph, anyOfIDs: []int{1, 2}},
	})
}

func TestFindClosestDumpsAlternateCommitGraph(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	// This database has the following commit graph:
	//
	// 1 --+-- [2] ---- 3
	//     |
	//     +--- 4 --+-- 5 -- 6
	//              |
	//              +-- 7 -- 8

	uploads := []shared.Upload{
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

	visibleUploads, links := commitgraph.NewGraph(graph, toCommitGraphView(uploads)).Gather()

	expectedVisibleUploads := map[string][]commitgraph.UploadMeta{
		makeCommit(2): {{UploadID: 1, Distance: 0}},
		makeCommit(3): {{UploadID: 1, Distance: 1}},
	}
	if diff := cmp.Diff(expectedVisibleUploads, normalizeVisibleUploads(visibleUploads)); diff != "" {
		t.Errorf("unexpected visible uploads (-want +got):\n%s", diff)
	}

	expectedLinks := map[string]commitgraph.LinkRelationship{}
	if diff := cmp.Diff(expectedLinks, links); diff != "" {
		t.Errorf("unexpected visible links (-want +got):\n%s", diff)
	}

	// Prep
	insertNearestUploads(t, db, 50, visibleUploads)
	insertLinks(t, db, 50, links)

	// Test
	testFindClosestDumps(t, store, []FindClosestDumpsTestCase{
		{commit: makeCommit(2), graph: graph, allOfIDs: []int{1}},
		{commit: makeCommit(3), graph: graph, allOfIDs: []int{1}},
		{commit: makeCommit(4), graph: graph},
		{commit: makeCommit(6), graph: graph},
		{commit: makeCommit(7), graph: graph},
		{commit: makeCommit(5), graph: graph},
		{commit: makeCommit(8), graph: graph},
	})
}

func TestFindClosestDumpsAlternateCommitGraphWithOverwrittenVisibleUploads(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	// This database has the following commit graph:
	//
	// 1 -- [2] -- 3 -- 4 -- [5]

	uploads := []shared.Upload{
		{ID: 1, Commit: makeCommit(2)},
		{ID: 2, Commit: makeCommit(5)},
	}
	insertUploads(t, db, uploads...)

	graph := gitdomain.ParseCommitGraph([]string{
		strings.Join([]string{makeCommit(5), makeCommit(4)}, " "),
		strings.Join([]string{makeCommit(4), makeCommit(3)}, " "),
		strings.Join([]string{makeCommit(3), makeCommit(2)}, " "),
		strings.Join([]string{makeCommit(2), makeCommit(1)}, " "),
		strings.Join([]string{makeCommit(1)}, " "),
	})

	visibleUploads, links := commitgraph.NewGraph(graph, toCommitGraphView(uploads)).Gather()

	expectedVisibleUploads := map[string][]commitgraph.UploadMeta{
		makeCommit(2): {{UploadID: 1, Distance: 0}},
		makeCommit(3): {{UploadID: 1, Distance: 1}},
		makeCommit(4): {{UploadID: 1, Distance: 2}},
		makeCommit(5): {{UploadID: 2, Distance: 0}},
	}
	if diff := cmp.Diff(expectedVisibleUploads, normalizeVisibleUploads(visibleUploads)); diff != "" {
		t.Errorf("unexpected visible uploads (-want +got):\n%s", diff)
	}

	expectedLinks := map[string]commitgraph.LinkRelationship{}
	if diff := cmp.Diff(expectedLinks, links); diff != "" {
		t.Errorf("unexpected visible links (-want +got):\n%s", diff)
	}

	// Prep
	insertNearestUploads(t, db, 50, visibleUploads)
	insertLinks(t, db, 50, links)

	// Test
	testFindClosestDumps(t, store, []FindClosestDumpsTestCase{
		{commit: makeCommit(2), graph: graph, allOfIDs: []int{1}},
		{commit: makeCommit(3), graph: graph, allOfIDs: []int{1}},
		{commit: makeCommit(4), graph: graph, allOfIDs: []int{1}},
		{commit: makeCommit(5), graph: graph, allOfIDs: []int{2}},
	})
}

func TestFindClosestDumpsDistinctRoots(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	// This database has the following commit graph:
	//
	// [1] -- 2

	uploads := []shared.Upload{
		{ID: 1, Commit: makeCommit(1), Root: "root1/"},
		{ID: 2, Commit: makeCommit(1), Root: "root2/"},
	}
	insertUploads(t, db, uploads...)

	graph := gitdomain.ParseCommitGraph([]string{
		strings.Join([]string{makeCommit(2), makeCommit(1)}, " "),
		strings.Join([]string{makeCommit(1)}, " "),
	})

	visibleUploads, links := commitgraph.NewGraph(graph, toCommitGraphView(uploads)).Gather()

	expectedVisibleUploads := map[string][]commitgraph.UploadMeta{
		makeCommit(1): {{UploadID: 1, Distance: 0}, {UploadID: 2, Distance: 0}},
	}
	if diff := cmp.Diff(expectedVisibleUploads, normalizeVisibleUploads(visibleUploads)); diff != "" {
		t.Errorf("unexpected visible uploads (-want +got):\n%s", diff)
	}

	expectedLinks := map[string]commitgraph.LinkRelationship{
		makeCommit(2): {Commit: makeCommit(2), AncestorCommit: makeCommit(1), Distance: 1},
	}
	if diff := cmp.Diff(expectedLinks, links); diff != "" {
		t.Errorf("unexpected visible links (-want +got):\n%s", diff)
	}

	// Prep
	insertNearestUploads(t, db, 50, visibleUploads)
	insertLinks(t, db, 50, links)

	// Test
	testFindClosestDumps(t, store, []FindClosestDumpsTestCase{
		{commit: makeCommit(1), file: "blah", rootMustEnclosePath: true, graph: graph},
		{commit: makeCommit(2), file: "root1/file.ts", rootMustEnclosePath: true, graph: graph, allOfIDs: []int{1}},
		{commit: makeCommit(1), file: "root2/file.ts", rootMustEnclosePath: true, graph: graph, allOfIDs: []int{2}},
		{commit: makeCommit(2), file: "root2/file.ts", rootMustEnclosePath: true, graph: graph, allOfIDs: []int{2}},
		{commit: makeCommit(1), file: "root3/file.ts", rootMustEnclosePath: true, graph: graph},
	})
}

func TestFindClosestDumpsOverlappingRoots(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

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
	// | 2        | 1      | root4/  | scip-python |
	// | 3        | 2      | root1/  | lsif-go |
	// | 4        | 2      | root2/  | lsif-go |
	// | 5        | 2      |         | scip-python | (overwrites root4/ at commit 1)
	// | 6        | 3      | root1/  | lsif-go | (overwrites root1/ at commit 2)
	// | 7        | 4      |         | scip-python | (overwrites (root) at commit 2)
	// | 8        | 5      | root2/  | lsif-go | (overwrites root2/ at commit 2)
	// | 9        | 6      | root1/  | lsif-go | (overwrites root1/ at commit 2)

	uploads := []shared.Upload{
		{ID: 1, Commit: makeCommit(1), Indexer: "lsif-go", Root: "root3/"},
		{ID: 2, Commit: makeCommit(1), Indexer: "scip-python", Root: "root4/"},
		{ID: 3, Commit: makeCommit(2), Indexer: "lsif-go", Root: "root1/"},
		{ID: 4, Commit: makeCommit(2), Indexer: "lsif-go", Root: "root2/"},
		{ID: 5, Commit: makeCommit(2), Indexer: "scip-python", Root: ""},
		{ID: 6, Commit: makeCommit(3), Indexer: "lsif-go", Root: "root1/"},
		{ID: 7, Commit: makeCommit(4), Indexer: "scip-python", Root: ""},
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

	visibleUploads, links := commitgraph.NewGraph(graph, toCommitGraphView(uploads)).Gather()

	expectedVisibleUploads := map[string][]commitgraph.UploadMeta{
		makeCommit(1): {{UploadID: 1, Distance: 0}, {UploadID: 2, Distance: 0}},
		makeCommit(2): {{UploadID: 1, Distance: 1}, {UploadID: 2, Distance: 1}, {UploadID: 3, Distance: 0}, {UploadID: 4, Distance: 0}, {UploadID: 5, Distance: 0}},
		makeCommit(3): {{UploadID: 1, Distance: 2}, {UploadID: 2, Distance: 2}, {UploadID: 4, Distance: 1}, {UploadID: 5, Distance: 1}, {UploadID: 6, Distance: 0}},
		makeCommit(4): {{UploadID: 1, Distance: 2}, {UploadID: 2, Distance: 2}, {UploadID: 3, Distance: 1}, {UploadID: 4, Distance: 1}, {UploadID: 7, Distance: 0}},
		makeCommit(5): {{UploadID: 1, Distance: 3}, {UploadID: 2, Distance: 3}, {UploadID: 6, Distance: 1}, {UploadID: 7, Distance: 1}, {UploadID: 8, Distance: 0}},
		makeCommit(6): {{UploadID: 1, Distance: 4}, {UploadID: 2, Distance: 4}, {UploadID: 7, Distance: 2}, {UploadID: 8, Distance: 1}, {UploadID: 9, Distance: 0}},
	}
	if diff := cmp.Diff(expectedVisibleUploads, normalizeVisibleUploads(visibleUploads)); diff != "" {
		t.Errorf("unexpected visible uploads (-want +got):\n%s", diff)
	}

	expectedLinks := map[string]commitgraph.LinkRelationship{}
	if diff := cmp.Diff(expectedLinks, links); diff != "" {
		t.Errorf("unexpected visible links (-want +got):\n%s", diff)
	}

	// Prep
	insertNearestUploads(t, db, 50, visibleUploads)
	insertLinks(t, db, 50, links)

	// Test
	testFindClosestDumps(t, store, []FindClosestDumpsTestCase{
		{commit: makeCommit(4), file: "root1/file.ts", rootMustEnclosePath: true, graph: graph, allOfIDs: []int{7, 3}},
		{commit: makeCommit(5), file: "root2/file.ts", rootMustEnclosePath: true, graph: graph, allOfIDs: []int{8, 7}},
		{commit: makeCommit(3), file: "root3/file.ts", rootMustEnclosePath: true, graph: graph, allOfIDs: []int{5, 1}},
		{commit: makeCommit(1), file: "root4/file.ts", rootMustEnclosePath: true, graph: graph, allOfIDs: []int{2}},
		{commit: makeCommit(2), file: "root4/file.ts", rootMustEnclosePath: true, graph: graph, allOfIDs: []int{2, 5}},
	})
}

func TestFindClosestDumpsIndexerName(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	// This database has the following commit graph:
	//
	// [1] --+-- [2] --+-- [3] --+-- [4] --+-- 5

	uploads := []shared.Upload{
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

	visibleUploads, links := commitgraph.NewGraph(graph, toCommitGraphView(uploads)).Gather()

	expectedVisibleUploads := map[string][]commitgraph.UploadMeta{
		makeCommit(1): {
			{UploadID: 1, Distance: 0},
			{UploadID: 5, Distance: 0},
		},
		makeCommit(2): {
			{UploadID: 1, Distance: 1},
			{UploadID: 2, Distance: 0},
			{UploadID: 5, Distance: 1},
			{UploadID: 6, Distance: 0},
		},
		makeCommit(3): {
			{UploadID: 1, Distance: 2},
			{UploadID: 2, Distance: 1},
			{UploadID: 3, Distance: 0},
			{UploadID: 5, Distance: 2},
			{UploadID: 6, Distance: 1},
			{UploadID: 7, Distance: 0},
		},
		makeCommit(4): {
			{UploadID: 1, Distance: 3},
			{UploadID: 2, Distance: 2},
			{UploadID: 3, Distance: 1},
			{UploadID: 4, Distance: 0},
			{UploadID: 5, Distance: 3},
			{UploadID: 6, Distance: 2},
			{UploadID: 7, Distance: 1},
			{UploadID: 8, Distance: 0},
		},
	}
	if diff := cmp.Diff(expectedVisibleUploads, normalizeVisibleUploads(visibleUploads)); diff != "" {
		t.Errorf("unexpected visible uploads (-want +got):\n%s", diff)
	}

	expectedLinks := map[string]commitgraph.LinkRelationship{
		makeCommit(5): {Commit: makeCommit(5), AncestorCommit: makeCommit(4), Distance: 1},
	}
	if diff := cmp.Diff(expectedLinks, links); diff != "" {
		t.Errorf("unexpected visible links (-want +got):\n%s", diff)
	}

	// Prep
	insertNearestUploads(t, db, 50, visibleUploads)
	insertLinks(t, db, 50, links)

	// Test
	testFindClosestDumps(t, store, []FindClosestDumpsTestCase{
		{commit: makeCommit(5), file: "root1/file.ts", indexer: "idx1", graph: graph, allOfIDs: []int{1}},
		{commit: makeCommit(5), file: "root2/file.ts", indexer: "idx1", graph: graph, allOfIDs: []int{2}},
		{commit: makeCommit(5), file: "root3/file.ts", indexer: "idx1", graph: graph, allOfIDs: []int{3}},
		{commit: makeCommit(5), file: "root4/file.ts", indexer: "idx1", graph: graph, allOfIDs: []int{4}},
		{commit: makeCommit(5), file: "root1/file.ts", indexer: "idx2", graph: graph, allOfIDs: []int{5}},
		{commit: makeCommit(5), file: "root2/file.ts", indexer: "idx2", graph: graph, allOfIDs: []int{6}},
		{commit: makeCommit(5), file: "root3/file.ts", indexer: "idx2", graph: graph, allOfIDs: []int{7}},
		{commit: makeCommit(5), file: "root4/file.ts", indexer: "idx2", graph: graph, allOfIDs: []int{8}},
	})
}

func TestFindClosestDumpsIntersectingPath(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	// This database has the following commit graph:
	//
	// [1]

	uploads := []shared.Upload{
		{ID: 1, Commit: makeCommit(1), Root: "web/src/", Indexer: "lsif-eslint"},
	}
	insertUploads(t, db, uploads...)

	graph := gitdomain.ParseCommitGraph([]string{
		strings.Join([]string{makeCommit(1)}, " "),
	})

	visibleUploads, links := commitgraph.NewGraph(graph, toCommitGraphView(uploads)).Gather()

	expectedVisibleUploads := map[string][]commitgraph.UploadMeta{
		makeCommit(1): {{UploadID: 1}},
	}
	if diff := cmp.Diff(expectedVisibleUploads, normalizeVisibleUploads(visibleUploads)); diff != "" {
		t.Errorf("unexpected visible uploads (-want +got):\n%s", diff)
	}

	expectedLinks := map[string]commitgraph.LinkRelationship{}
	if diff := cmp.Diff(expectedLinks, links); diff != "" {
		t.Errorf("unexpected visible links (-want +got):\n%s", diff)
	}

	// Prep
	insertNearestUploads(t, db, 50, visibleUploads)
	insertLinks(t, db, 50, links)

	// Test
	testFindClosestDumps(t, store, []FindClosestDumpsTestCase{
		{commit: makeCommit(1), file: "", rootMustEnclosePath: false, graph: graph, allOfIDs: []int{1}},
		{commit: makeCommit(1), file: "web/", rootMustEnclosePath: false, graph: graph, allOfIDs: []int{1}},
		{commit: makeCommit(1), file: "web/src/file.ts", rootMustEnclosePath: false, graph: graph, allOfIDs: []int{1}},
	})
}

func TestFindClosestDumpsFromGraphFragment(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	// This database has the following commit graph:
	//
	//       <- known commits || new commits ->
	//                        ||
	// [1] --+--- 2 --- 3 --  || -- 4 --+-- 7
	//       |                ||       /
	//       +-- [5] -- 6 --- || -----+

	uploads := []shared.Upload{
		{ID: 1, Commit: makeCommit(1)},
		{ID: 2, Commit: makeCommit(5)},
	}
	insertUploads(t, db, uploads...)

	currentGraph := gitdomain.ParseCommitGraph([]string{
		strings.Join([]string{makeCommit(6), makeCommit(5)}, " "),
		strings.Join([]string{makeCommit(5), makeCommit(1)}, " "),
		strings.Join([]string{makeCommit(3), makeCommit(2)}, " "),
		strings.Join([]string{makeCommit(2), makeCommit(1)}, " "),
		strings.Join([]string{makeCommit(1)}, " "),
	})

	visibleUploads, links := commitgraph.NewGraph(currentGraph, toCommitGraphView(uploads)).Gather()

	expectedVisibleUploads := map[string][]commitgraph.UploadMeta{
		makeCommit(1): {{UploadID: 1, Distance: 0}},
		makeCommit(2): {{UploadID: 1, Distance: 1}},
		makeCommit(3): {{UploadID: 1, Distance: 2}},
		makeCommit(5): {{UploadID: 2, Distance: 0}},
		makeCommit(6): {{UploadID: 2, Distance: 1}},
	}
	if diff := cmp.Diff(expectedVisibleUploads, normalizeVisibleUploads(visibleUploads)); diff != "" {
		t.Errorf("unexpected visible uploads (-want +got):\n%s", diff)
	}

	expectedLinks := map[string]commitgraph.LinkRelationship{}
	if diff := cmp.Diff(expectedLinks, links); diff != "" {
		t.Errorf("unexpected visible links (-want +got):\n%s", diff)
	}

	// Prep
	insertNearestUploads(t, db, 50, visibleUploads)
	insertLinks(t, db, 50, links)

	// Test
	graphFragment := gitdomain.ParseCommitGraph([]string{
		strings.Join([]string{makeCommit(7), makeCommit(4), makeCommit(6)}, " "),
		strings.Join([]string{makeCommit(4), makeCommit(3)}, " "),
		strings.Join([]string{makeCommit(6)}, " "),
		strings.Join([]string{makeCommit(3)}, " "),
	})

	testFindClosestDumps(t, store, []FindClosestDumpsTestCase{
		// Note: Can't query anything outside of the graph fragment
		{commit: makeCommit(3), file: "file.ts", rootMustEnclosePath: true, graph: graphFragment, anyOfIDs: []int{1}},
		{commit: makeCommit(6), file: "file.ts", rootMustEnclosePath: true, graph: graphFragment, anyOfIDs: []int{2}},
		{commit: makeCommit(4), file: "file.ts", rootMustEnclosePath: true, graph: graphFragment, graphFragmentOnly: true, anyOfIDs: []int{1}},
		{commit: makeCommit(7), file: "file.ts", rootMustEnclosePath: true, graph: graphFragment, graphFragmentOnly: true, anyOfIDs: []int{2}},
	})
}

func TestDefinitionDumps(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	moniker1 := precise.QualifiedMonikerData{
		MonikerData: precise.MonikerData{
			Scheme: "gomod",
		},
		PackageInformationData: precise.PackageInformationData{
			Name:    "leftpad",
			Version: "0.1.0",
		},
	}

	moniker2 := precise.QualifiedMonikerData{
		MonikerData: precise.MonikerData{
			Scheme: "npm",
		},
		PackageInformationData: precise.PackageInformationData{
			Name:    "rightpad",
			Version: "0.2.0",
		},
	}

	// Package does not exist initially
	if dumps, err := store.GetDumpsWithDefinitionsForMonikers(context.Background(), []precise.QualifiedMonikerData{moniker1}); err != nil {
		t.Fatalf("unexpected error getting package: %s", err)
	} else if len(dumps) != 0 {
		t.Fatal("unexpected record")
	}

	uploadedAt := time.Unix(1587396557, 0).UTC()
	startedAt := uploadedAt.Add(time.Minute)
	finishedAt := uploadedAt.Add(time.Minute * 2)
	expected1 := shared.Dump{
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
		IndexerVersion: "latest",
	}
	expected2 := shared.Dump{
		ID:                2,
		Commit:            makeCommit(2),
		Root:              "other/",
		VisibleAtTip:      false,
		UploadedAt:        uploadedAt,
		State:             "completed",
		FailureMessage:    nil,
		StartedAt:         &startedAt,
		FinishedAt:        &finishedAt,
		RepositoryID:      50,
		RepositoryName:    "n-50",
		Indexer:           "scip-typescript",
		IndexerVersion:    "1.2.3",
		AssociatedIndexID: nil,
	}
	expected3 := shared.Dump{
		ID:             3,
		Commit:         makeCommit(3),
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
		IndexerVersion: "latest",
	}

	insertUploads(t, db, dumpToUpload(expected1), dumpToUpload(expected2), dumpToUpload(expected3))
	insertVisibleAtTip(t, db, 50, 1)

	if err := store.UpdatePackages(context.Background(), 1, []precise.Package{
		{Scheme: "gomod", Name: "leftpad", Version: "0.1.0"},
	}); err != nil {
		t.Fatalf("unexpected error updating packages: %s", err)
	}

	if err := store.UpdatePackages(context.Background(), 2, []precise.Package{
		{Scheme: "npm", Name: "rightpad", Version: "0.2.0"},
	}); err != nil {
		t.Fatalf("unexpected error updating packages: %s", err)
	}

	// Duplicate package
	if err := store.UpdatePackages(context.Background(), 3, []precise.Package{
		{Scheme: "gomod", Name: "leftpad", Version: "0.1.0"},
	}); err != nil {
		t.Fatalf("unexpected error updating packages: %s", err)
	}

	if dumps, err := store.GetDumpsWithDefinitionsForMonikers(context.Background(), []precise.QualifiedMonikerData{moniker1}); err != nil {
		t.Fatalf("unexpected error getting package: %s", err)
	} else if len(dumps) != 1 {
		t.Fatal("expected one record")
	} else if diff := cmp.Diff(expected1, dumps[0]); diff != "" {
		t.Errorf("unexpected dump (-want +got):\n%s", diff)
	}

	if dumps, err := store.GetDumpsWithDefinitionsForMonikers(context.Background(), []precise.QualifiedMonikerData{moniker1, moniker2}); err != nil {
		t.Fatalf("unexpected error getting package: %s", err)
	} else if len(dumps) != 2 {
		t.Fatal("expected two records")
	} else if diff := cmp.Diff(expected1, dumps[0]); diff != "" {
		t.Errorf("unexpected dump (-want +got):\n%s", diff)
	} else if diff := cmp.Diff(expected2, dumps[1]); diff != "" {
		t.Errorf("unexpected dump (-want +got):\n%s", diff)
	}

	t.Run("enforce repository permissions", func(t *testing.T) {
		// Enable permissions user mapping forces checking repository permissions
		// against permissions tables in the database, which should effectively block
		// all access because permissions tables are empty.
		before := globals.PermissionsUserMapping()
		globals.SetPermissionsUserMapping(&schema.PermissionsUserMapping{Enabled: true})
		defer globals.SetPermissionsUserMapping(before)

		if dumps, err := store.GetDumpsWithDefinitionsForMonikers(context.Background(), []precise.QualifiedMonikerData{moniker1, moniker2}); err != nil {
			t.Fatalf("unexpected error getting package: %s", err)
		} else if len(dumps) != 0 {
			t.Errorf("unexpected count. want=%d have=%d", 0, len(dumps))
		}
	})
}

func dumpToUpload(expected shared.Dump) shared.Upload {
	return shared.Upload{
		ID:                expected.ID,
		Commit:            expected.Commit,
		Root:              expected.Root,
		UploadedAt:        expected.UploadedAt,
		State:             expected.State,
		FailureMessage:    expected.FailureMessage,
		StartedAt:         expected.StartedAt,
		FinishedAt:        expected.FinishedAt,
		ProcessAfter:      expected.ProcessAfter,
		NumResets:         expected.NumResets,
		NumFailures:       expected.NumFailures,
		RepositoryID:      expected.RepositoryID,
		RepositoryName:    expected.RepositoryName,
		Indexer:           expected.Indexer,
		IndexerVersion:    expected.IndexerVersion,
		AssociatedIndexID: expected.AssociatedIndexID,
	}
}

func toCommitGraphView(uploads []shared.Upload) *commitgraph.CommitGraphView {
	commitGraphView := commitgraph.NewCommitGraphView()
	for _, upload := range uploads {
		commitGraphView.Add(commitgraph.UploadMeta{UploadID: upload.ID}, upload.Commit, fmt.Sprintf("%s:%s", upload.Root, upload.Indexer))
	}

	return commitGraphView
}

func normalizeVisibleUploads(uploadMetas map[string][]commitgraph.UploadMeta) map[string][]commitgraph.UploadMeta {
	for _, uploads := range uploadMetas {
		sort.Slice(uploads, func(i, j int) bool {
			return uploads[i].UploadID-uploads[j].UploadID < 0
		})
	}

	return uploadMetas
}

// insertNearestUploads populates the lsif_nearest_uploads table with the given upload metadata.
func insertNearestUploads(t testing.TB, db database.DB, repositoryID int, uploads map[string][]commitgraph.UploadMeta) {
	var rows []*sqlf.Query
	for commit, uploadMetas := range uploads {
		uploadsByLength := make(map[int]int, len(uploadMetas))
		for _, uploadMeta := range uploadMetas {
			uploadsByLength[uploadMeta.UploadID] = int(uploadMeta.Distance)
		}

		serializedUploadMetas, err := json.Marshal(uploadsByLength)
		if err != nil {
			t.Fatalf("unexpected error marshalling uploads: %s", err)
		}

		rows = append(rows, sqlf.Sprintf(
			"(%s, %s, %s)",
			repositoryID,
			dbutil.CommitBytea(commit),
			serializedUploadMetas,
		))
	}

	query := sqlf.Sprintf(
		`INSERT INTO lsif_nearest_uploads (repository_id, commit_bytea, uploads) VALUES %s`,
		sqlf.Join(rows, ","),
	)
	if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error while updating commit graph: %s", err)
	}
}

//nolint:unparam // unparam complains that `repositoryID` always has same value across call-sites, but that's OK
func insertLinks(t testing.TB, db database.DB, repositoryID int, links map[string]commitgraph.LinkRelationship) {
	if len(links) == 0 {
		return
	}

	var rows []*sqlf.Query
	for commit, link := range links {
		rows = append(rows, sqlf.Sprintf(
			"(%s, %s, %s, %s)",
			repositoryID,
			dbutil.CommitBytea(commit),
			dbutil.CommitBytea(link.AncestorCommit),
			link.Distance,
		))
	}

	query := sqlf.Sprintf(
		`INSERT INTO lsif_nearest_uploads_links (repository_id, commit_bytea, ancestor_commit_bytea, distance) VALUES %s`,
		sqlf.Join(rows, ","),
	)
	if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error while updating links: %s %s", err, query.Query(sqlf.PostgresBindVar))
	}
}

type FindClosestDumpsTestCase struct {
	commit              string
	file                string
	rootMustEnclosePath bool
	indexer             string
	graph               *gitdomain.CommitGraph
	graphFragmentOnly   bool
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

		assertDumpIDs := func(t *testing.T, dumps []shared.Dump) {
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
		}

		if !testCase.graphFragmentOnly {
			t.Run(name, func(t *testing.T) {
				dumps, err := store.FindClosestDumps(context.Background(), 50, testCase.commit, testCase.file, testCase.rootMustEnclosePath, testCase.indexer)
				if err != nil {
					t.Fatalf("unexpected error finding closest dumps: %s", err)
				}

				assertDumpIDs(t, dumps)
			})
		}

		if testCase.graph != nil {
			t.Run(name+" [graph-fragment]", func(t *testing.T) {
				dumps, err := store.FindClosestDumpsFromGraphFragment(context.Background(), 50, testCase.commit, testCase.file, testCase.rootMustEnclosePath, testCase.indexer, testCase.graph)
				if err != nil {
					t.Fatalf("unexpected error finding closest dumps: %s", err)
				}

				assertDumpIDs(t, dumps)
			})
		}
	}
}

func testAnyOf(t *testing.T, dumps []shared.Dump, expectedIDs []int) {
	if len(dumps) != 1 {
		t.Errorf("unexpected nearest dump length. want=%d have=%d", 1, len(dumps))
		return
	}

	if !testPresence(dumps[0].ID, expectedIDs) {
		t.Errorf("unexpected nearest dump ids. want one of %v have=%v", expectedIDs, dumps[0].ID)
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

func testAllOf(t *testing.T, dumps []shared.Dump, expectedIDs []int) {
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
