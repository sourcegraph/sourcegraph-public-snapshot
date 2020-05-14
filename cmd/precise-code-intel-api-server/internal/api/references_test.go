package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	bundlemocks "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/mocks"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db/mocks"
	dbmocks "github.com/sourcegraph/sourcegraph/internal/codeintel/db/mocks"
)

func TestHandleSameDumpCursor(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	mockBundleClient := bundlemocks.NewMockBundleClient()

	setMockDBGetDumpByID(t, mockDB, map[int]db.Dump{42: testDump1})
	setMockBundleManagerClientBundleClient(t, mockBundleManagerClient, map[int]bundles.BundleClient{42: mockBundleClient})
	setMockBundleClientReferences(t, mockBundleClient, "main.go", 23, 34, []bundles.Location{
		{DumpID: 42, Path: "foo.go", Range: testRange1},
		{DumpID: 42, Path: "foo.go", Range: testRange2},
		{DumpID: 42, Path: "foo.go", Range: testRange3},
		{DumpID: 42, Path: "foo.go", Range: testRange4},
		{DumpID: 42, Path: "foo.go", Range: testRange5},
		{DumpID: 42, Path: "bar.go", Range: testRange1},
		{DumpID: 42, Path: "bar.go", Range: testRange2},
		{DumpID: 42, Path: "bar.go", Range: testRange3},
		{DumpID: 42, Path: "bar.go", Range: testRange4},
	})

	rpr := &ReferencePageResolver{
		db:                  mockDB,
		bundleManagerClient: mockBundleManagerClient,
		repositoryID:        100,
		commit:              testCommit,
		limit:               5,
	}

	t.Run("partial results", func(t *testing.T) {
		references, newCursor, hasNewCursor, err := rpr.dispatchCursorHandler(context.Background(), Cursor{
			Phase:       "same-dump",
			DumpID:      42,
			Path:        "main.go",
			Line:        23,
			Character:   34,
			Monikers:    []bundles.MonikerData{{Kind: "export", Scheme: "gomod", Identifier: "pad"}},
			SkipResults: 0,
		})
		if err != nil {
			t.Fatalf("expected error getting references: %s", err)
		}

		expectedReferences := []ResolvedLocation{
			{Dump: testDump1, Path: "sub1/foo.go", Range: testRange1},
			{Dump: testDump1, Path: "sub1/foo.go", Range: testRange2},
			{Dump: testDump1, Path: "sub1/foo.go", Range: testRange3},
			{Dump: testDump1, Path: "sub1/foo.go", Range: testRange4},
			{Dump: testDump1, Path: "sub1/foo.go", Range: testRange5},
		}
		if diff := cmp.Diff(expectedReferences, references); diff != "" {
			t.Errorf("unexpected references (-want +got):\n%s", diff)
		}

		expectedNewCursor := Cursor{
			Phase:       "same-dump",
			DumpID:      42,
			Path:        "main.go",
			Line:        23,
			Character:   34,
			Monikers:    []bundles.MonikerData{{Kind: "export", Scheme: "gomod", Identifier: "pad"}},
			SkipResults: 5,
		}
		if !hasNewCursor {
			t.Errorf("expected new cursor")
		} else if diff := cmp.Diff(expectedNewCursor, newCursor); diff != "" {
			t.Errorf("unexpected new cursor (-want +got):\n%s", diff)
		}
	})

	t.Run("end of result set", func(t *testing.T) {
		references, newCursor, hasNewCursor, err := rpr.dispatchCursorHandler(context.Background(), Cursor{
			Phase:       "same-dump",
			DumpID:      42,
			Path:        "main.go",
			Line:        23,
			Character:   34,
			Monikers:    []bundles.MonikerData{{Kind: "export", Scheme: "gomod", Identifier: "pad"}},
			SkipResults: 5,
		})
		if err != nil {
			t.Fatalf("expected error getting references: %s", err)
		}

		expectedReferences := []ResolvedLocation{
			{Dump: testDump1, Path: "sub1/bar.go", Range: testRange1},
			{Dump: testDump1, Path: "sub1/bar.go", Range: testRange2},
			{Dump: testDump1, Path: "sub1/bar.go", Range: testRange3},
			{Dump: testDump1, Path: "sub1/bar.go", Range: testRange4},
		}
		if diff := cmp.Diff(expectedReferences, references); diff != "" {
			t.Errorf("unexpected references (-want +got):\n%s", diff)
		}

		expectedNewCursor := Cursor{
			Phase:       "same-dump-monikers",
			DumpID:      42,
			Path:        "main.go",
			Line:        23,
			Character:   34,
			Monikers:    []bundles.MonikerData{{Kind: "export", Scheme: "gomod", Identifier: "pad"}},
			SkipResults: 0,
		}
		if !hasNewCursor {
			t.Errorf("expected new cursor")
		} else if diff := cmp.Diff(expectedNewCursor, newCursor); diff != "" {
			t.Errorf("unexpected new cursor (-want +got):\n%s", diff)
		}
	})
}

