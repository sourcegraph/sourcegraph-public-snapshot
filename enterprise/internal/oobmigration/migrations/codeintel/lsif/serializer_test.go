package lsif

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDocumentData(t *testing.T) {
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

	t.Run("current", func(t *testing.T) {
		serializer := newSerializer()

		recompressed, err := serializer.MarshalDocumentData(expected)
		if err != nil {
			t.Fatalf("unexpected error marshalling document data: %s", err)
		}

		roundtripActual, err := serializer.UnmarshalDocumentData(recompressed)
		if err != nil {
			t.Fatalf("unexpected error unmarshalling document data: %s", err)
		}

		if diff := cmp.Diff(expected, roundtripActual); diff != "" {
			t.Errorf("unexpected document data (-want +got):\n%s", diff)
		}
	})

	t.Run("legacy", func(t *testing.T) {
		serializer := newSerializer()

		recompressed, err := serializer.MarshalLegacyDocumentData(expected)
		if err != nil {
			t.Fatalf("unexpected error marshalling document data: %s", err)
		}

		roundtripActual, err := serializer.UnmarshalLegacyDocumentData(recompressed)
		if err != nil {
			t.Fatalf("unexpected error unmarshalling document data: %s", err)
		}

		if diff := cmp.Diff(expected, roundtripActual); diff != "" {
			t.Errorf("unexpected document data (-want +got):\n%s", diff)
		}
	})
}

func TestLocations(t *testing.T) {
	expected := []LocationData{
		{
			URI:            "internal/index/indexer.go",
			StartLine:      36,
			StartCharacter: 26,
			EndLine:        36,
			EndCharacter:   32,
		},
		{
			URI:            "protocol/writer.go",
			StartLine:      100,
			StartCharacter: 9,
			EndLine:        100,
			EndCharacter:   15,
		},
		{
			URI:            "protocol/writer.go",
			StartLine:      115,
			StartCharacter: 9,
			EndLine:        115,
			EndCharacter:   15,
		},
		{
			URI:            "protocol/writer.go",
			StartLine:      95,
			StartCharacter: 9,
			EndLine:        95,
			EndCharacter:   15,
		},
		{
			URI:            "protocol/writer.go",
			StartLine:      130,
			StartCharacter: 9,
			EndLine:        130,
			EndCharacter:   15,
		},
		{
			URI:            "protocol/writer.go",
			StartLine:      155,
			StartCharacter: 9,
			EndLine:        155,
			EndCharacter:   15,
		},
		{
			URI:            "protocol/writer.go",
			StartLine:      80,
			StartCharacter: 9,
			EndLine:        80,
			EndCharacter:   15,
		},
		{
			URI:            "protocol/writer.go",
			StartLine:      36,
			StartCharacter: 9,
			EndLine:        36,
			EndCharacter:   15,
		},
		{
			URI:            "protocol/writer.go",
			StartLine:      135,
			StartCharacter: 9,
			EndLine:        135,
			EndCharacter:   15,
		},
		{
			URI:            "protocol/writer.go",
			StartLine:      12,
			StartCharacter: 5,
			EndLine:        12,
			EndCharacter:   11,
		},
	}

	serializer := newSerializer()

	recompressed, err := serializer.MarshalLocations(expected)
	if err != nil {
		t.Fatalf("unexpected error marshalling locations: %s", err)
	}

	roundtripActual, err := serializer.UnmarshalLocations(recompressed)
	if err != nil {
		t.Fatalf("unexpected error unmarshalling locations: %s", err)
	}

	if diff := cmp.Diff(expected, roundtripActual); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}
}
