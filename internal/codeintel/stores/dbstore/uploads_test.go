package dbstore

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestInsertUploadUploading(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(db)

	insertRepo(t, db, 50, "")

	id, err := store.InsertUpload(context.Background(), Upload{
		Commit:       makeCommit(1),
		Root:         "sub/",
		State:        "uploading",
		RepositoryID: 50,
		Indexer:      "lsif-go",
		NumParts:     3,
	})
	if err != nil {
		t.Fatalf("unexpected error enqueueing upload: %s", err)
	}

	expected := Upload{
		ID:             id,
		Commit:         makeCommit(1),
		Root:           "sub/",
		VisibleAtTip:   false,
		UploadedAt:     time.Time{},
		State:          "uploading",
		FailureMessage: nil,
		StartedAt:      nil,
		FinishedAt:     nil,
		RepositoryID:   50,
		RepositoryName: "n-50",
		Indexer:        "lsif-go",
		NumParts:       3,
		UploadedParts:  []int{},
	}

	if upload, exists, err := store.GetUploadByID(context.Background(), id); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if !exists {
		t.Fatal("expected record to exist")
	} else {
		// Update auto-generated timestamp
		expected.UploadedAt = upload.UploadedAt

		if diff := cmp.Diff(expected, upload); diff != "" {
			t.Errorf("unexpected upload (-want +got):\n%s", diff)
		}
	}
}

func TestInsertUploadQueued(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(db)

	insertRepo(t, db, 50, "")

	id, err := store.InsertUpload(context.Background(), Upload{
		Commit:        makeCommit(1),
		Root:          "sub/",
		State:         "queued",
		RepositoryID:  50,
		Indexer:       "lsif-go",
		NumParts:      1,
		UploadedParts: []int{0},
	})
	if err != nil {
		t.Fatalf("unexpected error enqueueing upload: %s", err)
	}

	rank := 1
	expected := Upload{
		ID:             id,
		Commit:         makeCommit(1),
		Root:           "sub/",
		VisibleAtTip:   false,
		UploadedAt:     time.Time{},
		State:          "queued",
		FailureMessage: nil,
		StartedAt:      nil,
		FinishedAt:     nil,
		RepositoryID:   50,
		RepositoryName: "n-50",
		Indexer:        "lsif-go",
		NumParts:       1,
		UploadedParts:  []int{0},
		Rank:           &rank,
	}

	if upload, exists, err := store.GetUploadByID(context.Background(), id); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if !exists {
		t.Fatal("expected record to exist")
	} else {
		// Update auto-generated timestamp
		expected.UploadedAt = upload.UploadedAt

		if diff := cmp.Diff(expected, upload); diff != "" {
			t.Errorf("unexpected upload (-want +got):\n%s", diff)
		}
	}
}

func TestInsertUploadWithAssociatedIndexID(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(db)

	insertRepo(t, db, 50, "")

	associatedIndexIDArg := 42
	id, err := store.InsertUpload(context.Background(), Upload{
		Commit:            makeCommit(1),
		Root:              "sub/",
		State:             "queued",
		RepositoryID:      50,
		Indexer:           "lsif-go",
		NumParts:          1,
		UploadedParts:     []int{0},
		AssociatedIndexID: &associatedIndexIDArg,
	})
	if err != nil {
		t.Fatalf("unexpected error enqueueing upload: %s", err)
	}

	rank := 1
	associatedIndexIDResult := 42
	expected := Upload{
		ID:                id,
		Commit:            makeCommit(1),
		Root:              "sub/",
		VisibleAtTip:      false,
		UploadedAt:        time.Time{},
		State:             "queued",
		FailureMessage:    nil,
		StartedAt:         nil,
		FinishedAt:        nil,
		RepositoryID:      50,
		RepositoryName:    "n-50",
		Indexer:           "lsif-go",
		NumParts:          1,
		UploadedParts:     []int{0},
		Rank:              &rank,
		AssociatedIndexID: &associatedIndexIDResult,
	}

	if upload, exists, err := store.GetUploadByID(context.Background(), id); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if !exists {
		t.Fatal("expected record to exist")
	} else {
		// Update auto-generated timestamp
		expected.UploadedAt = upload.UploadedAt

		if diff := cmp.Diff(expected, upload); diff != "" {
			t.Errorf("unexpected upload (-want +got):\n%s", diff)
		}
	}
}

