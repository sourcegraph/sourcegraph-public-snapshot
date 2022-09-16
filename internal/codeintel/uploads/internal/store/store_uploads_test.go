package store

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
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
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGetUploads(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)
	ctx := context.Background()

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
	t11 := t1.Add(-time.Minute * 10)
	failureMessage := "unlucky 333"

	insertUploads(t, db,
		shared.Upload{ID: 1, Commit: makeCommit(3331), UploadedAt: t1, Root: "sub1/", State: "queued"},
		shared.Upload{ID: 2, UploadedAt: t2, FinishedAt: &t1, State: "errored", FailureMessage: &failureMessage, Indexer: "scip-typescript"},
		shared.Upload{ID: 3, Commit: makeCommit(3333), UploadedAt: t3, Root: "sub2/", State: "queued"},
		shared.Upload{ID: 4, UploadedAt: t4, State: "queued", RepositoryID: 51, RepositoryName: "foo bar x"},
		shared.Upload{ID: 5, Commit: makeCommit(3333), UploadedAt: t5, Root: "sub1/", State: "processing", Indexer: "scip-typescript"},
		shared.Upload{ID: 6, UploadedAt: t6, Root: "sub2/", State: "processing", RepositoryID: 52, RepositoryName: "foo bar y"},
		shared.Upload{ID: 7, UploadedAt: t7, FinishedAt: &t4, Root: "sub1/", Indexer: "scip-typescript"},
		shared.Upload{ID: 8, UploadedAt: t8, FinishedAt: &t4, Indexer: "scip-typescript"},
		shared.Upload{ID: 9, UploadedAt: t9, State: "queued"},
		shared.Upload{ID: 10, UploadedAt: t10, FinishedAt: &t6, Root: "sub1/", Indexer: "scip-typescript"},
		shared.Upload{ID: 11, UploadedAt: t11, FinishedAt: &t6, Root: "sub1/", Indexer: "scip-typescript"},

		// Deleted duplicates
		shared.Upload{ID: 12, Commit: makeCommit(3331), UploadedAt: t1, FinishedAt: &t1, Root: "sub1/", State: "deleted"},
		shared.Upload{ID: 13, UploadedAt: t2, FinishedAt: &t1, State: "deleted", FailureMessage: &failureMessage, Indexer: "scip-typescript"},
		shared.Upload{ID: 14, Commit: makeCommit(3333), UploadedAt: t3, FinishedAt: &t2, Root: "sub2/", State: "deleted"},

		// deleted repo
		shared.Upload{ID: 15, Commit: makeCommit(3334), UploadedAt: t4, State: "deleted", RepositoryID: 53, RepositoryName: "DELETED-barfoo"},

		// to-be hard deleted
		shared.Upload{ID: 16, Commit: makeCommit(3333), UploadedAt: t4, FinishedAt: &t3, State: "deleted"},
		shared.Upload{ID: 17, Commit: makeCommit(3334), UploadedAt: t4, FinishedAt: &t5, State: "deleting"},
	)
	insertVisibleAtTip(t, db, 50, 2, 5, 7, 8)

	updateUploads(t, db, shared.Upload{
		ID: 17, State: "deleted",
	})

	deleteUploads(t, db, 16)
	deleteUploads(t, db, 17)

	query := sqlf.Sprintf(
		`DELETE FROM lsif_uploads_audit_logs WHERE upload_id = %s
			AND sequence NOT IN (
				SELECT MAX(sequence) FROM lsif_uploads_audit_logs
				WHERE upload_id = %s
			)`,
		17, 17)
	if _, err := db.ExecContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// upload 10 depends on uploads 7 and 8
	insertPackages(t, store, []shared.Package{
		{DumpID: 7, Scheme: "npm", Name: "foo", Version: "0.1.0"},
		{DumpID: 8, Scheme: "npm", Name: "bar", Version: "1.2.3"},
		{DumpID: 11, Scheme: "npm", Name: "foo", Version: "0.1.0"}, // duplicate package
	})
	insertPackageReferences(t, store, []shared.PackageReference{
		{Package: shared.Package{DumpID: 7, Scheme: "npm", Name: "bar", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 10, Scheme: "npm", Name: "foo", Version: "0.1.0"}},
		{Package: shared.Package{DumpID: 10, Scheme: "npm", Name: "bar", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 11, Scheme: "npm", Name: "bar", Version: "1.2.3"}},
	})

	t.Logf("%v", sqlf.Sprintf(
		`INSERT INTO lsif_dirty_repositories(repository_id, update_token, dirty_token, updated_at) VALUES (%s, 10, 20, %s)`,
		50,
		t5,
	).Query(sqlf.PostgresBindVar))

	dirtyRepositoryQuery := sqlf.Sprintf(
		`INSERT INTO lsif_dirty_repositories(repository_id, update_token, dirty_token, updated_at) VALUES (%s, 10, 20, %s)`,
		50,
		t5,
	)
	if _, err := db.ExecContext(ctx, dirtyRepositoryQuery.Query(sqlf.PostgresBindVar), dirtyRepositoryQuery.Args()...); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	type testCase struct {
		repositoryID     int
		state            string
		term             string
		visibleAtTip     bool
		dependencyOf     int
		dependentOf      int
		uploadedBefore   *time.Time
		uploadedAfter    *time.Time
		inCommitGraph    bool
		oldestFirst      bool
		allowDeletedRepo bool
		expectedIDs      []int
	}
	testCases := []testCase{
		{expectedIDs: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}},
		{oldestFirst: true, expectedIDs: []int{11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}},
		{repositoryID: 50, expectedIDs: []int{1, 2, 3, 5, 7, 8, 9, 10, 11}},
		{state: "completed", expectedIDs: []int{7, 8, 10, 11}},
		{term: "sub", expectedIDs: []int{1, 3, 5, 6, 7, 10, 11}},     // searches root
		{term: "003", expectedIDs: []int{1, 3, 5}},                   // searches commits
		{term: "333", expectedIDs: []int{1, 2, 3, 5}},                // searches commits and failure message
		{term: "typescript", expectedIDs: []int{2, 5, 7, 8, 10, 11}}, // searches indexer
		{term: "QuEuEd", expectedIDs: []int{1, 3, 4, 9}},             // searches text status
		{term: "bAr", expectedIDs: []int{4, 6}},                      // search repo names
		{state: "failed", expectedIDs: []int{2}},                     // treats errored/failed states equivalently
		{visibleAtTip: true, expectedIDs: []int{2, 5, 7, 8}},
		{uploadedBefore: &t5, expectedIDs: []int{6, 7, 8, 9, 10, 11}},
		{uploadedAfter: &t4, expectedIDs: []int{1, 2, 3}},
		{inCommitGraph: true, expectedIDs: []int{10, 11}},
		{dependencyOf: 7, expectedIDs: []int{8}},
		{dependentOf: 7, expectedIDs: []int{10}},
		{dependencyOf: 8, expectedIDs: []int{}},
		{dependentOf: 8, expectedIDs: []int{7, 10, 11}},
		{dependencyOf: 10, expectedIDs: []int{7, 8}},
		{dependentOf: 10, expectedIDs: []int{}},
		{dependencyOf: 11, expectedIDs: []int{8}},
		{dependentOf: 11, expectedIDs: []int{}},
		{allowDeletedRepo: true, state: "deleted", expectedIDs: []int{12, 13, 14, 15, 16, 17}},
	}

	runTest := func(testCase testCase, lo, hi int) (errors int) {
		name := fmt.Sprintf(
			"repositoryID=%d|state='%s'|term='%s'|visibleAtTip=%v|dependencyOf=%d|dependentOf=%d|offset=%d",
			testCase.repositoryID,
			testCase.state,
			testCase.term,
			testCase.visibleAtTip,
			testCase.dependencyOf,
			testCase.dependentOf,
			lo,
		)

		t.Run(name, func(t *testing.T) {
			uploads, totalCount, err := store.GetUploads(ctx, shared.GetUploadsOptions{
				RepositoryID:     testCase.repositoryID,
				State:            testCase.state,
				Term:             testCase.term,
				VisibleAtTip:     testCase.visibleAtTip,
				DependencyOf:     testCase.dependencyOf,
				DependentOf:      testCase.dependentOf,
				UploadedBefore:   testCase.uploadedBefore,
				UploadedAfter:    testCase.uploadedAfter,
				InCommitGraph:    testCase.inCommitGraph,
				OldestFirst:      testCase.oldestFirst,
				AllowDeletedRepo: testCase.allowDeletedRepo,
				Limit:            3,
				Offset:           lo,
			})
			if err != nil {
				t.Fatalf("unexpected error getting uploads for repo: %s", err)
			}
			if totalCount != len(testCase.expectedIDs) {
				t.Errorf("unexpected total count. want=%d have=%d", len(testCase.expectedIDs), totalCount)
				errors++
			}

			if totalCount != 0 {
				var ids []int
				for _, upload := range uploads {
					ids = append(ids, upload.ID)
				}
				if diff := cmp.Diff(testCase.expectedIDs[lo:hi], ids); diff != "" {
					t.Errorf("unexpected upload ids at offset %d-%d (-want +got):\n%s", lo, hi, diff)
					errors++
				}
			}
		})

		return errors
	}

	for _, testCase := range testCases {
		if n := len(testCase.expectedIDs); n == 0 {
			runTest(testCase, 0, 0)
		} else {
			for lo := 0; lo < n; lo++ {
				if numErrors := runTest(testCase, lo, int(math.Min(float64(lo)+3, float64(n)))); numErrors > 0 {
					break
				}
			}
		}
	}

	t.Run("enforce repository permissions", func(t *testing.T) {
		// Enable permissions user mapping forces checking repository permissions
		// against permissions tables in the database, which should effectively block
		// all access because permissions tables are empty.
		before := globals.PermissionsUserMapping()
		globals.SetPermissionsUserMapping(&schema.PermissionsUserMapping{Enabled: true})
		defer globals.SetPermissionsUserMapping(before)

		uploads, totalCount, err := store.GetUploads(ctx,
			shared.GetUploadsOptions{
				Limit: 1,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if len(uploads) > 0 || totalCount > 0 {
			t.Fatalf("Want no upload but got %d uploads with totalCount %d", len(uploads), totalCount)
		}
	})
}

func TestGetUploadByID(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	// Upload does not exist initially
	if _, exists, err := store.GetUploadByID(ctx, 1); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if exists {
		t.Fatal("unexpected record")
	}

	uploadedAt := time.Unix(1587396557, 0).UTC()
	startedAt := uploadedAt.Add(time.Minute)
	expected := shared.Upload{
		ID:             1,
		Commit:         makeCommit(1),
		Root:           "sub/",
		VisibleAtTip:   true,
		UploadedAt:     uploadedAt,
		State:          "processing",
		FailureMessage: nil,
		StartedAt:      &startedAt,
		FinishedAt:     nil,
		RepositoryID:   123,
		RepositoryName: "n-123",
		Indexer:        "lsif-go",
		IndexerVersion: "1.2.3",
		NumParts:       1,
		UploadedParts:  []int{},
		Rank:           nil,
	}

	insertUploads(t, db, expected)
	insertVisibleAtTip(t, db, 123, 1)

	if upload, exists, err := store.GetUploadByID(ctx, 1); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if !exists {
		t.Fatal("expected record to exist")
	} else if diff := cmp.Diff(expected, upload); diff != "" {
		t.Errorf("unexpected upload (-want +got):\n%s", diff)
	}

	t.Run("enforce repository permissions", func(t *testing.T) {
		// Enable permissions user mapping forces checking repository permissions
		// against permissions tables in the database, which should effectively block
		// all access because permissions tables are empty.
		before := globals.PermissionsUserMapping()
		globals.SetPermissionsUserMapping(&schema.PermissionsUserMapping{Enabled: true})
		defer globals.SetPermissionsUserMapping(before)

		_, exists, err := store.GetUploadByID(ctx, 1)
		if err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Fatalf("exists: want false but got %v", exists)
		}
	})
}

