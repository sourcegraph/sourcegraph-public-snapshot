package reader

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	jsonserializer "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/serializer/json"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/sqliteutil"
)

func init() {
	sqliteutil.SetLocalLibpath()
	sqliteutil.MustRegisterSqlite3WithPcre()
}

func TestReadMeta(t *testing.T) {
	lsifVersion, sourcegraphVersion, numResultChunks, err := testReader(t).ReadMeta(context.Background())
	if err != nil {
		t.Fatalf("unexpected error reading meta: %s", err)
	}

	if lsifVersion != "0.4.3" {
		t.Errorf("unexpected lsifVersion. want=%s have=%s", "0.4.3", lsifVersion)
	}
	if sourcegraphVersion != "0.1.0" {
		t.Errorf("unexpected sourcegraphVersion. want=%s have=%s", "0.1.0", sourcegraphVersion)
	}
	if numResultChunks != 4 {
		t.Errorf("unexpected numResultChunks. want=%d have=%d", 4, numResultChunks)
	}
}

func TestReadDocument(t *testing.T) {
	data, exists, err := testReader(t).ReadDocument(context.Background(), "protocol/writer.go")
	if err != nil {
		t.Fatalf("unexpected error reading document: %s", err)
	}
	if !exists {
		t.Errorf("expected document to exist")
	}

	expectedRange := types.RangeData{
		StartLine:          145,
		StartCharacter:     17,
		EndLine:            145,
		EndCharacter:       28,
		DefinitionResultID: types.ID("2873"),
		ReferenceResultID:  types.ID("16518"),
		HoverResultID:      types.ID("2879"),
		MonikerIDs:         []types.ID{types.ID("2876")},
	}
	if diff := cmp.Diff(expectedRange, data.Ranges[types.ID("2870")]); diff != "" {
		t.Errorf("unexpected range data (-want +got):\n%s", diff)
	}

	expectedHoverData := "```go\n" + `func (*Writer).EmitMoniker(kind string, scheme string, identifier string) (string, error)` + "\n```"
	if diff := cmp.Diff(expectedHoverData, data.HoverResults[types.ID("2879")]); diff != "" {
		t.Errorf("unexpected hover data (-want +got):\n%s", diff)
	}

	expectedMoniker := types.MonikerData{
		Kind:                 "export",
		Scheme:               "gomod",
		Identifier:           "github.com/sourcegraph/lsif-go/protocol:EmitMoniker",
		PackageInformationID: types.ID("213"),
	}
	if diff := cmp.Diff(expectedMoniker, data.Monikers[types.ID("2876")]); diff != "" {
		t.Errorf("unexpected moniker data (-want +got):\n%s", diff)
	}

	expectedPackageInformation := types.PackageInformationData{
		Name:    "github.com/sourcegraph/lsif-go",
		Version: "v0.0.0-ad3507cbeb18",
	}
	if diff := cmp.Diff(expectedPackageInformation, data.PackageInformation[types.ID("213")]); diff != "" {
		t.Errorf("unexpected package information data (-want +got):\n%s", diff)
	}
}

func TestReadResultChunk(t *testing.T) {
	data, exists, err := testReader(t).ReadResultChunk(context.Background(), 3)
	if err != nil {
		t.Fatalf("unexpected error reading result chunk: %s", err)
	}
	if !exists {
		t.Errorf("expected result chunk to exist")
	}

	if path := data.DocumentPaths[types.ID("302")]; path != "protocol/protocol.go" {
		t.Errorf("unexpected document path. want=%s have=%s", "protocol/protocol.go", path)
	}

	expectedDocumentRanges := []types.DocumentIDRangeID{
		{DocumentID: "3981", RangeID: "4940"},
		{DocumentID: "3981", RangeID: "10759"},
		{DocumentID: "3981", RangeID: "10986"},
	}
	if diff := cmp.Diff(expectedDocumentRanges, data.DocumentIDRangeIDs[types.ID("14233")]); diff != "" {
		t.Errorf("unexpected document ranges (-want +got):\n%s", diff)
	}
}

func TestReadDefinitions(t *testing.T) {
	definitions, totalCount, err := testReader(t).ReadDefinitions(context.Background(), "gomod", "github.com/sourcegraph/lsif-go/protocol:Vertex", 3, 4)
	if err != nil {
		t.Fatalf("unexpected error getting definitions: %s", err)
	}
	if totalCount != 11 {
		t.Errorf("unexpected total count. want=%d have=%d", 11, totalCount)
	}

	expectedDefinitions := []types.DefinitionReferenceRow{
		{Scheme: "gomod", Identifier: "github.com/sourcegraph/lsif-go/protocol:Vertex", URI: "protocol/protocol.go", StartLine: 334, StartCharacter: 1, EndLine: 334, EndCharacter: 7},
		{Scheme: "gomod", Identifier: "github.com/sourcegraph/lsif-go/protocol:Vertex", URI: "protocol/protocol.go", StartLine: 139, StartCharacter: 1, EndLine: 139, EndCharacter: 7},
		{Scheme: "gomod", Identifier: "github.com/sourcegraph/lsif-go/protocol:Vertex", URI: "protocol/protocol.go", StartLine: 384, StartCharacter: 1, EndLine: 384, EndCharacter: 7},
		{Scheme: "gomod", Identifier: "github.com/sourcegraph/lsif-go/protocol:Vertex", URI: "protocol/protocol.go", StartLine: 357, StartCharacter: 1, EndLine: 357, EndCharacter: 7},
	}
	if diff := cmp.Diff(expectedDefinitions, definitions); diff != "" {
		t.Errorf("unexpected definitions (-want +got):\n%s", diff)
	}
}

func TestReadReferences(t *testing.T) {
	references, totalCount, err := testReader(t).ReadReferences(context.Background(), "gomod", "golang.org/x/tools/go/packages:Package", 3, 4)
	if err != nil {
		t.Fatalf("unexpected error getting references: %s", err)
	}
	if totalCount != 25 {
		t.Errorf("unexpected total count. want=%d have=%d", 25, totalCount)
	}

	expectedReferences := []types.DefinitionReferenceRow{
		{Scheme: "gomod", Identifier: "golang.org/x/tools/go/packages:Package", URI: "internal/index/helper.go", StartLine: 184, StartCharacter: 56, EndLine: 184, EndCharacter: 63},
		{Scheme: "gomod", Identifier: "golang.org/x/tools/go/packages:Package", URI: "internal/index/helper.go", StartLine: 35, StartCharacter: 56, EndLine: 35, EndCharacter: 63},
		{Scheme: "gomod", Identifier: "golang.org/x/tools/go/packages:Package", URI: "internal/index/helper.go", StartLine: 184, StartCharacter: 35, EndLine: 184, EndCharacter: 42},
		{Scheme: "gomod", Identifier: "golang.org/x/tools/go/packages:Package", URI: "internal/index/helper.go", StartLine: 48, StartCharacter: 44, EndLine: 48, EndCharacter: 51},
	}
	if diff := cmp.Diff(expectedReferences, references); diff != "" {
		t.Errorf("unexpected references (-want +got):\n%s", diff)
	}
}

func testReader(t *testing.T) Reader {
	reader, err := NewSQLiteReader("../testdata/lsif-go@ad3507cb.lsif.db", jsonserializer.New())
	if err != nil {
		t.Fatalf("unexpected error opening database: %s", err)
	}
	t.Cleanup(func() { _ = reader.Close() })

	// Wrap in observed, as that's how it's used in production
	return NewObserved(reader, &observation.TestContext)
}
