package lsifstore

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestReadWriteMeta(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)

	ctx := context.Background()
	store := NewStore(dbconn.Global, &observation.TestContext)

	if err := store.WriteMeta(ctx, 42, MetaData{NumResultChunks: 7}); err != nil {
		t.Fatalf("unexpected error while writing: %s", err)
	}

	meta, err := store.ReadMeta(ctx, 42)
	if err != nil {
		t.Fatalf("unexpected error reading from database: %s", err)
	}
	if meta.NumResultChunks != 7 {
		t.Errorf("unexpected num result chunks. want=%d have=%d", 7, meta.NumResultChunks)
	}
}

func TestReadWriteDocument(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)

	ctx := context.Background()
	store := NewStore(dbconn.Global, &observation.TestContext)

	expectedDocumentData := DocumentData{
		Ranges: map[ID]RangeData{
			"r01": {StartLine: 1, StartCharacter: 2, EndLine: 3, EndCharacter: 4, DefinitionResultID: "x01", MonikerIDs: []ID{"m01", "m02"}},
			"r02": {StartLine: 2, StartCharacter: 3, EndLine: 4, EndCharacter: 5, ReferenceResultID: "x06", MonikerIDs: []ID{"m03", "m04"}},
			"r03": {StartLine: 3, StartCharacter: 4, EndLine: 5, EndCharacter: 6, DefinitionResultID: "x02"},
		},
		HoverResults: map[ID]string{},
		Monikers: map[ID]MonikerData{
			"m01": {Kind: "import", Scheme: "scheme A", Identifier: "ident A", PackageInformationID: "p01"},
			"m02": {Kind: "import", Scheme: "scheme B", Identifier: "ident B"},
			"m03": {Kind: "export", Scheme: "scheme C", Identifier: "ident C", PackageInformationID: "p02"},
			"m04": {Kind: "export", Scheme: "scheme D", Identifier: "ident D"},
		},
		PackageInformation: map[ID]PackageInformationData{
			"p01": {Name: "pkg A", Version: "0.1.0"},
			"p02": {Name: "pkg B", Version: "1.2.3"},
		},
	}

	documentCh := make(chan KeyedDocumentData, 1)
	documentCh <- KeyedDocumentData{
		Path:     "foo.go",
		Document: expectedDocumentData,
	}
	close(documentCh)

	if err := store.WriteDocuments(ctx, 42, documentCh); err != nil {
		t.Fatalf("unexpected error while writing documents: %s", err)
	}

	documentData, _, err := store.ReadDocument(ctx, 42, "foo.go")
	if err != nil {
		t.Fatalf("unexpected error reading from database: %s", err)
	}
	if diff := cmp.Diff(expectedDocumentData, documentData); diff != "" {
		t.Errorf("unexpected document data (-want +got):\n%s", diff)
	}
}

func TestReadWriteResultChunk(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)

	ctx := context.Background()
	store := NewStore(dbconn.Global, &observation.TestContext)

	expectedResultChunkData := ResultChunkData{
		DocumentPaths: map[ID]string{
			"d01": "foo.go",
			"d02": "bar.go",
			"d03": "baz.go",
		},
		DocumentIDRangeIDs: map[ID][]DocumentIDRangeID{
			"x01": {
				{DocumentID: "d01", RangeID: "r03"},
				{DocumentID: "d02", RangeID: "r04"},
				{DocumentID: "d03", RangeID: "r07"},
			},
			"x02": {
				{DocumentID: "d01", RangeID: "r02"},
				{DocumentID: "d02", RangeID: "r05"},
				{DocumentID: "d03", RangeID: "r08"},
			},
			"x03": {
				{DocumentID: "d01", RangeID: "r01"},
				{DocumentID: "d02", RangeID: "r06"},
				{DocumentID: "d03", RangeID: "r09"},
			},
		},
	}

	resultChunkCh := make(chan IndexedResultChunkData, 1)
	resultChunkCh <- IndexedResultChunkData{
		Index:       7,
		ResultChunk: expectedResultChunkData,
	}
	close(resultChunkCh)

	if err := store.WriteResultChunks(ctx, 42, resultChunkCh); err != nil {
		t.Fatalf("unexpected error while writing result chunks: %s", err)
	}

	resultChunkData, _, err := store.ReadResultChunk(ctx, 42, 7)
	if err != nil {
		t.Fatalf("unexpected error reading from database: %s", err)
	}
	if diff := cmp.Diff(expectedResultChunkData, resultChunkData); diff != "" {
		t.Errorf("unexpected result chunk data (-want +got):\n%s", diff)
	}
}