func TestMarkQueued(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(db)

	insertUploads(t, db, Upload{ID: 1, State: "uploading"})

	uploadSize := int64(300)
	if err := store.MarkQueued(context.Background(), 1, &uploadSize); err != nil {
		t.Fatalf("unexpected error marking upload as queued: %s", err)
	}

	if upload, exists, err := store.GetUploadByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if !exists {
		t.Fatal("expected record to exist")
	} else if upload.State != "queued" {
		t.Errorf("unexpected state. want=%q have=%q", "queued", upload.State)
	} else if upload.UploadSize == nil || *upload.UploadSize != 300 {
		if upload.UploadSize == nil {
			t.Errorf("unexpected upload size. want=%v have=%v", 300, upload.UploadSize)
		} else {
			t.Errorf("unexpected upload size. want=%v have=%v", 300, *upload.UploadSize)
		}
	}
}

func TestMarkFailed(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(db)

	insertUploads(t, db, Upload{ID: 1, State: "uploading"})

	failureReason := "didn't like it"
	if err := store.MarkFailed(context.Background(), 1, failureReason); err != nil {
		t.Fatalf("unexpected error marking upload as failed: %s", err)
	}

	if upload, exists, err := store.GetUploadByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if !exists {
		t.Fatal("expected record to exist")
	} else if upload.State != "failed" {
		t.Errorf("unexpected state. want=%q have=%q", "failed", upload.State)
	} else if upload.NumFailures != 1 {
		t.Errorf("unexpected num failures. want=%v have=%v", 1, upload.NumFailures)
	} else if upload.FailureMessage == nil || *upload.FailureMessage != failureReason {
		if upload.FailureMessage == nil {
			t.Errorf("unexpected failure message. want='%s' have='%v'", failureReason, upload.FailureMessage)
		} else {
			t.Errorf("unexpected failure message. want='%s' have='%v'", failureReason, *upload.FailureMessage)
		}
	}
}

func TestAddUploadPart(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(db)

	insertUploads(t, db, Upload{ID: 1, State: "uploading"})

	for _, part := range []int{1, 5, 2, 3, 2, 2, 1, 6} {
		if err := store.AddUploadPart(context.Background(), 1, part); err != nil {
			t.Fatalf("unexpected error adding upload part: %s", err)
		}
	}
	if upload, exists, err := store.GetUploadByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if !exists {
		t.Fatal("expected record to exist")
	} else {
		sort.Ints(upload.UploadedParts)
		if diff := cmp.Diff([]int{1, 2, 3, 5, 6}, upload.UploadedParts); diff != "" {
			t.Errorf("unexpected upload parts (-want +got):\n%s", diff)
		}
	}
}