func TestHandleSameDumpMonikersCursor(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	mockBundleClient := bundlemocks.NewMockBundleClient()

	setMockDBGetDumpByID(t, mockDB, map[int]db.Dump{42: testDump1})
	setMockBundleManagerClientBundleClient(t, mockBundleManagerClient, map[int]bundles.BundleClient{42: mockBundleClient})
	setMockBundleClientReferences(t, mockBundleClient, "main.go", 23, 34, []bundles.Location{
		{DumpID: 42, Path: "foo.go", Range: testRange1},
		{DumpID: 42, Path: "foo.go", Range: testRange2},
		{DumpID: 42, Path: "foo.go", Range: testRange3},
	})

	rpr := &ReferencePageResolver{
		db:                  mockDB,
		bundleManagerClient: mockBundleManagerClient,
		repositoryID:        100,
		commit:              testCommit,
		limit:               5,
	}

	t.Run("partial results", func(t *testing.T) {
		setMockBundleClientMonikerResults(t, mockBundleClient, "reference", "gomod", "pad", 0, 5, []bundles.Location{
			{DumpID: 42, Path: "foo.go", Range: testRange1},
			{DumpID: 42, Path: "foo.go", Range: testRange2},
			{DumpID: 42, Path: "bar.go", Range: testRange2},
			{DumpID: 42, Path: "bar.go", Range: testRange3},
			{DumpID: 42, Path: "bar.go", Range: testRange4},
		}, 7)

		references, newCursor, hasNewCursor, err := rpr.dispatchCursorHandler(context.Background(), Cursor{
			Phase:       "same-dump-monikers",
			DumpID:      42,
			Path:        "main.go",
			Line:        23,
			Character:   34,
			Monikers:    []bundles.MonikerData{{Kind: "export", Scheme: "gomod", Identifier: "pad"}},
			SkipResults: 0,
		})
		if err != nil {
			t.Fatalf("expected error getting references: %s", err)
		}

		expectedReferences := []ResolvedLocation{
			{Dump: testDump1, Path: "sub1/bar.go", Range: testRange2},
			{Dump: testDump1, Path: "sub1/bar.go", Range: testRange3},
			{Dump: testDump1, Path: "sub1/bar.go", Range: testRange4},
		}
		if diff := cmp.Diff(expectedReferences, references); diff != "" {
			t.Errorf("unexpected references (-want +got):\n%s", diff)
		}

		expectedNewCursor := Cursor{
			Phase:       "same-dump-monikers",
			DumpID:      42,
			Path:        "main.go",
			Line:        23,
			Character:   34,
			Monikers:    []bundles.MonikerData{{Kind: "export", Scheme: "gomod", Identifier: "pad"}},
			SkipResults: 5,
		}
		if !hasNewCursor {
			t.Errorf("expected new cursor")
		} else if diff := cmp.Diff(expectedNewCursor, newCursor); diff != "" {
			t.Errorf("unexpected new cursor (-want +got):\n%s", diff)
		}
	})

	t.Run("end of result set", func(t *testing.T) {
		setMockBundleClientMonikerResults(t, mockBundleClient, "reference", "gomod", "pad", 5, 5, []bundles.Location{
			{DumpID: 42, Path: "baz.go", Range: testRange1},
			{DumpID: 42, Path: "baz.go", Range: testRange2},
		}, 7)

		references, newCursor, hasNewCursor, err := rpr.dispatchCursorHandler(context.Background(), Cursor{
			Phase:       "same-dump-monikers",
			DumpID:      42,
			Path:        "main.go",
			Line:        23,
			Character:   34,
			Monikers:    []bundles.MonikerData{{Kind: "export", Scheme: "gomod", Identifier: "pad"}},
			SkipResults: 5,
		})
		if err != nil {
			t.Fatalf("expected error getting references: %s", err)
		}

		expectedReferences := []ResolvedLocation{
			{Dump: testDump1, Path: "sub1/baz.go", Range: testRange1},
			{Dump: testDump1, Path: "sub1/baz.go", Range: testRange2},
		}
		if diff := cmp.Diff(expectedReferences, references); diff != "" {
			t.Errorf("unexpected references (-want +got):\n%s", diff)
		}

		expectedNewCursor := Cursor{
			Phase:       "definition-monikers",
			DumpID:      42,
			Path:        "main.go",
			Monikers:    []bundles.MonikerData{{Kind: "export", Scheme: "gomod", Identifier: "pad"}},
			SkipResults: 0,
		}
		if !hasNewCursor {
			t.Errorf("expected new cursor")
		} else if diff := cmp.Diff(expectedNewCursor, newCursor); diff != "" {
			t.Errorf("unexpected new cursor (-want +got):\n%s", diff)
		}
	})
}