func TestReadWriteDefinitions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)

	ctx := context.Background()
	store := NewStore(dbconn.Global, &observation.TestContext)

	expectedDefinitions := []LocationData{
		{URI: "bar.go", StartLine: 4, StartCharacter: 5, EndLine: 6, EndCharacter: 7},
		{URI: "baz.go", StartLine: 7, StartCharacter: 8, EndLine: 9, EndCharacter: 0},
		{URI: "foo.go", StartLine: 3, StartCharacter: 4, EndLine: 5, EndCharacter: 6},
	}

	definitionsCh := make(chan MonikerLocations, 1)
	definitionsCh <- MonikerLocations{
		Scheme:     "scheme A",
		Identifier: "ident A",
		Locations:  expectedDefinitions,
	}
	close(definitionsCh)

	if err := store.WriteDefinitions(ctx, 42, definitionsCh); err != nil {
		t.Fatalf("unexpected error while writing definitions: %s", err)
	}

	definitions, _, err := store.ReadDefinitions(ctx, 42, "scheme A", "ident A", 0, 100)
	if err != nil {
		t.Fatalf("unexpected error reading from database: %s", err)
	}
	if diff := cmp.Diff(expectedDefinitions, definitions); diff != "" {
		t.Errorf("unexpected definitions (-want +got):\n%s", diff)
	}
}

func TestReadWriteReferences(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)

	ctx := context.Background()
	store := NewStore(dbconn.Global, &observation.TestContext)

	expectedReferences := []LocationData{
		{URI: "baz.go", StartLine: 7, StartCharacter: 8, EndLine: 9, EndCharacter: 0},
		{URI: "baz.go", StartLine: 9, StartCharacter: 0, EndLine: 1, EndCharacter: 2},
		{URI: "foo.go", StartLine: 3, StartCharacter: 4, EndLine: 5, EndCharacter: 6},
	}

	referencesCh := make(chan MonikerLocations, 1)
	referencesCh <- MonikerLocations{
		Scheme:     "scheme C",
		Identifier: "ident C",
		Locations:  expectedReferences,
	}
	close(referencesCh)

	if err := store.WriteReferences(ctx, 42, referencesCh); err != nil {
		t.Fatalf("unexpected error while writing references: %s", err)
	}

	references, _, err := store.ReadReferences(ctx, 42, "scheme C", "ident C", 0, 100)
	if err != nil {
		t.Fatalf("unexpected error reading from database: %s", err)
	}
	if diff := cmp.Diff(expectedReferences, references); diff != "" {
		t.Errorf("unexpected references (-want +got):\n%s", diff)
	}
}

func TestPathsWithPrefix(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)

	ctx := context.Background()
	store := NewStore(dbconn.Global, &observation.TestContext)

	documentCh := make(chan KeyedDocumentData, 5)
	documentCh <- KeyedDocumentData{Path: "foo"}        // exact
	documentCh <- KeyedDocumentData{Path: "foo.go"}     // file prefix
	documentCh <- KeyedDocumentData{Path: "foo/bar.go"} // directory prefix
	documentCh <- KeyedDocumentData{Path: "bar/foo.go"} // contains, not prefixed
	documentCh <- KeyedDocumentData{Path: "bar/baz.go"} // does not contain
	close(documentCh)

	if err := store.WriteDocuments(ctx, 42, documentCh); err != nil {
		t.Fatalf("unexpected error while writing documents: %s", err)
	}

	paths, err := store.PathsWithPrefix(ctx, 42, "foo")
	if err != nil {
		t.Fatalf("unexpected error reading from database: %s", err)
	}

	expectedPaths := []string{
		"foo",
		"foo.go",
		"foo/bar.go",
	}
	if diff := cmp.Diff(expectedPaths, paths); diff != "" {
		t.Errorf("unexpected paths (-want +got):\n%s", diff)
	}
}