func TestHardDeleteUploadByID(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(db)

	insertUploads(t, db,
		Upload{ID: 51, State: "completed"},
		Upload{ID: 52, State: "completed"},
		Upload{ID: 53, State: "completed"},
		Upload{ID: 54, State: "completed"},
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

	if _, err := store.UpdateReferenceCounts(context.Background(), []int{51, 52, 53, 54}, DependencyReferenceCountUpdateTypeNone); err != nil {
		t.Fatalf("unexpected error updating reference counts: %s", err)
	}
	assertReferenceCounts(t, store, map[int]int{
		51: 0,
		52: 2, // referenced by 51, 54
		53: 2, // referenced by 51, 52
		54: 0,
	})

	if err := store.HardDeleteUploadByID(context.Background(), 51); err != nil {
		t.Fatalf("unexpected error deleting upload: %s", err)
	}
	assertReferenceCounts(t, store, map[int]int{
		// 51 was deleted
		52: 1, // referenced by 54
		53: 1, // referenced by 54
		54: 0,
	})
}

func TestHardDeleteUploadByIDPackageProvider(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(db)

	insertUploads(t, db,
		Upload{ID: 51, State: "completed"},
		Upload{ID: 52, State: "completed"},
		Upload{ID: 53, State: "completed"},
		Upload{ID: 54, State: "completed"},
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

	if _, err := store.UpdateReferenceCounts(context.Background(), []int{51, 52, 53, 54}, DependencyReferenceCountUpdateTypeNone); err != nil {
		t.Fatalf("unexpected error updating reference counts: %s", err)
	}
	assertReferenceCounts(t, store, map[int]int{
		51: 0,
		52: 2, // referenced by 51, 54
		53: 2, // referenced by 51, 54
		54: 0,
	})

	if err := store.HardDeleteUploadByID(context.Background(), 52); err != nil {
		t.Fatalf("unexpected error deleting upload: %s", err)
	}
	assertReferenceCounts(t, store, map[int]int{
		51: 0,
		// 52 was deleted
		53: 2, // referenced by 51, 54
		54: 0,
	})
}

func TestHardDeleteUploadByIDDuplicatePackageProvider(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(db)

	insertUploads(t, db,
		Upload{ID: 51, State: "completed"},
		Upload{ID: 52, State: "completed"},
		Upload{ID: 53, State: "completed"},
		Upload{ID: 54, State: "completed"},
		Upload{ID: 55, State: "completed"},
	)
	insertPackages(t, store, []shared.Package{
		{DumpID: 52, Scheme: "test", Name: "p1", Version: "1.2.3"},
		{DumpID: 53, Scheme: "test", Name: "p2", Version: "1.2.3"},
		{DumpID: 54, Scheme: "test", Name: "p1", Version: "1.2.3"},
		{DumpID: 55, Scheme: "test", Name: "p2", Version: "1.2.3"},
	})
	insertPackageReferences(t, store, []shared.PackageReference{
		{Package: shared.Package{DumpID: 51, Scheme: "test", Name: "p1", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 52, Scheme: "test", Name: "p2", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 53, Scheme: "test", Name: "p1", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 54, Scheme: "test", Name: "p2", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 55, Scheme: "test", Name: "p1", Version: "1.2.3"}},
	})

	if _, err := store.UpdateReferenceCounts(context.Background(), []int{51, 52, 53, 54, 55}, DependencyReferenceCountUpdateTypeNone); err != nil {
		t.Fatalf("unexpected error updating reference counts: %s", err)
	}
	assertReferenceCounts(t, store, map[int]int{
		51: 0,
		52: 3, // referenced by 51, 53, 55
		53: 2, // referenced by 52, 54
		54: 0,
		55: 0,
	})

	if err := store.HardDeleteUploadByID(context.Background(), 52); err != nil {
		t.Fatalf("unexpected error deleting upload: %s", err)
	}
	assertReferenceCounts(t, store, map[int]int{
		51: 0,
		// 52 was deleted
		53: 1, // referenced by 54
		54: 3, // referenced by 51, 53, 55
		55: 0,
	})
}

func TestUpdateUploadRetention(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(db)

	insertUploads(t, db,
		Upload{ID: 1, State: "completed"},
		Upload{ID: 2, State: "completed"},
		Upload{ID: 3, State: "completed"},
		Upload{ID: 4, State: "completed"},
		Upload{ID: 5, State: "completed"},
	)

	now := timeutil.Now()

	if err := store.updateUploadRetention(context.Background(), []int{}, []int{2, 3, 4}, now); err != nil {
		t.Fatalf("unexpected error marking uploads as expired: %s", err)
	}

	count, _, err := basestore.ScanFirstInt(db.QueryContext(context.Background(), `SELECT COUNT(*) FROM lsif_uploads WHERE expired`))
	if err != nil {
		t.Fatalf("unexpected error counting uploads: %s", err)
	}

	if count != 3 {
		t.Fatalf("unexpected count. want=%d have=%d", 3, count)
	}
}

func TestUpdateReferenceCounts(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(db)

	insertUploads(t, db,
		Upload{ID: 50, State: "completed"},
		Upload{ID: 51, State: "completed"},
		Upload{ID: 52, State: "completed"},
		Upload{ID: 53, State: "completed"},
		Upload{ID: 54, State: "completed"},
		Upload{ID: 55, State: "completed"},
		Upload{ID: 56, State: "completed"},
	)
	insertPackages(t, store, []shared.Package{
		{DumpID: 53, Scheme: "test", Name: "p1", Version: "1.2.3"},
		{DumpID: 54, Scheme: "test", Name: "p2", Version: "1.2.3"},
		{DumpID: 55, Scheme: "test", Name: "p3", Version: "1.2.3"},
		{DumpID: 56, Scheme: "test", Name: "p4", Version: "1.2.3"},
	})
	insertPackageReferences(t, store, []shared.PackageReference{
		{Package: shared.Package{DumpID: 51, Scheme: "test", Name: "p1", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 51, Scheme: "test", Name: "p2", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 51, Scheme: "test", Name: "p3", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 52, Scheme: "test", Name: "p1", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 52, Scheme: "test", Name: "p4", Version: "1.2.3"}},

		{Package: shared.Package{DumpID: 53, Scheme: "test", Name: "p4", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 54, Scheme: "test", Name: "p1", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 55, Scheme: "test", Name: "p1", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 56, Scheme: "test", Name: "p1", Version: "1.2.3"}},
		{Package: shared.Package{DumpID: 56, Scheme: "test", Name: "p3", Version: "1.2.4"}}, // future version
	})

	if _, err := store.UpdateReferenceCounts(context.Background(), []int{50, 51, 52, 53, 54, 55, 56}, DependencyReferenceCountUpdateTypeNone); err != nil {
		t.Fatalf("unexpected error updating reference counts: %s", err)
	}

	assertReferenceCounts(t, store, map[int]int{
		50: 0,
		51: 0,
		52: 0,
		53: 5, // referenced by 51, 52, 54, 55, 56
		54: 1, // referenced by 52
		55: 1, // referenced by 51
		56: 2, // referenced by 52, 53
	})

	t.Run("add uploads", func(t *testing.T) {
		insertUploads(t, db,
			Upload{ID: 62, State: "completed"},
			Upload{ID: 63, State: "completed"},
			Upload{ID: 64, State: "completed"},
		)
		insertPackages(t, store, []shared.Package{
			{DumpID: 62, Scheme: "test", Name: "p1", Version: "1.2.3"}, // duplicate version
			{DumpID: 63, Scheme: "test", Name: "p2", Version: "1.2.3"}, // duplicate version
			{DumpID: 64, Scheme: "test", Name: "p3", Version: "1.2.4"}, // new version
		})

		// Update commit dates so that the newly inserted uploads come first
		// in the commit graph. We use a heuristic to select the "oldest" upload
		// as the canonical provider ofa package for the same repository and root.
		// This ensures that we "usurp" the package provider with a younger upload.

		query := `
			UPDATE lsif_uploads
			SET committed_at = CASE
				WHEN id < 60 THEN NOW()
				ELSE              NOW() - '1 day'::interval
			END
		`
		if _, err := db.ExecContext(context.Background(), query); err != nil {
			t.Fatalf("unexpected error updating upload commit date: %s", err)
		}

		if _, err := store.UpdateReferenceCounts(context.Background(), []int{62, 63, 64}, DependencyReferenceCountUpdateTypeAdd); err != nil {
			t.Fatalf("unexpected error updating reference counts: %s", err)
		}

		assertReferenceCounts(t, store, map[int]int{
			50: 0,
			51: 0,
			52: 0,
			53: 0, // usurped by 62
			54: 0, // usurped by 63
			55: 1, // referenced by 51
			56: 2, // referenced by 52, 53
			62: 5, // referenced by 51, 52, 54, 55, 56 (usurped from 53)
			63: 1, // referenced by 52                 (usurped from 54)
			64: 1, // referenced by 56
		})
	})

	t.Run("remove uploads", func(t *testing.T) {
		if _, err := store.UpdateReferenceCounts(context.Background(), []int{53, 56, 63, 64}, DependencyReferenceCountUpdateTypeRemove); err != nil {
			t.Fatalf("unexpected error updating reference counts: %s", err)
		}

		if _, err := db.ExecContext(context.Background(), `DELETE FROM lsif_uploads WHERE id IN (53, 56, 63, 64)`); err != nil {
			t.Fatalf("unexpected error deleting uploads: %s", err)
		}

		assertReferenceCounts(t, store, map[int]int{
			50: 0,
			51: 0,
			52: 0,
			// 53 deleted
			54: 1, // referenced by 52             (usurped from 63)
			55: 1, // referenced by 51
			// 56 deleted
			62: 4, // referenced by 51, 52, 54, 55 (usurped from 53)
			// 63 deleted
			// 64 deleted
		})
	})
}

func TestUpdateCommitedAt(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(db)

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(time.Minute)
	t3 := t1.Add(time.Minute * 4)
	t4 := t1.Add(time.Minute * 6)

	insertUploads(t, db,
		Upload{ID: 1, State: "completed"},
		Upload{ID: 2, State: "completed"},
		Upload{ID: 3, State: "completed"},
		Upload{ID: 4, State: "completed"},
		Upload{ID: 5, State: "completed"},
		Upload{ID: 6, State: "completed"},
		Upload{ID: 7, State: "completed"},
		Upload{ID: 8, State: "completed"},
	)

	for uploadID, commitDate := range map[int]time.Time{
		1: t3,
		2: t4,
		4: t1,
		6: t2,
	} {
		if err := store.UpdateCommitedAt(context.Background(), uploadID, commitDate); err != nil {
			t.Fatalf("unexpected error updating commit date %s", err)
		}
	}

	commitDates, err := basestore.ScanTimes(db.QueryContext(context.Background(), "SELECT committed_at FROM lsif_uploads WHERE id IN (1, 2, 4, 6) ORDER BY id"))
	if err != nil {
		t.Fatalf("unexpected error querying commit dates: %s", err)
	}
	if diff := cmp.Diff([]time.Time{t3, t4, t1, t2}, commitDates); diff != "" {
		t.Errorf("unexpected commit dates(-want +got):\n%s", diff)
	}
}

func TestLastUploadRetentionScanForRepository(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(db)
	ctx := context.Background()

	ts, err := store.LastUploadRetentionScanForRepository(ctx, 50)
	if err != nil {
		t.Fatalf("unexpected error querying last upload retention scan: %s", err)
	}
	if ts != nil {
		t.Fatalf("unexpected timestamp for repository. want=%v have=%s", nil, ts)
	}

	expected := time.Unix(1587396557, 0).UTC()

	if err := store.Exec(ctx, sqlf.Sprintf(`
		INSERT INTO lsif_last_retention_scan (repository_id, last_retention_scan_at)
		VALUES (%s, %s)
	`, 50, expected)); err != nil {
		t.Fatalf("unexpected error inserting timestamp: %s", err)
	}

	ts, err = store.LastUploadRetentionScanForRepository(ctx, 50)
	if err != nil {
		t.Fatalf("unexpected error querying last upload retention scan: %s", err)
	}
	if ts == nil || !ts.Equal(expected) {
		t.Fatalf("unexpected timestamp for repository. want=%s have=%s", expected, ts)
	}
}