func TestHandleDefinitionMonikersCursor(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	mockBundleClient1 := bundlemocks.NewMockBundleClient()
	mockBundleClient2 := bundlemocks.NewMockBundleClient()

	setMockDBGetDumpByID(t, mockDB, map[int]db.Dump{42: testDump1, 50: testDump2})
	setMockBundleManagerClientBundleClient(t, mockBundleManagerClient, map[int]bundles.BundleClient{42: mockBundleClient1, 50: mockBundleClient2})
	setMockBundleClientPackageInformation(t, mockBundleClient1, "main.go", "1234", testPackageInformation)
	setMockDBGetPackage(t, mockDB, "gomod", "leftpad", "0.1.0", testDump2, true)

	rpr := &ReferencePageResolver{
		db:                  mockDB,
		bundleManagerClient: mockBundleManagerClient,
		repositoryID:        100,
		commit:              testCommit,
		limit:               5,
	}

	t.Run("partial results", func(t *testing.T) {
		setMockBundleClientMonikerResults(t, mockBundleClient2, "reference", "gomod", "pad", 0, 5, []bundles.Location{
			{DumpID: 50, Path: "foo.go", Range: testRange1},
			{DumpID: 50, Path: "bar.go", Range: testRange2},
			{DumpID: 50, Path: "baz.go", Range: testRange3},
			{DumpID: 50, Path: "bonk.go", Range: testRange4},
			{DumpID: 50, Path: "quux.go", Range: testRange5},
		}, 10)

		references, newCursor, hasNewCursor, err := rpr.dispatchCursorHandler(context.Background(), Cursor{
			Phase:       "definition-monikers",
			DumpID:      42,
			Path:        "main.go",
			Monikers:    []bundles.MonikerData{{Kind: "import", Scheme: "gomod", Identifier: "pad", PackageInformationID: "1234"}},
			SkipResults: 0,
		})
		if err != nil {
			t.Fatalf("expected error getting references: %s", err)
		}

		expectedReferences := []ResolvedLocation{
			{Dump: testDump2, Path: "sub2/foo.go", Range: testRange1},
			{Dump: testDump2, Path: "sub2/bar.go", Range: testRange2},
			{Dump: testDump2, Path: "sub2/baz.go", Range: testRange3},
			{Dump: testDump2, Path: "sub2/bonk.go", Range: testRange4},
			{Dump: testDump2, Path: "sub2/quux.go", Range: testRange5},
		}
		if diff := cmp.Diff(references, expectedReferences); diff != "" {
			t.Errorf("unexpected references (-want +got):\n%s", diff)
		}

		expectedNewCursor := Cursor{
			Phase:       "definition-monikers",
			DumpID:      42,
			Path:        "main.go",
			Monikers:    []bundles.MonikerData{{Kind: "import", Scheme: "gomod", Identifier: "pad", PackageInformationID: "1234"}},
			SkipResults: 5,
		}
		if !hasNewCursor {
			t.Errorf("expected new cursor")
		} else if diff := cmp.Diff(expectedNewCursor, newCursor); diff != "" {
			t.Errorf("unexpected new cursor (-want +got):\n%s", diff)
		}
	})

	t.Run("end of result set", func(t *testing.T) {
		setMockBundleClientMonikerResults(t, mockBundleClient2, "reference", "gomod", "pad", 5, 5, []bundles.Location{
			{DumpID: 50, Path: "foo.go", Range: testRange1},
			{DumpID: 50, Path: "bar.go", Range: testRange2},
			{DumpID: 50, Path: "baz.go", Range: testRange3},
			{DumpID: 50, Path: "bonk.go", Range: testRange4},
			{DumpID: 50, Path: "quux.go", Range: testRange5},
		}, 10)

		references, newCursor, hasNewCursor, err := rpr.dispatchCursorHandler(context.Background(), Cursor{
			Phase:       "definition-monikers",
			DumpID:      42,
			Path:        "main.go",
			Monikers:    []bundles.MonikerData{{Kind: "import", Scheme: "gomod", Identifier: "pad", PackageInformationID: "1234"}},
			SkipResults: 5,
		})
		if err != nil {
			t.Fatalf("expected error getting references: %s", err)
		}

		expectedReferences := []ResolvedLocation{
			{Dump: testDump2, Path: "sub2/foo.go", Range: testRange1},
			{Dump: testDump2, Path: "sub2/bar.go", Range: testRange2},
			{Dump: testDump2, Path: "sub2/baz.go", Range: testRange3},
			{Dump: testDump2, Path: "sub2/bonk.go", Range: testRange4},
			{Dump: testDump2, Path: "sub2/quux.go", Range: testRange5},
		}
		if diff := cmp.Diff(expectedReferences, references); diff != "" {
			t.Errorf("unexpected references (-want +got):\n%s", diff)
		}

		expectedNewCursor := Cursor{
			Phase:      "same-repo",
			DumpID:     42,
			Scheme:     "gomod",
			Identifier: "pad",
			Name:       "leftpad",
			Version:    "0.1.0",
		}
		if !hasNewCursor {
			t.Errorf("expected new cursor")
		} else if diff := cmp.Diff(expectedNewCursor, newCursor); diff != "" {
			t.Errorf("unexpected new cursor (-want +got):\n%s", diff)
		}
	})
}

