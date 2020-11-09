package api

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
)

func TestSerializationRoundTrip(t *testing.T) {
	c := Cursor{
		Phase:     "same-repo",
		DumpID:    42,
		Path:      "/foo/bar/baz.go",
		Line:      10,
		Character: 50,
		Monikers: []lsifstore.MonikerData{
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
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()

	setMockDBStoreGetDumpByID(t, mockDBStore, map[int]store.Dump{42: testDump1})
	setmockLSIFStoreMonikersByPosition(t, mockLSIFStore, 42, "main.go", 10, 20, [][]lsifstore.MonikerData{{testMoniker1}, {testMoniker2}})

	expectedCursor := Cursor{
		Phase:     "same-dump",
		DumpID:    42,
		Path:      "main.go",
		Line:      10,
		Character: 20,
		Monikers:  []lsifstore.MonikerData{testMoniker1, testMoniker2},
	}

	if cursor, err := DecodeOrCreateCursor("sub1/main.go", 10, 20, 42, "", mockDBStore, mockLSIFStore); err != nil {
		t.Fatalf("unexpected error decoding cursor: %s", err)
	} else if diff := cmp.Diff(expectedCursor, cursor); diff != "" {
		t.Errorf("unexpected cursor (-want +got):\n%s", diff)
	}
}

func TestDecodeOrCreateCursorUnknownDump(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	setMockDBStoreGetDumpByID(t, mockDBStore, nil)

	if _, err := DecodeOrCreateCursor("sub1/main.go", 10, 20, 42, "", mockDBStore, mockLSIFStore); err != ErrMissingDump {
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
		Monikers: []lsifstore.MonikerData{
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

	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()

	if cursor, err := DecodeOrCreateCursor("", 0, 0, 0, EncodeCursor(expectedCursor), mockDBStore, mockLSIFStore); err != nil {
		t.Fatalf("unexpected error decoding cursor: %s", err)
	} else if diff := cmp.Diff(expectedCursor, cursor); diff != "" {
		t.Errorf("unexpected cursor (-want +got):\n%s", diff)
	}
}
