package lsif

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/datastructures"
)

func TestUnmarshalDocumentData(t *testing.T) {
	element := Element{
		ID:    "02",
		Type:  "vertex",
		Label: "document",
		Raw:   json.RawMessage(`{"id": "02", "type": "vertex", "label": "document", "uri": "file:///test/root/foo.go"}`),
	}

	document, err := UnmarshalDocumentData(element, "file:///test/root/")
	if err != nil {
		t.Fatalf("unexpected error unmarshalling document data: %s", err)
	}

	expectedDocument := DocumentData{
		URI:      "foo.go",
		Contains: datastructures.IDSet{},
	}
	if diff := cmp.Diff(expectedDocument, document); diff != "" {
		t.Errorf("unexpected document (-want +got):\n%s", diff)
	}
}