func TestGetUploadByIDDeleted(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	// Upload does not exist initially
	if _, exists, err := store.GetUploadByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if exists {
		t.Fatal("unexpected record")
	}

	uploadedAt := time.Unix(1587396557, 0).UTC()
	startedAt := uploadedAt.Add(time.Minute)
	expected := shared.Upload{
		ID:             1,
		Commit:         makeCommit(1),
		Root:           "sub/",
		VisibleAtTip:   true,
		UploadedAt:     uploadedAt,
		State:          "deleted",
		FailureMessage: nil,
		StartedAt:      &startedAt,
		FinishedAt:     nil,
		RepositoryID:   123,
		RepositoryName: "n-123",
		Indexer:        "lsif-go",
		NumParts:       1,
		UploadedParts:  []int{},
		Rank:           nil,
	}

	insertUploads(t, db, expected)

	// Should still not be queryable
	if _, exists, err := store.GetUploadByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if exists {
		t.Fatal("unexpected record")
	}
}

func TestGetQueuedUploadRank(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(+time.Minute * 6)
	t3 := t1.Add(+time.Minute * 3)
	t4 := t1.Add(+time.Minute * 1)
	t5 := t1.Add(+time.Minute * 4)
	t6 := t1.Add(+time.Minute * 2)
	t7 := t1.Add(+time.Minute * 5)

	insertUploads(t, db,
		shared.Upload{ID: 1, UploadedAt: t1, State: "queued"},
		shared.Upload{ID: 2, UploadedAt: t2, State: "queued"},
		shared.Upload{ID: 3, UploadedAt: t3, State: "queued"},
		shared.Upload{ID: 4, UploadedAt: t4, State: "queued"},
		shared.Upload{ID: 5, UploadedAt: t5, State: "queued"},
		shared.Upload{ID: 6, UploadedAt: t6, State: "processing"},
		shared.Upload{ID: 7, UploadedAt: t1, State: "queued", ProcessAfter: &t7},
	)

	if upload, _, _ := store.GetUploadByID(context.Background(), 1); upload.Rank == nil || *upload.Rank != 1 {
		t.Errorf("unexpected rank. want=%d have=%s", 1, printableRank{upload.Rank})
	}
	if upload, _, _ := store.GetUploadByID(context.Background(), 2); upload.Rank == nil || *upload.Rank != 6 {
		t.Errorf("unexpected rank. want=%d have=%s", 5, printableRank{upload.Rank})
	}
	if upload, _, _ := store.GetUploadByID(context.Background(), 3); upload.Rank == nil || *upload.Rank != 3 {
		t.Errorf("unexpected rank. want=%d have=%s", 3, printableRank{upload.Rank})
	}
	if upload, _, _ := store.GetUploadByID(context.Background(), 4); upload.Rank == nil || *upload.Rank != 2 {
		t.Errorf("unexpected rank. want=%d have=%s", 2, printableRank{upload.Rank})
	}
	if upload, _, _ := store.GetUploadByID(context.Background(), 5); upload.Rank == nil || *upload.Rank != 4 {
		t.Errorf("unexpected rank. want=%d have=%s", 4, printableRank{upload.Rank})
	}

	// Only considers queued uploads to determine rank
	if upload, _, _ := store.GetUploadByID(context.Background(), 6); upload.Rank != nil {
		t.Errorf("unexpected rank. want=%s have=%s", "nil", printableRank{upload.Rank})
	}

	// Process after takes priority over upload time
	if upload, _, _ := store.GetUploadByID(context.Background(), 7); upload.Rank == nil || *upload.Rank != 5 {
		t.Errorf("unexpected rank. want=%d have=%s", 4, printableRank{upload.Rank})
	}
}