func TestHandleSameRepoCursor(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	mockBundleClient1 := bundlemocks.NewMockBundleClient()
	mockBundleClient2 := bundlemocks.NewMockBundleClient()
	mockBundleClient3 := bundlemocks.NewMockBundleClient()
	mockReferencePager := mocks.NewMockReferencePager()

	setMockDBGetDumpByID(t, mockDB, map[int]db.Dump{42: testDump1, 50: testDump2, 51: testDump3, 52: testDump4})
	setMockBundleManagerClientBundleClient(t, mockBundleManagerClient, map[int]bundles.BundleClient{50: mockBundleClient1, 51: mockBundleClient2, 52: mockBundleClient3})
	setMockDBSameRepoPager(t, mockDB, 100, testCommit, "gomod", "leftpad", "0.1.0", 5, 3, mockReferencePager)
	setMockReferencePagerPageFromOffset(t, mockReferencePager, 0, []types.PackageReference{
		{DumpID: 50, Filter: readTestFilter(t, "normal", "1")},
		{DumpID: 51, Filter: readTestFilter(t, "normal", "1")},
		{DumpID: 52, Filter: readTestFilter(t, "normal", "1")},
	})

	t.Run("partial results", func(t *testing.T) {
		setMockBundleClientMonikerResults(t, mockBundleClient1, "reference", "gomod", "bar", 0, 5, []bundles.Location{
			{DumpID: 50, Path: "foo.go", Range: testRange1},
			{DumpID: 50, Path: "bar.go", Range: testRange2},
			{DumpID: 51, Path: "baz.go", Range: testRange3},
			{DumpID: 51, Path: "bonk.go", Range: testRange4},
			{DumpID: 52, Path: "quux.go", Range: testRange5},
		}, 10)

		rpr := &ReferencePageResolver{
			db:                  mockDB,
			bundleManagerClient: mockBundleManagerClient,
			repositoryID:        100,
			commit:              testCommit,
			remoteDumpLimit:     5,
			limit:               5,
		}

		references, newCursor, hasNewCursor, err := rpr.resolvePage(context.Background(), Cursor{
			Phase:      "same-repo",
			DumpID:     42,
			Scheme:     "gomod",
			Identifier: "bar",
			Name:       "leftpad",
			Version:    "0.1.0",
		})
		if err != nil {
			t.Fatalf("expected error getting references: %s", err)
		}

		expectedReferences := []ResolvedLocation{
			{Dump: testDump2, Path: "sub2/foo.go", Range: testRange1},
			{Dump: testDump2, Path: "sub2/bar.go", Range: testRange2},
			{Dump: testDump2, Path: "sub2/baz.go", Range: testRange3},
			{Dump: testDump2, Path: "sub2/bonk.go", Range: testRange4},
			{Dump: testDump2, Path: "sub2/quux.go", Range: testRange5},
		}
		if diff := cmp.Diff(expectedReferences, references); diff != "" {
			t.Errorf("unexpected references (-want +got):\n%s", diff)
		}

		expectedNewCursor := Cursor{
			Phase:                  "same-repo",
			DumpID:                 42,
			Scheme:                 "gomod",
			Identifier:             "bar",
			Name:                   "leftpad",
			Version:                "0.1.0",
			DumpIDs:                []int{50, 51, 52},
			TotalDumpsWhenBatching: 3,
			SkipDumpsWhenBatching:  3,
			SkipDumpsInBatch:       0,
			SkipResultsInDump:      5,
		}
		if !hasNewCursor {
			t.Errorf("expected new cursor")
		} else if diff := cmp.Diff(expectedNewCursor, newCursor); diff != "" {
			t.Errorf("unexpected new cursor (-want +got):\n%s", diff)
		}
	})

	t.Run("multiple pages", func(t *testing.T) {
		setMockBundleClientMonikerResults(t, mockBundleClient1, "reference", "gomod", "bar", 0, 5, []bundles.Location{
			{DumpID: 50, Path: "foo.go", Range: testRange1},
			{DumpID: 50, Path: "bar.go", Range: testRange2},
		}, 2)
		setMockBundleClientMonikerResults(t, mockBundleClient2, "reference", "gomod", "bar", 0, 3, []bundles.Location{
			{DumpID: 51, Path: "baz.go", Range: testRange3},
			{DumpID: 51, Path: "bonk.go", Range: testRange4},
		}, 2)
		setMockBundleClientMonikerResults(t, mockBundleClient3, "reference", "gomod", "bar", 0, 1, []bundles.Location{
			{DumpID: 52, Path: "quux.go", Range: testRange5},
		}, 1)

		rpr := &ReferencePageResolver{
			db:                  mockDB,
			bundleManagerClient: mockBundleManagerClient,
			repositoryID:        100,
			commit:              testCommit,
			remoteDumpLimit:     5,
			limit:               5,
		}

		references, newCursor, hasNewCursor, err := rpr.resolvePage(context.Background(), Cursor{
			Phase:      "same-repo",
			DumpID:     42,
			Scheme:     "gomod",
			Identifier: "bar",
			Name:       "leftpad",
			Version:    "0.1.0",
		})
		if err != nil {
			t.Fatalf("expected error getting references: %s", err)
		}

		expectedReferences := []ResolvedLocation{
			{Dump: testDump2, Path: "sub2/foo.go", Range: testRange1},
			{Dump: testDump2, Path: "sub2/bar.go", Range: testRange2},
			{Dump: testDump3, Path: "sub3/baz.go", Range: testRange3},
			{Dump: testDump3, Path: "sub3/bonk.go", Range: testRange4},
			{Dump: testDump4, Path: "sub4/quux.go", Range: testRange5},
		}
		if diff := cmp.Diff(expectedReferences, references); diff != "" {
			t.Errorf("unexpected references (-want +got):\n%s", diff)
		}

		expectedNewCursor := Cursor{
			Phase:      "remote-repo",
			DumpID:     42,
			Scheme:     "gomod",
			Identifier: "bar",
			Name:       "leftpad",
			Version:    "0.1.0",
		}
		if !hasNewCursor {
			t.Errorf("expected new cursor")
		} else if diff := cmp.Diff(expectedNewCursor, newCursor); diff != "" {
			t.Errorf("unexpected new cursor (-want +got):\n%s", diff)
		}
	})
}

