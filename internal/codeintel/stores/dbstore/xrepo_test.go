package dbstore

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/commitgraph"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestDefinitionDumps(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

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
	if dumps, err := store.DefinitionDumps(context.Background(), []precise.QualifiedMonikerData{moniker1}); err != nil {
		t.Fatalf("unexpected error getting package: %s", err)
	} else if len(dumps) != 0 {
		t.Fatal("unexpected record")
	}

	uploadedAt := time.Unix(1587396557, 0).UTC()
	startedAt := uploadedAt.Add(time.Minute)
	finishedAt := uploadedAt.Add(time.Minute * 2)
	expected1 := Dump{
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
		Indexer:           "lsif-tsc",
		IndexerVersion:    "1.2.3",
		AssociatedIndexID: nil,
	}
	expected3 := Dump{
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

	if dumps, err := store.DefinitionDumps(context.Background(), []precise.QualifiedMonikerData{moniker1}); err != nil {
		t.Fatalf("unexpected error getting package: %s", err)
	} else if len(dumps) != 1 {
		t.Fatal("expected one record")
	} else if diff := cmp.Diff(expected1, dumps[0]); diff != "" {
		t.Errorf("unexpected dump (-want +got):\n%s", diff)
	}

	if dumps, err := store.DefinitionDumps(context.Background(), []precise.QualifiedMonikerData{moniker1, moniker2}); err != nil {
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

		if dumps, err := store.DefinitionDumps(context.Background(), []precise.QualifiedMonikerData{moniker1, moniker2}); err != nil {
			t.Fatalf("unexpected error getting package: %s", err)
		} else if len(dumps) != 0 {
			t.Errorf("unexpected count. want=%d have=%d", 0, len(dumps))
		}
	})
}

func TestReferenceIDs(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

	insertUploads(t, db,
		Upload{ID: 1, Commit: makeCommit(2), Root: "sub1/"},
		Upload{ID: 2, Commit: makeCommit(3), Root: "sub2/"},
		Upload{ID: 3, Commit: makeCommit(4), Root: "sub3/"},
		Upload{ID: 4, Commit: makeCommit(3), Root: "sub4/"},
		Upload{ID: 5, Commit: makeCommit(2), Root: "sub5/"},
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
			scanner, totalCount, err := store.ReferenceIDs(context.Background(), 50, makeCommit(1), []precise.QualifiedMonikerData{moniker}, testCase.limit, testCase.offset)
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

		_, totalCount, err := store.ReferenceIDs(context.Background(), 50, makeCommit(1), []precise.QualifiedMonikerData{moniker}, 50, 0)
		if err != nil {
			t.Fatalf("unexpected error getting filters: %s", err)
		}
		if totalCount != 0 {
			t.Errorf("unexpected count. want=%d have=%d", 0, totalCount)
		}
	})
}