func TestGetUploadsByIDs(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	insertUploads(t, db,
		shared.Upload{ID: 1},
		shared.Upload{ID: 2},
		shared.Upload{ID: 3},
		shared.Upload{ID: 4},
		shared.Upload{ID: 5},
		shared.Upload{ID: 6},
		shared.Upload{ID: 7},
		shared.Upload{ID: 8},
		shared.Upload{ID: 9},
		shared.Upload{ID: 10},
	)

	t.Run("fetch", func(t *testing.T) {
		indexes, err := store.GetUploadsByIDs(ctx, 2, 4, 6, 8, 12)
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

		indexes, err := store.GetUploadsByIDs(ctx, 1, 2, 3, 4)
		if err != nil {
			t.Fatal(err)
		}
		if len(indexes) > 0 {
			t.Fatalf("Want no index but got %d indexes", len(indexes))
		}
	})
}

func TestDeleteUploadsWithoutRepository(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	var uploads []shared.Upload
	for i := 0; i < 25; i++ {
		for j := 0; j < 10+i; j++ {
			uploads = append(uploads, shared.Upload{ID: len(uploads) + 1, RepositoryID: 50 + i})
		}
	}
	insertUploads(t, db, uploads...)

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

	deletedCounts, err := store.DeleteUploadsWithoutRepository(context.Background(), t1)
	if err != nil {
		t.Fatalf("unexpected error deleting uploads: %s", err)
	}

	expected := map[int]int{
		61: 21,
		63: 23,
		65: 25,
	}
	if diff := cmp.Diff(expected, deletedCounts); diff != "" {
		t.Errorf("unexpected deletedCounts (-want +got):\n%s", diff)
	}

	var uploadIDs []int
	for i := range uploads {
		uploadIDs = append(uploadIDs, i+1)
	}

	// Ensure records were deleted
	if states, err := getUploadStates(db, uploadIDs...); err != nil {
		t.Fatalf("unexpected error getting states: %s", err)
	} else {
		deletedStates := 0
		for _, state := range states {
			if state == "deleted" {
				deletedStates++
			}
		}

		expected := 0
		for _, deletedCount := range deletedCounts {
			expected += deletedCount
		}

		if deletedStates != expected {
			t.Errorf("unexpected number of deleted records. want=%d have=%d", expected, deletedStates)
		}
	}
}

