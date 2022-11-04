package dbstore

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestGetDumpsByIDs(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(db)

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
	expected1 := Dump{
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
	expected2 := Dump{
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

func TestDeleteOverlappingDumps(t *testing.T) {
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, sqlDB)
	store := testStore(db)

	insertUploads(t, db, Upload{
		ID:      1,
		Commit:  makeCommit(1),
		Root:    "cmd/",
		Indexer: "lsif-go",
	})

	err := store.DeleteOverlappingDumps(context.Background(), 50, makeCommit(1), "cmd/", "lsif-go")
	if err != nil {
		t.Fatalf("unexpected error deleting dump: %s", err)
	}

	// Ensure record was deleted
	if states, err := getUploadStates(db, 1); err != nil {
		t.Fatalf("unexpected error getting states: %s", err)
	} else if diff := cmp.Diff(map[int]string{1: "deleting"}, states); diff != "" {
		t.Errorf("unexpected dump (-want +got):\n%s", diff)
	}
}

func TestDeleteOverlappingDumpsNoMatches(t *testing.T) {
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, sqlDB)
	store := testStore(db)

	insertUploads(t, db, Upload{
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
		{makeCommit(1), "cmd/", "scip-typescript"},
	}

	for _, testCase := range testCases {
		err := store.DeleteOverlappingDumps(context.Background(), 50, testCase.commit, testCase.root, testCase.indexer)
		if err != nil {
			t.Fatalf("unexpected error deleting dump: %s", err)
		}
	}

	// Original dump still exists
	if dumps, err := store.GetDumpsByIDs(context.Background(), []int{1}); err != nil {
		t.Fatalf("unexpected error getting dump: %s", err)
	} else if len(dumps) != 1 {
		t.Fatal("expected dump record to still exist")
	}
}

func TestDeleteOverlappingDumpsIgnoresIncompleteUploads(t *testing.T) {
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, sqlDB)
	store := testStore(db)

	insertUploads(t, db, Upload{
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
