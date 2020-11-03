package lsifstore

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func init() {
	dbtesting.DBNameSuffix = "lsifstore"
}

func TestClear(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	for i := 0; i < 5; i++ {
		query := sqlf.Sprintf("INSERT INTO lsif_data_metadata (dump_id, num_result_chunks) VALUES (%s, 0)", i+1)

		if _, err := dbconn.Global.Exec(query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("unexpected error inserting repo: %s", err)
		}
	}

	if err := store.Clear(context.Background(), 2, 4); err != nil {
		t.Fatalf("unexpected error clearing bundle data: %s", err)
	}

	dumpIDs, err := basestore.ScanInts(dbconn.Global.Query("SELECT dump_id FROM lsif_data_metadata"))
	if err != nil {
		t.Fatalf("Unexpected error querying dump identifiers: %s", err)
	}

	if diff := cmp.Diff([]int{1, 3, 5}, dumpIDs); diff != "" {
		t.Errorf("unexpected dump identifiers (-want +got):\n%s", diff)
	}
}

func TestReadWriteMeta(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	if err := store.WriteMeta(context.Background(), 42, types.MetaData{NumResultChunks: 7}); err != nil {
		t.Fatalf("unexpected error while writing: %s", err)
	}

	meta, err := store.ReadMeta(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error reading from database: %s", err)
	}
	if meta.NumResultChunks != 7 {
		t.Errorf("unexpected num result chunks. want=%d have=%d", 7, meta.NumResultChunks)
	}
}

func TestReadWritePathsWithPrefix(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	documentCh := make(chan KeyedDocumentData, 13)
	documentCh <- KeyedDocumentData{Path: "foo.go"}
	documentCh <- KeyedDocumentData{Path: "bar.go"}
	documentCh <- KeyedDocumentData{Path: "baz.go"}
	documentCh <- KeyedDocumentData{Path: "foo/bar/bonk.go"}
	documentCh <- KeyedDocumentData{Path: "foo/baz/bonk.go"}
	documentCh <- KeyedDocumentData{Path: "bar/foo/bonk.go"}
	documentCh <- KeyedDocumentData{Path: "bar/baz/bonk.go"}
	documentCh <- KeyedDocumentData{Path: "baz/foo/bonk.go"}
	documentCh <- KeyedDocumentData{Path: "baz/bar/bonk.go"}
	close(documentCh)

	if err := store.WriteDocuments(context.Background(), 42, documentCh); err != nil {
		t.Fatalf("unexpected error while writing documents: %s", err)
	}

	paths, err := store.PathsWithPrefix(context.Background(), 42, "foo/")
	if err != nil {
		t.Fatalf("unexpected error fetching paths with prefix: %s", err)
	}

	expectedPaths := []string{
		"foo/bar/bonk.go",
		"foo/baz/bonk.go",
	}
	if diff := cmp.Diff(expectedPaths, paths); diff != "" {
		t.Errorf("unexpected paths (-want +got):\n%s", diff)
	}
}

func TestReadWriteDocuments(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	expectedDocumentData := types.DocumentData{
		Ranges: map[types.ID]types.RangeData{
			"r01": {StartLine: 1, StartCharacter: 2, EndLine: 3, EndCharacter: 4, DefinitionResultID: "x01", MonikerIDs: []types.ID{"m01", "m02"}},
			"r02": {StartLine: 2, StartCharacter: 3, EndLine: 4, EndCharacter: 5, ReferenceResultID: "x06", MonikerIDs: []types.ID{"m03", "m04"}},
			"r03": {StartLine: 3, StartCharacter: 4, EndLine: 5, EndCharacter: 6, DefinitionResultID: "x02"},
		},
		HoverResults: map[types.ID]string{},
		Monikers: map[types.ID]types.MonikerData{
			"m01": {Kind: "import", Scheme: "scheme A", Identifier: "ident A", PackageInformationID: "p01"},
			"m02": {Kind: "import", Scheme: "scheme B", Identifier: "ident B"},
			"m03": {Kind: "export", Scheme: "scheme C", Identifier: "ident C", PackageInformationID: "p02"},
			"m04": {Kind: "export", Scheme: "scheme D", Identifier: "ident D"},
		},
		PackageInformation: map[types.ID]types.PackageInformationData{
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

	if err := store.WriteDocuments(context.Background(), 42, documentCh); err != nil {
		t.Fatalf("unexpected error while writing documents: %s", err)
	}

	documentData, _, err := store.ReadDocument(context.Background(), 42, "foo.go")
	if err != nil {
		t.Fatalf("unexpected error reading from database: %s", err)
	}
	if diff := cmp.Diff(expectedDocumentData, documentData); diff != "" {
		t.Errorf("unexpected document data (-want +got):\n%s", diff)
	}
}

func TestReadWriteResultChunks(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	expectedResultChunkData := types.ResultChunkData{
		DocumentPaths: map[types.ID]string{
			"d01": "foo.go",
			"d02": "bar.go",
			"d03": "baz.go",
		},
		DocumentIDRangeIDs: map[types.ID][]types.DocumentIDRangeID{
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

	if err := store.WriteResultChunks(context.Background(), 42, resultChunkCh); err != nil {
		t.Fatalf("unexpected error while writing result chunks: %s", err)
	}

	resultChunkData, _, err := store.ReadResultChunk(context.Background(), 42, 7)
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
	store := testStore()

	expectedDefinitions := []types.Location{
		{URI: "bar.go", StartLine: 4, StartCharacter: 5, EndLine: 6, EndCharacter: 7},
		{URI: "baz.go", StartLine: 7, StartCharacter: 8, EndLine: 9, EndCharacter: 0},
		{URI: "foo.go", StartLine: 3, StartCharacter: 4, EndLine: 5, EndCharacter: 6},
	}

	definitionsCh := make(chan types.MonikerLocations, 1)
	definitionsCh <- types.MonikerLocations{
		Scheme:     "scheme A",
		Identifier: "ident A",
		Locations:  expectedDefinitions,
	}
	close(definitionsCh)

	if err := store.WriteDefinitions(context.Background(), 42, definitionsCh); err != nil {
		t.Fatalf("unexpected error while writing definitions: %s", err)
	}

	definitions, _, err := store.ReadDefinitions(context.Background(), 42, "scheme A", "ident A", 0, 100)
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
	store := testStore()

	expectedReferences := []types.Location{
		{URI: "baz.go", StartLine: 7, StartCharacter: 8, EndLine: 9, EndCharacter: 0},
		{URI: "baz.go", StartLine: 9, StartCharacter: 0, EndLine: 1, EndCharacter: 2},
		{URI: "foo.go", StartLine: 3, StartCharacter: 4, EndLine: 5, EndCharacter: 6},
	}

	referencesCh := make(chan types.MonikerLocations, 1)
	referencesCh <- types.MonikerLocations{
		Scheme:     "scheme C",
		Identifier: "ident C",
		Locations:  expectedReferences,
	}
	close(referencesCh)

	if err := store.WriteReferences(context.Background(), 42, referencesCh); err != nil {
		t.Fatalf("unexpected error while writing references: %s", err)
	}

	references, _, err := store.ReadReferences(context.Background(), 42, "scheme C", "ident C", 0, 100)
	if err != nil {
		t.Fatalf("unexpected error reading from database: %s", err)
	}
	if diff := cmp.Diff(expectedReferences, references); diff != "" {
		t.Errorf("unexpected references (-want +got):\n%s", diff)
	}
}

func testStore() Store {
	// Wrap in observed, as that's how it's used in production
	return NewObserved(New(dbconn.Global), &observation.TestContext)
}