func TestRecentUploadsSummary(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

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

	addDefaults := func(upload shared.Upload) shared.Upload {
		upload.Commit = makeCommit(upload.ID)
		upload.RepositoryID = 50
		upload.RepositoryName = "n-50"
		upload.IndexerVersion = "latest"
		upload.UploadedParts = []int{}
		return upload
	}

	uploads := []shared.Upload{
		addDefaults(shared.Upload{ID: 150, UploadedAt: t0, Root: "r1", Indexer: "i1", State: "queued", Rank: &r2}), // visible (group 1)
		addDefaults(shared.Upload{ID: 151, UploadedAt: t1, Root: "r1", Indexer: "i1", State: "queued", Rank: &r1}), // visible (group 1)
		addDefaults(shared.Upload{ID: 152, FinishedAt: &t2, Root: "r1", Indexer: "i1", State: "errored"}),          // visible (group 1)
		addDefaults(shared.Upload{ID: 153, FinishedAt: &t3, Root: "r1", Indexer: "i2", State: "completed"}),        // visible (group 2)
		addDefaults(shared.Upload{ID: 154, FinishedAt: &t4, Root: "r2", Indexer: "i1", State: "completed"}),        // visible (group 3)
		addDefaults(shared.Upload{ID: 155, FinishedAt: &t5, Root: "r2", Indexer: "i1", State: "errored"}),          // shadowed
		addDefaults(shared.Upload{ID: 156, FinishedAt: &t6, Root: "r2", Indexer: "i2", State: "completed"}),        // visible (group 4)
		addDefaults(shared.Upload{ID: 157, FinishedAt: &t7, Root: "r2", Indexer: "i2", State: "errored"}),          // shadowed
		addDefaults(shared.Upload{ID: 158, FinishedAt: &t8, Root: "r2", Indexer: "i2", State: "errored"}),          // shadowed
		addDefaults(shared.Upload{ID: 159, FinishedAt: &t9, Root: "r2", Indexer: "i2", State: "errored"}),          // shadowed
	}
	insertUploads(t, db, uploads...)

	summary, err := store.GetRecentUploadsSummary(ctx, 50)
	if err != nil {
		t.Fatalf("unexpected error querying recent upload summary: %s", err)
	}

	expected := []shared.UploadsWithRepositoryNamespace{
		{Root: "r1", Indexer: "i1", Uploads: []shared.Upload{uploads[0], uploads[1], uploads[2]}},
		{Root: "r1", Indexer: "i2", Uploads: []shared.Upload{uploads[3]}},
		{Root: "r2", Indexer: "i1", Uploads: []shared.Upload{uploads[4]}},
		{Root: "r2", Indexer: "i2", Uploads: []shared.Upload{uploads[6]}},
	}
	if diff := cmp.Diff(expected, summary); diff != "" {
		t.Errorf("unexpected upload summary (-want +got):\n%s", diff)
	}
}

func TestDeleteUploadsStuckUploading(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(time.Minute * 1)
	t3 := t1.Add(time.Minute * 2)
	t4 := t1.Add(time.Minute * 3)
	t5 := t1.Add(time.Minute * 4)

	insertUploads(t, db,
		shared.Upload{ID: 1, Commit: makeCommit(1111), UploadedAt: t1, State: "queued"},    // not uploading
		shared.Upload{ID: 2, Commit: makeCommit(1112), UploadedAt: t2, State: "uploading"}, // deleted
		shared.Upload{ID: 3, Commit: makeCommit(1113), UploadedAt: t3, State: "uploading"}, // deleted
		shared.Upload{ID: 4, Commit: makeCommit(1114), UploadedAt: t4, State: "completed"}, // old, not uploading
		shared.Upload{ID: 5, Commit: makeCommit(1115), UploadedAt: t5, State: "uploading"}, // old
	)

	count, err := store.DeleteUploadsStuckUploading(context.Background(), t1.Add(time.Minute*3))
	if err != nil {
		t.Fatalf("unexpected error deleting uploads stuck uploading: %s", err)
	}
	if count != 2 {
		t.Errorf("unexpected count. want=%d have=%d", 2, count)
	}

	uploads, totalCount, err := store.GetUploads(context.Background(), shared.GetUploadsOptions{Limit: 5})
	if err != nil {
		t.Fatalf("unexpected error getting uploads: %s", err)
	}

	var ids []int
	for _, upload := range uploads {
		ids = append(ids, upload.ID)
	}
	sort.Ints(ids)

	expectedIDs := []int{1, 4, 5}

	if totalCount != len(expectedIDs) {
		t.Errorf("unexpected total count. want=%d have=%d", len(expectedIDs), totalCount)
	}
	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Errorf("unexpected upload ids (-want +got):\n%s", diff)
	}
}

func TestHardDeleteUploadsByIDs(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	insertUploads(t, db,
		shared.Upload{ID: 51, State: "completed"},
		shared.Upload{ID: 52, State: "completed"},
		shared.Upload{ID: 53, State: "completed"},
		shared.Upload{ID: 54, State: "completed"},
	)
	insertPackages(t, store, []shared.Package{
		{DumpID: 52, Scheme: "test", Name: "p1", Version: "1.2.3"},
		{DumpID: 53, Scheme: "test", Name: "p2", Version: "1.2.3"},
	})
	insertPackageReferences(t, store, []shared.PackageReference{
		{Package: shared.Package{DumpID: 51, Scheme: "test", Name: "p1", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 51, Scheme: "test", Name: "p2", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 54, Scheme: "test", Name: "p1", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 54, Scheme: "test", Name: "p2", Version: "1.2.3"}},
	})

	if _, err := store.UpdateUploadsReferenceCounts(context.Background(), []int{51, 52, 53, 54}, shared.DependencyReferenceCountUpdateTypeNone); err != nil {
		t.Fatalf("unexpected error updating reference counts: %s", err)
	}
	assertReferenceCounts(t, db, map[int]int{
		51: 0,
		52: 2, // referenced by 51, 54
		53: 2, // referenced by 51, 52
		54: 0,
	})

	if err := store.HardDeleteUploadsByIDs(context.Background(), 51); err != nil {
		t.Fatalf("unexpected error deleting upload: %s", err)
	}
	assertReferenceCounts(t, db, map[int]int{
		// 51 was deleted
		52: 1, // referenced by 54
		53: 1, // referenced by 54
		54: 0,
	})
}