func TestHandleSameRepoCursorMultipleDumpBatches(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	mockBundleClient := bundlemocks.NewMockBundleClient()
	mockReferencePager := mocks.NewMockReferencePager()

	setMockDBGetDumpByID(t, mockDB, map[int]db.Dump{42: testDump1, 50: testDump2, 51: testDump3, 52: testDump4})
	setMockBundleManagerClientBundleClient(t, mockBundleManagerClient, map[int]bundles.BundleClient{51: mockBundleClient})
	setMockDBSameRepoPager(t, mockDB, 100, testCommit, "gomod", "leftpad", "0.1.0", 2, 3, mockReferencePager)
	setMockReferencePagerPageFromOffset(t, mockReferencePager, 0, []types.PackageReference{
		{DumpID: 50, Filter: readTestFilter(t, "normal", "1")},
		{DumpID: 51, Filter: readTestFilter(t, "normal", "1")},
	})
	setMockBundleClientMonikerResults(t, mockBundleClient, "reference", "gomod", "bar", 0, 5, []bundles.Location{
		{DumpID: 51, Path: "baz.go", Range: testRange3},
		{DumpID: 51, Path: "bonk.go", Range: testRange4},
	}, 2)

	rpr := &ReferencePageResolver{
		db:                  mockDB,
		bundleManagerClient: mockBundleManagerClient,
		repositoryID:        100,
		commit:              testCommit,
		remoteDumpLimit:     2,
		limit:               5,
	}

	references, newCursor, hasNewCursor, err := rpr.dispatchCursorHandler(context.Background(), Cursor{
		Phase:                  "same-repo",
		DumpID:                 42,
		Scheme:                 "gomod",
		Identifier:             "bar",
		Name:                   "leftpad",
		Version:                "0.1.0",
		DumpIDs:                []int{50, 51},
		TotalDumpsWhenBatching: 3,
		SkipDumpsWhenBatching:  2,
		SkipDumpsInBatch:       1,
		SkipResultsInDump:      0,
	})
	if err != nil {
		t.Fatalf("expected error getting references: %s", err)
	}

	expectedReferences := []ResolvedLocation{
		{Dump: testDump3, Path: "sub3/baz.go", Range: testRange3},
		{Dump: testDump3, Path: "sub3/bonk.go", Range: testRange4},
	}
	if diff := cmp.Diff(expectedReferences, references); diff != "" {
		t.Errorf("unexpected references (-want +got):\n%s", diff)
	}

	expectedNewCursor := Cursor{
		Phase:                  "same-repo",
		DumpID:                 42,
		Scheme:                 "gomod",
		Identifier:             "bar",
		Name:                   "leftpad",
		Version:                "0.1.0",
		DumpIDs:                []int{},
		TotalDumpsWhenBatching: 3,
		SkipDumpsWhenBatching:  2,
		SkipDumpsInBatch:       0,
		SkipResultsInDump:      0,
	}
	if !hasNewCursor {
		t.Errorf("expected new cursor")
	} else if diff := cmp.Diff(expectedNewCursor, newCursor); diff != "" {
		t.Errorf("unexpected new cursor (-want +got):\n%s", diff)
	}
}

