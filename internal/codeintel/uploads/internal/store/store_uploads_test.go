package store

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

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

func TestDeleteUploadsStuckUploading(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	// store := testStore(db)
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

// intsToQueries converts a slice of ints into a slice of queries.
func intsToQueries(values []int) []*sqlf.Query {
	var queries []*sqlf.Query
	for _, value := range values {
		queries = append(queries, sqlf.Sprintf("%d", value))
	}

	return queries
}
