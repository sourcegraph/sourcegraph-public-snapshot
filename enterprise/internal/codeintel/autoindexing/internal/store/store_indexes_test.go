package store

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestInsertIndexes(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	insertRepo(t, db, 50, "")

	indexes, err := store.InsertIndexes(ctx, []types.Index{
		{
			State:        "queued",
			Commit:       makeCommit(1),
			RepositoryID: 50,
			DockerSteps: []types.DockerStep{
				{
					Image:    "cimg/node:12.16",
					Commands: []string{"yarn install --frozen-lockfile --no-progress"},
				},
			},
			LocalSteps:  []string{"echo hello"},
			Root:        "/foo/bar",
			Indexer:     "sourcegraph/scip-typescript:latest",
			IndexerArgs: []string{"index", "--yarn-workspaces"},
			Outfile:     "dump.lsif",
			ExecutionLogs: []executor.ExecutionLogEntry{
				{Command: []string{"op", "1"}, Out: "Indexing\nUploading\nDone with 1.\n"},
				{Command: []string{"op", "2"}, Out: "Indexing\nUploading\nDone with 2.\n"},
			},
		},
		{
			State:        "queued",
			Commit:       makeCommit(2),
			RepositoryID: 50,
			DockerSteps: []types.DockerStep{
				{
					Image:    "cimg/rust:nightly",
					Commands: []string{"cargo install"},
				},
			},
			LocalSteps:  nil,
			Root:        "/baz",
			Indexer:     "sourcegraph/lsif-rust:15",
			IndexerArgs: []string{"-v"},
			Outfile:     "dump.lsif",
			ExecutionLogs: []executor.ExecutionLogEntry{
				{Command: []string{"op", "1"}, Out: "Done with 1.\n"},
				{Command: []string{"op", "2"}, Out: "Done with 2.\n"},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error enqueueing index: %s", err)
	}
	if len(indexes) == 0 {
		t.Fatalf("expected records to be inserted")
	}

	rank1 := 1
	rank2 := 2
	expected := []types.Index{
		{
			ID:             1,
			Commit:         makeCommit(1),
			QueuedAt:       time.Time{},
			State:          "queued",
			FailureMessage: nil,
			StartedAt:      nil,
			FinishedAt:     nil,
			RepositoryID:   50,
			RepositoryName: "n-50",
			DockerSteps: []types.DockerStep{
				{
					Image:    "cimg/node:12.16",
					Commands: []string{"yarn install --frozen-lockfile --no-progress"},
				},
			},
			LocalSteps:  []string{"echo hello"},
			Root:        "/foo/bar",
			Indexer:     "sourcegraph/scip-typescript:latest",
			IndexerArgs: []string{"index", "--yarn-workspaces"},
			Outfile:     "dump.lsif",
			ExecutionLogs: []executor.ExecutionLogEntry{
				{Command: []string{"op", "1"}, Out: "Indexing\nUploading\nDone with 1.\n"},
				{Command: []string{"op", "2"}, Out: "Indexing\nUploading\nDone with 2.\n"},
			},
			Rank: &rank1,
		},
		{
			ID:             2,
			Commit:         makeCommit(2),
			QueuedAt:       time.Time{},
			State:          "queued",
			FailureMessage: nil,
			StartedAt:      nil,
			FinishedAt:     nil,
			RepositoryID:   50,
			RepositoryName: "n-50",
			DockerSteps: []types.DockerStep{
				{
					Image:    "cimg/rust:nightly",
					Commands: []string{"cargo install"},
				},
			},
			LocalSteps:  []string{},
			Root:        "/baz",
			Indexer:     "sourcegraph/lsif-rust:15",
			IndexerArgs: []string{"-v"},
			Outfile:     "dump.lsif",
			ExecutionLogs: []executor.ExecutionLogEntry{
				{Command: []string{"op", "1"}, Out: "Done with 1.\n"},
				{Command: []string{"op", "2"}, Out: "Done with 2.\n"},
			},
			Rank: &rank2,
		},
	}

	for i := range expected {
		// Update auto-generated timestamp
		expected[i].QueuedAt = indexes[0].QueuedAt
	}

	if diff := cmp.Diff(expected, indexes); diff != "" {
		t.Errorf("unexpected indexes (-want +got):\n%s", diff)
	}
}

func TestGetIndexes(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(-time.Minute * 1)
	t3 := t1.Add(-time.Minute * 2)
	t4 := t1.Add(-time.Minute * 3)
	t5 := t1.Add(-time.Minute * 4)
	t6 := t1.Add(-time.Minute * 5)
	t7 := t1.Add(-time.Minute * 6)
	t8 := t1.Add(-time.Minute * 7)
	t9 := t1.Add(-time.Minute * 8)
	t10 := t1.Add(-time.Minute * 9)
	failureMessage := "unlucky 333"

	indexID1, indexID2, indexID3, indexID4 := 1, 3, 5, 5 // note the duplication
	uploadID1, uploadID2, uploadID3, uploadID4 := 10, 11, 12, 13

	insertIndexes(t, db,
		types.Index{ID: 1, Commit: makeCommit(3331), QueuedAt: t1, State: "queued", AssociatedUploadID: &uploadID1},
		types.Index{ID: 2, QueuedAt: t2, State: "errored", FailureMessage: &failureMessage},
		types.Index{ID: 3, Commit: makeCommit(3333), QueuedAt: t3, State: "queued", AssociatedUploadID: &uploadID1},
		types.Index{ID: 4, QueuedAt: t4, State: "queued", RepositoryID: 51, RepositoryName: "foo bar x"},
		types.Index{ID: 5, Commit: makeCommit(3333), QueuedAt: t5, State: "processing", AssociatedUploadID: &uploadID1},
		types.Index{ID: 6, QueuedAt: t6, State: "processing", RepositoryID: 52, RepositoryName: "foo bar y"},
		types.Index{ID: 7, QueuedAt: t7},
		types.Index{ID: 8, QueuedAt: t8},
		types.Index{ID: 9, QueuedAt: t9, State: "queued"},
		types.Index{ID: 10, QueuedAt: t10},
	)
	insertUploads(t, db,
		Upload{ID: uploadID1, AssociatedIndexID: &indexID1},
		Upload{ID: uploadID2, AssociatedIndexID: &indexID2},
		Upload{ID: uploadID3, AssociatedIndexID: &indexID3},
		Upload{ID: uploadID4, AssociatedIndexID: &indexID4},
	)

	testCases := []struct {
		repositoryID  int
		state         string
		states        []string
		term          string
		withoutUpload bool
		expectedIDs   []int
	}{
		{expectedIDs: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{repositoryID: 50, expectedIDs: []int{1, 2, 3, 5, 7, 8, 9, 10}},
		{state: "completed", expectedIDs: []int{7, 8, 10}},
		{term: "003", expectedIDs: []int{1, 3, 5}},                                 // searches commits
		{term: "333", expectedIDs: []int{1, 2, 3, 5}},                              // searches commits and failure message
		{term: "QuEuEd", expectedIDs: []int{1, 3, 4, 9}},                           // searches text status
		{term: "bAr", expectedIDs: []int{4, 6}},                                    // search repo names
		{state: "failed", expectedIDs: []int{2}},                                   // treats errored/failed states equivalently
		{states: []string{"completed", "failed"}, expectedIDs: []int{2, 7, 8, 10}}, // searches multiple states
		{withoutUpload: true, expectedIDs: []int{2, 4, 6, 7, 8, 9, 10}},            // anti-join with upload records
	}

	for _, testCase := range testCases {
		for lo := 0; lo < len(testCase.expectedIDs); lo++ {
			hi := lo + 3
			if hi > len(testCase.expectedIDs) {
				hi = len(testCase.expectedIDs)
			}

			name := fmt.Sprintf(
				"repositoryID=%d state=%s states=%s term=%s without_upload=%v offset=%d",
				testCase.repositoryID,
				testCase.state,
				strings.Join(testCase.states, ","),
				testCase.term,
				testCase.withoutUpload,
				lo,
			)

			t.Run(name, func(t *testing.T) {
				indexes, totalCount, err := store.GetIndexes(ctx, shared.GetIndexesOptions{
					RepositoryID:  testCase.repositoryID,
					State:         testCase.state,
					States:        testCase.states,
					Term:          testCase.term,
					WithoutUpload: testCase.withoutUpload,
					Limit:         3,
					Offset:        lo,
				})
				if err != nil {
					t.Fatalf("unexpected error getting indexes for repo: %s", err)
				}
				if totalCount != len(testCase.expectedIDs) {
					t.Errorf("unexpected total count. want=%d have=%d", len(testCase.expectedIDs), totalCount)
				}

				var ids []int
				for _, index := range indexes {
					ids = append(ids, index.ID)
				}

				if diff := cmp.Diff(testCase.expectedIDs[lo:hi], ids); diff != "" {
					t.Errorf("unexpected index ids at offset %d (-want +got):\n%s", lo, diff)
				}
			})
		}
	}

	t.Run("enforce repository permissions", func(t *testing.T) {
		// Enable permissions user mapping forces checking repository permissions
		// against permissions tables in the database, which should effectively block
		// all access because permissions tables are empty.
		before := globals.PermissionsUserMapping()
		globals.SetPermissionsUserMapping(&schema.PermissionsUserMapping{Enabled: true})
		defer globals.SetPermissionsUserMapping(before)

		indexes, totalCount, err := store.GetIndexes(ctx,
			shared.GetIndexesOptions{
				Limit: 1,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if len(indexes) > 0 || totalCount > 0 {
			t.Fatalf("Want no index but got %d indexes with totalCount %d", len(indexes), totalCount)
		}
	})
}

func TestGetIndexByID(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	// Index does not exist initially
	if _, exists, err := store.GetIndexByID(ctx, 1); err != nil {
		t.Fatalf("unexpected error getting index: %s", err)
	} else if exists {
		t.Fatal("unexpected record")
	}

	uploadID := 5
	queuedAt := time.Unix(1587396557, 0).UTC()
	startedAt := queuedAt.Add(time.Minute)
	expected := types.Index{
		ID:             1,
		Commit:         makeCommit(1),
		QueuedAt:       queuedAt,
		State:          "processing",
		FailureMessage: nil,
		StartedAt:      &startedAt,
		FinishedAt:     nil,
		RepositoryID:   123,
		RepositoryName: "n-123",
		DockerSteps: []types.DockerStep{
			{
				Image:    "cimg/node:12.16",
				Commands: []string{"yarn install --frozen-lockfile --no-progress"},
			},
		},
		LocalSteps:  []string{"echo hello"},
		Root:        "/foo/bar",
		Indexer:     "sourcegraph/scip-typescript:latest",
		IndexerArgs: []string{"index", "--yarn-workspaces"},
		Outfile:     "dump.lsif",
		ExecutionLogs: []executor.ExecutionLogEntry{
			{Command: []string{"op", "1"}, Out: "Indexing\nUploading\nDone with 1.\n"},
			{Command: []string{"op", "2"}, Out: "Indexing\nUploading\nDone with 2.\n"},
		},
		Rank:               nil,
		AssociatedUploadID: &uploadID,
	}

	insertIndexes(t, db, expected)
	insertUploads(t, db, Upload{ID: uploadID, AssociatedIndexID: &expected.ID})

	if index, exists, err := store.GetIndexByID(ctx, 1); err != nil {
		t.Fatalf("unexpected error getting index: %s", err)
	} else if !exists {
		t.Fatal("expected record to exist")
	} else if diff := cmp.Diff(expected, index); diff != "" {
		t.Errorf("unexpected index (-want +got):\n%s", diff)
	}

	t.Run("enforce repository permissions", func(t *testing.T) {
		// Enable permissions user mapping forces checking repository permissions
		// against permissions tables in the database, which should effectively block
		// all access because permissions tables are empty.
		before := globals.PermissionsUserMapping()
		globals.SetPermissionsUserMapping(&schema.PermissionsUserMapping{Enabled: true})
		defer globals.SetPermissionsUserMapping(before)

		_, exists, err := store.GetIndexByID(ctx, 1)
		if err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Fatalf("exists: want false but got %v", exists)
		}
	})
}

func TestGetIndexesByIDs(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	indexID1, indexID2, indexID3, indexID4 := 1, 3, 5, 5 // note the duplication
	uploadID1, uploadID2, uploadID3, uploadID4 := 10, 11, 12, 13

	insertIndexes(t, db,
		types.Index{ID: 1, AssociatedUploadID: &uploadID1},
		types.Index{ID: 2},
		types.Index{ID: 3, AssociatedUploadID: &uploadID1},
		types.Index{ID: 4},
		types.Index{ID: 5, AssociatedUploadID: &uploadID1},
		types.Index{ID: 6},
		types.Index{ID: 7},
		types.Index{ID: 8},
		types.Index{ID: 9},
		types.Index{ID: 10},
	)
	insertUploads(t, db,
		Upload{ID: uploadID1, AssociatedIndexID: &indexID1},
		Upload{ID: uploadID2, AssociatedIndexID: &indexID2},
		Upload{ID: uploadID3, AssociatedIndexID: &indexID3},
		Upload{ID: uploadID4, AssociatedIndexID: &indexID4},
	)

	t.Run("fetch", func(t *testing.T) {
		indexes, err := store.GetIndexesByIDs(ctx, 2, 4, 6, 8, 12)
		if err != nil {
			t.Fatalf("unexpected error getting indexes for repo: %s", err)
		}

		var ids []int
		for _, index := range indexes {
			ids = append(ids, index.ID)
		}
		sort.Ints(ids)

		if diff := cmp.Diff([]int{2, 4, 6, 8}, ids); diff != "" {
			t.Errorf("unexpected index ids (-want +got):\n%s", diff)
		}
	})

	t.Run("enforce repository permissions", func(t *testing.T) {
		// Enable permissions user mapping forces checking repository permissions
		// against permissions tables in the database, which should effectively block
		// all access because permissions tables are empty.
		before := globals.PermissionsUserMapping()
		globals.SetPermissionsUserMapping(&schema.PermissionsUserMapping{Enabled: true})
		defer globals.SetPermissionsUserMapping(before)

		indexes, err := store.GetIndexesByIDs(ctx, 1, 2, 3, 4)
		if err != nil {
			t.Fatal(err)
		}
		if len(indexes) > 0 {
			t.Fatalf("Want no index but got %d indexes", len(indexes))
		}
	})
}

func TestGetQueuedIndexRank(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(+time.Minute * 6)
	t3 := t1.Add(+time.Minute * 3)
	t4 := t1.Add(+time.Minute * 1)
	t5 := t1.Add(+time.Minute * 4)
	t6 := t1.Add(+time.Minute * 2)
	t7 := t1.Add(+time.Minute * 5)

	insertIndexes(t, db,
		types.Index{ID: 1, QueuedAt: t1, State: "queued"},
		types.Index{ID: 2, QueuedAt: t2, State: "queued"},
		types.Index{ID: 3, QueuedAt: t3, State: "queued"},
		types.Index{ID: 4, QueuedAt: t4, State: "queued"},
		types.Index{ID: 5, QueuedAt: t5, State: "queued"},
		types.Index{ID: 6, QueuedAt: t6, State: "processing"},
		types.Index{ID: 7, QueuedAt: t1, State: "queued", ProcessAfter: &t7},
	)

	if index, _, _ := store.GetIndexByID(context.Background(), 1); index.Rank == nil || *index.Rank != 1 {
		t.Errorf("unexpected rank. want=%d have=%s", 1, printableRank{index.Rank})
	}
	if index, _, _ := store.GetIndexByID(context.Background(), 2); index.Rank == nil || *index.Rank != 6 {
		t.Errorf("unexpected rank. want=%d have=%s", 5, printableRank{index.Rank})
	}
	if index, _, _ := store.GetIndexByID(context.Background(), 3); index.Rank == nil || *index.Rank != 3 {
		t.Errorf("unexpected rank. want=%d have=%s", 3, printableRank{index.Rank})
	}
	if index, _, _ := store.GetIndexByID(context.Background(), 4); index.Rank == nil || *index.Rank != 2 {
		t.Errorf("unexpected rank. want=%d have=%s", 2, printableRank{index.Rank})
	}
	if index, _, _ := store.GetIndexByID(context.Background(), 5); index.Rank == nil || *index.Rank != 4 {
		t.Errorf("unexpected rank. want=%d have=%s", 4, printableRank{index.Rank})
	}

	// Only considers queued indexes to determine rank
	if index, _, _ := store.GetIndexByID(context.Background(), 6); index.Rank != nil {
		t.Errorf("unexpected rank. want=%s have=%s", "nil", printableRank{index.Rank})
	}

	// Process after takes priority over upload time
	if upload, _, _ := store.GetIndexByID(context.Background(), 7); upload.Rank == nil || *upload.Rank != 5 {
		t.Errorf("unexpected rank. want=%d have=%s", 4, printableRank{upload.Rank})
	}
}

func TestRecentIndexesSummary(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	t0 := time.Unix(1587396557, 0).UTC()
	t1 := t0.Add(-time.Minute * 1)
	t2 := t0.Add(-time.Minute * 2)
	t3 := t0.Add(-time.Minute * 3)
	t4 := t0.Add(-time.Minute * 4)
	t5 := t0.Add(-time.Minute * 5)
	t6 := t0.Add(-time.Minute * 6)
	t7 := t0.Add(-time.Minute * 7)
	t8 := t0.Add(-time.Minute * 8)
	t9 := t0.Add(-time.Minute * 9)

	r1 := 1
	r2 := 2

	addDefaults := func(index types.Index) types.Index {
		index.Commit = makeCommit(index.ID)
		index.RepositoryID = 50
		index.RepositoryName = "n-50"
		index.DockerSteps = []types.DockerStep{}
		index.IndexerArgs = []string{}
		index.LocalSteps = []string{}
		return index
	}

	indexes := []types.Index{
		addDefaults(types.Index{ID: 150, QueuedAt: t0, Root: "r1", Indexer: "i1", State: "queued", Rank: &r2}), // visible (group 1)
		addDefaults(types.Index{ID: 151, QueuedAt: t1, Root: "r1", Indexer: "i1", State: "queued", Rank: &r1}), // visible (group 1)
		addDefaults(types.Index{ID: 152, FinishedAt: &t2, Root: "r1", Indexer: "i1", State: "errored"}),        // visible (group 1)
		addDefaults(types.Index{ID: 153, FinishedAt: &t3, Root: "r1", Indexer: "i2", State: "completed"}),      // visible (group 2)
		addDefaults(types.Index{ID: 154, FinishedAt: &t4, Root: "r2", Indexer: "i1", State: "completed"}),      // visible (group 3)
		addDefaults(types.Index{ID: 155, FinishedAt: &t5, Root: "r2", Indexer: "i1", State: "errored"}),        // shadowed
		addDefaults(types.Index{ID: 156, FinishedAt: &t6, Root: "r2", Indexer: "i2", State: "completed"}),      // visible (group 4)
		addDefaults(types.Index{ID: 157, FinishedAt: &t7, Root: "r2", Indexer: "i2", State: "errored"}),        // shadowed
		addDefaults(types.Index{ID: 158, FinishedAt: &t8, Root: "r2", Indexer: "i2", State: "errored"}),        // shadowed
		addDefaults(types.Index{ID: 159, FinishedAt: &t9, Root: "r2", Indexer: "i2", State: "errored"}),        // shadowed
	}
	insertIndexes(t, db, indexes...)

	summary, err := store.GetRecentIndexesSummary(ctx, 50)
	if err != nil {
		t.Fatalf("unexpected error querying recent index summary: %s", err)
	}

	expected := []shared.IndexesWithRepositoryNamespace{
		{Root: "r1", Indexer: "i1", Indexes: []types.Index{indexes[0], indexes[1], indexes[2]}},
		{Root: "r1", Indexer: "i2", Indexes: []types.Index{indexes[3]}},
		{Root: "r2", Indexer: "i1", Indexes: []types.Index{indexes[4]}},
		{Root: "r2", Indexer: "i2", Indexes: []types.Index{indexes[6]}},
	}
	if diff := cmp.Diff(expected, summary); diff != "" {
		t.Errorf("unexpected index summary (-want +got):\n%s", diff)
	}
}

func TestGetLastIndexScanForRepository(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	ts, err := store.GetLastIndexScanForRepository(ctx, 50)
	if err != nil {
		t.Fatalf("unexpected error querying last index scan: %s", err)
	}
	if ts != nil {
		t.Fatalf("unexpected timestamp for repository. want=%v have=%s", nil, ts)
	}

	expected := time.Unix(1587396557, 0).UTC()

	if err := basestore.NewWithHandle(db.Handle()).Exec(ctx, sqlf.Sprintf(`
		INSERT INTO lsif_last_index_scan (repository_id, last_index_scan_at)
		VALUES (%s, %s)
	`, 50, expected)); err != nil {
		t.Fatalf("unexpected error inserting timestamp: %s", err)
	}

	ts, err = store.GetLastIndexScanForRepository(ctx, 50)
	if err != nil {
		t.Fatalf("unexpected error querying last index scan: %s", err)
	}

	if ts == nil || !ts.Equal(expected) {
		t.Fatalf("unexpected timestamp for repository. want=%s have=%s", expected, ts)
	}
}

func TestDeleteIndexByID(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	insertIndexes(t, db, types.Index{ID: 1})

	if found, err := store.DeleteIndexByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error deleting index: %s", err)
	} else if !found {
		t.Fatalf("expected record to exist")
	}

	// Index no longer exists
	if _, exists, err := store.GetIndexByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting index: %s", err)
	} else if exists {
		t.Fatal("unexpected record")
	}
}

func TestDeleteIndexes(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	insertIndexes(t, db, types.Index{ID: 1, State: "completed"})
	insertIndexes(t, db, types.Index{ID: 2, State: "errored"})

	if err := store.DeleteIndexes(context.Background(), shared.DeleteIndexesOptions{
		States:       []string{"errored"},
		Term:         "",
		RepositoryID: 0,
	}); err != nil {
		t.Fatalf("unexpected error deleting indexes: %s", err)
	}

	// Index no longer exists
	if _, exists, err := store.GetIndexByID(context.Background(), 2); err != nil {
		t.Fatalf("unexpected error getting index: %s", err)
	} else if exists {
		t.Fatal("unexpected record")
	}
}

func TestDeleteIndexByIDMissingRow(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	if found, err := store.DeleteIndexByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error deleting index: %s", err)
	} else if found {
		t.Fatalf("unexpected record")
	}
}

func TestDeleteIndexesWithoutRepository(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	var indexes []types.Index
	for i := 0; i < 25; i++ {
		for j := 0; j < 10+i; j++ {
			indexes = append(indexes, types.Index{ID: len(indexes) + 1, RepositoryID: 50 + i})
		}
	}
	insertIndexes(t, db, indexes...)

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(-DeletedRepositoryGracePeriod + time.Minute)
	t3 := t1.Add(-DeletedRepositoryGracePeriod - time.Minute)

	deletions := map[int]time.Time{
		52: t2, 54: t2, 56: t2, // deleted too recently
		61: t3, 63: t3, 65: t3, // deleted
	}

	for repositoryID, deletedAt := range deletions {
		query := sqlf.Sprintf(`UPDATE repo SET deleted_at=%s WHERE id=%s`, deletedAt, repositoryID)

		if _, err := db.QueryContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("Failed to update repository: %s", err)
		}
	}

	ids, err := store.DeleteIndexesWithoutRepository(context.Background(), t1)
	if err != nil {
		t.Fatalf("unexpected error deleting indexes: %s", err)
	}

	expected := map[int]int{
		61: 21,
		63: 23,
		65: 25,
	}
	if diff := cmp.Diff(expected, ids); diff != "" {
		t.Errorf("unexpected ids (-want +got):\n%s", diff)
	}
}

func TestReindexIndexes(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	insertIndexes(t, db, types.Index{ID: 1, State: "completed"})
	insertIndexes(t, db, types.Index{ID: 2, State: "errored"})

	if err := store.ReindexIndexes(context.Background(), shared.ReindexIndexesOptions{
		States:       []string{"errored"},
		Term:         "",
		RepositoryID: 0,
	}); err != nil {
		t.Fatalf("unexpected error deleting indexes: %s", err)
	}

	// Index has been marked for reindexing
	if index, exists, err := store.GetIndexByID(context.Background(), 2); err != nil {
		t.Fatalf("unexpected error getting index: %s", err)
	} else if !exists {
		t.Fatal("index missing")
	} else if !index.ShouldReindex {
		t.Fatal("index not marked for reindexing")
	}
}

func TestReindexIndexByID(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	insertIndexes(t, db, types.Index{ID: 1, State: "completed"})
	insertIndexes(t, db, types.Index{ID: 2, State: "errored"})

	if err := store.ReindexIndexByID(context.Background(), 2); err != nil {
		t.Fatalf("unexpected error deleting indexes: %s", err)
	}

	// Index has been marked for reindexing
	if index, exists, err := store.GetIndexByID(context.Background(), 2); err != nil {
		t.Fatalf("unexpected error getting index: %s", err)
	} else if !exists {
		t.Fatal("index missing")
	} else if !index.ShouldReindex {
		t.Fatal("index not marked for reindexing")
	}
}

func TestIsQueued(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	insertIndexes(t, db, types.Index{ID: 1, RepositoryID: 1, Commit: makeCommit(1)})
	insertIndexes(t, db, types.Index{ID: 2, RepositoryID: 1, Commit: makeCommit(1), ShouldReindex: true})
	insertIndexes(t, db, types.Index{ID: 3, RepositoryID: 4, Commit: makeCommit(1), ShouldReindex: true})
	insertIndexes(t, db, types.Index{ID: 4, RepositoryID: 5, Commit: makeCommit(4), ShouldReindex: true})
	insertUploads(t, db, Upload{ID: 2, RepositoryID: 2, Commit: makeCommit(2)})
	insertUploads(t, db, Upload{ID: 3, RepositoryID: 3, Commit: makeCommit(3), State: "deleted"})
	insertUploads(t, db, Upload{ID: 4, RepositoryID: 5, Commit: makeCommit(4), ShouldReindex: true})

	testCases := []struct {
		repositoryID int
		commit       string
		expected     bool
	}{
		{1, makeCommit(1), true},
		{1, makeCommit(2), false},
		{2, makeCommit(1), false},
		{2, makeCommit(2), true},
		{3, makeCommit(1), false},
		{3, makeCommit(2), false},
		{3, makeCommit(3), false},
		{4, makeCommit(1), false},
		{5, makeCommit(4), false},
	}

	for _, testCase := range testCases {
		name := fmt.Sprintf("repositoryId=%d commit=%s", testCase.repositoryID, testCase.commit)

		t.Run(name, func(t *testing.T) {
			queued, err := store.IsQueued(context.Background(), testCase.repositoryID, testCase.commit)
			if err != nil {
				t.Fatalf("unexpected error checking if commit is queued: %s", err)
			}
			if queued != testCase.expected {
				t.Errorf("unexpected state. repo=%v commit=%v want=%v have=%v", testCase.repositoryID, testCase.commit, testCase.expected, queued)
			}
		})
	}
}

func TestIsQueuedRootIndexer(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	now := time.Now()
	insertIndexes(t, db, types.Index{ID: 1, RepositoryID: 1, Commit: makeCommit(1), Root: "/foo", Indexer: "i1", QueuedAt: now.Add(-time.Hour * 1)})
	insertIndexes(t, db, types.Index{ID: 2, RepositoryID: 1, Commit: makeCommit(1), Root: "/foo", Indexer: "i1", QueuedAt: now.Add(-time.Hour * 2)})
	insertIndexes(t, db, types.Index{ID: 3, RepositoryID: 2, Commit: makeCommit(2), Root: "/foo", Indexer: "i1", QueuedAt: now.Add(-time.Hour * 1), ShouldReindex: true})
	insertIndexes(t, db, types.Index{ID: 4, RepositoryID: 2, Commit: makeCommit(2), Root: "/foo", Indexer: "i1", QueuedAt: now.Add(-time.Hour * 2)})
	insertIndexes(t, db, types.Index{ID: 5, RepositoryID: 3, Commit: makeCommit(3), Root: "/foo", Indexer: "i1", QueuedAt: now.Add(-time.Hour * 1)})
	insertIndexes(t, db, types.Index{ID: 6, RepositoryID: 3, Commit: makeCommit(3), Root: "/foo", Indexer: "i1", QueuedAt: now.Add(-time.Hour * 2), ShouldReindex: true})

	testCases := []struct {
		repositoryID int
		commit       string
		root         string
		indexer      string
		expected     bool
	}{
		{1, makeCommit(1), "/foo", "i1", true},
		{1, makeCommit(1), "/bar", "i1", false}, // no index for root
		{2, makeCommit(2), "/foo", "i1", false}, // reindex (live)
		{3, makeCommit(3), "/foo", "i1", true},  // reindex (done)
	}

	for _, testCase := range testCases {
		name := fmt.Sprintf("repositoryId=%d commit=%s", testCase.repositoryID, testCase.commit)

		t.Run(name, func(t *testing.T) {
			queued, err := store.IsQueuedRootIndexer(context.Background(), testCase.repositoryID, testCase.commit, testCase.root, testCase.indexer)
			if err != nil {
				t.Fatalf("unexpected error checking if commit/root/indexer is queued: %s", err)
			}
			if queued != testCase.expected {
				t.Errorf("unexpected state. repo=%v commit=%v root=%v indexer=%v want=%v have=%v", testCase.repositoryID, testCase.commit, testCase.root, testCase.indexer, testCase.expected, queued)
			}
		})
	}
}

func TestGetQueuedRepoRev(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	expected := []RepoRev{
		{1, 50, "HEAD"},
		{2, 50, "HEAD~1"},
		{3, 50, "HEAD~2"},
		{4, 51, "HEAD"},
		{5, 51, "HEAD~1"},
		{6, 51, "HEAD~2"},
		{7, 52, "HEAD"},
		{8, 52, "HEAD~1"},
		{9, 52, "HEAD~2"},
	}
	for _, repoRev := range expected {
		if err := store.QueueRepoRev(ctx, repoRev.RepositoryID, repoRev.Rev); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	}

	// entire set
	repoRevs, err := store.GetQueuedRepoRev(ctx, 50)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if diff := cmp.Diff(expected, repoRevs); diff != "" {
		t.Errorf("unexpected repo revs (-want +got):\n%s", diff)
	}

	// smaller page size
	repoRevs, err = store.GetQueuedRepoRev(ctx, 5)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if diff := cmp.Diff(expected[:5], repoRevs); diff != "" {
		t.Errorf("unexpected repo revs (-want +got):\n%s", diff)
	}
}

func TestMarkRepoRevsAsProcessed(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	expected := []RepoRev{
		{1, 50, "HEAD"},
		{2, 50, "HEAD~1"},
		{3, 50, "HEAD~2"},
		{4, 51, "HEAD"},
		{5, 51, "HEAD~1"},
		{6, 51, "HEAD~2"},
		{7, 52, "HEAD"},
		{8, 52, "HEAD~1"},
		{9, 52, "HEAD~2"},
	}
	for _, repoRev := range expected {
		if err := store.QueueRepoRev(ctx, repoRev.RepositoryID, repoRev.Rev); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	}

	// entire set
	repoRevs, err := store.GetQueuedRepoRev(ctx, 50)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if diff := cmp.Diff(expected, repoRevs); diff != "" {
		t.Errorf("unexpected repo revs (-want +got):\n%s", diff)
	}

	// mark first elements as complete; re-request remaining
	if err := store.MarkRepoRevsAsProcessed(ctx, []int{1, 2, 3, 4, 5}); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	repoRevs, err = store.GetQueuedRepoRev(ctx, 50)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if diff := cmp.Diff(expected[5:], repoRevs); diff != "" {
		t.Errorf("unexpected repo revs (-want +got):\n%s", diff)
	}
}

// Upload is a subset of the lsif_uploads table and stores both processed and unprocessed
// records.
type Upload struct {
	ID                int
	Commit            string
	Root              string
	VisibleAtTip      bool
	UploadedAt        time.Time
	State             string
	FailureMessage    *string
	StartedAt         *time.Time
	FinishedAt        *time.Time
	ProcessAfter      *time.Time
	NumResets         int
	NumFailures       int
	RepositoryID      int
	RepositoryName    string
	Indexer           string
	IndexerVersion    string
	NumParts          int
	UploadedParts     []int
	UploadSize        *int64
	UncompressedSize  *int64
	Rank              *int
	AssociatedIndexID *int
	ShouldReindex     bool
}

// insertUploads populates the lsif_uploads table with the given upload models.
func insertUploads(t testing.TB, db database.DB, uploads ...Upload) {
	for _, upload := range uploads {
		if upload.Commit == "" {
			upload.Commit = makeCommit(upload.ID)
		}
		if upload.State == "" {
			upload.State = "completed"
		}
		if upload.RepositoryID == 0 {
			upload.RepositoryID = 50
		}
		if upload.Indexer == "" {
			upload.Indexer = "lsif-go"
		}
		if upload.IndexerVersion == "" {
			upload.IndexerVersion = "latest"
		}
		if upload.UploadedParts == nil {
			upload.UploadedParts = []int{}
		}

		// Ensure we have a repo for the inner join in select queries
		insertRepo(t, db, upload.RepositoryID, upload.RepositoryName)

		query := sqlf.Sprintf(`
			INSERT INTO lsif_uploads (
				id,
				commit,
				root,
				uploaded_at,
				state,
				failure_message,
				started_at,
				finished_at,
				process_after,
				num_resets,
				num_failures,
				repository_id,
				indexer,
				indexer_version,
				num_parts,
				uploaded_parts,
				upload_size,
				associated_index_id,
				should_reindex
			) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
		`,
			upload.ID,
			upload.Commit,
			upload.Root,
			upload.UploadedAt,
			upload.State,
			upload.FailureMessage,
			upload.StartedAt,
			upload.FinishedAt,
			upload.ProcessAfter,
			upload.NumResets,
			upload.NumFailures,
			upload.RepositoryID,
			upload.Indexer,
			upload.IndexerVersion,
			upload.NumParts,
			pq.Array(upload.UploadedParts),
			upload.UploadSize,
			upload.AssociatedIndexID,
			upload.ShouldReindex,
		)

		if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("unexpected error while inserting upload: %s", err)
		}
	}
}

type printableRank struct{ value *int }

func (r printableRank) String() string {
	if r.value == nil {
		return "nil"
	}
	return strconv.Itoa(*r.value)
}