func TestReferenceIDsVisibility(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

	insertUploads(t, db,
		Upload{ID: 1, Commit: makeCommit(1), Root: "sub1/"}, // not visible
		Upload{ID: 2, Commit: makeCommit(2), Root: "sub2/"}, // not visible
		Upload{ID: 3, Commit: makeCommit(3), Root: "sub1/"},
		Upload{ID: 4, Commit: makeCommit(4), Root: "sub2/"},
		Upload{ID: 5, Commit: makeCommit(5), Root: "sub5/"},
	)

	insertNearestUploads(t, db, 50, map[string][]commitgraph.UploadMeta{
		makeCommit(1): {{UploadID: 1, Distance: 0}},
		makeCommit(2): {{UploadID: 2, Distance: 0}},
		makeCommit(3): {{UploadID: 3, Distance: 0}},
		makeCommit(4): {{UploadID: 4, Distance: 0}},
		makeCommit(5): {{UploadID: 5, Distance: 0}},
		makeCommit(6): {{UploadID: 3, Distance: 3}, {UploadID: 4, Distance: 2}, {UploadID: 5, Distance: 1}},
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

	scanner, totalCount, err := store.ReferenceIDs(context.Background(), 50, makeCommit(6), []precise.QualifiedMonikerData{moniker}, 5, 0)
	if err != nil {
		t.Fatalf("unexpected error getting filters: %s", err)
	}

	if totalCount != 3 {
		t.Errorf("unexpected count. want=%d have=%d", 3, totalCount)
	}

	filters, err := consumeScanner(scanner)
	if err != nil {
		t.Fatalf("unexpected error from scanner: %s", err)
	}

	expected := []shared.PackageReference{
		{Package: shared.Package{DumpID: 3, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
		{Package: shared.Package{DumpID: 4, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
		{Package: shared.Package{DumpID: 5, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
	}
	if diff := cmp.Diff(expected, filters); diff != "" {
		t.Errorf("unexpected filters (-want +got):\n%s", diff)
	}
}

func TestReferenceIDsRemoteVisibility(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

	insertUploads(t, db,
		Upload{ID: 1, Commit: makeCommit(1)},
		Upload{ID: 2, Commit: makeCommit(2), RepositoryID: 51},
		Upload{ID: 3, Commit: makeCommit(3), RepositoryID: 52},
		Upload{ID: 4, Commit: makeCommit(4), RepositoryID: 53},
		Upload{ID: 5, Commit: makeCommit(5), RepositoryID: 54},
		Upload{ID: 6, Commit: makeCommit(6), RepositoryID: 55},
		Upload{ID: 7, Commit: makeCommit(6), RepositoryID: 56},
		Upload{ID: 8, Commit: makeCommit(7), RepositoryID: 57},
	)
	insertVisibleAtTip(t, db, 50, 1)
	insertVisibleAtTip(t, db, 51, 2)
	insertVisibleAtTip(t, db, 52, 3)
	insertVisibleAtTip(t, db, 53, 4)
	insertVisibleAtTip(t, db, 54, 5)
	insertVisibleAtTip(t, db, 56, 7)
	insertVisibleAtTipNonDefaultBranch(t, db, 57, 8)

	insertPackageReferences(t, store, []shared.PackageReference{
		{Package: shared.Package{DumpID: 1, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}}, // same repo, not visible in git
		{Package: shared.Package{DumpID: 2, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
		{Package: shared.Package{DumpID: 3, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
		{Package: shared.Package{DumpID: 4, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
		{Package: shared.Package{DumpID: 5, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
		{Package: shared.Package{DumpID: 6, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}}, // remote repo not visible at tip
		{Package: shared.Package{DumpID: 7, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
		{Package: shared.Package{DumpID: 8, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}}, // visible on non-default branch
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

	scanner, totalCount, err := store.ReferenceIDs(context.Background(), 50, makeCommit(6), []precise.QualifiedMonikerData{moniker}, 5, 0)
	if err != nil {
		t.Fatalf("unexpected error getting filters: %s", err)
	}

	if totalCount != 5 {
		t.Errorf("unexpected count. want=%d have=%d", 5, totalCount)
	}

	filters, err := consumeScanner(scanner)
	if err != nil {
		t.Fatalf("unexpected error from scanner: %s", err)
	}

	expected := []shared.PackageReference{
		{Package: shared.Package{DumpID: 2, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
		{Package: shared.Package{DumpID: 3, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
		{Package: shared.Package{DumpID: 4, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
		{Package: shared.Package{DumpID: 5, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
		{Package: shared.Package{DumpID: 7, Scheme: "gomod", Name: "leftpad", Version: "0.1.0"}},
	}
	if diff := cmp.Diff(expected, filters); diff != "" {
		t.Errorf("unexpected filters (-want +got):\n%s", diff)
	}
}

func TestReferencesForUpload(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

	insertUploads(t, db,
		Upload{ID: 1, Commit: makeCommit(2), Root: "sub1/"},
		Upload{ID: 2, Commit: makeCommit(3), Root: "sub2/"},
		Upload{ID: 3, Commit: makeCommit(4), Root: "sub3/"},
		Upload{ID: 4, Commit: makeCommit(3), Root: "sub4/"},
		Upload{ID: 5, Commit: makeCommit(2), Root: "sub5/"},
	)

	insertPackageReferences(t, store, []shared.PackageReference{
		{Package: shared.Package{DumpID: 1, Scheme: "gomod", Name: "leftpad", Version: "1.1.0"}},
		{Package: shared.Package{DumpID: 2, Scheme: "gomod", Name: "leftpad", Version: "2.1.0"}},
		{Package: shared.Package{DumpID: 2, Scheme: "gomod", Name: "leftpad", Version: "3.1.0"}},
		{Package: shared.Package{DumpID: 2, Scheme: "gomod", Name: "leftpad", Version: "4.1.0"}},
		{Package: shared.Package{DumpID: 3, Scheme: "gomod", Name: "leftpad", Version: "5.1.0"}},
	})

	scanner, err := store.ReferencesForUpload(context.Background(), 2)
	if err != nil {
		t.Fatalf("unexpected error getting filters: %s", err)
	}

	filters, err := consumeScanner(scanner)
	if err != nil {
		t.Fatalf("unexpected error from scanner: %s", err)
	}

	expected := []shared.PackageReference{
		{Package: shared.Package{DumpID: 2, Scheme: "gomod", Name: "leftpad", Version: "2.1.0"}},
		{Package: shared.Package{DumpID: 2, Scheme: "gomod", Name: "leftpad", Version: "3.1.0"}},
		{Package: shared.Package{DumpID: 2, Scheme: "gomod", Name: "leftpad", Version: "4.1.0"}},
	}
	if diff := cmp.Diff(expected, filters); diff != "" {
		t.Errorf("unexpected filters (-want +got):\n%s", diff)
	}
}

// consumeScanner reads all values from the scanner into memory.
func consumeScanner(scanner PackageReferenceScanner) (references []shared.PackageReference, _ error) {
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