func TestBackfillReferenceCountBatch(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	n := 150
	expectedReferenceCounts := make([]int, 0, n)
	for i := 0; i < n; i++ {
		expectedReferenceCounts = append(expectedReferenceCounts, n-i-1)
	}

	insertQuery := sqlf.Sprintf("INSERT INTO repo (id, name) VALUES (42, 'foo'), (43, 'bar')")
	if _, err := db.ExecContext(context.Background(), insertQuery.Query(sqlf.PostgresBindVar), insertQuery.Args()...); err != nil {
		t.Fatalf("unexpected error inserting repo: %s", err)
	}

	for i := 0; i < n; i++ {
		insertQuery := sqlf.Sprintf(
			"INSERT INTO lsif_uploads (repository_id, commit, state, indexer, num_parts, uploaded_parts) VALUES (%s, %s, 'completed', 'lsif-go', 0, '{}')",
			42+i/(n/2), // 50% id=42, 50% id=43
			fmt.Sprintf("%040d", i),
		)
		if _, err := db.ExecContext(context.Background(), insertQuery.Query(sqlf.PostgresBindVar), insertQuery.Args()...); err != nil {
			t.Fatalf("unexpected error inserting upload: %s", err)
		}

		insertQuery = sqlf.Sprintf(
			"INSERT INTO lsif_packages (scheme, name, version, dump_id) VALUES ('test', %s, '1.2.3', %s)",
			fmt.Sprintf("pkg-%03d", i),
			i+1,
		)
		if _, err := db.ExecContext(context.Background(), insertQuery.Query(sqlf.PostgresBindVar), insertQuery.Args()...); err != nil {
			t.Fatalf("unexpected error inserting upload: %s", err)
		}

		for j := i - 1; j >= 0; j-- {
			insertQuery := sqlf.Sprintf(
				"INSERT INTO lsif_references (scheme, name, version, dump_id) VALUES ('test', %s, '1.2.3', %s)",
				fmt.Sprintf("pkg-%03d", j),
				i+1,
			)
			if _, err := db.ExecContext(context.Background(), insertQuery.Query(sqlf.PostgresBindVar), insertQuery.Args()...); err != nil {
				t.Fatalf("unexpected error inserting upload: %s", err)
			}
		}
	}

	if err := store.BackfillReferenceCountBatch(context.Background(), n/2); err != nil {
		t.Fatalf("unexpected error performing up migration: %s", err)
	}
	referenceCountQuery := sqlf.Sprintf("SELECT u.reference_count FROM lsif_uploads u WHERE u.reference_count IS NOT NULL ORDER BY u.id")
	if referenceCounts, err := basestore.ScanInts(db.QueryContext(context.Background(), referenceCountQuery.Query(sqlf.PostgresBindVar), referenceCountQuery.Args()...)); err != nil {
		t.Fatalf("unexpected error querying uploads: %s", err)
	} else if diff := cmp.Diff(expectedReferenceCounts[:n/2], referenceCounts); diff != "" {
		t.Errorf("unexpected reference counts (-want +got):\n%s", diff)
	}

	if err := store.BackfillReferenceCountBatch(context.Background(), n/2); err != nil {
		t.Fatalf("unexpected error performing up migration: %s", err)
	}
	if referenceCounts, err := basestore.ScanInts(db.QueryContext(context.Background(), referenceCountQuery.Query(sqlf.PostgresBindVar), referenceCountQuery.Args()...)); err != nil {
		t.Fatalf("unexpected error querying uploads: %s", err)
	} else if diff := cmp.Diff(expectedReferenceCounts, referenceCounts); diff != "" {
		t.Errorf("unexpected reference counts (-want +got):\n%s", diff)
	}
}

func TestSourcedCommitsWithoutCommittedAt(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	now := time.Unix(1587396557, 0).UTC()

	insertUploads(t, db,
		shared.Upload{ID: 1, RepositoryID: 50, Commit: makeCommit(1), State: "completed"},
		shared.Upload{ID: 2, RepositoryID: 50, Commit: makeCommit(1), State: "completed", Root: "sub/"},
		shared.Upload{ID: 3, RepositoryID: 51, Commit: makeCommit(4), State: "completed"},
		shared.Upload{ID: 4, RepositoryID: 51, Commit: makeCommit(5), State: "completed"},
		shared.Upload{ID: 5, RepositoryID: 52, Commit: makeCommit(7), State: "completed"},
		shared.Upload{ID: 6, RepositoryID: 52, Commit: makeCommit(8), State: "completed"},
	)

	sourcedCommits, err := store.SourcedCommitsWithoutCommittedAt(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error getting stale sourced commits: %s", err)
	}
	expectedCommits := []shared.SourcedCommits{
		{RepositoryID: 50, RepositoryName: "n-50", Commits: []string{makeCommit(1)}},
		{RepositoryID: 51, RepositoryName: "n-51", Commits: []string{makeCommit(4), makeCommit(5)}},
		{RepositoryID: 52, RepositoryName: "n-52", Commits: []string{makeCommit(7), makeCommit(8)}},
	}
	if diff := cmp.Diff(expectedCommits, sourcedCommits); diff != "" {
		t.Errorf("unexpected sourced commits (-want +got):\n%s", diff)
	}

	// Update commits 1 and 4
	if err := store.UpdateCommittedAt(context.Background(), 50, makeCommit(1), now.Format(time.RFC3339)); err != nil {
		t.Fatalf("unexpected error refreshing commit resolvability: %s", err)
	}
	if err := store.UpdateCommittedAt(context.Background(), 51, makeCommit(4), now.Format(time.RFC3339)); err != nil {
		t.Fatalf("unexpected error refreshing commit resolvability: %s", err)
	}

	sourcedCommits, err = store.SourcedCommitsWithoutCommittedAt(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error getting stale sourced commits: %s", err)
	}
	expectedCommits = []shared.SourcedCommits{
		{RepositoryID: 51, RepositoryName: "n-51", Commits: []string{makeCommit(5)}},
		{RepositoryID: 52, RepositoryName: "n-52", Commits: []string{makeCommit(7), makeCommit(8)}},
	}
	if diff := cmp.Diff(expectedCommits, sourcedCommits); diff != "" {
		t.Errorf("unexpected sourced commits (-want +got):\n%s", diff)
	}
}

func TestSoftDeleteExpiredUploads(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	insertUploads(t, db,
		shared.Upload{ID: 50, State: "completed"},
		shared.Upload{ID: 51, State: "completed"},
		shared.Upload{ID: 52, State: "completed"},
		shared.Upload{ID: 53, State: "completed"}, // referenced by 51, 52, 54, 55, 56
		shared.Upload{ID: 54, State: "completed"}, // referenced by 52
		shared.Upload{ID: 55, State: "completed"}, // referenced by 51
		shared.Upload{ID: 56, State: "completed"}, // referenced by 52, 53
	)
	insertPackages(t, store, []shared.Package{
		{DumpID: 53, Scheme: "test", Name: "p1", Version: "1.2.3"},
		{DumpID: 54, Scheme: "test", Name: "p2", Version: "1.2.3"},
		{DumpID: 55, Scheme: "test", Name: "p3", Version: "1.2.3"},
		{DumpID: 56, Scheme: "test", Name: "p4", Version: "1.2.3"},
	})
	insertPackageReferences(t, store, []shared.PackageReference{
		// References removed
		{Package: shared.Package{DumpID: 51, Scheme: "test", Name: "p1", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 51, Scheme: "test", Name: "p2", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 51, Scheme: "test", Name: "p3", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 52, Scheme: "test", Name: "p1", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 52, Scheme: "test", Name: "p4", Version: "1.2.3"}},

		// Remaining references
		{Package: shared.Package{DumpID: 53, Scheme: "test", Name: "p4", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 54, Scheme: "test", Name: "p1", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 55, Scheme: "test", Name: "p1", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 56, Scheme: "test", Name: "p1", Version: "1.2.3"}},
	})

	if err := store.UpdateUploadRetention(context.Background(), []int{}, []int{51, 52, 53, 54}); err != nil {
		t.Fatalf("unexpected error marking uploads as expired: %s", err)
	}

	if _, err := store.UpdateUploadsReferenceCounts(context.Background(), []int{50, 51, 52, 53, 54, 55, 56}, shared.DependencyReferenceCountUpdateTypeAdd); err != nil {
		t.Fatalf("unexpected error updating reference counts: %s", err)
	}

	if count, err := store.SoftDeleteExpiredUploads(context.Background()); err != nil {
		t.Fatalf("unexpected error soft deleting uploads: %s", err)
	} else if count != 2 {
		t.Fatalf("unexpected number of uploads deleted: want=%d have=%d", 2, count)
	}

	// Ensure records were deleted
	expectedStates := map[int]string{
		50: "completed",
		51: "deleting",
		52: "deleting",
		53: "completed",
		54: "completed",
		55: "completed",
		56: "completed",
	}
	if states, err := getUploadStates(db, 50, 51, 52, 53, 54, 55, 56); err != nil {
		t.Fatalf("unexpected error getting states: %s", err)
	} else if diff := cmp.Diff(expectedStates, states); diff != "" {
		t.Errorf("unexpected upload states (-want +got):\n%s", diff)
	}

	// Ensure repository was marked as dirty
	repositoryIDs, err := store.GetDirtyRepositories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error listing dirty repositories: %s", err)
	}

	var keys []int
	for repositoryID := range repositoryIDs {
		keys = append(keys, repositoryID)
	}
	sort.Ints(keys)

	if len(keys) != 1 || keys[0] != 50 {
		t.Errorf("expected repository to be marked dirty")
	}
}

