package types

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshalDocumentData(t *testing.T) {
	contents, err := ioutil.ReadFile("../../testdata/documentdata.json")
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	compressed, err := compress(contents)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	actual, err := UnmarshalDocumentData(compressed)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	expected := DocumentData{
		Ranges: map[ID]RangeData{
			ID("7864"): {
				StartLine:          541,
				StartCharacter:     10,
				EndLine:            541,
				EndCharacter:       12,
				DefinitionResultID: ID("1266"),
				ReferenceResultID:  ID("15871"),
				HoverResultID:      ID("1269"),
				MonikerIDs:         nil,
			},
			ID("8265"): {
				StartLine:          266,
				StartCharacter:     10,
				EndLine:            266,
				EndCharacter:       16,
				DefinitionResultID: ID("311"),
				ReferenceResultID:  ID("15500"),
				HoverResultID:      ID("317"),
				MonikerIDs:         []ID{ID("314")},
			},
		},
		HoverResults: map[ID]string{
			ID("1269"): "```go\nvar id string\n```",
			ID("317"):  "```go\ntype Vertex struct\n```\n\n---\n\nVertex contains information of a vertex in the graph.\n\n---\n\n```go\nstruct {\n    Element\n    Label VertexLabel \"json:\\\"label\\\"\"\n}\n```",
		},
		Monikers: map[ID]MonikerData{
			ID("314"): {
				Kind:                 "export",
				Scheme:               "gomod",
				Identifier:           "github.com/sourcegraph/lsif-go/protocol:Vertex",
				PackageInformationID: ID("213"),
			},
			ID("2494"): {
				Kind:                 "export",
				Scheme:               "gomod",
				Identifier:           "github.com/sourcegraph/lsif-go/protocol:VertexLabel",
				PackageInformationID: ID("213"),
			},
		},
		PackageInformation: map[ID]PackageInformationData{
			ID("213"): {
				Name:    "github.com/sourcegraph/lsif-go",
				Version: "v0.0.0-ad3507cbeb18",
			},
		},
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected document data (-want +got):\n%s", diff)
	}
}

func TestUnmarshalResultChunkData(t *testing.T) {
	contents, err := ioutil.ReadFile("../../testdata/resultchunkdata.json")
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	compressed, err := compress(contents)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	actual, err := UnmarshalResultChunkData(compressed)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	expected := ResultChunkData{
		DocumentPaths: map[ID]string{
			ID("4"):   "internal/gomod/module.go",
			ID("302"): "protocol/protocol.go",
			ID("305"): "protocol/writer.go",
		},
		DocumentIDRangeIDs: map[ID][]DocumentIDRangeID{
			ID("34"): {
				{DocumentID: ID("4"), RangeID: ID("31")},
			},
			ID("14040"): {
				{DocumentID: ID("3978"), RangeID: ID("4544")},
			},
			ID("14051"): {
				{DocumentID: ID("3978"), RangeID: ID("4568")},
				{DocumentID: ID("3978"), RangeID: ID("9224")},
				{DocumentID: ID("3978"), RangeID: ID("9935")},
				{DocumentID: ID("3978"), RangeID: ID("9996")},
			},
		},
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected result chunk data (-want +got):\n%s", diff)
	}
}

func compress(uncompressed []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, err := io.Copy(gzipWriter, bytes.NewReader(uncompressed))
	gzipWriter.Close()
	return buf.Bytes(), err
}