//
//
//
//

func TestHandleRemoteRepoCursor(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	mockBundleClient1 := bundlemocks.NewMockBundleClient()
	mockBundleClient2 := bundlemocks.NewMockBundleClient()
	mockBundleClient3 := bundlemocks.NewMockBundleClient()
	mockReferencePager := mocks.NewMockReferencePager()

	setMockDBGetDumpByID(t, mockDB, map[int]db.Dump{42: testDump1, 50: testDump2, 51: testDump3, 52: testDump4})
	setMockBundleManagerClientBundleClient(t, mockBundleManagerClient, map[int]bundles.BundleClient{50: mockBundleClient1, 51: mockBundleClient2, 52: mockBundleClient3})
	setMockDBPackageReferencePager(t, mockDB, "gomod", "leftpad", "0.1.0", 100, 5, 3, mockReferencePager)
	setMockReferencePagerPageFromOffset(t, mockReferencePager, 0, []types.PackageReference{
		{DumpID: 50, Filter: readTestFilter(t, "normal", "1")},
		{DumpID: 51, Filter: readTestFilter(t, "normal", "1")},
		{DumpID: 52, Filter: readTestFilter(t, "normal", "1")},
	})

	t.Run("partial results", func(t *testing.T) {
		setMockBundleClientMonikerResults(t, mockBundleClient1, "reference", "gomod", "bar", 0, 5, []bundles.Location{
			{DumpID: 50, Path: "foo.go", Range: testRange1},
			{DumpID: 50, Path: "bar.go", Range: testRange2},
			{DumpID: 51, Path: "baz.go", Range: testRange3},
			{DumpID: 51, Path: "bonk.go", Range: testRange4},
			{DumpID: 52, Path: "quux.go", Range: testRange5},
		}, 10)

		rpr := &ReferencePageResolver{
			db:                  mockDB,
			bundleManagerClient: mockBundleManagerClient,
			repositoryID:        100,
			commit:              testCommit,
			remoteDumpLimit:     5,
			limit:               5,
		}

		references, newCursor, hasNewCursor, err := rpr.resolvePage(context.Background(), Cursor{
			Phase:      "remote-repo",
			DumpID:     42,
			Scheme:     "gomod",
			Identifier: "bar",
			Name:       "leftpad",
			Version:    "0.1.0",
		})
		if err != nil {
			t.Fatalf("expected error getting references: %s", err)
		}

		expectedReferences := []ResolvedLocation{
			{Dump: testDump2, Path: "sub2/foo.go", Range: testRange1},
			{Dump: testDump2, Path: "sub2/bar.go", Range: testRange2},
			{Dump: testDump2, Path: "sub2/baz.go", Range: testRange3},
			{Dump: testDump2, Path: "sub2/bonk.go", Range: testRange4},
			{Dump: testDump2, Path: "sub2/quux.go", Range: testRange5},
		}
		if diff := cmp.Diff(expectedReferences, references); diff != "" {
			t.Errorf("unexpected references (-want +got):\n%s", diff)
		}

		expectedNewCursor := Cursor{
			Phase:                  "remote-repo",
			DumpID:                 42,
			Scheme:                 "gomod",
			Identifier:             "bar",
			Name:                   "leftpad",
			Version:                "0.1.0",
			DumpIDs:                []int{50, 51, 52},
			TotalDumpsWhenBatching: 3,
			SkipDumpsWhenBatching:  3,
			SkipDumpsInBatch:       0,
			SkipResultsInDump:      5,
		}
		if !hasNewCursor {
			t.Errorf("expected new cursor")
		} else if diff := cmp.Diff(expectedNewCursor, newCursor); diff != "" {
			t.Errorf("unexpected new cursor (-want +got):\n%s", diff)
		}
	})

	t.Run("multiple pages", func(t *testing.T) {
		setMockBundleClientMonikerResults(t, mockBundleClient1, "reference", "gomod", "bar", 0, 5, []bundles.Location{
			{DumpID: 50, Path: "foo.go", Range: testRange1},
			{DumpID: 50, Path: "bar.go", Range: testRange2},
		}, 2)
		setMockBundleClientMonikerResults(t, mockBundleClient2, "reference", "gomod", "bar", 0, 3, []bundles.Location{
			{DumpID: 51, Path: "baz.go", Range: testRange3},
			{DumpID: 51, Path: "bonk.go", Range: testRange4},
		}, 2)
		setMockBundleClientMonikerResults(t, mockBundleClient3, "reference", "gomod", "bar", 0, 1, []bundles.Location{
			{DumpID: 52, Path: "quux.go", Range: testRange5},
		}, 1)

		rpr := &ReferencePageResolver{
			db:                  mockDB,
			bundleManagerClient: mockBundleManagerClient,
			repositoryID:        100,
			commit:              testCommit,
			remoteDumpLimit:     5,
			limit:               5,
		}

		references, _, hasNewCursor, err := rpr.resolvePage(context.Background(), Cursor{
			Phase:      "remote-repo",
			DumpID:     42,
			Scheme:     "gomod",
			Identifier: "bar",
			Name:       "leftpad",
			Version:    "0.1.0",
		})
		if err != nil {
			t.Fatalf("expected error getting references: %s", err)
		}

		expectedReferences := []ResolvedLocation{
			{Dump: testDump2, Path: "sub2/foo.go", Range: testRange1},
			{Dump: testDump2, Path: "sub2/bar.go", Range: testRange2},
			{Dump: testDump3, Path: "sub3/baz.go", Range: testRange3},
			{Dump: testDump3, Path: "sub3/bonk.go", Range: testRange4},
			{Dump: testDump4, Path: "sub4/quux.go", Range: testRange5},
		}
		if diff := cmp.Diff(expectedReferences, references); diff != "" {
			t.Errorf("unexpected references (-want +got):\n%s", diff)
		}
		if hasNewCursor {
			t.Errorf("unexpected new cursor")
		}
	})
}