func TestDeleteUploadByID(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	insertUploads(t, db,
		shared.Upload{ID: 1, RepositoryID: 50},
	)

	if found, err := store.DeleteUploadByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error deleting upload: %s", err)
	} else if !found {
		t.Fatalf("expected record to exist")
	}

	// Ensure record was deleted
	if states, err := getUploadStates(db, 1); err != nil {
		t.Fatalf("unexpected error getting states: %s", err)
	} else if diff := cmp.Diff(map[int]string{1: "deleting"}, states); diff != "" {
		t.Errorf("unexpected dump (-want +got):\n%s", diff)
	}

	repositoryIDs, err := store.GetDirtyRepositories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error listing dirty repositories: %s", err)
	}

	var keys []int
	for repositoryID := range repositoryIDs {
		keys = append(keys, repositoryID)
	}
	sort.Ints(keys)

	if len(keys) != 1 || keys[0] != 50 {
		t.Errorf("expected repository to be marked dirty")
	}
}

func TestDeleteUploadByIDNotCompleted(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	insertUploads(t, db,
		shared.Upload{ID: 1, RepositoryID: 50, State: "uploading"},
	)

	if found, err := store.DeleteUploadByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error deleting upload: %s", err)
	} else if !found {
		t.Fatalf("expected record to exist")
	}

	// Ensure record was deleted
	if states, err := getUploadStates(db, 1); err != nil {
		t.Fatalf("unexpected error getting states: %s", err)
	} else if diff := cmp.Diff(map[int]string{1: "deleted"}, states); diff != "" {
		t.Errorf("unexpected dump (-want +got):\n%s", diff)
	}

	repositoryIDs, err := store.GetDirtyRepositories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error listing dirty repositories: %s", err)
	}

	var keys []int
	for repositoryID := range repositoryIDs {
		keys = append(keys, repositoryID)
	}
	sort.Ints(keys)

	if len(keys) != 1 || keys[0] != 50 {
		t.Errorf("expected repository to be marked dirty")
	}
}

func TestDeleteUploadByIDMissingRow(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	if found, err := store.DeleteUploadByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error deleting upload: %s", err)
	} else if found {
		t.Fatalf("unexpected record")
	}
}

