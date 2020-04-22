package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestGetPackage(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	// Package does not exist initially
	if _, exists, err := db.GetPackage(context.Background(), "gomod", "leftpad", "0.1.0"); err != nil {
		t.Fatalf("unexpected error getting package: %s", err)
	} else if exists {
		t.Fatal("unexpected record")
	}

	uploadedAt := time.Unix(1587396557, 0).UTC()
	startedAt := uploadedAt.Add(time.Minute)
	finishedAt := uploadedAt.Add(time.Minute * 2)
	expected := Dump{
		ID:                1,
		Commit:            makeCommit(1),
		Root:              "sub/",
		VisibleAtTip:      true,
		UploadedAt:        uploadedAt,
		State:             "completed",
		FailureSummary:    nil,
		FailureStacktrace: nil,
		StartedAt:         &startedAt,
		FinishedAt:        &finishedAt,
		TracingContext:    `{"id": 42}`,
		RepositoryID:      50,
		Indexer:           "lsif-go",
	}

	insertUploads(t, db.db, Upload{
		ID:                expected.ID,
		Commit:            expected.Commit,
		Root:              expected.Root,
		VisibleAtTip:      expected.VisibleAtTip,
		UploadedAt:        expected.UploadedAt,
		State:             expected.State,
		FailureSummary:    expected.FailureSummary,
		FailureStacktrace: expected.FailureStacktrace,
		StartedAt:         expected.StartedAt,
		FinishedAt:        expected.FinishedAt,
		TracingContext:    expected.TracingContext,
		RepositoryID:      expected.RepositoryID,
		Indexer:           expected.Indexer,
	})

	insertPackages(t, db.db, PackageModel{
		Scheme:  "gomod",
		Name:    "leftpad",
		Version: "0.1.0",
		DumpID:  1,
	})

	if dump, exists, err := db.GetPackage(context.Background(), "gomod", "leftpad", "0.1.0"); err != nil {
		t.Fatalf("unexpected error getting package: %s", err)
	} else if !exists {
		t.Fatal("expected record to exist")
	} else if diff := cmp.Diff(expected, dump); diff != "" {
		t.Errorf("unexpected dump (-want +got):\n%s", diff)
	}
}
