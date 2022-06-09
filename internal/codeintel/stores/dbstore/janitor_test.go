package dbstore

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestStaleSourcedCommits(t *testing.T) {
	sqlDB := dbtest.NewDB(t)
	db := database.NewDB(sqlDB)
	store := testStore(db)

	now := time.Unix(1587396557, 0).UTC()

	insertUploads(t, sqlDB,
		Upload{ID: 1, RepositoryID: 50, Commit: makeCommit(1)},
		Upload{ID: 2, RepositoryID: 50, Commit: makeCommit(1), Root: "sub/"},
		Upload{ID: 3, RepositoryID: 51, Commit: makeCommit(4)},
		Upload{ID: 4, RepositoryID: 51, Commit: makeCommit(5)},
		Upload{ID: 5, RepositoryID: 52, Commit: makeCommit(7)},
	)
	insertIndexes(t, sqlDB,
		Index{ID: 1, RepositoryID: 50, Commit: makeCommit(1)},
		Index{ID: 2, RepositoryID: 50, Commit: makeCommit(2)},
		Index{ID: 3, RepositoryID: 50, Commit: makeCommit(3)},
		Index{ID: 4, RepositoryID: 51, Commit: makeCommit(6)},
		Index{ID: 5, RepositoryID: 52, Commit: makeCommit(7)},
	)

	sourcedCommits, err := store.StaleSourcedCommits(context.Background(), time.Minute, 5, now)
	if err != nil {
		t.Fatalf("unexpected error getting stale sourced commits: %s", err)
	}
	expectedCommits := []SourcedCommits{
		{RepositoryID: 50, RepositoryName: "n-50", Commits: []string{makeCommit(1), makeCommit(2), makeCommit(3)}},
		{RepositoryID: 51, RepositoryName: "n-51", Commits: []string{makeCommit(4), makeCommit(5)}},
	}
	if diff := cmp.Diff(expectedCommits, sourcedCommits); diff != "" {
		t.Errorf("unexpected sourced commits (-want +got):\n%s", diff)
	}

	// 120s away from next check (threshold is 60s)
	if _, _, err := store.UpdateSourcedCommits(context.Background(), 50, makeCommit(1), now); err != nil {
		t.Fatalf("unexpected error refreshing commit resolvability: %s", err)
	}

	// 30s away from next check (threshold is 60s)
	if _, _, err := store.UpdateSourcedCommits(context.Background(), 50, makeCommit(2), now.Add(time.Second*90)); err != nil {
		t.Fatalf("unexpected error refreshing commit resolvability: %s", err)
	}

	sourcedCommits, err = store.StaleSourcedCommits(context.Background(), time.Minute, 5, now.Add(time.Minute*2))
	if err != nil {
		t.Fatalf("unexpected error getting stale sourced commits: %s", err)
	}
	expectedCommits = []SourcedCommits{
		{RepositoryID: 50, RepositoryName: "n-50", Commits: []string{makeCommit(1), makeCommit(3)}},
		{RepositoryID: 51, RepositoryName: "n-51", Commits: []string{makeCommit(4), makeCommit(5), makeCommit(6)}},
	}
	if diff := cmp.Diff(expectedCommits, sourcedCommits); diff != "" {
		t.Errorf("unexpected sourced commits (-want +got):\n%s", diff)
	}
}

func TestUpdateSourcedCommits(t *testing.T) {
	sqlDB := dbtest.NewDB(t)
	db := database.NewDB(sqlDB)
	store := testStore(db)

	now := time.Unix(1587396557, 0).UTC()

	insertUploads(t, sqlDB,
		Upload{ID: 1, RepositoryID: 50, Commit: makeCommit(1)},
		Upload{ID: 2, RepositoryID: 50, Commit: makeCommit(1), Root: "sub/"},
		Upload{ID: 3, RepositoryID: 51, Commit: makeCommit(4)},
		Upload{ID: 4, RepositoryID: 51, Commit: makeCommit(5)},
		Upload{ID: 5, RepositoryID: 52, Commit: makeCommit(7)},
		Upload{ID: 6, RepositoryID: 52, Commit: makeCommit(7), State: "uploading"},
	)
	insertIndexes(t, sqlDB,
		Index{ID: 1, RepositoryID: 50, Commit: makeCommit(3)},
		Index{ID: 2, RepositoryID: 50, Commit: makeCommit(2)},
		Index{ID: 3, RepositoryID: 52, Commit: makeCommit(7)},
		Index{ID: 4, RepositoryID: 51, Commit: makeCommit(6)},
		Index{ID: 5, RepositoryID: 50, Commit: makeCommit(1)},
	)

	uploadsUpdated, indexesUpdated, err := store.UpdateSourcedCommits(context.Background(), 50, makeCommit(1), now)
	if err != nil {
		t.Fatalf("unexpected error refreshing commit resolvability: %s", err)
	}
	if uploadsUpdated != 2 {
		t.Fatalf("unexpected uploads updated. want=%d have=%d", 2, uploadsUpdated)
	}
	if indexesUpdated != 1 {
		t.Fatalf("unexpected indexes updated. want=%d have=%d", 1, indexesUpdated)
	}

	uploadStates, err := getUploadStates(db, 1, 2, 3, 4, 5, 6)
	if err != nil {
		t.Fatalf("unexpected error fetching upload states: %s", err)
	}
	expectedUploadStates := map[int]string{
		1: "completed",
		2: "completed",
		3: "completed",
		4: "completed",
		5: "completed",
		6: "uploading",
	}
	if diff := cmp.Diff(expectedUploadStates, uploadStates); diff != "" {
		t.Errorf("unexpected upload states (-want +got):\n%s", diff)
	}

	indexStates, err := getIndexStates(db, 1, 2, 3, 4, 5)
	if err != nil {
		t.Fatalf("unexpected error fetching index states: %s", err)
	}
	expectedIndexStates := map[int]string{
		1: "completed",
		2: "completed",
		3: "completed",
		4: "completed",
		5: "completed",
	}
	if diff := cmp.Diff(expectedIndexStates, indexStates); diff != "" {
		t.Errorf("unexpected index states (-want +got):\n%s", diff)
	}
}