func TestHandleRemoteRepoCursorMultipleDumpBatches(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	mockBundleClient := bundlemocks.NewMockBundleClient()
	mockReferencePager := mocks.NewMockReferencePager()

	setMockDBGetDumpByID(t, mockDB, map[int]db.Dump{42: testDump1, 50: testDump2, 51: testDump3, 52: testDump4})
	setMockBundleManagerClientBundleClient(t, mockBundleManagerClient, map[int]bundles.BundleClient{51: mockBundleClient})
	setMockDBPackageReferencePager(t, mockDB, "gomod", "leftpad", "0.1.0", 100, 2, 3, mockReferencePager)
	setMockReferencePagerPageFromOffset(t, mockReferencePager, 0, []types.PackageReference{
		{DumpID: 50, Filter: readTestFilter(t, "normal", "1")},
		{DumpID: 51, Filter: readTestFilter(t, "normal", "1")},
	})
	setMockBundleClientMonikerResults(t, mockBundleClient, "reference", "gomod", "bar", 0, 5, []bundles.Location{
		{DumpID: 51, Path: "baz.go", Range: testRange3},
		{DumpID: 51, Path: "bonk.go", Range: testRange4},
	}, 2)

	rpr := &ReferencePageResolver{
		db:                  mockDB,
		bundleManagerClient: mockBundleManagerClient,
		repositoryID:        100,
		commit:              testCommit,
		remoteDumpLimit:     2,
		limit:               5,
	}

	references, newCursor, hasNewCursor, err := rpr.dispatchCursorHandler(context.Background(), Cursor{
		Phase:                  "remote-repo",
		DumpID:                 42,
		Scheme:                 "gomod",
		Identifier:             "bar",
		Name:                   "leftpad",
		Version:                "0.1.0",
		DumpIDs:                []int{50, 51},
		TotalDumpsWhenBatching: 3,
		SkipDumpsWhenBatching:  2,
		SkipDumpsInBatch:       1,
		SkipResultsInDump:      0,
	})
	if err != nil {
		t.Fatalf("expected error getting references: %s", err)
	}

	expectedReferences := []ResolvedLocation{
		{Dump: testDump3, Path: "sub3/baz.go", Range: testRange3},
		{Dump: testDump3, Path: "sub3/bonk.go", Range: testRange4},
	}
	if diff := cmp.Diff(expectedReferences, references); diff != "" {
		t.Errorf("unexpected references (-want +got):\n%s", diff)
	}

	expectedNewCursor := Cursor{
		Phase:                  "remote-repo",
		DumpID:                 42,
		Scheme:                 "gomod",
		Identifier:             "bar",
		Name:                   "leftpad",
		Version:                "0.1.0",
		DumpIDs:                []int{},
		TotalDumpsWhenBatching: 3,
		SkipDumpsWhenBatching:  2,
		SkipDumpsInBatch:       0,
		SkipResultsInDump:      0,
	}
	if !hasNewCursor {
		t.Errorf("expected new cursor")
	} else if diff := cmp.Diff(expectedNewCursor, newCursor); diff != "" {
		t.Errorf("unexpected new cursor (-want +got):\n%s", diff)
	}
}

