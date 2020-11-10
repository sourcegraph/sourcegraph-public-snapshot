package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
)

func TestHandleSameDumpCursor(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()

	setMockDBStoreGetDumpByID(t, mockDBStore, map[int]store.Dump{42: testDump1})
	setmockLSIFStoreReferences(t, mockLSIFStore, 42, "main.go", 23, 34, []lsifstore.Location{
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
		dbStore:      mockDBStore,
		lsifStore:    mockLSIFStore,
		repositoryID: 100,
		commit:       testCommit,
		limit:        5,
	}

	t.Run("partial results", func(t *testing.T) {
		references, newCursor, hasNewCursor, err := rpr.dispatchCursorHandler(context.Background(), Cursor{
			Phase:       "same-dump",
			DumpID:      42,
			Path:        "main.go",
			Line:        23,
			Character:   34,
			Monikers:    []lsifstore.MonikerData{{Kind: "export", Scheme: "gomod", Identifier: "pad"}},
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
			Monikers:    []lsifstore.MonikerData{{Kind: "export", Scheme: "gomod", Identifier: "pad"}},
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
			Monikers:    []lsifstore.MonikerData{{Kind: "export", Scheme: "gomod", Identifier: "pad"}},
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
			Monikers:    []lsifstore.MonikerData{{Kind: "export", Scheme: "gomod", Identifier: "pad"}},
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
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()

	setMockDBStoreGetDumpByID(t, mockDBStore, map[int]store.Dump{42: testDump1})
	setmockLSIFStoreReferences(t, mockLSIFStore, 42, "main.go", 23, 34, []lsifstore.Location{
		{DumpID: 42, Path: "foo.go", Range: testRange1},
		{DumpID: 42, Path: "foo.go", Range: testRange2},
		{DumpID: 42, Path: "foo.go", Range: testRange3},
	})

	rpr := &ReferencePageResolver{
		dbStore:      mockDBStore,
		lsifStore:    mockLSIFStore,
		repositoryID: 100,
		commit:       testCommit,
		limit:        5,
	}

	t.Run("partial results", func(t *testing.T) {
		setmockLSIFStoreMonikerResults(t, mockLSIFStore, 42, "references", "gomod", "pad", 0, 5, []lsifstore.Location{
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
			Monikers:    []lsifstore.MonikerData{{Kind: "export", Scheme: "gomod", Identifier: "pad"}},
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
			Monikers:    []lsifstore.MonikerData{{Kind: "export", Scheme: "gomod", Identifier: "pad"}},
			SkipResults: 5,
		}
		if !hasNewCursor {
			t.Errorf("expected new cursor")
		} else if diff := cmp.Diff(expectedNewCursor, newCursor); diff != "" {
			t.Errorf("unexpected new cursor (-want +got):\n%s", diff)
		}
	})

	t.Run("end of result set", func(t *testing.T) {
		setmockLSIFStoreMonikerResults(t, mockLSIFStore, 42, "references", "gomod", "pad", 5, 5, []lsifstore.Location{
			{DumpID: 42, Path: "baz.go", Range: testRange1},
			{DumpID: 42, Path: "baz.go", Range: testRange2},
		}, 7)

		references, newCursor, hasNewCursor, err := rpr.dispatchCursorHandler(context.Background(), Cursor{
			Phase:       "same-dump-monikers",
			DumpID:      42,
			Path:        "main.go",
			Line:        23,
			Character:   34,
			Monikers:    []lsifstore.MonikerData{{Kind: "export", Scheme: "gomod", Identifier: "pad"}},
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
			Monikers:    []lsifstore.MonikerData{{Kind: "export", Scheme: "gomod", Identifier: "pad"}},
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
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()

	setMockDBStoreGetDumpByID(t, mockDBStore, map[int]store.Dump{42: testDump1, 50: testDump2})
	setmockLSIFStorePackageInformation(t, mockLSIFStore, 42, "main.go", "1234", testPackageInformation)
	setMockDBStoreGetPackage(t, mockDBStore, "gomod", "leftpad", "0.1.0", testDump2, true)

	rpr := &ReferencePageResolver{
		dbStore:      mockDBStore,
		lsifStore:    mockLSIFStore,
		repositoryID: 100,
		commit:       testCommit,
		limit:        5,
	}

	t.Run("partial results", func(t *testing.T) {
		setmockLSIFStoreMonikerResults(t, mockLSIFStore, 50, "references", "gomod", "pad", 0, 5, []lsifstore.Location{
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
			Monikers:    []lsifstore.MonikerData{{Kind: "import", Scheme: "gomod", Identifier: "pad", PackageInformationID: "1234"}},
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
			Monikers:    []lsifstore.MonikerData{{Kind: "import", Scheme: "gomod", Identifier: "pad", PackageInformationID: "1234"}},
			SkipResults: 5,
		}
		if !hasNewCursor {
			t.Errorf("expected new cursor")
		} else if diff := cmp.Diff(expectedNewCursor, newCursor); diff != "" {
			t.Errorf("unexpected new cursor (-want +got):\n%s", diff)
		}
	})

	t.Run("end of result set", func(t *testing.T) {
		setmockLSIFStoreMonikerResults(t, mockLSIFStore, 50, "references", "gomod", "pad", 5, 5, []lsifstore.Location{
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
			Monikers:    []lsifstore.MonikerData{{Kind: "import", Scheme: "gomod", Identifier: "pad", PackageInformationID: "1234"}},
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
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockReferencePager := NewMockReferencePager()

	setMockDBStoreGetDumpByID(t, mockDBStore, map[int]store.Dump{42: testDump1, 50: testDump2, 51: testDump3, 52: testDump4})
	setMockDBStoreSameRepoPager(t, mockDBStore, 100, testCommit, "gomod", "leftpad", "0.1.0", 5, 3, mockReferencePager)
	setMockReferencePagerPageFromOffset(t, mockReferencePager, 0, []lsifstore.PackageReference{
		{DumpID: 50, Filter: readTestFilter(t, "normal", "1")},
		{DumpID: 51, Filter: readTestFilter(t, "normal", "1")},
		{DumpID: 52, Filter: readTestFilter(t, "normal", "1")},
	})

	t.Run("partial results", func(t *testing.T) {
		setmockLSIFStoreMonikerResults(t, mockLSIFStore, 50, "references", "gomod", "bar", 0, 5, []lsifstore.Location{
			{DumpID: 50, Path: "foo.go", Range: testRange1},
			{DumpID: 50, Path: "bar.go", Range: testRange2},
			{DumpID: 51, Path: "baz.go", Range: testRange3},
			{DumpID: 51, Path: "bonk.go", Range: testRange4},
			{DumpID: 52, Path: "quux.go", Range: testRange5},
		}, 10)

		rpr := &ReferencePageResolver{
			dbStore:         mockDBStore,
			lsifStore:       mockLSIFStore,
			repositoryID:    100,
			commit:          testCommit,
			remoteDumpLimit: 5,
			limit:           5,
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
		setMultimockLSIFStoreMonikerResults(
			t,
			mockLSIFStore,
			monikerResultsSpec{
				50, "references", "gomod", "bar", 0, 5,
				[]lsifstore.Location{
					{DumpID: 50, Path: "foo.go", Range: testRange1},
					{DumpID: 50, Path: "bar.go", Range: testRange2},
				},
				2,
			},
			monikerResultsSpec{
				51, "references", "gomod", "bar", 0, 3,
				[]lsifstore.Location{
					{DumpID: 51, Path: "baz.go", Range: testRange3},
					{DumpID: 51, Path: "bonk.go", Range: testRange4},
				},
				2,
			},
			monikerResultsSpec{
				52, "references", "gomod", "bar", 0, 1,
				[]lsifstore.Location{
					{DumpID: 52, Path: "quux.go", Range: testRange5},
				},
				1,
			},
		)

		rpr := &ReferencePageResolver{
			dbStore:         mockDBStore,
			lsifStore:       mockLSIFStore,
			repositoryID:    100,
			commit:          testCommit,
			remoteDumpLimit: 5,
			limit:           5,
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
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockReferencePager := NewMockReferencePager()

	setMockDBStoreGetDumpByID(t, mockDBStore, map[int]store.Dump{42: testDump1, 50: testDump2, 51: testDump3, 52: testDump4})
	setMockDBStoreSameRepoPager(t, mockDBStore, 100, testCommit, "gomod", "leftpad", "0.1.0", 2, 3, mockReferencePager)
	setMockReferencePagerPageFromOffset(t, mockReferencePager, 0, []lsifstore.PackageReference{
		{DumpID: 50, Filter: readTestFilter(t, "normal", "1")},
		{DumpID: 51, Filter: readTestFilter(t, "normal", "1")},
	})
	setmockLSIFStoreMonikerResults(t, mockLSIFStore, 51, "references", "gomod", "bar", 0, 5, []lsifstore.Location{
		{DumpID: 51, Path: "baz.go", Range: testRange3},
		{DumpID: 51, Path: "bonk.go", Range: testRange4},
	}, 2)

	rpr := &ReferencePageResolver{
		dbStore:         mockDBStore,
		lsifStore:       mockLSIFStore,
		repositoryID:    100,
		commit:          testCommit,
		remoteDumpLimit: 2,
		limit:           5,
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
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockReferencePager := NewMockReferencePager()

	setMockDBStoreGetDumpByID(t, mockDBStore, map[int]store.Dump{42: testDump1, 50: testDump2, 51: testDump3, 52: testDump4})
	setMockDBStorePackageReferencePager(t, mockDBStore, "gomod", "leftpad", "0.1.0", 100, 5, 3, mockReferencePager)
	setMockReferencePagerPageFromOffset(t, mockReferencePager, 0, []lsifstore.PackageReference{
		{DumpID: 50, Filter: readTestFilter(t, "normal", "1")},
		{DumpID: 51, Filter: readTestFilter(t, "normal", "1")},
		{DumpID: 52, Filter: readTestFilter(t, "normal", "1")},
	})

	t.Run("partial results", func(t *testing.T) {
		setmockLSIFStoreMonikerResults(t, mockLSIFStore, 50, "references", "gomod", "bar", 0, 5, []lsifstore.Location{
			{DumpID: 50, Path: "foo.go", Range: testRange1},
			{DumpID: 50, Path: "bar.go", Range: testRange2},
			{DumpID: 51, Path: "baz.go", Range: testRange3},
			{DumpID: 51, Path: "bonk.go", Range: testRange4},
			{DumpID: 52, Path: "quux.go", Range: testRange5},
		}, 10)

		rpr := &ReferencePageResolver{
			dbStore:         mockDBStore,
			lsifStore:       mockLSIFStore,
			repositoryID:    100,
			commit:          testCommit,
			remoteDumpLimit: 5,
			limit:           5,
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
		setMultimockLSIFStoreMonikerResults(
			t,
			mockLSIFStore,
			monikerResultsSpec{
				50, "references", "gomod", "bar", 0, 5,
				[]lsifstore.Location{
					{DumpID: 50, Path: "foo.go", Range: testRange1},
					{DumpID: 50, Path: "bar.go", Range: testRange2},
				},
				2,
			},
			monikerResultsSpec{
				51, "references", "gomod", "bar", 0, 3,
				[]lsifstore.Location{
					{DumpID: 51, Path: "baz.go", Range: testRange3},
					{DumpID: 51, Path: "bonk.go", Range: testRange4},
				},
				2,
			},
			monikerResultsSpec{
				52, "references", "gomod", "bar", 0, 1,
				[]lsifstore.Location{
					{DumpID: 52, Path: "quux.go", Range: testRange5},
				},
				1,
			},
		)

		rpr := &ReferencePageResolver{
			dbStore:         mockDBStore,
			lsifStore:       mockLSIFStore,
			repositoryID:    100,
			commit:          testCommit,
			remoteDumpLimit: 5,
			limit:           5,
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
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockReferencePager := NewMockReferencePager()

	setMockDBStoreGetDumpByID(t, mockDBStore, map[int]store.Dump{42: testDump1, 50: testDump2, 51: testDump3, 52: testDump4})
	setMockDBStorePackageReferencePager(t, mockDBStore, "gomod", "leftpad", "0.1.0", 100, 2, 3, mockReferencePager)
	setMockReferencePagerPageFromOffset(t, mockReferencePager, 0, []lsifstore.PackageReference{
		{DumpID: 50, Filter: readTestFilter(t, "normal", "1")},
		{DumpID: 51, Filter: readTestFilter(t, "normal", "1")},
	})
	setmockLSIFStoreMonikerResults(t, mockLSIFStore, 51, "references", "gomod", "bar", 0, 5, []lsifstore.Location{
		{DumpID: 51, Path: "baz.go", Range: testRange3},
		{DumpID: 51, Path: "bonk.go", Range: testRange4},
	}, 2)

	rpr := &ReferencePageResolver{
		dbStore:         mockDBStore,
		lsifStore:       mockLSIFStore,
		repositoryID:    100,
		commit:          testCommit,
		remoteDumpLimit: 2,
		limit:           5,
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
	references := []lsifstore.PackageReference{
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