func TestUpdateUploadsVisibleToCommits(t *testing.T) {
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

	refDescriptions := map[string][]gitdomain.RefDescription{
		makeCommit(8): {{IsDefaultBranch: true}},
	}

	if err := store.UpdateUploadsVisibleToCommits(context.Background(), 50, graph, refDescriptions, time.Hour, time.Hour, 0, time.Now()); err != nil {
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

func TestUpdateUploadsVisibleToCommitsAlternateCommitGraph(t *testing.T) {
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

	refDescriptions := map[string][]gitdomain.RefDescription{
		makeCommit(3): {{IsDefaultBranch: true}},
	}

	if err := store.UpdateUploadsVisibleToCommits(context.Background(), 50, graph, refDescriptions, time.Hour, time.Hour, 0, time.Now()); err != nil {
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

func TestUpdateUploadsVisibleToCommitsDistinctRoots(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	// This database has the following commit graph:
	//
	// 1 -- [2]

	uploads := []shared.Upload{
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

	if err := store.UpdateUploadsVisibleToCommits(context.Background(), 50, graph, refDescriptions, time.Hour, time.Hour, 0, time.Now()); err != nil {
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

func TestUpdateUploadsVisibleToCommitsOverlappingRoots(t *testing.T) {
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

	refDescriptions := map[string][]gitdomain.RefDescription{
		makeCommit(6): {{IsDefaultBranch: true}},
	}

	if err := store.UpdateUploadsVisibleToCommits(context.Background(), 50, graph, refDescriptions, time.Hour, time.Hour, 0, time.Now()); err != nil {
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

func TestUpdateUploadsVisibleToCommitsIndexerName(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	// This database has the following commit graph:
	//
	// [1] -- [2] -- [3] -- [4] -- 5

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

	refDescriptions := map[string][]gitdomain.RefDescription{
		makeCommit(5): {{IsDefaultBranch: true}},
	}

	if err := store.UpdateUploadsVisibleToCommits(context.Background(), 50, graph, refDescriptions, time.Hour, time.Hour, 0, time.Now()); err != nil {
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

func TestUpdateUploadsVisibleToCommitsResetsDirtyFlag(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	uploads := []shared.Upload{
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
		if err := store.SetRepositoryAsDirty(context.Background(), 50); err != nil {
			t.Fatalf("unexpected error marking repository as dirty: %s", err)
		}
	}

	now := time.Unix(1587396557, 0).UTC()

	// Non-latest dirty token - should not clear flag
	if err := store.UpdateUploadsVisibleToCommits(context.Background(), 50, graph, refDescriptions, time.Hour, time.Hour, 2, now); err != nil {
		t.Fatalf("unexpected error while calculating visible uploads: %s", err)
	}
	repositoryIDs, err := store.GetDirtyRepositories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error listing dirty repositories: %s", err)
	}
	if len(repositoryIDs) == 0 {
		t.Errorf("did not expect repository to be unmarked")
	}

	// Latest dirty token - should clear flag
	if err := store.UpdateUploadsVisibleToCommits(context.Background(), 50, graph, refDescriptions, time.Hour, time.Hour, 3, now); err != nil {
		t.Fatalf("unexpected error while calculating visible uploads: %s", err)
	}
	repositoryIDs, err = store.GetDirtyRepositories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error listing dirty repositories: %s", err)
	}
	if len(repositoryIDs) != 0 {
		t.Errorf("expected repository to be unmarked")
	}

	stale, updatedAt, err := store.GetCommitGraphMetadata(context.Background(), 50)
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
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	uploads := []shared.Upload{
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
		if err := store.SetRepositoryAsDirty(context.Background(), 50); err != nil {
			t.Fatalf("unexpected error marking repository as dirty: %s", err)
		}
	}

	// This test is mainly a syntax check against `transaction_timestamp()`
	if err := store.UpdateUploadsVisibleToCommits(context.Background(), 50, graph, refDescriptions, time.Hour, time.Hour, 3, time.Now()); err != nil {
		t.Fatalf("unexpected error while calculating visible uploads: %s", err)
	}
}

func TestCalculateVisibleUploadsNonDefaultBranches(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

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

	uploads := []shared.Upload{
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

	if err := store.UpdateUploadsVisibleToCommits(context.Background(), 50, graph, refDescriptions, time.Hour, time.Hour, 0, time.Now()); err != nil {
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
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

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

	uploads := []shared.Upload{
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
	if _, err := db.ExecContext(context.Background(), retentionConfigurationQuery); err != nil {
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

	if err := store.UpdateUploadsVisibleToCommits(context.Background(), 50, graph, refDescriptions, time.Second, time.Second, 0, time.Now()); err != nil {
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

func TestGetVisibleUploadsMatchingMonikers(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	insertUploads(t, db,
		shared.Upload{ID: 1, Commit: makeCommit(2), Root: "sub1/"},
		shared.Upload{ID: 2, Commit: makeCommit(3), Root: "sub2/"},
		shared.Upload{ID: 3, Commit: makeCommit(4), Root: "sub3/"},
		shared.Upload{ID: 4, Commit: makeCommit(3), Root: "sub4/"},
		shared.Upload{ID: 5, Commit: makeCommit(2), Root: "sub5/"},
	)

	insertNearestUploads(t, db, 50, map[string][]commitgraph.UploadMeta{
		makeCommit(1): {
			{UploadID: 1, Distance: 1},
			{UploadID: 2, Distance: 2},
			{UploadID: 3, Distance: 3},
			{UploadID: 4, Distance: 2},
			{UploadID: 5, Distance: 1},
		},
		makeCommit(2): {
			{UploadID: 1, Distance: 0},
			{UploadID: 2, Distance: 1},
			{UploadID: 3, Distance: 2},
			{UploadID: 4, Distance: 1},
			{UploadID: 5, Distance: 0},
		},
		makeCommit(3): {
			{UploadID: 1, Distance: 1},
			{UploadID: 2, Distance: 0},
			{UploadID: 3, Distance: 1},
			{UploadID: 4, Distance: 0},
			{UploadID: 5, Distance: 1},
		},
		makeCommit(4): {
			{UploadID: 1, Distance: 2},
			{UploadID: 2, Distance: 1},
			{UploadID: 3, Distance: 0},
			{UploadID: 4, Distance: 1},
			{UploadID: 5, Distance: 2},
		},
	})

	insertPackageReferences(t, store, []shared.PackageReference{
		{Package: shared.Package{DumpID: 1, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
		{Package: shared.Package{DumpID: 2, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
		{Package: shared.Package{DumpID: 3, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
		{Package: shared.Package{DumpID: 4, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
		{Package: shared.Package{DumpID: 5, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
	})

	moniker := precise.QualifiedMonikerData{
		MonikerData: precise.MonikerData{
			Scheme: "gomod",
		},
		PackageInformationData: precise.PackageInformationData{
			Name:    "leftpad",
			Version: "0.1.0",
		},
	}

	refs := []shared.PackageReference{
		{Package: shared.Package{DumpID: 1, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
		{Package: shared.Package{DumpID: 2, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
		{Package: shared.Package{DumpID: 3, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
		{Package: shared.Package{DumpID: 4, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
		{Package: shared.Package{DumpID: 5, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
	}

	testCases := []struct {
		limit    int
		offset   int
		expected []shared.PackageReference
	}{
		{5, 0, refs},
		{5, 2, refs[2:]},
		{2, 1, refs[1:3]},
		{5, 5, nil},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("i=%d", i), func(t *testing.T) {
			scanner, totalCount, err := store.GetVisibleUploadsMatchingMonikers(context.Background(), 50, makeCommit(1), []precise.QualifiedMonikerData{moniker}, testCase.limit, testCase.offset)
			if err != nil {
				t.Fatalf("unexpected error getting scanner: %s", err)
			}

			if totalCount != 5 {
				t.Errorf("unexpected count. want=%d have=%d", 5, totalCount)
			}

			filters, err := consumeScanner(scanner)
			if err != nil {
				t.Fatalf("unexpected error from scanner: %s", err)
			}

			if diff := cmp.Diff(testCase.expected, filters); diff != "" {
				t.Errorf("unexpected filters (-want +got):\n%s", diff)
			}
		})
	}

	t.Run("enforce repository permissions", func(t *testing.T) {
		// Enable permissions user mapping forces checking repository permissions
		// against permissions tables in the database, which should effectively block
		// all access because permissions tables are empty.
		before := globals.PermissionsUserMapping()
		globals.SetPermissionsUserMapping(&schema.PermissionsUserMapping{Enabled: true})
		defer globals.SetPermissionsUserMapping(before)

		_, totalCount, err := store.GetVisibleUploadsMatchingMonikers(context.Background(), 50, makeCommit(1), []precise.QualifiedMonikerData{moniker}, 50, 0)
		if err != nil {
			t.Fatalf("unexpected error getting filters: %s", err)
		}
		if totalCount != 0 {
			t.Errorf("unexpected count. want=%d have=%d", 0, totalCount)
		}
	})
}

func TestCommitGraphMetadata(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	if err := store.SetRepositoryAsDirty(context.Background(), 50); err != nil {
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
			stale, updatedAt, err := store.GetCommitGraphMetadata(context.Background(), testCase.RepositoryID)
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

// consumeScanner reads all values from the scanner into memory.
func consumeScanner(scanner shared.PackageReferenceScanner) (references []shared.PackageReference, _ error) {
	for {
		reference, exists, err := scanner.Next()
		if err != nil {
			return nil, err
		}
		if !exists {
			break
		}

		references = append(references, reference)
	}
	if err := scanner.Close(); err != nil {
		return nil, err
	}

	return references, nil
}

// intsToQueries converts a slice of ints into a slice of queries.
func intsToQueries(values []int) []*sqlf.Query {
	var queries []*sqlf.Query
	for _, value := range values {
		queries = append(queries, sqlf.Sprintf("%d", value))
	}

	return queries
}

func getProtectedUploads(t testing.TB, db database.DB, repositoryID int) []int {
	query := sqlf.Sprintf(
		`SELECT DISTINCT upload_id FROM lsif_uploads_visible_at_tip WHERE repository_id = %s ORDER BY upload_id`,
		repositoryID,
	)

	ids, err := basestore.ScanInts(db.QueryContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...))
	if err != nil {
		t.Fatalf("unexpected error getting protected uploads: %s", err)
	}

	return ids
}

func assertReferenceCounts(t *testing.T, store database.DB, expectedReferenceCountsByID map[int]int) {
	db := basestore.NewWithHandle(store.Handle())

	referenceCountsByID, err := scanIntPairs(db.Query(context.Background(), sqlf.Sprintf(`SELECT id, reference_count FROM lsif_uploads`)))
	if err != nil {
		t.Fatalf("unexpected error querying reference counts: %s", err)
	}

	if diff := cmp.Diff(expectedReferenceCountsByID, referenceCountsByID); diff != "" {
		t.Errorf("unexpected reference count (-want +got):\n%s", diff)
	}
}

// insertVisibleAtTip populates rows of the lsif_uploads_visible_at_tip table for the given repository
// with the given identifiers. Each upload is assumed to refer to the tip of the default branch. To mark
// an upload as protected (visible to _some_ branch) butn ot visible from the default branch, use the
// insertVisibleAtTipNonDefaultBranch method instead.
func insertVisibleAtTip(t testing.TB, db database.DB, repositoryID int, uploadIDs ...int) {
	insertVisibleAtTipInternal(t, db, repositoryID, true, uploadIDs...)
}

// insertVisibleAtTipNonDefaultBranch populates rows of the lsif_uploads_visible_at_tip table for the
// given repository with the given identifiers. Each upload is assumed to refer to the tip of a branch
// distinct from the default branch or a tag.
func insertVisibleAtTipNonDefaultBranch(t testing.TB, db database.DB, repositoryID int, uploadIDs ...int) {
	insertVisibleAtTipInternal(t, db, repositoryID, false, uploadIDs...)
}

func insertVisibleAtTipInternal(t testing.TB, db database.DB, repositoryID int, isDefaultBranch bool, uploadIDs ...int) {
	var rows []*sqlf.Query
	for _, uploadID := range uploadIDs {
		rows = append(rows, sqlf.Sprintf("(%s, %s, %s)", repositoryID, uploadID, isDefaultBranch))
	}

	query := sqlf.Sprintf(
		`INSERT INTO lsif_uploads_visible_at_tip (repository_id, upload_id, is_default_branch) VALUES %s`,
		sqlf.Join(rows, ","),
	)
	if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error while updating uploads visible at tip: %s", err)
	}
}

//nolint:unparam // unparam complains that `repositoryID` always has same value across call-sites, but that's OK
func getVisibleUploads(t testing.TB, db database.DB, repositoryID int, commits []string) map[string][]int {
	idsByCommit := map[string][]int{}
	for _, commit := range commits {
		query := makeVisibleUploadsQuery(repositoryID, commit)

		uploadIDs, err := basestore.ScanInts(db.QueryContext(
			context.Background(),
			query.Query(sqlf.PostgresBindVar),
			query.Args()...,
		))
		if err != nil {
			t.Fatalf("unexpected error getting visible upload IDs: %s", err)
		}
		sort.Ints(uploadIDs)

		idsByCommit[commit] = uploadIDs
	}

	return idsByCommit
}

//nolint:unparam // unparam complains that `repositoryID` always has same value across call-sites, but that's OK
func getUploadsVisibleAtTip(t testing.TB, db database.DB, repositoryID int) []int {
	query := sqlf.Sprintf(
		`SELECT upload_id FROM lsif_uploads_visible_at_tip WHERE repository_id = %s AND is_default_branch ORDER BY upload_id`,
		repositoryID,
	)

	ids, err := basestore.ScanInts(db.QueryContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...))
	if err != nil {
		t.Fatalf("unexpected error getting uploads visible at tip: %s", err)
	}

	return ids
}

func assertCommitsVisibleFromUploads(t *testing.T, store Store, uploads []shared.Upload, expectedVisibleUploads map[string][]int) {
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
			commits, nextToken, err := store.GetCommitsVisibleToUpload(context.Background(), upload.ID, testPageSize, token)
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
	logger := logtest.Scoped(b)
	db := database.NewDB(logger, dbtest.NewDB(logger, b))
	store := New(db, &observation.TestContext)

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

	if err := store.UpdateUploadsVisibleToCommits(context.Background(), 50, graph, refDescriptions, time.Hour, time.Hour, 0, time.Now()); err != nil {
		b.Fatalf("unexpected error while calculating visible uploads: %s", err)
	}
}

func readBenchmarkCommitGraph() (*gitdomain.CommitGraph, error) {
	contents, err := readBenchmarkFile("../../../commitgraph/testdata/customer1/commits.txt.gz")
	if err != nil {
		return nil, err
	}

	return gitdomain.ParseCommitGraph(strings.Split(string(contents), "\n")), nil
}

func readBenchmarkCommitGraphView() ([]shared.Upload, error) {
	// contents, err := readBenchmarkFile("../../../commitgraph/testdata/uploads.txt.gz")
	contents, err := readBenchmarkFile("../../../../codeintel/commitgraph/testdata/customer1/uploads.csv.gz")
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(bytes.NewReader(contents))

	var uploads []shared.Upload
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

		uploads = append(uploads, shared.Upload{
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

type printableRank struct{ value *int }

func (r printableRank) String() string {
	if r.value == nil {
		return "nil"
	}
	return strconv.Itoa(*r.value)
}