func TestDeleteSourcedCommits(t *testing.T) {
	sqlDB := dbtest.NewDB(t)
	db := database.NewDB(sqlDB)
	store := testStore(db)

	now := time.Unix(1587396557, 0).UTC()

	insertUploads(t, sqlDB,
		Upload{ID: 1, RepositoryID: 50, Commit: makeCommit(1)},
		Upload{ID: 2, RepositoryID: 50, Commit: makeCommit(1), Root: "sub/"},
		Upload{ID: 3, RepositoryID: 51, Commit: makeCommit(4)},
		Upload{ID: 4, RepositoryID: 51, Commit: makeCommit(5)},
		Upload{ID: 5, RepositoryID: 52, Commit: makeCommit(7)},
		Upload{ID: 6, RepositoryID: 52, Commit: makeCommit(7), State: "uploading", UploadedAt: now.Add(-time.Minute * 90)},
		Upload{ID: 7, RepositoryID: 52, Commit: makeCommit(7), State: "queued", UploadedAt: now.Add(-time.Minute * 30)},
	)
	insertIndexes(t, sqlDB,
		Index{ID: 1, RepositoryID: 50, Commit: makeCommit(3)},
		Index{ID: 2, RepositoryID: 50, Commit: makeCommit(2)},
		Index{ID: 3, RepositoryID: 52, Commit: makeCommit(7)},
		Index{ID: 4, RepositoryID: 51, Commit: makeCommit(6)},
		Index{ID: 5, RepositoryID: 50, Commit: makeCommit(1)},
	)

	uploadsUpdated, uploadsDeleted, indexesDeleted, err := store.DeleteSourcedCommits(context.Background(), 52, makeCommit(7), time.Hour, now)
	if err != nil {
		t.Fatalf("unexpected error refreshing commit resolvability: %s", err)
	}
	if uploadsUpdated != 1 {
		t.Fatalf("unexpected number of uploads updated. want=%d have=%d", 1, uploadsUpdated)
	}
	if uploadsDeleted != 2 {
		t.Fatalf("unexpected number of uploads deleted. want=%d have=%d", 2, uploadsDeleted)
	}
	if indexesDeleted != 1 {
		t.Fatalf("unexpected number of indexes deleted. want=%d have=%d", 1, indexesDeleted)
	}

	uploadStates, err := getUploadStates(db, 1, 2, 3, 4, 5, 6, 7)
	if err != nil {
		t.Fatalf("unexpected error fetching upload states: %s", err)
	}
	expectedUploadStates := map[int]string{
		1: "completed",
		2: "completed",
		3: "completed",
		4: "completed",
		5: "deleting",
		6: "deleted",
		7: "queued",
	}
	if diff := cmp.Diff(expectedUploadStates, uploadStates); diff != "" {
		t.Errorf("unexpected upload states (-want +got):\n%s", diff)
	}

	indexStates, err := getIndexStates(db, 1, 2, 3, 4, 5)
	if err != nil {
		t.Fatalf("unexpected error fetching index states: %s", err)
	}
	expectedIndexStates := map[int]string{
		1: "completed",
		2: "completed",
		// 3 was deleted
		4: "completed",
		5: "completed",
	}
	if diff := cmp.Diff(expectedIndexStates, indexStates); diff != "" {
		t.Errorf("unexpected index states (-want +got):\n%s", diff)
	}
}

func TestDeleteOldAuditLogs(t *testing.T) {
	sqlDB := dbtest.NewDB(t)
	db := database.NewDB(sqlDB)
	store := testStore(db)

	// Sanity check for syntax only
	if _, err := store.DeleteOldAuditLogs(context.Background(), time.Second, time.Now()); err != nil {
		t.Fatalf("unexpected error deleting old audit logs: %s", err)
	}
}