func TestApplyBloomFilter(t *testing.T) {
	references := []types.PackageReference{
		{DumpID: 1, Filter: readTestFilter(t, "normal", "1")},   // bar
		{DumpID: 2, Filter: readTestFilter(t, "normal", "2")},   // no bar
		{DumpID: 3, Filter: readTestFilter(t, "normal", "3")},   // bar
		{DumpID: 4, Filter: readTestFilter(t, "normal", "4")},   // bar
		{DumpID: 5, Filter: readTestFilter(t, "normal", "5")},   // no bar
		{DumpID: 6, Filter: readTestFilter(t, "normal", "6")},   // bar
		{DumpID: 7, Filter: readTestFilter(t, "normal", "7")},   // bar
		{DumpID: 8, Filter: readTestFilter(t, "normal", "8")},   // no bar
		{DumpID: 9, Filter: readTestFilter(t, "normal", "9")},   // bar
		{DumpID: 10, Filter: readTestFilter(t, "normal", "10")}, // bar
		{DumpID: 11, Filter: readTestFilter(t, "normal", "11")}, // no bar
		{DumpID: 12, Filter: readTestFilter(t, "normal", "12")}, // bar
	}

	testCases := []struct {
		limit           int
		expectedScanned int
		expectedDumpIDs []int
	}{
		{1, 1, []int{1}},
		{2, 3, []int{1, 3}},
		{6, 9, []int{1, 3, 4, 6, 7, 9}},
		{7, 10, []int{1, 3, 4, 6, 7, 9, 10}},
		{8, 12, []int{1, 3, 4, 6, 7, 9, 10, 12}},
		{12, 12, []int{1, 3, 4, 6, 7, 9, 10, 12}},
	}

	for _, testCase := range testCases {
		name := fmt.Sprintf("limit=%d", testCase.limit)

		t.Run(name, func(t *testing.T) {
			filteredReferences, scanned := applyBloomFilter(references, "bar", testCase.limit)
			if scanned != testCase.expectedScanned {
				t.Errorf("unexpected scanned. want=%d have=%d", testCase.expectedScanned, scanned)
			}

			var filteredDumpIDs []int
			for _, reference := range filteredReferences {
				filteredDumpIDs = append(filteredDumpIDs, reference.DumpID)
			}

			if diff := cmp.Diff(testCase.expectedDumpIDs, filteredDumpIDs); diff != "" {
				t.Errorf("unexpected filtered references ids (-want +got):\n%s", diff)
			}
		})
	}
}
