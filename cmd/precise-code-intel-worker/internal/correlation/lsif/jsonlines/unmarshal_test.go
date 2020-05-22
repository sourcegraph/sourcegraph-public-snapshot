package jsonlines

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/datastructures"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/lsif"
)

func TestUnmarshalElement(t *testing.T) {
	element, err := unmarshalElement([]byte(`{"id": "47", "type": "vertex", "label": "test"}`))
	if err != nil {
		t.Fatalf("unexpected error unmarshalling element data: %s", err)
	}

	expectedElement := lsif.Element{
		ID:    "47",
		Type:  "vertex",
		Label: "test",
	}
	if diff := cmp.Diff(expectedElement, element); diff != "" {
		t.Errorf("unexpected element (-want +got):\n%s", diff)
	}
}

func TestUnmarshalElementNumericIDs(t *testing.T) {
	element, err := unmarshalElement([]byte(`{"id": 47, "type": "vertex", "label": "test"}`))
	if err != nil {
		t.Fatalf("unexpected error unmarshalling element data: %s", err)
	}

	expectedElement := lsif.Element{
		ID:    "47",
		Type:  "vertex",
		Label: "test",
	}
	if diff := cmp.Diff(expectedElement, element); diff != "" {
		t.Errorf("unexpected element (-want +got):\n%s", diff)
	}
}

func TestUnmarshalEdge(t *testing.T) {
	edge, err := unmarshalEdge([]byte(`{"id": "35", "type": "edge", "label": "item", "outV": "12", "inVs": ["07"], "document": "03"}`))
	if err != nil {
		t.Fatalf("unexpected error unmarshalling meta data: %s", err)
	}

	expectedEdge := lsif.Edge{
		OutV:     "12",
		InV:      "",
		InVs:     []string{"07"},
		Document: "03",
	}
	if diff := cmp.Diff(expectedEdge, edge); diff != "" {
		t.Errorf("unexpected edge (-want +got):\n%s", diff)
	}
}

func TestUnmarshalEdgeNumericIDs(t *testing.T) {
	edge, err := unmarshalEdge([]byte(`{"id": 35, "type": "edge", "label": "item", "outV": 12, "inVs": [7], "document": 3}`))
	if err != nil {
		t.Fatalf("unexpected error unmarshalling meta data: %s", err)
	}

	expectedEdge := lsif.Edge{
		OutV:     "12",
		InV:      "",
		InVs:     []string{"7"},
		Document: "3",
	}
	if diff := cmp.Diff(expectedEdge, edge); diff != "" {
		t.Errorf("unexpected edge (-want +got):\n%s", diff)
	}
}

func TestUnmarshalMetaData(t *testing.T) {
	metadata, err := unmarshalMetaData([]byte(`{"id": "01", "type": "vertex", "label": "metaData", "version": "0.4.3", "projectRoot": "file:///test"}`))
	if err != nil {
		t.Fatalf("unexpected error unmarshalling meta data: %s", err)
	}

	expectedMetadata := lsif.MetaData{
		Version:     "0.4.3",
		ProjectRoot: "file:///test",
	}
	if diff := cmp.Diff(expectedMetadata, metadata); diff != "" {
		t.Errorf("unexpected metadata (-want +got):\n%s", diff)
	}
}

func TestUnmarshalDocumentData(t *testing.T) {
	document, err := unmarshalDocumentData([]byte(`{"id": "02", "type": "vertex", "label": "document", "uri": "file:///test/root/foo.go"}`))
	if err != nil {
		t.Fatalf("unexpected error unmarshalling document data: %s", err)
	}

	expectedDocument := lsif.DocumentData{
		URI:      "file:///test/root/foo.go",
		Contains: datastructures.IDSet{},
	}
	if diff := cmp.Diff(expectedDocument, document); diff != "" {
		t.Errorf("unexpected document (-want +got):\n%s", diff)
	}
}

func TestUnmarshalRangeData(t *testing.T) {
	r, err := unmarshalRangeData([]byte(`{"id": "04", "type": "vertex", "label": "range", "start": {"line": 1, "character": 2}, "end": {"line": 3, "character": 4}}`))
	if err != nil {
		t.Fatalf("unexpected error unmarshalling range data: %s", err)
	}

	expectedRange := lsif.RangeData{
		StartLine:          1,
		StartCharacter:     2,
		EndLine:            3,
		EndCharacter:       4,
		DefinitionResultID: "",
		ReferenceResultID:  "",
		HoverResultID:      "",
		MonikerIDs:         datastructures.IDSet{},
	}
	if diff := cmp.Diff(expectedRange, r); diff != "" {
		t.Errorf("unexpected range (-want +got):\n%s", diff)
	}
}

func TestUnmarshalHoverData(t *testing.T) {
	testCases := []struct {
		contents      string
		expectedHover string
	}{
		{
			contents:      `"text"`,
			expectedHover: "text",
		},
		{
			contents:      `[{"kind": "markdown", "value": "text"}]`,
			expectedHover: "text",
		},
		{
			contents:      `[{"language": "go", "value": "text"}]`,
			expectedHover: "```go\ntext\n```",
		},
		{
			contents:      `[{"language": "go", "value": "text"}, {"language": "python", "value": "pext"}]`,
			expectedHover: "```go\ntext\n```\n\n---\n\n```python\npext\n```",
		},
	}

	for _, testCase := range testCases {
		name := fmt.Sprintf("contents=%s", testCase.contents)

		t.Run(name, func(t *testing.T) {
			hover, err := unmarshalHoverData([]byte(fmt.Sprintf(`{"id": "16", "type": "vertex", "label": "hoverResult", "result": {"contents": %s}}`, testCase.contents)))
			if err != nil {
				t.Fatalf("unexpected error unmarshalling hover data: %s", err)
			}

			if diff := cmp.Diff(testCase.expectedHover, hover); diff != "" {
				t.Errorf("unexpected hover text (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUnmarshalMonikerData(t *testing.T) {
	moniker, err := unmarshalMonikerData([]byte(`{"id": "18", "type": "vertex", "label": "moniker", "kind": "import", "scheme": "scheme A", "identifier": "ident A"}`))
	if err != nil {
		t.Fatalf("unexpected error unmarshalling moniker data: %s", err)
	}

	expectedMoniker := lsif.MonikerData{
		Kind:       "import",
		Scheme:     "scheme A",
		Identifier: "ident A",
	}
	if diff := cmp.Diff(expectedMoniker, moniker); diff != "" {
		t.Errorf("unexpected moniker (-want +got):\n%s", diff)
	}
}

func TestUnmarshalPackageInformationData(t *testing.T) {
	packageInformation, err := unmarshalPackageInformationData([]byte(`{"id": "22", "type": "vertex", "label": "packageInformation", "name": "pkg A", "version": "v0.1.0"}`))
	if err != nil {
		t.Fatalf("unexpected error unmarshalling package information data: %s", err)
	}

	expectedPackageInformation := lsif.PackageInformationData{
		Name:    "pkg A",
		Version: "v0.1.0",
	}
	if diff := cmp.Diff(expectedPackageInformation, packageInformation); diff != "" {
		t.Errorf("unexpected package information (-want +got):\n%s", diff)
	}
}
