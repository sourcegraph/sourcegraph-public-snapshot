package db

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestSameRepoPager(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	insertUploads(t, dbconn.Global,
		Upload{ID: 1, Commit: makeCommit(2), Root: "sub1/"},
		Upload{ID: 2, Commit: makeCommit(3), Root: "sub2/"},
		Upload{ID: 3, Commit: makeCommit(4), Root: "sub3/"},
		Upload{ID: 4, Commit: makeCommit(3), Root: "sub4/"},
		Upload{ID: 5, Commit: makeCommit(2), Root: "sub5/"},
	)

	expected := []types.PackageReference{
		{DumpID: 1, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f1")},
		{DumpID: 2, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f2")},
		{DumpID: 3, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f3")},
		{DumpID: 4, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f4")},
		{DumpID: 5, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f5")},
	}
	insertPackageReferences(t, db, expected)

	if err := db.UpdateCommits(context.Background(), 50, map[string][]string{
		makeCommit(1): {},
		makeCommit(2): {makeCommit(1)},
		makeCommit(3): {makeCommit(2)},
		makeCommit(4): {makeCommit(3)},
	}); err != nil {
		t.Fatalf("unexpected error updating commits: %s", err)
	}

	totalCount, pager, err := db.SameRepoPager(context.Background(), 50, makeCommit(1), "gomod", "leftpad", "0.1.0", 5)
	if err != nil {
		t.Fatalf("unexpected error getting pager: %s", err)
	}
	defer func() { _ = pager.Done(nil) }()

	if totalCount != 5 {
		t.Errorf("unexpected dump. want=%d have=%d", 5, totalCount)
	}

	if references, err := pager.PageFromOffset(context.Background(), 0); err != nil {
		t.Fatalf("unexpected error getting next page: %s", err)
	} else if diff := cmp.Diff(expected, references); diff != "" {
		t.Errorf("unexpected references (-want +got):\n%s", diff)
	}
}

func TestSameRepoPagerEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	totalCount, pager, err := db.SameRepoPager(context.Background(), 50, makeCommit(1), "gomod", "leftpad", "0.1.0", 5)
	if err != nil {
		t.Fatalf("unexpected error getting pager: %s", err)
	}
	defer func() { _ = pager.Done(nil) }()

	if totalCount != 0 {
		t.Errorf("unexpected dump. want=%d have=%d", 0, totalCount)
	}
}

func TestSameRepoPagerMultiplePages(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	insertUploads(t, dbconn.Global,
		Upload{ID: 1, Commit: makeCommit(1), Root: "sub1/"},
		Upload{ID: 2, Commit: makeCommit(1), Root: "sub2/"},
		Upload{ID: 3, Commit: makeCommit(1), Root: "sub3/"},
		Upload{ID: 4, Commit: makeCommit(1), Root: "sub4/"},
		Upload{ID: 5, Commit: makeCommit(1), Root: "sub5/"},
		Upload{ID: 6, Commit: makeCommit(1), Root: "sub6/"},
		Upload{ID: 7, Commit: makeCommit(1), Root: "sub7/"},
		Upload{ID: 8, Commit: makeCommit(1), Root: "sub8/"},
		Upload{ID: 9, Commit: makeCommit(1), Root: "sub9/"},
	)

	expected := []types.PackageReference{
		{DumpID: 1, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f1")},
		{DumpID: 2, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f2")},
		{DumpID: 3, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f3")},
		{DumpID: 4, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f4")},
		{DumpID: 5, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f5")},
		{DumpID: 6, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f6")},
		{DumpID: 7, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f7")},
		{DumpID: 8, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f8")},
		{DumpID: 9, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f9")},
	}
	insertPackageReferences(t, db, expected)

	if err := db.UpdateCommits(context.Background(), 50, map[string][]string{
		makeCommit(1): {},
	}); err != nil {
		t.Fatalf("unexpected error updating commits: %s", err)
	}

	totalCount, pager, err := db.SameRepoPager(context.Background(), 50, makeCommit(1), "gomod", "leftpad", "0.1.0", 3)
	if err != nil {
		t.Fatalf("unexpected error getting pager: %s", err)
	}
	defer func() { _ = pager.Done(nil) }()

	if totalCount != 9 {
		t.Errorf("unexpected dump. want=%d have=%d", 9, totalCount)
	}

	for lo := 0; lo < len(expected); lo++ {
		hi := lo + 3
		if hi > len(expected) {
			hi = len(expected)
		}

		if references, err := pager.PageFromOffset(context.Background(), lo); err != nil {
			t.Fatalf("unexpected error getting page at offset %d: %s", lo, err)
		} else if diff := cmp.Diff(expected[lo:hi], references); diff != "" {
			t.Errorf("unexpected references at offset %d (-want +got):\n%s", lo, diff)
		}
	}
}

func TestSameRepoPagerVisibility(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	insertUploads(t, dbconn.Global,
		Upload{ID: 1, Commit: makeCommit(1), Root: "sub1/"}, // not visible
		Upload{ID: 2, Commit: makeCommit(2), Root: "sub2/"}, // not visible
		Upload{ID: 3, Commit: makeCommit(3), Root: "sub1/"},
		Upload{ID: 4, Commit: makeCommit(4), Root: "sub2/"},
		Upload{ID: 5, Commit: makeCommit(5), Root: "sub5/"},
	)

	expected := []types.PackageReference{
		{DumpID: 3, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f3")},
		{DumpID: 4, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f4")},
		{DumpID: 5, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f5")},
	}
	insertPackageReferences(t, db, append([]types.PackageReference{
		{DumpID: 1, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f1")},
		{DumpID: 2, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f2")},
	}, expected...))

	if err := db.UpdateCommits(context.Background(), 50, map[string][]string{
		makeCommit(1): {},
		makeCommit(2): {makeCommit(1)},
		makeCommit(3): {makeCommit(2)},
		makeCommit(4): {makeCommit(3)},
		makeCommit(5): {makeCommit(4)},
		makeCommit(6): {makeCommit(5)},
	}); err != nil {
		t.Fatalf("unexpected error updating commits: %s", err)
	}

	totalCount, pager, err := db.SameRepoPager(context.Background(), 50, makeCommit(6), "gomod", "leftpad", "0.1.0", 5)
	if err != nil {
		t.Fatalf("unexpected error getting pager: %s", err)
	}
	defer func() { _ = pager.Done(nil) }()

	if totalCount != 3 {
		t.Errorf("unexpected dump. want=%d have=%d", 5, totalCount)
	}

	if references, err := pager.PageFromOffset(context.Background(), 0); err != nil {
		t.Fatalf("unexpected error getting next page: %s", err)
	} else if diff := cmp.Diff(expected, references); diff != "" {
		t.Errorf("unexpected references (-want +got):\n%s", diff)
	}
}

func TestPackageReferencePager(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	insertUploads(t, dbconn.Global,
		Upload{ID: 1, Commit: makeCommit(1), VisibleAtTip: true},
		Upload{ID: 2, Commit: makeCommit(2), VisibleAtTip: true, RepositoryID: 51},
		Upload{ID: 3, Commit: makeCommit(3), VisibleAtTip: true, RepositoryID: 52},
		Upload{ID: 4, Commit: makeCommit(4), VisibleAtTip: true, RepositoryID: 53},
		Upload{ID: 5, Commit: makeCommit(5), VisibleAtTip: true, RepositoryID: 54},
		Upload{ID: 6, Commit: makeCommit(6), VisibleAtTip: false, RepositoryID: 55},
		Upload{ID: 7, Commit: makeCommit(6), VisibleAtTip: true, RepositoryID: 56},
	)

	expected := []types.PackageReference{
		{DumpID: 2, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f2")},
		{DumpID: 3, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f3")},
		{DumpID: 4, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f4")},
		{DumpID: 5, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f5")},
		{DumpID: 7, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f7")},
	}
	insertPackageReferences(t, db, append([]types.PackageReference{
		{DumpID: 1, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f1")},
		{DumpID: 6, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f6")},
	}, expected...))

	totalCount, pager, err := db.PackageReferencePager(context.Background(), "gomod", "leftpad", "0.1.0", 50, 5)
	if err != nil {
		t.Fatalf("unexpected error getting pager: %s", err)
	}
	defer func() { _ = pager.Done(nil) }()

	if totalCount != 5 {
		t.Errorf("unexpected dump. want=%d have=%d", 5, totalCount)
	}

	if references, err := pager.PageFromOffset(context.Background(), 0); err != nil {
		t.Fatalf("unexpected error getting next page: %s", err)
	} else if diff := cmp.Diff(expected, references); diff != "" {
		t.Errorf("unexpected references (-want +got):\n%s", diff)
	}
}

func TestPackageReferencePagerEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	totalCount, pager, err := db.PackageReferencePager(context.Background(), "gomod", "leftpad", "0.1.0", 50, 5)
	if err != nil {
		t.Fatalf("unexpected error getting pager: %s", err)
	}
	defer func() { _ = pager.Done(nil) }()

	if totalCount != 0 {
		t.Errorf("unexpected dump. want=%d have=%d", 0, totalCount)
	}
}

func TestPackageReferencePagerPages(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	insertUploads(t, dbconn.Global,
		Upload{ID: 1, Commit: makeCommit(1), VisibleAtTip: true, RepositoryID: 51},
		Upload{ID: 2, Commit: makeCommit(2), VisibleAtTip: true, RepositoryID: 52},
		Upload{ID: 3, Commit: makeCommit(3), VisibleAtTip: true, RepositoryID: 53},
		Upload{ID: 4, Commit: makeCommit(4), VisibleAtTip: true, RepositoryID: 54},
		Upload{ID: 5, Commit: makeCommit(5), VisibleAtTip: true, RepositoryID: 55},
		Upload{ID: 6, Commit: makeCommit(6), VisibleAtTip: true, RepositoryID: 56},
		Upload{ID: 7, Commit: makeCommit(7), VisibleAtTip: true, RepositoryID: 57},
		Upload{ID: 8, Commit: makeCommit(8), VisibleAtTip: true, RepositoryID: 58},
		Upload{ID: 9, Commit: makeCommit(9), VisibleAtTip: true, RepositoryID: 59},
	)

	expected := []types.PackageReference{
		{DumpID: 1, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f1")},
		{DumpID: 2, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f2")},
		{DumpID: 3, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f3")},
		{DumpID: 4, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f4")},
		{DumpID: 5, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f5")},
		{DumpID: 6, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f6")},
		{DumpID: 7, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f7")},
		{DumpID: 8, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f8")},
		{DumpID: 9, Scheme: "gomod", Name: "leftpad", Version: "0.1.0", Filter: []byte("f9")},
	}
	insertPackageReferences(t, db, expected)

	totalCount, pager, err := db.PackageReferencePager(context.Background(), "gomod", "leftpad", "0.1.0", 50, 3)
	if err != nil {
		t.Fatalf("unexpected error getting pager: %s", err)
	}
	defer func() { _ = pager.Done(nil) }()

	if totalCount != 9 {
		t.Errorf("unexpected dump. want=%d have=%d", 9, totalCount)
	}

	testCases := []struct {
		offset int
		lo     int
		hi     int
	}{
		{0, 0, 3},
		{1, 1, 4},
		{2, 2, 5},
		{3, 3, 6},
		{4, 4, 7},
		{5, 5, 8},
		{6, 6, 9},
		{7, 7, 9},
		{8, 8, 9},
	}

	for _, testCase := range testCases {
		if references, err := pager.PageFromOffset(context.Background(), testCase.offset); err != nil {
			t.Fatalf("unexpected error getting page at offset %d: %s", testCase.offset, err)
		} else if diff := cmp.Diff(expected[testCase.lo:testCase.hi], references); diff != "" {
			t.Errorf("unexpected references at offset %d (-want +got):\n%s", testCase.offset, diff)
		}
	}
}

func TestUpdatePackageReferences(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	// for foreign key relation
	insertUploads(t, dbconn.Global, Upload{ID: 42})

	if err := db.UpdatePackageReferences(context.Background(), []types.PackageReference{
		{DumpID: 42, Scheme: "s0", Name: "n0", Version: "v0"},
		{DumpID: 42, Scheme: "s1", Name: "n1", Version: "v1"},
		{DumpID: 42, Scheme: "s2", Name: "n2", Version: "v2"},
		{DumpID: 42, Scheme: "s3", Name: "n3", Version: "v3"},
		{DumpID: 42, Scheme: "s4", Name: "n4", Version: "v4"},
		{DumpID: 42, Scheme: "s5", Name: "n5", Version: "v5"},
		{DumpID: 42, Scheme: "s6", Name: "n6", Version: "v6"},
		{DumpID: 42, Scheme: "s7", Name: "n7", Version: "v7"},
		{DumpID: 42, Scheme: "s8", Name: "n8", Version: "v8"},
		{DumpID: 42, Scheme: "s9", Name: "n9", Version: "v9"},
	}); err != nil {
		t.Fatalf("unexpected error updating references: %s", err)
	}

	count, _, err := scanFirstInt(dbconn.Global.Query("SELECT COUNT(*) FROM lsif_references"))
	if err != nil {
		t.Fatalf("unexpected error checking reference count: %s", err)
	}
	if count != 10 {
		t.Errorf("unexpected reference count. want=%d have=%d", 10, count)
	}
}

func TestUpdatePackageReferencesEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	if err := db.UpdatePackageReferences(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error updating references: %s", err)
	}

	count, _, err := scanFirstInt(dbconn.Global.Query("SELECT COUNT(*) FROM lsif_references"))
	if err != nil {
		t.Fatalf("unexpected error checking reference count: %s", err)
	}
	if count != 0 {
		t.Errorf("unexpected reference count. want=%d have=%d", 0, count)
	}
}

func TestUpdatePackageReferencesWithDuplicates(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	// for foreign key relation
	insertUploads(t, dbconn.Global, Upload{ID: 42})

	if err := db.UpdatePackageReferences(context.Background(), []types.PackageReference{
		{DumpID: 42, Scheme: "s0", Name: "n0", Version: "v0"},
		{DumpID: 42, Scheme: "s1", Name: "n1", Version: "v1"},
		{DumpID: 42, Scheme: "s2", Name: "n2", Version: "v2"},
		{DumpID: 42, Scheme: "s3", Name: "n3", Version: "v3"},
	}); err != nil {
		t.Fatalf("unexpected error updating references: %s", err)
	}

	if err := db.UpdatePackageReferences(context.Background(), []types.PackageReference{
		{DumpID: 42, Scheme: "s0", Name: "n0", Version: "v0"}, // two copies
		{DumpID: 42, Scheme: "s2", Name: "n2", Version: "v2"}, // two copies
		{DumpID: 42, Scheme: "s4", Name: "n4", Version: "v4"},
		{DumpID: 42, Scheme: "s5", Name: "n5", Version: "v5"},
		{DumpID: 42, Scheme: "s6", Name: "n6", Version: "v6"},
		{DumpID: 42, Scheme: "s7", Name: "n7", Version: "v7"},
		{DumpID: 42, Scheme: "s8", Name: "n8", Version: "v8"},
		{DumpID: 42, Scheme: "s9", Name: "n9", Version: "v9"},
	}); err != nil {
		t.Fatalf("unexpected error updating references: %s", err)
	}

	count, _, err := scanFirstInt(dbconn.Global.Query("SELECT COUNT(*) FROM lsif_references"))
	if err != nil {
		t.Fatalf("unexpected error checking reference count: %s", err)
	}
	if count != 12 {
		t.Errorf("unexpected reference count. want=%d have=%d", 12, count)
	}
}
