package api

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	bundlemocks "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/mocks"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	dbmocks "github.com/sourcegraph/sourcegraph/internal/codeintel/db/mocks"
)

func TestSerializationRoundTrip(t *testing.T) {
	c := Cursor{
		Phase:     "same-repo",
		DumpID:    42,
		Path:      "/foo/bar/baz.go",
		Line:      10,
		Character: 50,
		Monikers: []bundles.MonikerData{
			{Kind: "k1", Scheme: "s1", Identifier: "i1", PackageInformationID: "pid1"},
			{Kind: "k2", Scheme: "s2", Identifier: "i2", PackageInformationID: "pid2"},
			{Kind: "k3", Scheme: "s3", Identifier: "i3", PackageInformationID: "pid3"},
		},
		SkipResults:            1,
		Identifier:             "x",
		Scheme:                 "gomod",
		Name:                   "leftpad",
		Version:                "0.1.0",
		DumpIDs:                []int{1, 2, 3, 4, 5},
		TotalDumpsWhenBatching: 5,
		SkipDumpsWhenBatching:  4,
		SkipDumpsInBatch:       3,
		SkipResultsInDump:      2,
	}

	roundtripped, err := decodeCursor(EncodeCursor(c))
	if err != nil {
		t.Fatalf("unexpected error decoding cursor: %s", err)
	}

	if diff := cmp.Diff(c, roundtripped); diff != "" {
		t.Errorf("unexpected cursor (-want +got):\n%s", diff)
	}
}

func TestDecodeOrCreateCursor(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	mockBundleClient := bundlemocks.NewMockBundleClient()

	setMockDBGetDumpByID(t, mockDB, map[int]db.Dump{42: testDump1})
	setMockBundleManagerClientBundleClient(t, mockBundleManagerClient, map[int]bundles.BundleClient{42: mockBundleClient})
	setMockBundleClientMonikersByPosition(t, mockBundleClient, "main.go", 10, 20, [][]bundles.MonikerData{{testMoniker1}, {testMoniker2}})

	expectedCursor := Cursor{
		Phase:     "same-dump",
		DumpID:    42,
		Path:      "main.go",
		Line:      10,
		Character: 20,
		Monikers:  []bundles.MonikerData{testMoniker1, testMoniker2},
	}

	if cursor, err := DecodeOrCreateCursor("sub1/main.go", 10, 20, 42, "", mockDB, mockBundleManagerClient); err != nil {
		t.Fatalf("unexpected error decoding cursor: %s", err)
	} else if diff := cmp.Diff(expectedCursor, cursor); diff != "" {
		t.Errorf("unexpected cursor (-want +got):\n%s", diff)
	}
}

func TestDecodeOrCreateCursorUnknownDump(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	setMockDBGetDumpByID(t, mockDB, nil)

	if _, err := DecodeOrCreateCursor("sub1/main.go", 10, 20, 42, "", mockDB, mockBundleManagerClient); err != ErrMissingDump {
		t.Fatalf("unexpected error decoding cursor. want=%q have =%q", ErrMissingDump, err)
	}
}

func TestDecodeOrCreateCursorExisting(t *testing.T) {
	expectedCursor := Cursor{
		Phase:     "same-repo",
		DumpID:    42,
		Path:      "/foo/bar/baz.go",
		Line:      10,
		Character: 50,
		Monikers: []bundles.MonikerData{
			{Kind: "k1", Scheme: "s1", Identifier: "i1", PackageInformationID: "pid1"},
			{Kind: "k2", Scheme: "s2", Identifier: "i2", PackageInformationID: "pid2"},
			{Kind: "k3", Scheme: "s3", Identifier: "i3", PackageInformationID: "pid3"},
		},
		SkipResults:            1,
		Identifier:             "x",
		Scheme:                 "gomod",
		Name:                   "leftpad",
		Version:                "0.1.0",
		DumpIDs:                []int{1, 2, 3, 4, 5},
		TotalDumpsWhenBatching: 5,
		SkipDumpsWhenBatching:  4,
		SkipDumpsInBatch:       3,
		SkipResultsInDump:      2,
	}

	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()

	if cursor, err := DecodeOrCreateCursor("", 0, 0, 0, EncodeCursor(expectedCursor), mockDB, mockBundleManagerClient); err != nil {
		t.Fatalf("unexpected error decoding cursor: %s", err)
	} else if diff := cmp.Diff(expectedCursor, cursor); diff != "" {
		t.Errorf("unexpected cursor (-want +got):\n%s", diff)
	}
}
